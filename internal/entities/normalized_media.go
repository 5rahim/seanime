package entities

import (
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/result"
)

type NormalizedMedia struct {
	*anilist.BasicMedia
}

type NormalizedMediaCache struct {
	*result.Cache[int, *NormalizedMedia]
}

func NewNormalizedMedia(m *anilist.BasicMedia) *NormalizedMedia {
	return &NormalizedMedia{
		BasicMedia: m,
	}
}

func NewNormalizedMediaCache() *NormalizedMediaCache {
	return &NormalizedMediaCache{result.NewCache[int, *NormalizedMedia]()}
}
