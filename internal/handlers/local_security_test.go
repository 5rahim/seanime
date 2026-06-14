package handlers

import (
	"net/http"
	"net/http/httptest"
	"seanime/internal/core"
	"seanime/internal/database/models"
	"seanime/internal/security"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRequestHasTrustedLocalOrigin(t *testing.T) {
	tests := []struct {
		name    string
		origin  string
		referer string
		reqHost string
		want    bool
	}{
		{
			name:    "allows loopback origin",
			origin:  "http://127.0.0.1:43211",
			reqHost: "example.com",
			want:    true,
		},
		{
			name:    "allows localhost dev origin against loopback api",
			origin:  "http://localhost:43210",
			reqHost: "127.0.0.1:43001",
			want:    true,
		},
		{
			name:    "allows denshi app origin",
			origin:  "app://-",
			reqHost: "127.0.0.1:43211",
			want:    true,
		},
		{
			name:    "falls back to referer",
			referer: "http://[::1]:43211/settings",
			reqHost: "example.com",
			want:    true,
		},
		{
			name:    "allows same server lan origin",
			origin:  "http://192.168.1.10:43211",
			reqHost: "192.168.1.10:43211",
			want:    true,
		},
		{
			name:    "rejects arbitrary website origins",
			origin:  "https://evil.example",
			reqHost: "192.168.1.10:43211",
			want:    false,
		},
		{
			name:    "rejects different lan origins",
			origin:  "http://192.168.1.10:43211",
			reqHost: "192.168.1.20:43211",
			want:    false,
		},
		{
			name:    "allows tailscale origin",
			origin:  "http://100.64.1.10:43211",
			reqHost: "100.64.1.10:43211",
			want:    true,
		},
		{
			name:    "allows tailscale IPv6 origin",
			origin:  "http://[fd7a:115c:a1e0::1]:43211",
			reqHost: "[fd7a:115c:a1e0::1]:43211",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// local browser and denshi should still be able to change these settings.
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/settings", nil)
			req.Host = tt.reqHost
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.referer != "" {
				req.Header.Set("Referer", tt.referer)
			}

			assert.Equal(t, tt.want, isRequestFromTrustedOrigin(req))
		})
	}
}

func TestRequestHasStrictTrustedLocalBoundary(t *testing.T) {
	t.Cleanup(func() {
		security.SetRequestBoundaryConfig(nil, "")
	})

	tests := []struct {
		name       string
		origin     string
		reqHost    string
		remoteAddr string
		headers    map[string]string
		want       bool
	}{
		{
			name:       "allows direct loopback browser request",
			origin:     "http://127.0.0.1:43211",
			reqHost:    "127.0.0.1:43211",
			remoteAddr: "127.0.0.1:51111",
			want:       true,
		},
		{
			name:       "allows denshi from loopback",
			origin:     "app://-",
			reqHost:    "127.0.0.1:43211",
			remoteAddr: "127.0.0.1:51111",
			want:       true,
		},
		{
			name:       "rejects public client spoofing loopback headers",
			origin:     "http://127.0.0.1:43211",
			reqHost:    "127.0.0.1:43211",
			remoteAddr: "203.0.113.10:51111",
			want:       false,
		},
		{
			name:       "rejects public host through local proxy",
			origin:     "https://seanime.example",
			reqHost:    "seanime.example",
			remoteAddr: "127.0.0.1:51111",
			want:       false,
		},
		{
			name:       "rejects untrusted forwarded headers",
			origin:     "http://127.0.0.1:43211",
			reqHost:    "127.0.0.1:43211",
			remoteAddr: "127.0.0.1:51111",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.10",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// strict local-only actions should not trust spoofable headers without a local client boundary.
			req := httptest.NewRequest(http.MethodPost, "/api/v1/download-torrent-file", nil)
			req.Host = tt.reqHost
			req.RemoteAddr = tt.remoteAddr
			req.Header.Set("Origin", tt.origin)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			assert.Equal(t, tt.want, isRequestFromTrustedLocal(req))
		})
	}
}

func TestRequestHasTrustedLocalHost(t *testing.T) {
	tests := []struct {
		name    string
		reqHost string
		want    bool
	}{
		{
			name:    "allows localhost host",
			reqHost: "localhost:43211",
			want:    true,
		},
		{
			name:    "allows loopback host",
			reqHost: "127.0.0.1:43211",
			want:    true,
		},
		{
			name:    "allows ipv6 loopback host",
			reqHost: "[::1]:43211",
			want:    true,
		},
		{
			name:    "allows private lan host",
			reqHost: "192.168.1.10:43211",
			want:    true,
		},
		{
			name:    "rejects arbitrary domain host",
			reqHost: "evil.example",
			want:    false,
		},
		{
			name:    "allows tailscale IPv4 host",
			reqHost: "100.64.1.10:43211",
			want:    true,
		},
		{
			name:    "allows tailscale IPv6 host",
			reqHost: "[fd7a:115c:a1e0::1]:43211",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
			req.Host = tt.reqHost

			assert.Equal(t, tt.want, isTrustedRequestHost(req))
		})
	}
}

