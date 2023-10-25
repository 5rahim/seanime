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
		_, ok := FetchMediaFromLocalFiles(opts.AnilistClient, opts.LocalFiles, opts.BaseMediaCache, opts.AnizipCache)
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

// FetchMediaFromLocalFiles gets media and their relations from local files.
// It retrieves unique titles from local files,
// fetches mal.SearchResultAnime from MAL,
// uses these search results to get AniList IDs using anizip.Media mappings,
// queries AniList to retrieve all anilist.BaseMedia using anilist.GetBaseMediaById and their relations using anilist.FetchMediaTree.
// It does not return an error if one of the steps fails.
// It returns the scanned media and a boolean indicating whether the process was successful.
func FetchMediaFromLocalFiles(
	anilistClient *anilist.Client,
	localFiles []*LocalFile,
	baseMediaCache *anilist.BaseMediaCache,
	anizipCache *anizip.Cache,
) ([]*anilist.BaseMedia, bool) {
	rateLimiter := limiter.NewLimiter(time.Second, 20)
	rateLimiter2 := limiter.NewLimiter(time.Second, 20)
	anilistRateLimiter := limiter.NewAnilistLimiter()

	// Get titles
	titles := lop.Map(localFiles, func(file *LocalFile, index int) string {
		return file.GetParsedTitle()
	})
	titles = lo.Uniq(titles)

	//titles = titles[:8]
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
	// Get the MAL media IDs
	malIds := lop.Map(malMedia, func(n *mal.SearchResultAnime, index int) int { return n.ID })

	// Get AniZip mappings for each MAL ID and store them in `anizipCache`
	// This step is necessary because MAL doesn't provide AniList IDs and some MAL media don't exist on AniList
	lop.ForEach(malIds, func(id int, index int) {
		println("Fetching", id, "on AniZip")
		rateLimiter2.Wait()
		_, _ = anizipCache.GetOrSet(anizip.GetCacheKey("mal", id), func() (*anizip.Media, error) {
			res, err := anizip.FetchAniZipMedia("mal", id)
			return res, err
		})
	})

	// Retrieve the AniList IDs from the AniZip mappings stored in the cache
	anilistIds := make([]int, 0)
	anizipCache.Range(func(key string, value *anizip.Media) bool {
		if value != nil {
			anilistIds = append(anilistIds, value.GetMappings().AnilistID)
		}
		return true
	})

	// Fetch all media from the AniList IDs
	anilistMedia := make([]*anilist.BaseMedia, 0)
	lop.ForEach(anilistIds, func(id int, index int) {
		anilistRateLimiter.Wait()
		media, err := anilist.GetBaseMediaById(anilistClient, id)
		if err == nil {
			anilistMedia = append(anilistMedia, media)
		} else {
			println("error while fetching", id, err.Error())
		}
	})

	// Create a new tree that will hold the fetched relations
	// /!\ This is redundant because we already have a cache, but `FetchMediaTree` needs its
	tree := anilist.NewBaseMediaRelationTree()

	// For each media, fetch its relations
	// The relations are fetched in parallel and added to `baseMediaCache`
	lop.ForEach(anilistMedia, func(m *anilist.BaseMedia, index int) {
		// We ignore errors because we want to continue even if one of the media fails
		_ = m.FetchMediaTree(anilist.FetchMediaTreeAll, anilistClient, anilistRateLimiter, tree, baseMediaCache)
	})

	// Retrieve all media from the cache
	scanned := make([]*anilist.BaseMedia, 0)
	baseMediaCache.Range(func(key int, value *anilist.BaseMedia) bool {
		scanned = append(scanned, value)
		return true
	})

	return scanned, true
}
