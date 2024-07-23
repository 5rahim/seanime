package onlinestream

import (
	"errors"
	"fmt"
	"seanime/internal/onlinestream/providers"
	"seanime/internal/util/comparison"
	"strings"
)

var (
	ErrNoAnimeFound         = errors.New("no anime found")
	ErrNoEpisodes           = errors.New("no episodes found")
	errNoEpisodeSourceFound = errors.New("no source found for episode")
)

type (
	// episodeContainer contains results of fetching the episodes from the provider.
	episodeContainer struct {
		Provider onlinestream_providers.Provider
		// List of episode details from the provider.
		// It is used to get the episode servers.
		ProviderEpisodeList []*onlinestream_providers.EpisodeDetails
		// List of episodes with their servers.
		Episodes []*episodeData
	}

	// episodeData contains some details about a provider episode and all available servers.
	episodeData struct {
		Provider onlinestream_providers.Provider
		ID       string
		Number   int
		Title    string
		Servers  []*onlinestream_providers.EpisodeServer
	}
)

// getEpisodeContainer gets the episode details and servers from the specified provider.
// It takes the media ID, titles in order to fetch the episode details.
//   - This function can be used to only get the episode details by setting 'from' and 'to' to 0.
//
// Since the episode details are cached, we can request episode servers multiple times without fetching the episode details again.
func (r *Repository) getEpisodeContainer(provider onlinestream_providers.Provider, mId int, titles []*string, from int, to int, dubbed bool) (*episodeContainer, error) {

	r.logger.Debug().
		Str("provider", string(provider)).
		Int("mediaId", mId).
		Int("from", from).
		Int("to", to).
		Bool("dubbed", dubbed).
		Msg("onlinestream: Getting episode container")

	// Key identifying the provider episode list in the file cache.
	// It includes "dubbed" because Gogoanime has a different entry for dubbed anime.
	providerEpisodeListKey := fmt.Sprintf("%d$%s$%v", mId, string(provider), dubbed)

	ec := &episodeContainer{
		Provider:            provider,
		Episodes:            make([]*episodeData, 0),
		ProviderEpisodeList: make([]*onlinestream_providers.EpisodeDetails, 0),
	}

	// Get the episode details from the provider.
	r.logger.Debug().
		Str("key", providerEpisodeListKey).
		Msgf("onlinestream: Fetching %s episode list", provider)

	// Bucket
	fcEpisodeListBucket := r.getFcEpisodeListBucket(provider, mId)
	fcEpisodeDataBucket := r.getFcEpisodeDataBucket(provider, mId)

	var providerEpisodeList []*onlinestream_providers.EpisodeDetails
	if found, _ := r.fileCacher.Get(fcEpisodeListBucket, providerEpisodeListKey, &providerEpisodeList); !found {
		var err error
		providerEpisodeList, err = r.getProviderEpisodeListFromTitles(provider, titles, dubbed)
		if err != nil {
			r.logger.Error().Err(err).Msg("onlinestream: failed to get provider episodes")
			return nil, err // ErrNoAnimeFound or ErrNoEpisodes
		}
		_ = r.fileCacher.Set(fcEpisodeListBucket, providerEpisodeListKey, providerEpisodeList)
	} else {
		r.logger.Debug().
			Str("key", providerEpisodeListKey).
			Msg("onlinestream: Cache HIT for episode list")
	}

	ec.ProviderEpisodeList = providerEpisodeList

	for _, episodeDetails := range providerEpisodeList {

		if episodeDetails.Number >= from && episodeDetails.Number <= to {

			// Check if the episode is cached to avoid fetching the sources again.
			key := fmt.Sprintf("%d$%s$%d$%v", mId, provider, episodeDetails.Number, dubbed)

			r.logger.Debug().
				Str("key", key).
				Msgf("onlinestream: Fetching episode '%d' servers", episodeDetails.Number)

			// Check episode cache
			var cached *episodeData
			if found, _ := r.fileCacher.Get(fcEpisodeDataBucket, key, &cached); found {
				ec.Episodes = append(ec.Episodes, cached)

				r.logger.Debug().
					Str("key", key).
					Msgf("onlinestream: Cache HIT for episode '%d' servers", episodeDetails.Number)

				continue
			}

			// Zoro dubs
			if provider == onlinestream_providers.ZoroProvider && dubbed {
				// If the episode details have both sub and dub, we need to get the dub episode.
				if !strings.HasSuffix(episodeDetails.ID, string(onlinestream_providers.SubAndDub)) {
					// Skip sub-only episodes
					continue
				}
				// Replace "both" with "dub" so that [getProviderEpisodeServers] can find the dub episode.
				episodeDetails.ID = strings.Replace(episodeDetails.ID, string(onlinestream_providers.SubAndDub), string(onlinestream_providers.Dub), 1)
			}

			// Fetch episode servers
			servers, err := r.getProviderEpisodeServers(provider, episodeDetails)
			if err != nil {
				r.logger.Error().Err(err).Msgf("onlinestream: failed to get episode '%d' servers", episodeDetails.Number)
				continue
			}

			episode := &episodeData{
				ID:      episodeDetails.ID,
				Number:  episodeDetails.Number,
				Title:   episodeDetails.Title,
				Servers: servers,
			}
			ec.Episodes = append(ec.Episodes, episode)

			r.logger.Debug().
				Str("key", key).
				Msgf("onlinestream: Found %d servers for episode '%d'", len(servers), episodeDetails.Number)

			_ = r.fileCacher.Set(fcEpisodeDataBucket, key, episode)

		}

	}

	if from > 0 && to > 0 && len(ec.Episodes) == 0 {
		r.logger.Error().Msg("onlinestream: No episodes found")
		return nil, ErrNoEpisodes
	}

	if len(ec.ProviderEpisodeList) == 0 {
		r.logger.Error().Msg("onlinestream: No episodes found")
		return nil, ErrNoEpisodes
	}

	return ec, nil
}

