package extension_repo

import (
	"fmt"
	"seanime/internal/extension"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Online streaming
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) loadExternalOnlinestreamProviderExtension(ext *extension.Extension) (err error) {

	switch ext.Language {
	case extension.LanguageGo:
		err = r.loadExternalOnlinestreamProviderExtensionGo(ext)
	case extension.LanguageJavascript:
		err = r.loadExternalOnlinestreamExtensionJS(ext, extension.LanguageJavascript)
	case extension.LanguageTypescript:
		err = r.loadExternalOnlinestreamExtensionJS(ext, extension.LanguageTypescript)
	default:
		err = fmt.Errorf("unsupported language: %v", ext.Language)
	}

	if err != nil {
		return
	}

	//r.logger.Debug().Str("id", ext.ID).Msg("extensions: Loaded online streaming provider extension")
	return
}

func (r *Repository) loadExternalOnlinestreamProviderExtensionGo(ext *extension.Extension) error {

	provider, err := NewYaegiOnlinestreamProvider(r.yaegiInterp, ext, r.logger)
	if err != nil {
		return err
	}

	// Add the extension to the map
	r.onlinestreamProviderExtensionBank.Set(ext.ID, extension.NewOnlinestreamProviderExtension(ext, provider))
	return nil
}

func (r *Repository) loadExternalOnlinestreamExtensionJS(ext *extension.Extension, language extension.Language) error {

	provider, gojaExt, err := NewGojaOnlinestreamProvider(ext, language, r.logger)
	if err != nil {
		return err
	}

	// Add the goja extension pointer to the map
	r.gojaExtensions.Set(ext.ID, gojaExt)

	// Add the extension to the map
	r.onlinestreamProviderExtensionBank.Set(ext.ID, extension.NewOnlinestreamProviderExtension(ext, provider))
	return nil
}
