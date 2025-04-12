package handlers

import (
	"errors"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/database/models"
	debrid_client "seanime/internal/debrid/client"
	"seanime/internal/debrid/debrid"
	"seanime/internal/events"

	hibiketorrent "seanime/internal/extension/hibike/torrent"

	"github.com/labstack/echo/v4"
)

// HandleGetDebridSettings
//
//	@summary get debrid settings.
//	@desc This returns the debrid settings.
//	@returns models.DebridSettings
//	@route /api/v1/debrid/settings [GET]
func (h *Handler) HandleGetDebridSettings(c echo.Context) error {
	debridSettings, found := h.App.Database.GetDebridSettings()
	if !found {
		return h.RespondWithError(c, errors.New("debrid settings not found"))
	}

	return h.RespondWithData(c, debridSettings)
}

// HandleSaveDebridSettings
//
//	@summary save debrid settings.
//	@desc This saves the debrid settings.
//	@desc The client should refetch the server status.
//	@returns models.DebridSettings
//	@route /api/v1/debrid/settings [PATCH]
func (h *Handler) HandleSaveDebridSettings(c echo.Context) error {

	type body struct {
		Settings models.DebridSettings `json:"settings"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	settings, err := h.App.Database.UpsertDebridSettings(&b.Settings)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	h.App.InitOrRefreshDebridSettings()

	return h.RespondWithData(c, settings)
}

// HandleDebridAddTorrents
//
//	@summary add torrent to debrid.
//	@desc This adds a torrent to the debrid service.
//	@returns bool
//	@route /api/v1/debrid/torrents [POST]
func (h *Handler) HandleDebridAddTorrents(c echo.Context) error {

	type body struct {
		Torrents    []hibiketorrent.AnimeTorrent `json:"torrents"`
		Media       *anilist.BaseAnime           `json:"media"`
		Destination string                       `json:"destination"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	if !h.App.DebridClientRepository.HasProvider() {
		return h.RespondWithError(c, errors.New("debrid provider not set"))
	}

	for _, torrent := range b.Torrents {
		// Get the torrent's provider extension
		animeTorrentProviderExtension, ok := h.App.TorrentRepository.GetAnimeProviderExtension(torrent.Provider)
		if !ok {
			return h.RespondWithError(c, errors.New("provider extension not found for torrent"))
		}

		magnet, err := animeTorrentProviderExtension.GetProvider().GetTorrentMagnetLink(&torrent)
		if err != nil {
			if len(b.Torrents) == 1 {
				return h.RespondWithError(c, err)
			} else {
				h.App.Logger.Err(err).Msg("debrid: Failed to get magnet link")
				h.App.WSEventManager.SendEvent(events.ErrorToast, err.Error())
				continue
			}
		}

		torrent.MagnetLink = magnet

		// Add the torrent to the debrid service
		_, err = h.App.DebridClientRepository.AddAndQueueTorrent(debrid.AddTorrentOptions{
			MagnetLink:   magnet,
			SelectFileId: "all",
		}, b.Destination, b.Media.ID)
		if err != nil {
			// If there is only one torrent, return the error
			if len(b.Torrents) == 1 {
				return h.RespondWithError(c, err)
			} else {
				// If there are multiple torrents, send an error toast and continue to the next torrent
				h.App.Logger.Err(err).Msg("debrid: Failed to add torrent to debrid")
				h.App.WSEventManager.SendEvent(events.ErrorToast, err.Error())
				continue
			}
		}
	}

	return h.RespondWithData(c, true)
}

// HandleDebridDownloadTorrent
//
//	@summary download torrent from debrid.
//	@desc Manually downloads a torrent from the debrid service locally.
//	@returns bool
//	@route /api/v1/debrid/torrents/download [POST]
func (h *Handler) HandleDebridDownloadTorrent(c echo.Context) error {

	type body struct {
		TorrentItem debrid.TorrentItem `json:"torrentItem"`
		Destination string             `json:"destination"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	if !filepath.IsAbs(b.Destination) {
		return h.RespondWithError(c, errors.New("destination must be an absolute path"))
	}

	// Remove the torrent from the database
	// This is done so that the torrent is not downloaded automatically
	// We ignore the error here because the torrent might not be in the database
	_ = h.App.Database.DeleteDebridTorrentItemByTorrentItemId(b.TorrentItem.ID)

	// Download the torrent locally
	err := h.App.DebridClientRepository.DownloadTorrent(b.TorrentItem, b.Destination)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleDebridCancelDownload
//
//	@summary cancel download from debrid.
//	@desc This cancels a download from the debrid service.
//	@returns bool
//	@route /api/v1/debrid/torrents/cancel [POST]
func (h *Handler) HandleDebridCancelDownload(c echo.Context) error {

	type body struct {
		ItemID string `json:"itemID"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	err := h.App.DebridClientRepository.CancelDownload(b.ItemID)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleDebridDeleteTorrent
//
//	@summary remove torrent from debrid.
//	@desc This removes a torrent from the debrid service.
//	@returns bool
//	@route /api/v1/debrid/torrent [DELETE]
func (h *Handler) HandleDebridDeleteTorrent(c echo.Context) error {

	type body struct {
		TorrentItem debrid.TorrentItem `json:"torrentItem"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	provider, err := h.App.DebridClientRepository.GetProvider()
	if err != nil {
		return h.RespondWithError(c, err)
	}

	err = provider.DeleteTorrent(b.TorrentItem.ID)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleDebridGetTorrents
//
//	@summary get torrents from debrid.
//	@desc This gets the torrents from the debrid service.
//	@returns []debrid.TorrentItem
//	@route /api/v1/debrid/torrents [GET]
func (h *Handler) HandleDebridGetTorrents(c echo.Context) error {

	provider, err := h.App.DebridClientRepository.GetProvider()
	if err != nil {
		return h.RespondWithError(c, err)
	}

	torrents, err := provider.GetTorrents()
	if err != nil {
		h.App.Logger.Err(err).Msg("debrid: Failed to get torrents")
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, torrents)
}

// HandleDebridGetTorrentInfo
//
//	@summary get torrent info from debrid.
//	@desc This gets the torrent info from the debrid service.
//	@returns debrid.TorrentInfo
//	@route /api/v1/debrid/torrents/info [POST]
func (h *Handler) HandleDebridGetTorrentInfo(c echo.Context) error {
	type body struct {
		Torrent hibiketorrent.AnimeTorrent `json:"torrent"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	animeTorrentProviderExtension, ok := h.App.TorrentRepository.GetAnimeProviderExtension(b.Torrent.Provider)
	if !ok {
		return h.RespondWithError(c, errors.New("provider extension not found for torrent"))
	}

	magnet, err := animeTorrentProviderExtension.GetProvider().GetTorrentMagnetLink(&b.Torrent)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	b.Torrent.MagnetLink = magnet

	torrentInfo, err := h.App.DebridClientRepository.GetTorrentInfo(debrid.GetTorrentInfoOptions{
		MagnetLink: b.Torrent.MagnetLink,
		InfoHash:   b.Torrent.InfoHash,
	})
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, torrentInfo)
}

// HandleDebridGetTorrentFilePreviews
//
//	@summary get list of torrent files
//	@returns []debrid_client.FilePreview
//	@route /api/v1/debrid/torrents/file-previews [POST]
func (h *Handler) HandleDebridGetTorrentFilePreviews(c echo.Context) error {
	type body struct {
		Torrent       *hibiketorrent.AnimeTorrent `json:"torrent"`
		EpisodeNumber int                         `json:"episodeNumber"`
		Media         *anilist.BaseAnime          `json:"media"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	animeTorrentProviderExtension, ok := h.App.TorrentRepository.GetAnimeProviderExtension(b.Torrent.Provider)
	if !ok {
		return h.RespondWithError(c, errors.New("provider extension not found for torrent"))
	}

	magnet, err := animeTorrentProviderExtension.GetProvider().GetTorrentMagnetLink(b.Torrent)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	b.Torrent.MagnetLink = magnet

	// Get the media
	animeMetadata, _ := h.App.MetadataProvider.GetAnimeMetadata(metadata.AnilistPlatform, b.Media.ID)
	absoluteOffset := 0
	if animeMetadata != nil {
		absoluteOffset = animeMetadata.GetOffset()
	}

	torrentInfo, err := h.App.DebridClientRepository.GetTorrentFilePreviewsFromManualSelection(&debrid_client.GetTorrentFilePreviewsOptions{
		Torrent:        b.Torrent,
		Magnet:         magnet,
		EpisodeNumber:  b.EpisodeNumber,
		Media:          b.Media,
		AbsoluteOffset: absoluteOffset,
	})
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, torrentInfo)
}

// HandleDebridStartStream
//
//	@summary start stream from debrid.
//	@desc This starts streaming a torrent from the debrid service.
//	@returns bool
//	@route /api/v1/debrid/stream/start [POST]
func (h *Handler) HandleDebridStartStream(c echo.Context) error {
	type body struct {
		MediaId       int                              `json:"mediaId"`
		EpisodeNumber int                              `json:"episodeNumber"`
		AniDBEpisode  string                           `json:"aniDBEpisode"`
		AutoSelect    bool                             `json:"autoSelect"`
		Torrent       *hibiketorrent.AnimeTorrent      `json:"torrent"`
		FileId        string                           `json:"fileId"`
		FileIndex     *int                             `json:"fileIndex"`
		PlaybackType  debrid_client.StreamPlaybackType `json:"playbackType"` // "default" or "externalPlayerLink"
		ClientId      string                           `json:"clientId"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	userAgent := c.Request().Header.Get("User-Agent")

	if b.Torrent != nil {
		animeTorrentProviderExtension, ok := h.App.TorrentRepository.GetAnimeProviderExtension(b.Torrent.Provider)
		if !ok {
			return h.RespondWithError(c, errors.New("provider extension not found for torrent"))
		}

		magnet, err := animeTorrentProviderExtension.GetProvider().GetTorrentMagnetLink(b.Torrent)
		if err != nil {
			return h.RespondWithError(c, err)
		}

		b.Torrent.MagnetLink = magnet
	}

	err := h.App.DebridClientRepository.StartStream(&debrid_client.StartStreamOptions{
		MediaId:       b.MediaId,
		EpisodeNumber: b.EpisodeNumber,
		AniDBEpisode:  b.AniDBEpisode,
		Torrent:       b.Torrent,
		FileId:        b.FileId,
		FileIndex:     b.FileIndex,
		UserAgent:     userAgent,
		ClientId:      b.ClientId,
		PlaybackType:  b.PlaybackType,
		AutoSelect:    b.AutoSelect,
	})
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleDebridCancelStream
//
//	@summary cancel stream from debrid.
//	@desc This cancels a stream from the debrid service.
//	@returns bool
//	@route /api/v1/debrid/stream/cancel [POST]
func (h *Handler) HandleDebridCancelStream(c echo.Context) error {
	type body struct {
		Options *debrid_client.CancelStreamOptions `json:"options"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	h.App.DebridClientRepository.CancelStream(b.Options)

	return h.RespondWithData(c, true)
}
