package httputil

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func setTestTempDir(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()
	t.Setenv("TMPDIR", tempDir)
	t.Setenv("TMP", tempDir)
	t.Setenv("TEMP", tempDir)
	return tempDir
}

func TestNewFileStreamCreatesAndCleansUpTempFile(t *testing.T) {
	baseTemp := setTestTempDir(t)
	expectedDir := filepath.Join(baseTemp, "seanime", "filestream")

	_, err := os.Stat(expectedDir)
	require.ErrorIs(t, err, os.ErrNotExist)

	logger := zerolog.Nop()
	const contentLength int64 = 4096

	fs, err := NewFileStream(context.Background(), &logger, contentLength)
	require.NoError(t, err)
	require.NotNil(t, fs)

	filePath := fs.file.Name()
	require.Equal(t, expectedDir, filepath.Dir(filePath))
	require.True(t, strings.HasPrefix(filepath.Base(filePath), "filestream_"))
	require.True(t, strings.HasSuffix(filepath.Base(filePath), ".tmp"))

	info, err := os.Stat(filePath)
	require.NoError(t, err)
	require.Equal(t, contentLength, info.Size())

	require.NoError(t, fs.Close())

	_, err = os.Stat(filePath)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestCleanupStaleFileStreamsMissingDir(t *testing.T) {
	baseTemp := setTestTempDir(t)
	expectedDir := filepath.Join(baseTemp, "seanime", "filestream")
	logger := zerolog.Nop()

	require.NotPanics(t, func() {
		CleanupStaleFileStreams(&logger)
		CleanupStaleFileStreams(nil)
	})

	_, err := os.Stat(expectedDir)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestCleanupStaleFileStreamsRemovesOnlySubdirFiles(t *testing.T) {
	baseTemp := setTestTempDir(t)
	streamDir := filepath.Join(baseTemp, "seanime", "filestream")
	siblingDir := filepath.Join(baseTemp, "seanime", "torrentstream")

	require.NoError(t, os.MkdirAll(streamDir, 0755))
	require.NoError(t, os.MkdirAll(siblingDir, 0755))

	staleFile1 := filepath.Join(streamDir, "filestream_111.tmp")
	staleFile2 := filepath.Join(streamDir, "filestream_222.tmp")
	rootFile := filepath.Join(baseTemp, "filestream_root.tmp")
	siblingFile := filepath.Join(siblingDir, "keep.bin")

	for _, path := range []string{staleFile1, staleFile2, rootFile, siblingFile} {
		require.NoError(t, os.WriteFile(path, []byte("data"), 0644))
	}

	logger := zerolog.Nop()
	CleanupStaleFileStreams(&logger)

	_, err := os.Stat(staleFile1)
	require.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Stat(staleFile2)
	require.ErrorIs(t, err, os.ErrNotExist)

	_, err = os.Stat(rootFile)
	require.NoError(t, err)
	_, err = os.Stat(siblingFile)
	require.NoError(t, err)
}
