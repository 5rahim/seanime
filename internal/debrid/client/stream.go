package debrid_client

import (
	"context"
	"errors"
	"fmt"
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
	"seanime/internal/database/db_bridge"
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
		AutoSelect    bool
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

	//
	// Get the media info
	//
	media, _, err := s.getMediaInfo(opts.MediaId)
	if err != nil {
		return err
	}

	episodeNumber := opts.EpisodeNumber
	aniDbEpisode := strconv.Itoa(episodeNumber)

	selectedTorrent := opts.Torrent
	fileId := opts.FileId

	if opts.AutoSelect {

		s.repository.wsEventManager.SendEvent(events.DebridStreamState, StreamState{
			Status:      StreamStatusDownloading,
			TorrentName: "-",
			Message:     "Selecting best torrent...",
		})

		st, fi, err := s.repository.findBestTorrent(provider, media, opts.EpisodeNumber)
		if err != nil {
			s.repository.wsEventManager.SendEvent(events.DebridStreamState, StreamState{
				Status:      StreamStatusFailed,
				TorrentName: "-",
				Message:     fmt.Sprintf("Failed to select best torrent, %v", err),
			})
			return fmt.Errorf("debridstream: Failed to start stream: %w", err)
		}
		selectedTorrent = st
		fileId = fi
	}

	if selectedTorrent == nil {
		return fmt.Errorf("debridstream: Failed to start stream, no torrent provided")
	}

	s.repository.wsEventManager.SendEvent(events.DebridStreamState, StreamState{
		Status:      StreamStatusDownloading,
		TorrentName: selectedTorrent.Name,
		Message:     "Adding torrent...",
	})

	// Add the torrent to the debrid service
	torrentItemId, err := provider.AddTorrent(debrid.AddTorrentOptions{
		MagnetLink:   selectedTorrent.MagnetLink,
		InfoHash:     selectedTorrent.InfoHash,
		SelectFileId: fileId, // RD-only, download only the selected file
	})
	if err != nil {
		s.repository.wsEventManager.SendEvent(events.DebridStreamState, StreamState{
			Status:      StreamStatusFailed,
			TorrentName: selectedTorrent.Name,
			Message:     fmt.Sprintf("Failed to add torrent, %v", err),
		})
		return fmt.Errorf("debridstream: Failed to add torrent: %w", err)
	}

	time.Sleep(1 * time.Second)

	// Save the current torrent item id
	s.currentTorrentItemId = torrentItemId
	ctx, cancelCtx := context.WithCancel(context.Background())
	s.downloadCtxCancelFunc = cancelCtx

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
			TorrentName: selectedTorrent.Name,
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
			FileId: fileId,
		}, itemCh)

		go func() {
			close(itemCh)
		}()

		if ctx.Err() != nil {
			s.repository.logger.Debug().Msg("debridstream: Context cancelled, stopping stream")
			return
		}

		if err != nil {
			s.repository.logger.Err(err).Msg("debridstream: Failed to get stream URL")
			if !errors.Is(err, context.Canceled) {
				s.repository.wsEventManager.SendEvent(events.DebridStreamState, StreamState{
					Status:      StreamStatusFailed,
					TorrentName: selectedTorrent.Name,
					Message:     fmt.Sprintf("Failed to get stream URL, %v", err),
				})
			}
			return
		}

		s.repository.logger.Debug().Msg("debridstream: Stream URL received, checking stream file")
		s.repository.wsEventManager.SendEvent(events.DebridStreamState, StreamState{
			Status:      StreamStatusDownloading,
			TorrentName: selectedTorrent.Name,
			Message:     "Checking stream file...",
		})

		retries := 0

	streamUrlCheckLoop:
		for { // Retry loop for a total of 4 times (32 seconds)
			select {
			case <-ctx.Done():
				s.repository.logger.Debug().Msg("debridstream: Context cancelled, stopping stream")
				return
			default:
				// Check if we can stream the URL
				if canStream, reason := CanStream(streamUrl); !canStream {
					if retries >= 4 {
						s.repository.logger.Error().Msg("debridstream: Cannot stream the file")

						s.repository.wsEventManager.SendEvent(events.DebridStreamState, StreamState{
							Status:      StreamStatusFailed,
							TorrentName: selectedTorrent.Name,
							Message:     fmt.Sprintf("Cannot stream this file: %s", reason),
						})
						return
					}
					s.repository.logger.Warn().Msg("debridstream: Rechecking stream file in 8 seconds")
					s.repository.wsEventManager.SendEvent(events.DebridStreamState, StreamState{
						Status:      StreamStatusDownloading,
						TorrentName: selectedTorrent.Name,
						Message:     "Checking stream file...",
					})
					retries++
					time.Sleep(8 * time.Second)
					continue
				}
				break streamUrlCheckLoop
			}
		}

		s.repository.logger.Debug().Msg("debridstream: Stream is ready")

		// Signal to the client that the torrent is ready to stream
		s.repository.wsEventManager.SendEvent(events.DebridStreamState, StreamState{
			Status:      StreamStatusReady,
			TorrentName: selectedTorrent.Name,
			Message:     "Ready to stream the file",
		})

		if ctx.Err() != nil {
			s.repository.logger.Debug().Msg("debridstream: Context cancelled, stopping stream")
			return
		}

		switch opts.PlaybackType {
		case PlaybackTypeDefault:
			//
			// Start the stream
			//
			s.repository.logger.Debug().Msg("debridstream: Starting the media player")
			// Sends the stream to the media player
			// DEVNOTE: Events are handled by the torrentstream.Repository module
			err = s.repository.playbackManager.StartStreamingUsingMediaPlayer(fmt.Sprintf("%s - Episode %s", selectedTorrent.Name, aniDbEpisode), &playbackmanager.StartPlayingOptions{
				Payload:   streamUrl,
				UserAgent: opts.UserAgent,
				ClientId:  opts.ClientId,
			}, media.ToBaseAnime(), aniDbEpisode)
			if err != nil {
				// Failed to start the stream, we'll drop the torrents and stop the server
				s.repository.wsEventManager.SendEvent(events.DebridStreamState, StreamState{
					Status:      StreamStatusFailed,
					TorrentName: selectedTorrent.Name,
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
				TorrentName: selectedTorrent.Name,
				Message:     "External player link sent",
			})
		}

		go func() {
			defer util.HandlePanicInModuleThen("debridstream/AddBatchHistory", func() {})

			_ = db_bridge.InsertTorrentstreamHistory(s.repository.db, media.GetID(), selectedTorrent)
		}()
	}(ctx)

	s.repository.wsEventManager.SendEvent(events.DebridStreamState, StreamState{
		Status:      StreamStatusStarted,
		TorrentName: selectedTorrent.Name,
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
