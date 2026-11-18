package realdebrid

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

type capturedRequest struct {
	method      string
	path        string
	auth        string
	contentType string
	form        url.Values
}

func newTestProvider(t *testing.T, status int, body string) (*RealDebrid, *capturedRequest) {
	t.Helper()
	got := &capturedRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(raw))
		*got = capturedRequest{r.Method, r.URL.Path, r.Header.Get("Authorization"), r.Header.Get("Content-Type"), form}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	logger := zerolog.Nop()
	p := NewRealDebridT(&logger)
	p.baseUrl = srv.URL
	p.apiKey = mo.Some("test-token")
	return p, got
}

func TestFormRequests(t *testing.T) {
	magnet := "magnet:?xt=urn:btih:0123456789abcdef0123456789abcdef01234567&dn=Show+Name%20[Group]+100%25+A%26B=C&tr=udp%3A%2F%2Ftracker.example.org%3A1337%2Fannounce"
	link := "https://real-debrid.com/d/ABCDEF?name=a&b=c%20d#frag"

	tests := []struct {
		name   string
		status int
		body   string
		path   string
		field  string
		value  string
		call   func(p *RealDebrid) error
	}{
		{
			name:   "add magnet",
			status: http.StatusCreated,
			body:   `{"id":"ABC123","uri":"https://api.real-debrid.com/rest/1.0/torrents/info/ABC123"}`,
			path:   "/torrents/addMagnet",
			field:  "magnet",
			value:  magnet,
			call: func(p *RealDebrid) error {
				resp, err := p.addMagnet(magnet)
				if err == nil && resp.ID != "ABC123" {
					t.Errorf("id = %q, want ABC123", resp.ID)
				}
				return err
			},
		},
		{
			name:   "select files",
			status: http.StatusNoContent,
			path:   "/torrents/selectFiles/ABC123",
			field:  "files",
			value:  "1,3,12",
			call:   func(p *RealDebrid) error { return p._selectFiles("ABC123", "1,3,12") },
		},
		{
			name:   "unrestrict link",
			status: http.StatusOK,
			body:   `{"id":"XYZ","filename":"ep01.mkv","download":"https://example.com/d/ep01.mkv"}`,
			path:   "/unrestrict/link",
			field:  "link",
			value:  link,
			call: func(p *RealDebrid) error {
				resp, err := p.unrestrictLink(link)
				if err == nil && resp.Download != "https://example.com/d/ep01.mkv" {
					t.Errorf("download = %q", resp.Download)
				}
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, got := newTestProvider(t, tt.status, tt.body)
			if err := tt.call(p); err != nil {
				t.Fatal(err)
			}
			if got.method != http.MethodPost || got.path != tt.path {
				t.Errorf("request = %s %s, want POST %s", got.method, got.path, tt.path)
			}
			if got.auth != "Bearer test-token" {
				t.Errorf("auth = %q", got.auth)
			}
			if got.contentType != "application/x-www-form-urlencoded" {
				t.Errorf("content type = %q", got.contentType)
			}
			if len(got.form) != 1 || got.form.Get(tt.field) != tt.value {
				t.Errorf("form = %v, want only %s=%q", got.form, tt.field, tt.value)
			}
		})

		t.Run(tt.name+" error", func(t *testing.T) {
			p, _ := newTestProvider(t, http.StatusForbidden, `{"error":"bad_token","error_details":"Check parameters and/or cookies sent","error_code":9}`)
			err := tt.call(p)
			if err == nil || !strings.Contains(err.Error(), "403 Forbidden") {
				t.Errorf("err = %v, want 403", err)
			}
		})
	}
}
