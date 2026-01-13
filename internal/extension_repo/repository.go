package extension_repo

import (
	"context"
	"net/http"
	"os"
	"seanime/internal/events"
	"seanime/internal/extension"
	hibikecustomsource "seanime/internal/extension/hibike/customsource"
	hibikemanga "seanime/internal/extension/hibike/manga"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/goja/goja_runtime"
	"seanime/internal/hook"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"seanime/internal/util/result"
	"sync"
	"sync/atomic"
	"time"

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
		extensionBankRef *util.Ref[*extension.UnifiedBank]

		invalidExtensions *result.Map[string, *extension.InvalidExtension]

		hookManager hook.Manager

		client *http.Client

		// Cache the of all built-in extensions when they're first loaded
		// This is used to quickly determine if an extension is built-in or not and to reload them
		builtinExtensions *result.Map[string, *builtinExtension]

		updateData   []UpdateData
		updateDataMu sync.Mutex

		// Called when the external extensions are loaded for the first time
		firstExternalExtensionLoadedFunc context.CancelFunc

		loadOnlyType atomic.Value
	}

	builtinExtension struct {
		extension.Extension
		provider interface{}
	}

	AllExtensions struct {
		Extensions        []*extension.Extension        `json:"extensions"`
		InvalidExtensions []*extension.InvalidExtension `json:"invalidExtensions"`
		// List of extensions with invalid user config extensions, these extensions are still loaded
		InvalidUserConfigExtensions []*extension.InvalidExtension `json:"invalidUserConfigExtensions"`
		// List of extension IDs that have an update available
		// This is only populated when the user clicks on "Check for updates"
		HasUpdate []UpdateData `json:"hasUpdate"`
		// Extensions that use unsafe flags
		UnsafeExtensions map[string]bool `json:"unsafeExtensions"`
	}

	UpdateData struct {
		ExtensionID string `json:"extensionID"`
		ManifestURI string `json:"manifestURI"`
		Version     string `json:"version"`
		Payload     string `json:"payload"`
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

	CustomSourceExtensionItem struct {
		ID                  string                      `json:"id"`
		ExtensionIdentifier int                         `json:"extensionIdentifier"`
		Name                string                      `json:"name"`
		Lang                string                      `json:"lang"` // ISO 639-1 language code
		Settings            hibikecustomsource.Settings `json:"settings"`
	}
)

type NewRepositoryOptions struct {
	Logger           *zerolog.Logger
	ExtensionDir     string
	WSEventManager   events.WSEventManagerInterface
	FileCacher       *filecache.Cacher
	HookManager      hook.Manager
	ExtensionBankRef *util.Ref[*extension.UnifiedBank]
}

func NewRepository(opts *NewRepositoryOptions) *Repository {

	// Make sure the extension directory exists
	_ = os.MkdirAll(opts.ExtensionDir, os.ModePerm)

	ret := &Repository{
		logger:             opts.Logger,
		extensionDir:       opts.ExtensionDir,
		wsEventManager:     opts.WSEventManager,
		gojaExtensions:     result.NewMap[string, GojaExtension](),
		gojaRuntimeManager: goja_runtime.NewManager(opts.Logger),
		extensionBankRef:   opts.ExtensionBankRef,
		invalidExtensions:  result.NewMap[string, *extension.InvalidExtension](),
		fileCacher:         opts.FileCacher,
		hookManager:        opts.HookManager,
		client:             http.DefaultClient,
		builtinExtensions:  result.NewMap[string, *builtinExtension](),
		updateData:         make([]UpdateData, 0),
	}

	ret.loadOnlyType.Store([]extension.Type{})

	firstExtensionLoadedCtx, firstExtensionLoadedCancel := context.WithCancel(context.Background())
	ret.firstExternalExtensionLoadedFunc = firstExtensionLoadedCancel

	// Fetch extension updates at launch and every 12 hours
	go func(firstExtensionLoadedCtx context.Context) {
		defer util.HandlePanicInModuleThen("extension_repo/fetchExtensionUpdates", func() {
			ret.firstExternalExtensionLoadedFunc = nil
		})
		for {
			if ret.firstExternalExtensionLoadedFunc != nil {
				// Block until the first external extensions are loaded
				select {
				case <-firstExtensionLoadedCtx.Done():
				}
			}

			ret.firstExternalExtensionLoadedFunc = nil

			ret.updateData = ret.checkForUpdates()
			if len(ret.updateData) > 0 {
				// Signal the frontend that there are updates available
				ret.wsEventManager.SendEvent(events.ExtensionUpdatesFound, ret.updateData)
			}
			time.Sleep(12 * time.Hour)
		}
	}(firstExtensionLoadedCtx)

	return ret
}

