package torrentstream

import (
	"fmt"
	"github.com/samber/mo"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	itorrent "github.com/seanime-app/seanime/internal/torrents/torrent"
	"time"
)

type StartStreamOptions struct {
	MediaId       int                    `json:"mediaId"`
	EpisodeNumber int                    `json:"episodeNumber"` // RELATIVE Episode number to identify the file
	AniDBEpisode  string                 `json:"aniDBEpisode"`  // Anizip episode
	AutoSelect    bool                   `json:"autoSelect"`    // Automatically select the best file to stream
	Torrent       *itorrent.AnimeTorrent `json:"torrent"`       // Selected torrent
}

// StartStream is called by the client to start streaming a torrent
func (r *Repository) StartStream(opts *StartStreamOptions) error {
	// MY DUMBASS SHUT DOWN THE CLIENT BEFORE STARTING THE STREAM
	// NOT SHIT IT DIDN'T WORK! WASTED 2 DAYS TRYING TO DEBUG THIS SHIT
	//r.Shutdown()

	r.logger.Info().Int("mediaId", opts.MediaId).Msgf("torrentstream: Starting stream for episode %s", opts.AniDBEpisode)

	r.wsEventManager.SendEvent(eventTorrentLoading, nil)

	//
	// Get the media info
	//
	media, anizipMedia, err := r.getMediaInfo(opts.MediaId)
	if err != nil {
		return err
	}

	anizipEpisode, err := r.getEpisodeInfo(anizipMedia, opts.AniDBEpisode)
	if err != nil {
		return err
	}

	episodeNumber := opts.EpisodeNumber

	//
	// Find the best torrent / Select the torrent
	//
	var torrentToStream *playbackTorrent
	switch opts.AutoSelect {
	case true:
		torrentToStream, err = r.findBestTorrent(media, anizipMedia, anizipEpisode, episodeNumber)
		if err != nil {
			r.wsEventManager.SendEvent(eventTorrentLoadingFailed, nil)
			return err
		}
	case false:
		if opts.Torrent == nil {
			return fmt.Errorf("torrentstream: No torrent provided")
		}
		torrentToStream, err = r.findBestTorrentFromManualSelection(opts.Torrent.Link, media, anizipEpisode, episodeNumber)
		if err != nil {
			r.wsEventManager.SendEvent(eventTorrentLoadingFailed, nil)
			return err
		}
	}

	//
	// Set current file & torrent
	//
	r.client.currentFile = mo.Some(torrentToStream.File)
	r.client.currentTorrent = mo.Some(torrentToStream.Torrent)

	r.sendTorrentLoadingStatus(TLSStateStartingServer, "")

	//
	// Start the server
	//
	r.serverManager.startServer()

	r.sendTorrentLoadingStatus(TLSStateSendingStreamToMediaPlayer, "")

	go func() {
		for {
			// This is to make sure the client is ready to stream before we start the stream
			if r.client.readyToStream() {
				break
			}
			// If for some reason the torrent is dropped, we kill the goroutine
			if r.client.torrentClient.IsAbsent() || r.client.currentTorrent.IsAbsent() {
				return
			}
			r.logger.Debug().Msg("torrentstream: Waiting for playable threshold to be reached")
			time.Sleep(3 * time.Second)
		}

		//
		// Start the stream
		//
		r.logger.Debug().Msg("torrentstream: Starting the media player")
		err = r.playbackManager.StartStreamingUsingMediaPlayer(r.client.GetStreamingUrl(), media.ToBaseAnime(), anizipMedia, anizipEpisode)
		if err != nil {
			// Failed to start the stream, we'll drop the torrents and stop the server
			r.wsEventManager.SendEvent(eventTorrentLoadingFailed, nil)
			r.StopStream()
			r.logger.Error().Err(err).Msg("torrentstream: Failed to start the stream")
		}
	}()

	r.wsEventManager.SendEvent(eventTorrentLoaded, nil)
	r.logger.Info().Msg("torrentstream: Stream started")

	return nil
}

func (r *Repository) StopStream() error {
	r.logger.Info().Msg("torrentstream: Stopping stream")

	// Stop the client
	// This will stop the stream and close the server
	// This also sends the eventTorrentStopped event
	close(r.client.stopCh)

	r.logger.Info().Msg("torrentstream: Stream stopped")

	return nil
}

func (r *Repository) DropTorrent() error {
	r.logger.Info().Msg("torrentstream: Dropping last torrent")

	if r.client.torrentClient.IsAbsent() {
		return nil
	}

	for _, torrent := range r.client.torrentClient.MustGet().Torrents() {
		torrent.Drop()
	}

	// Also stop the server, since it's dropped
	r.serverManager.stopServer()
	r.mediaPlayerRepository.Stop()

	r.logger.Info().Msg("torrentstream: Dropped last torrent")

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) getMediaInfo(mediaId int) (media *anilist.CompleteAnime, anizipMedia *anizip.Media, err error) {
	// Get the media
	var found bool
	media, found = r.completeAnimeCache.Get(mediaId)
	if !found {
		// Fetch the media
		media, err = r.platform.GetAnimeWithRelations(mediaId)
		if err != nil {
			return nil, nil, fmt.Errorf("torrentstream: failed to fetch media: %w", err)
		}
	}

	// Get the media
	anizipMedia, err = anizip.FetchAniZipMediaC("anilist", mediaId, r.anizipCache)
	if err != nil {
		return nil, nil, fmt.Errorf("torrentstream: Could not fetch AniDB media: %w", err)
	}

	return
}

func (r *Repository) getEpisodeInfo(anizipMedia *anizip.Media, aniDBEpisode string) (episode *anizip.Episode, err error) {
	// Get the episode
	var found bool
	episode, found = anizipMedia.FindEpisode(aniDBEpisode)
	if !found {
		return nil, fmt.Errorf("torrentstream: Episode not found in the Anizip media")
	}
	return
}
