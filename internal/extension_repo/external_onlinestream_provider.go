package extension_repo

import (
	"fmt"
	hibikeonlinestream "github.com/5rahim/hibike/pkg/extension/onlinestream"
	"github.com/rs/zerolog"
	"seanime/internal/extension"
	"seanime/internal/util"
	"strings"
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

//
// Go
//

func (r *Repository) loadExternalOnlinestreamProviderExtensionGo(ext *extension.Extension) error {

	extensionPackageName := "ext_" + util.GenerateCryptoID()

	r.logger.Trace().Str("id", ext.ID).Str("language", "go").Str("packageName", extensionPackageName).Msg("extensions: Loading external online streaming provider")

	payload := strings.Replace(ext.Payload, "package main", "package "+extensionPackageName, 1)

	// Load the extension payload
	_, err := r.yaegiEval(payload)
	if err != nil {
		r.logger.Error().Err(err).Str("id", ext.ID).Msg(MsgYaegiFailedToEvaluateExtensionCode)
		return fmt.Errorf(MsgYaegiFailedToEvaluateExtensionCode+": %v", err)
	}

	// Get the provider
	newProviderFuncVal, err := r.yaegiEval(extensionPackageName + `.NewProvider`)
	if err != nil {
		r.logger.Error().Err(err).Str("id", ext.ID).Msg(MsgYaegiFailedToEvaluateExtensionCode)
		return fmt.Errorf(MsgYaegiFailedToEvaluateExtensionCode+": %v", err)
	}

	newProviderFunc, ok := newProviderFuncVal.Interface().(func(logger *zerolog.Logger) hibikeonlinestream.Provider)
	if !ok {
		r.logger.Error().Str("id", ext.ID).Msg(MsgYaegiFailedToInstantiateExtension)
		return fmt.Errorf(MsgYaegiFailedToInstantiateExtension)
	}

	provider := newProviderFunc(r.logger)

	// Add the extension to the map
	r.onlinestreamProviderExtensionBank.Set(ext.ID, extension.NewOnlinestreamProviderExtension(ext, provider))
	return nil
}

//
// Typescript / Javascript
//

func (r *Repository) loadExternalOnlinestreamExtensionJS(ext *extension.Extension, language extension.Language) error {

	r.logger.Trace().Str("id", ext.ID).Any("language", language).Msg("extensions: Loading external online streaming provider")

	provider, gojaExt, err := NewGojaOnlinestreamProvider(ext, language, r.logger)
	if err != nil {
		r.logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to create javascript VM for external online streaming provider")
		return err
	}

	// Add the goja extension pointer to the map
	r.gojaExtensions.Set(ext.ID, gojaExt)

	// Add the extension to the map
	r.onlinestreamProviderExtensionBank.Set(ext.ID, extension.NewOnlinestreamProviderExtension(ext, provider))
	return nil
}
