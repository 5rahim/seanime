package transcoder

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"seanime/internal/util"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
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
	kind     string
	handle   StreamHandle
	file     *FileStream
	segments []Segment
	heads    []Head
	// the lock used for the heads
	//lock sync.RWMutex

	segmentsLock sync.RWMutex
	headsLock    sync.RWMutex

	logger   *zerolog.Logger
	settings *Settings
	killCh   chan struct{}
	ctx      context.Context
	cancel   context.CancelFunc
}

type Segment struct {
	// channel open if the segment is not ready. closed if ready.
	// one can check if segment 1 is open by doing:
	//
	//  ts.isSegmentReady(1).
	//
	// You can also wait for it to be ready (non-blocking if already ready) by doing:
	//  <-ts.segments[i]
	channel chan struct{}
	encoder int
}

type Head struct {
	segment int32
	end     int32
	command *exec.Cmd
	stdin   io.WriteCloser
}

var DeletedHead = Head{
	segment: -1,
	end:     -1,
	command: nil,
}

var streamLogger = util.NewLogger()

func NewStream(
	kind string,
	file *FileStream,
	handle StreamHandle,
	ret *Stream,
	settings *Settings,
	logger *zerolog.Logger,
) {
	ret.kind = kind
	ret.handle = handle
	ret.file = file
	ret.heads = make([]Head, 0)
	ret.settings = settings
	ret.logger = logger
	ret.killCh = make(chan struct{})
	ret.ctx, ret.cancel = context.WithCancel(context.Background())

	length, isDone := file.Keyframes.Length()
	ret.segments = make([]Segment, length, max(length, 2000))
	for seg := range ret.segments {
		ret.segments[seg].channel = make(chan struct{})
	}

	if !isDone {
		file.Keyframes.AddListener(func(keyframes []float64) {
			ret.segmentsLock.Lock()
			defer ret.segmentsLock.Unlock()
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
	length, isDone := ts.file.Keyframes.Length()

	for segment := int32(0); segment < length-1; segment++ {
		index += fmt.Sprintf("#EXTINF:%.6f\n", ts.file.Keyframes.Get(segment+1)-ts.file.Keyframes.Get(segment))
		index += fmt.Sprintf("segment-%d.ts\n", segment)
	}
	// do not forget to add the last segment between the last keyframe and the end of the file
	// if the keyframes extraction is not done, do not bother to add it, it will be retrived on the next index retrival
	if isDone {
		index += fmt.Sprintf("#EXTINF:%.6f\n", float64(ts.file.Info.Duration)-ts.file.Keyframes.Get(length-1))
		index += fmt.Sprintf("segment-%d.ts\n", length-1)
		index += `#EXT-X-ENDLIST`
	}
	return index, nil
}

// GetSegment returns the path to the segment and waits for it to be ready.
func (ts *Stream) GetSegment(segment int32) (string, error) {
	// DEVNOTE: Reset the kill channel
	// This is needed because when the segment is needed again, this channel should be open
	ts.killCh = make(chan struct{})
	if debugStream {
		streamLogger.Trace().Msgf("transcoder: Getting segment %d [GetSegment]", segment)
		defer streamLogger.Trace().Msgf("transcoder: Retrieved segment %d [GetSegment]", segment)
	}

	ts.segmentsLock.RLock()
	ts.headsLock.RLock()
	ready := ts.isSegmentReady(segment)
	// we want to calculate distance in the same lock else it can be funky
	distance := 0.
	isScheduled := false
	if !ready {
		distance = ts.getMinEncoderDistance(segment)
		for _, head := range ts.heads {
			if head.segment <= segment && segment < head.end {
				isScheduled = true
				break
			}
		}
	}
	readyChan := ts.segments[segment].channel

	ts.segmentsLock.RUnlock()
	ts.headsLock.RUnlock()

	if !ready {
		// Only start a new encode if there is too big a distance between the current encoder and the segment.
		if distance > 60 || !isScheduled {
			streamLogger.Trace().Msgf("transcoder: New encoder for segment %d", segment)
			err := ts.run(segment)
			if err != nil {
				return "", err
			}
		} else {
			streamLogger.Trace().Msgf("transcoder: Awaiting segment %d - %.2fs gap", segment, distance)
		}

		select {
		// DEVNOTE: This can cause issues if the segment is called again but was "killed" beforehand
		// It's used to interrupt the waiting process but might not be needed since there's a timeout
		case <-ts.killCh:
			return "", fmt.Errorf("transcoder: Stream killed while waiting for segment %d", segment)
		case <-readyChan:
			break
		case <-time.After(25 * time.Second):
			streamLogger.Error().Msgf("transcoder: Could not retrieve %s segment %d (timeout)", ts.kind, segment)
			return "", errors.New("could not retrieve segment (timeout)")
		}
	}
	//go ts.prepareNextSegments(segment)
	ts.prepareNextSegments(segment)
	return fmt.Sprintf(filepath.ToSlash(ts.handle.getOutPath(ts.segments[segment].encoder)), segment), nil
}

// prepareNextSegments will start the next segments if they are not already started.
func (ts *Stream) prepareNextSegments(segment int32) {
	//if ts.IsKilled() {
	//	return
	//}
	// Audio is way cheaper to create than video, so we don't need to run them in advance
	// Running it in advance might actually slow down the video encode since less compute
	// power can be used, so we simply disable that.
	if ts.handle.getFlags()&VideoF == 0 {
		return
	}

	ts.segmentsLock.RLock()
	defer ts.segmentsLock.RUnlock()
	ts.headsLock.RLock()
	defer ts.headsLock.RUnlock()

	for i := segment + 1; i <= min(segment+10, int32(len(ts.segments)-1)); i++ {
		// If the segment is already ready, we don't need to start a new encoder.
		if ts.isSegmentReady(i) {
			continue
		}
		// only start encode for segments not planned (getMinEncoderDistance returns Inf for them)
		// or if they are 60s away (assume 5s per segments)
		if ts.getMinEncoderDistance(i) < 60+(5*float64(i-segment)) {
			continue
		}
		streamLogger.Trace().Msgf("transcoder: Creating new encoder head for future segment %d", i)
		go func() {
			_ = ts.run(i)
		}()
		return
	}
}

func (ts *Stream) getMinEncoderDistance(segment int32) float64 {
	t := ts.file.Keyframes.Get(segment)
	distances := lop.Map(ts.heads, func(head Head, _ int) float64 {
		// ignore killed heads or heads after the current time
		if head.segment < 0 || ts.file.Keyframes.Get(head.segment) > t || segment >= head.end {
			return math.Inf(1)
		}
		return t - ts.file.Keyframes.Get(head.segment)
	})
	if len(distances) == 0 {
		return math.Inf(1)
	}
	return slices.Min(distances)
}

func (ts *Stream) Kill() {
	streamLogger.Trace().Msgf("transcoder: Killing %s stream", ts.kind)
	defer streamLogger.Trace().Msg("transcoder: Stream killed")
	ts.lockHeads()
	defer ts.unlockHeads()

	for id := range ts.heads {
		ts.KillHead(id)
	}
}

func (ts *Stream) IsKilled() bool {
	select {
	case <-ts.killCh:
		// if the channel returned, it means it was closed
		return true
	default:
		return false
	}
}

// KillHead
// Stream is assumed to be locked
func (ts *Stream) KillHead(encoderId int) {
	//streamLogger.Trace().Int("eid", encoderId).Msgf("transcoder: Killing %s encoder head", ts.kind)
	defer streamLogger.Trace().Int("eid", encoderId).Msgf("transcoder: Killed %s encoder head", ts.kind)
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	close(ts.killCh)
	ts.cancel()
	if ts.heads[encoderId] == DeletedHead || ts.heads[encoderId].command == nil {
		return
	}
	ts.heads[encoderId].command.Process.Signal(os.Interrupt)
	//_, _ = ts.heads[encoderId].stdin.Write([]byte("q"))
	//_ = ts.heads[encoderId].stdin.Close()

	ts.heads[encoderId] = DeletedHead
}

func (ts *Stream) SetIsKilled() {
}

//////////////////////////////

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
	return strings.Join(lo.Map(segments, func(seg float64, _ int) string {
		return fmt.Sprintf("%.6f", seg)
	}), ",")
}

func (ts *Stream) run(start int32) error {
	//if ts.IsKilled() {
	//	return nil
	//}
	ts.logger.Trace().Msgf("transcoder: Running %s encoder head from %d", ts.kind, start)
	// Start the transcoder up to the 100th segment (or less)
	length, isDone := ts.file.Keyframes.Length()
	end := min(start+100, length)
	// if keyframes analysis is not finished, always have a 1-segment padding
	// for the extra segment needed for precise split (look comment before -to flag)
	if !isDone {
		end -= 2
	}
	// Stop at the first finished segment
	ts.lockSegments()
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
		ts.unlockSegments()
		return nil
	}
	ts.unlockSegments()

	ts.lockHeads()
	encoderId := len(ts.heads)
	ts.heads = append(ts.heads, Head{segment: start, end: end, command: nil})
	ts.unlockHeads()

	streamLogger.Trace().Any("eid", encoderId).Msgf(
		"transcoder: Transcoding %d-%d/%d segments for %s",
		start,
		end,
		length,
		ts.kind,
	)

	// Include both the start and end delimiter because -ss and -to are not accurate
	// Having an extra segment allows us to cut precisely the segments we want with the
	// -f segment that does cut the beginning and the end at the keyframe like asked
	startRef := float64(0)
	startSeg := start
	if start != 0 {
		// we always take on segment before the current one, for different reasons for audio/video:
		//  - Audio: we need context before the starting point, without that ffmpeg doesn't know what to do and leave ~100ms of silence
		//  - Video: if a segment is really short (between 20 and 100ms), the padding given in the else block bellow is not enough and
		// the previous segment is played another time. the -segment_times is way more precise, so it does not do the same with this one
		startSeg = start - 1
		if ts.handle.getFlags()&AudioF != 0 {
			startRef = ts.file.Keyframes.Get(startSeg)
		} else {
			// the param for the -ss takes the keyframe before the specified time
			// (if the specified time is a keyframe, it either takes that keyframe or the one before)
			// to prevent this weird behavior, we specify a bit after the keyframe that interest us

			// this can't be used with audio since we need to have context before the start-time
			// without this context, the cut loses a bit of audio (audio gap of ~100ms)
			if startSeg+1 == length {
				startRef = (ts.file.Keyframes.Get(startSeg) + float64(ts.file.Info.Duration)) / 2
			} else {
				startRef = (ts.file.Keyframes.Get(startSeg) + ts.file.Keyframes.Get(startSeg+1)) / 2
			}
		}
	}
	endPadding := int32(1)
	if end == length {
		endPadding = 0
	}
	segments := ts.file.Keyframes.Slice(start+1, end+endPadding)
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

	args = append(args, ts.settings.HwAccel.DecodeFlags...)

	if startRef != 0 {
		if ts.handle.getFlags()&VideoF != 0 {
			// This is the default behavior in transmux mode and needed to force pre/post segment to work
			// This must be disabled when processing only audio because it creates gaps in audio
			args = append(args, "-noaccurate_seek")
		}
		args = append(args,
			"-ss", fmt.Sprintf("%.6f", startRef),
		)
	}
	// do not include -to if we want the file to go to the end
	if end+1 < length {
		// sometimes, the duration is shorter than expected (only during transcode it seems)
		// always include more and use the -f segment to split the file where we want
		endRef := ts.file.Keyframes.Get(end + 1)
		// it seems that the -to is confused when -ss seek before the given time (because it searches for a keyframe)
		// add back the time that would be lost otherwise
		// this only happens when -to is before -i but having -to after -i gave a bug (not sure, don't remember)
		endRef += startRef - ts.file.Keyframes.Get(startSeg)
		args = append(args,
			"-to", fmt.Sprintf("%.6f", endRef),
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
		// recommended value is 1/(2*frame_rate), which for a 24fps is ~0.021
		// we take a little bit more than that to be extra safe but too much can be harmful
		// when segments are short (can make the video repeat itself)
		"-segment_time_delta", "0.05",
		"-segment_format", "mpegts",
		"-segment_times", toSegmentStr(lop.Map(segments, func(seg float64, _ int) float64 {
			// segment_times want durations, not timestamps so we must substract the -ss param
			// since we give a greater value to -ss to prevent wrong seeks but -segment_times
			// needs precise segments, we use the keyframe we want to seek to as a reference.
			return seg - ts.file.Keyframes.Get(startSeg)
		})),
		"-segment_list_type", "flat",
		"-segment_list", "pipe:1",
		"-segment_start_number", fmt.Sprint(start),
		outpath,
	)

	// Added logging for ffmpeg command and hardware transcoding state
	streamLogger.Trace().Msgf("transcoder: ffmpeg command: %s %s", ts.settings.FfmpegPath, strings.Join(args, " "))
	if len(ts.settings.HwAccel.DecodeFlags) > 0 {
		streamLogger.Trace().Msgf("transcoder: Hardware transcoding enabled with flags: %v", ts.settings.HwAccel.DecodeFlags)
	} else {
		streamLogger.Trace().Msg("transcoder: Hardware transcoding not enabled")
	}

	cmd := util.NewCmdCtx(context.Background(), ts.settings.FfmpegPath, args...)
	streamLogger.Trace().Msgf("transcoder: Executing ffmpeg for segments %d-%d of %s", start, end, ts.kind)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr

	err = cmd.Start()
	if err != nil {
		return err
	}
	ts.lockHeads()
	ts.heads[encoderId].command = cmd
	ts.heads[encoderId].stdin = stdin
	ts.unlockHeads()

	go func(stdin io.WriteCloser) {
		scanner := bufio.NewScanner(stdout)
		format := filepath.Base(outpath)
		shouldStop := false

		for scanner.Scan() {
			var segment int32
			_, _ = fmt.Sscanf(scanner.Text(), format, &segment)

			// If the segment number is less than the starting segment (start), it means it's not relevant for the current processing, so we skip it
			if segment < start {
				// This happens because we use -f segments for accurate cutting (since -ss is not)
				// check comment at beginning of function for more info
				continue
			}
			ts.lockHeads()
			ts.heads[encoderId].segment = segment
			ts.unlockHeads()
			if debugFfmpegOutput {
				streamLogger.Debug().Int("eid", encoderId).Msgf("t: \t ffmpeg finished segment %d/%d (%d-%d) of %s", segment, end, start, end, ts.kind)
			}

			ts.lockSegments()
			// If the segment is already marked as done, we can stop the ffmpeg process
			if ts.isSegmentReady(segment) {
				// the current segment is already marked as done so another process has already gone up to here.
				_, _ = stdin.Write([]byte("q"))
				_ = stdin.Close()
				//cmd.Process.Signal(os.Interrupt)
				if debugFfmpeg {
					streamLogger.Trace().Int("eid", encoderId).Msgf("transcoder: Terminated ffmpeg, segment %d is ready", segment)
				}
				shouldStop = true
			} else {
				// Mark the segment as ready
				ts.segments[segment].encoder = encoderId
				close(ts.segments[segment].channel)
				if segment == end-1 {
					// file finished, ffmpeg will finish soon on its own
					shouldStop = true
				} else if ts.isSegmentReady(segment + 1) {
					// If the next segment is already marked as done, we can stop the ffmpeg process
					_, _ = stdin.Write([]byte("q"))
					_ = stdin.Close()
					//cmd.Process.Signal(os.Interrupt)
					if debugFfmpeg {
						streamLogger.Trace().Int("eid", encoderId).Msgf("transcoder: Terminated ffmpeg, next segment %d is ready", segment)
					}
					shouldStop = true
				}
			}
			ts.unlockSegments()
			// we need this and not a return in the condition because we want to unlock
			// the lock (and can't defer since this is a loop)
			if shouldStop {
				if debugFfmpeg {
					streamLogger.Trace().Int("eid", encoderId).Msgf("transcoder: ffmpeg completed segments %d-%d/%d of %s", start, end, length, ts.kind)
				}
				return
			}
		}

		if err := scanner.Err(); err != nil {
			streamLogger.Error().Int("eid", encoderId).Err(err).Msg("transcoder: Error scanning ffmpeg output")
			return
		}
	}(stdin)

	// Listen for kill signal
	go func(stdin io.WriteCloser) {
		select {
		case <-ts.ctx.Done():
			streamLogger.Trace().Int("eid", encoderId).Msgf("transcoder: Aborting ffmpeg process for %s", ts.kind)
			_, _ = stdin.Write([]byte("q"))
			_ = stdin.Close()
			return
		}
	}(stdin)

	// Listen for process termination
	go func() {
		err := cmd.Wait()
		var exitErr *exec.ExitError
		// Check if hardware acceleration was attempted and if stderr indicates a failure to use it
		if len(ts.settings.HwAccel.DecodeFlags) > 0 {
			lowerOutput := strings.ToLower(stderr.String())
			if strings.Contains(lowerOutput, "failed") &&
				(strings.Contains(lowerOutput, "hwaccel") || strings.Contains(lowerOutput, "vaapi") || strings.Contains(lowerOutput, "cuvid") || strings.Contains(lowerOutput, "vdpau")) {
				streamLogger.Warn().Int("eid", encoderId).Msg("transcoder: ffmpeg failed to use hardware acceleration settings; falling back to CPU")
			}
		}

		if errors.As(err, &exitErr) && exitErr.ExitCode() == 255 {
			streamLogger.Trace().Int("eid", encoderId).Msgf("transcoder: ffmpeg process was terminated")
		} else if err != nil {
			streamLogger.Error().Int("eid", encoderId).Err(fmt.Errorf("%s: %s", err, stderr.String())).Msgf("transcoder: ffmpeg process failed")
		} else {
			streamLogger.Trace().Int("eid", encoderId).Msgf("transcoder: ffmpeg process for %s exited", ts.kind)
		}

		ts.lockHeads()
		defer ts.unlockHeads()
		// we can't delete the head directly because it would invalidate the others encoderId
		ts.heads[encoderId] = DeletedHead
	}()

	return nil
}

const debugLocks = false
const debugFfmpeg = true
const debugFfmpegOutput = false
const debugStream = false

func (ts *Stream) lockHeads() {
	if debugLocks {
		streamLogger.Debug().Msg("t: Locking heads")
	}
	ts.headsLock.Lock()
	if debugLocks {
		streamLogger.Debug().Msg("t: \t\tLocked heads")
	}
}

func (ts *Stream) unlockHeads() {
	if debugLocks {
		streamLogger.Debug().Msg("t: Unlocking heads")
	}
	ts.headsLock.Unlock()
	if debugLocks {
		streamLogger.Debug().Msg("t: \t\tUnlocked heads")
	}
}

func (ts *Stream) lockSegments() {
	if debugLocks {
		streamLogger.Debug().Msg("t: Locking segments")
	}
	ts.segmentsLock.Lock()
	if debugLocks {
		streamLogger.Debug().Msg("t: \t\tLocked segments")
	}
}

func (ts *Stream) unlockSegments() {
	if debugLocks {
		streamLogger.Debug().Msg("t: Unlocking segments")
	}
	ts.segmentsLock.Unlock()
	if debugLocks {
		streamLogger.Debug().Msg("t: \t\tUnlocked segments")
	}
}
