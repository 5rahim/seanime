package handlers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/qbittorrent/model"
	"github.com/seanime-app/seanime/internal/util"
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

func HandleTorrentAction(c *RouteCtx) error {

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
