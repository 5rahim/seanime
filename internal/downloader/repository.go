package downloader

import (
	"errors"
	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/events"
	"github.com/seanime-app/seanime-server/internal/nyaa"
	"github.com/seanime-app/seanime-server/internal/qbittorrent"
	"github.com/seanime-app/seanime-server/internal/qbittorrent/model"
	"time"
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

func (r *QbittorrentRepository) SmartSelect(opts *SmartSelect) error {
	if !opts.Enabled {
		return nil
	}

	// 3. on smartSelect, wait for torrents to finish loading, then select the appropriate files
	// 3.1 - on duplicate episode numbers, return error
	// 3.2 - on missing episode numbers, return error

	if len(opts.Magnets) != 1 {
		return errors.New("incorrect number of magnets")
	}

	if opts.Enabled && opts.Media == nil {
		return errors.New("no media found")
	}

	magnet := opts.Magnets[0]
	// get hash
	hash, ok := nyaa.ExtractHashFromMagnet(magnet)
	if !ok {
		return errors.New("could not extract hash")
	}

	// ticker
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// Set a timeout of 30 seconds
	timeout := time.After(30 * time.Second)

	// exit
	done := make(chan struct{})

	var err error

	var contents []*qbittorrent_model.TorrentContent

	contentsChan := make(chan []*qbittorrent_model.TorrentContent)

	// get torrent contents when it's done loading
	go func() {
		for {
			select {
			case <-ticker.C:
				ret, _ := r.Client.Torrent.GetContents(hash)
				if ret != nil && len(ret) > 0 {
					contentsChan <- ret
				}
			case <-timeout:
				return
			}
		}
	}()

workDone:
	for {
		select {
		case <-done:
			break workDone
		case <-timeout:
			err = errors.New("timeout occurred: unable to retrieve torrent content")
			close(done)
		case contents = <-contentsChan:
			close(done)
		}
	}

	if err != nil {
		return err
	}

	println(spew.Sdump(contents))

	return nil
}
