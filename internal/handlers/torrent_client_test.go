package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goccy/go-json"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"seanime/internal/api/metadata_provider"
	"seanime/internal/core"
	"seanime/internal/database/models"
	"seanime/internal/extension"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/library/autodownloader"
	"seanime/internal/torrent_clients/qbittorrent"
	"seanime/internal/torrent_clients/torrent_client"
	torrentrepo "seanime/internal/torrents/torrent"
	"seanime/internal/util"
)

type ruleMagnetTestProvider struct {
	magnet string
	err    error
	calls  int
}

func (p *ruleMagnetTestProvider) Search(hibiketorrent.AnimeSearchOptions) ([]*hibiketorrent.AnimeTorrent, error) {
	return nil, nil
}

func (p *ruleMagnetTestProvider) SmartSearch(hibiketorrent.AnimeSmartSearchOptions) ([]*hibiketorrent.AnimeTorrent, error) {
	return nil, nil
}

func (p *ruleMagnetTestProvider) GetTorrentInfoHash(torrent *hibiketorrent.AnimeTorrent) (string, error) {
	if torrent == nil {
		return "", nil
	}
	return torrent.InfoHash, nil
}

func (p *ruleMagnetTestProvider) GetTorrentMagnetLink(*hibiketorrent.AnimeTorrent) (string, error) {
	p.calls++
	if p.err != nil {
		return "", p.err
	}
	return p.magnet, nil
}

func (p *ruleMagnetTestProvider) GetLatest() ([]*hibiketorrent.AnimeTorrent, error) {
	return nil, nil
}

func (p *ruleMagnetTestProvider) GetSettings() hibiketorrent.AnimeProviderSettings {
	return hibiketorrent.AnimeProviderSettings{Type: hibiketorrent.AnimeProviderTypeMain}
}

func TestResolveAutoDownloaderItemMagnetUsesStoredTorrentExtension(t *testing.T) {
	provider := &ruleMagnetTestProvider{magnet: "magnet:?xt=urn:btih:resolved-from-provider"}
	repo := newTorrentRepositoryForRuleMagnetTests(map[string]*ruleMagnetTestProvider{"fake": provider})

	torrentData, err := json.Marshal(&autodownloader.NormalizedTorrent{
		AnimeTorrent: &hibiketorrent.AnimeTorrent{
			Name:     "Example torrent",
			InfoHash: "hash-from-torrent",
		},
		ExtensionID: "fake",
	})
	require.NoError(t, err)

	item := &models.AutoDownloaderItem{
		Hash:        "hash-from-item",
		TorrentData: torrentData,
	}

	magnet, err := resolveAutoDownloaderItemMagnet(item, repo)
	require.NoError(t, err)
	assert.Equal(t, "magnet:?xt=urn:btih:resolved-from-provider", magnet)
	assert.Equal(t, 1, provider.calls)
}

func TestResolveAutoDownloaderItemMagnetFallsBackToHash(t *testing.T) {
	provider := &ruleMagnetTestProvider{err: errors.New("provider failed")}
	repo := newTorrentRepositoryForRuleMagnetTests(map[string]*ruleMagnetTestProvider{"fake": provider})

	torrentData, err := json.Marshal(&autodownloader.NormalizedTorrent{
		AnimeTorrent: &hibiketorrent.AnimeTorrent{
			Name:     "Example torrent",
			InfoHash: "hash-from-torrent",
		},
		ExtensionID: "fake",
	})
	require.NoError(t, err)

	item := &models.AutoDownloaderItem{
		Hash:        "hash-from-item",
		TorrentData: torrentData,
	}

	magnet, err := resolveAutoDownloaderItemMagnet(item, repo)
	require.NoError(t, err)
	assert.Equal(t, "magnet:?xt=urn:btih:hash-from-item", magnet)
	assert.Equal(t, 1, provider.calls)
}

func newTorrentRepositoryForRuleMagnetTests(providers map[string]*ruleMagnetTestProvider) *torrentrepo.Repository {
	logger := zerolog.Nop()
	bank := extension.NewUnifiedBank()
	for id, provider := range providers {
		bank.Set(id, extension.NewAnimeTorrentProviderExtension(&extension.Extension{
			ID:          id,
			Name:        id,
			Version:     "1.0.0",
			ManifestURI: "builtin",
			Language:    extension.LanguageGo,
			Type:        extension.TypeAnimeTorrentProvider,
		}, provider))
	}

	var metadata metadata_provider.Provider
	repo := torrentrepo.NewRepository(&torrentrepo.NewRepositoryOptions{
		Logger:              &logger,
		MetadataProviderRef: util.NewRef[metadata_provider.Provider](metadata),
		ExtensionBankRef:    util.NewRef(bank),
	})
	repo.SetSettings(&torrentrepo.RepositorySettings{DefaultAnimeProvider: "fake"})

	return repo
}

func TestTorrentClientGetFilesStartsClient(t *testing.T) {
	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.URL.Path)
		switch r.URL.Path {
		case "/api/v2/app/version":
			_, _ = w.Write([]byte("v5.0.0"))
		case "/api/v2/torrents/properties":
			_, _ = w.Write([]byte(`{}`))
		case "/api/v2/torrents/files":
			_, _ = w.Write([]byte(`[{"name":"episode.mkv"}]`))
		default:
			t.Errorf("unexpected request: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	logger := zerolog.Nop()
	client := qbittorrent.NewClient(&qbittorrent.NewClientOptions{
		Logger: &logger,
		Host:   server.URL,
		Path:   "unused",
	})
	h := &Handler{App: &core.App{
		Logger: &logger,
		TorrentClientRepository: torrent_client.NewRepository(&torrent_client.NewRepositoryOptions{
			Logger:            &logger,
			Provider:          torrent_client.QbittorrentClient,
			QbittorrentClient: client,
		}),
		TorrentRepository: &torrentrepo.Repository{},
	}}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/torrent-client/get-files", strings.NewReader(`{"torrent":{"infoHash":"hash","magnetLink":"magnet:?xt=urn:btih:hash"}}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	require.NoError(t, h.HandleTorrentClientGetFiles(echo.New().NewContext(req, rec)))
	var res SeaResponse[[]string]
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	require.Empty(t, res.Error)
	require.Equal(t, []string{"episode.mkv"}, res.Data)
	require.Equal(t, []string{"/api/v2/app/version", "/api/v2/torrents/properties", "/api/v2/torrents/files"}, calls)
}

func TestTorrentClientGetFilesReportsUnavailableClient(t *testing.T) {
	h := &Handler{App: &core.App{
		TorrentClientRepository: torrent_client.NewRepository(&torrent_client.NewRepositoryOptions{
			Provider: "unavailable",
		}),
	}}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/torrent-client/get-files", strings.NewReader(`{"torrent":{"infoHash":"hash"}}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	require.NoError(t, h.HandleTorrentClientGetFiles(echo.New().NewContext(req, rec)))
	var res SeaResponse[any]
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	require.Contains(t, res.Error, "could not contact torrent client")
}
