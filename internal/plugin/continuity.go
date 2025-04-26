package plugin

import (
	"seanime/internal/continuity"
	"seanime/internal/extension"
	"seanime/internal/goja/goja_bindings"
	goja_util "seanime/internal/util/goja"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

func (a *AppContextImpl) BindContinuityToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler) {

	continuityObj := vm.NewObject()

	_ = continuityObj.Set("updateWatchHistoryItem", func(opts continuity.UpdateWatchHistoryItemOptions) goja.Value {
		manager, ok := a.continuityManager.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "continuity manager not set")
		}
		err := manager.UpdateWatchHistoryItem(&opts)
		if err != nil {
			goja_bindings.PanicThrowError(vm, err)
		}
		return goja.Undefined()
	})

	_ = continuityObj.Set("getWatchHistoryItem", func(mediaId int) goja.Value {
		manager, ok := a.continuityManager.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "continuity manager not set")
		}
		resp := manager.GetWatchHistoryItem(mediaId)
		if resp == nil || !resp.Found {
			return goja.Undefined()
		}
		return vm.ToValue(resp.Item)
	})

	_ = continuityObj.Set("getWatchHistory", func() goja.Value {
		manager, ok := a.continuityManager.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "continuity manager not set")
		}
		return vm.ToValue(manager.GetWatchHistory())
	})

	_ = continuityObj.Set("deleteWatchHistoryItem", func(mediaId int) goja.Value {
		manager, ok := a.continuityManager.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "continuity manager not set")
		}
		err := manager.DeleteWatchHistoryItem(mediaId)
		if err != nil {
			goja_bindings.PanicThrowError(vm, err)
		}
		return goja.Undefined()
	})

	_ = obj.Set("continuity", continuityObj)
}
