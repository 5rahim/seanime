package mediastream

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"seanime/internal/database/models"
	"seanime/internal/mediastream/videofile"
	"seanime/internal/testutil"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/samber/mo"
	"github.com/stretchr/testify/require"
)

// newDirectPlayTestRepository returns a repository with a single loaded file, along with the hash of
// another file that is not loaded.
func newDirectPlayTestRepository(t *testing.T) (repo *Repository, loadedHash string, otherHash string) {
	t.Helper()

	dir := t.TempDir()
	loadedPath := filepath.Join(dir, "episode-14.mkv")
	otherPath := filepath.Join(dir, "episode-13.mkv")
	require.NoError(t, os.WriteFile(loadedPath, []byte("episode-14-content"), 0644))
	require.NoError(t, os.WriteFile(otherPath, []byte("episode-13-content"), 0644))

	loadedHash, err := videofile.GetHashFromPath(loadedPath)
	require.NoError(t, err)
	otherHash, err = videofile.GetHashFromPath(otherPath)
	require.NoError(t, err)
	require.NotEqual(t, loadedHash, otherHash)

	repo = &Repository{
		logger:   testutil.NewTestEnv(t).Logger(),
		settings: mo.Some(&models.MediastreamSettings{}),
	}
	repo.playbackManager = NewPlaybackManager(repo)
	repo.playbackManager.currentMediaContainer = mo.Some(&MediaContainer{
		Filepath:   loadedPath,
		Hash:       loadedHash,
		StreamType: StreamTypeDirect,
		StreamUrl:  DirectPlayStreamUrl(loadedHash),
	})

	return repo, loadedHash, otherHash
}

func serveDirectPlay(t *testing.T, repo *Repository, query string) *httptest.ResponseRecorder {
	t.Helper()

	target := "/api/v1/mediastream/direct"
	if query != "" {
		target += "?" + query
	}

	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, target, nil), rec)

	if err := repo.ServeEchoDirectPlay(c, "1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return rec
}

func TestDirectPlayStreamUrlIsUniquePerFile(t *testing.T) {
	// the browser caches by URL, two files must never share one
	require.NotEqual(t, DirectPlayStreamUrl("hash-a"), DirectPlayStreamUrl("hash-b"))
	require.Contains(t, DirectPlayStreamUrl("hash-a"), "hash=hash-a")
}

func TestServeEchoDirectPlayServesLoadedFile(t *testing.T) {
	repo, loadedHash, _ := newDirectPlayTestRepository(t)

	rec := serveDirectPlay(t, repo, "hash="+loadedHash)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "episode-14-content", rec.Body.String())
}

// A request holding the URL of a file that is no longer loaded must fail instead of returning the
// content of the file that is loaded.
func TestServeEchoDirectPlayRejectsHashOfUnloadedFile(t *testing.T) {
	repo, _, otherHash := newDirectPlayTestRepository(t)

	rec := serveDirectPlay(t, repo, "hash="+otherHash)

	require.Equal(t, http.StatusNotFound, rec.Code)
	require.NotContains(t, rec.Body.String(), "episode-14-content")
	require.NotContains(t, rec.Body.String(), "episode-13-content")
}

// Clients that were served a stream URL without a hash keep working.
func TestServeEchoDirectPlayWithoutHashServesLoadedFile(t *testing.T) {
	repo, _, _ := newDirectPlayTestRepository(t)

	rec := serveDirectPlay(t, repo, "")

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "episode-14-content", rec.Body.String())
}
