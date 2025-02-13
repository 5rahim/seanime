package hook_event

import (
	"seanime/internal/api/anilist"
)

type GetBaseAnimeEvent struct {
	Event

	Anime *anilist.BaseAnime `json:"anime"`
}

type GetBaseAnimeErrorEvent struct {
	Event

	Error error
}