// getProviderEpisodeServers gets all the available servers for the episode.
// It returns errNoEpisodeSourceFound if no sources are found.
//
// Example:
//
//	episodeDetails, _ := getProviderEpisodeListFromTitles(provider, titles, dubbed)
//	episodeServers, err := getProviderEpisodeServers(provider, episodeDetails[0])
func (r *Repository) getProviderEpisodeServers(provider onlinestream_providers.Provider, episodeDetails *onlinestream_providers.EpisodeDetails) ([]*onlinestream_providers.EpisodeServer, error) {
	var providerServers []*onlinestream_providers.EpisodeServer
	switch provider {
	case onlinestream_providers.GogoanimeProvider:
		res, err := r.gogo.FindEpisodeServer(episodeDetails, onlinestream_providers.VidstreamingServer)
		if err == nil {
			providerServers = append(providerServers, res)
		}
		res, err = r.gogo.FindEpisodeServer(episodeDetails, onlinestream_providers.GogocdnServer)
		if err == nil {
			providerServers = append(providerServers, res)
		}
		//res, err = os.gogo.FindEpisodeServer(episodeDetails, onlinestream_providers.StreamSBServer)
		//if err == nil {
		//	providerServers = append(providerServers, res)
		//}
	case onlinestream_providers.ZoroProvider:
		res, err := r.zoro.FindEpisodeServer(episodeDetails, onlinestream_providers.VidcloudServer)
		if err == nil {
			providerServers = append(providerServers, res)
		}
		res, err = r.zoro.FindEpisodeServer(episodeDetails, onlinestream_providers.VidstreamingServer)
		if err == nil {
			providerServers = append(providerServers, res)
		}
		//res, err = os.zoro.FindEpisodeServer(episodeDetails, onlinestream_providers.StreamtapeServer)
		//if err == nil {
		//	providerServers = append(providerServers, res)
		//}
		//res, err = os.zoro.FindEpisodeServer(episodeDetails, onlinestream_providers.StreamSBServer)
		//if err == nil {
		//	providerServers = append(providerServers, res)
		//}
	}

	if len(providerServers) == 0 {
		return nil, errNoEpisodeSourceFound
	}

	return providerServers, nil
}

// getProviderEpisodeListFromTitles gets all the onlinestream_providers.EpisodeDetails from the provider based on the anime's titles.
// It returns ErrNoAnimeFound if the anime is not found or ErrNoEpisodes if no episodes are found.
func (r *Repository) getProviderEpisodeListFromTitles(provider onlinestream_providers.Provider, titles []*string, dubbed bool) ([]*onlinestream_providers.EpisodeDetails, error) {
	var ret []*onlinestream_providers.EpisodeDetails
	romajiTitle := strings.ReplaceAll(*titles[0], ":", "")
	englishTitle := ""
	if len(titles) > 1 {
		englishTitle = strings.ReplaceAll(*titles[1], ":", "")
	}

	// Get search results.
	var searchResults []*onlinestream_providers.SearchResult
	switch provider {
	case onlinestream_providers.GogoanimeProvider:
		res, err := r.gogo.Search(romajiTitle, dubbed)
		if err == nil {
			searchResults = res
		} else {
			res, err = r.gogo.Search(englishTitle, dubbed)
			if err == nil {
				searchResults = res
			}
		}
	case onlinestream_providers.ZoroProvider:
		res, err := r.zoro.Search(romajiTitle, dubbed)
		if err == nil {
			searchResults = res
		} else {
			res, err = r.zoro.Search(englishTitle, dubbed)
			if err == nil {
				searchResults = res
			}
		}
	}
	if len(searchResults) == 0 {
		return nil, ErrNoAnimeFound
	}

	// Filter results to get the best match.

	compBestResults := make([]*comparison.LevenshteinResult, 0, len(searchResults))
	for _, r := range searchResults {
		// Compare search result title with all titles.
		compBestResult, found := comparison.FindBestMatchWithLevenstein(&r.Title, titles)
		if found {
			compBestResults = append(compBestResults, compBestResult)
		}
	}
	compBestResult := compBestResults[0]
	for _, r := range compBestResults {
		if r.Distance < compBestResult.Distance {
			compBestResult = r
		}
	}

	// Get most accurate search result.
	var bestResult *onlinestream_providers.SearchResult
	for _, r := range searchResults {
		if r.Title == *compBestResult.OriginalValue {
			bestResult = r
			break
		}
	}

	// Fetch episodes.

	switch provider {
	case onlinestream_providers.GogoanimeProvider:
		res, err := r.gogo.FindEpisodeDetails(bestResult.ID)
		if err != nil {
			return nil, err
		}
		ret = res
	case onlinestream_providers.ZoroProvider:
		res, err := r.zoro.FindEpisodeDetails(bestResult.ID)
		if err != nil {
			return nil, err
		}
		ret = res
	}

	if len(ret) == 0 {
		return nil, ErrNoEpisodes
	}

	return ret, nil
}
