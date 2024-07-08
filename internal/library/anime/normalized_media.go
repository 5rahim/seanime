package anime

import (
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/util/result"
)

type NormalizedMedia struct {
	*anilist.BaseMedia
}

type NormalizedMediaCache struct {
	*result.Cache[int, *NormalizedMedia]
}

func NewNormalizedMedia(m *anilist.BaseMedia) *NormalizedMedia {
	return &NormalizedMedia{
		BaseMedia: m,
	}
}

func NewNormalizedMediaCache() *NormalizedMediaCache {
	return &NormalizedMediaCache{result.NewCache[int, *NormalizedMedia]()}
}
