package streamer

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

type Flags int32

const (
	AudioF   Flags = 1 << 0
	VideoF   Flags = 1 << 1
	Transmux Flags = 1 << 3
)

type StreamHandle interface {
	getTranscodeArgs(segments string) []string
	getOutPath(encoderId int) string
	getFlags() Flags
}

type Stream struct {
	handle   StreamHandle
	file     *FileStream
	segments []Segment
	heads    []Head
	// the lock used for the heads
	lock sync.RWMutex
}

type Segment struct {
	// channel open if the segment is not ready. closed if ready.
	// one can check if segment 1 is open by doing:
	//
	//  ts.isSegmentReady(1).
	//
	// You can also wait for it to be ready (non-blocking if already ready) by doing:
	//  <-ts.segments[i]
	channel chan (struct{})
	encoder int
}

type Head struct {
	segment int32
	end     int32
	command *exec.Cmd
}

var DeletedHead = Head{
	segment: -1,
	end:     -1,
	command: nil,
}

func NewStream(file *FileStream, handle StreamHandle, ret *Stream) {
	ret.handle = handle
	ret.file = file
	ret.heads = make([]Head, 0)

	length, isDone := file.Keyframes.Length()
	ret.segments = make([]Segment, length, 2000)
	for seg := range ret.segments {
		ret.segments[seg].channel = make(chan struct{})
	}

	if !isDone {
		file.Keyframes.AddListener(func(keyframes []float64) {
			ret.lock.Lock()
			defer ret.lock.Unlock()
			oldLength := len(ret.segments)
			if cap(ret.segments) > len(keyframes) {
				ret.segments = ret.segments[:len(keyframes)]
			} else {
				ret.segments = append(ret.segments, make([]Segment, len(keyframes)-oldLength)...)
			}
			for seg := oldLength; seg < len(keyframes); seg++ {
				ret.segments[seg].channel = make(chan struct{})
			}
		})
	}
}

// Remember to lock before calling this.
func (ts *Stream) isSegmentReady(segment int32) bool {
	select {
	case <-ts.segments[segment].channel:
		// if the channel returned, it means it was closed
		return true
	default:
		return false
	}
}

func (ts *Stream) isSegmentTranscoding(segment int32) bool {
	for _, head := range ts.heads {
		if head.segment == segment {
			return true
		}
	}
	return false
}

func toSegmentStr(segments []float64) string {
	return strings.Join(Map(segments, func(seg float64, _ int) string {
		return fmt.Sprintf("%.6f", seg)
	}), ",")
}

