package anizip

import (
	"errors"
	"io"
	"net/http"
	"seanime/internal/hook"
	"seanime/internal/util/result"
	"strconv"

	"github.com/goccy/go-json"
)

// AniZip is the API used for fetching anime metadata and mappings.

type (
	Episode struct {
		TvdbEid               int               `json:"tvdbEid,omitempty"`
		AirDate               string            `json:"airdate,omitempty"`
		SeasonNumber          int               `json:"seasonNumber,omitempty"`
		EpisodeNumber         int               `json:"episodeNumber,omitempty"`
		AbsoluteEpisodeNumber int               `json:"absoluteEpisodeNumber,omitempty"`
		Title                 map[string]string `json:"title,omitempty"`
		Image                 string            `json:"image,omitempty"`
		Summary               string            `json:"summary,omitempty"`
		Overview              string            `json:"overview,omitempty"`
		Runtime               int               `json:"runtime,omitempty"`
		Length                int               `json:"length,omitempty"`
		Episode               string            `json:"episode,omitempty"`
		AnidbEid              int               `json:"anidbEid,omitempty"`
		Rating                string            `json:"rating,omitempty"`
	}

	Mappings struct {
		AnimeplanetID string `json:"animeplanet_id,omitempty"`
		KitsuID       int    `json:"kitsu_id,omitempty"`
		MalID         int    `json:"mal_id,omitempty"`
		Type          string `json:"type,omitempty"`
		AnilistID     int    `json:"anilist_id,omitempty"`
		AnisearchID   int    `json:"anisearch_id,omitempty"`
		AnidbID       int    `json:"anidb_id,omitempty"`
		NotifymoeID   string `json:"notifymoe_id,omitempty"`
		LivechartID   int    `json:"livechart_id,omitempty"`
		ThetvdbID     int    `json:"thetvdb_id,omitempty"`
		ImdbID        string `json:"imdb_id,omitempty"`
		ThemoviedbID  string `json:"themoviedb_id,omitempty"`
	}

	Media struct {
		Titles       map[string]string  `json:"titles"`
		Episodes     map[string]Episode `json:"episodes"`
		EpisodeCount int                `json:"episodeCount"`
		SpecialCount int                `json:"specialCount"`
		Mappings     *Mappings          `json:"mappings"`
	}
)

//----------------------------------------------------------------------------------------------------------------------

type Cache struct {
	*result.Cache[string, *Media]
}

func NewCache() *Cache {
	return &Cache{result.NewCache[string, *Media]()}
}

func GetCacheKey(from string, id int) string {
	return from + strconv.Itoa(id)
}

//----------------------------------------------------------------------------------------------------------------------

// FetchAniZipMedia fetches anizip.Media from the AniZip API.
func FetchAniZipMedia(from string, id int) (*Media, error) {

	// Event
	reqEvent := &AnizipMediaRequestedEvent{
		From:  from,
		Id:    id,
		Media: &Media{},
	}
	err := hook.GlobalHookManager.OnAnizipMediaRequested().Trigger(reqEvent)
	if err != nil {
		return nil, err
	}

	// If the hook prevented the default behavior, return the data
	if reqEvent.DefaultPrevented {
		return reqEvent.Media, nil
	}

	from = reqEvent.From
	id = reqEvent.Id

	apiUrl := "https://api.ani.zip/v1/episodes?" + from + "_id=" + strconv.Itoa(id)

	// Send an HTTP GET request
	response, err := http.Get(apiUrl)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, errors.New("not found on AniZip")
	}

	// Read the response body
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON data into AniZipData
	var media Media
	if err := json.Unmarshal(responseBody, &media); err != nil {
		return nil, err
	}

	// Event
	event := &AnizipMediaEvent{
		Media: &media,
	}
	err = hook.GlobalHookManager.OnAnizipMedia().Trigger(event)
	if err != nil {
		return nil, err
	}

	// If the hook prevented the default behavior, return the data
	if event.DefaultPrevented {
		return event.Media, nil
	}

	return event.Media, nil
}

// FetchAniZipMediaC is the same as FetchAniZipMedia but uses a cache.
// If the media is found in the cache, it will be returned.
// If the media is not found in the cache, it will be fetched and then added to the cache.
func FetchAniZipMediaC(from string, id int, cache *Cache) (*Media, error) {

	cacheV, ok := cache.Get(GetCacheKey(from, id))
	if ok {
		return cacheV, nil
	}

	media, err := FetchAniZipMedia(from, id)
	if err != nil {
		return nil, err
	}

	cache.Set(GetCacheKey(from, id), media)

	return media, nil
}
