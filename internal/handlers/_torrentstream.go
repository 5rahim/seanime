package handlers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/database/models"
	"github.com/seanime-app/seanime/internal/torrents/torrent"
)

// HandleGetTorrentstreamEpisodeCollection
//
//	@summary get list of episodes
//	@desc This returns a list of episodes.
//	@returns torrentstream.EpisodeCollection
//	@param id - int - true - "AniList anime media ID"
//	@route /api/v1/torrentstream/episodes/{id} [GET]
func HandleGetTorrentstreamEpisodeCollection(c *RouteCtx) error {
	mId, err := c.Fiber.ParamsInt("id")
	if err != nil {
		return c.RespondWithError(err)
	}

	ec, err := c.App.TorrentstreamRepository.NewEpisodeCollection(mId)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(ec)
}

// HandleGetTorrentstreamSettings
//
//	@summary get torrentstream settings.
//	@desc This returns the torrentstream settings.
//	@returns models.TorrentstreamSettings
//	@route /api/v1/torrentstream/settings [GET]
func HandleGetTorrentstreamSettings(c *RouteCtx) error {
	torrentstreamSettings, found := c.App.Database.GetTorrentstreamSettings()
	if !found {
		return c.RespondWithError(errors.New("torrent streaming settings not found"))
	}

	return c.RespondWithData(torrentstreamSettings)
}

// HandleSaveTorrentstreamSettings
//
//	@summary save torrentstream settings.
//	@desc This saves the torrentstream settings.
//	@desc The client should refetch the server status.
//	@returns models.TorrentstreamSettings
//	@route /api/v1/torrentstream/settings [PATCH]
func HandleSaveTorrentstreamSettings(c *RouteCtx) error {

	type body struct {
		Settings models.TorrentstreamSettings `json:"settings"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	settings, err := c.App.Database.UpsertTorrentstreamSettings(&b.Settings)
	if err != nil {
		return c.RespondWithError(err)
	}

	c.App.InitOrRefreshTorrentstreamSettings()

	return c.RespondWithData(settings)
}

// HandleTorrentstreamStartStream
//
//	@summary starts a torrent stream.
//	@desc This starts the entire streaming process.
//	@returns bool
//	@route /api/v1/torrentstream/start [POST]
func HandleTorrentstreamStartStream(c *RouteCtx) error {

	type body struct {
		MediaId       int                   `json:"mediaId"`
		EpisodeNumber int                   `json:"episodeNumber"`
		AniDBEpisode  string                `json:"aniDBEpisode"`
		AutoSelect    bool                  `json:"autoSelect"`
		Torrent       *torrent.AnimeTorrent `json:"torrent"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	err := c.App.TorrentstreamRepository.StartStream(&torrentstream.StartStreamOptions{
		MediaId:       b.MediaId,
		EpisodeNumber: b.EpisodeNumber,
		AniDBEpisode:  b.AniDBEpisode,
		AutoSelect:    b.AutoSelect,
		Torrent:       b.Torrent,
	})

	return c.RespondWithError(err)
}
