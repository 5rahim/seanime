package extension_repo

import (
	"fmt"
	"seanime/internal/extension"
	"seanime/internal/util"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Online streaming
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) loadExternalOnlinestreamProviderExtension(ext *extension.Extension) (err error) {
	defer util.HandlePanicInModuleWithError("extension_repo/loadExternalOnlinestreamProviderExtension", &err)

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

	return
}

func (r *Repository) loadExternalOnlinestreamProviderExtensionGo(ext *extension.Extension) error {

	provider, err := NewYaegiOnlinestreamProvider(r.yaegiInterp, ext, r.logger)
	if err != nil {
		return err
	}

	// Add the extension to the map
	retExt := extension.NewOnlinestreamProviderExtension(ext, provider)
	r.extensionBank.Set(ext.ID, retExt)
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
	retExt := extension.NewOnlinestreamProviderExtension(ext, provider)
	r.extensionBank.Set(ext.ID, retExt)
	return nil
}
