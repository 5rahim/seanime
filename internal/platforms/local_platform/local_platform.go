package local_platform

import (
	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"seanime/internal/api/anilist"
	"seanime/internal/platforms/platform"
	"sync"
)

type LocalPlatform struct {
	logger             *zerolog.Logger
	anilistClient      anilist.AnilistClient
	animeCollection    mo.Option[*anilist.AnimeCollection]
	rawAnimeCollection mo.Option[*anilist.AnimeCollection]
	mangaCollection    mo.Option[*anilist.MangaCollection]
	rawMangaCollection mo.Option[*anilist.MangaCollection]
	mangaMu            sync.RWMutex
	animeMu            sync.RWMutex
	localDb            *LocalPlatformDatabase
}

func NewLocalPlatform(dataDir string, anilistClient anilist.AnilistClient, logger *zerolog.Logger) (platform.Platform, error) {

	localDb, err := newLocalPlatformDatabase(dataDir, "local", logger)
	if err != nil {
		return nil, err
	}

	ap := &LocalPlatform{
		localDb:            localDb,
		anilistClient:      anilistClient,
		logger:             logger,
		animeCollection:    mo.None[*anilist.AnimeCollection](),
		rawAnimeCollection: mo.None[*anilist.AnimeCollection](),
		mangaCollection:    mo.None[*anilist.MangaCollection](),
		rawMangaCollection: mo.None[*anilist.MangaCollection](),
		mangaMu:            sync.RWMutex{},
		animeMu:            sync.RWMutex{},
	}

	go ap.loadAnimeCollection()
	go ap.loadMangaCollection()

	return ap, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (pm *LocalPlatform) SetUsername(username string) {
	panic("todo")
}

func (pm *LocalPlatform) SetAnilistClient(client anilist.AnilistClient) {
	panic("todo")
}

func (pm *LocalPlatform) UpdateEntry(mediaID int, status *anilist.MediaListStatus, scoreRaw *int, progress *int, startedAt *anilist.FuzzyDateInput, completedAt *anilist.FuzzyDateInput) error {
	panic("todo")
}

func (pm *LocalPlatform) UpdateEntryProgress(mediaID int, progress int, totalEpisodes *int) error {
	panic("todo")
}

func (pm *LocalPlatform) DeleteEntry(mediaID int) error {
	panic("todo")
}

func (pm *LocalPlatform) GetAnime(mediaID int) (*anilist.BaseAnime, error) {
	panic("todo")
}

func (pm *LocalPlatform) GetAnimeByMalID(malID int) (*anilist.BaseAnime, error) {
	panic("todo")
}

func (pm *LocalPlatform) GetAnimeDetails(mediaID int) (*anilist.AnimeDetailsById_Media, error) {
	panic("todo")
}

func (pm *LocalPlatform) GetAnimeWithRelations(mediaID int) (*anilist.CompleteAnime, error) {
	panic("todo")
}

func (pm *LocalPlatform) GetManga(mediaID int) (*anilist.BaseManga, error) {
	panic("todo")
}

func (pm *LocalPlatform) GetMangaDetails(mediaID int) (*anilist.MangaDetailsById_Media, error) {
	panic("todo")
}

func (pm *LocalPlatform) GetAnimeCollection(bypassCache bool) (*anilist.AnimeCollection, error) {
	panic("todo")
}

func (pm *LocalPlatform) GetRawAnimeCollection(bypassCache bool) (*anilist.AnimeCollection, error) {
	panic("todo")
}

func (pm *LocalPlatform) RefreshAnimeCollection() (*anilist.AnimeCollection, error) {
	panic("todo")
}

func (pm *LocalPlatform) refreshAnimeCollection() error {
	panic("todo")
}

func (pm *LocalPlatform) GetAnimeCollectionWithRelations() (*anilist.AnimeCollectionWithRelations, error) {
	panic("todo")
}

func (pm *LocalPlatform) GetMangaCollection(bypassCache bool) (*anilist.MangaCollection, error) {
	panic("todo")
}

func (pm *LocalPlatform) GetRawMangaCollection(bypassCache bool) (*anilist.MangaCollection, error) {
	panic("todo")
}

func (pm *LocalPlatform) RefreshMangaCollection() (*anilist.MangaCollection, error) {
	panic("todo")
}

func (pm *LocalPlatform) refreshMangaCollection() error {
	panic("todo")
}

func (pm *LocalPlatform) AddMediaToCollection(mIds []int) error {
	panic("todo")
}

func (pm *LocalPlatform) GetStudioDetails(studioID int) (*anilist.StudioDetails, error) {
	panic("todo")
}

func (pm *LocalPlatform) GetAnilistClient() anilist.AnilistClient {
	panic("todo")
}
