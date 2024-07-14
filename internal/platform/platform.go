package platform

import (
	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"github.com/seanime-app/seanime/internal/api/anilist"
)

type Platform interface {
	SetUsername(username string)
	SetAnilistClientWrapper(client anilist.AnilistClient)
	UpdateEntry(mediaID int, status *anilist.MediaListStatus, scoreRaw *int, progress *int, startedAt *anilist.FuzzyDateInput, completedAt *anilist.FuzzyDateInput) error
	UpdateEntryProgress(mediaID int, progress int, totalEpisodes *int) error
	DeleteEntry(mediaID int) error
	GetAnime(mediaID int) (*anilist.BaseAnime, error)
	GetAnimeByMalID(malID int) (*anilist.BaseAnime, error)
	GetAnimeWithRelations(mediaID int) (*anilist.CompleteAnime, error)
	GetAnimeDetails(mediaID int) (*anilist.AnimeDetailsById_Media, error)
	GetManga(mediaID int) (*anilist.BaseManga, error)
	GetAnimeCollection(bypassCache bool) (*anilist.AnimeCollection, error)
	GetRawAnimeCollection(bypassCache bool) (*anilist.AnimeCollection, error)
	GetMangaDetails(mediaID int) (*anilist.MangaDetailsById_Media, error)
	GetAnimeCollectionWithRelations() (*anilist.AnimeCollectionWithRelations, error)
	GetMangaCollection(bypassCache bool) (*anilist.MangaCollection, error)
	GetRawMangaCollection(bypassCache bool) (*anilist.MangaCollection, error)
	AddMediaToCollection(mIds []int) error
	GetStudioDetails(studioID int) (*anilist.StudioDetails, error)
	GetAnilistClient() anilist.AnilistClient
	RefreshAnimeCollection() (*anilist.AnimeCollection, error)
	RefreshMangaCollection() (*anilist.MangaCollection, error)
}

func NewAnilistPlatform(anilistClientWrapper anilist.AnilistClient, logger *zerolog.Logger) Platform {
	ap := &AnilistPlatform{
		anilistClientWrapper: anilistClientWrapper,
		logger:               logger,
		username:             mo.None[string](),
		animeCollection:      mo.None[*anilist.AnimeCollection](),
		rawAnimeCollection:   mo.None[*anilist.AnimeCollection](),
		mangaCollection:      mo.None[*anilist.MangaCollection](),
		rawMangaCollection:   mo.None[*anilist.MangaCollection](),
	}

	return ap
}
