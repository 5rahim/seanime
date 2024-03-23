package tvdb

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type (
	TVDB struct {
		apiKey       string
		client       *http.Client
		currentToken string // Hydrated by getTokenWithKey
		logger       *zerolog.Logger
	}

	NewTVDBOptions struct {
		ApiKey string
		Logger *zerolog.Logger
	}

	FilterEpisodeMediaInfo struct {
		Year           *int
		Month          *int
		TotalEp        int // form anizip
		AbsoluteOffset int // from anizip
	}
)

func NewTVDB(opts *NewTVDBOptions) *TVDB {
	return &TVDB{
		apiKey: opts.ApiKey,
		client: &http.Client{},
		logger: opts.Logger,
	}
}

func (tvdb *TVDB) FetchSeriesEpisodes(id int, filter FilterEpisodeMediaInfo) (res []*Episode, err error) {
	// Get token
	_, err = tvdb.getTokenWithTries()
	if err != nil {
		return nil, err
	}

	// Fetch seasons
	seasons, err := tvdb.fetchSeasons(id)
	if err != nil {
		return nil, err
	}

	// Fetch episodes
	episodesF, err := tvdb.fetchEpisodes(seasons)
	if err != nil {
		return nil, err
	}

	// Filter episodes
	episodesF = tvdb.filterEpisodes(episodesF, filter)

	// Convert episodes
	res = make([]*Episode, len(episodesF), len(episodesF))
	for i, e := range episodesF {
		res[i] = &Episode{
			ID:      e.ID,
			Image:   e.Image,
			Number:  int(e.Number),
			AiredAt: e.Aired,
		}
	}

	tvdb.logger.Debug().Int("id", id).Int("episodes", len(res)).Msg("tvdb: Found episodes")

	return
}

// FetchMetadata fetches metadata for a series.
//   - id: The TVDB ID of the series.
func (tvdb *TVDB) fetchSeasons(id int) (res []*ExtendedSeriesResponse_Season, err error) {

	start := time.Now()
	tvdb.logger.Debug().Int("id", id).Msg("tvdb: Fetching seasons")

	// Fetch metadata
	resp, err := tvdb.doRequest(fmt.Sprintf("%s/series/%d/extended", ApiUrl, id), nil)
	if err != nil {
		return res, err
	}
	defer resp.Body.Close()

	// Parse response
	var data ExtendedSeriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		tvdb.logger.Error().Err(err).Msg("tvdb: Could not decode response")
		return res, err
	}

	if data.Data == nil || data.Data.Seasons == nil {
		tvdb.logger.Error().Msg("tvdb: Could not find seasons")
		return res, fmt.Errorf("could not find seasons")
	}

	tvdb.logger.Debug().Int("id", id).Int64("duration", time.Since(start).Milliseconds()).Msg("tvdb: Fetched seasons")

	res = data.Data.Seasons

	return res, nil
}

// fetchEpisodes returns a list of episodes based on a list of seasons.
func (tvdb *TVDB) fetchEpisodes(seasons []*ExtendedSeriesResponse_Season) (res []*ExtendedSeasonsResponse_Episode, err error) {

	tvdb.logger.Debug().Msg("tvdb: Fetching all possible episodes")

	_episodes := make([]*ExtendedSeasonsResponse_Episode, 0)

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	wg.Add(len(seasons))
	for _, _season := range seasons {

		go func(season *ExtendedSeriesResponse_Season) {
			defer wg.Done()

			tvdb.logger.Debug().Int64("seasonId", season.ID).Msg("tvdb: Fetching episodes for season")

			// Fetch season metadata
			resp, err := tvdb.doRequest(fmt.Sprintf("%s/seasons/%d/extended", ApiUrl, season.ID), nil)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			// Parse response
			var data ExtendedSeasonsResponse
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				tvdb.logger.Error().Int64("seasonId", season.ID).Err(err).Msg("tvdb: Could not decode response")
				return
			}

			if data.Data == nil || data.Data.Episodes == nil {
				tvdb.logger.Warn().Int64("seasonId", season.ID).Msg("tvdb: Could not find episodes for season")
				return
			}

			mu.Lock()
			_episodes = append(_episodes, data.Data.Episodes...)
			mu.Unlock()
		}(_season)

	}
	wg.Wait()

	res = _episodes

	return res, nil
}

