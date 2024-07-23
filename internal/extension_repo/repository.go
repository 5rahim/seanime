package extension_repo

import (
	hibikemanga "github.com/5rahim/hibike/pkg/extension/manga"
	hibikeonlinestream "github.com/5rahim/hibike/pkg/extension/onlinestream"
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
	"github.com/rs/zerolog"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"os"
	"seanime/internal/extension"
	"seanime/internal/util/result"
	"seanime/internal/yaegi_interp"
)

type (
	// Repository manages all extensions
	Repository struct {
		logger *zerolog.Logger
		// Absolute path to the directory containing all extensions
		extensionDir string
		// Yaegi interpreter for Go extensions
		yaegiInterp *interp.Interpreter
		// Map of manga provider extensions
		mangaProviderExtensions *result.Map[string, extension.MangaProviderExtension]
		// Map of torrent provider extensions
		torrentProviderExtensions *result.Map[string, extension.TorrentProviderExtension]
		// Map of online stream provider extensions
		onlinestreamProviderExtensions *result.Map[string, extension.OnlinestreamProviderExtension]
	}

	MangaProviderExtensionItem struct {
		ID   string `json:"id"`
		Name string `json:"name"`
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
		opts.Logger.Fatal().Err(err).Msg("extension repo: Failed to load yaegi stdlib")
	}

	// Load the extension symbols
	err := i.Use(yaegi_interp.Symbols)
	if err != nil {
		opts.Logger.Fatal().Err(err).Msg("extension repo: Failed to load extension symbols")
	}

	// Make sure the extension directory exists
	_ = os.MkdirAll(opts.ExtensionDir, os.ModePerm)

	ret := &Repository{
		yaegiInterp:                    i,
		logger:                         opts.Logger,
		extensionDir:                   opts.ExtensionDir,
		mangaProviderExtensions:        result.NewResultMap[string, extension.MangaProviderExtension](),
		torrentProviderExtensions:      result.NewResultMap[string, extension.TorrentProviderExtension](),
		onlinestreamProviderExtensions: result.NewResultMap[string, extension.OnlinestreamProviderExtension](),
	}

	return ret
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Lists
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) ListMangaProviderExtensions() []*MangaProviderExtensionItem {
	ret := make([]*MangaProviderExtensionItem, 0)

	r.mangaProviderExtensions.Range(func(key string, ext extension.MangaProviderExtension) bool {
		ret = append(ret, &MangaProviderExtensionItem{
			ID:   ext.GetID(),
			Name: ext.GetName(),
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

func (r *Repository) GetMangaExtensions() *result.Map[string, extension.MangaProviderExtension] {
	return r.mangaProviderExtensions
}

func (r *Repository) GetMangaExtensionByID(id string) (extension.MangaProviderExtension, bool) {
	ext, found := r.mangaProviderExtensions.Get(id)
	return ext, found
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Built-in extensions
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) LoadBuiltInMangaExtension(info extension.Extension, provider hibikemanga.Provider) {
	r.mangaProviderExtensions.Set(info.ID, extension.NewMangaProviderExtension(&info, provider))
}

func (r *Repository) LoadBuiltInTorrentExtension(info extension.Extension, provider hibiketorrent.Provider) {
	r.torrentProviderExtensions.Set(info.ID, extension.NewTorrentProviderExtension(&info, provider))
}

func (r *Repository) LoadBuiltInOnlinestreamExtension(info extension.Extension, provider hibikeonlinestream.Provider) {
	r.onlinestreamProviderExtensions.Set(info.ID, extension.NewOnlinestreamProviderExtension(&info, provider))
}
