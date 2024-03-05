package torrent_client

import (
	"github.com/hekmon/transmissionrpc/v3"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/qbittorrent"
	"github.com/seanime-app/seanime/internal/qbittorrent/model"
)

const (
	QbittorrentProvider  = "qbittorrent"
	TransmissionProvider = "transmission"
)

type (
	TorrentClientRepository struct {
		Logger             *zerolog.Logger
		QbittorrentClient  *qbittorrent.Client
		TransmissionClient *transmissionrpc.Client
		WSEventManager     events.IWSEventManager
		Destination        string
		Provider           string
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

func (r *TorrentClientRepository) getProvider() string {
	if r.Provider == "" {
		return QbittorrentProvider
	}
	return r.Provider
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *TorrentClientRepository) AddMagnets(magnets []string) error {

	r.Logger.Debug().Msg("downloader: Adding magnets")
	var err error

	if r.getProvider() == QbittorrentProvider {
		err = r.QbittorrentClient.Torrent.AddURLs(magnets, &qbittorrent_model.AddTorrentsOptions{
			Savepath: r.Destination,
		})
	}

	if err != nil {
		r.Logger.Err(err).Msg("downloader: Error while adding magnets")
	}

	return err

}

func (r *TorrentClientRepository) RemoveTorrents(hashes []string) error {

	var err error

	if r.getProvider() == QbittorrentProvider {
		err = r.QbittorrentClient.Torrent.DeleteTorrents(hashes, true)
	}

	if err != nil {
		r.Logger.Err(err).Msg("downloader: Error while removing torrents")
	}

	return err

}

func (r *TorrentClientRepository) PauseTorrents(hashes []string) error {

	var err error

	if r.getProvider() == QbittorrentProvider {
		err = r.QbittorrentClient.Torrent.StopTorrents(hashes)
	}

	if err != nil {
		r.Logger.Err(err).Msg("downloader: Error while pausing torrents")
	}

	return err

}

func (r *TorrentClientRepository) ResumeTorrents(hashes []string) error {

	var err error

	if r.getProvider() == QbittorrentProvider {
		err = r.QbittorrentClient.Torrent.StopTorrents(hashes)
	}

	if err != nil {
		r.Logger.Err(err).Msg("downloader: Error while resuming torrents")
	}

	return err

}
