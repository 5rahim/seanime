package plugin

import (
	discordrpc_presence "seanime/internal/discordrpc/presence"
	"seanime/internal/extension"
	goja_util "seanime/internal/util/goja"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

func (a *AppContextImpl) BindDiscordToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler) {

	discordObj := vm.NewObject()
	_ = discordObj.Set("setMangaActivity", func(opts discordrpc_presence.MangaActivity) goja.Value {
		presence, ok := a.discordPresence.Get()
		if !ok {
			return goja.Undefined()
		}
		presence.SetMangaActivity(&opts)
		return goja.Undefined()
	})
	_ = discordObj.Set("setAnimeActivity", func(opts discordrpc_presence.AnimeActivity) goja.Value {
		presence, ok := a.discordPresence.Get()
		if !ok {
			return goja.Undefined()
		}
		presence.SetAnimeActivity(&opts)
		return goja.Undefined()
	})
	_ = discordObj.Set("cancelActivity", func() goja.Value {
		presence, ok := a.discordPresence.Get()
		if !ok {
			return goja.Undefined()
		}
		presence.Close()
		return goja.Undefined()
	})
	_ = obj.Set("discord", discordObj)
}
