package plugin

import (
	"seanime/internal/constants"
	"seanime/internal/events"
	"seanime/internal/extension"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

func (a *AppContextImpl) BindApp(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension) {
	appObj := vm.NewObject()
	appObj.Set("getVersion", constants.Version)
	appObj.Set("getVersionName", constants.VersionName)

	appObj.Set("invalidateClientQuery", func(keys []string) {
		wsEventManager, ok := a.wsEventManager.Get()
		if !ok {
			return
		}
		wsEventManager.SendEvent(events.InvalidateQueries, keys)
	})

	_ = vm.Set("$app", appObj)
}
