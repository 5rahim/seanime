package transcoder

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/mediastream/videofile"
	"os"
	"path"
	"path/filepath"
)

type (
	Transcoder struct {
		// All file streams currently running, index is file path
		streams    CMap[string, *FileStream]
		clientChan chan ClientInfo
		tracker    *Tracker
		logger     *zerolog.Logger
		settings   Settings
	}

	Settings struct {
		StreamDir string
		HwAccel   HwAccelSettings
	}

	NewTranscoderOptions struct {
		Logger      *zerolog.Logger
		HwAccelKind string
		Preset      string
		TempOutDir  string
	}
)

func NewTranscoder(opts *NewTranscoderOptions) (*Transcoder, error) {

	// Create/clear the temp directory containing the streams
	streamDir := filepath.Join(opts.TempOutDir, "streams")
	_ = os.MkdirAll(streamDir, 0755)
	dir, err := os.ReadDir(streamDir)
	if err != nil {
		return nil, err
	}
	for _, d := range dir {
		_ = os.RemoveAll(path.Join(streamDir, d.Name()))
	}

	ret := &Transcoder{
		streams:    NewCMap[string, *FileStream](),
		clientChan: make(chan ClientInfo, 10),
		logger:     opts.Logger,
		settings: Settings{
			StreamDir: streamDir,
			HwAccel: GetHardwareAccelSettings(HwAccelOptions{
				Kind:   opts.HwAccelKind,
				Preset: opts.Preset,
			}),
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
	t.streams.lock.Lock()

	t.logger.Debug().Msg("transcoder: Destroying transcoder")
	for _, s := range t.streams.data {
		s.Destroy()
	}
	//close(t.clientChan)
	t.streams = NewCMap[string, *FileStream]()
	t.clientChan = make(chan ClientInfo, 10)
	t.logger.Debug().Msg("transcoder: Transcoder destroyed")
}

func (t *Transcoder) getFileStream(path string, hash string, mediaInfo *videofile.MediaInfo) (*FileStream, error) {
	var err error
	ret, _ := t.streams.GetOrCreate(path, func() *FileStream {
		return NewFileStream(path, hash, mediaInfo, &t.settings, t.logger)
	})
	ret.ready.Wait()
	if ret == nil {
		return nil, fmt.Errorf("could not get filestream, file may not exist")
	}
	if err != nil || ret.err != nil {
		t.streams.Remove(path)
		return nil, ret.err
	}
	return ret, nil
}

func (t *Transcoder) GetMaster(path string, hash string, mediaInfo *videofile.MediaInfo, client string) (string, error) {
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
	stream, err := t.getFileStream(path, hash, mediaInfo)
	if err != nil {
		return "", err
	}
	t.clientChan <- ClientInfo{
		client:  client,
		path:    path,
		quality: &quality,
		audio:   -1,
		head:    segment,
	}
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
