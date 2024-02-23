package torrent

import (
	"errors"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/animetosho"
	"github.com/seanime-app/seanime/internal/anizip"
	"github.com/seanime-app/seanime/internal/comparison"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/nyaa"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/seanime-parser"
	"github.com/sourcegraph/conc/pool"
	"sort"
	"strconv"
)

type (
	SmartSearchQueryOptions struct {
		QuickSearch    *bool              `json:"quickSearch"`
		Query          *string            `json:"query"`
		EpisodeNumber  *int               `json:"episodeNumber"`
		Batch          *bool              `json:"batch"`
		Media          *anilist.BaseMedia `json:"media"`
		AbsoluteOffset *int               `json:"absoluteOffset"`
		Resolution     *string            `json:"resolution"`
		Provider       string             `json:"provider"`
	}
	SmartSearchOptions struct {
		SmartSearchQueryOptions
		//
		NyaaSearchCache       *nyaa.SearchCache
		AnimeToshoSearchCache *animetosho.SearchCache
		AnizipCache           *anizip.Cache
	}
	// Preview is used to preview a torrent à la entities.MediaEntryEpisode.
	Preview struct {
		Episode *entities.MediaEntryEpisode `json:"episode"` // nil if batch
		Torrent *AnimeTorrent               `json:"torrent"`
	}
	// SearchData is the struct returned by NewSmartSearch
	SearchData struct {
		Torrents []*AnimeTorrent `json:"torrents"` // Torrents found
		Previews []*Preview      `json:"previews"` // TorrentPreview for each torrent
	}
)

func NewSmartSearch(opts *SmartSearchOptions) (*SearchData, error) {

	if opts.QuickSearch == nil ||
		opts.Media == nil ||
		opts.Batch == nil ||
		opts.EpisodeNumber == nil ||
		opts.AbsoluteOffset == nil ||
		opts.Resolution == nil ||
		opts.Query == nil {
		return nil, errors.New("missing arguments")
	}

	if opts.Provider == "" {
		opts.Provider = "nyaa"
	}

	retTorrents := make([]*AnimeTorrent, 0)

	// +---------------------+
	// |        Nyaa         |
	// +---------------------+

	if opts.Provider == "nyaa" {

		// Use quick search if the user turned it on OR has not specified a query
		if *opts.QuickSearch || len(*opts.Query) == 0 {
			queries, ok := nyaa.BuildSearchQuery(&nyaa.BuildSearchQueryOptions{
				Media:          opts.Media,
				Batch:          opts.Batch,
				EpisodeNumber:  opts.EpisodeNumber,
				Resolution:     opts.Resolution,
				AbsoluteOffset: opts.AbsoluteOffset,
				Title:          opts.Query,
			})
			if !ok {
				return nil, errors.New("could not build search query")
			}

			// +---------------------+
			// |   Search multiple   |
			// +---------------------+

			res, err := nyaa.SearchMultiple(nyaa.SearchMultipleOptions{
				Provider: "nyaa",
				Query:    queries,
				Category: "anime-eng",
				SortBy:   "seeders",
				Filter:   "",
				Cache:    opts.NyaaSearchCache,
			})
			if err != nil {
				return nil, err
			}

			for _, torrent := range res {
				retTorrents = append(retTorrents, NewAnimeTorrentFromNyaa(torrent))
			}

		} else {

			// +---------------------+
			// |       Query         |
			// +---------------------+

			res, err := nyaa.Search(nyaa.SearchOptions{
				Provider: "nyaa",
				Query:    *opts.Query,
				Category: "anime-eng",
				SortBy:   "seeders",
				Filter:   "",
				Cache:    opts.NyaaSearchCache,
			})
			if err != nil {
				return nil, err
			}

			for _, torrent := range res {
				retTorrents = append(retTorrents, NewAnimeTorrentFromNyaa(torrent))
			}
		}
	} else if opts.Provider == "animetosho" {

		// +---------------------+
		// |     AnimeTosho      |
		// +---------------------+

		if *opts.QuickSearch || len(*opts.Query) == 0 {

			res, err := animetosho.SearchQuery(&animetosho.BuildSearchQueryOptions{
				Media:          opts.Media,
				Batch:          opts.Batch,
				EpisodeNumber:  opts.EpisodeNumber,
				Resolution:     opts.Resolution,
				AbsoluteOffset: opts.AbsoluteOffset,
				Title:          opts.Query,
				Cache:          opts.AnimeToshoSearchCache,
			})
			if err != nil {
				return nil, err
			}

			for _, torrent := range res {
				retTorrents = append(retTorrents, NewAnimeTorrentFromAnimeTosho(torrent))
			}
		} else {
			res, err := animetosho.Search(*opts.Query)
			if err != nil {
				return nil, err
			}
			for _, torrent := range res {
				retTorrents = append(retTorrents, NewAnimeTorrentFromAnimeTosho(torrent))
			}
		}

	}

	// +---------------------+
	// |    Anizip Cache     |
	// +---------------------+

	// Verify that cache has the AniZip media
	// Note: It should because it is fetched when the user accesses the media entry page
	_, ok := opts.AnizipCache.Get(anizip.GetCacheKey("anilist", opts.Media.ID))
	if !ok {
		_, err := anizip.FetchAniZipMediaC("anilist", opts.Media.ID, opts.AnizipCache)
		if err != nil {
			// No AniZip media found
			// We will just return the torrent previews without AniZip metadata
		}
	}

	// +---------------------+
	// |   Torrent Preview   |
	// +---------------------+

	// Create torrent previews in parallel
	p := pool.NewWithResults[*Preview]()
	for _, torrent := range retTorrents {
		torrent := torrent
		p.Go(func() *Preview {
			tp, ok := createTorrentPreview(opts.Media, opts.AnizipCache, torrent, *opts.AbsoluteOffset)
			if !ok {
				return nil
			}
			return tp
		})
	}
	previews := p.Wait()
	previews = lo.Filter(previews, func(i *Preview, _ int) bool {
		return i != nil
	})

	// +---------------------+
	// |      Sorting        |
	// +---------------------+

	// sort both by seeders
	sort.Slice(retTorrents, func(i, j int) bool {
		return retTorrents[i].Seeders > retTorrents[j].Seeders
	})
	sort.Slice(previews, func(i, j int) bool {
		return previews[i].Torrent.Seeders > previews[j].Torrent.Seeders
	})

	return &SearchData{
		Torrents: retTorrents,
		Previews: previews,
	}, nil
}

