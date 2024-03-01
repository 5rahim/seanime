package streamer

import (
	"os"
	"path"
)

type Transcoder struct {
	// All file streams currently running, index is file path
	streams    CMap[string, *FileStream]
	clientChan chan ClientInfo
	tracker    *Tracker
}

func NewTranscoder() (*Transcoder, error) {
	out := Settings.Outpath
	dir, err := os.ReadDir(out)
	if err != nil {
		return nil, err
	}
	for _, d := range dir {
		err = os.RemoveAll(path.Join(out, d.Name()))
		if err != nil {
			return nil, err
		}
	}

	ret := &Transcoder{
		streams:    NewCMap[string, *FileStream](),
		clientChan: make(chan ClientInfo, 10),
	}
	ret.tracker = NewTracker(ret)
	return ret, nil
}

func (t *Transcoder) getFileStream(path string, route string) (*FileStream, error) {
	var err error
	ret, _ := t.streams.GetOrCreate(path, func() *FileStream {
		sha, err := GetHash(path)
		if err != nil {
			return nil
		}
		return NewFileStream(path, sha, route)
	})
	ret.ready.Wait()
	if err != nil || ret.err != nil {
		t.streams.Remove(path)
		return nil, ret.err
	}
	return ret, nil
}

func (t *Transcoder) GetMaster(path string, client string, route string) (string, error) {
	stream, err := t.getFileStream(path, route)
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
	quality Quality,
	client string,
	route string,
) (string, error) {
	stream, err := t.getFileStream(path, route)
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
	audio int32,
	client string,
	route string,
) (string, error) {
	stream, err := t.getFileStream(path, route)
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
	quality Quality,
	segment int32,
	client string,
	route string,
) (string, error) {
	stream, err := t.getFileStream(path, route)
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
	audio int32,
	segment int32,
	client string,
	route string,
) (string, error) {
	stream, err := t.getFileStream(path, route)
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
