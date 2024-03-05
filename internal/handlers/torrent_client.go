package handlers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/anilist"
	qbittorrent_model "github.com/seanime-app/seanime/internal/qbittorrent/model"
	"github.com/seanime-app/seanime/internal/torrent"
	"github.com/seanime-app/seanime/internal/torrent_client"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/sourcegraph/conc/pool"
	"time"
)

const (
	TorrentStatusDownloading TorrentStatus = "downloading"
	TorrentStatusSeeding     TorrentStatus = "seeding"
	TorrentStatusPaused      TorrentStatus = "paused"
)

type (
	Torrent struct {
		Name        string        `json:"name"`
		Hash        string        `json:"hash"`
		Seeds       int           `json:"seeds"`
		UpSpeed     string        `json:"upSpeed"`
		DownSpeed   string        `json:"downSpeed"`
		Progress    float64       `json:"progress"`
		Size        string        `json:"size"`
		Eta         string        `json:"eta"`
		Status      TorrentStatus `json:"status"`
		ContentPath string        `json:"contentPath"`
	}
	TorrentStatus string
)

// HandleGetActiveTorrentList will return all active qBittorrent torrents. (i.e. downloading or seeding)
// This handler is used by the client to display the active torrents.
//
// DEVNOTE: Could be modified to support other torrent clients.
//
//	GET /v1/torrent-client/list
func HandleGetActiveTorrentList(c *RouteCtx) error {

	res, err := c.App.QBittorrent.Torrent.GetList(&qbittorrent_model.GetTorrentListOptions{
		Filter: "all",
	})
	if err != nil {
		c.App.QBittorrent.Start()
		timeout := time.After(time.Second * 15)
		ticker := time.NewTicker(time.Second * 1)
		open := make(chan struct{})
		defer ticker.Stop()
		go func() {
			for {
				select {
				case <-ticker.C:
					res, err = c.App.QBittorrent.Torrent.GetList(&qbittorrent_model.GetTorrentListOptions{
						Filter: "all",
					})
					if err == nil {
						close(open)
						return
					}
				case <-timeout:
					ticker.Stop()
					return
				}
			}
		}()

	work:
		for {
			select {
			case <-open:
				break work
			case <-timeout:
				return c.RespondWithError(err)
			}
		}
	}

	var torrents []*Torrent
	for _, torrent := range res {
		// skip torrents that are not downloading or seeding
		if torrent.State == qbittorrent_model.StatePausedUP ||
			torrent.State == qbittorrent_model.StateCheckingResumeData ||
			torrent.State == qbittorrent_model.StateUnknown ||
			torrent.State == qbittorrent_model.StateMissingFiles ||
			torrent.State == qbittorrent_model.StateError ||
			torrent.State == qbittorrent_model.StateMoving {
			continue
		}
		torrents = append(torrents, &Torrent{
			Name:        torrent.Name,
			Hash:        torrent.Hash,
			Seeds:       torrent.NumSeeds,
			UpSpeed:     util.ToHumanReadableSpeed(torrent.Upspeed),
			DownSpeed:   util.ToHumanReadableSpeed(torrent.Dlspeed),
			Progress:    torrent.Progress,
			Size:        util.ToHumanReadableSize(torrent.Size),
			Eta:         util.FormatETA(torrent.Eta),
			ContentPath: torrent.ContentPath,
			Status:      getTorrentStatus(torrent.State),
		})
	}

	return c.RespondWithData(torrents)

}

// HandleTorrentClientAction will perform an action on a torrent.
// It returns true if the action was successful.
//
//	POST /v1/torrent-client/action
//
// FIXME Animetosho
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
		err := c.App.QBittorrent.Torrent.StopTorrents([]string{b.Hash})
		if err != nil {
			return c.RespondWithError(err)
		}
	case "resume":
		err := c.App.QBittorrent.Torrent.ResumeTorrents([]string{b.Hash})
		if err != nil {
			return c.RespondWithError(err)
		}
	case "open":
		if b.Dir == "" {
			return c.RespondWithError(errors.New("directory not found"))
		}
		openDirInExplorer(b.Dir)
	}

	return c.RespondWithData(true)

}

