package torrent_client

import (
	"context"
	"errors"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/events"
	"seanime/internal/torrent_clients/qbittorrent"
	"seanime/internal/torrent_clients/qbittorrent/model"
	"seanime/internal/torrent_clients/transmission"
	"seanime/internal/torrents/torrent"
	"seanime/internal/util"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hekmon/transmissionrpc/v3"
	"github.com/rs/zerolog"
)

const (
	QbittorrentClient  = "qbittorrent"
	TransmissionClient = "transmission"
	NoneClient         = "none"
)

type (
	Repository struct {
		logger                      *zerolog.Logger
		qBittorrentClient           *qbittorrent.Client
		transmission                *transmission.Transmission
		torrentRepository           *torrent.Repository
		provider                    string
		metadataProviderRef         *util.Ref[metadata_provider.Provider]
		activeTorrentCountCtxCancel context.CancelFunc
		activeTorrentCount          *ActiveCount
	}

	NewRepositoryOptions struct {
		Logger              *zerolog.Logger
		QbittorrentClient   *qbittorrent.Client
		Transmission        *transmission.Transmission
		TorrentRepository   *torrent.Repository
		Provider            string
		MetadataProviderRef *util.Ref[metadata_provider.Provider]
	}

	ActiveCount struct {
		Downloading int `json:"downloading"`
		Seeding     int `json:"seeding"`
		Paused      int `json:"paused"`
	}
)

func NewRepository(opts *NewRepositoryOptions) *Repository {
	if opts.Provider == "" {
		opts.Provider = QbittorrentClient
	}
	return &Repository{
		logger:              opts.Logger,
		qBittorrentClient:   opts.QbittorrentClient,
		transmission:        opts.Transmission,
		torrentRepository:   opts.TorrentRepository,
		provider:            opts.Provider,
		metadataProviderRef: opts.MetadataProviderRef,
		activeTorrentCount:  &ActiveCount{},
	}
}

func (r *Repository) Shutdown() {
	if r.activeTorrentCountCtxCancel != nil {
		r.activeTorrentCountCtxCancel()
		r.activeTorrentCountCtxCancel = nil
	}
}

func (r *Repository) InitActiveTorrentCount(enabled bool, wsEventManager events.WSEventManagerInterface) {
	if r.activeTorrentCountCtxCancel != nil {
		r.activeTorrentCountCtxCancel()
	}

	if !enabled {
		return
	}

	var ctx context.Context
	ctx, r.activeTorrentCountCtxCancel = context.WithCancel(context.Background())
	go func(ctx context.Context) {
		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.GetActiveCount(r.activeTorrentCount)
				wsEventManager.SendEvent(events.ActiveTorrentCountUpdated, r.activeTorrentCount)
			}
		}
	}(ctx)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) GetProvider() string {
	return r.provider
}

