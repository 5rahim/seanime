package extension_repo

import (
	"fmt"
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
	"github.com/rs/zerolog"
	"seanime/internal/extension"
	"seanime/internal/util"
	"strings"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Anime Torrent provider
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) loadExternalAnimeTorrentProviderExtension(ext *extension.Extension) (err error) {

	switch ext.Language {
	case extension.LanguageGo:
		err = r.loadExternalAnimeTorrentProviderExtensionGo(ext)
	case extension.LanguageJavascript:
		err = r.loadExternalAnimeTorrentProviderExtensionJS(ext, extension.LanguageJavascript)
	case extension.LanguageTypescript:
		err = r.loadExternalAnimeTorrentProviderExtensionJS(ext, extension.LanguageTypescript)
	default:
		err = fmt.Errorf("unsupported language: %v", ext.Language)
	}

	if err != nil {
		return
	}

	r.logger.Debug().Str("id", ext.ID).Msg("extensions: Loaded online streaming provider extension")
	return
}

//
// Go
//

func (r *Repository) loadExternalAnimeTorrentProviderExtensionGo(ext *extension.Extension) error {

	extensionPackageName := "ext_" + util.GenerateCryptoID()

	r.logger.Trace().Str("id", ext.ID).Str("packageName", extensionPackageName).Msg("extensions: Loading external torrent provider")

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

	newProviderFunc, ok := newProviderFuncVal.Interface().(func(logger *zerolog.Logger) hibiketorrent.AnimeProvider)
	if !ok {
		r.logger.Error().Str("id", ext.ID).Msg(MsgYaegiFailedToInstantiateExtension)
		return fmt.Errorf(MsgYaegiFailedToInstantiateExtension)
	}

	provider := newProviderFunc(r.logger)

	// Add the extension to the map
	r.animeTorrentProviderExtensionBank.Set(ext.ID, extension.NewAnimeTorrentProviderExtension(ext, provider))

	return nil
}

//
// Typescript / Javascript
//

func (r *Repository) loadExternalAnimeTorrentProviderExtensionJS(ext *extension.Extension, language extension.Language) error {

	provider, err := NewGojaAnimeTorrentProvider(ext, language, r.logger)
	if err != nil {
		r.logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to create javascript VM for external online streaming provider")
		return err
	}

	// Add the extension to the map
	r.animeTorrentProviderExtensionBank.Set(ext.ID, extension.NewAnimeTorrentProviderExtension(ext, provider))
	return nil
}
