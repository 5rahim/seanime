package builtin_client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/segments"
	"github.com/anacrolix/torrent/storage"
)

const (
	classicStorageDirPerm  os.FileMode = 0755
	classicStorageFilePerm os.FileMode = 0644
)

type classicFileStorage struct {
	baseDir string
	pc      storage.PieceCompletion
}

type classicTorrent struct {
	hash  metainfo.Hash
	files []classicFile
	index segments.Index
	pc    storage.PieceCompletion
}

type classicFile struct {
	path string
}

type classicPiece struct {
	t *classicTorrent
	p metainfo.Piece
}

func newClassicFileStorage(baseDir string, pc storage.PieceCompletion) storage.ClientImplCloser {
	return &classicFileStorage{baseDir: baseDir, pc: pc}
}

func (s *classicFileStorage) Close() error {
	return nil
}

func (s *classicFileStorage) OpenTorrent(_ context.Context, info *metainfo.Info, hash metainfo.Hash) (storage.TorrentImpl, error) {
	if s.baseDir == "" {
		return storage.TorrentImpl{}, errors.New("base directory is empty")
	}
	if err := os.MkdirAll(s.baseDir, classicStorageDirPerm); err != nil {
		return storage.TorrentImpl{}, err
	}

	metaFiles := info.UpvertedFiles()
	files := make([]classicFile, len(metaFiles))
	for i := range metaFiles {
		path, err := classicTorrentFilePath(s.baseDir, info, &metaFiles[i])
		if err != nil {
			return storage.TorrentImpl{}, fmt.Errorf("file %d: %w", i, err)
		}
		files[i] = classicFile{path: path}
		if metaFiles[i].Length == 0 {
			if err := createClassicStorageFile(path); err != nil {
				return storage.TorrentImpl{}, err
			}
		}
	}

	t := &classicTorrent{
		hash:  hash,
		files: files,
		index: info.FileSegmentsIndex(),
		pc:    s.pc,
	}
	return storage.TorrentImpl{
		Piece: t.Piece,
		Close: t.Close,
	}, nil
}

func classicTorrentFilePath(baseDir string, info *metainfo.Info, file *metainfo.FileInfo) (string, error) {
	parts := make([]string, 0, len(file.BestPath())+1)
	if name := info.BestName(); name != "" && name != metainfo.NoName {
		parts = append(parts, name)
	}
	parts = append(parts, file.BestPath()...)

	rel, err := storage.ToSafeFilePath(parts...)
	if err != nil {
		return "", err
	}
	if rel == "" {
		return "", errors.New("path is empty")
	}

	base, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(filepath.Join(base, rel))
	if err != nil {
		return "", err
	}
	if !isSubPath(base, path) {
		return "", errors.New("path escapes destination")
	}
	return path, nil
}

func isSubPath(base, path string) bool {
	rel, err := filepath.Rel(base, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func createClassicStorageFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), classicStorageDirPerm); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, classicStorageFilePerm)
	if err != nil {
		if stat, statErr := os.Stat(path); statErr == nil && stat.Mode().IsRegular() && stat.Size() == 0 {
			return nil
		}
		return err
	}
	return f.Close()
}

func (t *classicTorrent) Close() error {
	return nil
}

func (t *classicTorrent) Piece(p metainfo.Piece) storage.PieceImpl {
	return &classicPiece{t: t, p: p}
}

func (t *classicTorrent) readAt(b []byte, off int64) (int, error) {
	read := 0
	for i, extent := range t.index.LocateIter(segments.Extent{Start: off, Length: int64(len(b))}) {
		if i < 0 || i >= len(t.files) {
			return read, fmt.Errorf("file index %d is out of range", i)
		}
		n, err := readClassicStorageFile(t.files[i].path, b[:int(extent.Length)], extent.Start)
		read += n
		b = b[n:]
		if err != nil {
			return read, err
		}
		if int64(n) != extent.Length {
			return read, io.EOF
		}
	}
	if len(b) > 0 {
		return read, io.EOF
	}
	return read, nil
}

