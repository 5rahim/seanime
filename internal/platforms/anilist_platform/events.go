package anilist_platform

import (
	"seanime/internal/api/anilist"
	"seanime/internal/hook"
)

const (
	GetBaseAnimeEventName = "GetBaseAnimeEvent"
)

type GetBaseAnimeEvent struct {
	hook.Event

	Anime *anilist.BaseAnime
}

type GetBaseAnimeErrorEvent struct {
	hook.Event

	Error error
}
