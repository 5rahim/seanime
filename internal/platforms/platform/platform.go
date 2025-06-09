package platform

import (
	"context"
	"seanime/internal/api/anilist"
)

type Platform interface {
	SetUsername(username string)
	// SetAnilistClient sets the AniList client to use for the platform
	SetAnilistClient(client anilist.AnilistClient)
	// UpdateEntry updates the entry for the given media ID
	UpdateEntry(context context.Context, mediaID int, status *anilist.MediaListStatus, scoreRaw *int, progress *int, startedAt *anilist.FuzzyDateInput, completedAt *anilist.FuzzyDateInput) error
	// UpdateEntryProgress updates the entry progress for the given media ID
	UpdateEntryProgress(context context.Context, mediaID int, progress int, totalEpisodes *int) error
	// UpdateEntryRepeat updates the entry repeat number for the given media ID
	UpdateEntryRepeat(context context.Context, mediaID int, repeat int) error
	// DeleteEntry deletes the entry for the given media ID
	DeleteEntry(context context.Context, mediaID int) error
	// GetAnime gets the anime for the given media ID
	GetAnime(context context.Context, mediaID int) (*anilist.BaseAnime, error)
	// GetAnimeByMalID gets the anime by MAL ID
	GetAnimeByMalID(context context.Context, malID int) (*anilist.BaseAnime, error)
	// GetAnimeWithRelations gets the anime with relations for the given media ID
	// This is used for scanning purposes in order to build the relation tree
	GetAnimeWithRelations(context context.Context, mediaID int) (*anilist.CompleteAnime, error)
	// GetAnimeDetails gets the anime details for the given media ID
	// These details are only fetched by the anime page
	GetAnimeDetails(context context.Context, mediaID int) (*anilist.AnimeDetailsById_Media, error)
	// GetManga gets the manga for the given media ID
	GetManga(context context.Context, mediaID int) (*anilist.BaseManga, error)
	// GetAnimeCollection gets the anime collection without custom lists
	// This should not make any API calls and instead should be based on GetRawAnimeCollection
	GetAnimeCollection(context context.Context, bypassCache bool) (*anilist.AnimeCollection, error)
	// GetRawAnimeCollection gets the anime collection with custom lists
	GetRawAnimeCollection(context context.Context, bypassCache bool) (*anilist.AnimeCollection, error)
	// GetMangaDetails gets the manga details for the given media ID
	// These details are only fetched by the manga page
	GetMangaDetails(context context.Context, mediaID int) (*anilist.MangaDetailsById_Media, error)
	// GetAnimeCollectionWithRelations gets the anime collection with relations
	// This is used for scanning purposes in order to build the relation tree
	GetAnimeCollectionWithRelations(context context.Context) (*anilist.AnimeCollectionWithRelations, error)
	// GetMangaCollection gets the manga collection without custom lists
	// This should not make any API calls and instead should be based on GetRawMangaCollection
	GetMangaCollection(context context.Context, bypassCache bool) (*anilist.MangaCollection, error)
	// GetRawMangaCollection gets the manga collection with custom lists
	GetRawMangaCollection(context context.Context, bypassCache bool) (*anilist.MangaCollection, error)
	// AddMediaToCollection adds the media to the collection
	AddMediaToCollection(context context.Context, mIds []int) error
	// GetStudioDetails gets the studio details for the given studio ID
	GetStudioDetails(context context.Context, studioID int) (*anilist.StudioDetails, error)
	// GetAnilistClient gets the AniList client
	GetAnilistClient() anilist.AnilistClient
	// RefreshAnimeCollection refreshes the anime collection
	RefreshAnimeCollection(context context.Context) (*anilist.AnimeCollection, error)
	// RefreshMangaCollection refreshes the manga collection
	RefreshMangaCollection(context context.Context) (*anilist.MangaCollection, error)
	// GetViewerStats gets the viewer stats
	GetViewerStats(context context.Context) (*anilist.ViewerStats, error)
	// GetAnimeAiringSchedule gets the schedule for airing anime in the collection
	GetAnimeAiringSchedule(context context.Context) (*anilist.AnimeAiringSchedule, error)
}
