package goja_bindings

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
)

// inspired by figma

func TestFetchRedirectOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/redirect":
			http.Redirect(w, r, "/target", http.StatusFound)
		case "/target":
			_, _ = w.Write([]byte("target"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tests := []struct {
		name           string
		redirect       string
		expectedState  goja.PromiseState
		expectedStatus int
		expectedURL    string
		redirected     bool
		errorContains  string
	}{
		{
			name:           "follow",
			redirect:       "follow",
			expectedState:  goja.PromiseStateFulfilled,
			expectedStatus: http.StatusOK,
			expectedURL:    server.URL + "/target",
			redirected:     true,
		},
		{
			name:           "manual",
			redirect:       "manual",
			expectedState:  goja.PromiseStateFulfilled,
			expectedStatus: http.StatusFound,
			expectedURL:    server.URL + "/redirect",
			redirected:     false,
		},
		{
			name:          "error",
			redirect:      "error",
			expectedState: goja.PromiseStateRejected,
			errorContains: "redirect mode error",
		},
		{
			name:          "invalid",
			redirect:      "sideways",
			expectedState: goja.PromiseStateRejected,
			errorContains: "invalid redirect option",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := goja.New()
			fetch := BindFetch("", vm, []string{"*"})
			defer fetch.Close()

			val, err := vm.RunString(fmt.Sprintf(`fetch(%q, { redirect: %q })`, server.URL+"/redirect", tt.redirect))
			require.NoError(t, err)

			promise := requirePromise(t, val)
			waitForPromiseState(t, promise, tt.expectedState)

			if tt.expectedState == goja.PromiseStateRejected {
				require.Contains(t, promise.Result().String(), tt.errorContains)
				return
			}

			result := promise.Result().ToObject(vm)
			require.Equal(t, int64(tt.expectedStatus), result.Get("status").ToInteger())
			require.Equal(t, tt.expectedURL, result.Get("url").String())
			require.Equal(t, tt.redirected, result.Get("redirected").ToBoolean())
		})
	}
}

