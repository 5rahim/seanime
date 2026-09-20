package handlers

import (
	"net/http"
	"net/http/httptest"
	"seanime/internal/core"
	"seanime/internal/events"
	"seanime/internal/mediastream"
	"seanime/internal/testutil"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

// The direct play endpoint serves whichever file is currently loaded, so a cached response can
// outlive the file it was served for and end up playing the wrong episode.
func TestHandleMediastreamDirectPlayIsNeverCached(t *testing.T) {
	logger := testutil.NewTestEnv(t).Logger()

	h := &Handler{App: &core.App{
		MediastreamRepository: mediastream.NewRepository(&mediastream.NewRepositoryOptions{
			Logger:         logger,
			WSEventManager: events.NewWSEventManager(logger),
		}),
	}}

	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/api/v1/mediastream/direct", nil), rec)

	// the repository is not initialized, the response is an error but the headers must still be set
	_ = h.HandleMediastreamDirectPlay(c)

	require.Contains(t, strings.ToLower(rec.Header().Get("Cache-Control")), "no-store")
	require.Equal(t, "no-cache", rec.Header().Get("Pragma"))
	require.Equal(t, "0", rec.Header().Get("Expires"))
}
