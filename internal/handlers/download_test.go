package handlers

import (
	"archive/zip"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"seanime/internal/util"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDownloadTorrentFileUsesContentDispositionFilename(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", `attachment; filename="[SubsPlease] Example - 01.torrent"`)
		_, _ = w.Write([]byte("torrent-data"))
	}))
	defer server.Close()

	dest := t.TempDir()
	require.NoError(t, downloadTorrentFile(server.URL+"/download/12345", dest))

	content, err := os.ReadFile(filepath.Join(dest, "[SubsPlease] Example - 01.torrent"))
	require.NoError(t, err)
	require.Equal(t, "torrent-data", string(content))

	_, err = os.Stat(filepath.Join(dest, "12345"))
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestDownloadTorrentFileFallsBackToURLFilename(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("torrent-data"))
	}))
	defer server.Close()

	dest := t.TempDir()
	require.NoError(t, downloadTorrentFile(server.URL+"/download/%5BGroup%5D%20Example.torrent?token=123", dest))

	content, err := os.ReadFile(filepath.Join(dest, "[Group] Example.torrent"))
	require.NoError(t, err)
	require.Equal(t, "torrent-data", string(content))
}

func TestDownloadTorrentFileAddsTorrentExtensionForTorrentContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-bittorrent")
		_, _ = w.Write([]byte("torrent-data"))
	}))
	defer server.Close()

	dest := t.TempDir()
	require.NoError(t, downloadTorrentFile(server.URL+"/download/12345", dest))

	content, err := os.ReadFile(filepath.Join(dest, "12345.torrent"))
	require.NoError(t, err)
	require.Equal(t, "torrent-data", string(content))
}

func TestValidateMacAppArchive(t *testing.T) {
	t.Run("accepts a normal denshi app archive", func(t *testing.T) {
		archivePath := writeZipArchive(t, []zipArchiveEntry{
			{name: "Seanime Denshi.app/", mode: os.ModeDir | 0755},
			{name: "Seanime Denshi.app/Contents/", mode: os.ModeDir | 0755},
			{name: "Seanime Denshi.app/Contents/Frameworks/", mode: os.ModeDir | 0755},
			{name: "Seanime Denshi.app/Contents/Frameworks/Electron Framework.framework/", mode: os.ModeDir | 0755},
			{name: "Seanime Denshi.app/Contents/Frameworks/Electron Framework.framework/Versions/", mode: os.ModeDir | 0755},
			{name: "Seanime Denshi.app/Contents/Frameworks/Electron Framework.framework/Versions/Current", mode: os.ModeSymlink | 0755, content: "A"},
			{name: "Seanime Denshi.app/Contents/Frameworks/Electron Framework.framework/Electron Framework", mode: os.ModeSymlink | 0755, content: "Versions/Current/Electron Framework"},
			{name: "Seanime Denshi.app/Contents/Info.plist", mode: 0644, content: "plist"},
		})

		require.NoError(t, validateMacAppArchive(archivePath, "Seanime Denshi.app"))
	})

	t.Run("rejects archive traversal entries", func(t *testing.T) {
		archivePath := writeZipArchive(t, []zipArchiveEntry{
			{name: "Seanime Denshi.app/", mode: os.ModeDir | 0755},
			{name: "../escape.txt", mode: 0644, content: "nope"},
		})

		err := validateMacAppArchive(archivePath, "Seanime Denshi.app")
		require.Error(t, err)
		require.ErrorIs(t, err, util.ErrArchivePathTraversal)
	})

	t.Run("rejects symlink targets that escape the app bundle", func(t *testing.T) {
		archivePath := writeZipArchive(t, []zipArchiveEntry{
			{name: "Seanime Denshi.app/", mode: os.ModeDir | 0755},
			{name: "Seanime Denshi.app/Contents/link", mode: os.ModeSymlink | 0755, content: "../../../../escape"},
		})

		err := validateMacAppArchive(archivePath, "Seanime Denshi.app")
		require.Error(t, err)
		require.ErrorIs(t, err, util.ErrArchivePathTraversal)
	})

	t.Run("rejects unsupported special entries", func(t *testing.T) {
		archivePath := writeZipArchive(t, []zipArchiveEntry{
			{name: "Seanime Denshi.app/", mode: os.ModeDir | 0755},
			{name: "Seanime Denshi.app/Contents/special", mode: os.ModeNamedPipe | 0644},
		})

		err := validateMacAppArchive(archivePath, "Seanime Denshi.app")
		require.Error(t, err)
		require.ErrorIs(t, err, util.ErrUnsupportedArchiveEntry)
	})

	t.Run("rejects archives without the app bundle", func(t *testing.T) {
		archivePath := writeZipArchive(t, []zipArchiveEntry{
			{name: "Other.app/", mode: os.ModeDir | 0755},
			{name: "Other.app/Contents/Info.plist", mode: 0644, content: "plist"},
		})

		err := validateMacAppArchive(archivePath, "Seanime Denshi.app")
		require.EqualError(t, err, "app: Seanime Denshi.app not found in archive")
	})
}

