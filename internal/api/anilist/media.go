package anilist

import (
	"context"
	"github.com/seanime-app/seanime/internal/util/result"
)

type BaseMediaCache struct {
	*result.Cache[int, *BaseMedia]
}

// NewBaseMediaCache returns a new result.Cache[int, *BaseMedia].
// It is used to temporarily store the results of FetchMediaTree calls.
func NewBaseMediaCache() *BaseMediaCache {
	return &BaseMediaCache{result.NewCache[int, *BaseMedia]()}
}

type CompleteMediaCache struct {
	*result.Cache[int, *CompleteMedia]
}

// NewCompleteMediaCache returns a new result.Cache[int, *CompleteMedia].
// It is used to temporarily store the results of FetchMediaTree calls.
func NewCompleteMediaCache() *CompleteMediaCache {
	return &CompleteMediaCache{result.NewCache[int, *CompleteMedia]()}
}

//----------------------------------------------------------------------------------------------------------------------

func GetBaseMediaById(acw ClientWrapperInterface, id int) (*BaseMedia, error) {
	res, err := acw.BaseMediaByID(context.Background(), &id)
	if err != nil {
		return nil, err
	}

	return res.GetMedia(), nil
}

func GetCompleteMediaById(acw ClientWrapperInterface, id int) (*CompleteMedia, error) {
	res, err := acw.CompleteMediaByID(context.Background(), &id)
	if err != nil {
		return nil, err
	}

	return res.GetMedia(), nil
}

func GetBaseMediaByIdC(anilistClient *Client, id int, cache *BaseMediaCache) (*BaseMedia, error) {

	cacheV, ok := cache.Get(id)
	if ok {
		return cacheV, nil
	}

	res, err := anilistClient.BaseMediaByID(context.Background(), &id)
	if err != nil {
		return nil, err
	}

	cache.Set(id, res.GetMedia())

	return res.GetMedia(), nil
}
