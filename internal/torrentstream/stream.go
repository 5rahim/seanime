package torrentstream

import (
	"fmt"
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
	"github.com/anacrolix/torrent"
	"github.com/samber/mo"
	"seanime/internal/api/anilist"
	"seanime/internal/api/anizip"
	"seanime/internal/events"
	"seanime/internal/library/playbackmanager"
	"strconv"
	"time"
)

type PlaybackType string

const (
	PlaybackTypeDefault        PlaybackType = "default"
	PlaybackTypeExternalPlayer PlaybackType = "externalPlayerLink"
)

type StartStreamOptions struct {
	MediaId       int
	EpisodeNumber int                         // RELATIVE Episode number to identify the file
	AniDBEpisode  string                      // Anizip episode
	AutoSelect    bool                        // Automatically select the best file to stream
	Torrent       *hibiketorrent.AnimeTorrent // Selected torrent (Manual selection)
	FileIndex     *int                        // Index of the file to stream (Manual selection)
	UserAgent     string
	ClientId      string
	PlaybackType  PlaybackType
}

// StartStream is called by the client to start streaming a torrent
func (r *Repository) StartStream(opts *StartStreamOptions) error {
	// MY DUMBASS SHUT DOWN THE CLIENT BEFORE STARTING THE STREAM
	// NO SHIT IT DIDN'T WORK! WASTED 2 DAYS TRYING TO DEBUG THIS SHIT
	//r.Shutdown()

	r.logger.Info().
		Str("clientId", opts.ClientId).
		Any("playbackType", opts.PlaybackType).
		Int("mediaId", opts.MediaId).Msgf("torrentstream: Starting stream for episode %s", opts.AniDBEpisode)

	r.wsEventManager.SendEvent(eventTorrentLoading, nil)

	//
	// Get the media info
	//
	media, _, err := r.getMediaInfo(opts.MediaId)
	if err != nil {
		return err
	}

	episodeNumber := opts.EpisodeNumber
	aniDbEpisode := strconv.Itoa(episodeNumber)

	//
	// Find the best torrent / Select the torrent
	//
	var torrentToStream *playbackTorrent
	switch opts.AutoSelect {
	case true:
		torrentToStream, err = r.findBestTorrent(media, aniDbEpisode, episodeNumber)
		if err != nil {
			r.wsEventManager.SendEvent(eventTorrentLoadingFailed, nil)
			return err
		}
	case false:
		if opts.Torrent == nil {
			return fmt.Errorf("torrentstream: No torrent provided")
		}
		torrentToStream, err = r.findBestTorrentFromManualSelection(opts.Torrent, media, aniDbEpisode, opts.FileIndex)
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
		// Add the torrent to the history if it is a batch & manually selected
		if len(r.client.currentTorrent.MustGet().Files()) > 1 && opts.Torrent != nil {
			r.addBatchHistory(opts.MediaId, opts.Torrent) // ran in goroutine
		}

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
			time.Sleep(3 * time.Second) // Wait for 3 secs before checking again
		}

		switch opts.PlaybackType {
		case PlaybackTypeDefault:
			//
			// Start the stream
			//
			r.logger.Debug().Msg("torrentstream: Starting the media player")
			err = r.playbackManager.StartStreamingUsingMediaPlayer(&playbackmanager.StartPlayingOptions{
				Payload:   r.client.GetStreamingUrl(),
				UserAgent: opts.UserAgent,
				ClientId:  opts.ClientId,
			}, media.ToBaseAnime(), aniDbEpisode)
			if err != nil {
				// Failed to start the stream, we'll drop the torrents and stop the server
				r.wsEventManager.SendEvent(eventTorrentLoadingFailed, nil)
				_ = r.StopStream()
				r.logger.Error().Err(err).Msg("torrentstream: Failed to start the stream")
			}

		case PlaybackTypeExternalPlayer:
			// Send the external player link
			r.wsEventManager.SendEventTo(opts.ClientId, events.ExternalPlayerOpenURL, struct {
				Url           string `json:"url"`
				MediaId       int    `json:"mediaId"`
				EpisodeNumber int    `json:"episodeNumber"`
			}{
				Url:           r.client.GetStreamingUrl(),
				MediaId:       opts.MediaId,
				EpisodeNumber: opts.EpisodeNumber,
			})

			// Signal to the client that the torrent has started playing (remove loading status)
			// We can't know for sure
			r.wsEventManager.SendEvent(eventTorrentStartedPlaying, nil)
		}
	}()

	r.wsEventManager.SendEvent(eventTorrentLoaded, nil)
	r.logger.Info().Msg("torrentstream: Stream started")

	return nil
}

func (r *Repository) StopStream() error {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	r.logger.Info().Msg("torrentstream: Stopping stream")

	// Stop the client
	// This will stop the stream and close the server
	// This also sends the eventTorrentStopped event
	r.client.mu.Lock()
	//r.client.stopCh = make(chan struct{})
	r.client.repository.logger.Debug().Msg("torrentstream: Handling media player stopped event")
	// This is to prevent the client from downloading the whole torrent when the user stops watching
	// Also, the torrent might be a batch - so we don't want to download the whole thing
	if r.client.currentTorrent.IsPresent() {
		if r.client.currentTorrentStatus.ProgressPercentage < 70 {
			r.client.repository.logger.Debug().Msg("torrentstream: Dropping torrent, completion is less than 70%")
			r.client.dropTorrents()
		}
		r.client.repository.logger.Debug().Msg("torrentstream: Resetting current torrent and status")
	}
	r.client.currentTorrent = mo.None[*torrent.Torrent]()                  // Reset the current torrent
	r.client.currentFile = mo.None[*torrent.File]()                        // Reset the current file
	r.client.currentTorrentStatus = TorrentStatus{}                        // Reset the torrent status
	r.client.repository.serverManager.stopServer()                         // Stop streaming server
	r.client.repository.wsEventManager.SendEvent(eventTorrentStopped, nil) // Send torrent stopped event
	r.client.repository.mediaPlayerRepository.Stop()                       // Stop the media player gracefully if it's running
	r.client.mu.Unlock()

	r.logger.Info().Msg("torrentstream: Stream stopped")

	return nil
}

//func (r *Repository) StopStream() error {
//	defer func() {
//		if r := recover(); r != nil {
//		}
//	}()
//	r.logger.Info().Msg("torrentstream: Stopping stream")
//
//	// Stop the client
//	// This will stop the stream and close the server
//	// This also sends the eventTorrentStopped event
//	close(r.client.stopCh)
//
//	r.logger.Info().Msg("torrentstream: Stream stopped")
//
//	return nil
//}

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
	if anizipMedia == nil {
		return nil, fmt.Errorf("torrentstream: Anizip media is nil")
	}

	// Get the episode
	var found bool
	episode, found = anizipMedia.FindEpisode(aniDBEpisode)
	if !found {
		return nil, fmt.Errorf("torrentstream: Episode not found in the Anizip media")
	}
	return
}
