package extension_repo

import (
	hibikemanga "github.com/5rahim/hibike/pkg/extension/manga"
	hibikeonlinestream "github.com/5rahim/hibike/pkg/extension/onlinestream"
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/traefik/yaegi/interp"
	"os"
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/extension/vendoring/torrent"
	"seanime/internal/util/result"
)

type (
	// Repository manages all extensions
	Repository struct {
		logger         *zerolog.Logger
		wsEventManager events.WSEventManagerInterface
		// Absolute path to the directory containing all extensions
		extensionDir string
		// Yaegi interpreter for Go extensions
		yaegiInterp *interp.Interpreter
		// Store all active Goja VMs
		// - When reloading extensions, all VMs are interrupted
		gojaExtensions *result.Map[string, GojaExtension]
		// Extension bank
		// - When reloading extensions, external extensions are removed & re-added
		extensionBank *extension.UnifiedBank

		invalidExtensions *result.Map[string, *extension.InvalidExtension]
	}

	AllExtensions struct {
		Extensions        []*extension.Extension        `json:"extensions"`
		InvalidExtensions []*extension.InvalidExtension `json:"invalidExtensions"`
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
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	OnlinestreamProviderExtensionItem struct {
		ID             string   `json:"id"`
		Name           string   `json:"name"`
		EpisodeServers []string `json:"episodeServers"`
	}

	AnimeTorrentProviderExtensionItem struct {
		ID       string                                      `json:"id"`
		Name     string                                      `json:"name"`
		Settings vendor_hibike_torrent.AnimeProviderSettings `json:"settings"`
	}
)

type NewRepositoryOptions struct {
	Logger         *zerolog.Logger
	ExtensionDir   string
	WSEventManager events.WSEventManagerInterface
}

func NewRepository(opts *NewRepositoryOptions) *Repository {

	// Make sure the extension directory exists
	_ = os.MkdirAll(opts.ExtensionDir, os.ModePerm)

	ret := &Repository{
		logger:            opts.Logger,
		extensionDir:      opts.ExtensionDir,
		wsEventManager:    opts.WSEventManager,
		gojaExtensions:    result.NewResultMap[string, GojaExtension](),
		extensionBank:     extension.NewUnifiedBank(),
		invalidExtensions: result.NewResultMap[string, *extension.InvalidExtension](),
	}

	ret.loadYaegiInterpreter()

	return ret
}

func (r *Repository) GetAllExtensions(withUpdates bool) (ret *AllExtensions) {
	ret = &AllExtensions{
		Extensions:        r.ListExtensionData(),
		InvalidExtensions: r.ListInvalidExtensions(),
	}
	if withUpdates {
		ret.HasUpdate = r.checkForUpdates()
	}
	return
}

func (r *Repository) ListExtensionData() (ret []*extension.Extension) {
	r.extensionBank.Range(func(key string, ext extension.BaseExtension) bool {
		ret = append(ret, extension.InstalledToExtensionData(ext))
		return true
	})

	return ret
}

func (r *Repository) ListInvalidExtensions() (ret []*extension.InvalidExtension) {
	r.invalidExtensions.Range(func(key string, ext *extension.InvalidExtension) bool {
		//ext.Extension.Payload = "" // Remove the payload so the client knows the extension is installed
		ret = append(ret, ext)
		return true
	})

	return ret
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Lists
// - Lists are used to display available options to the user based on the extensions installed
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) ListMangaProviderExtensions() []*MangaProviderExtensionItem {
	ret := make([]*MangaProviderExtensionItem, 0)

	extension.RangeExtensions(r.extensionBank, func(key string, ext extension.MangaProviderExtension) bool {
		ret = append(ret, &MangaProviderExtensionItem{
			ID:   ext.GetID(),
			Name: ext.GetName(),
		})
		return true
	})

	return ret
}

func (r *Repository) ListOnlinestreamProviderExtensions() []*OnlinestreamProviderExtensionItem {
	ret := make([]*OnlinestreamProviderExtensionItem, 0)

	extension.RangeExtensions(r.extensionBank, func(key string, ext extension.OnlinestreamProviderExtension) bool {
		ret = append(ret, &OnlinestreamProviderExtensionItem{
			ID:             ext.GetID(),
			Name:           ext.GetName(),
			EpisodeServers: ext.GetProvider().GetEpisodeServers(),
		})
		return true
	})

	return ret
}

func (r *Repository) ListAnimeTorrentProviderExtensions() []*AnimeTorrentProviderExtensionItem {
	ret := make([]*AnimeTorrentProviderExtensionItem, 0)

	extension.RangeExtensions(r.extensionBank, func(key string, ext extension.AnimeTorrentProviderExtension) bool {
		ret = append(ret, &AnimeTorrentProviderExtensionItem{
			ID:   ext.GetID(),
			Name: ext.GetName(),
			Settings: vendor_hibike_torrent.AnimeProviderSettings{
				Type:           vendor_hibike_torrent.AnimeProviderType(ext.GetProvider().GetSettings().Type),
				CanSmartSearch: ext.GetProvider().GetSettings().CanSmartSearch,
				SupportsAdult:  ext.GetProvider().GetSettings().SupportsAdult,
				SmartSearchFilters: lo.Map(ext.GetProvider().GetSettings().SmartSearchFilters, func(value hibiketorrent.AnimeProviderSmartSearchFilter, _ int) vendor_hibike_torrent.AnimeProviderSmartSearchFilter {
					return vendor_hibike_torrent.AnimeProviderSmartSearchFilter(value)
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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
