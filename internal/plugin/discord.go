package plugin

import (
	discordrpc_presence "seanime/internal/discordrpc/presence"
	"seanime/internal/extension"
	"seanime/internal/goja/goja_bindings"
	goja_util "seanime/internal/util/goja"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

func (a *AppContextImpl) BindDiscordToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler) {

	discordObj := vm.NewObject()
	_ = discordObj.Set("setMangaActivity", func(opts discordrpc_presence.MangaActivity) goja.Value {
		presence, ok := a.discordPresence.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "discord rpc client not set")
		}
		presence.SetMangaActivity(&opts)
		return goja.Undefined()
	})
	_ = discordObj.Set("setAnimeActivity", func(opts discordrpc_presence.AnimeActivity) goja.Value {
		presence, ok := a.discordPresence.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "discord rpc client not set")
		}
		presence.SetAnimeActivity(&opts)
		return goja.Undefined()
	})
	_ = discordObj.Set("updateAnimeActivity", func(progress int, duration int, paused bool) goja.Value {
		presence, ok := a.discordPresence.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "discord rpc client not set")
		}
		presence.UpdateAnimeActivity(progress, duration, paused)
		return goja.Undefined()
	})
	_ = discordObj.Set("setLegacyAnimeActivity", func(opts discordrpc_presence.LegacyAnimeActivity) goja.Value {
		presence, ok := a.discordPresence.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "discord rpc client not set")
		}
		presence.LegacySetAnimeActivity(&opts)
		return goja.Undefined()
	})
	_ = discordObj.Set("cancelActivity", func() goja.Value {
		presence, ok := a.discordPresence.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "discord rpc client not set")
		}
		presence.Close()
		return goja.Undefined()
	})
	_ = obj.Set("discord", discordObj)
}
