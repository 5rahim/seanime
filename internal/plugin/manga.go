package plugin

import (
	"errors"
	"seanime/internal/extension"
	"seanime/internal/goja/goja_bindings"
	"seanime/internal/manga"
	goja_util "seanime/internal/util/goja"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

type Manga struct {
	ctx       *AppContextImpl
	vm        *goja.Runtime
	logger    *zerolog.Logger
	ext       *extension.Extension
	scheduler *goja_util.Scheduler
}

func (a *AppContextImpl) BindMangaToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler) {
	m := &Manga{
		ctx:       a,
		vm:        vm,
		logger:    logger,
		ext:       ext,
		scheduler: scheduler,
	}

	mangaObj := vm.NewObject()

	// Get downloaded chapter containers
	_ = mangaObj.Set("getDownloadedChapters", m.getDownloadedChapterContainers)
	_ = mangaObj.Set("getCollection", m.getCollection)
	_ = mangaObj.Set("refreshChapters", m.refreshChapterContainers)
	_ = mangaObj.Set("emptyCache", m.emptyCache)
	_ = mangaObj.Set("getChapterContainer", m.getChapterContainer)
	_ = mangaObj.Set("getProviders", m.getProviders)
	_ = obj.Set("manga", mangaObj)
}

func (m *Manga) getProviders() (map[string]string, error) {
	mangaRepo, ok := m.ctx.mangaRepository.Get()
	if !ok {
		return nil, errors.New("manga repository not found")
	}
	providers := make(map[string]string)
	extension.RangeExtensions(mangaRepo.GetProviderExtensionBank(), func(id string, ext extension.MangaProviderExtension) bool {
		providers[id] = ext.GetName()
		return true
	})
	return providers, nil
}

type GetChapterContainerOptions struct {
	MediaId  int
	Provider string
	Titles   []*string
	Year     int
}

func (m *Manga) getChapterContainer(opts *GetChapterContainerOptions) goja.Value {
	promise, resolve, reject := m.vm.NewPromise()

	mangaRepo, ok := m.ctx.mangaRepository.Get()
	if !ok {
		// reject(goja_bindings.NewErrorString(m.vm, "manga repository not set"))
		// return m.vm.ToValue(promise)
		goja_bindings.PanicThrowErrorString(m.vm, "manga repository not set")
	}

	go func() {
		ret, err := mangaRepo.GetMangaChapterContainer(&manga.GetMangaChapterContainerOptions{
			MediaId:  opts.MediaId,
			Provider: opts.Provider,
			Titles:   opts.Titles,
			Year:     opts.Year,
		})
		m.scheduler.ScheduleAsync(func() error {
			if err != nil {
				reject(err.Error())
			} else {
				resolve(ret)
			}
			return nil
		})
	}()

	return m.vm.ToValue(promise)
}

func (m *Manga) getDownloadedChapterContainers() ([]*manga.ChapterContainer, error) {
	mangaRepo, ok := m.ctx.mangaRepository.Get()
	if !ok {
		return nil, errors.New("manga repository not found")
	}
	anilistPlatform, foundAnilistPlatform := m.ctx.anilistPlatform.Get()
	if !foundAnilistPlatform {
		return nil, errors.New("anilist platform not found")
	}

	mangaCollection, err := anilistPlatform.GetMangaCollection(false)
	if err != nil {
		return nil, err
	}
	return mangaRepo.GetDownloadedChapterContainers(mangaCollection)
}

func (m *Manga) getCollection() (*manga.Collection, error) {
	anilistPlatform, foundAnilistPlatform := m.ctx.anilistPlatform.Get()
	if !foundAnilistPlatform {
		return nil, errors.New("anilist platform not found")
	}

	mangaCollection, err := anilistPlatform.GetMangaCollection(false)
	if err != nil {
		return nil, err
	}
	return manga.NewCollection(&manga.NewCollectionOptions{
		MangaCollection: mangaCollection,
		Platform:        anilistPlatform,
	})
}

func (m *Manga) refreshChapterContainers(selectedProviderMap map[int]string) goja.Value {
	promise, resolve, reject := m.vm.NewPromise()

	mangaRepo, ok := m.ctx.mangaRepository.Get()
	if !ok {
		jsErr := m.vm.NewGoError(errors.New("manga repository not found"))
		_ = reject(jsErr)
		return m.vm.ToValue(promise)
	}
	anilistPlatform, foundAnilistPlatform := m.ctx.anilistPlatform.Get()
	if !foundAnilistPlatform {
		jsErr := m.vm.NewGoError(errors.New("anilist platform not found"))
		_ = reject(jsErr)
		return m.vm.ToValue(promise)
	}

	mangaCollection, err := anilistPlatform.GetMangaCollection(false)
	if err != nil {
		reject(err.Error())
		return m.vm.ToValue(promise)
	}

	go func() {
		err := mangaRepo.RefreshChapterContainers(mangaCollection, selectedProviderMap)
		m.scheduler.ScheduleAsync(func() error {
			if err != nil {
				reject(err.Error())
			} else {
				resolve(nil)
			}
			return nil
		})
	}()

	return m.vm.ToValue(promise)
}

func (m *Manga) emptyCache(mediaId int) goja.Value {
	promise, resolve, reject := m.vm.NewPromise()

	mangaRepo, ok := m.ctx.mangaRepository.Get()
	if !ok {
		// reject(goja_bindings.NewErrorString(m.vm, "manga repository not found"))
		// return m.vm.ToValue(promise)
		goja_bindings.PanicThrowErrorString(m.vm, "manga repository not found")
	}

	go func() {
		err := mangaRepo.EmptyMangaCache(mediaId)
		m.scheduler.ScheduleAsync(func() error {
			if err != nil {
				reject(err.Error())
			} else {
				resolve(nil)
			}
			return nil
		})
	}()

	return m.vm.ToValue(promise)
}
