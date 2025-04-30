package plugin

import (
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/goja/goja_bindings"
	"seanime/internal/library/anime"
	"seanime/internal/onlinestream"
	goja_util "seanime/internal/util/goja"
	"strconv"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

// BindTorrentstreamToContextObj binds 'torrentstream' to the UI context object
func (a *AppContextImpl) BindTorrentstreamToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler) {

}

// BindOnlinestreamToContextObj binds 'onlinestream' to the UI context object
func (a *AppContextImpl) BindOnlinestreamToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler) {

}

// BindMediastreamToContextObj binds 'mediastream' to the UI context object
func (a *AppContextImpl) BindMediastreamToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler) {

}

// BindTorrentClientToContextObj binds 'torrentClient' to the UI context object
func (a *AppContextImpl) BindTorrentClientToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler) {

	torrentClientObj := vm.NewObject()
	_ = torrentClientObj.Set("getTorrents", func() goja.Value {
		promise, resolve, reject := vm.NewPromise()

		torrentClient, ok := a.torrentClientRepository.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "torrentClient not set")
		}

		go func() {
			torrents, err := torrentClient.GetList()
			scheduler.ScheduleAsync(func() error {
				if err != nil {
					reject(goja_bindings.NewErrorString(vm, "error getting torrents: "+err.Error()))
					return nil
				}
				resolve(vm.ToValue(torrents))
				return nil
			})
		}()

		return vm.ToValue(promise)
	})

	_ = torrentClientObj.Set("getActiveTorrents", func() goja.Value {
		promise, resolve, reject := vm.NewPromise()

		torrentClient, ok := a.torrentClientRepository.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "torrentClient not set")
		}

		go func() {
			activeTorrents, err := torrentClient.GetActiveTorrents()
			scheduler.ScheduleAsync(func() error {
				if err != nil {
					reject(goja_bindings.NewErrorString(vm, "error getting active torrents: "+err.Error()))
					return nil
				}
				resolve(vm.ToValue(activeTorrents))
				return nil
			})
		}()

		return vm.ToValue(promise)
	})

	_ = torrentClientObj.Set("addMagnets", func(magnets []string, dest string) goja.Value {
		promise, resolve, reject := vm.NewPromise()

		torrentClient, ok := a.torrentClientRepository.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "torrentClient not set")
		}

		go func() {
			err := torrentClient.AddMagnets(magnets, dest)
			scheduler.ScheduleAsync(func() error {
				if err != nil {
					reject(goja_bindings.NewErrorString(vm, "error adding magnets: "+err.Error()))
					return nil
				}
				resolve(goja.Undefined())
				return nil
			})
		}()

		return vm.ToValue(promise)
	})

	_ = torrentClientObj.Set("removeTorrents", func(hashes []string) goja.Value {
		promise, resolve, reject := vm.NewPromise()

		torrentClient, ok := a.torrentClientRepository.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "torrentClient not set")
		}

		go func() {
			err := torrentClient.RemoveTorrents(hashes)
			scheduler.ScheduleAsync(func() error {
				if err != nil {
					reject(goja_bindings.NewErrorString(vm, "error removing torrents: "+err.Error()))
					return nil
				}
				resolve(goja.Undefined())
				return nil
			})
		}()

		return vm.ToValue(promise)
	})

	_ = torrentClientObj.Set("pauseTorrents", func(hashes []string) goja.Value {
		promise, resolve, reject := vm.NewPromise()

		torrentClient, ok := a.torrentClientRepository.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "torrentClient not set")
		}

		go func() {
			err := torrentClient.PauseTorrents(hashes)
			scheduler.ScheduleAsync(func() error {
				if err != nil {
					reject(goja_bindings.NewErrorString(vm, "error pausing torrents: "+err.Error()))
					return nil
				}
				resolve(goja.Undefined())
				return nil
			})
		}()

		return vm.ToValue(promise)
	})

	_ = torrentClientObj.Set("resumeTorrents", func(hashes []string) goja.Value {
		promise, resolve, reject := vm.NewPromise()

		torrentClient, ok := a.torrentClientRepository.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "torrentClient not set")
		}

		go func() {
			err := torrentClient.ResumeTorrents(hashes)
			scheduler.ScheduleAsync(func() error {
				if err != nil {
					reject(goja_bindings.NewErrorString(vm, "error resuming torrents: "+err.Error()))
					return nil
				}
				resolve(goja.Undefined())
				return nil
			})
		}()

		return vm.ToValue(promise)
	})

	_ = torrentClientObj.Set("deselectFiles", func(hash string, indices []int) goja.Value {
		promise, resolve, reject := vm.NewPromise()

		torrentClient, ok := a.torrentClientRepository.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "torrentClient not set")
		}

		go func() {
			err := torrentClient.DeselectFiles(hash, indices)
			scheduler.ScheduleAsync(func() error {
				if err != nil {
					reject(goja_bindings.NewErrorString(vm, "error deselecting files: "+err.Error()))
					return nil
				}
				resolve(goja.Undefined())
				return nil
			})
		}()

		return vm.ToValue(promise)
	})

	_ = torrentClientObj.Set("getFiles", func(hash string) goja.Value {
		promise, resolve, reject := vm.NewPromise()

		torrentClient, ok := a.torrentClientRepository.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "torrentClient not set")
		}

		go func() {
			files, err := torrentClient.GetFiles(hash)
			scheduler.ScheduleAsync(func() error {
				if err != nil {
					reject(goja_bindings.NewErrorString(vm, "error getting files: "+err.Error()))
					return nil
				}
				resolve(vm.ToValue(files))
				return nil
			})
		}()

		return vm.ToValue(promise)
	})

	_ = obj.Set("torrentClient", torrentClientObj)

}

