package handlers

import (
	"errors"
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
	"seanime/internal/api/anilist"
	"seanime/internal/database/db_bridge"
	"seanime/internal/events"
	"seanime/internal/torrent_clients/torrent_client"
	"seanime/internal/util"
)

// HandleGetActiveTorrentList
//
//	@summary returns all active torrents.
//	@desc This handler is used by the client to display the active torrents.
//
//	@route /api/v1/torrent-client/list [GET]
//	@returns []torrent_client.Torrent
func HandleGetActiveTorrentList(c *RouteCtx) error {

	// Get torrent list
	res, err := c.App.TorrentClientRepository.GetActiveTorrents()
	// If an error occurred, try to start the torrent client and get the list again
	// DEVNOTE: We try to get the list first because this route is called repeatedly by the client.
	if err != nil {
		ok := c.App.TorrentClientRepository.Start()
		if !ok {
			return c.RespondWithError(errors.New("could not start torrent client, verify your settings"))
		}
		res, err = c.App.TorrentClientRepository.GetActiveTorrents()
	}

	return c.RespondWithData(res)

}

// HandleTorrentClientAction
//
//	@summary performs an action on a torrent.
//	@desc This handler is used to pause, resume or remove a torrent.
//	@route /api/v1/torrent-client/action [POST]
//	@returns bool
func HandleTorrentClientAction(c *RouteCtx) error {

	type body struct {
		Hash   string `json:"hash"`
		Action string `json:"action"`
		Dir    string `json:"dir"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	if b.Hash == "" || b.Action == "" {
		return c.RespondWithError(errors.New("missing arguments"))
	}

	switch b.Action {
	case "pause":
		err := c.App.TorrentClientRepository.PauseTorrents([]string{b.Hash})
		if err != nil {
			return c.RespondWithError(err)
		}
	case "resume":
		err := c.App.TorrentClientRepository.ResumeTorrents([]string{b.Hash})
		if err != nil {
			return c.RespondWithError(err)
		}
	case "remove":
		err := c.App.TorrentClientRepository.RemoveTorrents([]string{b.Hash})
		if err != nil {
			return c.RespondWithError(err)
		}
	case "open":
		if b.Dir == "" {
			return c.RespondWithError(errors.New("directory not found"))
		}
		OpenDirInExplorer(b.Dir)
	}

	return c.RespondWithData(true)

}

// HandleTorrentClientDownload
//
//	@summary adds torrents to the torrent client.
//	@desc It fetches the magnets from the provided URLs and adds them to the torrent client.
//	@desc If smart select is enabled, it will try to select the best torrent based on the missing episodes.
//	@route /api/v1/torrent-client/download [POST]
//	@returns bool
func HandleTorrentClientDownload(c *RouteCtx) error {

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
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	// try to start torrent client if it's not running
	ok := c.App.TorrentClientRepository.Start()
	if !ok {
		return c.RespondWithError(errors.New("could not contact torrent client, verify your settings or make sure it's running"))
	}

	completeAnime, err := c.App.AnilistPlatform.GetAnimeWithRelations(b.Media.ID)
	if err != nil {
		return c.RespondWithError(err)
	}

	if b.SmartSelect.Enabled {
		if len(b.Torrents) > 1 {
			return c.RespondWithError(errors.New("smart select is not supported for multiple torrents"))
		}

		// smart select
		err = c.App.TorrentClientRepository.SmartSelect(&torrent_client.SmartSelectParams{
			Torrent:          &b.Torrents[0],
			EpisodeNumbers:   b.SmartSelect.MissingEpisodeNumbers,
			Media:            completeAnime,
			Destination:      b.Destination,
			Platform:         c.App.AnilistPlatform,
			ShouldAddTorrent: true,
		})
		if err != nil {
			return c.RespondWithError(err)
		}
	} else {

		// Get magnets
		magnets := make([]string, 0)
		for _, t := range b.Torrents {
			// Get the torrent's provider extension
			providerExtension, ok := c.App.TorrentRepository.GetAnimeProviderExtension(t.Provider)
			if !ok {
				return c.RespondWithError(errors.New("provider extension not found for torrent"))
			}
			// Get the torrent magnet link
			magnet, err := providerExtension.GetProvider().GetTorrentMagnetLink(&t)
			if err != nil {
				return c.RespondWithError(err)
			}

			magnets = append(magnets, magnet)
		}

		// try to add torrents to client, on error return error
		err = c.App.TorrentClientRepository.AddMagnets(magnets, b.Destination)
		if err != nil {
			return c.RespondWithError(err)
		}
	}

	// Add the media to the collection (if it wasn't already)
	go func() {
		defer util.HandlePanicInModuleThen("handlers/HandleTorrentClientDownload", func() {})
		if b.Media != nil {
			// Check if the media is already in the collection
			animeCollection, err := c.App.GetAnimeCollection(false)
			if err != nil {
				return
			}
			_, found := animeCollection.FindAnime(b.Media.ID)
			if found {
				return
			}
			// Add the media to the collection
			err = c.App.AnilistPlatform.AddMediaToCollection([]int{b.Media.ID})
			if err != nil {
				c.App.Logger.Error().Err(err).Msg("anilist: Failed to add media to collection")
			}
			ac, _ := c.App.RefreshAnimeCollection()
			c.App.WSEventManager.SendEvent(events.RefreshedAnilistAnimeCollection, ac)
		}
	}()

	return c.RespondWithData(true)

}

// HandleTorrentClientAddMagnetFromRule
//
//	@summary adds magnets to the torrent client based on the AutoDownloader item.
//	@desc This is used to download torrents that were queued by the AutoDownloader.
//	@desc The item will be removed from the queue if the magnet was added successfully.
//	@desc The AutoDownloader items should be re-fetched after this.
//	@route /api/v1/torrent-client/rule-magnet [POST]
//	@returns bool
func HandleTorrentClientAddMagnetFromRule(c *RouteCtx) error {

	type body struct {
		MagnetUrl    string `json:"magnetUrl"`
		RuleId       uint   `json:"ruleId"`
		QueuedItemId uint   `json:"queuedItemId"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	if b.MagnetUrl == "" || b.RuleId == 0 {
		return c.RespondWithError(errors.New("missing parameters"))
	}

	// Get rule from database
	rule, err := db_bridge.GetAutoDownloaderRule(c.App.Database, b.RuleId)
	if err != nil {
		return c.RespondWithError(err)
	}

	// try to start torrent client if it's not running
	ok := c.App.TorrentClientRepository.Start()
	if !ok {
		return c.RespondWithError(errors.New("could not start torrent client, verify your settings"))
	}

	// try to add torrents to client, on error return error
	err = c.App.TorrentClientRepository.AddMagnets([]string{b.MagnetUrl}, rule.Destination)
	if err != nil {
		return c.RespondWithError(err)
	}

	if b.QueuedItemId > 0 {
		// the magnet was added successfully, remove the item from the queue
		err = c.App.Database.DeleteAutoDownloaderItem(b.QueuedItemId)
	}

	return c.RespondWithData(true)

}