func (r *Repository) LoadOnlyWrapper(only []extension.Type, loadFunc func()) {
	r.loadOnlyType.Store(only)
	defer r.loadOnlyType.Store([]extension.Type{})
	loadFunc()
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
		UnsafeExtensions:            make(map[string]bool),
	}

	for _, ext := range ret.Extensions {
		if ext.Plugin != nil && ext.Plugin.IsUnsafe() {
			ret.UnsafeExtensions[ext.ID] = true
		}
	}
	for _, ext := range ret.InvalidExtensions {
		if ext.Extension.Plugin != nil && ext.Extension.Plugin.IsUnsafe() {
			ret.UnsafeExtensions[ext.ID] = true
		}
	}

	// Send the update data to the frontend if there are any updates
	if len(r.updateData) > 0 {
		ret.HasUpdate = r.updateData
	}

	if withUpdates {
		ret.HasUpdate = r.checkForUpdates()
		r.updateData = ret.HasUpdate
	}
	return
}

func (r *Repository) GetUpdateData() (ret []UpdateData) {
	return r.updateData
}

func (r *Repository) ListExtensionData() (ret []*extension.Extension) {
	r.extensionBankRef.Get().Range(func(key string, ext extension.BaseExtension) bool {
		retExt := extension.ToExtensionData(ext)
		retExt.Payload = ""
		ret = append(ret, retExt)
		return true
	})

	return ret
}

func (r *Repository) ListDevelopmentModeExtensions() (ret []*extension.Extension) {
	r.extensionBankRef.Get().Range(func(key string, ext extension.BaseExtension) bool {
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
	ext, found := r.extensionBankRef.Get().Get(id)
	if !found {
		ie, found := r.invalidExtensions.Get(id)
		if found {
			ext := ie.Extension
			ret = ext.Payload
			if len(ret) > 0 {
				return
			}

			// Fetch from payload URI
			if ext.IsDevelopment {
				return
			}

			ret, _ = r.downloadPayload(ext.PayloadURI)
			return
		}
		return ""
	}

	ret = ext.GetPayload()
	if len(ret) > 0 {
		return
	}

	// Fetch from payload URI
	if ext.GetPayloadURI() == "" || ext.GetIsDevelopment() {
		return
	}

	ret, _ = r.downloadPayload(ext.GetPayloadURI())

	return
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Lists
// - Lists are used to display available options to the user based on the extensions installed
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) ListMangaProviderExtensions() []*MangaProviderExtensionItem {
	ret := make([]*MangaProviderExtensionItem, 0)

	extension.RangeExtensions(r.extensionBankRef.Get(), func(key string, ext extension.MangaProviderExtension) bool {
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

	extension.RangeExtensions(r.extensionBankRef.Get(), func(key string, ext extension.OnlinestreamProviderExtension) bool {
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

	extension.RangeExtensions(r.extensionBankRef.Get(), func(key string, ext extension.AnimeTorrentProviderExtension) bool {
		settings := ext.GetProvider().GetSettings()
		ret = append(ret, &AnimeTorrentProviderExtensionItem{
			ID:       ext.GetID(),
			Name:     ext.GetName(),
			Lang:     extension.GetExtensionLang(ext.GetLang()),
			Settings: settings,
		})

		return true
	})

	return ret
}

func (r *Repository) ListCustomSourceExtensions() []*CustomSourceExtensionItem {
	ret := make([]*CustomSourceExtensionItem, 0)

	extension.RangeExtensions(r.extensionBankRef.Get(), func(key string, ext extension.CustomSourceExtension) bool {
		settings := ext.GetProvider().GetSettings()
		ret = append(ret, &CustomSourceExtensionItem{
			ID:                  ext.GetID(),
			ExtensionIdentifier: ext.GetExtensionIdentifier(),
			Name:                ext.GetName(),
			Lang:                extension.GetExtensionLang(ext.GetLang()),
			Settings:            settings,
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
	ext, found := r.extensionBankRef.Get().Get(id)
	if found {
		return ext, true
	}

	return nil, false
}

func (r *Repository) GetExtensionBank() *extension.UnifiedBank {
	return r.extensionBankRef.Get()
}

func (r *Repository) GetMangaProviderExtensionByID(id string) (extension.MangaProviderExtension, bool) {
	ext, found := extension.GetExtension[extension.MangaProviderExtension](r.extensionBankRef.Get(), id)
	return ext, found
}

func (r *Repository) GetOnlinestreamProviderExtensionByID(id string) (extension.OnlinestreamProviderExtension, bool) {
	ext, found := extension.GetExtension[extension.OnlinestreamProviderExtension](r.extensionBankRef.Get(), id)
	return ext, found
}

func (r *Repository) GetAnimeTorrentProviderExtensionByID(id string) (extension.AnimeTorrentProviderExtension, bool) {
	ext, found := extension.GetExtension[extension.AnimeTorrentProviderExtension](r.extensionBankRef.Get(), id)
	return ext, found
}

func (r *Repository) GetCustomSourceExtensionByID(id string) (extension.CustomSourceExtension, bool) {
	ext, found := extension.GetExtension[extension.CustomSourceExtension](r.extensionBankRef.Get(), id)
	return ext, found
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
