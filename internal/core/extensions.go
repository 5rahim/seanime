package core

import (
	"seanime/internal/extension"
	"seanime/internal/extension_repo"
	manga_providers "seanime/internal/manga/providers"
	onlinestream_providers "seanime/internal/onlinestream/providers"
	"seanime/internal/torrents/animetosho"
	"seanime/internal/torrents/nyaa"
	"seanime/internal/torrents/seadex"

	"github.com/rs/zerolog"
)

func LoadExtensions(extensionRepository *extension_repo.Repository, logger *zerolog.Logger) {

	//
	// Built-in manga providers
	//

	extensionRepository.LoadBuiltInMangaProviderExtension(extension.Extension{
		ID:          "comick",
		Name:        "ComicK",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeMangaProvider,
		Author:      "Seanime",
		Description: "",
		Lang:        "en",
		Icon:        "https://raw.githubusercontent.com/5rahim/hibike/main/icons/comick.webp",
	}, manga_providers.NewComicK(logger))

	extensionRepository.LoadBuiltInMangaProviderExtension(extension.Extension{
		ID:          "comick-multi",
		Name:        "ComicK (Multi)",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeMangaProvider,
		Author:      "Seanime",
		Description: "",
		Lang:        "multi",
		Icon:        "https://raw.githubusercontent.com/5rahim/hibike/main/icons/comick.webp",
	}, manga_providers.NewComicKMulti(logger))

	extensionRepository.LoadBuiltInMangaProviderExtension(extension.Extension{
		ID:          "mangapill",
		Name:        "Mangapill",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeMangaProvider,
		Author:      "Seanime",
		Lang:        "en",
		Icon:        "https://raw.githubusercontent.com/5rahim/hibike/main/icons/mangapill.png",
	}, manga_providers.NewMangapill(logger))

	extensionRepository.LoadBuiltInMangaProviderExtension(extension.Extension{
		ID:          "weebcentral",
		Name:        "WeebCentral",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeMangaProvider,
		Author:      "Seanime",
		Lang:        "en",
		Icon:        "https://raw.githubusercontent.com/5rahim/hibike/main/icons/weebcentral.png",
	}, manga_providers.NewWeebCentral(logger))

	extensionRepository.LoadBuiltInMangaProviderExtension(extension.Extension{
		ID:          "mangadex",
		Name:        "Mangadex",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeMangaProvider,
		Author:      "Seanime",
		Lang:        "en",
		Icon:        "https://raw.githubusercontent.com/5rahim/hibike/main/icons/mangadex.png",
	}, manga_providers.NewMangadex(logger))

	extensionRepository.LoadBuiltInMangaProviderExtension(extension.Extension{
		ID:          "manganato",
		Name:        "Manganato",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeMangaProvider,
		Author:      "Seanime",
		Lang:        "en",
		Icon:        "https://raw.githubusercontent.com/5rahim/hibike/main/icons/manganato.png",
	}, manga_providers.NewManganato(logger))

	//
	// Built-in online stream providers
	//

	//extensionRepository.LoadBuiltInOnlinestreamProviderExtension(extension.Extension{
	//	ID:          "gogoanime",
	//	Name:        "Gogoanime",
	//	Version:     "",
	//	ManifestURI: "builtin",
	//	Language:    extension.LanguageGo,
	//	Type:        extension.TypeOnlinestreamProvider,
	//	Author:      "Seanime",
	//	Lang:        "en",
	//	Icon:        "https://raw.githubusercontent.com/5rahim/hibike/main/icons/gogoanime.png",
	//}, onlinestream_providers.NewGogoanime(logger))

	//extensionRepository.LoadBuiltInOnlinestreamProviderExtension(extension.Extension{
	//	ID:          "zoro",
	//	Name:        "Hianime",
	//	Version:     "",
	//	ManifestURI: "builtin",
	//	Language:    extension.LanguageGo,
	//	Type:        extension.TypeOnlinestreamProvider,
	//	Author:      "Seanime",
	//	Lang:        "en",
	//	Icon:        "https://raw.githubusercontent.com/5rahim/hibike/main/icons/hianime.png",
	//}, onlinestream_providers.NewZoro(logger))

	extensionRepository.LoadBuiltInOnlinestreamProviderExtensionJS(extension.Extension{
		ID:          "animepahe",
		Name:        "Animepahe",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageTypescript,
		Type:        extension.TypeOnlinestreamProvider,
		Author:      "Seanime",
		Lang:        "en",
		Icon:        "https://raw.githubusercontent.com/5rahim/hibike/main/icons/animepahe.png",
		Payload:     onlinestream_providers.AnimepahePayload,
	})

	//
	// Built-in torrent providers
	//

	extensionRepository.LoadBuiltInAnimeTorrentProviderExtension(extension.Extension{
		ID:          "nyaa",
		Name:        "Nyaa",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeAnimeTorrentProvider,
		Author:      "Seanime",
		Lang:        "en",
		Icon:        "https://raw.githubusercontent.com/5rahim/hibike/main/icons/nyaa.png",
	}, nyaa.NewProvider(logger))

	extensionRepository.LoadBuiltInAnimeTorrentProviderExtension(extension.Extension{
		ID:          "nyaa-sukebei",
		Name:        "Nyaa Sukebei",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeAnimeTorrentProvider,
		Author:      "Seanime",
		Lang:        "en",
		Icon:        "https://raw.githubusercontent.com/5rahim/hibike/main/icons/nyaa.png",
	}, nyaa.NewSukebeiProvider(logger))

	extensionRepository.LoadBuiltInAnimeTorrentProviderExtension(extension.Extension{
		ID:          "animetosho",
		Name:        "AnimeTosho",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeAnimeTorrentProvider,
		Author:      "Seanime",
		Lang:        "en",
		Icon:        "https://raw.githubusercontent.com/5rahim/hibike/main/icons/animetosho.png",
	}, animetosho.NewProvider(logger))

	extensionRepository.LoadBuiltInAnimeTorrentProviderExtension(extension.Extension{
		ID:          "seadex",
		Name:        "SeaDex",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeAnimeTorrentProvider,
		Author:      "Seanime",
		Lang:        "en",
		Icon:        "https://raw.githubusercontent.com/5rahim/hibike/main/icons/seadex.png",
	}, seadex.NewProvider(logger))

	extensionRepository.ReloadExternalExtensions()
}

func (a *App) AddExtensionBankToConsumers() {

	var consumers = []extension.Consumer{
		a.MangaRepository,
		a.OnlinestreamRepository,
		a.TorrentRepository,
	}

	for _, consumer := range consumers {
		consumer.InitExtensionBank(a.ExtensionRepository.GetExtensionBank())
	}
}
