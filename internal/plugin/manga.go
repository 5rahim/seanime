package plugin

import (
	"seanime/internal/extension"
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
	// m := &Manga{
	// 	ctx: a,
	// 	vm: vm,
	// 	logger: logger,
	// 	ext: ext,
	// 	scheduler: scheduler,
	// }

	mangaObj := vm.NewObject()
	// mangaRepo, ok := a.mangaRepository.Get()
	// anilistPlatform, foundAnilistPlatform := a.anilistPlatform.Get()

	// if ok && foundAnilistPlatform {

	// Get downloaded chapter containers
	// _ = mangaObj.Set("getDownloadedChapterContainers", func() ([]*manga.ChapterContainer, error) {
	// 	mangaCollection, err := anilistPlatform.GetMangaCollection(false)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	return mangaRepo.GetDownloadedChapterContainers(mangaCollection)
	// })
	// }
	_ = obj.Set("manga", mangaObj)
}
