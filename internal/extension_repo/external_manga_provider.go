package extension_repo

import (
	"seanime/internal/extension"
	"seanime/internal/util"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Manga
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) loadExternalMangaExtension(ext *extension.Extension) (err error) {
	defer util.HandlePanicInModuleWithError("extension_repo/loadExternalMangaExtension", &err)

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

	return
}

func (r *Repository) loadExternalMangaExtensionGo(ext *extension.Extension) error {

	provider, err := NewYaegiMangaProvider(r.yaegiInterp, ext, r.logger)
	if err != nil {
		return err
	}

	// Add the extension to the map
	retExt := extension.NewMangaProviderExtension(ext, provider)
	r.extensionBank.Set(ext.ID, retExt)
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
	retExt := extension.NewMangaProviderExtension(ext, provider)
	r.extensionBank.Set(ext.ID, retExt)
	return nil
}
