package torrent_client

import (
	"context"
	"errors"
	"github.com/hekmon/transmissionrpc/v3"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/qbittorrent"
	"github.com/seanime-app/seanime/internal/qbittorrent/model"
	"github.com/seanime-app/seanime/internal/transmission"
)

const (
	QbittorrentProvider  = "qbittorrent"
	TransmissionProvider = "transmission"
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
