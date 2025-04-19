package extension_repo

import (
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/manga/providers"
	"seanime/internal/onlinestream/providers"
	"seanime/internal/torrents/animetosho"
	"seanime/internal/torrents/nyaa"
	"seanime/internal/torrents/seadex"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"testing"
)

func GetMockExtensionRepository(t *testing.T) *Repository {
	logger := util.NewLogger()
	filecacher, _ := filecache.NewCacher(t.TempDir())
	extensionRepository := NewRepository(&NewRepositoryOptions{
		Logger:         logger,
		ExtensionDir:   t.TempDir(),
		WSEventManager: events.NewMockWSEventManager(logger),
		FileCacher:     filecacher,
	})

	extensionRepository.ReloadBuiltInExtension(extension.Extension{
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

	extensionRepository.ReloadBuiltInExtension(extension.Extension{
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

	extensionRepository.ReloadBuiltInExtension(extension.Extension{
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

	extensionRepository.ReloadBuiltInExtension(extension.Extension{
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

	extensionRepository.ReloadBuiltInExtension(extension.Extension{
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

	extensionRepository.ReloadBuiltInExtension(extension.Extension{
		ID:          "gogoanime",
		Name:        "Gogoanime",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeOnlinestreamProvider,
		Author:      "Seanime",
		Lang:        "en",
		Icon:        "https://raw.githubusercontent.com/5rahim/hibike/main/icons/gogoanime.png",
	}, onlinestream_providers.NewGogoanime(logger))

	extensionRepository.ReloadBuiltInExtension(extension.Extension{
		ID:          "zoro",
		Name:        "Hianime",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeOnlinestreamProvider,
		Author:      "Seanime",
		Lang:        "en",
		Icon:        "https://raw.githubusercontent.com/5rahim/hibike/main/icons/hianime.png",
	}, onlinestream_providers.NewZoro(logger))

	//
	// Built-in torrent providers
	//

	extensionRepository.ReloadBuiltInExtension(extension.Extension{
		ID:          "nyaa",
		Name:        "Nyaa",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeAnimeTorrentProvider,
		Author:      "Seanime",
		Lang:        "en",
		Icon:        "https://raw.githubusercontent.com/5rahim/hibike/main/icons/nyaa.png",
	}, nyaa.NewProvider(logger, nyaa.CategoryAnimeEng))

	extensionRepository.ReloadBuiltInExtension(extension.Extension{
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

	extensionRepository.ReloadBuiltInExtension(extension.Extension{
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

	extensionRepository.ReloadBuiltInExtension(extension.Extension{
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

	return extensionRepository
}
