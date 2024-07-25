package torrent

import (
	"cmp"
	"fmt"
	"github.com/samber/mo"
	"seanime/internal/api/anilist"
	"seanime/internal/api/anizip"
	"seanime/internal/library/anime"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"seanime/seanime-parser"
	"slices"
	"strconv"
	"sync"

	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
)

const (
	AnimeSearchTypeSmart  AnimeSearchType = "smart"
	AnimeSearchTypeSimple AnimeSearchType = "simple"
)

type (
	AnimeSearchType string

	AnimeSearchOptions struct {
		// Provider extension ID
		Provider string
		Type     AnimeSearchType
		Media    *anilist.BaseAnime
		// Search options
		Query string
		// Filter options
		Batch         bool
		EpisodeNumber int
		BestReleases  bool
		Resolution    string
	}

	// Preview contains the torrent and episode information
	Preview struct {
		Episode *anime.AnimeEntryEpisode    `json:"episode"` // nil if batch
		Torrent *hibiketorrent.AnimeTorrent `json:"torrent"`
	}
	// SearchData is the struct returned by NewSmartSearch
	SearchData struct {
		Torrents []*hibiketorrent.AnimeTorrent `json:"torrents"` // Torrents found
		Previews []*Preview                    `json:"previews"` // TorrentPreview for each torrent
	}
)