// BindFillerManagerToContextObj binds 'fillerManager' to the UI context object
func (a *AppContextImpl) BindFillerManagerToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler) {

	fillerManagerObj := vm.NewObject()
	_ = fillerManagerObj.Set("getFillerEpisodes", func(mediaId int) goja.Value {
		fillerManager, ok := a.fillerManager.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "fillerManager not set")
		}
		fillerEpisodes, ok := fillerManager.GetFillerEpisodes(mediaId)
		if !ok {
			return goja.Undefined()
		}
		return vm.ToValue(fillerEpisodes)
	})

	_ = fillerManagerObj.Set("removeFillerData", func(mediaId int) goja.Value {
		fillerManager, ok := a.fillerManager.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "fillerManager not set")
		}
		fillerManager.RemoveFillerData(mediaId)
		return goja.Undefined()
	})

	_ = fillerManagerObj.Set("setFillerEpisodes", func(mediaId int, fillerEpisodes []string) goja.Value {
		fillerManager, ok := a.fillerManager.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "fillerManager not set")
		}
		fillerManager.StoreFillerData("plugin", strconv.Itoa(mediaId), mediaId, fillerEpisodes)
		return goja.Undefined()
	})

	_ = fillerManagerObj.Set("isEpisodeFiller", func(mediaId int, episodeNumber int) goja.Value {
		fillerManager, ok := a.fillerManager.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "fillerManager not set")
		}
		return vm.ToValue(fillerManager.IsEpisodeFiller(mediaId, episodeNumber))
	})

	_ = fillerManagerObj.Set("hydrateFillerData", func(e *anime.Entry) goja.Value {
		fillerManager, ok := a.fillerManager.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "fillerManager not set")
		}
		fillerManager.HydrateFillerData(e)
		return goja.Undefined()
	})

	_ = fillerManagerObj.Set("hydrateOnlinestreamFillerData", func(mId int, episodes []*onlinestream.Episode) goja.Value {
		fillerManager, ok := a.fillerManager.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "fillerManager not set")
		}
		fillerManager.HydrateOnlinestreamFillerData(mId, episodes)
		return goja.Undefined()
	})

	_ = obj.Set("fillerManager", fillerManagerObj)

}

// BindAutoDownloaderToContextObj binds 'autoDownloader' to the UI context object
func (a *AppContextImpl) BindAutoDownloaderToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler) {

	autoDownloaderObj := vm.NewObject()
	_ = autoDownloaderObj.Set("run", func() goja.Value {
		autoDownloader, ok := a.autoDownloader.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "autoDownloader not set")
		}
		autoDownloader.Run()
		return goja.Undefined()
	})
	_ = obj.Set("autoDownloader", autoDownloaderObj)
}

// BindAutoScannerToContextObj binds 'autoScanner' to the UI context object
func (a *AppContextImpl) BindAutoScannerToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler) {

	autoScannerObj := vm.NewObject()
	_ = autoScannerObj.Set("notify", func() goja.Value {
		autoScanner, ok := a.autoScanner.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "autoScanner not set")
		}
		autoScanner.Notify()
		return goja.Undefined()
	})
	_ = obj.Set("autoScanner", autoScannerObj)

}

// BindFileCacherToContextObj binds 'fileCacher' to the UI context object
func (a *AppContextImpl) BindFileCacherToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler) {

}

// BindExternalPlayerLinkToContextObj binds 'externalPlayerLink' to the UI context object
func (a *AppContextImpl) BindExternalPlayerLinkToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler) {

	externalPlayerLinkObj := vm.NewObject()
	_ = externalPlayerLinkObj.Set("open", func(url string, mediaId int, episodeNumber int) goja.Value {
		wsEventManager, ok := a.wsEventManager.Get()
		if !ok {
			goja_bindings.PanicThrowErrorString(vm, "wsEventManager not set")
		}
		// Send the external player link
		wsEventManager.SendEvent(events.ExternalPlayerOpenURL, struct {
			Url           string `json:"url"`
			MediaId       int    `json:"mediaId"`
			EpisodeNumber int    `json:"episodeNumber"`
		}{
			Url:           url,
			MediaId:       mediaId,
			EpisodeNumber: episodeNumber,
		})
		return goja.Undefined()
	})
	_ = obj.Set("externalPlayerLink", externalPlayerLinkObj)
}
