package torrent

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/api/metadata"
	"github.com/seanime-app/seanime/internal/library/anime"
	"github.com/seanime-app/seanime/internal/torrents/animetosho"
	"github.com/seanime-app/seanime/internal/torrents/nyaa"
	"github.com/seanime-app/seanime/internal/torrents/seadex"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/internal/util/comparison"
	"github.com/seanime-app/seanime/seanime-parser"
	"github.com/sourcegraph/conc/pool"
	"sort"
	"strconv"
)

// Smart Search is a service that searches for torrents based on the media and the user's query.
// It is implemented for Nyaa and AnimeTosho.

type (
	SmartSearchQueryOptions struct {
		SmartSearch    *bool              `json:"smartSearch"`
		Query          *string            `json:"query"`
		EpisodeNumber  *int               `json:"episodeNumber"`
		Batch          *bool              `json:"batch"`
		Media          *anilist.BaseMedia `json:"media"`
		AbsoluteOffset *int               `json:"absoluteOffset"`
		Resolution     *string            `json:"resolution"`
		Best           *bool              `json:"best"`
		Provider       string             `json:"provider"`
	}
	SmartSearchOptions struct {
		SmartSearchQueryOptions
		NyaaSearchCache       *nyaa.SearchCache
		AnimeToshoSearchCache *animetosho.SearchCache
		AnizipCache           *anizip.Cache
		Logger                *zerolog.Logger
		MetadataProvider      *metadata.Provider
	}
	// Preview is used to preview a torrent à la anime.MediaEntryEpisode.
	Preview struct {
		Episode *anime.MediaEntryEpisode `json:"episode"` // nil if batch
		Torrent *AnimeTorrent            `json:"torrent"`
	}
	// SearchData is the struct returned by NewSmartSearch
	SearchData struct {
		Torrents []*AnimeTorrent `json:"torrents"` // Torrents found
		Previews []*Preview      `json:"previews"` // TorrentPreview for each torrent
	}
)

