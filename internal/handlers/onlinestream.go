package handlers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/onlinestream"
)

// HandleGetOnlineStreamEpisodeList returns the episodes.
// It returns the best available episodes from the online stream providers.
//
//	POST /v1/onlinestream/episode-list
func HandleGetOnlineStreamEpisodeList(c *RouteCtx) error {

	type body struct {
		MediaId  int    `json:"mediaId"`
		Dubbed   bool   `json:"dubbed"`
		Provider string `json:"provider"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	if !c.App.Settings.Library.EnableOnlinestream {
		return c.RespondWithError(errors.New("enable online streaming in the settings"))
	}

	// Get media
	// This is cached
	media, err := c.App.Onlinestream.GetMedia(b.MediaId)
	if err != nil {
		return c.RespondWithError(err)
	}

	if media.Status == nil || *media.Status == anilist.MediaStatusNotYetReleased {
		return c.RespondWithError(errors.New("unavailable"))
	}

	// Get episode list
	// This is cached using file cache
	episodes, err := c.App.Onlinestream.GetMediaEpisodes(b.Provider, media, b.Dubbed)
	if err != nil {
		return c.RespondWithError(err)
	}

	ret := struct {
		Episodes []*onlinestream.Episode `json:"episodes"`
		Media    *anilist.BaseMedia      `json:"media"`
	}{
		Episodes: episodes,
		Media:    media,
	}

	return c.RespondWithData(ret)
}

// HandleGetOnlineStreamEpisodeSource
//
//	POST /v1/onlinestream/episode-sources
func HandleGetOnlineStreamEpisodeSource(c *RouteCtx) error {

	type body struct {
		EpisodeNumber int    `json:"episodeNumber"`
		MediaId       int    `json:"mediaId"`
		Provider      string `json:"provider"`
		Dubbed        bool   `json:"dubbed"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	sources, err := c.App.Onlinestream.GetEpisodeSources(b.Provider, b.MediaId, b.EpisodeNumber, b.Dubbed)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(sources)
}

// HandleOnlineStreamEmptyCache
//
//	DELETE /v1/onlinestream/cache
func HandleOnlineStreamEmptyCache(c *RouteCtx) error {

	err := c.App.Onlinestream.EmptyCache()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}
