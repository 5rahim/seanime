package streamer

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type Keyframe struct {
	Sha         string
	Keyframes   []float64
	CanTransmux bool
	IsDone      bool
	mutex       sync.RWMutex
	ready       sync.WaitGroup
	listeners   []func(keyframes []float64)
}

func (kf *Keyframe) Get(idx int32) float64 {
	kf.mutex.RLock()
	defer kf.mutex.RUnlock()
	return kf.Keyframes[idx]
}

func (kf *Keyframe) Slice(start int32, end int32) []float64 {
	if end <= start {
		return []float64{}
	}
	kf.mutex.RLock()
	defer kf.mutex.RUnlock()
	ref := kf.Keyframes[start:end]
	ret := make([]float64, end-start)
	copy(ret, ref)
	return ret
}

func (kf *Keyframe) Length() (int32, bool) {
	kf.mutex.RLock()
	defer kf.mutex.RUnlock()
	return int32(len(kf.Keyframes)), kf.IsDone
}

func (kf *Keyframe) add(values []float64) {
	kf.mutex.Lock()
	defer kf.mutex.Unlock()
	kf.Keyframes = append(kf.Keyframes, values...)
	for _, listener := range kf.listeners {
		listener(kf.Keyframes)
	}
}

func (kf *Keyframe) AddListener(callback func(keyframes []float64)) {
	kf.mutex.Lock()
	defer kf.mutex.Unlock()
	kf.listeners = append(kf.listeners, callback)
}

var keyframes = NewCMap[string, *Keyframe]()

func GetKeyframes(sha string, path string) *Keyframe {
	ret, _ := keyframes.GetOrCreate(sha, func() *Keyframe {
		kf := &Keyframe{
			Sha:    sha,
			IsDone: false,
		}
		kf.ready.Add(1)
		go func() {
			save_path := fmt.Sprintf("%s/%s/keyframes.json", Settings.Metadata, sha)
			if err := getSavedInfo(save_path, kf); err == nil {
				log.Printf("Using keyframes cache on filesystem for %s", path)
				kf.ready.Done()
				return
			}

			err := getKeyframes(path, kf)
			if err == nil {
				saveInfo(save_path, kf)
			}
		}()
		return kf
	})
	ret.ready.Wait()
	return ret
}

func getKeyframes(path string, kf *Keyframe) error {
	defer printExecTime("ffprobe analysis for %s", path)()
	// run ffprobe to return all IFrames, IFrames are points where we can split the video in segments.
	// We ask ffprobe to return the time of each frame and it's flags
	// We could ask it to return only i-frames (keyframes) with the -skip_frame nokey but using it is extremly slow
	// since ffmpeg parses every frames when this flag is set.
	cmd := exec.Command(
		"ffprobe",
		"-loglevel", "error",
		"-select_streams", "v:0",
		"-show_entries", "packet=pts_time,flags",
		"-of", "csv=print_section=0",
		path,
	)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)

	ret := make([]float64, 0, 1000)
	max := 100
	done := 0
	for scanner.Scan() {
		frame := scanner.Text()
		if frame == "" {
			continue
		}

		x := strings.Split(frame, ",")
		pts, flags := x[0], x[1]

		// Only take keyframes
		if flags[0] != 'K' {
			continue
		}

		fpts, err := strconv.ParseFloat(pts, 64)
		if err != nil {
			return err
		}

		// Before, we wanted to only save keyframes with at least 3s betweens
		// to prevent segments of 0.2s but sometimes, the -f segment muxer discards
		// the segment time and decide to cut at a random keyframe. Having every keyframe
		// handled as a segment prevents that.

		if done == 0 && len(ret) == 0 {
			// sometimes, videos can start at a timing greater than 0:00. We need to take that into account
			// and only list keyframes that come after the start of the video (without that, our segments count
			// mismatch and we can have the same segment twice on the stream).

			// we hardcode 0 as the first keyframe (even if this is fake) because it makes the code way easier to follow.
			// this value is actually never sent to ffmpeg anyways.
			ret = append(ret, 0)
			continue
		}
		ret = append(ret, fpts)

		if len(ret) == max {
			kf.add(ret)
			if done == 0 {
				kf.ready.Done()
			} else if done >= 500 {
				max = 500
			}
			done += max
			// clear the array without reallocating it
			ret = ret[:0]
		}
	}
	kf.add(ret)
	if done == 0 {
		kf.ready.Done()
	}
	kf.IsDone = true
	return nil
}