func TestCanAccessLocalServer(t *testing.T) {
	tests := []struct {
		name            string
		origin          string
		reqHost         string
		remoteAddr      string
		serverPassword  string
		accessAllowlist []string
		secureMode      string
		want            bool
	}{
		{
			name:       "allows passwordless local host without browser metadata",
			reqHost:    "127.0.0.1:43211",
			remoteAddr: "127.0.0.1:51111",
			want:       true,
		},
		{
			name:       "allows passwordless trusted local origin",
			origin:     "http://127.0.0.1:43211",
			reqHost:    "127.0.0.1:43211",
			remoteAddr: "127.0.0.1:51111",
			want:       true,
		},
		{
			name:       "rejects passwordless spoofed local host without local client boundary",
			reqHost:    "127.0.0.1:43211",
			remoteAddr: "203.0.113.10:51111",
			want:       true,
		},
		{
			name:       "rejects passwordless spoofed trusted local origin without local client boundary",
			origin:     "http://127.0.0.1:43211",
			reqHost:    "127.0.0.1:43211",
			remoteAddr: "203.0.113.10:51111",
			want:       true,
		},
		{
			name:       "rejects passwordless spoofed local host in hardened mode",
			reqHost:    "127.0.0.1:43211",
			remoteAddr: "203.0.113.10:51111",
			secureMode: security.SecureModeHardened,
			want:       false,
		},
		{
			name:       "rejects passwordless spoofed trusted local origin in hardened mode",
			origin:     "http://127.0.0.1:43211",
			reqHost:    "127.0.0.1:43211",
			remoteAddr: "203.0.113.10:51111",
			secureMode: security.SecureModeHardened,
			want:       false,
		},
		{
			name:    "rejects passwordless untrusted origin even on local host",
			origin:  "https://evil.example",
			reqHost: "127.0.0.1:43211",
			want:    false,
		},
		{
			name:       "allows passwordless same-server lan host by default",
			reqHost:    "192.168.1.10:43211",
			remoteAddr: "192.168.1.10:51111",
			want:       true,
		},
		{
			name:       "rejects passwordless same-server lan host in hardened mode",
			reqHost:    "192.168.1.10:43211",
			remoteAddr: "192.168.1.10:51111",
			secureMode: security.SecureModeHardened,
			want:       false,
		},
		{
			name:       "rejects passwordless same-server lan host in strict mode",
			reqHost:    "192.168.1.10:43211",
			remoteAddr: "192.168.1.10:51111",
			secureMode: security.SecureModeStrict,
			want:       false,
		},
		{
			name:       "allows passwordless tailscale host by default",
			reqHost:    "100.64.1.10:43211",
			remoteAddr: "100.64.1.10:51111",
			want:       true,
		},
		{
			name:       "rejects passwordless tailscale host in hardened mode",
			reqHost:    "100.64.1.10:43211",
			remoteAddr: "100.64.1.10:51111",
			secureMode: security.SecureModeHardened,
			want:       false,
		},
		{
			name:    "rejects passwordless arbitrary domain host",
			reqHost: "evil.example",
			want:    false,
		},
		{
			name:    "rejects passwordless cross-site browser requests without origin metadata",
			reqHost: "127.0.0.1:43211",
			origin:  "",
			want:    false,
		},
		{
			name:           "allows authenticated requests regardless of host",
			reqHost:        "evil.example",
			serverPassword: "configured",
			want:           true,
		},
		{
			name:           "allows authenticated requests regardless of host in strict mode",
			reqHost:        "evil.example",
			serverPassword: "configured",
			secureMode:     security.SecureModeStrict,
			want:           true,
		},
		{
			name:            "allows passwordless public host when allowlisted",
			reqHost:         "demo.example",
			accessAllowlist: []string{"demo.example"},
			want:            true,
		},
		{
			name:            "allows passwordless public origin when allowlisted",
			origin:          "https://demo.example",
			reqHost:         "demo.example",
			accessAllowlist: []string{"https://demo.example"},
			want:            true,
		},
		{
			name:            "allows passwordless public subdomain when wildcard allowlisted",
			origin:          "https://live.demo.example",
			reqHost:         "live.demo.example",
			accessAllowlist: []string{"*.demo.example"},
			want:            true,
		},
		{
			name:       "allows arbitrary public host in lax mode",
			reqHost:    "demo.example",
			secureMode: security.SecureModeLax,
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			security.SetSecureMode(tt.secureMode)
			t.Cleanup(func() {
				security.SetSecureMode("")
			})

			req := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
			req.Host = tt.reqHost
			if tt.remoteAddr != "" {
				req.RemoteAddr = tt.remoteAddr
			}
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			} else if tt.name == "rejects passwordless cross-site browser requests without origin metadata" {
				req.Header.Set("Sec-Fetch-Site", "cross-site")
			}

			assert.Equal(t, tt.want, isRequestPermitted(req, tt.serverPassword, tt.accessAllowlist))
		})
	}
}

