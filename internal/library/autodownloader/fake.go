package autodownloader

import (
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/events"
	"seanime/internal/extension"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/test_utils"
	"seanime/internal/torrents/torrent"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"testing"

	"github.com/stretchr/testify/require"
)

type Fake struct {
	SearchResults    []*hibiketorrent.AnimeTorrent
	GetLatestResults []*hibiketorrent.AnimeTorrent
	Database         *db.Database
}

type FakeTorrentProvider struct {
	fake *Fake
}

func (f FakeTorrentProvider) Search(opts hibiketorrent.AnimeSearchOptions) ([]*hibiketorrent.AnimeTorrent, error) {
	return f.fake.SearchResults, nil
}

func (f FakeTorrentProvider) SmartSearch(opts hibiketorrent.AnimeSmartSearchOptions) ([]*hibiketorrent.AnimeTorrent, error) {
	return f.fake.SearchResults, nil
}

func (f FakeTorrentProvider) GetTorrentInfoHash(torrent *hibiketorrent.AnimeTorrent) (string, error) {
	return torrent.InfoHash, nil
}

func (f FakeTorrentProvider) GetTorrentMagnetLink(torrent *hibiketorrent.AnimeTorrent) (string, error) {
	return torrent.MagnetLink, nil
}

func (f FakeTorrentProvider) GetLatest() ([]*hibiketorrent.AnimeTorrent, error) {
	return f.fake.GetLatestResults, nil
}

func (f FakeTorrentProvider) GetSettings() hibiketorrent.AnimeProviderSettings {
	return hibiketorrent.AnimeProviderSettings{
		CanSmartSearch:     false,
		SmartSearchFilters: nil,
		SupportsAdult:      false,
		Type:               "main",
	}
}

var _ hibiketorrent.AnimeProvider = (*FakeTorrentProvider)(nil)

func (f *Fake) New(t *testing.T) *AutoDownloader {
	logger := util.NewLogger()
	database, err := db.NewDatabase("", test_utils.ConfigData.Database.Name, logger)
	require.NoError(t, err)

	f.Database = database

	filecacher, err := filecache.NewCacher(t.TempDir())
	require.NoError(t, err)

	extensionBankRef := util.NewRef(extension.NewUnifiedBank())

	// Fake Extension
	provider := FakeTorrentProvider{fake: f}
	ext := extension.NewAnimeTorrentProviderExtension(&extension.Extension{
		ID:   "fake",
		Type: extension.TypeAnimeTorrentProvider,
		Name: "Fake Provider",
	}, provider)

	extensionBankRef.Get().Set("fake", ext)

	metadataProvider := metadata_provider.NewProvider(&metadata_provider.NewProviderImplOptions{
		Logger:           logger,
		FileCacher:       filecacher,
		Database:         database,
		ExtensionBankRef: extensionBankRef,
	})

	torrentRepository := torrent.NewRepository(&torrent.NewRepositoryOptions{
		Logger:              logger,
		MetadataProviderRef: util.NewRef(metadataProvider),
		ExtensionBankRef:    extensionBankRef,
	})

	metadataProviderRef := util.NewRef[metadata_provider.Provider](metadataProvider)
	//torrentClientRepository := torrent_client.NewRepository(&torrent_client.NewRepositoryOptions{
	//	Logger:              logger,
	//	QbittorrentClient:   &qbittorrent.Client{},
	//	Transmission:        &transmission.Transmission{},
	//	TorrentRepository:   torrentRepository,
	//	Provider:            "",
	//	MetadataProviderRef: nil,
	//})
	ad := New(&NewAutoDownloaderOptions{
		Logger:                  logger,
		TorrentClientRepository: nil,
		TorrentRepository:       torrentRepository,
		WSEventManager:          events.NewMockWSEventManager(logger),
		Database:                database,
		MetadataProviderRef:     metadataProviderRef,
		DebridClientRepository:  nil,
		IsOfflineRef:            util.NewRef(false),
	})

	ad.SetSettings(&models.AutoDownloaderSettings{
		Provider:              "fake",
		Interval:              15,
		Enabled:               true,
		DownloadAutomatically: false,
		EnableEnhancedQueries: false,
		EnableSeasonCheck:     false,
		UseDebrid:             false,
	})
	ad.SetAnimeCollection(&anilist.AnimeCollection{})

	return ad
}
