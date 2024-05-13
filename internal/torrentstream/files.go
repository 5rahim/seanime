package torrentstream

import (
	"github.com/anacrolix/torrent"
	"io"
)

type SeekableFile interface {
	io.ReadSeeker
	io.Closer
}

type ReadableFile struct {
	*torrent.File
	torrent.Reader
}

func (f *ReadableFile) Seek(offset int64, whence int) (int64, error) {
	return f.Reader.Seek(offset+f.File.Offset(), whence)
}

func NewReadableFile(f *torrent.File) (SeekableFile, error) {
	t := f.Torrent()
	reader := t.NewReader()

	reader.SetReadahead(f.Length() / 100)
	reader.SetResponsive()
	_, err := reader.Seek(f.Offset(), io.SeekStart)

	return &ReadableFile{
		File:   f,
		Reader: reader,
	}, err
}
