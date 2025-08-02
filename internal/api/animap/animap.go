package animap

import (
	"errors"
	"io"
	"net/http"
	"seanime/internal/constants"
	"seanime/internal/hook"
	"seanime/internal/util/result"
	"strconv"

	"github.com/goccy/go-json"
)

type (
	Anime struct {
		Title     string              `json:"title"`
		Titles    map[string]string   `json:"titles,omitempty"`
		StartDate string              `json:"startDate,omitempty"` // YYYY-MM-DD
		EndDate   string              `json:"endDate,omitempty"`   // YYYY-MM-DD
		Status    string              `json:"status"`              // Finished, Airing, Upcoming, etc.
		Type      string              `json:"type"`                // TV, OVA, Movie, etc.
		Episodes  map[string]*Episode `json:"episodes,omitzero"`   // Indexed by AniDB episode number, "1", "S1", etc.
		Mappings  *AnimeMapping       `json:"mappings,omitzero"`
	}

	AnimeMapping struct {
		AnidbID          int    `json:"anidb_id,omitempty"`
		AnilistID        int    `json:"anilist_id,omitempty"`
		KitsuID          int    `json:"kitsu_id,omitempty"`
		TheTvdbID        int    `json:"thetvdb_id,omitempty"`
		TheMovieDbID     string `json:"themoviedb_id,omitempty"` // Can be int or string, forced to string
		MalID            int    `json:"mal_id,omitempty"`
		LivechartID      int    `json:"livechart_id,omitempty"`
		AnimePlanetID    string `json:"anime-planet_id,omitempty"` // Can be int or string, forced to string
		AnisearchID      int    `json:"anisearch_id,omitempty"`
		SimklID          int    `json:"simkl_id,omitempty"`
		NotifyMoeID      string `json:"notify.moe_id,omitempty"`
		AnimecountdownID int    `json:"animecountdown_id,omitempty"`
		Type             string `json:"type,omitempty"`
	}

	Episode struct {
		AnidbEpisode   string `json:"anidbEpisode"`
		AnidbId        int    `json:"anidbEid"`
		TvdbId         int    `json:"tvdbEid,omitempty"`
		TvdbShowId     int    `json:"tvdbShowId,omitempty"`
		AirDate        string `json:"airDate,omitempty"`    // YYYY-MM-DD
		AnidbTitle     string `json:"anidbTitle,omitempty"` // Title of the episode from AniDB
		TvdbTitle      string `json:"tvdbTitle,omitempty"`  // Title of the episode from TVDB
		Overview       string `json:"overview,omitempty"`
		Image          string `json:"image,omitempty"`
		Runtime        int    `json:"runtime,omitempty"` // minutes
		Length         string `json:"length,omitempty"`  // Xm
		SeasonNumber   int    `json:"seasonNumber,omitempty"`
		SeasonName     string `json:"seasonName,omitempty"`
		Number         int    `json:"number"`
		AbsoluteNumber int    `json:"absoluteNumber,omitempty"`
	}
)

//----------------------------------------------------------------------------------------------------------------------

type Cache struct {
	*result.Cache[string, *Anime]
}

func NewCache() *Cache {
	return &Cache{result.NewCache[string, *Anime]()}
}

func GetCacheKey(from string, id int) string {
	return from + strconv.Itoa(id)
}

//----------------------------------------------------------------------------------------------------------------------

// FetchAnimapMedia fetches animap.Anime from the Animap API.
func FetchAnimapMedia(from string, id int) (*Anime, error) {

	// Event
	reqEvent := &AnimapMediaRequestedEvent{
		From:  from,
		Id:    id,
		Media: &Anime{},
	}
	err := hook.GlobalHookManager.OnAnimapMediaRequested().Trigger(reqEvent)
	if err != nil {
		return nil, err
	}

	// If the hook prevented the default behavior, return the data
	if reqEvent.DefaultPrevented {
		return reqEvent.Media, nil
	}

	from = reqEvent.From
	id = reqEvent.Id

	apiUrl := constants.InternalMetadataURL + "/entry?" + from + "_id=" + strconv.Itoa(id)

	request, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("X-Seanime-Version", "Seanime/"+constants.Version)

	// Send an HTTP GET request
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, errors.New("not found on Animap")
	}

	// Read the response body
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON data into AnimapData
	var media Anime
	if err := json.Unmarshal(responseBody, &media); err != nil {
		return nil, err
	}

	// Event
	event := &AnimapMediaEvent{
		Media: &media,
	}
	err = hook.GlobalHookManager.OnAnimapMedia().Trigger(event)
	if err != nil {
		return nil, err
	}

	// If the hook prevented the default behavior, return the data
	if event.DefaultPrevented {
		return event.Media, nil
	}

	return event.Media, nil
}

// FetchAnimapMediaC is the same as FetchAnimapMedia but uses a cache.
// If the media is found in the cache, it will be returned.
// If the media is not found in the cache, it will be fetched and then added to the cache.
func FetchAnimapMediaC(from string, id int, cache *Cache) (*Anime, error) {

	cacheV, ok := cache.Get(GetCacheKey(from, id))
	if ok {
		return cacheV, nil
	}

	media, err := FetchAnimapMedia(from, id)
	if err != nil {
		return nil, err
	}

	cache.Set(GetCacheKey(from, id), media)

	return media, nil
}
