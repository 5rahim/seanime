package alldebrid

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllDebrid_Authenticate(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/user", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{
			Status: "success",
			Data: map[string]interface{}{
				"user": map[string]string{"username": "testuser"},
			},
		})
	}))
	defer server.Close()

	logger := zerolog.Nop()
	ad := NewAllDebrid(&logger).(*AllDebrid)
	ad.baseUrl = server.URL // Override base URL for testing

	err := ad.Authenticate("test-api-key")
	assert.NoError(t, err)
}

func TestAllDebrid_GetTorrents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/../v4.1/magnet/status", r.URL.Path)
		assert.Equal(t, "Bearer key", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{
			Status: "success",
			Data: map[string]interface{}{
				"magnets": []map[string]interface{}{
					{
						"id":            123,
						"filename":      "test.mkv",
						"size":          1000,
						"hash":          "hash123",
						"status":        "Ready",
						"statusCode":    4,
						"downloaded":    1000,
						"uploaded":      0,
						"seeders":       10,
						"downloadSpeed": 0,
						"uploadSpeed":   0,
						"links":         []interface{}{},
					},
				},
			},
		})
	}))
	defer server.Close()

	logger := zerolog.Nop()
	ad := NewAllDebrid(&logger).(*AllDebrid)
	ad.baseUrl = server.URL
	ad.apiKey = mo.Some("key")

	torrents, err := ad.GetTorrents()
	assert.NoError(t, err)
	assert.Len(t, torrents, 1)
	assert.Equal(t, "123", torrents[0].ID)
	assert.Equal(t, "test.mkv", torrents[0].Name)
	assert.True(t, torrents[0].IsReady)
}

func TestAllDebrid_Integration(t *testing.T) {
	apiKey := os.Getenv("ALLDEBRID_API_KEY")
	if apiKey == "" {
		t.Skip("ALLDEBRID_API_KEY not set")
	}

	logger := zerolog.Nop()
	ad := NewAllDebrid(&logger).(*AllDebrid)

	// Test Authenticate
	err := ad.Authenticate(apiKey)
	require.NoError(t, err)

	// Test GetTorrents
	torrents, err := ad.GetTorrents()
	require.NoError(t, err)
	t.Logf("Found %d torrents", len(torrents))

	// Optional: Add a magnet and check info
	// magnet := "magnet:?xt=urn:btih:..."
	// id, err := ad.AddTorrent(debrid.AddTorrentOptions{MagnetLink: magnet})
	// ...
}
