package plugin

import (
	"errors"
	"seanime/internal/extension"
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
	_ = obj.Set("manga", mangaObj)
}

type GetChapterContainerOptions struct {
	MediaId  int
	Provider string
	Titles   []*string
	Year     int
}

func (m *Manga) getChapterContainer(opts *GetChapterContainerOptions) *goja.Promise {
	promise, resolve, reject := m.vm.NewPromise()

	mangaRepo, ok := m.ctx.mangaRepository.Get()
	if !ok {
		reject(errors.New("manga repository not found"))
		return promise
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

	return promise
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

func (m *Manga) refreshChapterContainers(selectedProviderMap map[int]string) *goja.Promise {
	promise, resolve, reject := m.vm.NewPromise()

	mangaRepo, ok := m.ctx.mangaRepository.Get()
	if !ok {
		reject(errors.New("manga repository not found"))
		return promise
	}
	anilistPlatform, foundAnilistPlatform := m.ctx.anilistPlatform.Get()
	if !foundAnilistPlatform {
		reject(errors.New("anilist platform not found"))
		return promise
	}

	mangaCollection, err := anilistPlatform.GetMangaCollection(false)
	if err != nil {
		reject(err.Error())
		return promise
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

	return promise
}

func (m *Manga) emptyCache(mediaId int) *goja.Promise {
	promise, resolve, reject := m.vm.NewPromise()

	mangaRepo, ok := m.ctx.mangaRepository.Get()
	if !ok {
		reject(errors.New("manga repository not found"))
		return promise
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

	return promise
}
