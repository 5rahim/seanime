package handlers

import (
	"errors"
	"net/http"
	"os"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/database/models"
	"seanime/internal/events"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/library/anime"
	"seanime/internal/torrentstream"
	"strconv"

	"github.com/labstack/echo/v4"
	lop "github.com/samber/lo/parallel"
)

// HandleGetTorrentstreamEpisodeCollection
//
//	@summary get list of episodes
//	@desc This returns a list of episodes.
//	@returns torrentstream.EpisodeCollection
//	@param id - int - true - "AniList anime media ID"
//	@route /api/v1/torrentstream/episodes/{id} [GET]
func (h *Handler) HandleGetTorrentstreamEpisodeCollection(c echo.Context) error {
	mId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return h.RespondWithError(c, err)
	}

	ec, err := h.App.TorrentstreamRepository.NewEpisodeCollection(mId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	lop.ForEach(ec.Episodes, func(e *anime.Episode, _ int) {
		h.App.FillerManager.HydrateEpisodeFillerData(mId, e)
	})

	return h.RespondWithData(c, ec)
}

// HandleGetTorrentstreamSettings
//
//	@summary get torrentstream settings.
//	@desc This returns the torrentstream settings.
//	@returns models.TorrentstreamSettings
//	@route /api/v1/torrentstream/settings [GET]
func (h *Handler) HandleGetTorrentstreamSettings(c echo.Context) error {
	torrentstreamSettings, found := h.App.Database.GetTorrentstreamSettings()
	if !found {
		return h.RespondWithError(c, errors.New("torrent streaming settings not found"))
	}

	return h.RespondWithData(c, torrentstreamSettings)
}

// HandleSaveTorrentstreamSettings
//
//	@summary save torrentstream settings.
//	@desc This saves the torrentstream settings.
//	@desc The client should refetch the server status.
//	@returns models.TorrentstreamSettings
//	@route /api/v1/torrentstream/settings [PATCH]
func (h *Handler) HandleSaveTorrentstreamSettings(c echo.Context) error {

	type body struct {
		Settings models.TorrentstreamSettings `json:"settings"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	// Validate the download directory
	if b.Settings.DownloadDir != "" {
		dir, err := os.Stat(b.Settings.DownloadDir)
		if err != nil {
			h.App.Logger.Error().Err(err).Msgf("torrentstream: Download directory %s does not exist", b.Settings.DownloadDir)
			h.App.WSEventManager.SendEvent(events.ErrorToast, "Download directory does not exist")
			b.Settings.DownloadDir = ""
		}
		if !dir.IsDir() {
			h.App.Logger.Error().Msgf("torrentstream: Download directory %s is not a directory", b.Settings.DownloadDir)
			h.App.WSEventManager.SendEvent(events.ErrorToast, "Download directory is not a directory")
			b.Settings.DownloadDir = ""
		}
	}

	settings, err := h.App.Database.UpsertTorrentstreamSettings(&b.Settings)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	h.App.InitOrRefreshTorrentstreamSettings()

	return h.RespondWithData(c, settings)
}

// HandleGetTorrentstreamTorrentFilePreviews
//
//	@summary get list of torrent files from a batch
//	@desc This returns a list of file previews from the torrent
//	@returns []torrentstream.FilePreview
//	@route /api/v1/torrentstream/torrent-file-previews [POST]
func (h *Handler) HandleGetTorrentstreamTorrentFilePreviews(c echo.Context) error {
	type body struct {
		Torrent       *hibiketorrent.AnimeTorrent `json:"torrent"`
		EpisodeNumber int                         `json:"episodeNumber"`
		Media         *anilist.BaseAnime          `json:"media"`
	}
	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	providerExtension, ok := h.App.ExtensionRepository.GetAnimeTorrentProviderExtensionByID(b.Torrent.Provider)
	if !ok {
		return h.RespondWithError(c, errors.New("torrentstream: Torrent provider extension not found"))
	}

	magnet, err := providerExtension.GetProvider().GetTorrentMagnetLink(b.Torrent)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	// Get the media metadata
	animeMetadata, _ := h.App.MetadataProvider.GetAnimeMetadata(metadata.AnilistPlatform, b.Media.ID)
	absoluteOffset := 0
	if animeMetadata != nil {
		absoluteOffset = animeMetadata.GetOffset()
	}

	files, err := h.App.TorrentstreamRepository.GetTorrentFilePreviewsFromManualSelection(&torrentstream.GetTorrentFilePreviewsOptions{
		Torrent:        b.Torrent,
		Magnet:         magnet,
		EpisodeNumber:  b.EpisodeNumber,
		AbsoluteOffset: absoluteOffset,
		Media:          b.Media,
	})
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, files)
}

// HandleTorrentstreamStartStream
//
//	@summary starts a torrent stream.
//	@desc This starts the entire streaming process.
//	@returns bool
//	@route /api/v1/torrentstream/start [POST]
func (h *Handler) HandleTorrentstreamStartStream(c echo.Context) error {

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
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	userAgent := c.Request().Header.Get("User-Agent")

	err := h.App.TorrentstreamRepository.StartStream(&torrentstream.StartStreamOptions{
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
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleTorrentstreamStopStream
//
//	@summary stop a torrent stream.
//	@desc This stops the entire streaming process and drops the torrent if it's below a threshold.
//	@desc This is made to be used while the stream is running.
//	@returns bool
//	@route /api/v1/torrentstream/stop [POST]
func (h *Handler) HandleTorrentstreamStopStream(c echo.Context) error {

	err := h.App.TorrentstreamRepository.StopStream()
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleTorrentstreamDropTorrent
//
//	@summary drops a torrent stream.
//	@desc This stops the entire streaming process and drops the torrent completely.
//	@desc This is made to be used to force drop a torrent.
//	@returns bool
//	@route /api/v1/torrentstream/drop [POST]
func (h *Handler) HandleTorrentstreamDropTorrent(c echo.Context) error {

	err := h.App.TorrentstreamRepository.DropTorrent()
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleGetTorrentstreamBatchHistory
//
//	@summary returns the most recent batch selected.
//	@desc This returns the most recent batch selected.
//	@returns torrentstream.BatchHistoryResponse
//	@route /api/v1/torrentstream/batch-history [POST]
func (h *Handler) HandleGetTorrentstreamBatchHistory(c echo.Context) error {
	type body struct {
		MediaID int `json:"mediaId"`
	}
	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	ret := h.App.TorrentstreamRepository.GetBatchHistory(b.MediaID)
	return h.RespondWithData(c, ret)
}

// route /api/v1/torrentstream/stream/*
func (h *Handler) HandleTorrentstreamServeStream() http.Handler {
	return h.App.TorrentstreamRepository.HTTPStreamHandler()
}
