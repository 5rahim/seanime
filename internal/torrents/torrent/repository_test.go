package torrent

import (
	"seanime/internal/api/metadata"
	"seanime/internal/extension"
	"seanime/internal/torrents/animetosho"
	"seanime/internal/torrents/nyaa"
	"seanime/internal/torrents/seadex"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"testing"
)

func getTestRepo(t *testing.T) *Repository {
	logger := util.NewLogger()
	metadataProvider := metadata.TestGetMockProvider(t)

	extensions := result.NewResultMap[string, extension.AnimeTorrentProviderExtension]()

	extensions.Set("nyaa", extension.NewAnimeTorrentProviderExtension(&extension.Extension{
		ID:       "nyaa",
		Name:     "Nyaa",
		Version:  "1.0.0",
		Language: extension.LanguageGo,
		Type:     extension.TypeTorrentProvider,
		Author:   "Seanime",
	}, nyaa.NewProvider(logger)))

	extensions.Set("nyaa-sukebei", extension.NewAnimeTorrentProviderExtension(&extension.Extension{
		ID:       "nyaa-sukebei",
		Name:     "Nyaa Sukebei",
		Version:  "1.0.0",
		Language: extension.LanguageGo,
		Type:     extension.TypeTorrentProvider,
		Author:   "Seanime",
	}, nyaa.NewSukebeiProvider(logger)))

	extensions.Set("animetosho", extension.NewAnimeTorrentProviderExtension(&extension.Extension{
		ID:       "animetosho",
		Name:     "AnimeTosho",
		Version:  "1.0.0",
		Language: extension.LanguageGo,
		Type:     extension.TypeTorrentProvider,
		Author:   "Seanime",
	}, animetosho.NewProvider(logger)))

	extensions.Set("seadex", extension.NewAnimeTorrentProviderExtension(&extension.Extension{
		ID:       "seadex",
		Name:     "SeaDex",
		Version:  "1.0.0",
		Language: extension.LanguageGo,
		Type:     extension.TypeTorrentProvider,
		Author:   "Seanime",
	}, seadex.NewProvider(logger)))

	repo := NewRepository(&NewRepositoryOptions{
		Logger:           logger,
		MetadataProvider: metadataProvider,
	})

	repo.SetAnimeProviderExtensions(extensions)

	repo.SetSettings(&RepositorySettings{
		DefaultAnimeProvider: ProviderAnimeTosho,
	})

	return repo
}