func TestTrustedCORSOrigin(t *testing.T) {
	tests := []struct {
		name            string
		origin          string
		serverPassword  string
		accessAllowlist []string
		secureMode      string
		want            bool
	}{
		{
			name:   "allows trusted local origin",
			origin: "http://127.0.0.1:43211",
			want:   true,
		},
		{
			name:   "allows private lan origin by default",
			origin: "http://192.168.1.10:43211",
			want:   true,
		},
		{
			name:       "rejects private lan origin in hardened mode",
			origin:     "http://192.168.1.10:43211",
			secureMode: security.SecureModeHardened,
			want:       false,
		},
		{
			name:       "rejects private lan origin in strict mode",
			origin:     "http://192.168.1.10:43211",
			secureMode: security.SecureModeStrict,
			want:       false,
		},
		{
			name:   "allows tailscale origin by default",
			origin: "http://100.64.1.10:43211",
			want:   true,
		},
		{
			name:       "rejects tailscale origin in hardened mode",
			origin:     "http://100.64.1.10:43211",
			secureMode: security.SecureModeHardened,
			want:       false,
		},
		{
			name:            "allows allowlisted public origin",
			origin:          "https://demo.example",
			accessAllowlist: []string{"https://demo.example"},
			want:            true,
		},
		{
			name:            "allows wildcard allowlisted public origin",
			origin:          "https://live.demo.example",
			accessAllowlist: []string{"*.demo.example"},
			want:            true,
		},
		{
			name:   "rejects arbitrary public origin without allowlist",
			origin: "https://demo.example",
			want:   false,
		},
		{
			name:       "allows any origin in lax mode",
			origin:     "https://demo.example",
			secureMode: security.SecureModeLax,
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			security.SetSecureMode(tt.secureMode)
			t.Cleanup(func() {
				security.SetSecureMode("")
			})

			assert.Equal(t, tt.want, isTrustedCORSOrigin(tt.origin, tt.serverPassword, tt.accessAllowlist))
		})
	}
}

