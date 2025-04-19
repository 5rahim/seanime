package torrent_client

import (
	"seanime/internal/torrent_clients/qbittorrent/model"
	"seanime/internal/util"

	"github.com/hekmon/transmissionrpc/v3"
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

//var torrentPool = util.NewPool[*Torrent](func() *Torrent {
//	return &Torrent{}
//})

func (r *Repository) FromTransmissionTorrents(t []transmissionrpc.Torrent) []*Torrent {
	ret := make([]*Torrent, 0, len(t))
	for _, t := range t {
		ret = append(ret, r.FromTransmissionTorrent(&t))
	}
	return ret
}

func (r *Repository) FromTransmissionTorrent(t *transmissionrpc.Torrent) *Torrent {
	torrent := &Torrent{}

	torrent.Name = "N/A"
	if t.Name != nil {
		torrent.Name = *t.Name
	}

	torrent.Hash = "N/A"
	if t.HashString != nil {
		torrent.Hash = *t.HashString
	}

	torrent.Seeds = 0
	if t.PeersSendingToUs != nil {
		torrent.Seeds = int(*t.PeersSendingToUs)
	}

	torrent.UpSpeed = "0 KB/s"
	if t.RateUpload != nil {
		torrent.UpSpeed = util.ToHumanReadableSpeed(int(*t.RateUpload))
	}

	torrent.DownSpeed = "0 KB/s"
	if t.RateDownload != nil {
		torrent.DownSpeed = util.ToHumanReadableSpeed(int(*t.RateDownload))
	}

	torrent.Progress = 0.0
	if t.PercentDone != nil {
		torrent.Progress = *t.PercentDone
	}

	torrent.Size = "N/A"
	if t.TotalSize != nil {
		torrent.Size = util.Bytes(uint64(*t.TotalSize))
	}

	torrent.Eta = "???"
	if t.ETA != nil {
		torrent.Eta = util.FormatETA(int(*t.ETA))
	}

	torrent.ContentPath = ""
	if t.DownloadDir != nil {
		torrent.ContentPath = *t.DownloadDir
	}

	torrent.Status = TorrentStatusOther
	if t.Status != nil && t.IsFinished != nil {
		torrent.Status = fromTransmissionTorrentStatus(*t.Status, *t.IsFinished)
	}

	return torrent
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
	torrent := &Torrent{}

	torrent.Name = t.Name
	torrent.Hash = t.Hash
	torrent.Seeds = t.NumSeeds
	torrent.UpSpeed = util.ToHumanReadableSpeed(t.Upspeed)
	torrent.DownSpeed = util.ToHumanReadableSpeed(t.Dlspeed)
	torrent.Progress = t.Progress
	torrent.Size = util.Bytes(uint64(t.Size))
	torrent.Eta = util.FormatETA(t.Eta)
	torrent.ContentPath = t.ContentPath
	torrent.Status = fromQbitTorrentStatus(t.State)

	return torrent
}

// fromQbitTorrentStatus returns a normalized status for the torrent.
func fromQbitTorrentStatus(st qbittorrent_model.TorrentState) TorrentStatus {
	if st == qbittorrent_model.StateQueuedUP ||
		st == qbittorrent_model.StateStalledUP ||
		st == qbittorrent_model.StateForcedUP ||
		st == qbittorrent_model.StateCheckingUP ||
		st == qbittorrent_model.StateUploading {
		return TorrentStatusSeeding
	} else if st == qbittorrent_model.StatePausedDL || st == qbittorrent_model.StateStoppedDL {
		return TorrentStatusPaused
	} else if st == qbittorrent_model.StateDownloading ||
		st == qbittorrent_model.StateCheckingDL ||
		st == qbittorrent_model.StateStalledDL ||
		st == qbittorrent_model.StateQueuedDL ||
		st == qbittorrent_model.StateMetaDL ||
		st == qbittorrent_model.StateAllocating ||
		st == qbittorrent_model.StateForceDL {
		return TorrentStatusDownloading
	} else if st == qbittorrent_model.StatePausedUP || st == qbittorrent_model.StateStoppedUP {
		return TorrentStatusStopped
	} else {
		return TorrentStatusOther
	}
}
