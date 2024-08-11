package extension_repo

import (
	"fmt"
	"seanime/internal/extension"
	"seanime/internal/util"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Media player
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) loadExternalMediaPlayerExtension(ext *extension.Extension) (err error) {
	defer util.HandlePanicInModuleWithError("extension_repo/loadExternalMediaPlayerExtension", &err)

	// Check if the extension ID is not already in use by built-in code
	switch ext.ID {
	case "mpv", "vlc", "mpc-hc":
		err = fmt.Errorf("extension ID '%s' is a reserved ID", ext.ID)
		return
	default:
	}

	switch ext.Language {
	case extension.LanguageGo:
		err = r.loadExternalMediaPlayerExtensionGo(ext)
	case extension.LanguageJavascript:
		err = r.loadExternalMediaPlayerExtensionJS(ext, extension.LanguageJavascript)
	case extension.LanguageTypescript:
		err = r.loadExternalMediaPlayerExtensionJS(ext, extension.LanguageTypescript)
	}

	if err != nil {
		return
	}

	return
}

func (r *Repository) loadExternalMediaPlayerExtensionGo(ext *extension.Extension) error {

	provider, err := NewYaegiMediaPlayer(r.yaegiInterp, ext, r.logger)
	if err != nil {
		return err
	}

	// Add the extension to the map
	retExt := extension.NewMediaPlayerExtension(ext, provider)
	r.extensionBank.Set(ext.ID, retExt)
	return nil
}

func (r *Repository) loadExternalMediaPlayerExtensionJS(ext *extension.Extension, language extension.Language) error {

	provider, gojaExt, err := NewGojaMediaPlayer(ext, language, r.logger)
	if err != nil {
		return err
	}

	// Add the goja extension pointer to the map
	r.gojaExtensions.Set(ext.ID, gojaExt)

	// Add the extension to the map
	retExt := extension.NewMediaPlayerExtension(ext, provider)
	r.extensionBank.Set(ext.ID, retExt)
	return nil
}
