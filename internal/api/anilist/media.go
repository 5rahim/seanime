package anilist

import (
	"seanime/internal/util/result"
)

type BaseAnimeCache struct {
	*result.Cache[int, *BaseAnime]
}

// NewBaseAnimeCache returns a new result.Cache[int, *BaseAnime].
// It is used to temporarily store the results of FetchMediaTree calls.
func NewBaseAnimeCache() *BaseAnimeCache {
	return &BaseAnimeCache{result.NewCache[int, *BaseAnime]()}
}

type CompleteAnimeCache struct {
	*result.Cache[int, *CompleteAnime]
}

// NewCompleteAnimeCache returns a new result.Cache[int, *CompleteAnime].
// It is used to temporarily store the results of FetchMediaTree calls.
func NewCompleteAnimeCache() *CompleteAnimeCache {
	return &CompleteAnimeCache{result.NewCache[int, *CompleteAnime]()}
}
