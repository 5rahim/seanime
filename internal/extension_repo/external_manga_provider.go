package extension_repo

import (
	"seanime/internal/extension"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Manga
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) loadExternalMangaExtension(ext *extension.Extension) (err error) {

	switch ext.Language {
	case extension.LanguageGo:
		err = r.loadExternalMangaExtensionGo(ext)
	case extension.LanguageJavascript:
		err = r.loadExternalMangaExtensionJS(ext, extension.LanguageJavascript)
	case extension.LanguageTypescript:
		err = r.loadExternalMangaExtensionJS(ext, extension.LanguageTypescript)
	}

	if err != nil {
		return
	}

	//r.logger.Debug().Str("id", ext.ID).Msg("extensions: Loaded manga provider extension")
	return
}

func (r *Repository) loadExternalMangaExtensionGo(ext *extension.Extension) error {

	provider, err := NewYaegiMangaProvider(r.yaegiInterp, ext, r.logger)
	if err != nil {
		return err
	}

	// Add the extension to the map
	r.mangaProviderExtensionBank.Set(ext.ID, extension.NewMangaProviderExtension(ext, provider))
	return nil
}

func (r *Repository) loadExternalMangaExtensionJS(ext *extension.Extension, language extension.Language) error {

	provider, gojaExt, err := NewGojaMangaProvider(ext, language, r.logger)
	if err != nil {
		return err
	}

	// Add the goja extension pointer to the map
	r.gojaExtensions.Set(ext.ID, gojaExt)

	// Add the extension to the map
	r.mangaProviderExtensionBank.Set(ext.ID, extension.NewMangaProviderExtension(ext, provider))
	return nil
}
