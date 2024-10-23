package debrid_client

import (
	"context"
	"errors"
	"fmt"
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
	"seanime/internal/debrid/debrid"
	"seanime/internal/events"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/util"
	"strconv"
	"time"
)

type (
	StreamManager struct {
		repository            *Repository
		currentTorrentItemId  string
		downloadCtxCancelFunc context.CancelFunc
	}

	StreamPlaybackType string

	StreamStatus string

	StreamState struct {
		Status      StreamStatus `json:"status"`
		TorrentName string       `json:"torrentName"`
		Message     string       `json:"message"`
	}

	StartStreamOptions struct {
		MediaId       int
		EpisodeNumber int                         // RELATIVE Episode number to identify the file
		AniDBEpisode  string                      // Anizip episode
		Torrent       *hibiketorrent.AnimeTorrent // Selected torrent
		FileId        string                      // File ID or index
		UserAgent     string
		ClientId      string
		PlaybackType  StreamPlaybackType
	}

	CancelStreamOptions struct {
		// Whether to remove the torrent from the debrid service
		RemoveTorrent bool `json:"removeTorrent"`
	}
)

const (
	StreamStatusDownloading StreamStatus = "downloading"
	StreamStatusReady       StreamStatus = "ready"
	StreamStatusFailed      StreamStatus = "failed"
	StreamStatusStarted     StreamStatus = "started"
)

