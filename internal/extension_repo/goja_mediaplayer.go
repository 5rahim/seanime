package extension_repo

import (
	"fmt"
	"github.com/dop251/goja"
	"github.com/rs/zerolog"
	"seanime/internal/extension"

	hibikemediaplayer "github.com/5rahim/hibike/pkg/extension/mediaplayer"
)

type (
	GojaMediaPlayer struct {
		gojaExtensionImpl
	}
)

func NewGojaMediaPlayer(ext *extension.Extension, language extension.Language, logger *zerolog.Logger) (hibikemediaplayer.MediaPlayer, *GojaMediaPlayer, error) {
	logger.Trace().Str("id", ext.ID).Any("language", language).Msg("extensions: Loading external online streaming provider")

	vm, err := SetupGojaExtensionVM(ext, language, logger)
	if err != nil {
		logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to create javascript VM")
		return nil, nil, err
	}

	// Create the provider
	_, err = vm.RunString(`function NewMediaPlayer() {
   return new MediaPlayer()
}`)
	if err != nil {
		vm.ClearInterrupt()
		logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to create online streaming provider")
		return nil, nil, err
	}

	newMediaPlayerFunc, ok := goja.AssertFunction(vm.Get("NewMediaPlayer"))
	if !ok {
		vm.ClearInterrupt()
		logger.Error().Str("id", ext.ID).Msg("extensions: Failed to invoke online streaming provider constructor")
		return nil, nil, fmt.Errorf("failed to invoke online streaming provider constructor")
	}

	classObjVal, err := newMediaPlayerFunc(goja.Undefined())
	if err != nil {
		vm.ClearInterrupt()
		logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to create online streaming provider")
		return nil, nil, err
	}

	classObj := classObjVal.ToObject(vm)

	ret := &GojaMediaPlayer{
		gojaExtensionImpl: gojaExtensionImpl{
			vm:       vm,
			logger:   logger,
			ext:      ext,
			classObj: classObj,
		},
	}
	return ret, ret, nil
}

func (g *GojaMediaPlayer) GetVM() *goja.Runtime {
	return g.vm
}

func (g *GojaMediaPlayer) GetSettings() (ret hibikemediaplayer.Settings) {

	res, err := g.callClassMethod("getSettings")
	if err != nil {
		return hibikemediaplayer.Settings{}
	}

	err = g.unmarshalValue(res, &ret)
	if err != nil {
		return hibikemediaplayer.Settings{}
	}

	return
}

func (g *GojaMediaPlayer) InitConfig(config map[string]interface{}) {
	_, err := g.callClassMethod("initConfig", g.vm.ToValue(config))
	if err != nil {
		return
	}
}

func (g *GojaMediaPlayer) Start() error {
	_, err := g.callClassMethod("start")
	if err != nil {
		return err
	}
	return nil
}

func (g *GojaMediaPlayer) Stop() error {
	_, err := g.callClassMethod("stop")
	if err != nil {
		return err
	}
	return nil
}

func (g *GojaMediaPlayer) Play(req hibikemediaplayer.PlayRequest) (*hibikemediaplayer.PlayResponse, error) {
	res, err := g.callClassMethod("play", g.vm.ToValue(structToMap(req)))
	if err != nil {
		return nil, err
	}

	var ret hibikemediaplayer.PlayResponse
	err = g.unmarshalValue(res, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (g *GojaMediaPlayer) Stream(req hibikemediaplayer.PlayRequest) (*hibikemediaplayer.PlayResponse, error) {
	res, err := g.callClassMethod("stream", g.vm.ToValue(structToMap(req)))
	if err != nil {
		return nil, err
	}

	var ret hibikemediaplayer.PlayResponse
	err = g.unmarshalValue(res, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (g *GojaMediaPlayer) GetPlaybackStatus() (*hibikemediaplayer.PlaybackStatus, error) {
	res, err := g.callClassMethod("getPlaybackStatus")
	if err != nil {
		return nil, err
	}

	var ret hibikemediaplayer.PlaybackStatus
	err = g.unmarshalValue(res, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}
