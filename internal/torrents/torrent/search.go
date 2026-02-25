package torrent

import (
	"cmp"
	"context"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/debrid/debrid"
	"seanime/internal/extension"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/library/anime"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"seanime/internal/util/result"
	"slices"
	"strconv"
	"sync"

	"github.com/5rahim/habari"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

const (
	AnimeSearchTypeSmart  AnimeSearchType = "smart"
	AnimeSearchTypeSimple AnimeSearchType = "simple"
)

var (
	metadataCache = result.NewBoundedCache[string, *TorrentMetadata](100)
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
		Batch                   bool
		EpisodeNumber           int
		BestReleases            bool
		Resolution              string
		IncludeSpecialProviders bool
		SkipPreviews            bool
	}

	// Preview contains the torrent and episode information
	Preview struct {
		Episode *anime.Episode              `json:"episode"` // nil if batch
		Torrent *hibiketorrent.AnimeTorrent `json:"torrent"`
	}

	TorrentMetadata struct {
		Distance int              `json:"distance"`
		Metadata *habari.Metadata `json:"metadata"`
	}

	// SearchData is the struct returned by NewSmartSearch
	SearchData struct {
		Torrents                  []*hibiketorrent.AnimeTorrent                    `json:"torrents"`                  // Torrents found
		Previews                  []*Preview                                       `json:"previews"`                  // TorrentPreview for each torrent
		TorrentMetadata           map[string]*TorrentMetadata                      `json:"torrentMetadata"`           // Torrent metadata
		DebridInstantAvailability map[string]debrid.TorrentItemInstantAvailability `json:"debridInstantAvailability"` // Debrid instant availability
		AnimeMetadata             *metadata.AnimeMetadata                          `json:"animeMetadata"`             // Animap media
		IncludedSpecialProviders  []string                                         `json:"includedSpecialProviders"`
	}
)

