package torrent_client

import (
	"github.com/hekmon/transmissionrpc/v3"
	"seanime/internal/torrents/qbittorrent/model"
	"seanime/internal/util"
)

const (
	TorrentStatusDownloading TorrentStatus = "downloading"
	TorrentStatusSeeding     TorrentStatus = "seeding"
	TorrentStatusPaused      TorrentStatus = "paused"
	TorrentStatusOther       TorrentStatus = "other"
	TorrentStatusStopped     TorrentStatus = "stopped"
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

func (r *Repository) FromTransmissionTorrents(t []transmissionrpc.Torrent) []*Torrent {
	ret := make([]*Torrent, 0, len(t))
	for _, t := range t {
		ret = append(ret, r.FromTransmissionTorrent(&t))
	}
	return ret
}

func (r *Repository) FromTransmissionTorrent(t *transmissionrpc.Torrent) *Torrent {
	name := "N/A"
	if t.Name != nil {
		name = *t.Name
	}

	hash := "N/A"
	if t.HashString != nil {
		hash = *t.HashString
	}

	seeds := 0
	if t.PeersSendingToUs != nil {
		seeds = int(*t.PeersSendingToUs)
	}

	upSpeed := "0 KB/s"
	if t.RateUpload != nil {
		upSpeed = util.ToHumanReadableSpeed(int(*t.RateUpload))
	}

	downSpeed := "0 KB/s"
	if t.RateDownload != nil {
		downSpeed = util.ToHumanReadableSpeed(int(*t.RateDownload))
	}

	progress := 0.0
	if t.PercentDone != nil {
		progress = *t.PercentDone
	}

	size := "N/A"
	if t.TotalSize != nil {
		size = util.ToHumanReadableSize(int64(*t.TotalSize))
	}

	eta := "???"
	if t.ETA != nil {
		eta = util.FormatETA(int(*t.ETA))
	}

	contentPath := ""
	if t.DownloadDir != nil {
		contentPath = *t.DownloadDir
	}

	status := TorrentStatusOther
	if t.Status != nil && t.IsFinished != nil {
		status = fromTransmissionTorrentStatus(*t.Status, *t.IsFinished)
	}

	return &Torrent{
		Name:        name,
		Hash:        hash,
		Seeds:       seeds,
		UpSpeed:     upSpeed,
		DownSpeed:   downSpeed,
		Progress:    progress,
		Size:        size,
		Eta:         eta,
		ContentPath: contentPath,
		Status:      status,
	}
}

// fromTransmissionTorrentStatus returns a normalized status for the torrent.
func fromTransmissionTorrentStatus(st transmissionrpc.TorrentStatus, isFinished bool) TorrentStatus {
	if st == transmissionrpc.TorrentStatusSeed || st == transmissionrpc.TorrentStatusSeedWait {
		return TorrentStatusSeeding
	} else if st == transmissionrpc.TorrentStatusStopped && isFinished {
		return TorrentStatusStopped
	} else if st == transmissionrpc.TorrentStatusStopped && !isFinished {
		return TorrentStatusPaused
	} else if st == transmissionrpc.TorrentStatusDownload || st == transmissionrpc.TorrentStatusDownloadWait {
		return TorrentStatusDownloading
	} else {
		return TorrentStatusOther
	}
}

func (r *Repository) FromQbitTorrents(t []*qbittorrent_model.Torrent) []*Torrent {
	ret := make([]*Torrent, 0, len(t))
	for _, t := range t {
		ret = append(ret, r.FromQbitTorrent(t))
	}
	return ret
}
func (r *Repository) FromQbitTorrent(t *qbittorrent_model.Torrent) *Torrent {
	return &Torrent{
		Name:        t.Name,
		Hash:        t.Hash,
		Seeds:       t.NumSeeds,
		UpSpeed:     util.ToHumanReadableSpeed(t.Upspeed),
		DownSpeed:   util.ToHumanReadableSpeed(t.Dlspeed),
		Progress:    t.Progress,
		Size:        util.ToHumanReadableSize(int64(t.Size)),
		Eta:         util.FormatETA(t.Eta),
		ContentPath: t.ContentPath,
		Status:      fromQbitTorrentStatus(t.State),
	}
}

// fromQbitTorrentStatus returns a normalized status for the torrent.
func fromQbitTorrentStatus(st qbittorrent_model.TorrentState) TorrentStatus {
	if st == qbittorrent_model.StateQueuedUP ||
		st == qbittorrent_model.StateStalledUP ||
		st == qbittorrent_model.StateForcedUP ||
		st == qbittorrent_model.StateCheckingUP ||
		st == qbittorrent_model.StateUploading {
		return TorrentStatusSeeding
	} else if st == qbittorrent_model.StatePausedDL {
		return TorrentStatusPaused
	} else if st == qbittorrent_model.StateDownloading ||
		st == qbittorrent_model.StateCheckingDL ||
		st == qbittorrent_model.StateStalledDL ||
		st == qbittorrent_model.StateQueuedDL ||
		st == qbittorrent_model.StateMetaDL ||
		st == qbittorrent_model.StateAllocating ||
		st == qbittorrent_model.StateForceDL {
		return TorrentStatusDownloading
	} else if st == qbittorrent_model.StatePausedUP {
		return TorrentStatusStopped
	} else {
		return TorrentStatusOther
	}
}