func (tvdb *TVDB) filterEpisodes(episodes []*ExtendedSeasonsResponse_Episode, mediaInfo FilterEpisodeMediaInfo) (res []*ExtendedSeasonsResponse_Episode) {
	// The media doesn't have a year, just return a dedup list of episodes :shrug:
	episodeMap := make(map[int64]*ExtendedSeasonsResponse_Episode)

	if mediaInfo.Year == nil || *mediaInfo.Year == 0 {
		for _, episode := range episodes {
			if e, ok := episodeMap[episode.Number]; ok {
				if e.Image == "" && episode.Image != "" {
					episodeMap[episode.Number] = episode
				}
			} else {
				episodeMap[episode.Number] = episode
			}
		}

	} else {

		// The media has a year, we need to filter episodes based on the year
		for _, episode := range episodes {
			if episode.Year == "" {
				continue
			}
			episodeYear, _ := strconv.Atoi(episode.Year)
			if episodeYear >= *mediaInfo.Year {
				if e, ok := episodeMap[episode.Number]; ok {
					if e.Image == "" && episode.Image != "" {
						episodeMap[episode.Number] = episode
					}
				} else {
					episodeMap[episode.Number] = episode
				}
			}
		}

	}

	// Convert map to slice
	res = make([]*ExtendedSeasonsResponse_Episode, 0)
	for _, episode := range episodeMap {
		res = append(res, episode)
	}

	// If we find episodes that are over the total episode count, we need to apply the offset
	if mediaInfo.TotalEp > 0 {

		// If the media has a month, we don't need to apply the offset
		// just filter the episodes based on the month
		if mediaInfo.Month != nil && *mediaInfo.Month > 0 {

			// Filter episodes
			filteredEpisodes := make([]*ExtendedSeasonsResponse_Episode, 0)
			for _, episode := range res {
				if episode.Aired == "" || episode.Year == "" {
					continue
				}
				airedParts := strings.Split(episode.Aired, "-")
				episodeMonth, _ := strconv.Atoi(airedParts[1])
				episodeYear, _ := strconv.Atoi(episode.Year)
				// If the episode aired AFTER the month/year, we can include it
				if episodeYear > *mediaInfo.Year || (episodeYear == *mediaInfo.Year && episodeMonth >= *mediaInfo.Month) {
					filteredEpisodes = append(filteredEpisodes, episode)
				}
			}

			if len(filteredEpisodes) > 0 {
				res = filteredEpisodes
			}

		}

		// Get the offset
		offset := mediaInfo.AbsoluteOffset

		// Filter episodes
		filteredEpisodes := make([]*ExtendedSeasonsResponse_Episode, 0)
		for _, episode := range res {
			if episode.Number > int64(offset) && episode.Number <= int64(mediaInfo.TotalEp+offset) {
				episode.Number = episode.Number - int64(offset)
				filteredEpisodes = append(filteredEpisodes, episode)
			}
		}

		if len(filteredEpisodes) > 0 {
			res = filteredEpisodes
		}

	}

	return res
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (tvdb *TVDB) doRequest(url string, body io.Reader) (res *http.Response, err error) {
	req, err := http.NewRequest("GET", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tvdb.currentToken))

	return tvdb.client.Do(req)
}

// getTokenWithTries tries to get a token with all available API keys.
// If an API key is provided in the options, it will be tried first.
func (tvdb *TVDB) getTokenWithTries() (token string, err error) {

	if tvdb.apiKey != "" {
		token, err := tvdb.getTokenWithKey(tvdb.apiKey)
		if err == nil {
			return token, nil
		}
	}

	for _, key := range ApiKeys {
		token, err := tvdb.getTokenWithKey(key)
		if err != nil {
			continue
		}

		return token, nil
	}

	return "", fmt.Errorf("could not get authentication token")
}

// getTokenWithKey gets a token with a specific API key.
func (tvdb *TVDB) getTokenWithKey(key string) (token string, err error) {
	req, err := tvdb.client.Post(fmt.Sprintf("%s/login", ApiUrl), "application/json", strings.NewReader(fmt.Sprintf(`{"apikey":"%s"}`, key)))
	if err != nil {
		return "", err
	}
	defer req.Body.Close()

	b, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}

	var res map[string]interface{}
	if err := json.Unmarshal(b, &res); err != nil {
		return "", err
	}

	data, ok := res["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("could not get token")
	}

	token, ok = data["token"].(string)
	if !ok {
		return "", fmt.Errorf("could not get token")
	}

	tvdb.currentToken = token

	return token, nil
}