func TestCanMutatePrivilegedSettings(t *testing.T) {
	prev := &models.Settings{
		MediaPlayer: &models.MediaPlayerSettings{
			Default: "vlc",
			VlcPath: "/Applications/VLC.app/Contents/MacOS/VLC",
			MpvArgs: "--no-config",
		},
		Torrent: &models.TorrentSettings{
			Default:         "qbittorrent",
			QBittorrentPath: "/Applications/qBittorrent.app/Contents/MacOS/qbittorrent",
		},
	}

	tests := []struct {
		name           string
		origin         string
		reqHost        string
		remoteAddr     string
		secureMode     string
		serverPassword string
		nextMedia      *models.MediaPlayerSettings
		nextTorrent    *models.TorrentSettings
		want           bool
	}{
		{
			name:    "allows unrelated settings changes without trusted origin",
			origin:  "https://evil.example",
			reqHost: "192.168.1.10:43211",
			nextMedia: &models.MediaPlayerSettings{
				Default: "vlc",
				Host:    "127.0.0.1",
				VlcPath: "/Applications/VLC.app/Contents/MacOS/VLC",
				MpvArgs: "--no-config",
			},
			nextTorrent: &models.TorrentSettings{
				Default:         "qbittorrent",
				QBittorrentPath: "/Applications/qBittorrent.app/Contents/MacOS/qbittorrent",
			},
			want: true,
		},
		{
			name:       "allows trusted local origin when passwordless",
			origin:     "http://127.0.0.1:43211",
			reqHost:    "127.0.0.1:43211",
			remoteAddr: "127.0.0.1:51111",
			nextMedia: &models.MediaPlayerSettings{
				Default: "mpv",
				VlcPath: "/Applications/VLC.app/Contents/MacOS/VLC",
				MpvPath: "/Applications/mpv.app/Contents/MacOS/mpv",
				MpvArgs: "--no-config",
			},
			nextTorrent: &models.TorrentSettings{
				Default:         "qbittorrent",
				QBittorrentPath: "/Applications/qBittorrent.app/Contents/MacOS/qbittorrent",
			},
			want: true,
		},
		{
			name:       "allows spoofed trusted local origin by default",
			origin:     "http://127.0.0.1:43211",
			reqHost:    "127.0.0.1:43211",
			remoteAddr: "203.0.113.10:51111",
			nextMedia: &models.MediaPlayerSettings{
				Default: "mpv",
				VlcPath: "/Applications/VLC.app/Contents/MacOS/VLC",
				MpvPath: "/tmp/mpv-wrapper",
				MpvArgs: "--no-config",
			},
			nextTorrent: &models.TorrentSettings{
				Default:         "qbittorrent",
				QBittorrentPath: "/Applications/qBittorrent.app/Contents/MacOS/qbittorrent",
			},
			want: true,
		},
		{
			name:       "rejects spoofed trusted local origin in hardened mode",
			origin:     "http://127.0.0.1:43211",
			reqHost:    "127.0.0.1:43211",
			remoteAddr: "203.0.113.10:51111",
			secureMode: security.SecureModeHardened,
			nextMedia: &models.MediaPlayerSettings{
				Default: "mpv",
				VlcPath: "/Applications/VLC.app/Contents/MacOS/VLC",
				MpvPath: "/tmp/mpv-wrapper",
				MpvArgs: "--no-config",
			},
			nextTorrent: &models.TorrentSettings{
				Default:         "qbittorrent",
				QBittorrentPath: "/Applications/qBittorrent.app/Contents/MacOS/qbittorrent",
			},
			want: false,
		},
		{
			name:    "rejects untrusted origin when command sinks change",
			origin:  "https://evil.example",
			reqHost: "192.168.1.10:43211",
			nextMedia: &models.MediaPlayerSettings{
				Default: "mpv",
				VlcPath: "/Applications/VLC.app/Contents/MacOS/VLC",
				MpvPath: "/tmp/mpv-wrapper",
				MpvArgs: "--no-config",
			},
			nextTorrent: &models.TorrentSettings{
				Default:         "qbittorrent",
				QBittorrentPath: "/Applications/qBittorrent.app/Contents/MacOS/qbittorrent",
			},
			want: false,
		},
		{
			name:    "rejects untrusted origin when compatible translation endpoint changes",
			origin:  "https://evil.example",
			reqHost: "192.168.1.10:43211",
			nextMedia: &models.MediaPlayerSettings{
				Default:             "vlc",
				VlcPath:             "/Applications/VLC.app/Contents/MacOS/VLC",
				MpvArgs:             "--no-config",
				VcTranslate:         true,
				VcTranslateProvider: "openai-compatible",
				VcTranslateBaseUrl:  "http://localhost:1234/v1",
			},
			nextTorrent: &models.TorrentSettings{
				Default:         "qbittorrent",
				QBittorrentPath: "/Applications/qBittorrent.app/Contents/MacOS/qbittorrent",
			},
			want: false,
		},
		{
			name:           "allows authenticated writes even without trusted origin",
			origin:         "https://evil.example",
			reqHost:        "192.168.1.10:43211",
			serverPassword: "configured",
			nextMedia: &models.MediaPlayerSettings{
				Default: "mpv",
				VlcPath: "/Applications/VLC.app/Contents/MacOS/VLC",
				MpvPath: "/tmp/mpv-wrapper",
				MpvArgs: "--no-config",
			},
			nextTorrent: &models.TorrentSettings{
				Default:         "qbittorrent",
				QBittorrentPath: "/Applications/qBittorrent.app/Contents/MacOS/qbittorrent",
			},
			want: true,
		},
		{
			name:    "rejects missing origin when command sinks change and no password is set",
			reqHost: "192.168.1.10:43211",
			nextMedia: &models.MediaPlayerSettings{
				Default: "mpv",
				VlcPath: "/Applications/VLC.app/Contents/MacOS/VLC",
				MpvPath: "/tmp/mpv-wrapper",
				MpvArgs: "--no-config",
			},
			nextTorrent: &models.TorrentSettings{
				Default:         "qbittorrent",
				QBittorrentPath: "/Applications/qBittorrent.app/Contents/MacOS/qbittorrent",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// random websites should not be able to flip command sinks on passwordless servers.
			security.SetSecureMode(tt.secureMode)
			t.Cleanup(func() {
				security.SetSecureMode("")
			})

			req := httptest.NewRequest(http.MethodPatch, "/api/v1/settings", nil)
			req.Host = tt.reqHost
			if tt.remoteAddr != "" {
				req.RemoteAddr = tt.remoteAddr
			}
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			assert.Equal(t, tt.want, canMutatePrivilegedSettings(req, tt.serverPassword, prev, tt.nextMedia, tt.nextTorrent))
		})
	}

	t.Run("rejects authenticated public clients in strict mode", func(t *testing.T) {
		security.SetSecureMode(security.SecureModeStrict)
		t.Cleanup(func() {
			security.SetSecureMode("")
		})

		req := httptest.NewRequest(http.MethodPatch, "/api/v1/settings", nil)
		req.Host = "127.0.0.1:43211"
		req.RemoteAddr = "203.0.113.10:51111"
		req.Header.Set("Origin", "http://127.0.0.1:43211")

		nextMedia := &models.MediaPlayerSettings{
			Default: "mpv",
			MpvPath: "/tmp/mpv-wrapper",
			MpvArgs: "--no-config",
		}

		assert.False(t, canMutatePrivilegedSettings(req, "configured", prev, nextMedia, nil))
	})
}

