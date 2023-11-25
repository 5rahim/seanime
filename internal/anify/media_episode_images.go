package anify

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"net/http"
	"strconv"
)

const episodesUrl = "https://api.anify.tv/episodes"

type (
	EpisodeRequestResults []*EpisodeRequestResult

	EpisodeRequestResult struct {
		ProviderId string `json:"providerId"`
		Episodes   []struct {
			ID          string `json:"id"`
			IsFiller    bool   `json:"isFiller,omitempty"`
			Number      int    `json:"number"`
			Image       string `json:"img"`
			HasDub      bool   `json:"hasDub"`
			Description string `json:"description,omitempty"`
			Rating      int    `json:"rating"`
			UpdatedAt   int    `json:"updatedAt,omitempty"`
		} `json:"episodes"`
	}

	MediaEpisodeImagesEntry struct {
		MediaId          int                  `json:"mediaId"`
		EpisodeImageData []*MediaEpisodeImage `json:"episodeImageData"`
	}

	MediaEpisodeImage struct {
		EpisodeNumber int    `json:"episodeNumber"`
		Image         string `json:"image"`
	}
)

func FetchMediaEpisodeImagesEntry(mId int) (*MediaEpisodeImagesEntry, error) {

	res, err := http.Get(episodesUrl + "?id=" + strconv.Itoa(mId))
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("could not fetch media image data, (%s)", res.Status))
	}

	buff := bytes.NewBufferString("")
	_, err = buff.ReadFrom(res.Body)
	if err != nil {
		return nil, err
	}

	var a EpisodeRequestResults
	if err := json.Unmarshal(buff.Bytes(), &a); err != nil {
		return nil, err
	}

	if len(a) == 0 {
		return nil, errors.New("no results")
	}

	r := new(MediaEpisodeImagesEntry)

	r.MediaId = mId
	r.EpisodeImageData = make([]*MediaEpisodeImage, 0)

	for _, provider := range a {
		for _, ep := range provider.Episodes {
			// Add the episode only if it's not already included
			if !r.HasEpisode(ep.Number) && ep.Image != "" {
				r.EpisodeImageData = append(r.EpisodeImageData, &MediaEpisodeImage{
					EpisodeNumber: ep.Number,
					Image:         ep.Image,
				})
			}
		}
	}

	return r, nil

}
