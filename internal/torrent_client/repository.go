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

func (r *Repository) Start() bool {
	switch r.Provider {
	case QbittorrentProvider:
		return r.QbittorrentClient.CheckStart()
	case TransmissionProvider:
		return r.Transmission.CheckStart()
	default:
		return false
	}
}
func (r *Repository) TorrentExists(hash string) bool {
	switch r.Provider {
	case QbittorrentProvider:
		p, err := r.QbittorrentClient.Torrent.GetProperties(hash)
		return err == nil && p != nil
	case TransmissionProvider:
		torrents, err := r.Transmission.Client.TorrentGetAllForHashes(context.Background(), []string{hash})
		return err == nil && len(torrents) > 0
	default:
		return false
	}
}

// GetList will return all torrents from the torrent client.
func (r *Repository) GetList() ([]*Torrent, error) {
	switch r.Provider {
	case QbittorrentProvider:
		torrents, err := r.QbittorrentClient.Torrent.GetList(&qbittorrent_model.GetTorrentListOptions{Filter: "all"})
		if err != nil {
			r.Logger.Err(err).Msg("torrent client: Error while getting torrent list (qBittorrent)")
			return nil, err
		}
		return r.FromQbitTorrents(torrents), nil
	case TransmissionProvider:
		torrents, err := r.Transmission.Client.TorrentGetAll(context.Background())
		if err != nil {
			r.Logger.Err(err).Msg("torrent client: Error while getting torrent list (Transmission)")
			return nil, err
		}
		return r.FromTransmissionTorrents(torrents), nil
	default:
		return nil, errors.New("torrent client: No torrent client provider found")
	}
}

// GetActiveTorrents will return all torrents that are currently downloading, paused or seeding.
func (r *Repository) GetActiveTorrents() ([]*Torrent, error) {
	torrents, err := r.GetList()
	if err != nil {
		return nil, err
	}
	var active []*Torrent
	for _, t := range torrents {
		if t.Status == TorrentStatusDownloading || t.Status == TorrentStatusSeeding || t.Status == TorrentStatusPaused {
			active = append(active, t)
		}
	}
	return active, nil
}
func (r *Repository) AddMagnets(magnets []string, dest string) error {
	r.Logger.Debug().Msg("torrent client: Adding magnets")

	var err error
	switch r.Provider {
	case QbittorrentProvider:
		err = r.QbittorrentClient.Torrent.AddURLs(magnets, &qbittorrent_model.AddTorrentsOptions{
			Savepath: dest,
		})
	case TransmissionProvider:
		for _, magnet := range magnets {
			_, err = r.Transmission.Client.TorrentAdd(context.Background(), transmissionrpc.TorrentAddPayload{
				Filename:    &magnet,
				DownloadDir: &dest,
			})
			if err != nil {
				r.Logger.Err(err).Msg("torrent client: Error while adding magnets (Transmission)")
				break
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
	switch r.Provider {
	case QbittorrentProvider:
		err = r.QbittorrentClient.Torrent.DeleteTorrents(hashes, true)
	case TransmissionProvider:
		torrents, err := r.Transmission.Client.TorrentGetAllForHashes(context.Background(), hashes)
		if err != nil {
			r.Logger.Err(err).Msg("torrent client: Error while fetching torrents (Transmission)")
			return err
		}
		ids := make([]int64, len(torrents))
		for i, t := range torrents {
			ids[i] = *t.ID
		}
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
	switch r.Provider {
	case QbittorrentProvider:
		err = r.QbittorrentClient.Torrent.StopTorrents(hashes)
	case TransmissionProvider:
		err = r.Transmission.Client.TorrentStopHashes(context.Background(), hashes)
	}

	if err != nil {
		r.Logger.Err(err).Msg("torrent client: Error while pausing torrents")
	}

	return err
}

func (r *Repository) ResumeTorrents(hashes []string) error {
	var err error
	switch r.Provider {
	case QbittorrentProvider:
		err = r.QbittorrentClient.Torrent.StopTorrents(hashes)
	case TransmissionProvider:
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
		size = util.ToHumanReadableSize(int(*t.TotalSize))
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
