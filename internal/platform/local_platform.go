package platform

import (
	"github.com/samber/mo"
	"github.com/seanime-app/seanime/internal/api/anilist"
)

type LocalPlatform struct {
	localAnimeCollection mo.Option[*anilist.AnimeCollection]
	localMangaCollection mo.Option[*anilist.MangaCollection]
}

func (ap *LocalPlatform) SetUsername(username string) {
	panic("todo")
}

func (ap *LocalPlatform) SetAnilistClient(client anilist.AnilistClient) {
	panic("todo")
}

func (ap *LocalPlatform) UpdateEntry(mediaID int, status *anilist.MediaListStatus, scoreRaw *int, progress *int, startedAt *anilist.FuzzyDateInput, completedAt *anilist.FuzzyDateInput) error {
	panic("todo")
}

func (ap *LocalPlatform) UpdateEntryProgress(mediaID int, progress int, totalEpisodes *int) error {
	panic("todo")
}

func (ap *LocalPlatform) DeleteEntry(mediaID int) error {
	panic("todo")
}

func (ap *LocalPlatform) GetAnime(mediaID int) (*anilist.BaseAnime, error) {
	panic("todo")
}

func (ap *LocalPlatform) GetAnimeByMalID(malID int) (*anilist.BaseAnime, error) {
	panic("todo")
}

func (ap *LocalPlatform) GetAnimeDetails(mediaID int) (*anilist.AnimeDetailsById_Media, error) {
	panic("todo")
}

func (ap *LocalPlatform) GetAnimeWithRelations(mediaID int) (*anilist.CompleteAnime, error) {
	panic("todo")
}

func (ap *LocalPlatform) GetManga(mediaID int) (*anilist.BaseManga, error) {
	panic("todo")
}

func (ap *LocalPlatform) GetMangaDetails(mediaID int) (*anilist.MangaDetailsById_Media, error) {
	panic("todo")
}

func (ap *LocalPlatform) GetAnimeCollection(bypassCache bool) (*anilist.AnimeCollection, error) {
	panic("todo")
}

func (ap *LocalPlatform) GetRawAnimeCollection(bypassCache bool) (*anilist.AnimeCollection, error) {
	panic("todo")
}

func (ap *LocalPlatform) RefreshAnimeCollection() (*anilist.AnimeCollection, error) {
	panic("todo")
}

func (ap *LocalPlatform) refreshAnimeCollection() error {
	panic("todo")
}

func (ap *LocalPlatform) GetAnimeCollectionWithRelations() (*anilist.AnimeCollectionWithRelations, error) {
	panic("todo")
}

func (ap *LocalPlatform) GetMangaCollection(bypassCache bool) (*anilist.MangaCollection, error) {
	panic("todo")
}

func (ap *LocalPlatform) GetRawMangaCollection(bypassCache bool) (*anilist.MangaCollection, error) {
	panic("todo")
}

func (ap *LocalPlatform) RefreshMangaCollection() (*anilist.MangaCollection, error) {
	panic("todo")
}

func (ap *LocalPlatform) refreshMangaCollection() error {
	panic("todo")
}

func (ap *LocalPlatform) AddMediaToCollection(mIds []int) error {
	panic("todo")
}

func (ap *LocalPlatform) GetStudioDetails(studioID int) (*anilist.StudioDetails, error) {
	panic("todo")
}

func (ap *LocalPlatform) GetAnilistClient() anilist.AnilistClient {
	panic("todo")
}
