package core

import (
	"seanime/internal/extension"
	"seanime/internal/extension_repo"
	manga_providers "seanime/internal/manga/providers"

	"github.com/rs/zerolog"
)

func (a *App) AddExtensionBankToConsumers() {

	var consumers = []extension.Consumer{
		a.MangaRepository,
		a.OnlinestreamRepository,
		a.TorrentRepository,
		a.AnilistPlatform,
		a.MetadataProvider,
	}

	for _, consumer := range consumers {
		consumer.InitExtensionBank(a.ExtensionRepository.GetExtensionBank())
	}
}

func LoadExtensions(extensionRepository *extension_repo.Repository, logger *zerolog.Logger, config *Config) {

	//
	// Built-in manga providers
	//

	extensionRepository.ReloadBuiltInExtension(extension.Extension{
		ID:          manga_providers.LocalProvider,
		Name:        "Local",
		Version:     "",
		ManifestURI: "builtin",
		Language:    extension.LanguageGo,
		Type:        extension.TypeMangaProvider,
		Author:      "Seanime",
		Lang:        "multi",
		Icon:        "https://raw.githubusercontent.com/5rahim/hibike/main/icons/local-manga.png",
	}, manga_providers.NewLocal(config.Manga.LocalDir, logger))

	extensionRepository.ReloadExternalExtensions()
}
