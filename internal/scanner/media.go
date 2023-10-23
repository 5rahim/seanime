package scanner

import (
	"context"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
	"github.com/seanime-app/seanime-server/internal/limiter"
	"github.com/seanime-app/seanime-server/internal/mal"
	"github.com/seanime-app/seanime-server/internal/util/parallel"
	"time"
)

// MediaContainer
//
// Hold all anilist.BaseMedia that will be used for the matching process
type MediaContainer struct {
	allMedia []*anilist.BaseMedia
}

// NewMediaContainer
//
// When enhancing is off, MediaContainer.allMedia will hold all anilist.BaseMedia from the user's AniList collection.
// When enhancing is on, MediaContainer.allMedia will hold anilist.BaseMedia for each unique, parsed anime title and their relations.
func NewMediaContainer(anilistClient *anilist.Client, localFiles []*LocalFile, enhancing bool, userName string) (*MediaContainer, error) {

	mc := new(MediaContainer)

	// Fetch user's AniList collection
	animeCollection, err := anilistClient.AnimeCollection(context.Background(), &userName)
	if err != nil {
		return nil, err
	}

	mc.allMedia = make([]*anilist.BaseMedia, 0)

	// For each collection entry, append the media to allMedia
	for _, list := range animeCollection.GetMediaListCollection().GetLists() {
		for _, entry := range list.GetEntries() {
			mc.allMedia = append(mc.allMedia, entry.GetMedia())
		}
	}

	if enhancing {
		//mc.fetchMediaTrees(localFiles)
	}

	return mc, nil
}

// FetchMediaTrees gets unique titles from local files.
// It then fetches mal.SearchResultAnime from MAL.
// It then uses these search results to get AniList IDs using anizip.Media mappings.
// Next, it queries AniList to retrieve anilist.BaseMedia's
func FetchMediaTrees(anilistClient *anilist.Client, localFiles []*LocalFile, baseMediaCache *anilist.BaseMediaCache, anizipCache *anizip.Cache) ([]*anilist.BaseMedia, bool) {
	rateLimiter := limiter.NewLimiter(time.Second, 2)
	anilistRateLimiter := limiter.NewLimiter(time.Minute, 90)

	// Get titles
	//titles := lop.Map(localFiles, func(file *LocalFile, index int) string {
	//	return file.GetParsedTitle()
	//})
	//titles = lo.Uniq(titles)

	//titles := []string{"Blue Lock", "One Piece", "Jujutsu Kaisen", "Hyouka", "Sousou no Frieren"}
	titles := []string{"Jujutsu Kaisen", "Sousou no Frieren"}

	// Get MAL media

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

	// Use the AniList IDs to get the AniList media
	anilistMediaResults := parallel.NewSettledResults[int, []*anilist.BaseMedia](anilistIds)
	anilistMediaResults.AllSettled(func(id int, index int) ([]*anilist.BaseMedia, error) {
		// Wait for the rate limiter
		anilistRateLimiter.Wait()

		// Fetch the media
		println("Fetching", id, "on AniList")
		m, err := anilist.GetBaseMediaById(anilistClient, id)
		if err != nil {
			return nil, err
		}

		// Fetch the media's relations
		tree := anilist.NewBaseMediaRelationTree()
		_ = m.FetchMediaTreeC(anilist.FetchMediaTreeAll, anilistClient, anilistRateLimiter, tree, baseMediaCache)

		// Get and return all media from the tree
		media := make([]*anilist.BaseMedia, 0)
		tree.Range(func(key int, value *anilist.BaseMedia) bool {
			media = append(media, value)
			return true
		})
		return media, nil
	})

	allMedia, ok := anilistMediaResults.GetFulfilledResults()

	if !ok {
		return nil, false
	}

	// Fetch media tree for each AniList media

	return lo.Flatten(*allMedia), true
}
