package core

import (
	"seanime/internal/extension"
	"seanime/internal/manga/providers"
	"seanime/internal/onlinestream/providers"
	"seanime/internal/torrents/animetosho"
	"seanime/internal/torrents/nyaa"
	"seanime/internal/torrents/seadex"
)

func (a *App) LoadBuiltInExtensions() {

	// Load the extensions to the manga repository
	a.MangaRepository.SetProviderExtensions(a.ExtensionRepository.GetMangaProviderExtensions())
	// Load the extensions to the online stream repository
	a.OnlinestreamRepository.SetProviderExtensions(a.ExtensionRepository.GetOnlinestreamProviderExtensions())
	// Load the extensions to the torrent repository
	a.TorrentRepository.SetAnimeProviderExtensions(a.ExtensionRepository.GetAnimeTorrentProviderExtensions())

	//
	// Built-in manga providers
	//

	a.ExtensionRepository.LoadBuiltInMangaProviderExtension(extension.Extension{
		ID:            "comick",
		Name:          "ComicK",
		Version:       "1.0.0",
		RepositoryURI: "",
		Language:      extension.LanguageGo,
		Type:          extension.TypeMangaProvider,
		Author:        "Seanime",
		Description:   "",
	}, manga_providers.NewComicK(a.Logger))

	a.ExtensionRepository.LoadBuiltInMangaProviderExtension(extension.Extension{
		ID:       "mangapill",
		Name:     "Mangapill",
		Version:  "1.0.0",
		Language: extension.LanguageGo,
		Type:     extension.TypeMangaProvider,
		Author:   "Seanime",
	}, manga_providers.NewMangapill(a.Logger))

	a.ExtensionRepository.LoadBuiltInMangaProviderExtension(extension.Extension{
		ID:       "mangasee",
		Name:     "Mangasee",
		Version:  "1.0.0",
		Language: extension.LanguageGo,
		Type:     extension.TypeMangaProvider,
		Author:   "Seanime",
	}, manga_providers.NewMangasee(a.Logger))

	a.ExtensionRepository.LoadBuiltInMangaProviderExtension(extension.Extension{
		ID:       "mangadex",
		Name:     "Mangadex",
		Version:  "1.0.0",
		Language: extension.LanguageGo,
		Type:     extension.TypeMangaProvider,
		Author:   "Seanime",
	}, manga_providers.NewMangadex(a.Logger))

	a.ExtensionRepository.LoadBuiltInMangaProviderExtension(extension.Extension{
		ID:       "manganato",
		Name:     "Manganato",
		Version:  "1.0.0",
		Language: extension.LanguageGo,
		Type:     extension.TypeMangaProvider,
		Author:   "Seanime",
	}, manga_providers.NewManganato(a.Logger))

	//
	// Built-in online stream providers
	//

	a.ExtensionRepository.LoadBuiltInOnlinestreamProviderExtension(extension.Extension{
		ID:       "gogoanime",
		Name:     "Gogoanime",
		Version:  "1.0.0",
		Language: extension.LanguageGo,
		Type:     extension.TypeOnlinestreamProvider,
		Author:   "Seanime",
	}, onlinestream_providers.NewGogoanime(a.Logger))

	a.ExtensionRepository.LoadBuiltInOnlinestreamProviderExtension(extension.Extension{
		ID:       "zoro",
		Name:     "Hianime",
		Version:  "1.0.0",
		Language: extension.LanguageGo,
		Type:     extension.TypeOnlinestreamProvider,
		Author:   "Seanime",
	}, onlinestream_providers.NewZoro(a.Logger))

	//
	// Built-in torrent providers
	//

	a.ExtensionRepository.LoadBuiltInAnimeTorrentProviderExtension(extension.Extension{
		ID:       "nyaa",
		Name:     "Nyaa",
		Version:  "1.0.0",
		Language: extension.LanguageGo,
		Type:     extension.TypeTorrentProvider,
		Author:   "Seanime",
	}, nyaa.NewProvider(a.Logger))

	a.ExtensionRepository.LoadBuiltInAnimeTorrentProviderExtension(extension.Extension{
		ID:       "nyaa-sukebei",
		Name:     "Nyaa Sukebei",
		Version:  "1.0.0",
		Language: extension.LanguageGo,
		Type:     extension.TypeTorrentProvider,
		Author:   "Seanime",
	}, nyaa.NewSukebeiProvider(a.Logger))

	a.ExtensionRepository.LoadBuiltInAnimeTorrentProviderExtension(extension.Extension{
		ID:       "animetosho",
		Name:     "AnimeTosho",
		Version:  "1.0.0",
		Language: extension.LanguageGo,
		Type:     extension.TypeTorrentProvider,
		Author:   "Seanime",
	}, animetosho.NewProvider(a.Logger))

	a.ExtensionRepository.LoadBuiltInAnimeTorrentProviderExtension(extension.Extension{
		ID:       "seadex",
		Name:     "SeaDex",
		Version:  "1.0.0",
		Language: extension.LanguageGo,
		Type:     extension.TypeTorrentProvider,
		Author:   "Seanime",
	}, seadex.NewProvider(a.Logger))

}

func (a *App) LoadOrRefreshExternalExtensions() {

	// Always called after loading built-in extensions
	a.ExtensionRepository.LoadExternalExtensions()

}
