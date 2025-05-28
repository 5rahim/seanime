package directstream

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/neilotoole/streamcache"
)

func ServeLocalFile(w http.ResponseWriter, r *http.Request, lfStream *LocalFileStream) {
	if lfStream.serveContentCancelFunc != nil {
		lfStream.serveContentCancelFunc()
	}

	_, cancel := context.WithCancel(lfStream.manager.playbackCtx)
	lfStream.serveContentCancelFunc = cancel

	reader, err := lfStream.newReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	playbackInfo, err := lfStream.LoadPlaybackInfo()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	size := playbackInfo.ContentLength
	w.Header().Set("Content-Length", fmt.Sprint(size))

	http.ServeContent(w, r, lfStream.localFile.Path, time.Now(), reader)
}

///////////////////////////////////////////////////////////////////////////////////////////////////////

type StreamCacheReadSeekCloser struct {
	stream         *streamcache.Stream
	streamReader   *streamcache.Reader
	originalReader io.ReadSeekCloser
}

var _ io.ReadSeekCloser = (*StreamCacheReadSeekCloser)(nil)

func NewStreamCacheReadSeekCloser(ctx context.Context, reader io.ReadSeekCloser) StreamCacheReadSeekCloser {
	stream := streamcache.New(reader)
	return StreamCacheReadSeekCloser{
		stream:         stream,
		streamReader:   stream.NewReader(ctx),
		originalReader: reader,
	}
}

func (s StreamCacheReadSeekCloser) Read(p []byte) (n int, err error) {
	return s.streamReader.Read(p)
}

func (s StreamCacheReadSeekCloser) Seek(offset int64, whence int) (int64, error) {
	return s.originalReader.Seek(offset, whence)
}

func (s StreamCacheReadSeekCloser) Close() error {
	return s.originalReader.Close()
}
