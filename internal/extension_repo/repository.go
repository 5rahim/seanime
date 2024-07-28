package extension_repo

import (
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"os"
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/extension/vendoring/torrent"
	"seanime/internal/util/result"
	"seanime/internal/yaegi_interp"

	"github.com/dop251/goja"

	hibikemanga "github.com/5rahim/hibike/pkg/extension/manga"
	hibikeonlinestream "github.com/5rahim/hibike/pkg/extension/onlinestream"
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
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
		// Goja VMs for JS extensions
		gojaVMs *result.Map[string, *goja.Runtime]
		// Extension banks
		mangaProviderExtensionBank        *extension.Bank[extension.MangaProviderExtension]
		animeTorrentProviderExtensionBank *extension.Bank[extension.AnimeTorrentProviderExtension]
		onlinestreamProviderExtensionBank *extension.Bank[extension.OnlinestreamProviderExtension]
		invalidExtensions                 *result.Map[string, *extension.InvalidExtension]
	}

	AllExtensions struct {
		Extensions        []*extension.Extension        `json:"extensions"`
		InvalidExtensions []*extension.InvalidExtension `json:"invalidExtensions"`
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
		logger:                            opts.Logger,
		extensionDir:                      opts.ExtensionDir,
		wsEventManager:                    opts.WSEventManager,
		gojaVMs:                           result.NewResultMap[string, *goja.Runtime](),
		mangaProviderExtensionBank:        extension.NewBank[extension.MangaProviderExtension](),
		animeTorrentProviderExtensionBank: extension.NewBank[extension.AnimeTorrentProviderExtension](),
		onlinestreamProviderExtensionBank: extension.NewBank[extension.OnlinestreamProviderExtension](),
		invalidExtensions:                 result.NewResultMap[string, *extension.InvalidExtension](),
	}

	ret.loadYaegiInterpreter()

	return ret
}

func (r *Repository) GetAllExtensions() (ret *AllExtensions) {
	ret = &AllExtensions{
		Extensions:        r.ListExtensionData(),
		InvalidExtensions: r.ListInvalidExtensions(),
	}

	return
}

func (r *Repository) ListExtensionData() (ret []*extension.Extension) {
	r.mangaProviderExtensionBank.Range(func(key string, ext extension.MangaProviderExtension) bool {
		ret = append(ret, extension.InstalledToExtensionData(ext))
		return true
	})

	r.animeTorrentProviderExtensionBank.Range(func(key string, ext extension.AnimeTorrentProviderExtension) bool {
		ret = append(ret, extension.InstalledToExtensionData(ext))
		return true
	})

	r.onlinestreamProviderExtensionBank.Range(func(key string, ext extension.OnlinestreamProviderExtension) bool {
		ret = append(ret, extension.InstalledToExtensionData(ext))
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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) loadYaegiInterpreter() {
	i := interp.New(interp.Options{
		Unrestricted: false,
	})

	symbols := stdlib.Symbols
	// Remove symbols from stdlib that are risky to give to extensions
	delete(symbols, "os/os")
	delete(symbols, "io/fs/fs")
	delete(symbols, "os/exec/exec")
	delete(symbols, "os/signal/signal")
	delete(symbols, "os/user/user")
	delete(symbols, "os/signal/signal")
	delete(symbols, "io/ioutil/ioutil")
	delete(symbols, "runtime/runtime")
	delete(symbols, "syscall/syscall")
	delete(symbols, "archive/tar/tar")
	delete(symbols, "archive/zip/zip")
	delete(symbols, "compress/gzip/gzip")
	delete(symbols, "compress/zlib/zlib")

	if err := i.Use(symbols); err != nil {
		r.logger.Fatal().Err(err).Msg("extensions: Failed to load yaegi stdlib")
	}

	// Load the extension symbols
	err := i.Use(yaegi_interp.Symbols)
	if err != nil {
		r.logger.Fatal().Err(err).Msg("extensions: Failed to load extension symbols")
	}

	r.yaegiInterp = i
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Lists
// - Lists are used to display available options to the user based on the extensions installed
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) ListMangaProviderExtensions() []*MangaProviderExtensionItem {
	ret := make([]*MangaProviderExtensionItem, 0)

	r.mangaProviderExtensionBank.Range(func(key string, ext extension.MangaProviderExtension) bool {
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

	r.onlinestreamProviderExtensionBank.Range(func(key string, ext extension.OnlinestreamProviderExtension) bool {
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

	r.animeTorrentProviderExtensionBank.Range(func(key string, ext extension.AnimeTorrentProviderExtension) bool {
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
	ext, found := r.mangaProviderExtensionBank.Get(id)
	if found {
		return ext, true
	}

	ext, found = r.animeTorrentProviderExtensionBank.Get(id)
	if found {
		return ext, true
	}

	ext, found = r.onlinestreamProviderExtensionBank.Get(id)
	if found {
		return ext, true
	}

	return nil, false
}

func (r *Repository) GetMangaProviderExtensionBank() *extension.Bank[extension.MangaProviderExtension] {
	return r.mangaProviderExtensionBank
}

func (r *Repository) GetMangaProviderExtensionByID(id string) (extension.MangaProviderExtension, bool) {
	ext, found := r.mangaProviderExtensionBank.Get(id)
	return ext, found
}

func (r *Repository) GetOnlinestreamProviderExtensionBank() *extension.Bank[extension.OnlinestreamProviderExtension] {
	return r.onlinestreamProviderExtensionBank
}

func (r *Repository) GetOnlinestreamProviderExtensionByID(id string) (extension.OnlinestreamProviderExtension, bool) {
	ext, found := r.onlinestreamProviderExtensionBank.Get(id)
	return ext, found
}

func (r *Repository) GetAnimeTorrentProviderExtensionBank() *extension.Bank[extension.AnimeTorrentProviderExtension] {
	return r.animeTorrentProviderExtensionBank
}

func (r *Repository) GetAnimeTorrentProviderExtensionByID(id string) (extension.AnimeTorrentProviderExtension, bool) {
	ext, found := r.animeTorrentProviderExtensionBank.Get(id)
	return ext, found
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Built-in extensions
// - Built-in extensions are loaded once, on application startup
// - The "manifestURI" field is set to "builtin" to indicate that the extension is not external
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) LoadBuiltInMangaProviderExtension(info extension.Extension, provider hibikemanga.Provider) {
	r.mangaProviderExtensionBank.Set(info.ID, extension.NewMangaProviderExtension(&info, provider))
	r.logger.Debug().Str("id", info.ID).Msg("extensions: Loaded built-in manga provider extension")
}

func (r *Repository) LoadBuiltInAnimeTorrentProviderExtension(info extension.Extension, provider hibiketorrent.AnimeProvider) {
	r.animeTorrentProviderExtensionBank.Set(info.ID, extension.NewAnimeTorrentProviderExtension(&info, provider))
	r.logger.Debug().Str("id", info.ID).Msg("extensions: Loaded built-in anime torrent provider extension")
}

func (r *Repository) LoadBuiltInOnlinestreamProviderExtension(info extension.Extension, provider hibikeonlinestream.Provider) {
	r.onlinestreamProviderExtensionBank.Set(info.ID, extension.NewOnlinestreamProviderExtension(&info, provider))
	r.logger.Debug().Str("id", info.ID).Msg("extensions: Loaded built-in onlinestream provider extension")
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
