package torrent_client

import (
	"context"
	"errors"
	"github.com/hekmon/transmissionrpc/v3"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/qbittorrent"
	"github.com/seanime-app/seanime/internal/qbittorrent/model"
	"github.com/seanime-app/seanime/internal/transmission"
	"github.com/seanime-app/seanime/internal/util"
)

const (
	QbittorrentProvider  = "qbittorrent"
	TransmissionProvider = "transmission"
)

const (
	TorrentStatusDownloading TorrentStatus = "downloading"
	TorrentStatusSeeding     TorrentStatus = "seeding"
	TorrentStatusPaused      TorrentStatus = "paused"
	TorrentStatusOther       TorrentStatus = "other"
	TorrentStatusStopped     TorrentStatus = "stopped"
)

type (
	Repository struct {
		Logger            *zerolog.Logger
		QbittorrentClient *qbittorrent.Client
		Transmission      *transmission.Transmission
		Provider          string
	}

	NewRepositoryOptions struct {
		Logger            *zerolog.Logger
		QbittorrentClient *qbittorrent.Client
		Transmission      *transmission.Transmission
		Provider          string
	}

	SmartSelect struct {
		Magnets               []string
		Enabled               bool
		MissingEpisodeNumbers []int
		AbsoluteOffset        int
		Media                 *anilist.BaseMedia
		Destination           string
	}

	TmpLocalFile struct {
		torrentContent *qbittorrent_model.TorrentContent
		localFile      *entities.LocalFile
		index          int
	}

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
	TorrentStatus     string
	TorrentProperties struct {
	}
)

func NewRepository(opts *NewRepositoryOptions) *Repository {
	if opts.Provider == "" {
		opts.Provider = QbittorrentProvider
	}
	return &Repository{
		Logger:            opts.Logger,
		QbittorrentClient: opts.QbittorrentClient,
		Transmission:      opts.Transmission,
		Provider:          opts.Provider,
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) Start() error {
	panic("not implemented")
}
func (r *Repository) CheckStart() bool {
	panic("not implemented")
}
func (r *Repository) GetProperties(hash string) (*TorrentProperties, error) {
	panic("not implemented")
}

// GetList will return all torrents from the torrent client.
func (r *Repository) GetList() ([]*Torrent, error) {
	if r.Provider == QbittorrentProvider {

		torrents, err := r.QbittorrentClient.Torrent.GetList(&qbittorrent_model.GetTorrentListOptions{
			Filter: "all",
		})
		if err != nil {
			r.Logger.Err(err).Msg("torrent client: Error while getting torrent list (qBittorrent)")
			return nil, err
		}
		return r.FromQbitTorrents(torrents), nil

	} else if r.Provider == TransmissionProvider {

		torrents, err := r.Transmission.Client.TorrentGetAll(context.Background())
		if err != nil {
			r.Logger.Err(err).Msg("torrent client: Error while getting torrent list (Transmission)")
			return nil, err
		}
		return r.FromTransmissionTorrents(torrents), nil

	} else {
		return nil, errors.New("torrent client: No torrent client provider found")
	}
}
func (r *Repository) AddMagnets(magnets []string, dest string) error {

	r.Logger.Debug().Msg("torrent client: Adding magnets")
	var err error

	if r.Provider == QbittorrentProvider {
		err = r.QbittorrentClient.Torrent.AddURLs(magnets, &qbittorrent_model.AddTorrentsOptions{
			Savepath: dest,
		})
	} else if r.Provider == TransmissionProvider {
		for _, magnet := range magnets {
			_, err := r.Transmission.Client.TorrentAdd(context.Background(), transmissionrpc.TorrentAddPayload{
				Filename:    &magnet,
				DownloadDir: &dest,
			})
			if err != nil {
				r.Logger.Err(err).Msg("torrent client: Error while adding magnets (Transmission)")
				return err
			}
		}
	}

	if err != nil {
		r.Logger.Err(err).Msg("torrent client: Error while adding magnets")
	}

	return err

}

func (r *Repository) RemoveTorrents(hashes []string) error {

	var err error

	if r.Provider == QbittorrentProvider {

		err = r.QbittorrentClient.Torrent.DeleteTorrents(hashes, true)

	} else if r.Provider == TransmissionProvider {

		// 1. Get ids
		torrents, err := r.Transmission.Client.TorrentGetAllForHashes(context.Background(), hashes)
		if err != nil {
			r.Logger.Err(err).Msg("torrent client: Error while fetching torrents (Transmission)")
			return err
		}
		ids := make([]int64, 0, len(torrents))
		for _, t := range torrents {
			ids = append(ids, *t.ID)
		}
		// 2. Remove
		err = r.Transmission.Client.TorrentRemove(context.Background(), transmissionrpc.TorrentRemovePayload{
			IDs:             ids,
			DeleteLocalData: true,
		})
		if err != nil {
			r.Logger.Err(err).Msg("torrent client: Error while removing torrents (Transmission)")
			return err
		}

	}

	if err != nil {
		r.Logger.Err(err).Msg("torrent client: Error while removing torrents")
	}

	return err

}

func (r *Repository) PauseTorrents(hashes []string) error {

	var err error

	if r.Provider == QbittorrentProvider {
		err = r.QbittorrentClient.Torrent.StopTorrents(hashes)
	} else if r.Provider == TransmissionProvider {
		err = r.Transmission.Client.TorrentStopHashes(context.Background(), hashes)
	}

	if err != nil {
		r.Logger.Err(err).Msg("torrent client: Error while pausing torrents")
	}

	return err

}

func (r *Repository) ResumeTorrents(hashes []string) error {

	var err error

	if r.Provider == QbittorrentProvider {
		err = r.QbittorrentClient.Torrent.StopTorrents(hashes)
	} else if r.Provider == TransmissionProvider {
		err = r.Transmission.Client.TorrentStartHashes(context.Background(), hashes)
	}

	if err != nil {
		r.Logger.Err(err).Msg("torrent client: Error while resuming torrents")
	}

	return err

}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) FromTransmissionTorrents(t []transmissionrpc.Torrent) []*Torrent {
	ret := make([]*Torrent, 0, len(t))
	for _, t := range t {
		ret = append(ret, r.FromTransmissionTorrent(&t))
	}
	return ret
}

