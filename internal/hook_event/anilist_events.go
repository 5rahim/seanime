package hook_event

import (
	"seanime/internal/api/anilist"
)

/////////////////////////////
// AniList Events
/////////////////////////////

type GetAnimeEvent struct {
	Event
	Anime *anilist.BaseAnime `json:"anime"`
}

type GetAnimeDetailsEvent struct {
	Event
	Anime *anilist.AnimeDetailsById_Media `json:"anime"`
}

type GetMangaEvent struct {
	Event
	Manga *anilist.BaseManga `json:"manga"`
}

type GetMangaDetailsEvent struct {
	Event
	Manga *anilist.MangaDetailsById_Media `json:"manga"`
}

type GetAnimeCollectionEvent struct {
	Event
	AnimeCollection *anilist.AnimeCollection `json:"animeCollection"`
}

type GetMangaCollectionEvent struct {
	Event
	MangaCollection *anilist.MangaCollection `json:"mangaCollection"`
}

type GetRawAnimeCollectionEvent struct {
	Event
	AnimeCollection *anilist.AnimeCollection `json:"animeCollection"`
}

type GetRawMangaCollectionEvent struct {
	Event
	MangaCollection *anilist.MangaCollection `json:"mangaCollection"`
}

type GetStudioDetailsEvent struct {
	Event
	Studio *anilist.StudioDetails `json:"studio"`
}

type PreUpdateEntryEvent struct {
	Event
	MediaID     *int                     `json:"mediaId"`
	Status      *anilist.MediaListStatus `json:"status"`
	ScoreRaw    *int                     `json:"scoreRaw"`
	Progress    *int                     `json:"progress"`
	StartedAt   *anilist.FuzzyDateInput  `json:"startedAt"`
	CompletedAt *anilist.FuzzyDateInput  `json:"completedAt"`
}

type PostUpdateEntryEvent struct {
	Event
	MediaID *int `json:"mediaId"`
}

type PreUpdateEntryProgressEvent struct {
	Event
	// When true, Seanime's default logic for updating the progress will be overridden
	// This means the status will not be updated and the progress will not be clamped
	SkipDefault *bool `json:"skipDefault"`
	MediaID     *int  `json:"mediaId"`
	Progress    *int  `json:"progress"`
	TotalCount  *int  `json:"totalCount"`
	// Defaults to anilist.MediaListStatusCurrent
	Status *anilist.MediaListStatus `json:"status"`
}

type PostUpdateEntryProgressEvent struct {
	Event
	MediaID *int `json:"mediaId"`
}

type PreUpdateEntryRepeatEvent struct {
	Event
	MediaID *int `json:"mediaId"`
	Repeat  *int `json:"repeat"`
}

type PostUpdateEntryRepeatEvent struct {
	Event
	MediaID *int `json:"mediaId"`
}