func NewStreamManager(repository *Repository) *StreamManager {
	return &StreamManager{
		repository:           repository,
		currentTorrentItemId: "",
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const (
	PlaybackTypeDefault        StreamPlaybackType = "default"
	PlaybackTypeExternalPlayer StreamPlaybackType = "externalPlayerLink"
)

// startStream is called by the client to start streaming a torrent
func (s *StreamManager) startStream(opts *StartStreamOptions) (err error) {
	defer util.HandlePanicInModuleWithError("debrid/client/StartStream", &err)

	s.repository.logger.Info().
		Str("clientId", opts.ClientId).
		Any("playbackType", opts.PlaybackType).
		Int("mediaId", opts.MediaId).Msgf("debridstream: Starting stream for episode %s", opts.AniDBEpisode)

	// Cancel the download context if it's running
	if s.downloadCtxCancelFunc != nil {
		s.downloadCtxCancelFunc()
		s.downloadCtxCancelFunc = nil
	}

	provider, err := s.repository.GetProvider()
	if err != nil {
		return fmt.Errorf("debridstream: Failed to start stream: %w", err)
	}

	s.repository.wsEventManager.SendEvent(events.DebridStreamState, StreamState{
		Status:      StreamStatusDownloading,
		TorrentName: opts.Torrent.Name,
		Message:     "Adding torrent...",
	})

	//
	// Get the media info
	//
	media, _, err := s.getMediaInfo(opts.MediaId)
	if err != nil {
		return err
	}

	episodeNumber := opts.EpisodeNumber
	aniDbEpisode := strconv.Itoa(episodeNumber)

	ctx, cancelCtx := context.WithCancel(context.Background())
	s.downloadCtxCancelFunc = cancelCtx

	// Add the torrent to the debrid service
	// For Torbox, this will automatically start downloading the torrent
	// For Real Debrid, this will just add the torrent to the user's account
	torrentItemId, err := provider.AddTorrent(debrid.AddTorrentOptions{
		MagnetLink: opts.Torrent.MagnetLink,
		InfoHash:   opts.Torrent.InfoHash,
	})
	if err != nil {
		return fmt.Errorf("debridstream: Failed to add torrent: %w", err)
	}

	time.Sleep(1 * time.Second)

	// Save the current torrent item id
	s.currentTorrentItemId = torrentItemId

	// Launch a goroutine that will listen to the added torrent's status
	go func(ctx context.Context) {
		defer util.HandlePanicInModuleThen("debrid/client/StartStream", func() {})

		defer func() {
			// Cancel the context
			if s.downloadCtxCancelFunc != nil {
				s.downloadCtxCancelFunc()
				s.downloadCtxCancelFunc = nil
			}
		}()

		s.repository.logger.Debug().Msg("debridstream: Listening to torrent status")

		s.repository.wsEventManager.SendEvent(events.DebridStreamState, StreamState{
			Status:      StreamStatusDownloading,
			TorrentName: opts.Torrent.Name,
			Message:     fmt.Sprintf("Downloading torrent..."),
		})

		itemCh := make(chan debrid.TorrentItem, 1)

		go func() {
			for item := range itemCh {
				s.repository.wsEventManager.SendEvent(events.DebridStreamState, StreamState{
					Status:      StreamStatusDownloading,
					TorrentName: item.Name,
					Message:     fmt.Sprintf("Downloading torrent: %d%%", item.CompletionPercentage),
				})
			}
		}()

		// Await the stream URL
		// For Torbox, this will wait until the entire torrent is downloaded
		streamUrl, err := provider.GetTorrentStreamUrl(ctx, debrid.StreamTorrentOptions{
			ID:     torrentItemId,
			FileId: opts.FileId,
		}, itemCh)

		go func() {
			close(itemCh)
		}()

		if err != nil {
			s.repository.logger.Err(err).Msg("debridstream: Failed to get stream URL")
			if !errors.Is(err, context.Canceled) {
				s.repository.wsEventManager.SendEvent(events.DebridStreamState, StreamState{
					Status:      StreamStatusFailed,
					TorrentName: opts.Torrent.Name,
					Message:     fmt.Sprintf("Failed to get stream URL, %v", err),
				})
			}
			return
		}

		s.repository.logger.Debug().Msg("debridstream: Stream URL received, checking stream file")

		// Check if we can stream the URL
		if canStream, reason := CanStream(streamUrl); !canStream {
			s.repository.logger.Warn().Msg("debridstream: Cannot stream the file")

			s.repository.wsEventManager.SendEvent(events.DebridStreamState, StreamState{
				Status:      StreamStatusFailed,
				TorrentName: opts.Torrent.Name,
				Message:     fmt.Sprintf("Cannot stream this file: %s", reason),
			})
			return
		}

		s.repository.logger.Debug().Msg("debridstream: Stream is ready")

		// Signal to the client that the torrent is ready to stream
		s.repository.wsEventManager.SendEvent(events.DebridStreamState, StreamState{
			Status:      StreamStatusReady,
			TorrentName: opts.Torrent.Name,
			Message:     "Ready to stream the file",
		})

		switch opts.PlaybackType {
		case PlaybackTypeDefault:
			//
			// Start the stream
			//
			s.repository.logger.Debug().Msg("debridstream: Starting the media player")
			// Sends the stream to the media player
			// DEVNOTE: Events are handled by the torrentstream.Repository module
			err = s.repository.playbackManager.StartStreamingUsingMediaPlayer(fmt.Sprintf("%s - Episode %s", opts.Torrent.Name, aniDbEpisode), &playbackmanager.StartPlayingOptions{
				Payload:   streamUrl,
				UserAgent: opts.UserAgent,
				ClientId:  opts.ClientId,
			}, media.ToBaseAnime(), aniDbEpisode)
			if err != nil {
				// Failed to start the stream, we'll drop the torrents and stop the server
				s.repository.wsEventManager.SendEvent(events.DebridStreamState, StreamState{
					Status:      StreamStatusFailed,
					TorrentName: opts.Torrent.Name,
					Message:     "Failed to send the stream to the media player",
				})
			}

		case PlaybackTypeExternalPlayer:
			// Send the external player link
			s.repository.wsEventManager.SendEventTo(opts.ClientId, events.ExternalPlayerOpenURL, struct {
				Url           string `json:"url"`
				MediaId       int    `json:"mediaId"`
				EpisodeNumber int    `json:"episodeNumber"`
			}{
				Url:           streamUrl,
				MediaId:       opts.MediaId,
				EpisodeNumber: opts.EpisodeNumber,
			})

			// Signal to the client that the torrent has started playing (remove loading status)
			// We can't know for sure
			s.repository.wsEventManager.SendEvent(events.DebridStreamState, StreamState{
				Status:      StreamStatusReady,
				TorrentName: opts.Torrent.Name,
				Message:     "External player link sent",
			})
		}
	}(ctx)

	s.repository.wsEventManager.SendEvent(events.DebridStreamState, StreamState{
		Status:      StreamStatusStarted,
		TorrentName: opts.Torrent.Name,
		Message:     "Stream started",
	})
	s.repository.logger.Info().Msg("debridstream: Stream started")

	return nil
}

func (s *StreamManager) cancelStream(opts *CancelStreamOptions) {
	if s.downloadCtxCancelFunc != nil {
		s.downloadCtxCancelFunc()
		s.downloadCtxCancelFunc = nil
	}

	if opts.RemoveTorrent && s.currentTorrentItemId != "" {
		// Remove the torrent from the debrid service
		provider, err := s.repository.GetProvider()
		if err != nil {
			s.repository.logger.Err(err).Msg("debridstream: Failed to remove torrent")
			return
		}

		// Remove the torrent from the debrid service
		err = provider.DeleteTorrent(s.currentTorrentItemId)
		if err != nil {
			s.repository.logger.Err(err).Msg("debridstream: Failed to remove torrent")
		}
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
