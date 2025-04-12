package extension_repo

import (
	"os"
	"seanime/internal/events"
	"seanime/internal/extension"
	hibikemanga "seanime/internal/extension/hibike/manga"
	hibikeonlinestream "seanime/internal/extension/hibike/onlinestream"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/goja/goja_runtime"
	"seanime/internal/hook"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"seanime/internal/util/result"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

type (
	// Repository manages all extensions
	Repository struct {
		logger         *zerolog.Logger
		fileCacher     *filecache.Cacher
		wsEventManager events.WSEventManagerInterface
		// Absolute path to the directory containing all extensions
		extensionDir string
		// Store all active Goja VMs
		// - When reloading extensions, all VMs are interrupted
		gojaExtensions *result.Map[string, GojaExtension]

		gojaRuntimeManager *goja_runtime.Manager
		// Extension bank
		// - When reloading extensions, external extensions are removed & re-added
		extensionBank *extension.UnifiedBank

		invalidExtensions *result.Map[string, *extension.InvalidExtension]

		hookManager hook.Manager
	}

	AllExtensions struct {
		Extensions        []*extension.Extension        `json:"extensions"`
		InvalidExtensions []*extension.InvalidExtension `json:"invalidExtensions"`
		// List of extensions with invalid user config extensions, these extensions are still loaded
		InvalidUserConfigExtensions []*extension.InvalidExtension `json:"invalidUserConfigExtensions"`
		// List of extension IDs that have an update available
		// This is only populated when the user clicks on "Check for updates"
		HasUpdate []UpdateData `json:"hasUpdate"`
	}

	UpdateData struct {
		ExtensionID string `json:"extensionID"`
		ManifestURI string `json:"manifestURI"`
		Version     string `json:"version"`
	}

	MangaProviderExtensionItem struct {
		ID       string               `json:"id"`
		Name     string               `json:"name"`
		Lang     string               `json:"lang"` // ISO 639-1 language code
		Settings hibikemanga.Settings `json:"settings"`
	}

	OnlinestreamProviderExtensionItem struct {
		ID             string   `json:"id"`
		Name           string   `json:"name"`
		Lang           string   `json:"lang"` // ISO 639-1 language code
		EpisodeServers []string `json:"episodeServers"`
		SupportsDub    bool     `json:"supportsDub"`
	}

	AnimeTorrentProviderExtensionItem struct {
		ID       string                              `json:"id"`
		Name     string                              `json:"name"`
		Lang     string                              `json:"lang"` // ISO 639-1 language code
		Settings hibiketorrent.AnimeProviderSettings `json:"settings"`
	}
)

type NewRepositoryOptions struct {
	Logger         *zerolog.Logger
	ExtensionDir   string
	WSEventManager events.WSEventManagerInterface
	FileCacher     *filecache.Cacher
	HookManager    hook.Manager
}

func NewRepository(opts *NewRepositoryOptions) *Repository {

	// Make sure the extension directory exists
	_ = os.MkdirAll(opts.ExtensionDir, os.ModePerm)

	ret := &Repository{
		logger:             opts.Logger,
		extensionDir:       opts.ExtensionDir,
		wsEventManager:     opts.WSEventManager,
		gojaExtensions:     result.NewResultMap[string, GojaExtension](),
		gojaRuntimeManager: goja_runtime.NewManager(opts.Logger, 20),
		extensionBank:      extension.NewUnifiedBank(),
		invalidExtensions:  result.NewResultMap[string, *extension.InvalidExtension](),
		fileCacher:         opts.FileCacher,
		hookManager:        opts.HookManager,
	}

	return ret
}

func (r *Repository) GetAllExtensions(withUpdates bool) (ret *AllExtensions) {
	invalidExtensions := r.ListInvalidExtensions()

	fatalInvalidExtensions := lo.Filter(invalidExtensions, func(ext *extension.InvalidExtension, _ int) bool {
		return ext.Code != extension.InvalidExtensionUserConfigError
	})

	userConfigInvalidExtensions := lo.Filter(invalidExtensions, func(ext *extension.InvalidExtension, _ int) bool {
		return ext.Code == extension.InvalidExtensionUserConfigError
	})

	ret = &AllExtensions{
		Extensions:                  r.ListExtensionData(),
		InvalidExtensions:           fatalInvalidExtensions,
		InvalidUserConfigExtensions: userConfigInvalidExtensions,
	}
	if withUpdates {
		ret.HasUpdate = r.checkForUpdates()
	}
	return
}

func (r *Repository) ListExtensionData() (ret []*extension.Extension) {
	r.extensionBank.Range(func(key string, ext extension.BaseExtension) bool {
		retExt := extension.ToExtensionData(ext)
		retExt.Payload = ""
		ret = append(ret, retExt)
		return true
	})

	return ret
}

func (r *Repository) ListDevelopmentModeExtensions() (ret []*extension.Extension) {
	r.extensionBank.Range(func(key string, ext extension.BaseExtension) bool {
		if ext.GetIsDevelopment() {
			retExt := extension.ToExtensionData(ext)
			retExt.Payload = ""
			ret = append(ret, retExt)
		}
		return true
	})

	return ret
}

func (r *Repository) ListInvalidExtensions() (ret []*extension.InvalidExtension) {
	r.invalidExtensions.Range(func(key string, ext *extension.InvalidExtension) bool {
		ext.Extension.Payload = ""
		ret = append(ret, ext)
		return true
	})

	return ret
}

func (r *Repository) GetExtensionPayload(id string) (ret string) {
	ext, found := r.extensionBank.Get(id)
	if !found {
		return ""
	}
	return ext.GetPayload()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Lists
// - Lists are used to display available options to the user based on the extensions installed
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) ListMangaProviderExtensions() []*MangaProviderExtensionItem {
	ret := make([]*MangaProviderExtensionItem, 0)

	extension.RangeExtensions(r.extensionBank, func(key string, ext extension.MangaProviderExtension) bool {
		settings := ext.GetProvider().GetSettings()
		ret = append(ret, &MangaProviderExtensionItem{
			ID:       ext.GetID(),
			Name:     ext.GetName(),
			Lang:     extension.GetExtensionLang(ext.GetLang()),
			Settings: settings,
		})
		return true
	})

	return ret
}

func (r *Repository) ListOnlinestreamProviderExtensions() []*OnlinestreamProviderExtensionItem {
	ret := make([]*OnlinestreamProviderExtensionItem, 0)

	extension.RangeExtensions(r.extensionBank, func(key string, ext extension.OnlinestreamProviderExtension) bool {
		settings := ext.GetProvider().GetSettings()
		ret = append(ret, &OnlinestreamProviderExtensionItem{
			ID:             ext.GetID(),
			Name:           ext.GetName(),
			Lang:           extension.GetExtensionLang(ext.GetLang()),
			EpisodeServers: settings.EpisodeServers,
			SupportsDub:    settings.SupportsDub,
		})
		return true
	})

	return ret
}

func (r *Repository) ListAnimeTorrentProviderExtensions() []*AnimeTorrentProviderExtensionItem {
	ret := make([]*AnimeTorrentProviderExtensionItem, 0)

	extension.RangeExtensions(r.extensionBank, func(key string, ext extension.AnimeTorrentProviderExtension) bool {
		settings := ext.GetProvider().GetSettings()
		ret = append(ret, &AnimeTorrentProviderExtensionItem{
			ID:   ext.GetID(),
			Name: ext.GetName(),
			Lang: extension.GetExtensionLang(ext.GetLang()),
			Settings: hibiketorrent.AnimeProviderSettings{
				Type:           settings.Type,
				CanSmartSearch: settings.CanSmartSearch,
				SupportsAdult:  settings.SupportsAdult,
				SmartSearchFilters: lo.Map(settings.SmartSearchFilters, func(value hibiketorrent.AnimeProviderSmartSearchFilter, _ int) hibiketorrent.AnimeProviderSmartSearchFilter {
					return value
				}),
			},
		})

		return true
	})

	return ret
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// GetLoadedExtension returns the loaded extension by ID.
// It returns an extension.BaseExtension interface, so it can be used to get the extension's details.
func (r *Repository) GetLoadedExtension(id string) (extension.BaseExtension, bool) {
	var ext extension.BaseExtension
	ext, found := r.extensionBank.Get(id)
	if found {
		return ext, true
	}

	return nil, false
}

func (r *Repository) GetExtensionBank() *extension.UnifiedBank {
	return r.extensionBank
}

func (r *Repository) GetMangaProviderExtensionByID(id string) (extension.MangaProviderExtension, bool) {
	ext, found := extension.GetExtension[extension.MangaProviderExtension](r.extensionBank, id)
	return ext, found
}

func (r *Repository) GetOnlinestreamProviderExtensionByID(id string) (extension.OnlinestreamProviderExtension, bool) {
	ext, found := extension.GetExtension[extension.OnlinestreamProviderExtension](r.extensionBank, id)
	return ext, found
}

func (r *Repository) GetAnimeTorrentProviderExtensionByID(id string) (extension.AnimeTorrentProviderExtension, bool) {
	ext, found := extension.GetExtension[extension.AnimeTorrentProviderExtension](r.extensionBank, id)
	return ext, found
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Built-in extensions
// - Built-in extensions are loaded once, on application startup
// - The "manifestURI" field is set to "builtin" to indicate that the extension is not external
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) LoadBuiltInMangaProviderExtension(info extension.Extension, provider hibikemanga.Provider) {
	r.extensionBank.Set(info.ID, extension.NewMangaProviderExtension(&info, provider))
	r.logger.Debug().Str("id", info.ID).Msg("extensions: Loaded built-in manga provider extension")
}

func (r *Repository) LoadBuiltInAnimeTorrentProviderExtension(info extension.Extension, provider hibiketorrent.AnimeProvider) {
	r.extensionBank.Set(info.ID, extension.NewAnimeTorrentProviderExtension(&info, provider))
	r.logger.Debug().Str("id", info.ID).Msg("extensions: Loaded built-in anime torrent provider extension")
}

func (r *Repository) LoadBuiltInOnlinestreamProviderExtension(info extension.Extension, provider hibikeonlinestream.Provider) {
	r.extensionBank.Set(info.ID, extension.NewOnlinestreamProviderExtension(&info, provider))
	r.logger.Debug().Str("id", info.ID).Msg("extensions: Loaded built-in onlinestream provider extension")
}

func (r *Repository) LoadBuiltInOnlinestreamProviderExtensionJS(info extension.Extension) {
	err := r.loadExternalOnlinestreamExtensionJS(&info, info.Language)
	if err != nil {
		r.logger.Error().Err(err).Str("id", info.ID).Msg("extensions: Failed to load built-in JS onlinestream provider extension")
		return
	}
	r.logger.Debug().Str("id", info.ID).Msg("extensions: Loaded built-in onlinestream provider extension")
}

func (r *Repository) loadPlugin(ext *extension.Extension) (err error) {
	defer util.HandlePanicInModuleWithError("extension_repo/loadPlugin", &err)

	err = r.loadPluginExtension(ext)
	if err != nil {
		r.logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to load plugin")
		return err
	}

	r.logger.Debug().Str("id", ext.ID).Msg("extensions: Loaded plugin")
	return
}
