package extension_repo

import (
	hibikeonlinestream "github.com/5rahim/hibike/pkg/extension/onlinestream"
	"github.com/rs/zerolog"
	"seanime/internal/extension"
	"seanime/internal/util"
	"strings"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Online streaming
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) loadExternalOnlinestreamProviderExtension(ext *extension.Extension) {

	switch ext.Language {
	case extension.LanguageGo:
		r.loadExternalOnlinestreamProviderExtensionGo(ext)
	case extension.LanguageJavascript:
		// TODO
	}

	r.logger.Debug().Str("id", ext.ID).Msg("extension repo: Loaded online streaming provider extension")
}

//
// Go
//

func (r *Repository) loadExternalOnlinestreamProviderExtensionGo(ext *extension.Extension) {

	extensionPackageName := "ext_" + util.GenerateCryptoID()

	r.logger.Debug().Str("id", ext.ID).Str("packageName", extensionPackageName).Msg("extension repo: Loading online streaming provider")

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

	newProviderFunc, ok := newProviderFuncVal.Interface().(func(logger *zerolog.Logger) hibikeonlinestream.Provider)
	if !ok {
		r.logger.Error().Str("id", ext.ID).Msg(MsgYaegiFailedToInstantiateExtension)
		return
	}

	provider := newProviderFunc(r.logger)

	// Add the extension to the map
	r.onlinestreamProviderExtensions.Set(ext.ID, extension.NewOnlinestreamProviderExtension(ext, provider))
}