func TestCanMutatePrivilegedMediastreamSettings(t *testing.T) {
	prev := &models.MediastreamSettings{
		FfmpegPath:  "ffmpeg",
		FfprobePath: "ffprobe",
	}

	tests := []struct {
		name           string
		origin         string
		reqHost        string
		remoteAddr     string
		secureMode     string
		serverPassword string
		next           *models.MediastreamSettings
		want           bool
	}{
		{
			name:    "allows unrelated mediastream changes without trusted origin",
			origin:  "https://evil.example",
			reqHost: "192.168.1.10:43211",
			next: &models.MediastreamSettings{
				TranscodeEnabled: true,
				FfmpegPath:       "ffmpeg",
				FfprobePath:      "ffprobe",
			},
			want: true,
		},
		{
			name:       "allows trusted local origin when ffmpeg path changes",
			origin:     "http://127.0.0.1:43211",
			reqHost:    "127.0.0.1:43211",
			remoteAddr: "127.0.0.1:51111",
			next: &models.MediastreamSettings{
				FfmpegPath:  "/tmp/ffmpeg-wrapper",
				FfprobePath: "ffprobe",
			},
			want: true,
		},
		{
			name:       "allows spoofed trusted local origin when ffmpeg path changes by default",
			origin:     "http://127.0.0.1:43211",
			reqHost:    "127.0.0.1:43211",
			remoteAddr: "203.0.113.10:51111",
			next: &models.MediastreamSettings{
				FfmpegPath:  "/tmp/ffmpeg-wrapper",
				FfprobePath: "ffprobe",
			},
			want: true,
		},
		{
			name:       "rejects spoofed trusted local origin when ffmpeg path changes in hardened mode",
			origin:     "http://127.0.0.1:43211",
			reqHost:    "127.0.0.1:43211",
			remoteAddr: "203.0.113.10:51111",
			secureMode: security.SecureModeHardened,
			next: &models.MediastreamSettings{
				FfmpegPath:  "/tmp/ffmpeg-wrapper",
				FfprobePath: "ffprobe",
			},
			want: false,
		},
		{
			name:    "rejects untrusted origin when ffprobe path changes",
			origin:  "https://evil.example",
			reqHost: "192.168.1.10:43211",
			next: &models.MediastreamSettings{
				FfmpegPath:  "ffmpeg",
				FfprobePath: "/tmp/ffprobe-wrapper",
			},
			want: false,
		},
		{
			name:           "allows authenticated mediastream writes without trusted origin",
			origin:         "https://evil.example",
			reqHost:        "192.168.1.10:43211",
			serverPassword: "configured",
			next: &models.MediastreamSettings{
				FfmpegPath:  "/tmp/ffmpeg-wrapper",
				FfprobePath: "/tmp/ffprobe-wrapper",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// remote sites should not be able to repoint ffmpeg or ffprobe on passwordless servers.
			security.SetSecureMode(tt.secureMode)
			t.Cleanup(func() {
				security.SetSecureMode("")
			})

			req := httptest.NewRequest(http.MethodPatch, "/api/v1/mediastream/settings", nil)
			req.Host = tt.reqHost
			if tt.remoteAddr != "" {
				req.RemoteAddr = tt.remoteAddr
			}
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			assert.Equal(t, tt.want, canMutatePrivilegedMediastreamSettings(req, tt.serverPassword, prev, tt.next))
		})
	}

	t.Run("rejects authenticated public clients in strict mode", func(t *testing.T) {
		security.SetSecureMode(security.SecureModeStrict)
		t.Cleanup(func() {
			security.SetSecureMode("")
		})

		req := httptest.NewRequest(http.MethodPatch, "/api/v1/mediastream/settings", nil)
		req.Host = "127.0.0.1:43211"
		req.RemoteAddr = "203.0.113.10:51111"
		req.Header.Set("Origin", "http://127.0.0.1:43211")

		next := &models.MediastreamSettings{
			FfmpegPath:  "/tmp/ffmpeg-wrapper",
			FfprobePath: "ffprobe",
		}

		assert.False(t, canMutatePrivilegedMediastreamSettings(req, "configured", prev, next))
	})
}