func (r *Repository) SearchAnime(opts AnimeSearchOptions) (ret *SearchData, err error) {
	defer util.HandlePanicInModuleWithError("torrents/torrent/SearchAnime", &err)

	r.logger.Debug().Str("provider", opts.Provider).Str("type", string(opts.Type)).Str("query", opts.Query).Msg("torrent repo: Searching for anime torrents")

	// Find the provider by ID
	providerExtension, ok := r.animeProviderExtensions.Get(opts.Provider)
	if !ok {
		// Get the default provider
		providerExtension, ok = r.GetDefaultAnimeProviderExtension()
		if !ok {
			return nil, fmt.Errorf("torrent provider not found")
		}
	}

	if opts.Type == AnimeSearchTypeSmart && !providerExtension.GetProvider().CanSmartSearch() {
		return nil, fmt.Errorf("provider does not support smart search")
	}

	var torrents []*hibiketorrent.AnimeTorrent

	// Fetch Anizip media
	anizipMedia := mo.None[*anizip.Media]()
	anizipMediaF, err := anizip.FetchAniZipMediaC("anilist", opts.Media.ID, r.anizipCache)
	if err == nil {
		anizipMedia = mo.Some(anizipMediaF)
	}

	queryMedia := hibiketorrent.Media{
		ID:                   opts.Media.GetID(),
		IDMal:                opts.Media.GetIDMal(),
		Status:               string(*opts.Media.GetStatus()),
		Format:               string(*opts.Media.GetFormat()),
		EnglishTitle:         opts.Media.GetTitle().GetEnglish(),
		RomajiTitle:          opts.Media.GetRomajiTitleSafe(),
		EpisodeCount:         opts.Media.GetTotalEpisodeCount(),
		AbsoluteSeasonOffset: 0,
		Synonyms:             opts.Media.GetSynonymsContainingSeason(),
		IsAdult:              *opts.Media.GetIsAdult(),
		StartDate: &hibiketorrent.FuzzyDate{
			Year:  *opts.Media.GetStartDate().GetYear(),
			Month: opts.Media.GetStartDate().GetMonth(),
			Day:   opts.Media.GetStartDate().GetDay(),
		},
	}

	//// Force simple search if AniZip media is absent
	//if opts.Type == AnimeSearchTypeSmart && anizipMedia.IsAbsent() {
	//	opts.Type = AnimeSearchTypeSimple
	//}

	var queryKey string

	switch opts.Type {
	case AnimeSearchTypeSmart:
		anidbAID := 0
		anidbEID := 0

		// Get the AniDB Anime ID and Episode ID
		if anizipMedia.IsPresent() {
			// Override absolute offset value of queryMedia
			queryMedia.AbsoluteSeasonOffset = anizipMedia.MustGet().GetOffset()

			if anizipMedia.MustGet().GetMappings() != nil {

				anidbAID = anizipMedia.MustGet().GetMappings().AnidbID
				// Find Anizip Episode based on inputted episode number
				anizipEpisode, found := anizipMedia.MustGet().FindEpisode(strconv.Itoa(opts.EpisodeNumber))
				if found {
					anidbEID = anizipEpisode.AnidbEid
				}
			}
		}

		queryKey = fmt.Sprintf("%d-%s-%d-%d-%d-%s-%t-%t", opts.Media.GetID(), opts.Query, opts.EpisodeNumber, anidbAID, anidbEID, opts.Resolution, opts.BestReleases, opts.Batch)
		if cache, found := r.animeProviderSmartSearchCaches.Get(opts.Provider); found {
			// Check the cache
			data, found := cache.Get(queryKey)
			if found {
				r.logger.Debug().Str("provider", opts.Provider).Str("type", string(opts.Type)).Msg("torrent repo: Cache HIT")
				return data, nil
			}
		}

		torrents, err = providerExtension.GetProvider().SmartSearch(hibiketorrent.AnimeSmartSearchOptions{
			Media:         queryMedia,
			Query:         opts.Query,
			Batch:         opts.Batch,
			EpisodeNumber: opts.EpisodeNumber,
			Resolution:    opts.Resolution,
			AniDbAID:      anidbAID,
			AniDbEID:      anidbEID,
			BestReleases:  opts.BestReleases,
		})

	case AnimeSearchTypeSimple:

		queryKey = fmt.Sprintf("%d-%s", opts.Media.GetID(), opts.Query)
		if cache, found := r.animeProviderSearchCaches.Get(opts.Provider); found {
			// Check the cache
			data, found := cache.Get(queryKey)
			if found {
				r.logger.Debug().Str("provider", opts.Provider).Str("type", string(opts.Type)).Msg("torrent repo: Cache HIT")
				return data, nil
			}
		}

		torrents, err = providerExtension.GetProvider().Search(hibiketorrent.AnimeSearchOptions{
			Media: queryMedia,
			Query: opts.Query,
		})
	}
	if err != nil {
		return nil, err
	}

	// Add preview for smart search
	previews := make([]*Preview, 0)

	if opts.Type == AnimeSearchTypeSmart {

		wg := sync.WaitGroup{}
		wg.Add(len(torrents))
		for _, t := range torrents {
			go func(t *hibiketorrent.AnimeTorrent) {
				defer wg.Done()

				preview := r.createAnimeTorrentPreview(createAnimeTorrentPreviewOptions{
					torrent:     t,
					media:       opts.Media,
					anizipMedia: anizipMedia,
					searchOpts:  &opts,
				})
				if preview != nil {
					previews = append(previews, preview)
				}
			}(t)
		}
		wg.Wait()

	}

	// sort both by seeders
	slices.SortFunc(torrents, func(i, j *hibiketorrent.AnimeTorrent) int {
		return cmp.Compare(j.Seeders, i.Seeders)
	})
	slices.SortFunc(previews, func(i, j *Preview) int {
		return cmp.Compare(j.Torrent.Seeders, i.Torrent.Seeders)
	})

	ret = &SearchData{
		Torrents: torrents,
		Previews: previews,
	}

	// Store the data in the cache
	switch opts.Type {
	case AnimeSearchTypeSmart:
		if cache, found := r.animeProviderSmartSearchCaches.Get(opts.Provider); found {
			cache.Set(queryKey, ret)
		}
	case AnimeSearchTypeSimple:
		if cache, found := r.animeProviderSearchCaches.Get(opts.Provider); found {
			cache.Set(queryKey, ret)
		}
	}

	return
}