func TestIsURLAllowed(t *testing.T) {
	vm := goja.New()

	tests := []struct {
		name           string
		allowedDomains []string
		url            string
		expected       bool
	}{
		// Empty allowedDomains, should deny everything
		{
			name:           "empty allowed domains denies all",
			allowedDomains: []string{},
			url:            "https://example.com",
			expected:       false,
		},

		// Wildcard "*", allows everything
		{
			name:           "wildcard allows any URL",
			allowedDomains: []string{"*"},
			url:            "https://example.com/api/data",
			expected:       true,
		},
		{
			name:           "wildcard allows localhost",
			allowedDomains: []string{"*"},
			url:            "http://localhost:3000",
			expected:       true,
		},

		// Exact domain matches
		{
			name:           "exact domain match",
			allowedDomains: []string{"example.com"},
			url:            "https://example.com",
			expected:       true,
		},
		{
			name:           "exact domain with path",
			allowedDomains: []string{"example.com"},
			url:            "https://example.com/api/data",
			expected:       true,
		},
		{
			name:           "exact domain mismatch",
			allowedDomains: []string{"example.com"},
			url:            "https://other.com",
			expected:       false,
		},
		{
			name:           "exact domain with different subdomain",
			allowedDomains: []string{"example.com"},
			url:            "https://api.example.com",
			expected:       false,
		},

		// Subdomain wildcard (*.example.com)
		{
			name:           "subdomain wildcard matches subdomain",
			allowedDomains: []string{"*.example.com"},
			url:            "https://api.example.com",
			expected:       true,
		},
		{
			name:           "subdomain wildcard matches nested subdomain",
			allowedDomains: []string{"*.example.com"},
			url:            "https://api.v2.example.com",
			expected:       true,
		},
		{
			name:           "subdomain wildcard matches base domain",
			allowedDomains: []string{"*.example.com"},
			url:            "https://example.com",
			expected:       true,
		},
		{
			name:           "subdomain wildcard does not match different domain",
			allowedDomains: []string{"*.example.com"},
			url:            "https://example.org",
			expected:       false,
		},
		{
			name:           "subdomain wildcard with path",
			allowedDomains: []string{"*.example.com"},
			url:            "https://api.example.com/data",
			expected:       true,
		},

		// Scheme-specific patterns
		{
			name:           "http scheme matches",
			allowedDomains: []string{"http://example.com"},
			url:            "http://example.com",
			expected:       true,
		},
		{
			name:           "http scheme does not match https",
			allowedDomains: []string{"http://example.com"},
			url:            "https://example.com",
			expected:       false,
		},
		{
			name:           "https scheme matches",
			allowedDomains: []string{"https://example.com"},
			url:            "https://example.com",
			expected:       true,
		},
		{
			name:           "https scheme does not match http",
			allowedDomains: []string{"https://example.com"},
			url:            "http://example.com",
			expected:       false,
		},
		{
			name:           "ws scheme matches",
			allowedDomains: []string{"ws://example.com"},
			url:            "ws://example.com",
			expected:       true,
		},
		{
			name:           "wss scheme matches",
			allowedDomains: []string{"wss://socket.io"},
			url:            "wss://socket.io",
			expected:       true,
		},

		// Path-specific patterns with trailing slash
		{
			name:           "path with trailing slash allows deeper paths",
			allowedDomains: []string{"example.com/api/"},
			url:            "https://example.com/api/users",
			expected:       true,
		},
		{
			name:           "path with trailing slash allows exact match",
			allowedDomains: []string{"example.com/api/"},
			url:            "https://example.com/api/",
			expected:       true,
		},
		{
			name:           "path with trailing slash denies different path",
			allowedDomains: []string{"example.com/api/"},
			url:            "https://example.com/other",
			expected:       false,
		},
		{
			name:           "path with trailing slash denies parent path",
			allowedDomains: []string{"example.com/api/data/"},
			url:            "https://example.com/api",
			expected:       false,
		},

		// Path-specific patterns without trailing slash
		{
			name:           "path without trailing slash exact match",
			allowedDomains: []string{"api.example.com/rest/get"},
			url:            "https://api.example.com/rest/get",
			expected:       true,
		},
		{
			name:           "path without trailing slash denies deeper path",
			allowedDomains: []string{"api.example.com/rest/get"},
			url:            "https://api.example.com/rest/get/exampleresource.json",
			expected:       false,
		},
		{
			name:           "path without trailing slash denies different path",
			allowedDomains: []string{"api.example.com/rest/get"},
			url:            "https://api.example.com/rest/post",
			expected:       false,
		},

		// Localhost patterns
		{
			name:           "localhost without port",
			allowedDomains: []string{"http://localhost"},
			url:            "http://localhost",
			expected:       true,
		},
		{
			name:           "localhost with specific port matches",
			allowedDomains: []string{"http://localhost:3000"},
			url:            "http://localhost:3000",
			expected:       true,
		},
		{
			name:           "localhost with different port matches base",
			allowedDomains: []string{"http://localhost"},
			url:            "http://localhost:8080",
			expected:       true,
		},
		{
			name:           "localhost https",
			allowedDomains: []string{"https://localhost"},
			url:            "https://localhost",
			expected:       true,
		},
		{
			name:           "localhost with path",
			allowedDomains: []string{"http://localhost:3000"},
			url:            "http://localhost:3000/api/test",
			expected:       true,
		},

		// Specific resource URLs
		{
			name:           "specific resource URL with trailing slash",
			allowedDomains: []string{"www.example.com/images/"},
			url:            "https://www.example.com/images/img1.png",
			expected:       true,
		},
		{
			name:           "specific resource URL matches subdirectory",
			allowedDomains: []string{"www.example.com/images/"},
			url:            "https://www.example.com/images/avatars/img2.png",
			expected:       true,
		},
		{
			name:           "specific resource URL denies sibling path",
			allowedDomains: []string{"www.example.com/images/"},
			url:            "https://www.example.com/videos/video.mp4",
			expected:       false,
		},
		{
			name:           "CDN with https scheme",
			allowedDomains: []string{"https://my-app.cdn.com"},
			url:            "https://my-app.cdn.com/assets/style.css",
			expected:       true,
		},
		{
			name:           "S3 bucket path",
			allowedDomains: []string{"http://s3.amazonaws.com/example_bucket/"},
			url:            "http://s3.amazonaws.com/example_bucket/file.json",
			expected:       true,
		},

		// Multiple domains
		{
			name:           "multiple domains, first matches",
			allowedDomains: []string{"example.com", "figma.com"},
			url:            "https://example.com",
			expected:       true,
		},
		{
			name:           "multiple domains, second matches",
			allowedDomains: []string{"example.com", "figma.com"},
			url:            "https://figma.com/api",
			expected:       true,
		},
		{
			name:           "multiple domains, none match",
			allowedDomains: []string{"example.com", "figma.com"},
			url:            "https://other.com",
			expected:       false,
		},

		// Complex real-world scenarios
		{
			name: "complex mix of patterns",
			allowedDomains: []string{
				"figma.com",
				"*.google.com",
				"https://my-app.cdn.com",
				"wss://socket.io",
				"example.com/api/",
				"exact-path.com/content",
			},
			url:      "https://maps.google.com",
			expected: true,
		},
		{
			name: "complex mix, CDN match",
			allowedDomains: []string{
				"figma.com",
				"*.google.com",
				"https://my-app.cdn.com",
				"wss://socket.io",
				"example.com/api/",
				"exact-path.com/content",
			},
			url:      "https://my-app.cdn.com/bundle.js",
			expected: true,
		},
		{
			name: "complex mix, path prefix match",
			allowedDomains: []string{
				"figma.com",
				"*.google.com",
				"https://my-app.cdn.com",
				"wss://socket.io",
				"example.com/api/",
				"exact-path.com/content",
			},
			url:      "https://example.com/api/users/123",
			expected: true,
		},
		{
			name: "complex mix, exact path only",
			allowedDomains: []string{
				"figma.com",
				"*.google.com",
				"https://my-app.cdn.com",
				"wss://socket.io",
				"example.com/api/",
				"exact-path.com/content",
			},
			url:      "https://exact-path.com/content",
			expected: true,
		},
		{
			name: "complex mix, exact path blocks deeper",
			allowedDomains: []string{
				"figma.com",
				"*.google.com",
				"https://my-app.cdn.com",
				"wss://socket.io",
				"example.com/api/",
				"exact-path.com/content",
			},
			url:      "https://exact-path.com/content/deep",
			expected: false,
		},

		// Edge cases
		{
			name:           "invalid URL",
			allowedDomains: []string{"example.com"},
			url:            "not a valid url",
			expected:       false,
		},
		{
			name:           "empty URL",
			allowedDomains: []string{"example.com"},
			url:            "",
			expected:       false,
		},
		{
			name:           "URL with query parameters",
			allowedDomains: []string{"example.com"},
			url:            "https://example.com/api?key=value",
			expected:       true,
		},
		{
			name:           "URL with fragment",
			allowedDomains: []string{"example.com"},
			url:            "https://example.com/page#section",
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFetch("", vm, tt.allowedDomains)
			result := f.isURLAllowed(tt.url)
			if result != tt.expected {
				t.Errorf("isURLAllowed(%q) with domains %v = %v, expected %v",
					tt.url, tt.allowedDomains, result, tt.expected)
			}
		})
	}
}
