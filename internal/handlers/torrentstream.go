package handlers

import (
	"errors"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime/internal/database/models"
	"github.com/seanime-app/seanime/internal/library/anime"
	"github.com/seanime-app/seanime/internal/torrents/torrent"
	"github.com/seanime-app/seanime/internal/torrentstream"
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

	lop.ForEach(ec.Episodes, func(e *anime.AnimeEntryEpisode, _ int) {
		c.App.FillerManager.HydrateEpisodeFillerData(mId, e)
	})

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
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

// HandleTorrentstreamStopStream
//
//	@summary stop a torrent stream.
//	@desc This stops the entire streaming process and drops the torrent if it's below a threshold.
//	@desc This is made to be used while the stream is running.
//	@returns bool
//	@route /api/v1/torrentstream/stop [POST]
func HandleTorrentstreamStopStream(c *RouteCtx) error {

	err := c.App.TorrentstreamRepository.StopStream()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

// HandleTorrentstreamDropTorrent
//
//	@summary drops a torrent stream.
//	@desc This stops the entire streaming process and drops the torrent completely.
//	@desc This is made to be used to force drop a torrent.
//	@returns bool
//	@route /api/v1/torrentstream/drop [POST]
func HandleTorrentstreamDropTorrent(c *RouteCtx) error {

	err := c.App.TorrentstreamRepository.DropTorrent()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

// HandleTorrentstreamDONOTUSE
//
//	@summary used to generate typescript types
//	@returns torrentstream.TorrentLoadingStatus
//	@route /api/v1/torrentstream/DONOTUSE
func HandleTorrentstreamDONOTUSE(c *RouteCtx) error {
	return c.RespondWithData(true)
}

// HandleTorrentstreamDONOTUSE2
//
//	@summary used to generate typescript types
//	@returns torrentstream.TorrentStatus
//	@route /api/v1/torrentstream/DONOTUSE
func HandleTorrentstreamDONOTUSE2(c *RouteCtx) error {
	return c.RespondWithData(true)
}