func TestCanUsePrivilegedExtensionManagement(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		reqHost        string
		path           string
		remoteAddr     string
		secureMode     string
		serverPassword string
		want           bool
	}{
		{
			name:       "allows trusted local origin when passwordless",
			origin:     "http://127.0.0.1:43211",
			reqHost:    "127.0.0.1:43211",
			path:       "/api/v1/extensions/external/install",
			remoteAddr: "127.0.0.1:51111",
			want:       true,
		},
		{
			name:    "rejects untrusted origin when passwordless",
			origin:  "https://evil.example",
			reqHost: "192.168.1.10:43211",
			path:    "/api/v1/extensions/external/install",
			want:    false,
		},
		{
			name:       "allows spoofed trusted local origin when passwordless by default",
			origin:     "http://127.0.0.1:43211",
			reqHost:    "127.0.0.1:43211",
			path:       "/api/v1/extensions/external/install",
			remoteAddr: "203.0.113.10:51111",
			want:       true,
		},
		{
			name:       "rejects spoofed trusted local origin when passwordless in hardened mode",
			origin:     "http://127.0.0.1:43211",
			reqHost:    "127.0.0.1:43211",
			path:       "/api/v1/extensions/external/install",
			remoteAddr: "203.0.113.10:51111",
			secureMode: security.SecureModeHardened,
			want:       false,
		},
		{
			name:           "allows authenticated extension management without trusted origin",
			origin:         "https://evil.example",
			reqHost:        "192.168.1.10:43211",
			path:           "/api/v1/extensions/external/install",
			serverPassword: "configured",
			want:           true,
		},
		{
			name:    "rejects untrusted origin for playground execution when passwordless",
			origin:  "https://evil.example",
			reqHost: "192.168.1.10:43211",
			path:    "/api/v1/extensions/playground/run",
			want:    false,
		},
		{
			name:       "allows trusted local origin for playground execution when passwordless",
			origin:     "http://127.0.0.1:43211",
			reqHost:    "127.0.0.1:43211",
			path:       "/api/v1/extensions/playground/run",
			remoteAddr: "127.0.0.1:51111",
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			security.SetSecureMode(tt.secureMode)
			t.Cleanup(func() {
				security.SetSecureMode("")
			})

			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			req.Host = tt.reqHost
			if tt.remoteAddr != "" {
				req.RemoteAddr = tt.remoteAddr
			}
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			assert.Equal(t, tt.want, canUsePrivilegedExtensionManagement(req, tt.serverPassword))
		})
	}

	t.Run("rejects authenticated public clients in strict mode", func(t *testing.T) {
		security.SetSecureMode(security.SecureModeStrict)
		t.Cleanup(func() {
			security.SetSecureMode("")
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/extensions/external/install", nil)
		req.Host = "127.0.0.1:43211"
		req.RemoteAddr = "203.0.113.10:51111"
		req.Header.Set("Origin", "http://127.0.0.1:43211")

		assert.False(t, canUsePrivilegedExtensionManagement(req, "configured"))
	})
}

func TestCanConsumeMedia(t *testing.T) {
	tests := []struct {
		name            string
		origin          string
		reqHost         string
		remoteAddr      string
		serverPassword  string
		accessAllowlist []string
		secureMode      string
		want            bool
	}{
		{
			name:           "allows authenticated public playback in strict mode",
			origin:         "https://demo.example",
			reqHost:        "demo.example",
			remoteAddr:     "203.0.113.10:51111",
			serverPassword: "configured",
			secureMode:     security.SecureModeStrict,
			want:           true,
		},
		{
			name:            "allows allowlisted public playback in strict mode",
			origin:          "https://demo.example",
			reqHost:         "demo.example",
			remoteAddr:      "203.0.113.10:51111",
			accessAllowlist: []string{"https://demo.example"},
			secureMode:      security.SecureModeStrict,
			want:            true,
		},
		{
			name:       "rejects public playback without auth or allowlist in strict mode",
			origin:     "https://demo.example",
			reqHost:    "demo.example",
			remoteAddr: "203.0.113.10:51111",
			secureMode: security.SecureModeStrict,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			security.SetSecureMode(tt.secureMode)
			t.Cleanup(func() {
				security.SetSecureMode("")
			})

			req := httptest.NewRequest(http.MethodPost, "/api/v1/mediastream/request", nil)
			req.Host = tt.reqHost
			if tt.remoteAddr != "" {
				req.RemoteAddr = tt.remoteAddr
			}
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			assert.Equal(t, tt.want, canConsumeMedia(req, tt.serverPassword, tt.accessAllowlist))
		})
	}
}

