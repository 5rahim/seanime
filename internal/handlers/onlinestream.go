package handlers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/onlinestream"
)

// HandleGetOnlineStreamEpisodeList
//
//	@summary returns the episode list for the given media and provider.
//	@desc It returns the episode list for the given media and provider.
//	@desc The episodes are cached using a file cache.
//	@desc The episode list is just a list of episodes with no video sources, it's what the client uses to display the episodes and subsequently fetch the sources.
//	@desc The episode list might be nil or empty if nothing could be found, but the media will always be returned.
//	@route /api/v1/onlinestream/episode-list [POST]
//	@returns {episodes: Episode[], media: BaseMedia}
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
	//if err != nil {
	//	return c.RespondWithError(err)
	//}

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
//	@summary returns the video sources for the given media, episode number and provider.
//	@route /api/v1/onlinestream/episode-sources [POST]
//	@returns onlinestream.EpisodeSource
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
//	@summary empties the cache for the given media.
//	@route /api/v1/onlinestream/cache [DELETE]
//	@returns bool
func HandleOnlineStreamEmptyCache(c *RouteCtx) error {

	type body struct {
		MediaId int `json:"mediaId"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	err := c.App.Onlinestream.EmptyCache(b.MediaId)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}
