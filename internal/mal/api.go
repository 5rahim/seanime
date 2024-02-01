package mal

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/goccy/go-json"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	ApiBaseURL                     string          = "https://api.myanimelist.net/v2"
	MediaTypeTV                    MediaType       = "tv"
	MediaTypeOVA                   MediaType       = "ova"
	MediaTypeMovie                 MediaType       = "movie"
	MediaTypeSpecial               MediaType       = "special"
	MediaTypeONA                   MediaType       = "ona"
	MediaTypeMusic                 MediaType       = "music"
	MediaTypeManga                 MediaType       = "manga"
	MediaTypeNovel                 MediaType       = "novel"
	MediaTypeOneShot               MediaType       = "oneshot"
	MediaStatusFinishedAiring      MediaStatus     = "finished_airing"
	MediaStatusCurrentlyAiring     MediaStatus     = "currently_airing"
	MediaStatusNotYetAired         MediaStatus     = "not_yet_aired"
	MediaStatusFinished            MediaStatus     = "finished"
	MediaStatusCurrentlyPublishing MediaStatus     = "currently_publishing"
	MediaStatusNotYetPublished     MediaStatus     = "not_yet_published"
	MediaListStatusWatching        MediaListStatus = "watching"
	MediaListStatusCompleted       MediaListStatus = "completed"
	MediaListStatusOnHold          MediaListStatus = "on_hold"
	MediaListStatusDropped         MediaListStatus = "dropped"
	MediaListStatusPlanToWatch     MediaListStatus = "plan_to_watch"

	BaseAnimeFields string = "id,title,main_picture,alternative_titles,start_date,end_date,start_season,synopsis,num_episodes,mean,rank,popularity,media_type,status"
)

type (
	MediaType       string
	MediaStatus     string
	MediaListStatus string

	RequestOptions struct {
		AccessToken  string
		RefreshToken string
		ExpiresAt    time.Time
	}

	BasicAnime struct {
		ID          int    `json:"id"`
		Title       string `json:"title"`
		MainPicture struct {
			Medium string `json:"medium"`
			Large  string `json:"large"`
		} `json:"main_picture"`
		AlternativeTitles struct {
			Synonyms []string `json:"synonyms"`
			En       string   `json:"en"`
			Ja       string   `json:"ja"`
		} `json:"alternative_titles"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		StartSeason struct {
			Year   int    `json:"year"`
			Season string `json:"season"`
		} `json:"start_season"`
		Synopsis    string      `json:"synopsis"`
		NumEpisodes int         `json:"num_episodes"`
		Mean        float32     `json:"mean"`
		Rank        int         `json:"rank"`
		Popularity  int         `json:"popularity"`
		MediaType   MediaType   `json:"media_type"`
		Status      MediaStatus `json:"status"`
	}
)

func GetAnimeDetails(accessToken string, mId int) (*BasicAnime, error) {

	reqUrl := fmt.Sprintf("%s/anime/%d?fields=%s", ApiBaseURL, mId, BaseAnimeFields)

	// Create a new HTTP GET request
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return nil, err
	}

	if accessToken == "" {
		return nil, fmt.Errorf("access token is empty")
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Make the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response status %s", resp.Status)
	}

	// Decode the response
	var anime BasicAnime
	if err := json.NewDecoder(resp.Body).Decode(&anime); err != nil {
		return nil, err
	}

	return &anime, nil
}

type AnimeListStatusParams struct {
	Status             *MediaListStatus
	IsRewatching       *bool
	NumWatchedEpisodes *int
	Score              *int
}

func UpdateAnimeListStatus(accessToken string, opts *AnimeListStatusParams, mId int) error {

	reqUrl := fmt.Sprintf("%s/anime/%d/my_list_status", ApiBaseURL, mId)

	client := &http.Client{}

	// Build URL
	urlData := url.Values{}
	if opts.Status != nil {
		urlData.Set("status", string(*opts.Status))
	}
	if opts.IsRewatching != nil {
		urlData.Set("is_rewatching", fmt.Sprintf("%t", *opts.IsRewatching))
	}
	if opts.NumWatchedEpisodes != nil {
		urlData.Set("num_watched_episodes", fmt.Sprintf("%d", *opts.NumWatchedEpisodes))
	}
	if opts.Score != nil {
		urlData.Set("score", fmt.Sprintf("%d", *opts.Score))
	}
	encodedData := urlData.Encode()

	spew.Dump(reqUrl)

	req, err := http.NewRequest("PATCH", reqUrl, strings.NewReader(encodedData))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+accessToken)

	// Response
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid response status %s", res.Status)
	}

	return nil
}
