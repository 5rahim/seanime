package extension_repo

import (
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"os"
	"seanime/internal/extension"
	vendor_hibike_torrent "seanime/internal/extension/vendoring/torrent"
	"seanime/internal/yaegi_interp"

	hibikemanga "github.com/5rahim/hibike/pkg/extension/manga"
	hibikeonlinestream "github.com/5rahim/hibike/pkg/extension/onlinestream"
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
)

type (
	// Repository manages all extensions
	Repository struct {
		logger *zerolog.Logger
		// Absolute path to the directory containing all extensions
		extensionDir string
		// Yaegi interpreter for Go extensions
		yaegiInterp *interp.Interpreter
		// Extension banks
		mangaProviderExtensionBank        *extension.Bank[extension.MangaProviderExtension]
		animeTorrentProviderExtensionBank *extension.Bank[extension.AnimeTorrentProviderExtension]
		onlinestreamProviderExtensionBank *extension.Bank[extension.OnlinestreamProviderExtension]
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
	Logger       *zerolog.Logger
	ExtensionDir string
}

func NewRepository(opts *NewRepositoryOptions) *Repository {
	// Load the extension
	i := interp.New(interp.Options{})
	if err := i.Use(stdlib.Symbols); err != nil {
		opts.Logger.Fatal().Err(err).Msg("extensions: Failed to load yaegi stdlib")
	}

	// Load the extension symbols
	err := i.Use(yaegi_interp.Symbols)
	if err != nil {
		opts.Logger.Fatal().Err(err).Msg("extensions: Failed to load extension symbols")
	}

	// Make sure the extension directory exists
	_ = os.MkdirAll(opts.ExtensionDir, os.ModePerm)

	ret := &Repository{
		yaegiInterp:                       i,
		logger:                            opts.Logger,
		extensionDir:                      opts.ExtensionDir,
		mangaProviderExtensionBank:        extension.NewBank[extension.MangaProviderExtension](),
		animeTorrentProviderExtensionBank: extension.NewBank[extension.AnimeTorrentProviderExtension](),
		onlinestreamProviderExtensionBank: extension.NewBank[extension.OnlinestreamProviderExtension](),
	}

	return ret
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
// External extensions
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) InstallExternalExtension(repositoryURI string) {

	// 1. Get the json from the URI
	// 2. Parse the json
	// 3. Check if the extension is already installed
	// 4. If not, install the extension | If yes, update the extension
	// 5. Load the extension

}

// CheckForUpdates checks all extensions for updates by querying their respective repositories
func (r *Repository) CheckForUpdates() {

}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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
