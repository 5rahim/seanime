package handlers

import (
	"errors"
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
	lop "github.com/samber/lo/parallel"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/database/models"
	"seanime/internal/library/anime"
	"seanime/internal/torrentstream"
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

	lop.ForEach(ec.Episodes, func(e *anime.Episode, _ int) {
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

// HandleGetTorrentstreamTorrentFilePreviews
//
//	@summary get list of torrent files from a batch
//	@desc This returns a list of file previews from the torrent
//	@returns []torrentstream.FilePreview
//	@route /api/v1/torrentstream/torrent-file-previews [POST]
func HandleGetTorrentstreamTorrentFilePreviews(c *RouteCtx) error {
	type body struct {
		Torrent       *hibiketorrent.AnimeTorrent `json:"torrent"`
		EpisodeNumber int                         `json:"episodeNumber"`
		Media         *anilist.BaseAnime          `json:"media"`
	}
	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	providerExtension, ok := c.App.ExtensionRepository.GetAnimeTorrentProviderExtensionByID(b.Torrent.Provider)
	if !ok {
		return c.RespondWithError(errors.New("torrentstream: Torrent provider extension not found"))
	}

	magnet, err := providerExtension.GetProvider().GetTorrentMagnetLink(b.Torrent)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Get the media metadata
	animeMetadata, _ := c.App.MetadataProvider.GetAnimeMetadata(metadata.AnilistPlatform, b.Media.ID)
	absoluteOffset := 0
	if animeMetadata != nil {
		absoluteOffset = animeMetadata.GetOffset()
	}

	files, err := c.App.TorrentstreamRepository.GetTorrentFilePreviewsFromManualSelection(&torrentstream.GetTorrentFilePreviewsOptions{
		Torrent:        b.Torrent,
		Magnet:         magnet,
		EpisodeNumber:  b.EpisodeNumber,
		AbsoluteOffset: absoluteOffset,
		Media:          b.Media,
	})
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(files)
}

// HandleTorrentstreamStartStream
//
//	@summary starts a torrent stream.
//	@desc This starts the entire streaming process.
//	@returns bool
//	@route /api/v1/torrentstream/start [POST]
func HandleTorrentstreamStartStream(c *RouteCtx) error {

	type body struct {
		MediaId       int                         `json:"mediaId"`
		EpisodeNumber int                         `json:"episodeNumber"`
		AniDBEpisode  string                      `json:"aniDBEpisode"`
		AutoSelect    bool                        `json:"autoSelect"`
		Torrent       *hibiketorrent.AnimeTorrent `json:"torrent,omitempty"` // Nil if autoSelect is true
		FileIndex     *int                        `json:"fileIndex,omitempty"`
		PlaybackType  torrentstream.PlaybackType  `json:"playbackType"` // "default" or "externalPlayerLink"
		ClientId      string                      `json:"clientId"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	userAgent := c.Fiber.Get("User-Agent")

	err := c.App.TorrentstreamRepository.StartStream(&torrentstream.StartStreamOptions{
		MediaId:       b.MediaId,
		EpisodeNumber: b.EpisodeNumber,
		AniDBEpisode:  b.AniDBEpisode,
		AutoSelect:    b.AutoSelect,
		Torrent:       b.Torrent,
		FileIndex:     b.FileIndex,
		UserAgent:     userAgent,
		ClientId:      b.ClientId,
		PlaybackType:  b.PlaybackType,
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

// HandleGetTorrentstreamBatchHistory
//
//	@summary returns the most recent batch selected.
//	@desc This returns the most recent batch selected.
//	@returns torrentstream.BatchHistoryResponse
//	@route /api/v1/torrentstream/batch-history [POST]
func HandleGetTorrentstreamBatchHistory(c *RouteCtx) error {
	type body struct {
		MediaID int `json:"mediaId"`
	}
	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	ret := c.App.TorrentstreamRepository.GetBatchHistory(b.MediaID)
	return c.RespondWithData(ret)
}