func TestExtractMacAppArchive(t *testing.T) {
	archivePath := writeZipArchive(t, []zipArchiveEntry{
		{name: "Seanime Denshi.app/", mode: os.ModeDir | 0755},
		{name: "Seanime Denshi.app/Contents/", mode: os.ModeDir | 0755},
		{name: "Seanime Denshi.app/Contents/Frameworks/", mode: os.ModeDir | 0755},
		{name: "Seanime Denshi.app/Contents/Frameworks/Electron Framework.framework/", mode: os.ModeDir | 0755},
		{name: "Seanime Denshi.app/Contents/Frameworks/Electron Framework.framework/Versions/", mode: os.ModeDir | 0755},
		{name: "Seanime Denshi.app/Contents/Frameworks/Electron Framework.framework/Versions/A/", mode: os.ModeDir | 0755},
		{name: "Seanime Denshi.app/Contents/Frameworks/Electron Framework.framework/Versions/A/Electron Framework", mode: 0755, content: "binary"},
		{name: "Seanime Denshi.app/Contents/Frameworks/Electron Framework.framework/Versions/Current", mode: os.ModeSymlink | 0755, content: "A"},
		{name: "Seanime Denshi.app/Contents/Frameworks/Electron Framework.framework/Electron Framework", mode: os.ModeSymlink | 0755, content: "Versions/Current/Electron Framework"},
		{name: "Seanime Denshi.app/Contents/Info.plist", mode: 0644, content: "plist"},
	})

	dest := filepath.Join(t.TempDir(), "extracted")
	require.NoError(t, os.MkdirAll(dest, 0755))
	require.NoError(t, extractMacAppArchive(archivePath, dest, "Seanime Denshi.app"))

	target, err := os.Readlink(filepath.Join(dest, "Seanime Denshi.app", "Contents", "Frameworks", "Electron Framework.framework", "Versions", "Current"))
	require.NoError(t, err)
	require.Equal(t, "A", target)

	binaryPath := filepath.Join(dest, "Seanime Denshi.app", "Contents", "Frameworks", "Electron Framework.framework", "Versions", "A", "Electron Framework")
	binaryContent, err := os.ReadFile(binaryPath)
	require.NoError(t, err)
	require.Equal(t, "binary", string(binaryContent))
}

func TestInstallMacAppBundle(t *testing.T) {
	t.Run("replaces the existing app and clears the backup", func(t *testing.T) {
		root := t.TempDir()
		applicationsPath := filepath.Join(root, "Applications", "Seanime Denshi.app")
		stagedPath := filepath.Join(root, "staged", "Seanime Denshi.app")

		writeAppMarker(t, applicationsPath, "old")
		writeAppMarker(t, stagedPath, "new")

		err := installMacAppBundle(stagedPath, applicationsPath, func(src string, dst string) error {
			return os.Rename(src, dst)
		})
		require.NoError(t, err)

		marker, err := os.ReadFile(filepath.Join(applicationsPath, "Contents", "marker.txt"))
		require.NoError(t, err)
		require.Equal(t, "new", string(marker))

		backups, err := filepath.Glob(applicationsPath + ".backup-*")
		require.NoError(t, err)
		require.Empty(t, backups)
	})

	t.Run("restores the previous app if the install move fails", func(t *testing.T) {
		root := t.TempDir()
		applicationsPath := filepath.Join(root, "Applications", "Seanime Denshi.app")
		stagedPath := filepath.Join(root, "staged", "Seanime Denshi.app")

		writeAppMarker(t, applicationsPath, "old")
		writeAppMarker(t, stagedPath, "new")

		calls := 0
		err := installMacAppBundle(stagedPath, applicationsPath, func(src string, dst string) error {
			calls++
			switch calls {
			case 1:
				return os.Rename(src, dst)
			case 2:
				return errors.New("install failed")
			case 3:
				return os.Rename(src, dst)
			default:
				return nil
			}
		})

		require.ErrorContains(t, err, "failed to move app to Applications")

		marker, readErr := os.ReadFile(filepath.Join(applicationsPath, "Contents", "marker.txt"))
		require.NoError(t, readErr)
		require.Equal(t, "old", string(marker))

		_, statErr := os.Stat(stagedPath)
		require.NoError(t, statErr)
	})
}

type zipArchiveEntry struct {
	name    string
	mode    os.FileMode
	content string
}

func writeZipArchive(t *testing.T, entries []zipArchiveEntry) string {
	t.Helper()

	archivePath := filepath.Join(t.TempDir(), "update.zip")
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}

	writer := zip.NewWriter(archiveFile)
	for _, entry := range entries {
		header := &zip.FileHeader{Name: entry.name, Method: zip.Store}
		header.SetMode(entry.mode)
		zipEntryWriter, err := writer.CreateHeader(header)
		if err != nil {
			_ = writer.Close()
			_ = archiveFile.Close()
			t.Fatalf("create archive entry %s: %v", entry.name, err)
		}

		if entry.mode.IsRegular() || entry.mode&os.ModeSymlink != 0 {
			if _, err := zipEntryWriter.Write([]byte(entry.content)); err != nil {
				_ = writer.Close()
				_ = archiveFile.Close()
				t.Fatalf("write archive entry %s: %v", entry.name, err)
			}
		}
	}

	require.NoError(t, writer.Close())
	require.NoError(t, archiveFile.Close())

	return archivePath
}

func writeAppMarker(t *testing.T, appPath string, marker string) {
	t.Helper()

	markerPath := filepath.Join(appPath, "Contents", "marker.txt")
	require.NoError(t, os.MkdirAll(filepath.Dir(markerPath), 0755))
	require.NoError(t, os.WriteFile(markerPath, []byte(marker), 0644))
}