//----------------------------------------------------------------------------------------------------------------------

// createTorrentPreview creates a Preview from a torrent.
// It also uses the AniZip cache and the media to create the preview.
func createTorrentPreview(
	media *anilist.BaseMedia,
	anizipCache *anizip.Cache,
	torrent *AnimeTorrent,
	absoluteOffset int,
) (*Preview, bool) {

	anizipMedia, _ := anizipCache.Get(anizip.GetCacheKey("anilist", media.ID)) // can be nil

	elements := seanime_parser.Parse(torrent.Name)
	if len(elements.Title) == 0 {
		return nil, false
	}

	// -1 = error
	// -2 = batch
	torrent.EpisodeNumber = -1

	if len(elements.EpisodeNumber) == 1 {
		asInt, ok := util.StringToInt(elements.EpisodeNumber[0])
		if ok {
			torrent.EpisodeNumber = asInt
		}
	} else if len(elements.EpisodeNumber) > 1 {
		torrent.EpisodeNumber = -2
	}

	// Check if the torrent is a batch, if we still have no episode number
	if torrent.EpisodeNumber < 0 {
		if comparison.ValueContainsBatchKeywords(torrent.Name) {
			torrent.EpisodeNumber = -2
		}
	}

	// normalize episode number
	if torrent.EpisodeNumber >= 0 && torrent.EpisodeNumber > media.GetCurrentEpisodeCount() {
		torrent.EpisodeNumber = torrent.EpisodeNumber - absoluteOffset
	}

	if *media.GetFormat() == anilist.MediaFormatMovie {
		torrent.EpisodeNumber = 1
	}

	ret := &Preview{
		Torrent: torrent,
	}

	// If the torrent is a batch, we don't need to set the episode
	if torrent.EpisodeNumber != -2 {
		ret.Episode = entities.NewMediaEntryEpisode(&entities.NewMediaEntryEpisodeOptions{
			LocalFile:            nil,
			OptionalAniDBEpisode: strconv.Itoa(torrent.EpisodeNumber),
			AnizipMedia:          anizipMedia,
			Media:                media,
			ProgressOffset:       0,
			IsDownloaded:         false,
		})
		if ret.Episode.IsInvalid { // remove invalid episodes
			return nil, false
		}
	}

	return ret, true

}