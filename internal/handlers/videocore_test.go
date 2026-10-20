package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"seanime/internal/core"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/security"
	"seanime/internal/testutil"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleVideoCoreSaveScreenshotRejectsPathFilename(t *testing.T) {
	e := echo.New()
	body := `{"dir":"/tmp","filename":"../escape.png","base64Data":"aGk="}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/videocore/screenshot", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &Handler{}
	err := h.HandleVideoCoreSaveScreenshot(c)
	assert.NoError(t, err)
	assert.Contains(t, rec.Body.String(), "filename must not contain a path")
}

func TestHandleVideoCoreSaveScreenshot(t *testing.T) {
	prev := db.CurrSettings
	security.SetSecureMode("")
	t.Cleanup(func() {
		db.CurrSettings = prev
		security.SetSecureMode("")
	})
	env := testutil.NewTestEnv(t)
	database := env.NewDatabase("")
	e := echo.New()

	tests := []struct {
		name     string
		origin   string
		password string
		otherDir bool
		noDir    bool
		link     bool
		want     bool
	}{
		{name: "allows same server LAN browser", origin: "http://192.168.1.10:43211", want: true},
		{name: "allows localhost", origin: "http://localhost:43211", want: true},
		{name: "allows authenticated request", password: "configured", want: true},
		{name: "rejects missing origin"},
		{name: "rejects foreign origin", origin: "https://evil.example"},
		{name: "rejects another folder", origin: "http://192.168.1.10:43211", otherDir: true},
		{name: "rejects unset folder", origin: "http://192.168.1.10:43211", noDir: true},
		{name: "rejects symlink outside folder", origin: "http://192.168.1.10:43211", link: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "screenshots")
			settings := &models.Settings{MediaPlayer: &models.MediaPlayerSettings{ScreenshotDir: dir}}
			if tt.noDir {
				settings.MediaPlayer.ScreenshotDir = ""
			}
			_, err := database.UpsertSettings(settings)
			require.NoError(t, err)
			if tt.otherDir {
				dir = t.TempDir()
			}
			var target string
			if tt.link {
				if runtime.GOOS == "windows" {
					t.Skip("symlinks need extra permissions on Windows")
				}
				target = filepath.Join(t.TempDir(), "target")
				require.NoError(t, os.WriteFile(target, []byte("original"), 0644))
				require.NoError(t, os.MkdirAll(dir, 0755))
				require.NoError(t, os.Symlink(target, filepath.Join(dir, "image.png")))
			}
			body, err := json.Marshal(map[string]string{"dir": dir, "filename": "image.png", "base64Data": "aGk="})
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "http://192.168.1.10:43211/api/v1/videocore/screenshot", strings.NewReader(string(body)))
			req.RemoteAddr = "192.168.1.20:51111"
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Origin", tt.origin)
			rec := httptest.NewRecorder()
			h := &Handler{App: &core.App{Config: &core.Config{}, Database: database}}
			h.App.Config.Server.Password = tt.password
			err = h.HandleVideoCoreSaveScreenshot(e.NewContext(req, rec))
			if tt.want {
				require.NoError(t, err)
				assert.JSONEq(t, `{"data":true}`, rec.Body.String())
				data, err := os.ReadFile(filepath.Join(dir, "image.png"))
				require.NoError(t, err)
				assert.Equal(t, "hi", string(data))
			} else {
				assert.Contains(t, rec.Body.String(), `"error"`)
				if tt.link {
					data, err := os.ReadFile(target)
					require.NoError(t, err)
					assert.Equal(t, "original", string(data))
				} else {
					_, err := os.Stat(filepath.Join(dir, "image.png"))
					assert.True(t, os.IsNotExist(err))
				}
			}
		})
	}
}
