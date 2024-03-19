package onlinestream

import (
	"errors"
	"fmt"
	"github.com/seanime-app/seanime/internal/onlinestream/providers"
	"github.com/seanime-app/seanime/internal/util/comparison"
	"strconv"
)

const (
	ProviderGogoanime Provider = "gogoanime"
	ProviderZoro      Provider = "zoro"
)

var (
	Providers = []Provider{
		ProviderGogoanime,
		ProviderZoro,
	}

	ErrNoAnimeFound         = errors.New("no anime found")
	ErrNoEpisodes           = errors.New("no episodes found")
	errNoEpisodeSourceFound = errors.New("no source found for episode")
)

type (
	Provider string

	episodeContainer struct {
		ProviderEpisodes []*extractedProviderEpisodes
	}

	extractedProviderEpisodes struct {
		Provider          Provider
		ExtractedEpisodes []*extractedEpisode
	}

	// extractedEpisode contains the episode data from a provider.
	extractedEpisode struct {
		ID            string
		Number        int
		Title         string
		ServerSources []*onlinestream_providers.ProviderServerSources
	}
)

func (os *OnlineStream) getEpisodeContainer(provider Provider, mId int, titles []*string, from int, to int, dubbed bool) (*episodeContainer, bool) {
	providerEpisodesInfoKey := strconv.Itoa(mId) + "$" + string(provider)

	ae := &episodeContainer{
		ProviderEpisodes: make([]*extractedProviderEpisodes, 0, len(Providers)),
	}

	episodes := make([]*extractedEpisode, 0)
	var providerEpisodesInfo []*onlinestream_providers.ProviderEpisodeInfo

	if found, _ := os.fileCacher.Get(os.fcProviderEpisodesInfoBucket, providerEpisodesInfoKey, &providerEpisodesInfo); !found {
		var err error
		providerEpisodesInfo, err = os.getProviderEpisodes(provider, titles, dubbed)
		if err != nil {
			os.logger.Error().Err(err).Msg("onlinestream: failed to get provider episodes")
			return nil, false
		}
		_ = os.fileCacher.Set(os.fcProviderEpisodesInfoBucket, providerEpisodesInfoKey, providerEpisodesInfo)
	}

	for _, providerEpisodeInfo := range providerEpisodesInfo {

		if providerEpisodeInfo.Number >= from && providerEpisodeInfo.Number <= to {
			// Check if the episode is cached to avoid fetching the sources again.
			key := fmt.Sprintf("%d$%s$%d$%v", mId, provider, providerEpisodeInfo.Number, dubbed)

			var cached *extractedEpisode
			if found, _ := os.fileCacher.Get(os.fcEpisodeBucket, key, &cached); found {
				episodes = append(episodes, cached)
				continue
			}

			// Fetch episode sources
			episodeSources, err := os.getEpisodeSources(provider, providerEpisodeInfo)
			if err != nil {
				continue
			}

			episode := &extractedEpisode{
				ID:            providerEpisodeInfo.ID,
				Number:        providerEpisodeInfo.Number,
				Title:         providerEpisodeInfo.Title,
				ServerSources: episodeSources,
			}
			episodes = append(episodes, episode)

			//os.episodeCache.SetT(key, episode, 10*time.Minute)
			_ = os.fileCacher.Set(os.fcEpisodeBucket, key, episode)

		}

	}

	if len(episodes) > 0 {
		ae.ProviderEpisodes = append(ae.ProviderEpisodes, &extractedProviderEpisodes{
			Provider:          provider,
			ExtractedEpisodes: episodes,
		})
	}

	if len(ae.ProviderEpisodes) == 0 {
		return nil, false
	}

	return ae, true
}

// getEpisodeSources gets the onlinestream_providers.ProviderEpisodeInfo server sources from the provider.
// It returns errNoEpisodeSourceFound if no sources are found.
func (os *OnlineStream) getEpisodeSources(provider Provider, providerEpisodeInfo *onlinestream_providers.ProviderEpisodeInfo) ([]*onlinestream_providers.ProviderServerSources, error) {
	var providerServers []*onlinestream_providers.ProviderServerSources
	switch provider {
	case ProviderGogoanime:
		res, err := os.gogo.FindEpisodeServerSources(providerEpisodeInfo, onlinestream_providers.VidstreamingServer)
		if err == nil {
			providerServers = append(providerServers, res)
		}
		res, err = os.gogo.FindEpisodeServerSources(providerEpisodeInfo, onlinestream_providers.GogocdnServer)
		if err == nil {
			providerServers = append(providerServers, res)
		}
		//res, err = os.gogo.FindEpisodeServerSources(providerEpisodeInfo, onlinestream_providers.StreamSBServer)
		//if err == nil {
		//	providerServers = append(providerServers, res)
		//}
	case ProviderZoro:
		res, err := os.zoro.FindEpisodeServerSources(providerEpisodeInfo, onlinestream_providers.VidcloudServer)
		if err == nil {
			providerServers = append(providerServers, res)
		}
		res, err = os.zoro.FindEpisodeServerSources(providerEpisodeInfo, onlinestream_providers.VidstreamingServer)
		if err == nil {
			providerServers = append(providerServers, res)
		}
		//res, err = os.zoro.FindEpisodeServerSources(providerEpisodeInfo, onlinestream_providers.StreamtapeServer)
		//if err == nil {
		//	providerServers = append(providerServers, res)
		//}
		//res, err = os.zoro.FindEpisodeServerSources(providerEpisodeInfo, onlinestream_providers.StreamSBServer)
		//if err == nil {
		//	providerServers = append(providerServers, res)
		//}
	}

	if len(providerServers) == 0 {
		return nil, errNoEpisodeSourceFound
	}

	return providerServers, nil
}

// getProviderEpisodes gets the anime episodes from provider based of the titles.
// It returns ErrNoAnimeFound if the anime is not found or ErrNoEpisodes if no episodes are found.
func (os *OnlineStream) getProviderEpisodes(provider Provider, titles []*string, dubbed bool) ([]*onlinestream_providers.ProviderEpisodeInfo, error) {
	var ret []*onlinestream_providers.ProviderEpisodeInfo
	romajiTitle := titles[0]

	// Get search results.
	var searchResults []*onlinestream_providers.SearchResult
	switch provider {
	case ProviderGogoanime:
		res, err := os.gogo.Search(*romajiTitle, dubbed)
		if err == nil {
			searchResults = res
		}
	case ProviderZoro:
		res, err := os.zoro.Search(*romajiTitle, dubbed)
		if err == nil {
			searchResults = res
		}
	}
	if len(searchResults) == 0 {
		return nil, ErrNoAnimeFound
	}

	// Filter results to get the best match.

	compBestResults := make([]*comparison.SorensenDiceResult, 0, len(searchResults))
	for _, r := range searchResults {
		// Compare search result title with all titles.
		compBestResult, found := comparison.FindBestMatchWithSorensenDice(&r.Title, titles)
		if found {
			compBestResults = append(compBestResults, compBestResult)
		}
	}
	compBestResult := compBestResults[0]
	for _, r := range compBestResults {
		if r.Rating > compBestResult.Rating {
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
	case ProviderGogoanime:
		res, err := os.gogo.FindEpisodesInfo(bestResult.ID)
		if err != nil {
			return nil, err
		}
		ret = res
	case ProviderZoro:
		res, err := os.zoro.FindEpisodesInfo(bestResult.ID)
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
