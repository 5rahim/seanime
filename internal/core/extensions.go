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
	a.MangaRepository.InitProviderExtensionBank(a.ExtensionRepository.GetMangaProviderExtensionBank())
	// Load the extensions to the online stream repository
	a.OnlinestreamRepository.InitProviderExtensionBank(a.ExtensionRepository.GetOnlinestreamProviderExtensionBank())
	// Load the extensions to the torrent repository
	a.TorrentRepository.InitAnimeProviderExtensionBank(a.ExtensionRepository.GetAnimeTorrentProviderExtensionBank())

	//
	// Built-in manga providers
	//

	a.ExtensionRepository.LoadBuiltInMangaProviderExtension(extension.Extension{
		ID:          "comick",
		Name:        "ComicK",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeMangaProvider,
		Author:      "Seanime",
		Description: "",
	}, manga_providers.NewComicK(a.Logger))

	a.ExtensionRepository.LoadBuiltInMangaProviderExtension(extension.Extension{
		ID:          "mangapill",
		Name:        "Mangapill",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeMangaProvider,
		Author:      "Seanime",
	}, manga_providers.NewMangapill(a.Logger))

	a.ExtensionRepository.LoadBuiltInMangaProviderExtension(extension.Extension{
		ID:          "mangasee",
		Name:        "Mangasee",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeMangaProvider,
		Author:      "Seanime",
	}, manga_providers.NewMangasee(a.Logger))

	a.ExtensionRepository.LoadBuiltInMangaProviderExtension(extension.Extension{
		ID:          "mangadex",
		Name:        "Mangadex",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeMangaProvider,
		Author:      "Seanime",
	}, manga_providers.NewMangadex(a.Logger))

	a.ExtensionRepository.LoadBuiltInMangaProviderExtension(extension.Extension{
		ID:          "manganato",
		Name:        "Manganato",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeMangaProvider,
		Author:      "Seanime",
	}, manga_providers.NewManganato(a.Logger))

	//
	// Built-in online stream providers
	//

	a.ExtensionRepository.LoadBuiltInOnlinestreamProviderExtension(extension.Extension{
		ID:          "gogoanime",
		Name:        "Gogoanime",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeOnlinestreamProvider,
		Author:      "Seanime",
	}, onlinestream_providers.NewGogoanime(a.Logger))

	a.ExtensionRepository.LoadBuiltInOnlinestreamProviderExtension(extension.Extension{
		ID:          "zoro",
		Name:        "Hianime",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeOnlinestreamProvider,
		Author:      "Seanime",
	}, onlinestream_providers.NewZoro(a.Logger))

	//
	// Built-in torrent providers
	//

	a.ExtensionRepository.LoadBuiltInAnimeTorrentProviderExtension(extension.Extension{
		ID:          "nyaa",
		Name:        "Nyaa",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeAnimeTorrentProvider,
		Author:      "Seanime",
		Meta: extension.Meta{
			Icon: "https://files.catbox.moe/dlrljx.png",
		},
	}, nyaa.NewProvider(a.Logger))

	a.ExtensionRepository.LoadBuiltInAnimeTorrentProviderExtension(extension.Extension{
		ID:          "nyaa-sukebei",
		Name:        "Nyaa Sukebei",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeAnimeTorrentProvider,
		Author:      "Seanime",
		Meta: extension.Meta{
			Icon: "https://files.catbox.moe/dlrljx.png",
		},
	}, nyaa.NewSukebeiProvider(a.Logger))

	a.ExtensionRepository.LoadBuiltInAnimeTorrentProviderExtension(extension.Extension{
		ID:          "animetosho",
		Name:        "AnimeTosho",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeAnimeTorrentProvider,
		Author:      "Seanime",
		Meta:        extension.Meta{
			//Icon: "https://files.catbox.moe/xf9jl6.ico",
		},
	}, animetosho.NewProvider(a.Logger))

	a.ExtensionRepository.LoadBuiltInAnimeTorrentProviderExtension(extension.Extension{
		ID:          "seadex",
		Name:        "SeaDex",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeAnimeTorrentProvider,
		Author:      "Seanime",
		Meta: extension.Meta{
			Icon: "https://files.catbox.moe/6fax26.png",
		},
	}, seadex.NewProvider(a.Logger))

}

func (a *App) LoadOrRefreshExternalExtensions() {

	// Always called after loading built-in extensions
	a.ExtensionRepository.ReloadExternalExtensions()

}