type createAnimeTorrentPreviewOptions struct {
	torrent     *hibiketorrent.AnimeTorrent
	media       *anilist.BaseAnime
	anizipMedia mo.Option[*anizip.Media]
	searchOpts  *AnimeSearchOptions
}

func (r *Repository) createAnimeTorrentPreview(opts createAnimeTorrentPreviewOptions) *Preview {

	parsedData := seanime_parser.Parse(opts.torrent.Name)

	isBatch := opts.torrent.IsBestRelease ||
		opts.torrent.IsBatch ||
		comparison.ValueContainsBatchKeywords(opts.torrent.Name) || // Contains batch keywords
		(!opts.media.IsMovieOrSingleEpisode() && len(parsedData.EpisodeNumber) > 1) // Multiple episodes parsed & not a movie

	if isBatch {
		return &Preview{
			Episode: nil, // Will be displayed as batch
			Torrent: opts.torrent,
		}
	}

	// If past this point we haven't detected a batch but the episode number returned from the provider is -1
	// we will parse it from the torrent name
	if opts.torrent.EpisodeNumber == -1 && len(parsedData.EpisodeNumber) == 1 {
		opts.torrent.EpisodeNumber = util.StringToIntMust(parsedData.EpisodeNumber[0])
	}

	// If the torrent is confirmed, use the episode number from the search options
	// because it could be absolute
	if opts.torrent.Confirmed {
		opts.torrent.EpisodeNumber = opts.searchOpts.EpisodeNumber
	}

	// If there was no single episode number parsed but the media is movie, set the episode number to 1
	if opts.torrent.EpisodeNumber == -1 && opts.media.IsMovieOrSingleEpisode() {
		opts.torrent.EpisodeNumber = 1
	}

	if opts.anizipMedia.IsPresent() {

		// normalize episode number
		if opts.torrent.EpisodeNumber >= 0 && opts.torrent.EpisodeNumber > opts.media.GetCurrentEpisodeCount() {
			opts.torrent.EpisodeNumber = opts.torrent.EpisodeNumber - opts.anizipMedia.MustGet().GetOffset()
		}

		anizipMedia := opts.anizipMedia.MustGet()
		_, foundEp := anizipMedia.FindEpisode(strconv.Itoa(opts.searchOpts.EpisodeNumber))

		if foundEp {
			episode := anime.NewAnimeEntryEpisode(&anime.NewAnimeEntryEpisodeOptions{
				LocalFile:            nil,
				OptionalAniDBEpisode: strconv.Itoa(opts.torrent.EpisodeNumber),
				AnizipMedia:          anizipMedia,
				Media:                opts.media,
				ProgressOffset:       0,
				IsDownloaded:         false,
				MetadataProvider:     r.metadataProvider,
			})
			episode.IsInvalid = false

			return &Preview{
				Episode: episode,
				Torrent: opts.torrent,
			}
		}

		// If the episode number could not be found in the AniZip media, create a new episode
		episode := anime.AnimeEntryEpisode{
			Type:                  anime.LocalFileTypeMain,
			DisplayTitle:          fmt.Sprintf("Episode %d", opts.searchOpts.EpisodeNumber),
			EpisodeTitle:          "",
			EpisodeNumber:         opts.searchOpts.EpisodeNumber,
			ProgressNumber:        opts.searchOpts.EpisodeNumber,
			AniDBEpisode:          "",
			AbsoluteEpisodeNumber: 0,
			LocalFile:             nil,
			IsDownloaded:          false,
			EpisodeMetadata:       anime.NewEpisodeMetadata(opts.anizipMedia.MustGet(), nil, opts.media, r.metadataProvider),
			FileMetadata:          nil,
			IsInvalid:             false,
			MetadataIssue:         "",
			BaseAnime:             opts.media,
		}

		return &Preview{
			Episode: &episode,
			Torrent: opts.torrent,
		}

	}

	return &Preview{
		Episode: nil,
		Torrent: opts.torrent,
	}
}
