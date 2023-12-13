package handlers

import (
	"errors"
	"github.com/5rahim/tanuki"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/anizip"
	"github.com/seanime-app/seanime/internal/comparison"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/nyaa"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/sourcegraph/conc/pool"
	"sort"
	"strconv"
)

type (
	TorrentPreview struct {
		Episode       *entities.MediaEntryEpisode `json:"episode"`                 // nil if batch
		EpisodeNumber *int                        `json:"episodeNumber,omitempty"` // nil if batch
		IsBatch       bool                        `json:"isBatch"`
		Resolution    string                      `json:"resolution"`
		ReleaseGroup  string                      `json:"releaseGroup"`
		Torrent       nyaa.DetailedTorrent        `json:"torrent"`
	}
	TorrentSearchData struct {
		Previews []*TorrentPreview       `json:"previews"`
		Torrents []*nyaa.DetailedTorrent `json:"torrents"`
	}
)

func HandleNyaaSearch(c *RouteCtx) error {

	type body struct {
		QuickSearch    *bool              `json:"quickSearch"`
		Query          *string            `json:"query"`
		EpisodeNumber  *int               `json:"episodeNumber"`
		Batch          *bool              `json:"batch"`
		Media          *anilist.BaseMedia `json:"media"`
		AbsoluteOffset *int               `json:"absoluteOffset"`
		Resolution     *string            `json:"resolution"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	if b.QuickSearch == nil ||
		b.Media == nil ||
		b.Batch == nil ||
		b.EpisodeNumber == nil ||
		b.AbsoluteOffset == nil ||
		b.Resolution == nil ||
		b.Query == nil {
		return c.RespondWithError(errors.New("missing arguments"))
	}

	ret := make([]*nyaa.DetailedTorrent, 0)

	// +---------------------+
	// | Build Search query  |
	// +---------------------+
	// Use quick search if the user turned it on OR has not specified a query
	if *b.QuickSearch || len(*b.Query) == 0 {
		queries, ok := nyaa.BuildSearchQuery(&nyaa.BuildSearchQueryOptions{
			Media:          b.Media,
			Batch:          b.Batch,
			EpisodeNumber:  b.EpisodeNumber,
			Resolution:     b.Resolution,
			AbsoluteOffset: b.AbsoluteOffset,
			Title:          b.Query,
		})
		if !ok {
			return c.RespondWithError(errors.New("could not build search query"))
		}
		c.App.Logger.Trace().Msgf("nyaa query: %+v", queries)

		// +---------------------+
		// |   Search multiple   |
		// +---------------------+

		res, err := nyaa.SearchMultiple(nyaa.SearchMultipleOptions{
			Provider: "nyaa",
			Query:    queries,
			Category: "anime-eng",
			SortBy:   "seeders",
			Filter:   "",
			Cache:    c.App.NyaaSearchCache,
		})
		if err != nil {
			return c.RespondWithError(err)
		}
		ret = res
	} else {

		// +---------------------+
		// |       Query         |
		// +---------------------+

		res, err := nyaa.Search(nyaa.SearchOptions{
			Provider: "nyaa",
			Query:    *b.Query,
			Category: "anime-eng",
			SortBy:   "seeders",
			Filter:   "",
			Cache:    c.App.NyaaSearchCache,
		})
		if err != nil {
			return c.RespondWithError(err)
		}
		ret = res
	}

	// +---------------------+
	// |    Anizip Cache     |
	// +---------------------+

	// Verify that cache has the AniZip media
	_, ok := c.App.AnizipCache.Get(anizip.GetCacheKey("anilist", b.Media.ID))
	if !ok {
		_, err := anizip.FetchAniZipMediaC("anilist", b.Media.ID, c.App.AnizipCache)
		if err != nil {
			return c.RespondWithError(err)
		}
	}

	// +---------------------+
	// |   Torrent Preview   |
	// +---------------------+

	// Create torrent previews in parallel
	p := pool.NewWithResults[*TorrentPreview]()
	for _, torrent := range ret {
		torrent := torrent
		p.Go(func() *TorrentPreview {
			tp, ok := createTorrentPreview(b.Media, c.App.AnizipCache, torrent, *b.AbsoluteOffset)
			if !ok {
				return nil
			}
			return tp
		})
	}
	previews := p.Wait()
	previews = lo.Filter(previews, func(i *TorrentPreview, _ int) bool {
		return i != nil
	})

	/*
		Sorting
	*/

	// sort both by seeders
	sort.Slice(ret, func(i, j int) bool {
		iS, _ := strconv.Atoi(ret[i].Seeders)
		jS, _ := strconv.Atoi(ret[j].Seeders)
		return iS > jS
	})
	sort.Slice(previews, func(i, j int) bool {
		iS, _ := strconv.Atoi(previews[i].Torrent.Seeders)
		jS, _ := strconv.Atoi(previews[j].Torrent.Seeders)
		return iS > jS
	})

	return c.RespondWithData(TorrentSearchData{
		Previews: previews,
		Torrents: ret,
	})

}

//----------------------------------------------------------------------------------------------------------------------

func createTorrentPreview(
	media *anilist.BaseMedia,
	anizipCache *anizip.Cache,
	torrent *nyaa.DetailedTorrent,
	absoluteOffset int,
) (*TorrentPreview, bool) {

	anizipMedia, ok := anizipCache.Get(anizip.GetCacheKey("anilist", media.ID))
	if !ok {
		return nil, false
	}

	elements := tanuki.Parse(torrent.Name, tanuki.DefaultOptions)
	if len(elements.AnimeTitle) == 0 {
		return nil, false
	}

	// -1 = error
	// -2 = batch
	episodeNumber := -1

	if len(elements.EpisodeNumber) == 1 {
		asInt, ok := util.StringToInt(elements.EpisodeNumber[0])
		if ok {
			episodeNumber = asInt
		}
	} else if len(elements.EpisodeNumber) > 1 {
		episodeNumber = -2
	}

	// Check if the torrent is a batch, if we still have no episode number
	if episodeNumber < 0 {
		if comparison.ValueContainsBatchKeywords(torrent.Name) {
			episodeNumber = -2
		}
	}

	// normalize episode number
	if episodeNumber >= 0 && episodeNumber > media.GetCurrentEpisodeCount() {
		episodeNumber = episodeNumber - absoluteOffset
	}

	if *media.GetFormat() == anilist.MediaFormatMovie {
		episodeNumber = 1
	}

	ret := &TorrentPreview{
		IsBatch:      episodeNumber == -2,
		Resolution:   elements.VideoResolution,
		ReleaseGroup: elements.ReleaseGroup,
		Torrent:      *torrent,
	}

	// If the torrent is a batch, we don't need to set the episode
	if episodeNumber != -2 {
		ret.Episode = entities.NewMediaEntryEpisode(&entities.NewMediaEntryEpisodeOptions{
			LocalFile:            nil,
			OptionalAniDBEpisode: strconv.Itoa(episodeNumber),
			AnizipMedia:          anizipMedia,
			Media:                media,
			ProgressOffset:       0,
			IsDownloaded:         false,
		})
		if ret.Episode.IsInvalid { // remove invalid episodes
			return nil, false
		}
		ret.EpisodeNumber = lo.ToPtr(episodeNumber)
	}

	return ret, true

}
