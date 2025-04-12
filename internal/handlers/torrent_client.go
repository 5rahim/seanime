package handlers

import (
	"errors"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/database/db_bridge"
	"seanime/internal/events"
	"seanime/internal/torrent_clients/torrent_client"
	"seanime/internal/util"

	"github.com/labstack/echo/v4"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
)

// HandleGetActiveTorrentList
//
//	@summary returns all active torrents.
//	@desc This handler is used by the client to display the active torrents.
//
//	@route /api/v1/torrent-client/list [GET]
//	@returns []torrent_client.Torrent
func (h *Handler) HandleGetActiveTorrentList(c echo.Context) error {

	// Get torrent list
	res, err := h.App.TorrentClientRepository.GetActiveTorrents()
	// If an error occurred, try to start the torrent client and get the list again
	// DEVNOTE: We try to get the list first because this route is called repeatedly by the client.
	if err != nil {
		ok := h.App.TorrentClientRepository.Start()
		if !ok {
			return h.RespondWithError(c, errors.New("could not start torrent client, verify your settings"))
		}
		res, err = h.App.TorrentClientRepository.GetActiveTorrents()
	}

	return h.RespondWithData(c, res)

}

// HandleTorrentClientAction
//
//	@summary performs an action on a torrent.
//	@desc This handler is used to pause, resume or remove a torrent.
//	@route /api/v1/torrent-client/action [POST]
//	@returns bool
func (h *Handler) HandleTorrentClientAction(c echo.Context) error {

	type body struct {
		Hash   string `json:"hash"`
		Action string `json:"action"`
		Dir    string `json:"dir"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	if b.Hash == "" || b.Action == "" {
		return h.RespondWithError(c, errors.New("missing arguments"))
	}

	switch b.Action {
	case "pause":
		err := h.App.TorrentClientRepository.PauseTorrents([]string{b.Hash})
		if err != nil {
			return h.RespondWithError(c, err)
		}
	case "resume":
		err := h.App.TorrentClientRepository.ResumeTorrents([]string{b.Hash})
		if err != nil {
			return h.RespondWithError(c, err)
		}
	case "remove":
		err := h.App.TorrentClientRepository.RemoveTorrents([]string{b.Hash})
		if err != nil {
			return h.RespondWithError(c, err)
		}
	case "open":
		if b.Dir == "" {
			return h.RespondWithError(c, errors.New("directory not found"))
		}
		OpenDirInExplorer(b.Dir)
	}

	return h.RespondWithData(c, true)

}

// HandleTorrentClientDownload
//
//	@summary adds torrents to the torrent client.
//	@desc It fetches the magnets from the provided URLs and adds them to the torrent client.
//	@desc If smart select is enabled, it will try to select the best torrent based on the missing episodes.
//	@route /api/v1/torrent-client/download [POST]
//	@returns bool
func (h *Handler) HandleTorrentClientDownload(c echo.Context) error {

	type body struct {
		Torrents    []hibiketorrent.AnimeTorrent `json:"torrents"`
		Destination string                       `json:"destination"`
		SmartSelect struct {
			Enabled               bool  `json:"enabled"`
			MissingEpisodeNumbers []int `json:"missingEpisodeNumbers"`
		} `json:"smartSelect"`
		Media *anilist.BaseAnime `json:"media"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	if b.Destination == "" {
		return h.RespondWithError(c, errors.New("destination not found"))
	}

	if !filepath.IsAbs(b.Destination) {
		return h.RespondWithError(c, errors.New("destination path must be absolute"))
	}

	// Check that the destination path is a library path
	//libraryPaths, err := h.App.Database.GetAllLibraryPathsFromSettings()
	//if err != nil {
	//	return h.RespondWithError(c, err)
	//}
	//isInLibrary := util.IsSubdirectoryOfAny(libraryPaths, b.Destination)
	//if !isInLibrary {
	//	return h.RespondWithError(c, errors.New("destination path is not a library path"))
	//}

	// try to start torrent client if it's not running
	ok := h.App.TorrentClientRepository.Start()
	if !ok {
		return h.RespondWithError(c, errors.New("could not contact torrent client, verify your settings or make sure it's running"))
	}

	completeAnime, err := h.App.AnilistPlatform.GetAnimeWithRelations(b.Media.ID)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	if b.SmartSelect.Enabled {
		if len(b.Torrents) > 1 {
			return h.RespondWithError(c, errors.New("smart select is not supported for multiple torrents"))
		}

		// smart select
		err = h.App.TorrentClientRepository.SmartSelect(&torrent_client.SmartSelectParams{
			Torrent:          &b.Torrents[0],
			EpisodeNumbers:   b.SmartSelect.MissingEpisodeNumbers,
			Media:            completeAnime,
			Destination:      b.Destination,
			Platform:         h.App.AnilistPlatform,
			ShouldAddTorrent: true,
		})
		if err != nil {
			return h.RespondWithError(c, err)
		}
	} else {

		// Get magnets
		magnets := make([]string, 0)
		for _, t := range b.Torrents {
			// Get the torrent's provider extension
			providerExtension, ok := h.App.TorrentRepository.GetAnimeProviderExtension(t.Provider)
			if !ok {
				return h.RespondWithError(c, errors.New("provider extension not found for torrent"))
			}
			// Get the torrent magnet link
			magnet, err := providerExtension.GetProvider().GetTorrentMagnetLink(&t)
			if err != nil {
				return h.RespondWithError(c, err)
			}

			magnets = append(magnets, magnet)
		}

		// try to add torrents to client, on error return error
		err = h.App.TorrentClientRepository.AddMagnets(magnets, b.Destination)
		if err != nil {
			return h.RespondWithError(c, err)
		}
	}

	// Add the media to the collection (if it wasn't already)
	go func() {
		defer util.HandlePanicInModuleThen("handlers/HandleTorrentClientDownload", func() {})
		if b.Media != nil {
			// Check if the media is already in the collection
			animeCollection, err := h.App.GetAnimeCollection(false)
			if err != nil {
				return
			}
			_, found := animeCollection.FindAnime(b.Media.ID)
			if found {
				return
			}
			// Add the media to the collection
			err = h.App.AnilistPlatform.AddMediaToCollection([]int{b.Media.ID})
			if err != nil {
				h.App.Logger.Error().Err(err).Msg("anilist: Failed to add media to collection")
			}
			ac, _ := h.App.RefreshAnimeCollection()
			h.App.WSEventManager.SendEvent(events.RefreshedAnilistAnimeCollection, ac)
		}
	}()

	return h.RespondWithData(c, true)

}

// HandleTorrentClientAddMagnetFromRule
//
//	@summary adds magnets to the torrent client based on the AutoDownloader item.
//	@desc This is used to download torrents that were queued by the AutoDownloader.
//	@desc The item will be removed from the queue if the magnet was added successfully.
//	@desc The AutoDownloader items should be re-fetched after this.
//	@route /api/v1/torrent-client/rule-magnet [POST]
//	@returns bool
func (h *Handler) HandleTorrentClientAddMagnetFromRule(c echo.Context) error {

	type body struct {
		MagnetUrl    string `json:"magnetUrl"`
		RuleId       uint   `json:"ruleId"`
		QueuedItemId uint   `json:"queuedItemId"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	if b.MagnetUrl == "" || b.RuleId == 0 {
		return h.RespondWithError(c, errors.New("missing parameters"))
	}

	// Get rule from database
	rule, err := db_bridge.GetAutoDownloaderRule(h.App.Database, b.RuleId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	// try to start torrent client if it's not running
	ok := h.App.TorrentClientRepository.Start()
	if !ok {
		return h.RespondWithError(c, errors.New("could not start torrent client, verify your settings"))
	}

	// try to add torrents to client, on error return error
	err = h.App.TorrentClientRepository.AddMagnets([]string{b.MagnetUrl}, rule.Destination)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	if b.QueuedItemId > 0 {
		// the magnet was added successfully, remove the item from the queue
		err = h.App.Database.DeleteAutoDownloaderItem(b.QueuedItemId)
	}

	return h.RespondWithData(c, true)

}
