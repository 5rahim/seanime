package handlers

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/database/db_bridge"
	"seanime/internal/database/models"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/library/autodownloader"
	"seanime/internal/torrent_clients/torrent_client"
	torrentrepo "seanime/internal/torrents/torrent"
	"seanime/internal/util"

	"github.com/goccy/go-json"
	"github.com/labstack/echo/v4"
)

// HandleGetActiveTorrentList
//
//	@summary returns all active torrents.
//	@desc This handler is used by the client to display the active torrents.
//
//	@route /api/v1/torrent-client/list [GET]
//	@returns []torrent_client.Torrent
func (h *Handler) HandleGetActiveTorrentList(c echo.Context) error {
	var category *string
	if v := c.QueryParam("category"); v != "" {
		category = &v
	}
	sort := c.QueryParam("sort")

	// Get torrent list
	res, err := h.App.TorrentClientRepository.GetActiveTorrents(&torrent_client.GetListOptions{
		Category: category,
		Sort:     sort,
	})
	// If an error occurred, try to start the torrent client and get the list again
	// DEVNOTE: We try to get the list first because this route is called repeatedly by the client.
	if err != nil {
		if err := h.guardPrivilegedTorrentClient(c, h.App.Settings); err != nil {
			return err
		}
		ok := h.App.TorrentClientRepository.Start()
		if !ok {
			return h.RespondWithError(c, errors.New("could not start torrent client, verify your settings"))
		}
		res, err = h.App.TorrentClientRepository.GetActiveTorrents(&torrent_client.GetListOptions{
			Category: category,
			Sort:     sort,
		})
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
		Hash          string `json:"hash,omitempty"`
		Action        string `json:"action"`
		Dir           string `json:"dir,omitempty"`
		Tracker       string `json:"tracker,omitempty"`
		Name          string `json:"name,omitempty"`
		Value         bool   `json:"value,omitempty"`
		Index         int    `json:"index,omitempty"`
		Priority      int    `json:"priority,omitempty"`
		DownloadLimit int    `json:"downloadLimit,omitempty"`
		UploadLimit   int    `json:"uploadLimit,omitempty"`
		Magnet        string `json:"magnet,omitempty"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	if b.Action == "" {
		return h.RespondWithError(c, errors.New("missing arguments"))
	}
	globalAction := b.Action == "pause-all" || b.Action == "resume-all" || b.Action == "set-limits" || b.Action == "add-magnet"
	if !globalAction && b.Hash == "" {
		return h.RespondWithError(c, errors.New("missing torrent hash"))
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
		// Ensure the directory exists before attempting to open it, avoiding arbitrary string execution
		if b.Dir == "" {
			return h.RespondWithError(c, errors.New("directory not found"))
		}
		stat, err := os.Stat(b.Dir)
		if err != nil {
			return h.RespondWithError(c, errors.New("directory does not exist"))
		}
		// If it's a file, open its parent directory
		if !stat.IsDir() {
			b.Dir = filepath.Dir(b.Dir)
		}

		if err := h.guardPrivilegedLocalExecution(c); err != nil {
			return err
		}
		OpenDirInExplorer(b.Dir)
	case "pause-all", "resume-all", "force-start", "queue-up", "queue-down", "move-storage", "recheck", "reannounce", "add-tracker", "remove-tracker", "set-file-priority", "set-sequential", "rename", "set-limits", "add-magnet":
		client := h.App.TorrentClientRepository.GetSeanimeClient()
		if h.App.TorrentClientRepository.GetProvider() != torrent_client.SeanimeClient || client == nil {
			return h.RespondWithError(c, errors.New("action is only available for the Seanime torrent client"))
		}
		var err error
		switch b.Action {
		case "pause-all":
			err = client.PauseAll()
		case "resume-all":
			err = client.ResumeAll()
		case "force-start":
			err = client.SetForceStart(b.Hash, b.Value)
		case "queue-up":
			err = client.MoveQueue(b.Hash, -1)
		case "queue-down":
			err = client.MoveQueue(b.Hash, 1)
		case "move-storage":
			err = client.MoveStorage(b.Hash, b.Dir)
		case "recheck":
			go func(hash string) {
				if verifyErr := client.RecheckTorrent(hash); verifyErr != nil {
					h.App.Logger.Error().Err(verifyErr).Str("hash", hash).Msg("torrent client: recheck failed")
				}
			}(b.Hash)
		case "reannounce":
			err = client.ReannounceTorrent(b.Hash)
		case "add-tracker":
			err = client.AddTracker(b.Hash, b.Tracker)
		case "remove-tracker":
			err = client.RemoveTracker(b.Hash, b.Tracker)
		case "set-file-priority":
			err = client.SetFilePriority(b.Hash, b.Index, b.Priority)
		case "set-sequential":
			err = client.SetSequential(b.Hash, b.Value)
		case "rename":
			err = client.RenameTorrent(b.Hash, b.Name)
		case "set-limits":
			client.SetLimits(b.DownloadLimit, b.UploadLimit)
		case "add-magnet":
			_, err = client.AddMagnet(b.Magnet, b.Dir)
		}
		if err != nil {
			return h.RespondWithError(c, err)
		}
	}

	return h.RespondWithData(c, true)

}

// HandleGetBuiltInTorrentDetails
//
//	@summary returns details for a torrent managed by the Seanime torrent client.
//	@route /api/v1/torrent-client/details [GET]
//	@returns seanime.TorrentDetails
func (h *Handler) HandleGetBuiltInTorrentDetails(c echo.Context) error {
	hash := c.QueryParam("hash")
	if hash == "" {
		return h.RespondWithError(c, errors.New("missing torrent hash"))
	}
	client := h.App.TorrentClientRepository.GetSeanimeClient()
	if h.App.TorrentClientRepository.GetProvider() != torrent_client.SeanimeClient || client == nil {
		return h.RespondWithError(c, errors.New("Seanime torrent client is not active"))
	}
	details, err := client.GetTorrentDetails(hash)
	if err != nil {
		return h.RespondWithError(c, err)
	}
	return h.RespondWithData(c, details)
}

// HandleTorrentClientGetFiles
//
//	@summary gets the files of a torrent.
//	@desc This handler is used to get the files of a torrent.
//	@route /api/v1/torrent-client/get-files [POST]
//	@returns []string
func (h *Handler) HandleTorrentClientGetFiles(c echo.Context) error {

	type body struct {
		Torrent  *hibiketorrent.AnimeTorrent `json:"torrent"`
		Provider string                      `json:"provider"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	if b.Torrent == nil || b.Torrent.InfoHash == "" {
		return h.RespondWithError(c, errors.New("missing arguments"))
	}

	tempDir, err := os.MkdirTemp("", "torrent-")
	if err != nil {
		return h.RespondWithError(c, err)
	}
	defer os.RemoveAll(tempDir)

	// Get the magnet
	magnet, err := h.App.TorrentRepository.ResolveMagnetLink(b.Torrent)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	exists := h.App.TorrentClientRepository.TorrentExists(b.Torrent.InfoHash)

	if !exists {
		h.App.Logger.Info().Msgf("torrent client: Torrent %s does not exist, adding", b.Torrent.InfoHash)
		// Add the torrent
		err = h.App.TorrentClientRepository.AddMagnets([]string{magnet}, tempDir)
		if err != nil {
			return err
		}
	}

	h.App.Logger.Info().Msgf("torrent client: Getting files for %s", b.Torrent.InfoHash)
	files, err := h.App.TorrentClientRepository.GetFiles(b.Torrent.InfoHash)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	if !exists {
		h.App.Logger.Info().Msgf("torrent client: Removing torrent %s", b.Torrent.InfoHash)
		_ = h.App.TorrentClientRepository.RemoveTorrents([]string{b.Torrent.InfoHash})
	}

	return h.RespondWithData(c, files)
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
		Deselect struct {
			Enabled bool  `json:"enabled"`
			Indices []int `json:"indices"`
		} `json:"deselect,omitempty"`
		Media *anilist.BaseAnime `json:"media"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	if err := h.guardStrictLocalOnlyAction(c); err != nil {
		return err
	}

	if err := h.guardStrictFilesystemPath(c, b.Destination); err != nil {
		return err
	}

	if b.Destination == "" {
		return h.RespondWithError(c, errors.New("destination not found"))
	}

	if !filepath.IsAbs(b.Destination) {
		return h.RespondWithError(c, errors.New("destination path must be absolute"))
	}

	if err := h.guardStrictFilesystemPath(c, b.Destination); err != nil {
		return err
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
	if err := h.guardPrivilegedTorrentClient(c, h.App.Settings); err != nil {
		return err
	}
	ok := h.App.TorrentClientRepository.Start()
	if !ok {
		return h.RespondWithError(c, errors.New("could not contact torrent client, verify your settings or make sure it's running"))
	}

	var completeAnime *anilist.CompleteAnime
	var err error
	completeAnime, err = h.App.AnilistPlatformRef.Get().GetAnimeWithRelations(c.Request().Context(), b.Media.ID)
	if err != nil {
		completeAnime = b.Media.ToCompleteAnime()
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
			PlatformRef:      h.App.AnilistPlatformRef,
			ShouldAddTorrent: true,
		})
		if err != nil {
			return h.RespondWithError(c, err)
		}
	}

	if b.Deselect.Enabled {
		err = h.App.TorrentClientRepository.DeselectAndDownload(&torrent_client.DeselectAndDownloadParams{
			Torrent:          &b.Torrents[0],
			FileIndices:      b.Deselect.Indices,
			Destination:      b.Destination,
			ShouldAddTorrent: true,
		})
		if err != nil {
			return h.RespondWithError(c, err)
		}
	} else {

		// Get magnets
		magnets := make([]string, 0)
		for _, t := range b.Torrents {
			// Get the torrent magnet link
			magnet, err := h.App.TorrentRepository.ResolveMagnetLink(&t)
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
			err = h.App.AnilistPlatformRef.Get().AddMediaToCollection(c.Request().Context(), []int{b.Media.ID})
			if err != nil {
				h.App.Logger.Error().Err(err).Msg("anilist: Failed to add media to collection")
			}
			_, _ = h.App.RefreshAnimeCollection()
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

	if err := h.guardStrictLocalOnlyAction(c); err != nil {
		return err
	}

	if b.RuleId == 0 || (b.MagnetUrl == "" && b.QueuedItemId == 0) {
		return h.RespondWithError(c, errors.New("missing parameters"))
	}

	magnetURL := b.MagnetUrl
	if magnetURL == "" {
		item, err := h.App.Database.GetAutoDownloaderItem(b.QueuedItemId)
		if err != nil {
			return h.RespondWithError(c, err)
		}

		magnetURL, err = resolveAutoDownloaderItemMagnet(item, h.App.TorrentRepository)
		if err != nil {
			return h.RespondWithError(c, err)
		}

		if item.Magnet != magnetURL {
			item.Magnet = magnetURL
			if err := h.App.Database.UpdateAutoDownloaderItem(item.ID, item); err != nil {
				h.App.Logger.Warn().Err(err).Uint("queuedItemId", item.ID).Msg("torrent client: Failed to cache resolved queued magnet")
			}
		}
	}

	// Get rule from database
	rule, err := db_bridge.GetAutoDownloaderRule(h.App.Database, b.RuleId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	if !filepath.IsAbs(rule.Destination) {
		return h.RespondWithError(c, errors.New("destination path must be absolute"))
	}

	if err := h.guardStrictFilesystemPath(c, rule.Destination); err != nil {
		return err
	}

	// try to start torrent client if it's not running
	if err := h.guardPrivilegedTorrentClient(c, h.App.Settings); err != nil {
		return err
	}
	ok := h.App.TorrentClientRepository.Start()
	if !ok {
		return h.RespondWithError(c, errors.New("could not start torrent client, verify your settings"))
	}

	// try to add torrents to client, on error return error
	err = h.App.TorrentClientRepository.AddMagnets([]string{magnetURL}, rule.Destination)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	if b.QueuedItemId > 0 {
		// the magnet was added successfully, remove the item from the queue
		err = h.App.Database.DeleteAutoDownloaderItem(b.QueuedItemId)
	}

	return h.RespondWithData(c, true)

}

func resolveAutoDownloaderItemMagnet(item *models.AutoDownloaderItem, torrentRepository *torrentrepo.Repository) (string, error) {
	if item == nil {
		return "", errors.New("queued item not found")
	}

	if item.Magnet != "" {
		return item.Magnet, nil
	}

	fallbackHash := item.Hash
	var resolveErr error

	if len(item.TorrentData) > 0 {
		var storedTorrent autodownloader.NormalizedTorrent
		if err := json.Unmarshal(item.TorrentData, &storedTorrent); err != nil {
			resolveErr = err
		} else if storedTorrent.AnimeTorrent != nil {
			if fallbackHash == "" {
				fallbackHash = storedTorrent.AnimeTorrent.InfoHash
			}

			if storedTorrent.AnimeTorrent.Provider == "" && storedTorrent.ExtensionID != "" {
				storedTorrent.AnimeTorrent.Provider = storedTorrent.ExtensionID
			}

			if torrentRepository != nil {
				magnet, err := torrentRepository.ResolveMagnetLink(storedTorrent.AnimeTorrent)
				if err == nil && magnet != "" {
					return magnet, nil
				}
				resolveErr = err
			}
		}
	}

	if fallbackHash != "" {
		return fmt.Sprintf("magnet:?xt=urn:btih:%s", fallbackHash), nil
	}

	if resolveErr != nil {
		return "", resolveErr
	}

	return "", errors.New("magnet link not found")
}
