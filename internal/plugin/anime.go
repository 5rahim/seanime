package plugin

import (
	"context"
	"seanime/internal/api/metadata"
	"seanime/internal/database/db_bridge"
	"seanime/internal/extension"
	"seanime/internal/goja/goja_bindings"
	"seanime/internal/hook"
	"seanime/internal/library/anime"
	gojautil "seanime/internal/util/goja"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

type Anime struct {
	ctx       *AppContextImpl
	vm        *goja.Runtime
	logger    *zerolog.Logger
	ext       *extension.Extension
	scheduler *gojautil.Scheduler
}

func (a *AppContextImpl) BindAnimeToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler) {
	m := &Anime{
		ctx:       a,
		vm:        vm,
		logger:    logger,
		ext:       ext,
		scheduler: scheduler,
	}

	animeObj := vm.NewObject()

	// Get downloaded chapter containers
	_ = animeObj.Set("getAnimeEntry", m.getAnimeEntry)
	_ = animeObj.Set("getAnimeMetadata", m.getAnimeMetadata)
	_ = animeObj.Set("clearEpisodeMetadataCache", func(call goja.FunctionCall) goja.Value {
		metadataProviderRef, ok := a.metadataProviderRef.Get()
		if ok {
			metadataProviderRef.Get().ClearCache()
			anime.ClearEpisodeCollectionCache()
		}
		return goja.Undefined()
	})
	_ = obj.Set("anime", animeObj)
}

func (m *Anime) getAnimeMetadata(call goja.FunctionCall) goja.Value {
	promise, resolve, reject := m.vm.NewPromise()

	from := gojautil.ExpectStringArg(m.vm, call, 0)
	mediaId := int(gojautil.ExpectIntArg(m.vm, call, 1))

	metadataProviderRef, ok := m.ctx.metadataProviderRef.Get()
	if !ok {
		_ = reject(goja_bindings.NewErrorString(m.vm, "metadata provider not found"))
		return m.vm.ToValue(promise)
	}
	go func() {
		ret, err := metadataProviderRef.Get().GetAnimeMetadata(metadata.Platform(from), mediaId)
		if err != nil {
			_ = reject(m.vm.ToValue(err.Error()))
			return
		}

		m.scheduler.ScheduleAsync(func() error {
			_ = resolve(m.vm.ToValue(ret))
			return nil
		})
	}()

	return m.vm.ToValue(promise)
}

func (m *Anime) getAnimeEntry(call goja.FunctionCall) goja.Value {
	promise, resolve, reject := m.vm.NewPromise()

	mediaId := call.Argument(0).ToInteger()

	if mediaId == 0 {
		_ = reject(goja_bindings.NewErrorString(m.vm, "invalid media id"))
		return m.vm.ToValue(promise)
	}

	database, ok := m.ctx.database.Get()
	if !ok {
		_ = reject(goja_bindings.NewErrorString(m.vm, "database not found"))
		return m.vm.ToValue(promise)
	}

	anilistPlatformRef, ok := m.ctx.anilistPlatformRef.Get()
	if !ok {
		_ = reject(goja_bindings.NewErrorString(m.vm, "anilist platform not found"))
		return m.vm.ToValue(promise)
	}

	metadataProviderRef, ok := m.ctx.metadataProviderRef.Get()
	if !ok {
		_ = reject(goja_bindings.NewErrorString(m.vm, "metadata provider not found"))
		return m.vm.ToValue(promise)
	}

	fillerManager, ok := m.ctx.fillerManager.Get()
	if !ok {
		_ = reject(goja_bindings.NewErrorString(m.vm, "filler manager not found"))
		return m.vm.ToValue(promise)
	}

	go func() {
		// Get all the local files
		lfs, _, err := db_bridge.GetLocalFiles(database)
		if err != nil {
			_ = reject(m.vm.ToValue(err.Error()))
			return
		}

		// Get the user's anilist collection
		animeCollection, err := anilistPlatformRef.Get().GetAnimeCollection(context.Background(), false)
		if err != nil {
			_ = reject(m.vm.ToValue(err.Error()))
			return
		}

		if animeCollection == nil {
			_ = reject(goja_bindings.NewErrorString(m.vm, "anilist collection not found"))
			return
		}

		// Create a new media entry
		entry, err := anime.NewEntry(context.Background(), &anime.NewEntryOptions{
			MediaId:             int(mediaId),
			LocalFiles:          lfs,
			AnimeCollection:     animeCollection,
			PlatformRef:         anilistPlatformRef,
			MetadataProviderRef: metadataProviderRef,
		})
		if err != nil {
			_ = reject(goja_bindings.NewError(m.vm, err))
			return
		}

		fillerEvent := new(anime.AnimeEntryFillerHydrationEvent)
		fillerEvent.Entry = entry
		err = hook.GlobalHookManager.OnAnimeEntryFillerHydration().Trigger(fillerEvent)
		if err != nil {
			_ = reject(goja_bindings.NewError(m.vm, err))
			return
		}
		entry = fillerEvent.Entry

		if !fillerEvent.DefaultPrevented {
			fillerManager.HydrateFillerData(fillerEvent.Entry)
		}

		m.scheduler.ScheduleAsync(func() error {
			_ = resolve(m.vm.ToValue(entry))
			return nil
		})
	}()

	return m.vm.ToValue(promise)
}
