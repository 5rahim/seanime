package extension_repo

import (
	"fmt"
	"math/rand"
	"seanime/internal/extension"
	"seanime/internal/util"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Online streaming
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) loadExternalCustomSourceProviderExtension(ext *extension.Extension) (err error) {
	defer util.HandlePanicInModuleWithError("extension_repo/loadExternalCustomSourceProviderExtension", &err)

	switch ext.Language {
	case extension.LanguageJavascript, extension.LanguageTypescript:
		err = r.loadExternalCustomSourceExtensionJS(ext, ext.Language)
	default:
		err = fmt.Errorf("unsupported language: %v", ext.Language)
	}

	if err != nil {
		return
	}

	return
}

// generateExtensionIdentifier generates a unique extension identifier for a custom source provider extension
// it ensures that the extension identifier is unique across all custom source provider extensions
func (r *Repository) generateExtensionIdentifier() int {
	customSourceProviderExtensions := r.ListCustomSourceExtensions()

	//return rand.Intn(65535) + 1

	identifier := 1
	for {
		found := false
		for _, ext := range customSourceProviderExtensions {
			if ext.ExtensionIdentifier == identifier {
				found = true
				break
			}
		}
		if !found {
			return identifier
		}
		identifier++

		if identifier > 65535 {
			return rand.Intn(65535) + 1
		}
	}
}

func (r *Repository) loadExternalCustomSourceExtensionJS(ext *extension.Extension, language extension.Language) error {
	provider, gojaExt, err := NewGojaCustomSource(ext, language, r.logger, r.gojaRuntimeManager)
	if err != nil {
		return err
	}

	// Add the extension to the map
	retExt := extension.NewCustomSourceExtension(ext, provider)
	retExt.SetExtensionIdentifier(r.generateExtensionIdentifier())
	gojaExt.extensionIdentifier = retExt.GetExtensionIdentifier()
	r.extensionBank.Set(ext.ID, retExt)
	r.gojaExtensions.Set(ext.ID, gojaExt)

	r.logger.Trace().Str("id", ext.ID).Int("identifier", gojaExt.extensionIdentifier).Msg("extensions: Loaded external custom source extension")

	return nil
}
