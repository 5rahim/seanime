package plugin_ui

import (
	"sync"

	"github.com/dop251/goja"
)

type ScreenManager struct {
	ctx *Context
	mu  sync.RWMutex
}

func NewScreenManager(ctx *Context) *ScreenManager {
	return &ScreenManager{
		ctx: ctx,
	}
}

func (s *ScreenManager) bind(vm *goja.Runtime, ctxObj *goja.Object) {
	screenObj := vm.NewObject()
	_ = screenObj.Set("onNavigate", s.jsOnNavigate)

	_ = ctxObj.Set("screen", screenObj)
}

// jsOnNavigate registers a callback to be called when the current screen changes
//
//	Example:
//	const onNavigate = (event) => {
//		console.log(event.screen);
//	};
//	ctx.screen.onNavigate(onNavigate);
func (s *ScreenManager) jsOnNavigate(callback goja.Callable) {
	eventListener := s.ctx.RegisterEventListener()
	var payload ScreenChangedEventPayload

	go func(payload ScreenChangedEventPayload) {
		for event := range eventListener.Channel {
			if event.ParsePayloadAs(ScreenChangedEvent, &payload) {
				_ = s.ctx.scheduler.Schedule(func() error {
					_, err := callback(goja.Undefined(), s.ctx.vm.ToValue(payload))
					return err
				})
			}
		}
	}(payload)
}
