package anime

import (
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/util/result"
)

type NormalizedMedia struct {
	*anilist.BaseAnime
}

type NormalizedMediaCache struct {
	*result.Cache[int, *NormalizedMedia]
}

func NewNormalizedMedia(m *anilist.BaseAnime) *NormalizedMedia {
	return &NormalizedMedia{
		BaseAnime: m,
	}
}

func NewNormalizedMediaCache() *NormalizedMediaCache {
	return &NormalizedMediaCache{result.NewCache[int, *NormalizedMedia]()}
}