func (r *Repository) FromTransmissionTorrent(t *transmissionrpc.Torrent) *Torrent {
	return &Torrent{
		Name:        *t.Name,
		Hash:        *t.HashString,
		Seeds:       int(*t.PeersSendingToUs),
		UpSpeed:     util.ToHumanReadableSpeed(int(*t.RateUpload)),
		DownSpeed:   util.ToHumanReadableSpeed(int(*t.RateDownload)),
		Progress:    *t.PercentDone,
		Size:        util.ToHumanReadableSize(int(*t.TotalSize)),
		Eta:         util.FormatETA(int(*t.ETA)),
		ContentPath: *t.DownloadDir,
		Status:      fromTransmissionTorrentStatus(*t.Status, *t.IsFinished),
	}
}

// fromTransmissionTorrentStatus returns a normalized status for the torrent.
func fromTransmissionTorrentStatus(st transmissionrpc.TorrentStatus, isFinished bool) TorrentStatus {
	if st == transmissionrpc.TorrentStatusSeed || st == transmissionrpc.TorrentStatusSeedWait {
		return TorrentStatusSeeding
	} else if st == transmissionrpc.TorrentStatusStopped && isFinished {
		return TorrentStatusStopped
	} else if st == transmissionrpc.TorrentStatusStopped && !isFinished { // TODO Verify this
		return TorrentStatusPaused
	} else if st == transmissionrpc.TorrentStatusDownload || st == transmissionrpc.TorrentStatusDownloadWait {
		return TorrentStatusDownloading
	} else {
		return TorrentStatusOther
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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
		Size:        util.ToHumanReadableSize(t.Size),
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
