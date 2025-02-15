package hook_event

import (
	"seanime/internal/api/anilist"
	"seanime/internal/library/anime"
)

/////////////////////////////
// Anime Library Events
/////////////////////////////

type PreGetAnimeEntryEvent struct {
	Event
	MediaId         *int                     `json:"mediaId"`
	LocalFiles      []*anime.LocalFile       `json:"localFiles"`
	AnimeCollection *anilist.AnimeCollection `json:"animeCollection"`
}

type AnimeEntryEvent struct {
	Event
	Entry *anime.Entry `json:"entry"`
}

type AnimeEntryFillerHydrationEvent struct {
	Event
	SkipDefault *bool        `json:"skipDefault"`
	Entry       *anime.Entry `json:"entry"`
}

type AnimeEntryErrorEvent struct {
	Event
	PreGetAnimeEntryEvent PreGetAnimeEntryEvent `json:"preGetAnimeEntryEvent"`
	Error                 error                 `json:"error"`
}
