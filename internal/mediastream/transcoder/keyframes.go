package transcoder

import (
	"bufio"
	"github.com/rs/zerolog"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type Keyframe struct {
	Sha         string
	Keyframes   []float64
	CanTransmux bool
	IsDone      bool
	info        *KeyframeInfo
}
type KeyframeInfo struct {
	mutex     sync.RWMutex
	ready     sync.WaitGroup
	listeners []func(keyframes []float64)
}

func (kf *Keyframe) Get(idx int32) float64 {
	kf.info.mutex.RLock()
	defer kf.info.mutex.RUnlock()
	return kf.Keyframes[idx]
}

func (kf *Keyframe) Slice(start int32, end int32) []float64 {
	if end <= start {
		return []float64{}
	}
	kf.info.mutex.RLock()
	defer kf.info.mutex.RUnlock()
	ref := kf.Keyframes[start:end]
	ret := make([]float64, end-start)
	copy(ret, ref)
	return ret
}

func (kf *Keyframe) Length() (int32, bool) {
	kf.info.mutex.RLock()
	defer kf.info.mutex.RUnlock()
	return int32(len(kf.Keyframes)), kf.IsDone
}

func (kf *Keyframe) add(values []float64) {
	kf.info.mutex.Lock()
	defer kf.info.mutex.Unlock()
	kf.Keyframes = append(kf.Keyframes, values...)
	for _, listener := range kf.info.listeners {
		listener(kf.Keyframes)
	}
}

func (kf *Keyframe) AddListener(callback func(keyframes []float64)) {
	kf.info.mutex.Lock()
	defer kf.info.mutex.Unlock()
	kf.info.listeners = append(kf.info.listeners, callback)
}

var keyframes = NewCMap[string, *Keyframe]()

func GetKeyframes(
	sha string,
	path string,
	logger *zerolog.Logger,
	settings *Settings,
) *Keyframe {
	ret, _ := keyframes.GetOrCreate(sha, func() *Keyframe {
		kf := &Keyframe{
			Sha:    sha,
			IsDone: false,
			info:   &KeyframeInfo{},
		}
		kf.info.ready.Add(1)
		go func() {
			keyframesPath := filepath.Join(settings.MetadataDir, sha, "keyframes.json")
			if err := getSavedInfo(keyframesPath, kf); err == nil {
				logger.Trace().Msgf("transcoder: Keyframes Cache HIT")
				kf.info.ready.Done()
				return
			}

			err := getKeyframes(path, kf, logger)
			if err == nil {
				saveInfo(keyframesPath, kf)
			}
		}()
		return kf
	})
	ret.info.ready.Wait()
	return ret
}

func getKeyframes(path string, kf *Keyframe, logger *zerolog.Logger) error {
	defer printExecTime(logger, "ffprobe analysis for %s", path)()
	// Execute ffprobe to retrieve all IFrames. IFrames are specific points in the video we can divide it into segments.
	// We instruct ffprobe to return the timestamp and flags of each frame.
	// Although it's possible to request ffprobe to return only i-frames (keyframes) using the -skip_frame nokey option, this approach is highly inefficient.
	// The inefficiency arises because when this option is used, ffmpeg processes every single frame, which significantly slows down the operation.
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

		// Previously, the aim was to save only those keyframes that had a minimum gap of 3 seconds between them.
		// This was to avoid creating segments as short as 0.2 seconds.
		// However, there were instances where the -f segment muxer would ignore the specified segment time and choose a random keyframe to cut at.
		// To counter this, treat every keyframe as a potential segment.
		if done == 0 && len(ret) == 0 {

			// There are instances where videos may not start exactly at 0:00. This needs to be considered,
			// and we should only include keyframes that occur after the video's start time. If not done so,
			// it can lead to a discrepancy in our segment count and potentially duplicate the same segment in the stream.

			// For simplicity in code comprehension, we designate 0 as the initial keyframe, even though it's not genuine.
			// This value is never actually passed to ffmpeg.
			ret = append(ret, 0)
			continue
		}
		ret = append(ret, fpts)

		if len(ret) == max {
			kf.add(ret)
			if done == 0 {
				kf.info.ready.Done()
			} else if done >= 500 {
				max = 500
			}
			done += max
			// clear the array without reallocing it
			ret = ret[:0]
		}
	}
	kf.add(ret)
	if done == 0 {
		kf.info.ready.Done()
	}
	kf.IsDone = true
	return nil
}
