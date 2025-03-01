package plugin

import (
	"seanime/internal/extension"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

type Playback struct {
	ctx *AppContextImpl
}

func (p *Playback) Bind(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension) {
	playbackObj := vm.NewObject()
	vm.Set("$playback", playbackObj)

}