// getTorrentStatus returns a normalized status for the torrent.
func getTorrentStatus(st qbittorrent_model.TorrentState) TorrentStatus {
	if st == qbittorrent_model.StateQueuedUP ||
		st == qbittorrent_model.StateStalledUP ||
		st == qbittorrent_model.StateForcedUP ||
		st == qbittorrent_model.StateCheckingUP ||
		st == qbittorrent_model.StateUploading {
		return TorrentStatusSeeding
	} else if st == qbittorrent_model.StatePausedDL {
		return TorrentStatusPaused
	} else {
		return TorrentStatusDownloading
	}
}

// HandleTorrentClientDownload will get magnets from Nyaa and add them to qBittorrent.
// It also handles smart selection (torrent_client.SmartSelect).
//
//	POST /v1/torrent-client/download
//
// FIXME Animetosho
func HandleTorrentClientDownload(c *RouteCtx) error {

	type body struct {
		Urls        []string `json:"urls"`
		Destination string   `json:"destination"`
		SmartSelect struct {
			Enabled               bool  `json:"enabled"`
			MissingEpisodeNumbers []int `json:"missingEpisodeNumbers"`
			AbsoluteOffset        int   `json:"absoluteOffset"`
		} `json:"smartSelect"`
		Media *anilist.BaseMedia `json:"media"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	// try to start qbittorrent if it's not running
	err := c.App.QBittorrent.Start()
	if err != nil {
		return c.RespondWithError(err)
	}

	// get magnets
	p := pool.NewWithResults[string]().WithErrors()

	for _, url := range b.Urls {
		p.Go(func() (string, error) {
			return torrent.GetTorrentMagnetFromUrl(url)
		})
	}

	// if we couldn't get a magnet, return error
	magnets, err := p.Wait()
	if err != nil {
		return c.RespondWithError(err)
	}

	// create repository
	repo := &torrent_client.TorrentClientRepository{
		Logger:            c.App.Logger,
		QbittorrentClient: c.App.QBittorrent,
		WSEventManager:    c.App.WSEventManager,
		Destination:       b.Destination,
	}

	// try to add torrents to qbittorrent, on error return error
	err = repo.AddMagnets(magnets)
	if err != nil {
		return c.RespondWithError(err)
	}

	err = repo.SmartSelect(&torrent_client.SmartSelect{
		Magnets:               magnets,
		Enabled:               b.SmartSelect.Enabled,
		MissingEpisodeNumbers: b.SmartSelect.MissingEpisodeNumbers,
		AbsoluteOffset:        b.SmartSelect.AbsoluteOffset,
		Media:                 b.Media,
	})
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)

}

// HandleTorrentClientAddMagnetFromRule will add the magnets to the torrent client based on the queued rule item.
//
// CLIENT: The AutoDownloader items should be re-fetched after this.
//
//	POST /v1/torrent-client/rule-magnet
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
	rule, err := c.App.Database.GetAutoDownloaderRule(b.RuleId)
	if err != nil {
		return c.RespondWithError(err)
	}

	// try to start qbittorrent if it's not running
	err = c.App.QBittorrent.Start()
	if err != nil {
		return c.RespondWithError(err)
	}

	// create repository
	repo := &torrent_client.TorrentClientRepository{
		Logger:            c.App.Logger,
		QbittorrentClient: c.App.QBittorrent,
		WSEventManager:    c.App.WSEventManager,
		Destination:       rule.Destination,
	}

	// try to add torrents to client, on error return error
	err = repo.AddMagnets([]string{b.MagnetUrl})
	if err != nil {
		return c.RespondWithError(err)
	}

	if b.QueuedItemId > 0 {
		// the magnet was added successfully, remove the item from the queue
		err = c.App.Database.DeleteAutoDownloaderItem(b.QueuedItemId)
	}

	return c.RespondWithData(true)

}
