package anilist_platform

import (
	"seanime/internal/api/anilist"
	"seanime/internal/hook_resolver"
)

/////////////////////////////
// AniList Events
/////////////////////////////

type GetAnimeEvent struct {
	hook_resolver.Event
	Anime *anilist.BaseAnime `json:"anime"`
}

type GetAnimeDetailsEvent struct {
	hook_resolver.Event
	Anime *anilist.AnimeDetailsById_Media `json:"anime"`
}

type GetMangaEvent struct {
	hook_resolver.Event
	Manga *anilist.BaseManga `json:"manga"`
}

type GetMangaDetailsEvent struct {
	hook_resolver.Event
	Manga *anilist.MangaDetailsById_Media `json:"manga"`
}

type GetCachedAnimeCollectionEvent struct {
	hook_resolver.Event
	AnimeCollection *anilist.AnimeCollection `json:"animeCollection"`
}

type GetCachedMangaCollectionEvent struct {
	hook_resolver.Event
	MangaCollection *anilist.MangaCollection `json:"mangaCollection"`
}

type GetAnimeCollectionEvent struct {
	hook_resolver.Event
	AnimeCollection *anilist.AnimeCollection `json:"animeCollection"`
}

type GetMangaCollectionEvent struct {
	hook_resolver.Event
	MangaCollection *anilist.MangaCollection `json:"mangaCollection"`
}

type GetCachedRawAnimeCollectionEvent struct {
	hook_resolver.Event
	AnimeCollection *anilist.AnimeCollection `json:"animeCollection"`
}

type GetCachedRawMangaCollectionEvent struct {
	hook_resolver.Event
	MangaCollection *anilist.MangaCollection `json:"mangaCollection"`
}

type GetRawAnimeCollectionEvent struct {
	hook_resolver.Event
	AnimeCollection *anilist.AnimeCollection `json:"animeCollection"`
}

type GetRawMangaCollectionEvent struct {
	hook_resolver.Event
	MangaCollection *anilist.MangaCollection `json:"mangaCollection"`
}

type GetStudioDetailsEvent struct {
	hook_resolver.Event
	Studio *anilist.StudioDetails `json:"studio"`
}

// PreUpdateEntryEvent is triggered when an entry is about to be updated.
// Prevent default to skip the default update and override the update.
type PreUpdateEntryEvent struct {
	hook_resolver.Event
	MediaID     *int                     `json:"mediaId"`
	Status      *anilist.MediaListStatus `json:"status"`
	ScoreRaw    *int                     `json:"scoreRaw"`
	Progress    *int                     `json:"progress"`
	StartedAt   *anilist.FuzzyDateInput  `json:"startedAt"`
	CompletedAt *anilist.FuzzyDateInput  `json:"completedAt"`
}

type PostUpdateEntryEvent struct {
	hook_resolver.Event
	MediaID *int `json:"mediaId"`
}

// PreUpdateEntryProgressEvent is triggered when an entry's progress is about to be updated.
// Prevent default to skip the default update and override the update.
type PreUpdateEntryProgressEvent struct {
	hook_resolver.Event
	MediaID    *int `json:"mediaId"`
	Progress   *int `json:"progress"`
	TotalCount *int `json:"totalCount"`
	// Defaults to anilist.MediaListStatusCurrent
	Status *anilist.MediaListStatus `json:"status"`
}

type PostUpdateEntryProgressEvent struct {
	hook_resolver.Event
	MediaID *int `json:"mediaId"`
}

// PreUpdateEntryRepeatEvent is triggered when an entry's repeat is about to be updated.
// Prevent default to skip the default update and override the update.
type PreUpdateEntryRepeatEvent struct {
	hook_resolver.Event
	MediaID *int `json:"mediaId"`
	Repeat  *int `json:"repeat"`
}

type PostUpdateEntryRepeatEvent struct {
	hook_resolver.Event
	MediaID *int `json:"mediaId"`
}