func (ts *Stream) run(start int32) error {
	// Start the transcoder up to the 100th segment (or less)
	length, isDone := ts.file.Keyframes.Length()
	end := min(start+100, length)
	// if keyframes analysis is not finished, always have a 1-segment padding
	// for the extra segment needed for precise split (look comment before -to flag)
	if !isDone {
		end -= 2
	}
	// Stop at the first finished segment
	ts.lock.Lock()
	for i := start; i < end; i++ {
		if ts.isSegmentReady(i) || ts.isSegmentTranscoding(i) {
			end = i
			break
		}
	}
	if start >= end {
		// this can happen if the start segment was finished between the check
		// to call run() and the actual call.
		// since most checks are done in a RLock() instead of a Lock() this can
		// happens when two goroutines try to make the same segment ready
		ts.lock.Unlock()
		return nil
	}
	encoderId := len(ts.heads)
	ts.heads = append(ts.heads, Head{segment: start, end: end, command: nil})
	ts.lock.Unlock()

	log.Printf(
		"Transcoding %d for %s (from %d to %d out of %d segments)",
		encoderId,
		ts.file.Path,
		start,
		end,
		length,
	)

	// Include both the start and end delimiter because -ss and -to are not accurate
	// Having an extra segment allows us to cut precisely the segments we want with the
	// -f segment that does cut the beginning and the end at the keyframe like asked
	start_ref := float64(0)
	start_segment := start
	if start != 0 {
		if ts.handle.getFlags()&AudioF != 0 {
			// when segmenting audio, we need -ss to have the context before the start time
			// without it, the cut loses a bit of audio (audio gap of ~100ms)
			start_segment = start - 1
		}
		// the param for the -ss takes the keyframe before the specified time
		// (if the specified time is a keyframe, it either takes that keyframe or the one before)
		// to prevent this weird behavior, we specify a bit after the keyframe that interest us
		if start_segment+1 == length {
			start_ref = (ts.file.Keyframes.Get(start_segment) + float64(ts.file.Info.Duration)) / 2
		} else {
			start_ref = (ts.file.Keyframes.Get(start_segment) + ts.file.Keyframes.Get(start_segment+1)) / 2
		}
	}
	end_padding := int32(1)
	if end == length {
		end_padding = 0
	}
	segments := ts.file.Keyframes.Slice(start+1, end+end_padding)
	if len(segments) == 0 {
		// we can't leave that empty else ffmpeg errors out.
		segments = []float64{9999999}
	}

	outpath := ts.handle.getOutPath(encoderId)
	err := os.MkdirAll(filepath.Dir(outpath), 0755)
	if err != nil {
		return err
	}

	args := []string{
		"-nostats", "-hide_banner", "-loglevel", "warning",
	}

	args = append(args, Settings.HwAccel.DecodeFlags...)

	if start_ref != 0 {
		if ts.handle.getFlags()&VideoF != 0 {
			// This is the default behavior in transmux mode and needed to force pre/post segment to work
			// This must be disabled when processing only audio because it creates gaps in audio
			args = append(args, "-noaccurate_seek")
		}
		args = append(args,
			"-ss", fmt.Sprintf("%.6f", start_ref),
		)
	}
	// do not include -to if we want the file to go to the end
	if end+1 < length {
		// sometimes, the duration is shorter than expected (only during transcode it seems)
		// always include more and use the -f segment to split the file where we want
		end_ref := ts.file.Keyframes.Get(end + 1)
		// it seems that the -to is confused when -ss seek before the given time (because it searches for a keyframe)
		// add back the time that would be lost otherwise
		// this only appens when -to is before -i but having -to after -i gave a bug (not sure, don't remember)
		end_ref += start_ref - ts.file.Keyframes.Get(start_segment)
		args = append(args,
			"-to", fmt.Sprintf("%.6f", end_ref),
		)
	}
	args = append(args,
		"-i", ts.file.Path,
		// this makes behaviors consistent between soft and hardware decodes.
		// this also means that after a -ss 50, the output video will start at 50s
		"-start_at_zero",
		// for hls streams, -copyts is mandatory
		"-copyts",
		// this makes output file start at 0s instead of a random delay + the -ss value
		// this also cancel -start_at_zero weird delay.
		// this is not always respected, but generally it gives better results.
		// even when this is not respected, it does not result in a bugged experience but this is something
		// to keep in mind when debugging
		"-muxdelay", "0",
	)
	args = append(args, ts.handle.getTranscodeArgs(toSegmentStr(segments))...)
	args = append(args,
		"-f", "segment",
		// needed for rounding issues when forcing keyframes
		"-segment_time_delta", "0.2",
		"-segment_format", "mpegts",
		"-segment_times", toSegmentStr(Map(segments, func(seg float64, _ int) float64 {
			// segment_times want durations, not timestamps so we must substract the -ss param
			// since we give a greater value to -ss to prevent wrong seeks but -segment_times
			// needs precise segments, we use the keyframe we want to seek to as a reference.
			return seg - ts.file.Keyframes.Get(start_segment)
		})),
		"-segment_list_type", "flat",
		"-segment_list", "pipe:1",
		"-segment_start_number", fmt.Sprint(start),
		outpath,
	)

	cmd := exec.Command("ffmpeg", args...)
	log.Printf("Running %s", strings.Join(cmd.Args, " "))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr

	err = cmd.Start()
	if err != nil {
		return err
	}
	ts.lock.Lock()
	ts.heads[encoderId].command = cmd
	ts.lock.Unlock()

	go func() {
		scanner := bufio.NewScanner(stdout)
		format := filepath.Base(outpath)
		shouldStop := false

		for scanner.Scan() {
			var segment int32
			_, _ = fmt.Sscanf(scanner.Text(), format, &segment)

			if segment < start {
				// This happens because we use -f segments for accurate cutting (since -ss is not)
				// check comment at beginning of function for more info
				continue
			}
			ts.lock.Lock()
			ts.heads[encoderId].segment = segment
			log.Printf("Segment %d ready (EID: %d)", segment, encoderId)
			if ts.isSegmentReady(segment) {
				// the current segment is already marked at done so another process has already gone up to here.
				cmd.Process.Signal(os.Interrupt)
				log.Printf("Killing ffmpeg because segment %d is already ready", segment)
				shouldStop = true
			} else {
				ts.segments[segment].encoder = encoderId
				close(ts.segments[segment].channel)
				if segment == end-1 {
					// file finished, ffmped will finish soon on its own
					shouldStop = true
				} else if ts.isSegmentReady(segment + 1) {
					cmd.Process.Signal(os.Interrupt)
					log.Printf("Killing ffmpeg because next segment %d is ready", segment)
					shouldStop = true
				}
			}
			ts.lock.Unlock()
			// we need this and not a return in the condition because we want to unlock
			// the lock (and can't defer since this is a loop)
			if shouldStop {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			log.Println("Error reading stdout of ffmpeg", err)
		}
	}()

	go func() {
		err := cmd.Wait()
		if exiterr, ok := err.(*exec.ExitError); ok && exiterr.ExitCode() == 255 {
			log.Printf("ffmpeg %d was killed by us", encoderId)
		} else if err != nil {
			log.Printf("ffmpeg %d occured an error: %s: %s", encoderId, err, stderr.String())
		} else {
			log.Printf("ffmpeg %d finished successfully", encoderId)
		}

		ts.lock.Lock()
		defer ts.lock.Unlock()
		// we can't delete the head directly because it would invalidate the others encoder_id
		ts.heads[encoderId] = DeletedHead
	}()

	return nil
}

func (ts *Stream) GetIndex() (string, error) {
	// playlist type is event since we can append to the list if Keyframe.IsDone is false.
	// start time offset makes the stream start at 0s instead of ~3segments from the end (requires version 6 of hls)
	index := `#EXTM3U
#EXT-X-VERSION:6
#EXT-X-PLAYLIST-TYPE:EVENT
#EXT-X-START:TIME-OFFSET=0
#EXT-X-TARGETDURATION:4
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-INDEPENDENT-SEGMENTS
`
	length, is_done := ts.file.Keyframes.Length()

	for segment := int32(0); segment < length-1; segment++ {
		index += fmt.Sprintf("#EXTINF:%.6f\n", ts.file.Keyframes.Get(segment+1)-ts.file.Keyframes.Get(segment))
		index += fmt.Sprintf("segment-%d.ts\n", segment)
	}
	// do not forget to add the last segment between the last keyframe and the end of the file
	// if the keyframes extraction is not done, do not bother to add it, it will be retrived on the next index retrival
	if is_done {
		index += fmt.Sprintf("#EXTINF:%.6f\n", float64(ts.file.Info.Duration)-ts.file.Keyframes.Get(length-1))
		index += fmt.Sprintf("segment-%d.ts\n", length-1)
		index += `#EXT-X-ENDLIST`
	}
	return index, nil
}

func (ts *Stream) GetSegment(segment int32) (string, error) {
	ts.lock.RLock()
	ready := ts.isSegmentReady(segment)
	// we want to calculate distance in the same lock else it can be funky
	distance := 0.
	is_scheduled := false
	if !ready {
		distance = ts.getMinEncoderDistance(segment)
		for _, head := range ts.heads {
			if head.segment <= segment && segment < head.end {
				is_scheduled = true
				break
			}
		}
	}
	readyChan := ts.segments[segment].channel
	ts.lock.RUnlock()

	if !ready {
		// Only start a new encode if there is too big a distance between the current encoder and the segment.
		if distance > 60 || !is_scheduled {
			log.Printf("Creating new head for %d since closest head is %fs aways", segment, distance)
			err := ts.run(segment)
			if err != nil {
				return "", err
			}
		} else {
			log.Printf("Waiting for segment %d since encoder head is %fs aways", segment, distance)
		}

		select {
		case <-readyChan:
		case <-time.After(60 * time.Second):
			return "", errors.New("could not retrieve the selected segment (timeout)")
		}
	}
	ts.prepareNextSegments(segment)
	return fmt.Sprintf(filepath.ToSlash(ts.handle.getOutPath(ts.segments[segment].encoder)), segment), nil
}

func (ts *Stream) prepareNextSegments(segment int32) {
	// Audio is way cheaper to create than video, so we don't need to run them in advance
	// Running it in advance might actually slow down the video encode since less compute
	// power can be used, so we simply disable that.
	if ts.handle.getFlags()&VideoF == 0 {
		return
	}
	ts.lock.RLock()
	defer ts.lock.RUnlock()
	for i := segment + 1; i <= min(segment+10, int32(len(ts.segments)-1)); i++ {
		if ts.isSegmentReady(i) {
			continue
		}
		// only start encode for segments not planned (getMinEncoderDistance returns Inf for them)
		// or if they are 60s away (assume 5s per segments)
		if ts.getMinEncoderDistance(i) < 60+(5*float64(i-segment)) {
			continue
		}
		log.Printf("Creating new head for future segment (%d)", i)
		go ts.run(i)
		return
	}
}

func (ts *Stream) getMinEncoderDistance(segment int32) float64 {
	time := ts.file.Keyframes.Get(segment)
	distances := Map(ts.heads, func(head Head, _ int) float64 {
		// ignore killed heads or heads after the current time
		if head.segment < 0 || ts.file.Keyframes.Get(head.segment) > time || segment >= head.end {
			return math.Inf(1)
		}
		return time - ts.file.Keyframes.Get(head.segment)
	})
	if len(distances) == 0 {
		return math.Inf(1)
	}
	return slices.Min(distances)
}

func (ts *Stream) Kill() {
	ts.lock.Lock()
	defer ts.lock.Unlock()

	for id := range ts.heads {
		ts.KillHead(id)
	}
}

// Stream assume to be locked
func (ts *Stream) KillHead(encoder_id int) {
	if ts.heads[encoder_id] == DeletedHead || ts.heads[encoder_id].command == nil {
		return
	}
	ts.heads[encoder_id].command.Process.Signal(os.Interrupt)
	ts.heads[encoder_id] = DeletedHead
}
