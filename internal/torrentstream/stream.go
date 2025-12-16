package torrentstream

import (
	"context"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/directstream"
	"seanime/internal/events"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/hook"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/util"
	"seanime/internal/videocore"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/samber/mo"
)

type PlaybackType string

const (
	PlaybackTypeExternal           PlaybackType = "default" // External player
	PlaybackTypeExternalPlayerLink PlaybackType = "externalPlayerLink"
	PlaybackTypeNativePlayer       PlaybackType = "nativeplayer"
	PlaybackTypeNone               PlaybackType = "none"
	PlaybackTypeNoneAndAwait       PlaybackType = "noneAndAwait"
)

type StartStreamOptions struct {
	MediaId           int
	EpisodeNumber     int                         // RELATIVE Episode number to identify the file
	AniDBEpisode      string                      // Animap episode
	AutoSelect        bool                        // Automatically select the best file to stream
	Torrent           *hibiketorrent.AnimeTorrent // Selected torrent (Manual selection)
	FileIndex         *int                        // Index of the file to stream (Manual selection)
	UserAgent         string
	ClientId          string
	PlaybackType      PlaybackType
	BatchEpisodeFiles *hibiketorrent.BatchEpisodeFiles
}

// StartStream is called by the client to start streaming a torrent
func (r *Repository) StartStream(ctx context.Context, opts *StartStreamOptions) (err error) {
	defer util.HandlePanicInModuleWithError("torrentstream/stream/StartStream", &err)
	// DEVNOTE: Do not
	//r.Shutdown()

	r.previousStreamOptions = mo.Some(opts)

	r.logger.Info().
		Str("clientId", opts.ClientId).
		Any("playbackType", opts.PlaybackType).
		Int("mediaId", opts.MediaId).Msgf("torrentstream: Starting stream for episode %s", opts.AniDBEpisode)

	r.sendStateEvent(eventLoading)
	r.wsEventManager.SendEvent(events.ShowIndefiniteLoader, "torrentstream")
	defer func() {
		r.wsEventManager.SendEvent(events.HideIndefiniteLoader, "torrentstream")
	}()

	if opts.PlaybackType == PlaybackTypeNativePlayer {
		r.directStreamManager.PrepareNewStream(opts.ClientId, "Selecting torrent...")
	}

	//
	// Get the media info
	//
	media, _, err := r.GetMediaInfo(ctx, opts.MediaId)
	if err != nil {
		return err
	}

	episodeNumber := opts.EpisodeNumber
	aniDbEpisode := opts.AniDBEpisode
	//
	// Check if there's a prepared stream that matches this request
	//
	var torrentToStream *playbackTorrent
	usedPreparedStream := false

	if prepared, ok := r.preloadedStream.Get(); ok {
		if streamOptionsMatch(opts, prepared.Options) {
			r.logger.Info().Msg("torrentstream: Using pre-downloaded stream")
			torrentToStream = &playbackTorrent{
				Torrent: prepared.Torrent,
				File:    prepared.File,
			}
			usedPreparedStream = true

			// Cancel the prepared stream context and clear it
			if prepared.CancelFunc != nil {
				prepared.CancelFunc()
			}
			r.preloadedStream = mo.None[*preloadedStream]()
		} else {
			// Different episode requested, cancel and drop the prepared stream
			r.logger.Debug().Msg("torrentstream: Prepared stream doesn't match request, cancelling it")
			r.CancelPreparedStream()
		}
	}

	//
	// Find the best torrent / Select the torrent (only if not using prepared stream)
	//
	if !usedPreparedStream {
		if opts.AutoSelect {
			torrentToStream, err = r.findBestTorrent(media, aniDbEpisode, episodeNumber)
			if err != nil {
				if opts.PlaybackType == PlaybackTypeNativePlayer {
					r.directStreamManager.AbortOpen(opts.ClientId, err)
				}
				r.sendStateEvent(eventLoadingFailed)
				return err
			}
		} else {
			if opts.Torrent == nil {
				return fmt.Errorf("torrentstream: No torrent provided")
			}
			torrentToStream, err = r.findBestTorrentFromManualSelection(opts.Torrent, media, aniDbEpisode, opts.FileIndex)
			if err != nil {
				if opts.PlaybackType == PlaybackTypeNativePlayer {
					r.directStreamManager.AbortOpen(opts.ClientId, err)
				}
				r.sendStateEvent(eventLoadingFailed)
				return err
			}
		}
	}

	if torrentToStream == nil {
		if opts.PlaybackType == PlaybackTypeNativePlayer {
			r.directStreamManager.AbortOpen(opts.ClientId, fmt.Errorf("torrentstream: No torrent found"))
		}
		r.sendStateEvent(eventLoadingFailed)
		return fmt.Errorf("torrentstream: No torrent selected")
	}

	//
	// Set current file & torrent
	//
	r.client.currentFile = mo.Some(torrentToStream.File)
	r.client.currentTorrent = mo.Some(torrentToStream.Torrent)

	r.sendStateEvent(eventLoading, TLSStateSendingStreamToMediaPlayer)

	go func() {
		// Add the torrent to the history if it is a batch & manually selected
		if len(r.client.currentTorrent.MustGet().Files()) > 1 && opts.Torrent != nil && opts.Torrent.IsBatch {
			r.AddBatchHistory(opts.MediaId, opts.Torrent, opts.BatchEpisodeFiles) // ran in goroutine
		}
	}()

	//
	// Start the playback
	//
	go func() {
		switch opts.PlaybackType {
		case PlaybackTypeNone:
			r.logger.Warn().Msg("torrentstream: Playback type is set to 'none'")
			// Signal to the client that the torrent has started playing (remove loading status)
			// There will be no tracking
			r.sendStateEvent(eventTorrentStartedPlaying)
		case PlaybackTypeNoneAndAwait:
			r.logger.Warn().Msg("torrentstream: Playback type is set to 'noneAndAwait'")
			// Signal to the client that the torrent has started playing (remove loading status)
			// There will be no tracking
			for {
				if r.client.readyToStream() {
					break
				}
				time.Sleep(3 * time.Second) // Wait for 3 secs before checking again
			}
			r.sendStateEvent(eventTorrentStartedPlaying)
		//
		// External player
		//
		case PlaybackTypeExternal, PlaybackTypeExternalPlayerLink:
			r.sendStreamToExternalPlayer(opts, media, aniDbEpisode)
		//
		// Direct stream
		//
		case PlaybackTypeNativePlayer:
			readyCh, err := r.directStreamManager.PlayTorrentStream(ctx, directstream.PlayTorrentStreamOptions{
				ClientId:      opts.ClientId,
				EpisodeNumber: opts.EpisodeNumber,
				AnidbEpisode:  opts.AniDBEpisode,
				Media:         media.ToBaseAnime(),
				Torrent:       r.client.currentTorrent.MustGet(),
				File:          r.client.currentFile.MustGet(),
				OnTerminate: func() {
					_ = r.StopStream(true)
				},
			})
			if err != nil {
				r.logger.Error().Err(err).Msg("torrentstream: Failed to prepare new stream")
				r.sendStateEvent(eventLoadingFailed)
				return
			}

			if opts.PlaybackType == PlaybackTypeNativePlayer {
				r.directStreamManager.PrepareNewStream(opts.ClientId, "Downloading metadata...")
			}

			// Make sure the client is ready and the torrent is partially downloaded
			for {
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
			close(readyCh)
		}
	}()

	r.sendStateEvent(eventTorrentLoaded)
	r.logger.Info().Msg("torrentstream: Stream started")

	return nil
}

// sendStreamToExternalPlayer sends the stream to the desktop player or external player link.
// It blocks until the some pieces have been downloaded before sending the stream for faster playback.
func (r *Repository) sendStreamToExternalPlayer(opts *StartStreamOptions, completeAnime *anilist.CompleteAnime, aniDbEpisode string) {

	baseAnime := completeAnime.ToBaseAnime()

	r.wsEventManager.SendEvent(events.ShowIndefiniteLoader, "torrentstream")
	defer func() {
		r.wsEventManager.SendEvent(events.HideIndefiniteLoader, "torrentstream")
	}()

	// Make sure the client is ready and the torrent is partially downloaded
	for {
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

	event := &TorrentStreamSendStreamToMediaPlayerEvent{
		WindowTitle:  "",
		StreamURL:    r.client.GetStreamingUrl(),
		Media:        baseAnime,
		AniDbEpisode: aniDbEpisode,
		PlaybackType: string(opts.PlaybackType),
	}
	err := hook.GlobalHookManager.OnTorrentStreamSendStreamToMediaPlayer().Trigger(event)
	if err != nil {
		r.logger.Error().Err(err).Msg("torrentstream: Failed to trigger hook")
		return
	}
	windowTitle := event.WindowTitle
	streamURL := event.StreamURL
	baseAnime = event.Media
	aniDbEpisode = event.AniDbEpisode
	playbackType := PlaybackType(event.PlaybackType)

	if event.DefaultPrevented {
		r.logger.Debug().Msg("torrentstream: Stream prevented by hook")
		return
	}

	switch playbackType {
	//
	// Desktop player
	//
	case PlaybackTypeExternal:
		r.logger.Debug().Msgf("torrentstream: Starting the media player %s", streamURL)
		err = r.playbackManager.StartStreamingUsingMediaPlayer(windowTitle, &playbackmanager.StartPlayingOptions{
			Payload:   streamURL,
			UserAgent: opts.UserAgent,
			ClientId:  opts.ClientId,
		}, baseAnime, aniDbEpisode)
		if err != nil {
			// Failed to start the stream, we'll drop the torrents and stop the server
			r.sendStateEvent(eventLoadingFailed)
			_ = r.StopStream()
			r.logger.Error().Err(err).Msg("torrentstream: Failed to start the stream")
			r.wsEventManager.SendEventTo(opts.ClientId, events.ErrorToast, err.Error())
		}

		r.wsEventManager.SendEvent(events.ShowIndefiniteLoader, "torrentstream")
		defer func() {
			r.wsEventManager.SendEvent(events.HideIndefiniteLoader, "torrentstream")
		}()

		r.playbackManager.RegisterMediaPlayerCallback(func(event playbackmanager.PlaybackEvent) bool {
			switch event.(type) {
			case playbackmanager.StreamStartedEvent:
				r.logger.Debug().Msg("torrentstream: Media player started playing")
				r.wsEventManager.SendEvent(events.HideIndefiniteLoader, "torrentstream")
				return false
			}
			return true
		})

	//
	// External player link
	//
	case PlaybackTypeExternalPlayerLink:
		r.logger.Debug().Msgf("torrentstream: Sending stream to external player %s", streamURL)
		r.wsEventManager.SendEventTo(opts.ClientId, events.ExternalPlayerOpenURL, struct {
			Url           string `json:"url"`
			MediaId       int    `json:"mediaId"`
			EpisodeNumber int    `json:"episodeNumber"`
			MediaTitle    string `json:"mediaTitle"`
		}{
			Url:           r.client.GetExternalPlayerStreamingUrl(),
			MediaId:       opts.MediaId,
			EpisodeNumber: opts.EpisodeNumber,
			MediaTitle:    baseAnime.GetPreferredTitle(),
		})

		// Signal to the client that the torrent has started playing (remove loading status)
		// We can't know for sure
		r.sendStateEvent(eventTorrentStartedPlaying)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type StartUntrackedStreamOptions struct {
	Magnet       string
	FileIndex    int
	WindowTitle  string
	UserAgent    string
	ClientId     string
	PlaybackType PlaybackType
}

// StopStream stops the stream and closes the server.
// If fromNativePlayer is true, it will not stop the native player again.
func (r *Repository) StopStream(fromNativePlayer ...bool) error {
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
		currentTorrent := r.client.currentTorrent.MustGet()
		shouldDrop := r.client.currentTorrentStatus.ProgressPercentage < 70

		// Don't drop if this is the prepared torrent
		if r.preloadedStream.IsPresent() {
			prepared := r.preloadedStream.MustGet()
			if currentTorrent.InfoHash() == prepared.Torrent.InfoHash() {
				r.client.repository.logger.Debug().Msg("torrentstream: Not dropping torrent as it's being prepared for next episode")
				shouldDrop = false
			}
		}

		if shouldDrop {
			r.client.repository.logger.Debug().Msg("torrentstream: Dropping torrent, completion is less than 70%")
			r.client.dropTorrents()
		}
		r.client.repository.logger.Debug().Msg("torrentstream: Resetting current torrent and status")
	}
	r.client.currentTorrent = mo.None[*torrent.Torrent]()        // Reset the current torrent
	r.client.currentFile = mo.None[*torrent.File]()              // Reset the current file
	r.client.currentTorrentStatus = TorrentStatus{}              // Reset the torrent status
	r.client.repository.sendStateEvent(eventTorrentStopped, nil) // Send torrent stopped event
	r.client.repository.mediaPlayerRepository.Stop()             // Stop the media player gracefully if it's running
	r.client.mu.Unlock()

	if len(fromNativePlayer) == 0 || fromNativePlayer[0] == false {
		go func() {
			if playbackType, ok := r.nativePlayer.VideoCore().GetCurrentPlaybackType(); ok && playbackType == videocore.PlaybackTypeTorrent {
				r.nativePlayer.Stop()
			}
		}()
	}

	r.logger.Info().Msg("torrentstream: Stream stopped")

	return nil
}

func (r *Repository) DropTorrent() error {
	r.logger.Info().Msg("torrentstream: Dropping last torrent")

	if r.client.torrentClient.IsAbsent() {
		return nil
	}

	for _, t := range r.client.torrentClient.MustGet().Torrents() {
		t.Drop()
	}

	r.mediaPlayerRepository.Stop()

	r.logger.Info().Msg("torrentstream: Dropped last torrent")

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) GetMediaInfo(ctx context.Context, mediaId int) (media *anilist.CompleteAnime, animeMetadata *metadata.AnimeMetadata, err error) {
	// Get the media
	var found bool
	media, found = r.completeAnimeCache.Get(mediaId)
	if !found {
		// Fetch the media
		media, err = r.platformRef.Get().GetAnimeWithRelations(ctx, mediaId)
		if err != nil {
			baseAnime, lErr := r.platformRef.Get().GetAnime(ctx, mediaId)
			if lErr != nil {
				return nil, nil, fmt.Errorf("torrentstream: Failed to fetch media: %w", err)
			}
			media = baseAnime.ToCompleteAnime()
			err = nil
		}
	}

	// Get the media
	animeMetadata, err = r.metadataProviderRef.Get().GetAnimeMetadata(metadata.AnilistPlatform, mediaId)
	if err != nil {
		//return nil, nil, fmt.Errorf("torrentstream: Could not fetch AniDB media: %w", err)
		animeMetadata = &metadata.AnimeMetadata{
			Titles:       make(map[string]string),
			Episodes:     make(map[string]*metadata.EpisodeMetadata),
			EpisodeCount: 0,
			SpecialCount: 0,
			Mappings: &metadata.AnimeMappings{
				AnilistId: media.GetID(),
			},
		}
		animeMetadata.Titles["en"] = media.GetTitleSafe()
		animeMetadata.Titles["x-jat"] = media.GetRomajiTitleSafe()
		err = nil
	}

	return
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// PreloadStream
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// streamOptionsMatch checks if two stream options represent the same episode
func streamOptionsMatch(a, b *StartStreamOptions) bool {
	if a == nil || b == nil {
		return false
	}
	return a.MediaId == b.MediaId && a.EpisodeNumber == b.EpisodeNumber
}

// PreloadStream starts pre-downloading a stream at reduced speed to avoid interfering with current playback
func (r *Repository) PreloadStream(ctx context.Context, opts *StartStreamOptions) (err error) {
	defer util.HandlePanicInModuleWithError("torrentstream/stream/PreloadStream", &err)

	r.logger.Info().
		Int("mediaId", opts.MediaId).
		Int("episodeNumber", opts.EpisodeNumber).
		Msg("torrentstream: Preloading stream for future playback")

	// Cancel any existing prepared stream
	if r.preloadedStream.IsPresent() {
		r.logger.Debug().Msg("torrentstream: Cancelling existing preloaded stream")
		prepared := r.preloadedStream.MustGet()
		if prepared.CancelFunc != nil {
			prepared.CancelFunc()
		}
		r.preloadedStream = mo.None[*preloadedStream]()
	}

	// Get media info
	media, _, err := r.GetMediaInfo(ctx, opts.MediaId)
	if err != nil {
		return err
	}

	// Find best torrent
	var torrentToStream *playbackTorrent
	if opts.AutoSelect {
		torrentToStream, err = r.findBestTorrent(media, opts.AniDBEpisode, opts.EpisodeNumber)
		if err != nil {
			r.logger.Error().Err(err).Msg("torrentstream: Failed to find torrent for preloading")
			return err
		}
	} else {
		if opts.Torrent == nil {
			return fmt.Errorf("torrentstream: No torrent provided")
		}
		torrentToStream, err = r.findBestTorrentFromManualSelection(opts.Torrent, media, opts.AniDBEpisode, opts.FileIndex)
		if err != nil {
			r.logger.Error().Err(err).Msg("torrentstream: Failed to select torrent for preloading")
			return err
		}
	}

	if torrentToStream == nil {
		return fmt.Errorf("torrentstream: No torrent selected for preloading")
	}

	// Create a cancellable context for this prepared stream
	prepareCtx, cancelFunc := context.WithCancel(ctx)

	r.logger.Info().
		Str("torrent", torrentToStream.Torrent.Name()).
		Msg("torrentstream: Started preloading stream")

	// Store prepared stream info
	r.preloadedStream = mo.Some(&preloadedStream{
		Torrent:    torrentToStream.Torrent,
		File:       torrentToStream.File,
		Options:    opts,
		CancelFunc: cancelFunc,
	})

	// Start downloading in background
	go func() {
		<-prepareCtx.Done()
		r.logger.Debug().Msg("torrentstream: Prepared stream context cancelled")
	}()

	return nil
}

// CancelPreparedStream cancels any ongoing stream preloading
func (r *Repository) CancelPreparedStream() {
	if prepared, ok := r.preloadedStream.Get(); ok {
		r.logger.Debug().Msg("torrentstream: Cancelling prepared stream")
		if prepared.CancelFunc != nil {
			prepared.CancelFunc()
		}
		// Drop the prepared torrent if it's not the current one
		if r.client.currentTorrent.IsAbsent() ||
			r.client.currentTorrent.MustGet().InfoHash() != prepared.Torrent.InfoHash() {
			prepared.Torrent.Drop()
		}
		r.preloadedStream = mo.None[*preloadedStream]()
	}
}