func TestGuardPrivilegedMediaPlayer(t *testing.T) {
	t.Cleanup(func() {
		security.SetSecureMode("")
	})

	e := echo.New()
	h := &Handler{App: &core.App{Config: &core.Config{}}}

	t.Run("allows authenticated public playback in strict mode", func(t *testing.T) {
		// hosted playback should still work once the normal request boundary is satisfied.
		security.SetSecureMode(security.SecureModeStrict)
		h.App.Config.Server.Password = "configured"

		req := httptest.NewRequest(http.MethodPost, "/api/v1/playback-manager/play", nil)
		req.Host = "demo.example"
		req.RemoteAddr = "203.0.113.10:51111"
		req.Header.Set("Origin", "https://demo.example")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		settings := &models.Settings{
			MediaPlayer: &models.MediaPlayerSettings{
				Default: "mpv",
				MpvPath: "/tmp/mpv-wrapper",
				MpvArgs: "--no-config",
			},
		}

		err := h.guardPrivilegedMediaPlayer(c, settings)
		assert.NoError(t, err)
	})

	t.Run("rejects passwordless public playback with privileged media-player settings", func(t *testing.T) {
		// custom executables should still stay behind the normal privileged request check.
		security.SetSecureMode(security.SecureModeStrict)
		h.App.Config.Server.Password = ""

		req := httptest.NewRequest(http.MethodPost, "/api/v1/playback-manager/play", nil)
		req.Host = "demo.example"
		req.RemoteAddr = "203.0.113.10:51111"
		req.Header.Set("Origin", "https://demo.example")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		settings := &models.Settings{
			MediaPlayer: &models.MediaPlayerSettings{
				Default: "mpv",
				MpvPath: "/tmp/mpv-wrapper",
				MpvArgs: "--no-config",
			},
		}

		err := h.guardPrivilegedMediaPlayer(c, settings)
		assert.ErrorIs(t, err, errGuardResponseWritten)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}

func TestMediaConsumptionHandlersDoNotUseStrictLocalOnlyBoundary(t *testing.T) {
	t.Cleanup(func() {
		security.SetSecureMode("")
	})

	security.SetSecureMode(security.SecureModeStrict)
	e := echo.New()
	h := &Handler{App: &core.App{Config: &core.Config{}}}
	h.App.Config.Server.Password = "configured"

	t.Run("directstream play local file falls through to binding for authenticated hosted requests", func(t *testing.T) {
		// this should no longer short-circuit on the old strict local-only guard.
		req := httptest.NewRequest(http.MethodPost, "/api/v1/directstream/play/localfile", strings.NewReader(`{"path":`))
		req.Host = "demo.example"
		req.RemoteAddr = "203.0.113.10:51111"
		req.Header.Set("Origin", "https://demo.example")
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.HandleDirectstreamPlayLocalFile(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("mediastream request falls through to binding for authenticated hosted requests", func(t *testing.T) {
		// hosted playback requests should hit normal request validation instead of a strict local-only block.
		req := httptest.NewRequest(http.MethodPost, "/api/v1/mediastream/request", strings.NewReader(`{"path":`))
		req.Host = "demo.example"
		req.RemoteAddr = "203.0.113.10:51111"
		req.Header.Set("Origin", "https://demo.example")
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.HandleRequestMediastreamMediaContainer(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestShouldRestrictSensitiveLocalInfo(t *testing.T) {
	t.Cleanup(func() {
		security.SetSecureMode("")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
	req.Host = "192.168.1.10:43211"
	req.RemoteAddr = "203.0.113.10:51111"
	req.Header.Set("Origin", "https://evil.example")

	security.SetSecureMode("")
	assert.False(t, isStrictModeSensitive(req, ""))

	security.SetSecureMode(security.SecureModeStrict)
	assert.True(t, isStrictModeSensitive(req, ""))
	assert.False(t, isStrictModeSensitive(req, "configured"))

	security.SetSecureMode(security.SecureModeLax)
	assert.False(t, isStrictModeSensitive(req, ""))

	req.Host = "127.0.0.1:43211"
	req.RemoteAddr = "127.0.0.1:51111"
	req.Header.Set("Origin", "http://127.0.0.1:43211")
	assert.False(t, isStrictModeSensitive(req, ""))
}

func TestGuardStrictLocalOnlyAction(t *testing.T) {
	t.Cleanup(func() {
		security.SetSecureMode("")
	})

	h := &Handler{App: &core.App{Config: &core.Config{}}}
	e := echo.New()

	t.Run("rejects requests without a trusted local origin in strict mode", func(t *testing.T) {
		security.SetSecureMode(security.SecureModeStrict)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/download-torrent-file", nil)
		req.Host = "127.0.0.1:43211"
		req.RemoteAddr = "203.0.113.10:51111"
		req.Header.Set("Origin", "http://127.0.0.1:43211")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.guardStrictLocalOnlyAction(c)
		assert.ErrorIs(t, err, errGuardResponseWritten)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("allows trusted local origins in strict mode", func(t *testing.T) {
		security.SetSecureMode(security.SecureModeStrict)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/download-torrent-file", nil)
		req.Host = "127.0.0.1:43211"
		req.RemoteAddr = "127.0.0.1:51111"
		req.Header.Set("Origin", "http://127.0.0.1:43211")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.guardStrictLocalOnlyAction(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestHandleDirectorySelectorStrictMode(t *testing.T) {
	t.Cleanup(func() {
		security.SetSecureMode("")
	})

	security.SetSecureMode(security.SecureModeStrict)
	h := &Handler{App: &core.App{Config: &core.Config{}}}
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/directory-selector", strings.NewReader(`{"input":"/tmp"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.HandleDirectorySelector(c)
	assert.ErrorIs(t, err, errGuardResponseWritten)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestHandleTestDumpStrictMode(t *testing.T) {
	t.Cleanup(func() {
		security.SetSecureMode("")
	})

	security.SetSecureMode(security.SecureModeStrict)
	h := &Handler{App: &core.App{Config: &core.Config{}}}
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/test-dump", strings.NewReader(`{"dir":"/tmp","userName":"test"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.HandleTestDump(c)
	assert.ErrorIs(t, err, errGuardResponseWritten)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequestMatchescontextClientId(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	req.AddCookie(&http.Cookie{Name: clientIdCookieName, Value: "client-1"})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(clientIdCookieName, "client-1")

	assert.True(t, isSameContextClientId(c, "client-1"))
	assert.False(t, isSameContextClientId(c, "client-2"))
	assert.Equal(t, "client-1", getContextClientId(c))
	assert.False(t, isSameContextClientId(c, ""))
}

func TestResolveRequestClientId(t *testing.T) {
	t.Run("prefers server context", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/mediastream/request", nil)
		req.AddCookie(&http.Cookie{Name: clientIdCookieName, Value: "cookie-client"})
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(clientIdCookieName, "server-client")

		assert.Equal(t, "server-client", getRequestClientId(c, "body-client"))
	})

	t.Run("falls back to claimed id when context is missing", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/mediastream/request", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		assert.Equal(t, "body-client", getRequestClientId(c, " body-client "))
	})
}

func TestUsesPrivilegedCommandSettings(t *testing.T) {
	t.Run("ignores default executable paths", func(t *testing.T) {
		// default app paths should still behave like the built-in integration.
		settings := &models.Settings{
			MediaPlayer: &models.MediaPlayerSettings{
				Default: "vlc",
				VlcPath: "/Applications/VLC.app/Contents/MacOS/VLC",
			},
			Torrent: &models.TorrentSettings{
				Default:         "qbittorrent",
				QBittorrentPath: "/Applications/qbittorrent.app/Contents/MacOS/qbittorrent",
			},
		}
		mediastreamSettings := &models.MediastreamSettings{
			FfmpegPath:  "ffmpeg",
			FfprobePath: "ffprobe",
		}

		assert.False(t, isPrivilegedMediaPlayer(settings))
		assert.False(t, isPrivilegedTorrentClient(settings))
		assert.False(t, isPrivilegedMediastream(mediastreamSettings))
	})

	t.Run("detects custom paths and custom args", func(t *testing.T) {
		// custom wrappers and launch args stay behind the trusted-origin gate.
		settings := &models.Settings{
			MediaPlayer: &models.MediaPlayerSettings{
				Default: "mpv",
				MpvPath: "/tmp/mpv-wrapper",
				MpvArgs: "--script=/tmp/hook.lua",
			},
			Torrent: &models.TorrentSettings{
				Default:         "qbittorrent",
				QBittorrentPath: "/tmp/qbit-wrapper",
			},
		}
		mediastreamSettings := &models.MediastreamSettings{
			FfmpegPath:  "/tmp/ffmpeg-wrapper",
			FfprobePath: "/tmp/ffprobe-wrapper",
		}

		assert.True(t, isPrivilegedMediaPlayer(settings))
		assert.True(t, isPrivilegedTorrentClient(settings))
		assert.True(t, isPrivilegedMediastream(mediastreamSettings))
	})
}
