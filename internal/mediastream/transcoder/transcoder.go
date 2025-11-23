package transcoder

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"seanime/internal/mediastream/videofile"
	"seanime/internal/util/result"
	"time"

	"github.com/rs/zerolog"
)

type (
	Transcoder struct {
		// All file streams currently running, index is file path
		streams    *result.Map[string, *FileStream]
		clientChan chan ClientInfo
		tracker    *Tracker
		logger     *zerolog.Logger
		settings   Settings
	}

	Settings struct {
		StreamDir   string
		HwAccel     HwAccelSettings
		FfmpegPath  string
		FfprobePath string
	}

	NewTranscoderOptions struct {
		Logger                *zerolog.Logger
		HwAccelKind           string
		Preset                string
		TempOutDir            string
		FfmpegPath            string
		FfprobePath           string
		HwAccelCustomSettings string
	}
)

func NewTranscoder(opts *NewTranscoderOptions) (*Transcoder, error) {

	// Create a directory that'll hold the stream segments if it doesn't exist
	streamDir := filepath.Join(opts.TempOutDir, "streams")
	_ = os.MkdirAll(streamDir, 0755)

	// Clear the directory containing the streams
	dir, err := os.ReadDir(streamDir)
	if err != nil {
		return nil, err
	}
	for _, d := range dir {
		_ = os.RemoveAll(path.Join(streamDir, d.Name()))
	}

	ret := &Transcoder{
		streams:    result.NewMap[string, *FileStream](),
		clientChan: make(chan ClientInfo, 1000),
		logger:     opts.Logger,
		settings: Settings{
			StreamDir: streamDir,
			HwAccel: GetHardwareAccelSettings(HwAccelOptions{
				Kind:           opts.HwAccelKind,
				Preset:         opts.Preset,
				CustomSettings: opts.HwAccelCustomSettings,
			}),
			FfmpegPath:  opts.FfmpegPath,
			FfprobePath: opts.FfprobePath,
		},
	}
	ret.tracker = NewTracker(ret)

	ret.logger.Info().Msg("transcoder: Initialized")
	return ret, nil
}

func (t *Transcoder) GetSettings() *Settings {
	return &t.settings
}

// Destroy stops all streams and removes the output directory.
// A new transcoder should be created after calling this function.
func (t *Transcoder) Destroy() {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	t.tracker.Stop()

	t.logger.Debug().Msg("transcoder: Destroying transcoder")
	for _, s := range t.streams.Values() {
		s.Destroy()
	}
	t.streams.Clear()
	//close(t.clientChan)
	t.streams = result.NewMap[string, *FileStream]()
	t.clientChan = make(chan ClientInfo, 10)
	t.logger.Debug().Msg("transcoder: Transcoder destroyed")
}

func (t *Transcoder) getFileStream(path string, hash string, mediaInfo *videofile.MediaInfo) (*FileStream, error) {
	if debugStream {
		start := time.Now()
		t.logger.Trace().Msgf("transcoder: Getting filestream")
		defer t.logger.Trace().Msgf("transcoder: Filestream retrieved in %.2fs", time.Since(start).Seconds())
	}
	ret, _ := t.streams.GetOrSet(path, func() (*FileStream, error) {
		return NewFileStream(path, hash, mediaInfo, &t.settings, t.logger), nil
	})
	if ret == nil {
		return nil, fmt.Errorf("could not get filestream, file may not exist")
	}
	ret.ready.Wait()
	if ret.err != nil {
		t.streams.Delete(path)
		return nil, ret.err
	}
	return ret, nil
}

func (t *Transcoder) GetMaster(path string, hash string, mediaInfo *videofile.MediaInfo, client string) (string, error) {
	if debugStream {
		start := time.Now()
		t.logger.Trace().Msgf("transcoder: Retrieving master file")
		defer t.logger.Trace().Msgf("transcoder: Master file retrieved in %.2fs", time.Since(start).Seconds())
	}
	stream, err := t.getFileStream(path, hash, mediaInfo)
	if err != nil {
		return "", err
	}
	t.clientChan <- ClientInfo{
		client:  client,
		path:    path,
		quality: nil,
		audio:   -1,
		head:    -1,
	}
	return stream.GetMaster(), nil
}

func (t *Transcoder) GetVideoIndex(
	path string,
	hash string,
	mediaInfo *videofile.MediaInfo,
	quality Quality,
	client string,
) (string, error) {
	if debugStream {
		start := time.Now()
		t.logger.Trace().Msgf("transcoder: Retrieving video index file (%s)", quality)
		defer t.logger.Trace().Msgf("transcoder: Video index file retrieved in %.2fs", time.Since(start).Seconds())
	}
	stream, err := t.getFileStream(path, hash, mediaInfo)
	if err != nil {
		return "", err
	}
	t.clientChan <- ClientInfo{
		client:  client,
		path:    path,
		quality: &quality,
		audio:   -1,
		head:    -1,
	}
	return stream.GetVideoIndex(quality)
}

func (t *Transcoder) GetAudioIndex(
	path string,
	hash string,
	mediaInfo *videofile.MediaInfo,
	audio int32,
	client string,
) (string, error) {
	if debugStream {
		start := time.Now()
		t.logger.Trace().Msgf("transcoder: Retrieving audio index file (%d)", audio)
		defer t.logger.Trace().Msgf("transcoder: Audio index file retrieved in %.2fs", time.Since(start).Seconds())
	}
	stream, err := t.getFileStream(path, hash, mediaInfo)
	if err != nil {
		return "", err
	}
	t.clientChan <- ClientInfo{
		client: client,
		path:   path,
		audio:  audio,
		head:   -1,
	}
	return stream.GetAudioIndex(audio)
}

func (t *Transcoder) GetVideoSegment(
	path string,
	hash string,
	mediaInfo *videofile.MediaInfo,
	quality Quality,
	segment int32,
	client string,
) (string, error) {
	if debugStream {
		start := time.Now()
		t.logger.Trace().Msgf("transcoder: Retrieving video segment %d (%s) [GetVideoSegment]", segment, quality)
		defer t.logger.Trace().Msgf("transcoder: Video segment retrieved in %.2fs", time.Since(start).Seconds())
	}
	stream, err := t.getFileStream(path, hash, mediaInfo)
	if err != nil {
		return "", err
	}
	//t.logger.Trace().Msgf("transcoder: Sending client info, segment %d (%s) [GetVideoSegment]", segment, quality)
	t.clientChan <- ClientInfo{
		client:  client,
		path:    path,
		quality: &quality,
		audio:   -1,
		head:    segment,
	}
	//t.logger.Trace().Msgf("transcoder: Getting video segment %d (%s) [GetVideoSegment]", segment, quality)
	return stream.GetVideoSegment(quality, segment)
}

func (t *Transcoder) GetAudioSegment(
	path string,
	hash string,
	mediaInfo *videofile.MediaInfo,
	audio int32,
	segment int32,
	client string,
) (string, error) {
	if debugStream {
		start := time.Now()
		t.logger.Trace().Msgf("transcoder: Retrieving audio segment %d (%d)", segment, audio)
		defer t.logger.Trace().Msgf("transcoder: Audio segment %d (%d) retrieved in %.2fs", segment, audio, time.Since(start).Seconds())
	}
	stream, err := t.getFileStream(path, hash, mediaInfo)
	if err != nil {
		return "", err
	}
	t.clientChan <- ClientInfo{
		client: client,
		path:   path,
		audio:  audio,
		head:   segment,
	}
	return stream.GetAudioSegment(audio, segment)
}
