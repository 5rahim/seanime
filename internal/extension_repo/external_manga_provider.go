package extension_repo

import (
	"fmt"
	hibikemanga "github.com/5rahim/hibike/pkg/extension/manga"
	"github.com/rs/zerolog"
	"seanime/internal/extension"
	"seanime/internal/util"
	"strings"
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

	r.logger.Debug().Str("id", ext.ID).Msg("extensions: Loaded manga provider extension")
	return
}

//
// Go
//

func (r *Repository) loadExternalMangaExtensionGo(ext *extension.Extension) error {

	extensionPackageName := "ext_" + util.GenerateCryptoID()

	r.logger.Trace().Str("id", ext.ID).Str("packageName", extensionPackageName).Msg("extensions: Loading external manga provider")

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

	newProviderFunc, ok := newProviderFuncVal.Interface().(func(logger *zerolog.Logger) hibikemanga.Provider)
	if !ok {
		r.logger.Error().Str("id", ext.ID).Msg(MsgYaegiFailedToInstantiateExtension)
		return fmt.Errorf(MsgYaegiFailedToInstantiateExtension)
	}

	provider := newProviderFunc(r.logger)

	// Add the extension to the map
	r.mangaProviderExtensionBank.Set(ext.ID, extension.NewMangaProviderExtension(ext, provider))
	return nil
}

//
// Typescript / Javascript
//

func (r *Repository) loadExternalMangaExtensionJS(ext *extension.Extension, language extension.Language) error {

	provider, err := NewGojaMangaProvider(ext, language, r.logger)
	if err != nil {
		r.logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to create javascript VM for external manga provider")
		return err
	}

	// Add the extension to the map
	r.mangaProviderExtensionBank.Set(ext.ID, extension.NewMangaProviderExtension(ext, provider))
	return nil
}