func (t *classicTorrent) writeAt(b []byte, off int64) (int, error) {
	written := 0
	for i, extent := range t.index.LocateIter(segments.Extent{Start: off, Length: int64(len(b))}) {
		if i < 0 || i >= len(t.files) {
			return written, fmt.Errorf("file index %d is out of range", i)
		}
		n, err := writeClassicStorageFile(t.files[i].path, b[:int(extent.Length)], extent.Start)
		written += n
		b = b[n:]
		if err != nil {
			return written, err
		}
		if int64(n) != extent.Length {
			return written, io.ErrShortWrite
		}
	}
	if len(b) > 0 {
		return written, io.ErrShortWrite
	}
	return written, nil
}

func readClassicStorageFile(path string, b []byte, off int64) (int, error) {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return 0, io.EOF
	}
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return f.ReadAt(b, off)
}

func writeClassicStorageFile(path string, b []byte, off int64) (int, error) {
	if err := os.MkdirAll(filepath.Dir(path), classicStorageDirPerm); err != nil {
		return 0, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, classicStorageFilePerm)
	if err != nil {
		return 0, err
	}
	n, writeErr := f.WriteAt(b, off)
	closeErr := f.Close()
	if writeErr != nil {
		return n, writeErr
	}
	return n, closeErr
}

func (p *classicPiece) ReadAt(b []byte, off int64) (int, error) {
	if off < 0 {
		return 0, os.ErrInvalid
	}
	if off >= p.p.Length() {
		return 0, io.EOF
	}
	requested := len(b)
	if maxLen := p.p.Length() - off; int64(len(b)) > maxLen {
		b = b[:int(maxLen)]
	}
	n, err := p.t.readAt(b, p.p.Offset()+off)
	if err == nil && n < requested {
		err = io.EOF
	}
	return n, err
}

func (p *classicPiece) WriteAt(b []byte, off int64) (int, error) {
	if off < 0 {
		return 0, os.ErrInvalid
	}
	if off+int64(len(b)) > p.p.Length() {
		return 0, io.ErrShortWrite
	}
	return p.t.writeAt(b, p.p.Offset()+off)
}

func (p *classicPiece) MarkComplete() error {
	if err := p.t.pc.Set(p.key(), true); err != nil {
		return err
	}
	return p.sync()
}

func (p *classicPiece) MarkNotComplete() error {
	return p.t.pc.Set(p.key(), false)
}

func (p *classicPiece) Completion() storage.Completion {
	c, err := p.t.pc.Get(p.key())
	c.Err = errors.Join(c.Err, err)
	if c.Err != nil || !c.Ok || !c.Complete {
		return c
	}

	ok, err := p.filesComplete()
	if err != nil {
		c.Err = errors.Join(c.Err, err)
		c.Complete = false
		return c
	}
	if !ok {
		c.Complete = false
		_ = p.MarkNotComplete()
	}
	return c
}

func (p *classicPiece) key() metainfo.PieceKey {
	return metainfo.PieceKey{InfoHash: p.t.hash, Index: p.p.Index()}
}

func (p *classicPiece) filesComplete() (bool, error) {
	for i, extent := range p.t.index.LocateIter(pieceExtent(p.p)) {
		if i < 0 || i >= len(p.t.files) {
			return false, fmt.Errorf("file index %d is out of range", i)
		}
		stat, err := os.Stat(p.t.files[i].path)
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if stat.Size() < extent.Start+extent.Length {
			return false, nil
		}
	}
	return true, nil
}

func (p *classicPiece) sync() error {
	seen := make(map[string]struct{})
	for i := range p.t.index.LocateIter(pieceExtent(p.p)) {
		if i < 0 || i >= len(p.t.files) {
			return fmt.Errorf("file index %d is out of range", i)
		}
		path := p.t.files[i].path
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		f, err := os.OpenFile(path, os.O_WRONLY, classicStorageFilePerm)
		if err != nil {
			return err
		}
		syncErr := f.Sync()
		closeErr := f.Close()
		if syncErr != nil {
			return syncErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func pieceExtent(piece metainfo.Piece) segments.Extent {
	return segments.Extent{Start: piece.Offset(), Length: piece.Length()}
}
