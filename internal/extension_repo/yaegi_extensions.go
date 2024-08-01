package extension_repo

import (
	"fmt"
	hibikemanga "github.com/5rahim/hibike/pkg/extension/manga"
	hibikemediaplayer "github.com/5rahim/hibike/pkg/extension/mediaplayer"
	hibikeonlinestream "github.com/5rahim/hibike/pkg/extension/onlinestream"
	"github.com/rs/zerolog"
	"github.com/traefik/yaegi/interp"
	"seanime/internal/extension"
	"seanime/internal/util"
	"strings"

	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
)

func NewYaegiAnimeTorrentProvider(interp *interp.Interpreter, ext *extension.Extension, logger *zerolog.Logger) (hibiketorrent.AnimeProvider, error) {

	extensionPackageName := "ext_" + util.GenerateCryptoID()

	logger.Trace().Str("id", ext.ID).Str("language", "go").Str("packageName", extensionPackageName).Msg("extensions: Loading anime torrent provider extension")

	payload := strings.Replace(ext.Payload, "package main", "package "+extensionPackageName, 1)

	// Load the extension payload
	_, err := yaegiEval(interp, payload)
	if err != nil {
		logger.Error().Err(err).Str("id", ext.ID).Msg(MsgYaegiFailedToEvaluateExtensionCode)
		return nil, fmt.Errorf(MsgYaegiFailedToEvaluateExtensionCode+": %v", err)
	}

	// Get the provider
	newProviderFuncVal, err := yaegiEval(interp, extensionPackageName+`.NewProvider`)
	if err != nil {
		logger.Error().Err(err).Str("id", ext.ID).Msg(MsgYaegiFailedToEvaluateExtensionCode)
		return nil, fmt.Errorf(MsgYaegiFailedToEvaluateExtensionCode+": %v", err)
	}

	newProviderFunc, ok := newProviderFuncVal.Interface().(func(logger *zerolog.Logger) hibiketorrent.AnimeProvider)
	if !ok {
		logger.Error().Str("id", ext.ID).Msg(MsgYaegiFailedToInstantiateExtension)
		return nil, fmt.Errorf(MsgYaegiFailedToInstantiateExtension)
	}

	provider := newProviderFunc(logger)

	return provider, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func NewYaegiOnlinestreamProvider(interp *interp.Interpreter, ext *extension.Extension, logger *zerolog.Logger) (hibikeonlinestream.Provider, error) {

	extensionPackageName := "ext_" + util.GenerateCryptoID()

	logger.Trace().Str("id", ext.ID).Str("language", "go").Str("packageName", extensionPackageName).Msg("extensions: Loading online streaming provider extension")

	payload := strings.Replace(ext.Payload, "package main", "package "+extensionPackageName, 1)

	// Load the extension payload
	_, err := yaegiEval(interp, payload)
	if err != nil {
		logger.Error().Err(err).Str("id", ext.ID).Msg(MsgYaegiFailedToEvaluateExtensionCode)
		return nil, fmt.Errorf(MsgYaegiFailedToEvaluateExtensionCode+": %v", err)
	}

	// Get the provider
	newProviderFuncVal, err := yaegiEval(interp, extensionPackageName+`.NewProvider`)
	if err != nil {
		logger.Error().Err(err).Str("id", ext.ID).Msg(MsgYaegiFailedToEvaluateExtensionCode)
		return nil, fmt.Errorf(MsgYaegiFailedToEvaluateExtensionCode+": %v", err)
	}

	newProviderFunc, ok := newProviderFuncVal.Interface().(func(logger *zerolog.Logger) hibikeonlinestream.Provider)
	if !ok {
		logger.Error().Str("id", ext.ID).Msg(MsgYaegiFailedToInstantiateExtension)
		return nil, fmt.Errorf(MsgYaegiFailedToInstantiateExtension)
	}

	provider := newProviderFunc(logger)

	return provider, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func NewYaegiMangaProvider(interp *interp.Interpreter, ext *extension.Extension, logger *zerolog.Logger) (hibikemanga.Provider, error) {

	extensionPackageName := "ext_" + util.GenerateCryptoID()

	logger.Trace().Str("id", ext.ID).Str("language", "go").Str("packageName", extensionPackageName).Msg("extensions: Loading manga provider extension")

	payload := strings.Replace(ext.Payload, "package main", "package "+extensionPackageName, 1)

	// Load the extension payload
	_, err := yaegiEval(interp, payload)
	if err != nil {
		logger.Error().Err(err).Str("id", ext.ID).Msg(MsgYaegiFailedToEvaluateExtensionCode)
		return nil, fmt.Errorf(MsgYaegiFailedToEvaluateExtensionCode+": %v", err)
	}

	// Get the provider
	newProviderFuncVal, err := yaegiEval(interp, extensionPackageName+`.NewProvider`)
	if err != nil {
		logger.Error().Err(err).Str("id", ext.ID).Msg(MsgYaegiFailedToEvaluateExtensionCode)
		return nil, fmt.Errorf(MsgYaegiFailedToEvaluateExtensionCode+": %v", err)
	}

	newProviderFunc, ok := newProviderFuncVal.Interface().(func(logger *zerolog.Logger) hibikemanga.Provider)
	if !ok {
		logger.Error().Str("id", ext.ID).Msg(MsgYaegiFailedToInstantiateExtension)
		return nil, fmt.Errorf(MsgYaegiFailedToInstantiateExtension)
	}

	provider := newProviderFunc(logger)

	return provider, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func NewYaegiMediaPlayer(interp *interp.Interpreter, ext *extension.Extension, logger *zerolog.Logger) (hibikemediaplayer.MediaPlayer, error) {

	extensionPackageName := "ext_" + util.GenerateCryptoID()

	logger.Trace().Str("id", ext.ID).Str("language", "go").Str("packageName", extensionPackageName).Msg("extensions: Loading media player extension")

	payload := strings.Replace(ext.Payload, "package main", "package "+extensionPackageName, 1)

	// Load the extension payload
	_, err := yaegiEval(interp, payload)
	if err != nil {
		logger.Error().Err(err).Str("id", ext.ID).Msg(MsgYaegiFailedToEvaluateExtensionCode)
		return nil, fmt.Errorf(MsgYaegiFailedToEvaluateExtensionCode+": %v", err)
	}

	// Get the provider
	newProviderFuncVal, err := yaegiEval(interp, extensionPackageName+`.NewProvider`)
	if err != nil {
		logger.Error().Err(err).Str("id", ext.ID).Msg(MsgYaegiFailedToEvaluateExtensionCode)
		return nil, fmt.Errorf(MsgYaegiFailedToEvaluateExtensionCode+": %v", err)
	}

	newProviderFunc, ok := newProviderFuncVal.Interface().(func(logger *zerolog.Logger) hibikemediaplayer.MediaPlayer)
	if !ok {
		logger.Error().Str("id", ext.ID).Msg(MsgYaegiFailedToInstantiateExtension)
		return nil, fmt.Errorf(MsgYaegiFailedToInstantiateExtension)
	}

	provider := newProviderFunc(logger)

	return provider, nil
}
