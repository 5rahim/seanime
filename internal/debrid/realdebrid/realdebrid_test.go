package realdebrid

import (
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

const testAPIKey = "test-token"

type capturedRequest struct {
	method      string
	path        string
	auth        string
	contentType string
	form        url.Values
}

func newTestProvider(t *testing.T, status int, body string) (*RealDebrid, *capturedRequest) {
	t.Helper()
	captured := &capturedRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(raw))
		*captured = capturedRequest{
			method:      r.Method,
			path:        r.URL.Path,
			auth:        r.Header.Get("Authorization"),
			contentType: r.Header.Get("Content-Type"),
			form:        form,
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	logger := zerolog.Nop()
	p := NewRealDebridT(&logger)
	p.baseUrl = srv.URL
	p.apiKey = mo.Some(testAPIKey)
	return p, captured
}

func assertFormRequest(t *testing.T, got *capturedRequest, path string) {
	t.Helper()
	if got.method != http.MethodPost {
		t.Errorf("method = %q, want POST", got.method)
	}
	if got.path != path {
		t.Errorf("path = %q, want %q", got.path, path)
	}
	if got.auth != "Bearer "+testAPIKey {
		t.Errorf("Authorization = %q, want bearer token", got.auth)
	}
	if got.contentType != "application/x-www-form-urlencoded" {
		t.Errorf("Content-Type = %q, want application/x-www-form-urlencoded", got.contentType)
	}
	if mt, _, _ := mime.ParseMediaType(got.contentType); strings.HasPrefix(mt, "multipart/") {
		t.Errorf("request must not be multipart, got %q", mt)
	}
}

func TestAddMagnet(t *testing.T) {
	magnets := map[string]string{
		"simple": "magnet:?xt=urn:btih:0123456789abcdef0123456789abcdef01234567",
		"special characters": "magnet:?xt=urn:btih:0123456789abcdef0123456789abcdef01234567" +
			"&dn=Show+Name%20[Group]+100%25+A%26B=C&tr=udp%3A%2F%2Ftracker.example.org%3A1337%2Fannounce" +
			"&tr=https://tracker.example.org/announce?k=v&x=1",
	}
	for name, magnet := range magnets {
		t.Run(name, func(t *testing.T) {
			p, got := newTestProvider(t, http.StatusCreated, `{"id":"ABC123","uri":"https://api.real-debrid.com/rest/1.0/torrents/info/ABC123"}`)

			resp, err := p.addMagnet(magnet)
			if err != nil {
				t.Fatalf("addMagnet: %v", err)
			}
			assertFormRequest(t, got, "/torrents/addMagnet")
			if len(got.form) != 1 || got.form.Get("magnet") != magnet {
				t.Errorf("form = %v, want only magnet=%q", got.form, magnet)
			}
			if resp.ID != "ABC123" || !strings.HasSuffix(resp.URI, "/torrents/info/ABC123") {
				t.Errorf("unexpected response: %+v", resp)
			}
		})
	}
}

func TestSelectFiles(t *testing.T) {
	for _, files := range []string{"all", "1,3,12"} {
		t.Run(files, func(t *testing.T) {
			p, got := newTestProvider(t, http.StatusNoContent, "")

			if err := p._selectFiles("ABC123", files); err != nil {
				t.Fatalf("_selectFiles: %v", err)
			}
			assertFormRequest(t, got, "/torrents/selectFiles/ABC123")
			if len(got.form) != 1 || got.form.Get("files") != files {
				t.Errorf("form = %v, want only files=%q", got.form, files)
			}
		})
	}
}

func TestUnrestrictLink(t *testing.T) {
	link := "https://real-debrid.com/d/ABCDEF?name=a&b=c%20d#frag"
	p, got := newTestProvider(t, http.StatusOK, `{"id":"XYZ","filename":"ep01.mkv","filesize":1234,"link":"https://real-debrid.com/d/ABCDEF","download":"https://example.com/d/ep01.mkv","streamable":1}`)

	resp, err := p.unrestrictLink(link)
	if err != nil {
		t.Fatalf("unrestrictLink: %v", err)
	}
	assertFormRequest(t, got, "/unrestrict/link")
	if len(got.form) != 1 || got.form.Get("link") != link {
		t.Errorf("form = %v, want only link=%q", got.form, link)
	}
	if resp.ID != "XYZ" || resp.Filename != "ep01.mkv" || resp.Filesize != 1234 ||
		resp.Download != "https://example.com/d/ep01.mkv" || resp.Streamable != 1 {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestFormRequestsPropagateAPIError(t *testing.T) {
	const apiErr = `{"error":"bad_token","error_details":"Check parameters and/or cookies sent","error_code":9}`

	calls := map[string]func(p *RealDebrid) error{
		"addMagnet":      func(p *RealDebrid) error { _, err := p.addMagnet("magnet:?xt=urn:btih:abc"); return err },
		"_selectFiles":   func(p *RealDebrid) error { return p._selectFiles("ABC123", "all") },
		"unrestrictLink": func(p *RealDebrid) error { _, err := p.unrestrictLink("https://real-debrid.com/d/ABC"); return err },
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			p, _ := newTestProvider(t, http.StatusForbidden, apiErr)

			err := call(p)
			if err == nil {
				t.Fatal("expected an error")
			}
			for _, want := range []string{"403 Forbidden", "Check parameters and/or cookies sent"} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not contain %q", err, want)
				}
			}
		})
	}
}
