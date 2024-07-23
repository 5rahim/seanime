package core

import (
	"seanime/internal/extension"
	"seanime/internal/manga/providers"
)

func (a *App) LoadBuiltInExtensions() {

	// Load the extensions to the manga repository
	a.MangaRepository.SetProviderExtensions(a.ExtensionRepository.GetMangaExtensions())

	//
	// Built-in manga providers
	//

	a.ExtensionRepository.LoadBuiltInMangaExtension(extension.Extension{
		ID:            "comick",
		Name:          "ComicK",
		Version:       "1.0.0",
		RepositoryURI: "",
		Language:      extension.LanguageGo,
		Type:          extension.TypeMangaProvider,
		Author:        "Seanime",
		Description:   "",
	}, manga_providers.NewComicK(a.Logger))

	a.ExtensionRepository.LoadBuiltInMangaExtension(extension.Extension{
		ID:       "mangapill",
		Name:     "Mangapill",
		Version:  "1.0.0",
		Language: extension.LanguageGo,
		Type:     extension.TypeMangaProvider,
		Author:   "Seanime",
	}, manga_providers.NewMangapill(a.Logger))

	a.ExtensionRepository.LoadBuiltInMangaExtension(extension.Extension{
		ID:       "mangasee",
		Name:     "Mangasee",
		Version:  "1.0.0",
		Language: extension.LanguageGo,
		Type:     extension.TypeMangaProvider,
		Author:   "Seanime",
	}, manga_providers.NewMangasee(a.Logger))

	a.ExtensionRepository.LoadBuiltInMangaExtension(extension.Extension{
		ID:       "mangadex",
		Name:     "Mangadex",
		Version:  "1.0.0",
		Language: extension.LanguageGo,
		Type:     extension.TypeMangaProvider,
		Author:   "Seanime",
	}, manga_providers.NewMangadex(a.Logger))

	a.ExtensionRepository.LoadBuiltInMangaExtension(extension.Extension{
		ID:       "manganato",
		Name:     "Manganato",
		Version:  "1.0.0",
		Language: extension.LanguageGo,
		Type:     extension.TypeMangaProvider,
		Author:   "Seanime",
	}, manga_providers.NewManganato(a.Logger))

}

func (a *App) LoadOrRefreshExternalExtensions() {

	// Always called after loading built-in extensions
	a.ExtensionRepository.LoadExternalExtensions()

}
