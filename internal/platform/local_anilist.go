package platform

import (
	"github.com/samber/mo"
	"github.com/seanime-app/seanime/internal/api/anilist"
)

type LocalPlatform struct {
	animeCollection mo.Option[*anilist.AnimeCollection]
	mangaCollection mo.Option[*anilist.MangaCollection]
}

func (lp *LocalPlatform) UpdateEntry(mediaID int, status *anilist.MediaListStatus, scoreRaw *int, progress *int, startedAt *anilist.FuzzyDateInput, completedAt *anilist.FuzzyDateInput) error {
	panic("not implemented")
}

func (lp *LocalPlatform) UpdateEntryProgress(mediaID int, progress int, totalEpisodes *int) error {
	panic("not implemented")
}

func (lp *LocalPlatform) DeleteEntry(mediaID int) error {
	panic("not implemented")
}

func (lp *LocalPlatform) GetAnime(mediaID int) (*anilist.BaseAnime, error) {
	panic("not implemented")
}

func (lp *LocalPlatform) GetAnimeWithRelations(mediaID int) (*anilist.CompleteAnime, error) {
	panic("not implemented")
}

func (lp *LocalPlatform) GetManga(mediaID int) (*anilist.BaseManga, error) {
	panic("not implemented")
}

func (lp *LocalPlatform) GetAnimeCollection() (*anilist.AnimeCollection, error) {
	panic("not implemented")
}

func (lp *LocalPlatform) GetAnimeCollectionWithRelations() (*anilist.AnimeCollectionWithRelations, error) {
	panic("not implemented")
}

func (lp *LocalPlatform) GetMangaCollection() (*anilist.MangaCollection, error) {
	panic("not implemented")
}

func (lp *LocalPlatform) AddMediaToCollection(mIds []int) error {
	panic("not implemented")
}
