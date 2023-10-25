package scanner

import (
	"context"
	"errors"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
	"github.com/seanime-app/seanime-server/internal/limiter"
	"github.com/seanime-app/seanime-server/internal/mal"
	"github.com/seanime-app/seanime-server/internal/util/parallel"
	"sync"
	"time"
)

// MediaContainer holds all anilist.BaseMedia that will be used for the matching process
type MediaContainer struct {
	AllMedia []*anilist.BaseMedia
}

type MediaContainerOptions struct {
	Enhanced       bool
	Username       string
	AnilistClient  *anilist.Client
	LocalFiles     []*LocalFile
	BaseMediaCache *anilist.BaseMediaCache
	AnizipCache    *anizip.Cache
}

// NewMediaContainer
// When enhancing is off, MediaContainer.AllMedia will hold all anilist.BaseMedia from the user's AniList collection.
// When enhancing is on, MediaContainer.AllMedia will hold anilist.BaseMedia for each unique, parsed anime title and their relations.
func NewMediaContainer(opts *MediaContainerOptions) (*MediaContainer, error) {

	if opts.AnilistClient == nil ||
		opts.Username == "" ||
		opts.LocalFiles == nil ||
		opts.BaseMediaCache == nil ||
		opts.AnizipCache == nil {
		return nil, errors.New("missing options")
	}

	mc := new(MediaContainer)

	// Fetch user's AniList collection
	animeCollection, err := opts.AnilistClient.AnimeCollection(context.Background(), &opts.Username)
	if err != nil {
		return nil, err
	}

	mc.AllMedia = make([]*anilist.BaseMedia, 0)

	// For each collection entry, append the media to AllMedia
	for _, list := range animeCollection.GetMediaListCollection().GetLists() {
		for _, entry := range list.GetEntries() {
			mc.AllMedia = append(mc.AllMedia, entry.GetMedia())
			// We assume the BaseMediaCache is empty. Add media to cache.
			opts.BaseMediaCache.Set(entry.GetMedia().ID, entry.GetMedia())
		}
	}

	// If enhancing is on, scan media from local files and get their relations
	if opts.Enhanced {
		_, ok := FetchMediaTrees(opts.AnilistClient, opts.LocalFiles, opts.BaseMediaCache, opts.AnizipCache)
		if ok {
			// We assume the BaseMediaCache is populated. We overwrite AllMedia with the cache content.
			// This is because the cache will contain all media from the user's collection and the local files.
			mc.AllMedia = make([]*anilist.BaseMedia, 0)
			opts.BaseMediaCache.Range(func(key int, value *anilist.BaseMedia) bool {
				mc.AllMedia = append(mc.AllMedia, value)
				return true
			})
		}
	}

	return mc, nil
}

