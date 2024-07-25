package torrent

import (
	"github.com/rs/zerolog"
	"seanime/internal/api/anizip"
	"seanime/internal/api/metadata"
	"seanime/internal/extension"
	"seanime/internal/util/result"
	"sync"
)

type (
	Repository struct {
		logger                         *zerolog.Logger
		animeProviderExtensions        *result.Map[string, extension.AnimeTorrentProviderExtension]
		animeProviderSearchCaches      *result.Map[string, *result.Cache[string, *SearchData]]
		animeProviderSmartSearchCaches *result.Map[string, *result.Cache[string, *SearchData]]
		anizipCache                    *anizip.Cache
		settings                       RepositorySettings
		metadataProvider               *metadata.Provider
		mu                             sync.Mutex
	}

	RepositorySettings struct {
		DefaultAnimeProvider string // Default torrent provider
	}
)

type NewRepositoryOptions struct {
	Logger           *zerolog.Logger
	MetadataProvider *metadata.Provider
}

func NewRepository(opts *NewRepositoryOptions) *Repository {
	ret := &Repository{
		logger:                         opts.Logger,
		metadataProvider:               opts.MetadataProvider,
		animeProviderExtensions:        result.NewResultMap[string, extension.AnimeTorrentProviderExtension](),
		animeProviderSearchCaches:      result.NewResultMap[string, *result.Cache[string, *SearchData]](),
		animeProviderSmartSearchCaches: result.NewResultMap[string, *result.Cache[string, *SearchData]](),
		anizipCache:                    anizip.NewCache(),
		settings:                       RepositorySettings{},
		mu:                             sync.Mutex{},
	}

	return ret
}

func (r *Repository) SetAnimeProviderExtensions(extensions *result.Map[string, extension.AnimeTorrentProviderExtension]) {
	r.animeProviderExtensions = extensions

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if the default provider is in the list of providers
	if r.settings.DefaultAnimeProvider != "" && r.settings.DefaultAnimeProvider != "none" {
		if _, ok := r.animeProviderExtensions.Get(r.settings.DefaultAnimeProvider); !ok {
			r.logger.Error().Str("defaultProvider", r.settings.DefaultAnimeProvider).Msg("torrent repo: Default torrent provider not found in extensions")
			// Set the default provider to empty
			r.settings.DefaultAnimeProvider = ""
		}
	}

	// Clear the search caches
	r.animeProviderSearchCaches = result.NewResultMap[string, *result.Cache[string, *SearchData]]()
	r.animeProviderSmartSearchCaches = result.NewResultMap[string, *result.Cache[string, *SearchData]]()

	r.animeProviderExtensions.Range(func(provider string, value extension.AnimeTorrentProviderExtension) bool {
		r.animeProviderSearchCaches.Set(provider, result.NewCache[string, *SearchData]())
		r.animeProviderSmartSearchCaches.Set(provider, result.NewCache[string, *SearchData]())
		return true
	})
}

// SetSettings should be called after the repository is created and settings are refreshed
func (r *Repository) SetSettings(s *RepositorySettings) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if s == nil {
		r.settings = RepositorySettings{
			DefaultAnimeProvider: "",
		}
	}
	r.settings = *s

	if r.settings.DefaultAnimeProvider == "none" {
		r.settings.DefaultAnimeProvider = ""
	}
}

func (r *Repository) GetDefaultAnimeProviderExtension() (extension.AnimeTorrentProviderExtension, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.settings.DefaultAnimeProvider == "" {
		return nil, false
	}
	return r.animeProviderExtensions.Get(r.settings.DefaultAnimeProvider)
}

func (r *Repository) GetAnimeProviderExtension(id string) (extension.AnimeTorrentProviderExtension, bool) {
	return r.animeProviderExtensions.Get(id)
}