func (r *Repository) SearchAnime(ctx context.Context, opts AnimeSearchOptions) (ret *SearchData, retErr error) {
	defer util.HandlePanicInModuleWithError("torrents/torrent/SearchAnime", &retErr)
	var torrents []*hibiketorrent.AnimeTorrent

	// Find the provider by ID
	providerExtension, ok := extension.GetExtension[extension.AnimeTorrentProviderExtension](r.extensionBankRef.Get(), opts.Provider)
	if !ok {
		providerExtension, ok = r.GetDefaultAnimeProviderExtension()
		if !ok {
			return nil, fmt.Errorf("torrent provider not found")
		}
	}

	providers := []extension.AnimeTorrentProviderExtension{providerExtension}
	includedSpecialProviders := make([]string, 0)

	// Add other special providers if requested
	if opts.IncludeSpecialProviders && opts.Type == AnimeSearchTypeSmart && opts.Batch == true {
		specialAdded := 0
		specialProviders := r.GetAnimeProviderExtensionsBy(func(ext extension.AnimeTorrentProviderExtension) bool {
			if specialAdded >= 2 {
				return false
			}
			if opts.Provider == ext.GetID() {
				return false
			}
			if ext.GetProvider().GetSettings().Type == hibiketorrent.AnimeProviderTypeSpecial {
				specialAdded++
				includedSpecialProviders = append(includedSpecialProviders, ext.GetID())
				return true
			}
			return false
		})
		providers = append(providers, specialProviders...)
	}

	r.logger.Debug().Str("provider", opts.Provider).Interface("specialProviders", includedSpecialProviders).Msg("torrent search: Searching for anime torrents")

	// Fetch Animap media, this is cached
	animeMetadata := mo.None[*metadata.AnimeMetadata]()
	animeMetadataF, errM := r.metadataProviderRef.Get().GetAnimeMetadata(metadata.AnilistPlatform, opts.Media.GetID())
	if errM == nil {
		animeMetadata = mo.Some(animeMetadataF)
	}

	status := anilist.MediaStatusNotYetReleased
	if opts.Media.GetStatus() != nil {
		status = *opts.Media.GetStatus()
	}
	format := anilist.MediaFormatTv
	if opts.Media.GetFormat() != nil {
		format = *opts.Media.GetFormat()
	}
	var year int
	if opts.Media.GetStartDate() != nil && opts.Media.GetStartDate().GetYear() != nil {
		year = *opts.Media.GetStartDate().GetYear()
	}

	queryMedia := hibiketorrent.Media{
		ID:                   opts.Media.GetID(),
		IDMal:                opts.Media.GetIDMal(),
		Status:               string(status),
		Format:               string(format),
		EnglishTitle:         opts.Media.GetTitle().GetEnglish(),
		RomajiTitle:          opts.Media.GetRomajiTitleSafe(),
		EpisodeCount:         opts.Media.GetTotalEpisodeCount(),
		AbsoluteSeasonOffset: 0,
		Synonyms:             opts.Media.GetSynonymsContainingSeason(),
		IsAdult:              *opts.Media.GetIsAdult(),
		StartDate: &hibiketorrent.FuzzyDate{
			Year:  year,
			Month: opts.Media.GetStartDate().GetMonth(),
			Day:   opts.Media.GetStartDate().GetDay(),
		},
	}

	var mainErr error

	smartQueryKey := fmt.Sprintf("%d-%s-%d-%s-%t-%t", opts.Media.GetID(), opts.Query, opts.EpisodeNumber, opts.Resolution, opts.BestReleases, opts.Batch)
	simpleQueryKey := fmt.Sprintf("%d-%s", opts.Media.GetID(), opts.Query)

	if opts.Type == AnimeSearchTypeSmart {
		if cache, found := r.animeProviderSmartSearchCaches.Get(opts.Provider); found {
			// Check the cache
			data, found := cache.Get(smartQueryKey)
			if found {
				r.logger.Debug().Str("provider", opts.Provider).Str("type", string(opts.Type)).Msg("torrent search: Cache HIT")
				return data, nil
			}
		}
	} else if opts.Type == AnimeSearchTypeSimple {
		if cache, found := r.animeProviderSearchCaches.Get(opts.Provider); found {
			// Check the cache
			data, found := cache.Get(simpleQueryKey)
			if found {
				r.logger.Debug().Str("provider", opts.Provider).Str("type", string(opts.Type)).Msg("torrent search: Cache HIT")
				return data, nil
			}
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(len(providers))
	mu := sync.Mutex{}
	for i, provider := range providers {
		go func() {
			defer util.HandlePanicInModuleThen("torrents/torrent/SearchAnime", func() {})
			defer wg.Done()

			isMain := i == 0
			r.logger.Debug().Str("provider", provider.GetID()).Str("type", string(opts.Type)).Str("query", opts.Query).Msg("torrent search: Searching for anime torrents")

			searchType := opts.Type
			if isMain && opts.Type == AnimeSearchTypeSmart && !provider.GetProvider().GetSettings().CanSmartSearch {
				mainErr = fmt.Errorf("provider %s does not support smart search", provider.GetID())
				return
			}
			if !isMain && opts.Type == AnimeSearchTypeSmart && !provider.GetProvider().GetSettings().CanSmartSearch {
				searchType = AnimeSearchTypeSimple
			}

			//// Force simple search if Animap media is absent
			//if opts.Type == AnimeSearchTypeSmart && animeMetadata.IsAbsent() {
			//	opts.Type = AnimeSearchTypeSimple
			//}

			switch searchType {
			case AnimeSearchTypeSmart:
				anidbAID := 0
				anidbEID := 0

				// Get the AniDB Anime ID and Episode ID
				if animeMetadata.IsPresent() {
					// Override absolute offset value of queryMedia
					queryMedia.AbsoluteSeasonOffset = animeMetadata.MustGet().GetOffset()

					if animeMetadata.MustGet().GetMappings() != nil {

						anidbAID = animeMetadata.MustGet().GetMappings().AnidbId
						// Find Animap Episode based on requested episode number
						episodeMetadata, found := animeMetadata.MustGet().FindEpisode(strconv.Itoa(opts.EpisodeNumber))
						if found {
							anidbEID = episodeMetadata.AnidbEid
						}
					}
				}

				// Check for context cancellation before making the request
				select {
				case <-ctx.Done():
					return
				default:
				}

				res, err := provider.GetProvider().SmartSearch(hibiketorrent.AnimeSmartSearchOptions{
					Media:         queryMedia,
					Query:         opts.Query,
					Batch:         opts.Batch,
					EpisodeNumber: opts.EpisodeNumber,
					Resolution:    opts.Resolution,
					AnidbAID:      anidbAID,
					AnidbEID:      anidbEID,
					BestReleases:  opts.BestReleases,
				})
				if err != nil {
					if isMain {
						mainErr = err
					}
					return
				}

				mu.Lock()
				r.logger.Debug().Str("provider", provider.GetID()).Int("found", len(res)).Msg("torrent search: Found torrents")
				torrents = append(torrents, res...)
				mu.Unlock()

				return
			case AnimeSearchTypeSimple:

				// Check for context cancellation before making the request
				select {
				case <-ctx.Done():
					return
				default:
				}

				res, err := provider.GetProvider().Search(hibiketorrent.AnimeSearchOptions{
					Media: queryMedia,
					Query: opts.Query,
				})
				if err != nil {
					if isMain {
						mainErr = err
					}
					return
				}

				mu.Lock()
				r.logger.Debug().Str("provider", provider.GetID()).Int("found", len(res)).Msg("torrent search: Found torrents")
				torrents = append(torrents, res...)
				mu.Unlock()

				return
			}

		}()
	}
	wg.Wait()

	if mainErr != nil {
		return nil, mainErr
	}

	// Place best torrents on top, deduplicate
	bestReleases := make([]*hibiketorrent.AnimeTorrent, 0)
	other := make([]*hibiketorrent.AnimeTorrent, 0)
	for _, t := range torrents {
		if t.InfoHash == "" { // make sure it's never empty
			t.InfoHash = t.Name
		}
		if t.IsBestRelease {
			bestReleases = append(bestReleases, t)
		} else {
			other = append(other, t)
		}
	}
	torrents = append(bestReleases, other...)

	torrents = lo.UniqBy(torrents, func(t *hibiketorrent.AnimeTorrent) string {
		return t.InfoHash
	})

	// Parse all torrents
	torrentMetadata := make(map[string]*TorrentMetadata)
	wg.Add(len(torrents))
	for _, t := range torrents {
		go func(t *hibiketorrent.AnimeTorrent) {
			defer wg.Done()
			tMetadata, found := metadataCache.Get(t.Name)
			if !found {
				m := habari.Parse(t.Name)
				var distance *comparison.LevenshteinResult
				distance, ok := comparison.FindBestMatchWithLevenshtein(&m.Title, opts.Media.GetAllTitles())
				if !ok {
					distance = &comparison.LevenshteinResult{
						Distance: 1000,
					}
				}
				tMetadata = &TorrentMetadata{
					Distance: distance.Distance,
					Metadata: m,
				}
				metadataCache.Set(t.Name, tMetadata)
			}
			mu.Lock()
			torrentMetadata[t.InfoHash] = tMetadata
			mu.Unlock()
		}(t)
	}
	wg.Wait()

	//
	// Previews
	//
	previews := make([]*Preview, 0)

	if opts.Type == AnimeSearchTypeSmart && !opts.SkipPreviews {
		wg := sync.WaitGroup{}
		wg.Add(len(torrents))
		mu := sync.Mutex{}
		for _, t := range torrents {
			go func(t *hibiketorrent.AnimeTorrent) {
				defer wg.Done()

				// Check for context cancellation in each goroutine
				select {
				case <-ctx.Done():
					return
				default:
				}

				preview := r.createAnimeTorrentPreview(createAnimeTorrentPreviewOptions{
					torrent:       t,
					media:         opts.Media,
					animeMetadata: animeMetadata,
					searchOpts:    &opts,
				})
				if preview != nil {
					mu.Lock()
					previews = append(previews, preview)
					mu.Unlock()
				}
			}(t)
		}
		wg.Wait()

		// Check if context was cancelled during preview creation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}

	// sort both by seeders, put best releases on top
	slices.SortFunc(torrents, func(i, j *hibiketorrent.AnimeTorrent) int {
		if i.IsBestRelease != j.IsBestRelease {
			if i.IsBestRelease {
				return -1
			}
			return 1
		}
		return cmp.Compare(j.Seeders, i.Seeders)
	})
	previews = lo.Filter(previews, func(p *Preview, _ int) bool {
		return p != nil && p.Torrent != nil
	})
	slices.SortFunc(previews, func(i, j *Preview) int {
		if i.Torrent.IsBestRelease != j.Torrent.IsBestRelease {
			if i.Torrent.IsBestRelease {
				return -1
			}
			return 1
		}
		return cmp.Compare(j.Torrent.Seeders, i.Torrent.Seeders)
	})

	ret = &SearchData{
		Torrents:                 torrents,
		Previews:                 previews,
		TorrentMetadata:          torrentMetadata,
		IncludedSpecialProviders: includedSpecialProviders,
	}

	if animeMetadata.IsPresent() {
		ret.AnimeMetadata = animeMetadata.MustGet()
	}

	// Store the data in the cache
	switch opts.Type {
	case AnimeSearchTypeSmart:
		if cache, found := r.animeProviderSmartSearchCaches.Get(opts.Provider); found {
			cache.Set(smartQueryKey, ret)
		}
	case AnimeSearchTypeSimple:
		if cache, found := r.animeProviderSearchCaches.Get(opts.Provider); found {
			cache.Set(simpleQueryKey, ret)
		}
	}

	return
}

type createAnimeTorrentPreviewOptions struct {
	torrent       *hibiketorrent.AnimeTorrent
	media         *anilist.BaseAnime
	animeMetadata mo.Option[*metadata.AnimeMetadata]
	searchOpts    *AnimeSearchOptions
}

func (r *Repository) createAnimeTorrentPreview(opts createAnimeTorrentPreviewOptions) *Preview {
	defer util.HandlePanicInModuleThen("torrents/torrent/createAnimeTorrentPreview", func() {})

	var parsedData *habari.Metadata
	tMetadata, found := metadataCache.Get(opts.torrent.Name)
	if !found { // Should always be found
		parsedData = habari.Parse(opts.torrent.Name)
		metadataCache.Set(opts.torrent.Name, &TorrentMetadata{
			Distance: 1000,
			Metadata: parsedData,
		})
	}
	parsedData = tMetadata.Metadata

	isBatch := opts.torrent.IsBestRelease ||
		opts.torrent.IsBatch ||
		//comparison.ValueContainsBatchKeywords(opts.torrent.Name) || // Contains batch keywords
		(!opts.media.IsMovieOrSingleEpisode() && (len(parsedData.EpisodeNumber) > 1 || len(parsedData.EpisodeNumber) == 0)) // Multiple episodes parsed & not a movie

	if isBatch && !opts.torrent.IsBatch {
		opts.torrent.IsBatch = true
	}

	if opts.torrent.ReleaseGroup == "" {
		opts.torrent.ReleaseGroup = parsedData.ReleaseGroup
	}

	if opts.torrent.Resolution == "" {
		opts.torrent.Resolution = parsedData.VideoResolution
	}

	if opts.torrent.FormattedSize == "" {
		opts.torrent.FormattedSize = util.Bytes(uint64(opts.torrent.Size))
	}

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

	if opts.animeMetadata.IsPresent() {

		// normalize episode number
		if opts.torrent.EpisodeNumber >= 0 && opts.torrent.EpisodeNumber > opts.media.GetCurrentEpisodeCount() {
			opts.torrent.EpisodeNumber = opts.torrent.EpisodeNumber - opts.animeMetadata.MustGet().GetOffset()
		}

		animeMetadata := opts.animeMetadata.MustGet()
		_, foundEp := animeMetadata.FindEpisode(strconv.Itoa(opts.searchOpts.EpisodeNumber))

		amw := r.metadataProviderRef.Get().GetAnimeMetadataWrapper(opts.media, animeMetadata)

		if foundEp {
			var episode *anime.Episode

			// Remove the episode if the parsed episode number is not the same as the search option
			if isProbablySameEpisode(parsedData.EpisodeNumber, opts.searchOpts.EpisodeNumber, opts.animeMetadata.MustGet().GetOffset()) {
				ep := opts.searchOpts.EpisodeNumber
				episode = anime.NewEpisode(&anime.NewEpisodeOptions{
					LocalFile:            nil,
					OptionalAniDBEpisode: strconv.Itoa(ep),
					AnimeMetadata:        animeMetadata,
					Media:                opts.media,
					ProgressOffset:       0,
					IsDownloaded:         false,
					MetadataProvider:     r.metadataProviderRef.Get(),
					MetadataWrapper:      amw,
				})
				episode.IsInvalid = false

				if episode.DisplayTitle == "" {
					episode.DisplayTitle = parsedData.Title
				}
			}

			return &Preview{
				Episode: episode,
				Torrent: opts.torrent,
			}
		}

		var episode *anime.Episode

		// Remove the episode if the parsed episode number is not the same as the search option
		if isProbablySameEpisode(parsedData.EpisodeNumber, opts.searchOpts.EpisodeNumber, opts.animeMetadata.MustGet().GetOffset()) {
			displayTitle := ""
			if len(parsedData.EpisodeNumber) == 1 && parsedData.EpisodeNumber[0] != strconv.Itoa(opts.searchOpts.EpisodeNumber) {
				displayTitle = fmt.Sprintf("Episode %s", parsedData.EpisodeNumber[0])
			}
			// If the episode number could not be found in the Animap media, create a new episode
			episode = &anime.Episode{
				Type:                  anime.LocalFileTypeMain,
				DisplayTitle:          displayTitle,
				EpisodeTitle:          "",
				EpisodeNumber:         opts.searchOpts.EpisodeNumber,
				ProgressNumber:        opts.searchOpts.EpisodeNumber,
				AniDBEpisode:          "",
				AbsoluteEpisodeNumber: 0,
				LocalFile:             nil,
				IsDownloaded:          false,
				EpisodeMetadata:       anime.NewEpisodeMetadata(amw, nil, strconv.Itoa(opts.searchOpts.EpisodeNumber), opts.media),
				FileMetadata:          nil,
				IsInvalid:             false,
				MetadataIssue:         "",
				BaseAnime:             opts.media,
			}
		}

		return &Preview{
			Episode: episode,
			Torrent: opts.torrent,
		}

	}

	return &Preview{
		Episode: nil,
		Torrent: opts.torrent,
	}
}

func isProbablySameEpisode(parsedEpisode []string, searchEpisode int, absoluteOffset int) bool {
	if len(parsedEpisode) == 1 {
		if util.StringToIntMust(parsedEpisode[0]) == searchEpisode || util.StringToIntMust(parsedEpisode[0]) == searchEpisode+absoluteOffset {
			return true
		}
	}

	return false
}
