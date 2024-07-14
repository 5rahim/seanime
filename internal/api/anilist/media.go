package anilist

import (
	"context"
	"github.com/seanime-app/seanime/internal/util/result"
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

//----------------------------------------------------------------------------------------------------------------------

func GetBaseAnimeById(acw AnilistClient, id int) (*BaseAnime, error) {
	res, err := acw.BaseAnimeByID(context.Background(), &id)
	if err != nil {
		return nil, err
	}

	return res.GetMedia(), nil
}

func GetCompleteAnimeById(acw AnilistClient, id int) (*CompleteAnime, error) {
	res, err := acw.CompleteAnimeByID(context.Background(), &id)
	if err != nil {
		return nil, err
	}

	return res.GetMedia(), nil
}

func GetBaseAnimeByIdC(anilistClient *Client, id int, cache *BaseAnimeCache) (*BaseAnime, error) {

	cacheV, ok := cache.Get(id)
	if ok {
		return cacheV, nil
	}

	res, err := anilistClient.BaseAnimeByID(context.Background(), &id)
	if err != nil {
		return nil, err
	}

	cache.Set(id, res.GetMedia())

	return res.GetMedia(), nil
}