func NewSmartSearch(opts *SmartSearchOptions) (*SearchData, error) {

	if opts.SmartSearch == nil ||
		opts.Media == nil ||
		opts.Batch == nil ||
		opts.EpisodeNumber == nil ||
		opts.AbsoluteOffset == nil ||
		opts.Resolution == nil ||
		opts.MetadataProvider == nil ||
		opts.Best == nil ||
		opts.Query == nil {
		return nil, errors.New("missing arguments")
	}

	// Default to AnimeTosho if the provider is not specified or is none
	if opts.Provider == "" || opts.Provider == ProviderNone {
		opts.Provider = ProviderAnimeTosho
	}

	retTorrents := make([]*AnimeTorrent, 0)

	// +---------------------+
	// |    Anizip Cache     |
	// +---------------------+

	// Verify that cache has the AniZip media
	// Note: It should because it is fetched when the user accesses the media entry page
	anizipMedia, found := opts.AnizipCache.Get(anizip.GetCacheKey("anilist", opts.Media.ID))
	if !found {
		var err error
		anizipMedia, err = anizip.FetchAniZipMediaC("anilist", opts.Media.ID, opts.AnizipCache)
		if err != nil {
			// No AniZip media found
			// We will just return the torrent previews without AniZip metadata
		}
	}

	var episodesFoundByID bool

	if *opts.Best {

		torrents, err := seadex.New(opts.Logger).FetchTorrents(opts.Media.ID, opts.Media.GetRomajiTitleSafe())
		if err != nil {
			return nil, err
		}

		for _, torrent := range torrents {
			retTorrents = append(retTorrents, NewAnimeTorrentFromSeaDex(torrent))
		}

	} else if opts.Provider == "nyaa" {
		// +---------------------+
		// |        Nyaa         |
		// +---------------------+

		// Use quick search if the user turned it on OR has not specified a query
		if *opts.SmartSearch || len(*opts.Query) == 0 {
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

		if *opts.SmartSearch || len(*opts.Query) == 0 {

			animetoshoTorrents := make([]*animetosho.Torrent, 0)
			var err error

			// Search by EID
			if anizipMedia != nil && !*opts.Batch {

				opts.Logger.Debug().Int("eid", anizipMedia.Mappings.AnidbID).Msg("smartsearch: Searching by EID (AnimeTosho)")

				episode, foundEp := anizipMedia.FindEpisode(strconv.Itoa(*opts.EpisodeNumber))

				if foundEp && episode.AnidbEid > 0 {
					animetoshoTorrents, err = animetosho.SearchByEID(episode.AnidbEid, *opts.Resolution)
					episodesFoundByID = true
				}

			}

			// Could not find by EID, search by query
			if err != nil || len(animetoshoTorrents) == 0 {

				opts.Logger.Debug().Str("query", *opts.Query).Msg("smartsearch: Searching by query (AnimeTosho)")

				animetoshoTorrents, err = animetosho.SearchQuery(&animetosho.BuildSearchQueryOptions{
					Media:          opts.Media,
					Batch:          opts.Batch,
					EpisodeNumber:  opts.EpisodeNumber,
					Resolution:     opts.Resolution,
					AbsoluteOffset: opts.AbsoluteOffset,
					Title:          opts.Query,
					Cache:          opts.AnimeToshoSearchCache,
					Logger:         opts.Logger,
				})

				// If we still have no torrents, and the user wants to batch search, we will search by AID
				if anizipMedia != nil && *opts.Batch && (err != nil || len(animetoshoTorrents) == 0) {

					opts.Logger.Debug().Int("aid", anizipMedia.Mappings.AnidbID).Msg("smartsearch: Searching by AID (AnimeTosho)")

					animetoshoTorrents, err = animetosho.SearchByAIDLikelyBatch(anizipMedia.Mappings.AnidbID, *opts.Resolution)
					// Filter if found
					if err == nil && len(animetoshoTorrents) > 0 {
						newAT := make([]*animetosho.Torrent, 0)
						for _, t := range animetoshoTorrents {
							m := seanime_parser.Parse(t.Title)
							if len(m.EpisodeNumber) < 2 && !comparison.ValueContainsBatchKeywords(t.Title) {
								continue
							}
							newAT = append(newAT, t)
						}
						animetoshoTorrents = newAT
						episodesFoundByID = true
					}

				}
			}

			if err != nil {
				return nil, err
			}

			for _, torrent := range animetoshoTorrents {
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
	// |   Torrent Preview   |
	// +---------------------+

	// Create torrent previews in parallel
	p := pool.NewWithResults[*Preview]()
	for _, torrent := range retTorrents {
		p.Go(func() *Preview {
			tp, ok := createTorrentPreview(
				opts.Media,
				opts.MetadataProvider,
				opts.AnizipCache,
				torrent,
				*opts.AbsoluteOffset,
				episodesFoundByID,
				*opts.EpisodeNumber,
				*opts.Best,
			)
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
	metadataProvider *metadata.Provider,
	anizipCache *anizip.Cache,
	torrent *AnimeTorrent,
	absoluteOffset int,
	foundByID bool,
	searchEpNumber int,
	isBestRelease bool,
) (*Preview, bool) {

	anizipMedia, _ := anizipCache.Get(anizip.GetCacheKey("anilist", media.ID)) // can be nil

	if isBestRelease && anizipMedia != nil {
		ep := anime.NewMediaEntryEpisode(&anime.NewMediaEntryEpisodeOptions{
			LocalFile:            nil,
			OptionalAniDBEpisode: strconv.Itoa(torrent.EpisodeNumber),
			AnizipMedia:          anizipMedia,
			Media:                media,
			ProgressOffset:       0,
			IsDownloaded:         false,
			MetadataProvider:     metadataProvider,
		})
		ep.DisplayTitle = media.GetRomajiTitleSafe()
		ep.EpisodeMetadata = anime.NewEpisodeMetadata(anizipMedia, nil, media, metadataProvider)
		ep.EpisodeTitle = media.GetRomajiTitleSafe()
		return &Preview{
			Torrent: torrent,
			Episode: ep,
		}, true
	} else if isBestRelease {
		return nil, false
	}

	elements := seanime_parser.Parse(torrent.Name)
	if len(elements.Title) == 0 {
		return nil, false
	}

	// -1 = error
	// -2 = batch
	torrent.EpisodeNumber = -1

	// Check if the torrent is a batch, if we still have no episode number
	if comparison.ValueContainsBatchKeywords(torrent.Name) {
		torrent.EpisodeNumber = -2
	} else if foundByID {
		// If we found the episode number by ID, we don't need to parse the title
		torrent.EpisodeNumber = searchEpNumber

	} else if !foundByID {
		// Torrent isn't a batch, but we couldn't find the torrent by ID, so we need to parse the title
		if len(elements.EpisodeNumber) == 1 {
			asInt, ok := util.StringToInt(elements.EpisodeNumber[0])
			if ok {
				torrent.EpisodeNumber = asInt
			}
		} else if len(elements.EpisodeNumber) > 1 {
			torrent.EpisodeNumber = -2
		}

		// normalize episode number
		if torrent.EpisodeNumber >= 0 && torrent.EpisodeNumber > media.GetCurrentEpisodeCount() {
			torrent.EpisodeNumber = torrent.EpisodeNumber - absoluteOffset
		}
	}

	if *media.GetFormat() == anilist.MediaFormatMovie {
		torrent.EpisodeNumber = 1
	}

	ret := &Preview{
		Torrent: torrent,
	}

	// If the torrent is a batch, we don't need to set the episode
	if torrent.EpisodeNumber != -2 {
		ret.Episode = anime.NewMediaEntryEpisode(&anime.NewMediaEntryEpisodeOptions{
			LocalFile:            nil,
			OptionalAniDBEpisode: strconv.Itoa(torrent.EpisodeNumber),
			AnizipMedia:          anizipMedia,
			Media:                media,
			ProgressOffset:       0,
			IsDownloaded:         false,
			MetadataProvider:     metadataProvider,
		})
		if ret.Episode.IsInvalid { // remove invalid episodes
			return nil, false
		}
	}

	return ret, true

}
