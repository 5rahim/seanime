package torrent_client

import (
	"context"
	"errors"
	"github.com/hekmon/transmissionrpc/v3"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/torrents/qbittorrent"
	"github.com/seanime-app/seanime/internal/torrents/qbittorrent/model"
	"github.com/seanime-app/seanime/internal/torrents/transmission"
	"strconv"
	"time"
)

const (
	QbittorrentClient  = "qbittorrent"
	TransmissionClient = "transmission"
)

type (
	Repository struct {
		logger            *zerolog.Logger
		qBittorrentClient *qbittorrent.Client
		transmission      *transmission.Transmission
		provider          string
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
		opts.Provider = QbittorrentClient
	}
	return &Repository{
		logger:            opts.Logger,
		qBittorrentClient: opts.QbittorrentClient,
		transmission:      opts.Transmission,
		provider:          opts.Provider,
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) Start() bool {
	switch r.provider {
	case QbittorrentClient:
		return r.qBittorrentClient.CheckStart()
	case TransmissionClient:
		return r.transmission.CheckStart()
	default:
		return false
	}
}
func (r *Repository) TorrentExists(hash string) bool {
	switch r.provider {
	case QbittorrentClient:
		p, err := r.qBittorrentClient.Torrent.GetProperties(hash)
		return err == nil && p != nil
	case TransmissionClient:
		torrents, err := r.transmission.Client.TorrentGetAllForHashes(context.Background(), []string{hash})
		return err == nil && len(torrents) > 0
	default:
		return false
	}
}

// GetList will return all torrents from the torrent client.
func (r *Repository) GetList() ([]*Torrent, error) {
	switch r.provider {
	case QbittorrentClient:
		torrents, err := r.qBittorrentClient.Torrent.GetList(&qbittorrent_model.GetTorrentListOptions{Filter: "all"})
		if err != nil {
			r.logger.Err(err).Msg("torrent client: Error while getting torrent list (qBittorrent)")
			return nil, err
		}
		return r.FromQbitTorrents(torrents), nil
	case TransmissionClient:
		torrents, err := r.transmission.Client.TorrentGetAll(context.Background())
		if err != nil {
			r.logger.Err(err).Msg("torrent client: Error while getting torrent list (Transmission)")
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
	r.logger.Debug().Msg("torrent client: Adding magnets")

	var err error
	switch r.provider {
	case QbittorrentClient:
		err = r.qBittorrentClient.Torrent.AddURLs(magnets, &qbittorrent_model.AddTorrentsOptions{
			Savepath: dest,
		})
	case TransmissionClient:
		for _, magnet := range magnets {
			_, err = r.transmission.Client.TorrentAdd(context.Background(), transmissionrpc.TorrentAddPayload{
				Filename:    &magnet,
				DownloadDir: &dest,
			})
			if err != nil {
				r.logger.Err(err).Msg("torrent client: Error while adding magnets (Transmission)")
				break
			}
		}
	}

	if err != nil {
		r.logger.Err(err).Msg("torrent client: Error while adding magnets")
		return err
	}

	r.logger.Debug().Msg("torrent client: Added torrents")

	return nil
}

func (r *Repository) RemoveTorrents(hashes []string) error {
	var err error
	switch r.provider {
	case QbittorrentClient:
		err = r.qBittorrentClient.Torrent.DeleteTorrents(hashes, true)
	case TransmissionClient:
		torrents, err := r.transmission.Client.TorrentGetAllForHashes(context.Background(), hashes)
		if err != nil {
			r.logger.Err(err).Msg("torrent client: Error while fetching torrents (Transmission)")
			return err
		}
		ids := make([]int64, len(torrents))
		for i, t := range torrents {
			ids[i] = *t.ID
		}
		err = r.transmission.Client.TorrentRemove(context.Background(), transmissionrpc.TorrentRemovePayload{
			IDs:             ids,
			DeleteLocalData: true,
		})
		if err != nil {
			r.logger.Err(err).Msg("torrent client: Error while removing torrents (Transmission)")
			return err
		}
	}
	if err != nil {
		r.logger.Err(err).Msg("torrent client: Error while removing torrents")
		return err
	}

	r.logger.Debug().Any("hashes", hashes).Msg("torrent client: Removed torrents")
	return nil
}

func (r *Repository) PauseTorrents(hashes []string) error {
	var err error
	switch r.provider {
	case QbittorrentClient:
		err = r.qBittorrentClient.Torrent.StopTorrents(hashes)
	case TransmissionClient:
		err = r.transmission.Client.TorrentStopHashes(context.Background(), hashes)
	}

	if err != nil {
		r.logger.Err(err).Msg("torrent client: Error while pausing torrents")
		return err
	}

	r.logger.Debug().Any("hashes", hashes).Msg("torrent client: Paused torrents")

	return nil
}

func (r *Repository) ResumeTorrents(hashes []string) error {
	var err error
	switch r.provider {
	case QbittorrentClient:
		err = r.qBittorrentClient.Torrent.ResumeTorrents(hashes)
	case TransmissionClient:
		err = r.transmission.Client.TorrentStartHashes(context.Background(), hashes)
	}

	if err != nil {
		r.logger.Err(err).Msg("torrent client: Error while resuming torrents")
		return err
	}

	r.logger.Debug().Any("hashes", hashes).Msg("torrent client: Resumed torrents")

	return nil
}

func (r *Repository) DeselectFiles(hash string, indices []int) error {

	var err error
	switch r.provider {
	case QbittorrentClient:
		strIndices := make([]string, len(indices), len(indices))
		for i, v := range indices {
			strIndices[i] = strconv.Itoa(v)
		}
		err = r.qBittorrentClient.Torrent.SetFilePriorities(hash, strIndices, 0)
	case TransmissionClient:
		torrents, err := r.transmission.Client.TorrentGetAllForHashes(context.Background(), []string{hash})
		if err != nil || torrents[0].ID == nil {
			r.logger.Err(err).Msg("torrent client: Error while deselecting files (Transmission)")
			return err
		}
		id := *torrents[0].ID
		ind := make([]int64, len(indices), len(indices))
		for i, v := range indices {
			ind[i] = int64(v)
		}
		err = r.transmission.Client.TorrentSet(context.Background(), transmissionrpc.TorrentSetPayload{
			FilesUnwanted: ind,
			IDs:           []int64{id},
		})
	}

	if err != nil {
		r.logger.Err(err).Msg("torrent client: Error while deselecting files")
		return err
	}

	r.logger.Debug().Str("hash", hash).Any("indices", indices).Msg("torrent client: Deselected torrent files")

	return nil
}

// GetFiles blocks until the files are retrieved, or until timeout.
func (r *Repository) GetFiles(hash string) (filenames []string, err error) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	filenames = make([]string, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	done := make(chan struct{})

	go func() {
		r.logger.Debug().Str("hash", hash).Msg("torrent client: Getting torrent files")
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				err = errors.New("torrent client: Unable to retrieve torrent files (timeout)")
				return
			case <-ticker.C:
				switch r.provider {
				case QbittorrentClient:
					qbitFiles, err := r.qBittorrentClient.Torrent.GetContents(hash)
					if err == nil && qbitFiles != nil && len(qbitFiles) > 0 {
						r.logger.Debug().Str("hash", hash).Int("count", len(qbitFiles)).Msg("torrent client: Retrieved torrent files")
						for _, f := range qbitFiles {
							filenames = append(filenames, f.Name)
						}
						return
					}
				case TransmissionClient:
					torrents, err := r.transmission.Client.TorrentGetAllForHashes(context.Background(), []string{hash})
					if err == nil && len(torrents) > 0 && torrents[0].Files != nil && len(torrents[0].Files) > 0 {
						transmissionFiles := torrents[0].Files
						r.logger.Debug().Str("hash", hash).Int("count", len(transmissionFiles)).Msg("torrent client: Retrieved torrent files")
						for _, f := range transmissionFiles {
							filenames = append(filenames, f.Name)
						}
						return
					}
				}
			}
		}
	}()

	<-done // wait for the files to be retrieved

	return
}
