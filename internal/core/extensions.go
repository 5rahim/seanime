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
	var consumers = []extension.Consumer{
		a.MangaRepository,
		a.OnlinestreamRepository,
		a.TorrentRepository,
		a.MediaPlayerRepository,
	}

	for _, consumer := range consumers {
		consumer.InitExtensionBank(a.ExtensionRepository.GetExtensionBank())
	}

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
		Icon:        "https://files.catbox.moe/wi5e0s.webp",
	}, manga_providers.NewComicK(a.Logger))

	a.ExtensionRepository.LoadBuiltInMangaProviderExtension(extension.Extension{
		ID:          "mangapill",
		Name:        "Mangapill",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeMangaProvider,
		Author:      "Seanime",
		Icon:        "https://files.catbox.moe/pwhp89.png",
	}, manga_providers.NewMangapill(a.Logger))

	a.ExtensionRepository.LoadBuiltInMangaProviderExtension(extension.Extension{
		ID:          "mangasee",
		Name:        "Mangasee",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeMangaProvider,
		Author:      "Seanime",
		Icon:        "https://files.catbox.moe/xerps6.png",
	}, manga_providers.NewMangasee(a.Logger))

	a.ExtensionRepository.LoadBuiltInMangaProviderExtension(extension.Extension{
		ID:          "mangadex",
		Name:        "Mangadex",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeMangaProvider,
		Author:      "Seanime",
		Icon:        "https://files.catbox.moe/cbn07p.png",
	}, manga_providers.NewMangadex(a.Logger))

	a.ExtensionRepository.LoadBuiltInMangaProviderExtension(extension.Extension{
		ID:          "manganato",
		Name:        "Manganato",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeMangaProvider,
		Author:      "Seanime",
		Icon:        "https://files.catbox.moe/sd8reg.png",
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
		Icon:        "https://files.catbox.moe/gzy8ip.png",
	}, onlinestream_providers.NewGogoanime(a.Logger))

	a.ExtensionRepository.LoadBuiltInOnlinestreamProviderExtension(extension.Extension{
		ID:          "zoro",
		Name:        "Hianime",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeOnlinestreamProvider,
		Author:      "Seanime",
		Icon:        "https://files.catbox.moe/ko0b5v.png",
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
		Icon:        "https://files.catbox.moe/dlrljx.png",
	}, nyaa.NewProvider(a.Logger))

	a.ExtensionRepository.LoadBuiltInAnimeTorrentProviderExtension(extension.Extension{
		ID:          "nyaa-sukebei",
		Name:        "Nyaa Sukebei",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeAnimeTorrentProvider,
		Author:      "Seanime",
		Icon:        "https://files.catbox.moe/dlrljx.png",
	}, nyaa.NewSukebeiProvider(a.Logger))

	a.ExtensionRepository.LoadBuiltInAnimeTorrentProviderExtension(extension.Extension{
		ID:          "animetosho",
		Name:        "AnimeTosho",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeAnimeTorrentProvider,
		Author:      "Seanime",
		Icon:        "https://files.catbox.moe/s506vk.jpg",
	}, animetosho.NewProvider(a.Logger))

	a.ExtensionRepository.LoadBuiltInAnimeTorrentProviderExtension(extension.Extension{
		ID:          "seadex",
		Name:        "SeaDex",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeAnimeTorrentProvider,
		Author:      "Seanime",
		Icon:        "https://files.catbox.moe/6fax26.png",
	}, seadex.NewProvider(a.Logger))

}

func (a *App) LoadOrRefreshExternalExtensions() {

	// Always called after loading built-in extensions
	a.ExtensionRepository.ReloadExternalExtensions()

}
