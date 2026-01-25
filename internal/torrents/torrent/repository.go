package torrent

import (
	"seanime/internal/api/metadata_provider"
	"seanime/internal/extension"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"slices"
	"sync"

	"github.com/rs/zerolog"
)

type (
	Repository struct {
		logger                         *zerolog.Logger
		extensionBankRef               *util.Ref[*extension.UnifiedBank]
		animeProviderSearchCaches      *result.Map[string, *result.Cache[string, *SearchData]]
		animeProviderSmartSearchCaches *result.Map[string, *result.Cache[string, *SearchData]]
		settings                       RepositorySettings
		metadataProviderRef            *util.Ref[metadata_provider.Provider]
		mu                             sync.Mutex
	}

	RepositorySettings struct {
		DefaultAnimeProvider string // Default torrent provider
		AutoSelectProvider   string
	}
)

type NewRepositoryOptions struct {
	Logger              *zerolog.Logger
	MetadataProviderRef *util.Ref[metadata_provider.Provider]
	ExtensionBankRef    *util.Ref[*extension.UnifiedBank]
}

func NewRepository(opts *NewRepositoryOptions) *Repository {
	ret := &Repository{
		logger:                         opts.Logger,
		metadataProviderRef:            opts.MetadataProviderRef,
		extensionBankRef:               opts.ExtensionBankRef,
		animeProviderSearchCaches:      result.NewMap[string, *result.Cache[string, *SearchData]](),
		animeProviderSmartSearchCaches: result.NewMap[string, *result.Cache[string, *SearchData]](),
		settings:                       RepositorySettings{},
		mu:                             sync.Mutex{},
	}

	sub := ret.extensionBankRef.Get().Subscribe("torrent-repository")

	go func() {
		for {
			select {
			case <-sub.OnExtensionAdded():
				//r.logger.Debug().Msg("torrent repo: Anime provider extension added")
				ret.OnExtensionReloaded()
			case <-sub.OnExtensionRemoved():
				ret.OnExtensionReloaded()
			}
		}
	}()

	return ret
}
func (r *Repository) OnExtensionReloaded() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reloadExtensions()
}

// This is called each time a new extension is added or removed
func (r *Repository) reloadExtensions() {
	// Clear the search caches
	r.animeProviderSearchCaches = result.NewMap[string, *result.Cache[string, *SearchData]]()
	r.animeProviderSmartSearchCaches = result.NewMap[string, *result.Cache[string, *SearchData]]()

	go func() {
		// Create new caches for each provider
		extension.RangeExtensions(r.extensionBankRef.Get(), func(provider string, value extension.AnimeTorrentProviderExtension) bool {
			r.animeProviderSearchCaches.Set(provider, result.NewCache[string, *SearchData]())
			r.animeProviderSmartSearchCaches.Set(provider, result.NewCache[string, *SearchData]())
			return true
		})
	}()

	// Check if the default provider is in the list of providers
	//if r.settings.DefaultAnimeProvider != "" && r.settings.DefaultAnimeProvider != "none" {
	//	if _, ok := r.extensionBankRef.Get().Get(r.settings.DefaultAnimeProvider); !ok {
	//		//r.logger.Error().Str("defaultProvider", r.settings.DefaultAnimeProvider).Msg("torrent repo: Default torrent provider not found in extensions")
	//		// Set the default provider to empty
	//		r.settings.DefaultAnimeProvider = ""
	//	}
	//}

	//r.logger.Trace().Str("defaultProvider", r.settings.DefaultAnimeProvider).Msg("torrent repo: Reloaded extensions")
}

// SetSettings should be called after the repository is created and settings are refreshed
func (r *Repository) SetSettings(s *RepositorySettings) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.logger.Trace().Msg("torrent repo: Setting settings")

	if s != nil {
		r.settings = *s
	} else {
		r.settings = RepositorySettings{
			DefaultAnimeProvider: "",
		}
	}

	if r.settings.DefaultAnimeProvider == "none" {
		r.settings.DefaultAnimeProvider = ""
	}

	// Reload extensions after settings change
	r.reloadExtensions()
}

func (r *Repository) GetDefaultAnimeProviderExtension() (extension.AnimeTorrentProviderExtension, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := r.settings.DefaultAnimeProvider
	// if no selected default provider, get the first one
	if id == "" {
		ids := r.GetAllAnimeProviderExtensionIds()
		if len(ids) > 0 {
			id = ids[0]
		}
		if id == "" {
			return nil, false
		}
	}
	return r.GetAnimeProviderExtensionOrFirst(id)
}

// DEPRECATED: Use GetDefaultAnimeProviderExtension instead
func (r *Repository) GetAutoSelectProviderExtension() (extension.AnimeTorrentProviderExtension, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	provider := r.settings.AutoSelectProvider
	if provider == "" {
		id := r.settings.DefaultAnimeProvider
		if id == "" {
			ids := r.GetAllAnimeProviderExtensionIds()
			if len(ids) > 0 {
				id = ids[0]
			}
			if id == "" {
				return nil, false
			}
		}
		provider = id
	}
	return r.GetAnimeProviderExtensionOrFirst(provider)
}

func (r *Repository) GetAnimeProviderExtension(id string) (extension.AnimeTorrentProviderExtension, bool) {
	return extension.GetExtension[extension.AnimeTorrentProviderExtension](r.extensionBankRef.Get(), id)
}

// GetAnimeProviderExtensionOrFirst returns the extension with the given ID, or the first one if the extension is not found
func (r *Repository) GetAnimeProviderExtensionOrFirst(id string) (extension.AnimeTorrentProviderExtension, bool) {
	ext, found := extension.GetExtension[extension.AnimeTorrentProviderExtension](r.extensionBankRef.Get(), id)
	if !found {
		ids := r.GetAllAnimeProviderExtensionIds()
		// check if we find the default extension
		if r.settings.DefaultAnimeProvider != "" && slices.Contains(ids, r.settings.DefaultAnimeProvider) {
			id = r.settings.DefaultAnimeProvider
		} else if len(ids) > 0 {
			id = ids[0]
		} else {
			return nil, false
		}
		ext, _ = extension.GetExtension[extension.AnimeTorrentProviderExtension](r.extensionBankRef.Get(), id)
	}
	return ext, true
}

// GetAnimeProviderExtensionOrDefault returns the extension with the given ID, or the default one
func (r *Repository) GetAnimeProviderExtensionOrDefault(id string) (extension.AnimeTorrentProviderExtension, bool) {
	ext, found := extension.GetExtension[extension.AnimeTorrentProviderExtension](r.extensionBankRef.Get(), id)
	if !found {
		return r.GetDefaultAnimeProviderExtension()
	}
	return ext, true
}

func (r *Repository) GetAllAnimeProviderExtensionIds() []string {
	ids := make([]string, 0)
	extension.RangeExtensions[extension.AnimeTorrentProviderExtension](r.extensionBankRef.Get(), func(id string, ext extension.AnimeTorrentProviderExtension) bool {
		ids = append(ids, id)
		return true
	})
	return ids
}

func (r *Repository) GetAnimeProviderExtensionsBy(pred func(ext extension.AnimeTorrentProviderExtension) bool) []extension.AnimeTorrentProviderExtension {
	exts := make([]extension.AnimeTorrentProviderExtension, 0)
	extension.RangeExtensions[extension.AnimeTorrentProviderExtension](r.extensionBankRef.Get(), func(id string, ext extension.AnimeTorrentProviderExtension) bool {
		if pred(ext) {
			exts = append(exts, ext)
		}
		return true
	})
	return exts
}