func (r *Repository) Start() bool {
	switch r.provider {
	case QbittorrentClient:
		return r.qBittorrentClient.CheckStart()
	case TransmissionClient:
		return r.transmission.CheckStart()
	case NoneClient:
		return true
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

type GetListOptions struct {
	Category *string // qbittorrent only
	Sort     string  // name, name-desc, newest, oldest
}

var transmissionTorrentFields = []string{
	"name", "hashString", "peersSendingToUs", "rateUpload", "rateDownload",
	"percentDone", "totalSize", "eta", "status", "downloadDir", "addedDate", "isFinished",
}

// GetList will return all torrents from the torrent client.
func (r *Repository) GetList(opts *GetListOptions) ([]*Torrent, error) {
	// Normalize sort options
	sortBy := "added_on"
	reverse := true

	if opts.Sort != "" {
		switch opts.Sort {
		case "name":
			sortBy = "name"
			reverse = false
		case "name-desc":
			sortBy = "name"
			reverse = true
		case "newest":
			sortBy = "added_on"
			reverse = true
		case "oldest":
			sortBy = "added_on"
			reverse = false
		default:
		}
	}

	switch r.provider {
	case QbittorrentClient:
		torrents, err := r.qBittorrentClient.Torrent.GetList(&qbittorrent_model.GetTorrentListOptions{
			Filter:   "all",
			Category: opts.Category,
			Sort:     sortBy,
			Reverse:  reverse,
		})
		if err != nil {
			r.logger.Err(err).Msg("torrent client: Error while getting torrent list (qBittorrent)")
			return nil, err
		}
		return r.FromQbitTorrents(torrents), nil

	case TransmissionClient:
		torrents, err := r.transmission.Client.TorrentGet(context.Background(), transmissionTorrentFields, nil)
		if err != nil {
			r.logger.Err(err).Msg("torrent client: Error while getting torrent list (Transmission)")
			return nil, err
		}

		// Transmission does not sort server-side
		sort.Slice(torrents, func(i, j int) bool {
			t1, t2 := torrents[i], torrents[j]

			switch sortBy {
			case "added_on":
				// Handle nil dates safely
				var d1, d2 time.Time
				if t1.AddedDate != nil {
					d1 = *t1.AddedDate
				}
				if t2.AddedDate != nil {
					d2 = *t2.AddedDate
				}
				if reverse {
					return d1.After(d2) // Newest first
				}
				return d1.Before(d2) // Oldest first
			default: // "name"
				var n1, n2 string
				if t1.Name != nil {
					n1 = strings.ToLower(*t1.Name)
				}
				if t2.Name != nil {
					n2 = strings.ToLower(*t2.Name)
				}
				if reverse {
					return n1 > n2
				}
				return n1 < n2
			}
		})

		return r.FromTransmissionTorrents(torrents), nil

	default:
		return nil, errors.New("torrent client: No torrent client provider found")
	}
}

// GetActiveCount will return the count of active torrents (downloading, seeding, paused).
func (r *Repository) GetActiveCount(ret *ActiveCount) {
	ret.Seeding = 0
	ret.Downloading = 0
	ret.Paused = 0
	switch r.provider {
	case QbittorrentClient:
		torrents, err := r.qBittorrentClient.Torrent.GetList(&qbittorrent_model.GetTorrentListOptions{Filter: "downloading"})
		if err != nil {
			return
		}
		torrents2, err := r.qBittorrentClient.Torrent.GetList(&qbittorrent_model.GetTorrentListOptions{Filter: "seeding"})
		if err != nil {
			return
		}
		torrents = append(torrents, torrents2...)
		for _, t := range torrents {
			switch fromQbitTorrentStatus(t.State) {
			case TorrentStatusDownloading:
				ret.Downloading++
			case TorrentStatusSeeding:
				ret.Seeding++
			case TorrentStatusPaused:
				ret.Paused++
			}
		}
	case TransmissionClient:
		torrents, err := r.transmission.Client.TorrentGet(context.Background(), []string{"id", "status", "isFinished"}, nil)
		if err != nil {
			return
		}
		for _, t := range torrents {
			if t.Status == nil || t.IsFinished == nil {
				continue
			}
			switch fromTransmissionTorrentStatus(*t.Status, *t.IsFinished) {
			case TorrentStatusDownloading:
				ret.Downloading++
			case TorrentStatusSeeding:
				ret.Seeding++
			case TorrentStatusPaused:
				ret.Paused++
			}
		}
		return
	default:
		return
	}
}

// GetActiveTorrents will return all torrents that are currently downloading, paused or seeding.
func (r *Repository) GetActiveTorrents(opts *GetListOptions) ([]*Torrent, error) {
	torrents, err := r.GetList(opts)
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
	r.logger.Trace().Any("magnets", magnets).Msg("torrent client: Adding magnets")

	if len(magnets) == 0 {
		r.logger.Debug().Msg("torrent client: No magnets to add")
		return nil
	}

	var err error
	switch r.provider {
	case QbittorrentClient:
		err = r.qBittorrentClient.Torrent.AddURLs(magnets, &qbittorrent_model.AddTorrentsOptions{
			Savepath: dest,
			Tags:     r.qBittorrentClient.Tags,
			Category: r.qBittorrentClient.Category,
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
	case NoneClient:
		return errors.New("torrent client: No torrent client selected")
	}

	if err != nil {
		r.logger.Err(err).Msg("torrent client: Error while adding magnets")
		return err
	}

	r.logger.Debug().Msg("torrent client: Added torrents")

	return nil
}

func (r *Repository) RemoveTorrents(hashes []string) error {
	r.logger.Trace().Msg("torrent client: Removing torrents")

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
	r.logger.Trace().Msg("torrent client: Pausing torrents")

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
	r.logger.Trace().Msg("torrent client: Resuming torrents")

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