// FetchMediaTrees gets unique titles from local files.
// It then fetches mal.SearchResultAnime from MAL.
// It then uses these search results to get AniList IDs using anizip.Media mappings.
// Next, it queries AniList to retrieve anilist.BaseMedia's
func FetchMediaTrees(
	anilistClient *anilist.Client,
	localFiles []*LocalFile,
	baseMediaCache *anilist.BaseMediaCache,
	anizipCache *anizip.Cache,
) ([]*anilist.BaseMedia, bool) {
	rateLimiter := limiter.NewLimiter(time.Second, 20)
	anilistRateLimiter := limiter.NewAnilistLimiter()

	// Get titles
	titles := lop.Map(localFiles, func(file *LocalFile, index int) string {
		return file.GetParsedTitle()
	})
	titles = lo.Uniq(titles)

	//titles := []string{"Bungou Stray Dogs", "Jujutsu Kaisen", "Sousou no Frieren"}

	// Get MAL media from titles
	malSR := parallel.NewSettledResults[string, *mal.SearchResultAnime](titles)
	malSR.AllSettled(func(title string, index int) (*mal.SearchResultAnime, error) {
		rateLimiter.Wait()
		println("Fetching", title, "on MAL")
		return mal.AdvancedSearchWithMAL(title)
	})
	malRes, ok := malSR.GetFulfilledResults()

	if !ok {
		return nil, false
	}

	// Get duplicate-free version of MAL media
	malMedia := lo.UniqBy(*malRes, func(res *mal.SearchResultAnime) int { return res.ID })
	// Get their IDs
	malIds := lop.Map(malMedia, func(n *mal.SearchResultAnime, index int) int { return n.ID })

	// Get AniZip mappings for each MAL ID
	// This step is necessary because MAL doesn't provide AniList IDs and some MAL media don't exist on AniList

	parallel.EachTask(malIds, func(id int, index int) {
		println("Fetching", id, "on AniZip")
		_, _ = anizipCache.GetOrSet(anizip.GetCacheKey("mal", id), func() (*anizip.Media, error) {
			res, err := anizip.FetchAniZipMedia("mal", id)
			return res, err
		})
	})

	// Get the AniList IDs from the AniZip mappings
	anilistIds := make([]int, 0)

	anizipCache.Range(func(key string, value *anizip.Media) bool {
		if value != nil {
			anilistIds = append(anilistIds, value.GetMappings().AnilistID)
		}
		return true
	})

	// Use the AniList IDs to get the AniList media and their relations
	//anilistMediaResults := parallel.NewSettledResults[int, []*anilist.BaseMedia](anilistIds)
	//anilistMediaResults.AllSettled(func(id int, index int) ([]*anilist.BaseMedia, error) {
	//	// Wait for the rate limiter
	//	anilistRateLimiter.Wait()
	//
	//	// Fetch the media
	//	println("Fetching", id, "on AniList")
	//	media, err := anilist.GetBaseMediaById(anilistClient, id)
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	// Fetch the media's relations
	//	tree := anilist.NewBaseMediaRelationTree()
	//	_ = media.FetchMediaTree(anilist.FetchMediaTreeAll, anilistClient, anilistRateLimiter, tree, baseMediaCache)
	//
	//	// Get and return all media from the tree
	//	ret := make([]*anilist.BaseMedia, 0)
	//	tree.Range(func(key int, m *anilist.BaseMedia) bool {
	//		println("Adding", *m.GetTitleSafe(), "from tree to results")
	//		ret = append(ret, m)
	//		return true
	//	})
	//
	//	return ret, nil
	//})
	//
	//scanned, ok := anilistMediaResults.GetFulfilledResults()

	anilistMediaResults := parallel.NewSettledResults[int, *anilist.BaseMedia](anilistIds)
	anilistMediaResults.AllSettled(func(id int, index int) (*anilist.BaseMedia, error) {
		// Wait for the rate limiter
		anilistRateLimiter.Wait()

		// Fetch the media
		println("Fetching", id, "on AniList")
		media, err := anilist.GetBaseMediaById(anilistClient, id)
		return media, err
	})

	anilistMedia, ok := anilistMediaResults.GetFulfilledResults()
	if !ok {
		return nil, false
	}

	tree := anilist.NewBaseMediaRelationTree()
	wg := sync.WaitGroup{}

	for _, m := range *anilistMedia {
		wg.Add(1)
		go func(_m *anilist.BaseMedia) {
			defer wg.Done()

			err := _m.FetchMediaTree(anilist.FetchMediaTreeAll, anilistClient, rateLimiter, tree, baseMediaCache)
			if err != nil {
				return
			}

		}(m)
	}

	wg.Wait()

	tree.Range(func(key int, value *anilist.BaseMedia) bool {
		baseMediaCache.Set(key, value)
		return true
	})

	scanned := make([]*anilist.BaseMedia, 0)
	baseMediaCache.Range(func(key int, value *anilist.BaseMedia) bool {
		scanned = append(scanned, value)
		return true
	})

	if !ok {
		return nil, false
	}

	return scanned, true
}
