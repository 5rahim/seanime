package extension_repo

import (
	hibikemanga "github.com/5rahim/hibike/pkg/extension/manga"
	"github.com/rs/zerolog"
	"seanime/internal/extension"
	"seanime/internal/util"
	"strings"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Manga
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) loadExternalMangaExtension(ext *extension.Extension) {

	switch ext.Language {
	case extension.LanguageGo:
		r.loadExternalMangaExtensionGo(ext)
	case extension.LanguageJavascript:
		// TODO
	}

	r.logger.Debug().Str("id", ext.ID).Msg("extensions: Loaded manga provider extension")
}

//
// Go
//

func (r *Repository) loadExternalMangaExtensionGo(ext *extension.Extension) {

	extensionPackageName := "ext_" + util.GenerateCryptoID()

	r.logger.Trace().Str("id", ext.ID).Str("packageName", extensionPackageName).Msg("extensions: Loading external manga provider")

	payload := strings.Replace(ext.Payload, "package main", "package "+extensionPackageName, 1)

	// Load the extension payload
	_, err := r.yaegiEval(payload)
	if err != nil {
		r.logger.Error().Err(err).Str("id", ext.ID).Msg(MsgYaegiFailedToEvaluateExtensionCode)
		return
	}

	// Get the provider
	newProviderFuncVal, err := r.yaegiEval(extensionPackageName + `.NewProvider`)
	if err != nil {
		r.logger.Error().Err(err).Str("id", ext.ID).Msg(MsgYaegiFailedToEvaluateExtensionCode)
		return
	}

	newProviderFunc, ok := newProviderFuncVal.Interface().(func(logger *zerolog.Logger) hibikemanga.Provider)
	if !ok {
		r.logger.Error().Str("id", ext.ID).Msg(MsgYaegiFailedToInstantiateExtension)
		return
	}

	provider := newProviderFunc(r.logger)

	// Add the extension to the map
	r.mangaProviderExtensionBank.Set(ext.ID, extension.NewMangaProviderExtension(ext, provider))
}
