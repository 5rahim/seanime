package downloader

import (
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/entities"
	"github.com/seanime-app/seanime-server/internal/events"
	"github.com/seanime-app/seanime-server/internal/qbittorrent"
	"github.com/seanime-app/seanime-server/internal/qbittorrent/model"
)

type (
	QbittorrentRepository struct {
		Logger         *zerolog.Logger
		Client         *qbittorrent.Client
		WSEventManager events.IWSEventManager
		Destination    string
	}

	SmartSelect struct {
		Magnets               []string
		Enabled               bool
		MissingEpisodeNumbers []int
		AbsoluteOffset        int
		Media                 *anilist.BaseMedia
	}

	TmpLocalFile struct {
		torrentContent *qbittorrent_model.TorrentContent
		localFile      *entities.LocalFile
		index          int
	}
)

func (r *QbittorrentRepository) AddMagnets(magnets []string) error {

	r.Logger.Debug().Msg("downloader: Adding magnets")

	err := r.Client.Torrent.AddURLs(magnets, &qbittorrent_model.AddTorrentsOptions{
		Savepath: r.Destination,
	})

	if err != nil {
		r.Logger.Err(err).Msg("downloader: Error while adding magnets")
	}

	return err

}

func (r *QbittorrentRepository) RemoveTorrents(hashes []string) error {

	err := r.Client.Torrent.DeleteTorrents(hashes, true)

	if err != nil {
		r.Logger.Err(err).Msg("downloader: Error while removing torrents")
	}

	return err

}

func (r *QbittorrentRepository) PauseTorrents(hashes []string) error {

	err := r.Client.Torrent.StopTorrents(hashes)

	if err != nil {
		r.Logger.Err(err).Msg("downloader: Error while pausing torrents")
	}

	return err

}

func (r *QbittorrentRepository) ResumeTorrents(hashes []string) error {

	err := r.Client.Torrent.StopTorrents(hashes)

	if err != nil {
		r.Logger.Err(err).Msg("downloader: Error while resuming torrents")
	}

	return err

}
