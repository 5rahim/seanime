package torrent_client

import (
	"seanime/internal/torrent_clients/builtin_client"
	qbittorrent_model "seanime/internal/torrent_clients/qbittorrent/model"
	"seanime/internal/util"
	"time"

	"github.com/hekmon/transmissionrpc/v3"
)

const (
	TorrentStatusDownloading TorrentStatus = "downloading"
	TorrentStatusSeeding     TorrentStatus = "seeding"
	TorrentStatusPaused      TorrentStatus = "paused"
	TorrentStatusOther       TorrentStatus = "other"
	TorrentStatusStopped     TorrentStatus = "stopped"
	TorrentStatusQueued      TorrentStatus = "queued"
	TorrentStatusError       TorrentStatus = "error"
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
		Peers       int           `json:"peers"`
		Ratio       float64       `json:"ratio"`
		AddedAt     time.Time     `json:"addedAt"`
		QueueIndex  int           `json:"queueIndex"`
		ForceStart  bool          `json:"forceStart"`
		Sequential  bool          `json:"sequential"`
		Error       string        `json:"error"`
	}
	TorrentStatus string
)

func (r *Repository) FromSeanimeTorrents(items []builtin_client.TorrentSnapshot) []*Torrent {
	ret := make([]*Torrent, 0, len(items))
	for _, item := range items {
		progress := 0.0
		if item.Length > 0 {
			progress = float64(item.Completed) / float64(item.Length)
		}
		status := TorrentStatusDownloading
		switch {
		case item.Error != "":
			status = TorrentStatusError
		case item.Paused:
			status = TorrentStatusPaused
		case item.Queued:
			status = TorrentStatusQueued
		case item.Length > 0 && item.Completed >= item.Length:
			status = TorrentStatusSeeding
		}
		eta := "N/A"
		if item.DownSpeed > 0 && item.Length > item.Completed {
			eta = util.FormatETA(int((item.Length - item.Completed) / item.DownSpeed))
		}
		ratio := 0.0
		if item.Downloaded > 0 {
			ratio = float64(item.Uploaded) / float64(item.Downloaded)
		}
		ret = append(ret, &Torrent{
			Name: item.Name, Hash: item.Hash, Seeds: item.Seeds, Peers: item.Peers,
			UpSpeed: util.ToHumanReadableSpeed(int(item.UpSpeed)), DownSpeed: util.ToHumanReadableSpeed(int(item.DownSpeed)),
			Progress: progress, Size: util.Bytes(uint64(item.Length)), Eta: eta, Status: status,
			ContentPath: item.Destination, Ratio: ratio, AddedAt: item.AddedAt, QueueIndex: item.QueueIndex,
			ForceStart: item.ForceStart, Sequential: item.Sequential, Error: item.Error,
		})
	}
	return ret
}

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
