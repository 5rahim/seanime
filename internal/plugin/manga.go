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
	_ = mangaObj.Set("getDownloadedChapterContainers", m.getDownloadedChapterContainers)
	_ = mangaObj.Set("getCollection", m.getCollection)
	_ = mangaObj.Set("refreshChapterContainers", m.refreshChapterContainers)
	_ = obj.Set("manga", mangaObj)
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

func (m *Manga) refreshChapterContainers() *goja.Promise {
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
		err := mangaRepo.RefreshChapterContainers(mangaCollection)
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
