package torrentstream

import (
	"context"
	"errors"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/directstream"
	"seanime/internal/events"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/hook"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/player"
	"seanime/internal/util"
	"sync"
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
	MediaId           int                              `json:"mediaId"`
	EpisodeNumber     int                              `json:"episodeNumber"` // RELATIVE Episode number to identify the file
	AniDBEpisode      string                           `json:"aniDbEpisode"`  // Animap episode
	AutoSelect        bool                             `json:"autoSelect"`    // Automatically select the best file to stream
	Torrent           *hibiketorrent.AnimeTorrent      `json:"torrent"`       // Selected torrent (Manual selection)
	FileIndex         *int                             `json:"fileIndex"`     // Index of the file to stream (Manual selection)
	UserAgent         string                           `json:"userAgent"`
	ClientId          string                           `json:"clientId"`
	PlaybackType      PlaybackType                     `json:"playbackType"`
	BatchEpisodeFiles *hibiketorrent.BatchEpisodeFiles `json:"batchEpisodeFiles"`
	media             *anilist.BaseAnime               `json:"-"`
}

func (opts *StartStreamOptions) SetMedia(media *anilist.BaseAnime) {
	opts.media = media
}

func (r *Repository) incStartRequestId() uint64 {
	return r.startRequestId.Add(1)
}

func (r *Repository) isLatestStartRequest(requestId uint64) bool {
	return r.startRequestId.Load() == requestId
}

func (r *Repository) beginStartRequest(ctx context.Context, requestId uint64) (context.Context, func()) {
	if ctx == nil {
		ctx = context.Background()
	}

	startCtx, cancel := context.WithCancel(ctx)
	r.startCancelMu.Lock()
	if r.startCancel != nil {
		r.startCancel()
	}
	r.startCancel = cancel
	r.startCancelId = requestId
	r.startCancelMu.Unlock()

	return startCtx, func() {
		r.startCancelMu.Lock()
		if r.startCancelId == requestId {
			r.startCancel = nil
			r.startCancelId = 0
		}
		r.startCancelMu.Unlock()
		cancel()
	}
}

func (r *Repository) cancelStartRequest() {
	r.startRequestId.Add(1)

	r.startCancelMu.Lock()
	if r.startCancel != nil {
		r.startCancel()
	}
	r.startCancel = nil
	r.startCancelId = 0
	r.startCancelMu.Unlock()
}

func (r *Repository) isStaleStartError(err error, requestID uint64) bool {
	return errors.Is(err, context.Canceled) && !r.isLatestStartRequest(requestID)
}

func (r *Repository) dropStalePlaybackTorrent(stream *playbackTorrent) {
	if stream == nil || stream.Torrent == nil {
		return
	}

	infoHash := stream.Torrent.InfoHash()
	if r.client.currentTorrent.IsPresent() && r.client.currentTorrent.MustGet().InfoHash() == infoHash {
		return
	}
	if prepared, ok := r.preloadedStream.Get(); ok && prepared.Torrent.InfoHash() == infoHash {
		return
	}

	stream.Torrent.Drop()
	r.client.removeTorrentFiles(infoHash)
}

// StartStream is called by the client to start streaming a torrent
func (r *Repository) StartStream(ctx context.Context, opts *StartStreamOptions) (err error) {
	defer util.HandlePanicInModuleWithError("torrentstream/stream/StartStream", &err)
	startLaunchTime := time.Now()
	requestId := r.incStartRequestId()
	ctx, finishStart := r.beginStartRequest(ctx, requestId)
	defer finishStart()

	r.playback.currentVideoDuration = 0
	// DEVNOTE: Do not
	//r.Shutdown()

	r.logger.Info().
		Str("clientId", opts.ClientId).
		Any("playbackType", opts.PlaybackType).
		Int("mediaId", opts.MediaId).Msgf("torrentstream: Starting stream for episode %s", opts.AniDBEpisode)

	r.sendStateEvent(eventLoading)
	r.wsEventManager.SendEvent(events.ShowIndefiniteLoader, "torrentstream")
	defer func() {
		r.wsEventManager.SendEvent(events.HideIndefiniteLoader, "torrentstream")
	}()

	var readyCh chan struct{}
	var readyChOnce sync.Once
	closeReadyCh := func() {
		if readyCh == nil {
			return
		}
		readyChOnce.Do(func() {
			close(readyCh)
		})
	}

	if opts.PlaybackType == PlaybackTypeNativePlayer {
		r.directStreamManager.BeginOpen(opts.ClientId, "Selecting torrent...", func() {
			closeReadyCh()
			_ = r.StopStream(true)
		})
	}

	r.streamActionMu.Lock()
	defer r.streamActionMu.Unlock()

	if !r.isLatestStartRequest(requestId) {
		r.logger.Debug().Msg("torrentstream: Ignoring stale stream request")
		return nil
	}
	if opts.PlaybackType == PlaybackTypeNativePlayer && !r.directStreamManager.IsOpenActive(opts.ClientId) {
		r.logger.Debug().Msg("torrentstream: Stream opening was cancelled before selection")
		return nil
	}
	r.previousStreamOptions = mo.Some(opts)

	//
	// Get the media info
	//
	media, _, err := r.GetMediaInfoFromOptions(ctx, opts)
	if err != nil {
		if r.isStaleStartError(err, requestId) {
			return nil
		}
		return err
	}
	if !r.isLatestStartRequest(requestId) {
		r.logger.Debug().Msg("torrentstream: Ignoring stale stream request after media lookup")
		return nil
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
			r.logger.Debug().Msgf(
				"torrentstream: Prepared stream doesn't match request, cancelling it requestMediaId=%d requestEpisode=%d requestAniDBEpisode=%s requestFileIndex=%d preparedMediaId=%d preparedEpisode=%d preparedAniDBEpisode=%s preparedFileIndex=%d",
				opts.MediaId,
				opts.EpisodeNumber,
				opts.AniDBEpisode,
				streamOptionFileIndex(opts),
				prepared.Options.MediaId,
				prepared.Options.EpisodeNumber,
				prepared.Options.AniDBEpisode,
				streamOptionFileIndex(prepared.Options),
			)
			r.cancelPreparedStream()
		}
	}

	//
	// Find the best torrent / Select the torrent (only if not using prepared stream)
	//
	if !usedPreparedStream {
		if opts.AutoSelect {
			torrentToStream, err = r.findBestTorrent(ctx, media, aniDbEpisode, episodeNumber)
			if err != nil {
				if r.isStaleStartError(err, requestId) {
					return nil
				}
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
			torrentToStream, err = r.findBestTorrentFromManualSelection(ctx, opts.Torrent, media, aniDbEpisode, opts.FileIndex)
			if err != nil {
				if r.isStaleStartError(err, requestId) {
					return nil
				}
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

	torrentSelectionDuration := time.Since(startLaunchTime)
	var metadataRetrievalDuration time.Duration
	if usedPreparedStream {
		metadataRetrievalDuration = 0
	} else {
		metadataRetrievalDuration = r.client.lastMetadataDuration
	}

	if !r.isLatestStartRequest(requestId) {
		r.logger.Debug().Msg("torrentstream: Dropping stale stream selection")
		r.dropStalePlaybackTorrent(torrentToStream)
		return nil
	}

	if opts.PlaybackType == PlaybackTypeNativePlayer && !r.directStreamManager.IsOpenActive(opts.ClientId) {
		r.logger.Debug().Msg("torrentstream: Stream opening was cancelled before playback")
		if torrentToStream.Torrent != nil {
			r.dropStalePlaybackTorrent(torrentToStream)
		}
		return nil
	}

	//
	// Set current file & torrent
	//
	r.client.currentFile = mo.Some(torrentToStream.File)
	r.client.currentTorrent = mo.Some(torrentToStream.Torrent)
	r.client.ResetBaselines()
	r.resetPreloadFlag()
	r.client.cleanupActiveTorrentFiles()

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
			firstUsefulBytesTime, waitErr := r.waitForReadyToStream(context.Background(), "", false, startLaunchTime)
			if waitErr != nil {
				r.logger.Error().Err(waitErr).Msg("torrentstream: wait for ready failed")
				return
			}
			r.logDiagnostics(startLaunchTime, torrentSelectionDuration, metadataRetrievalDuration, firstUsefulBytesTime)
			r.sendStateEvent(eventTorrentStartedPlaying)
		//
		// External player
		//
		case PlaybackTypeExternal, PlaybackTypeExternalPlayerLink:
			r.sendStreamToExternalPlayer(context.Background(), opts, media, aniDbEpisode, startLaunchTime, torrentSelectionDuration, metadataRetrievalDuration)
		//
		// Direct stream
		//
		case PlaybackTypeNativePlayer:
			readyCh, err = r.directStreamManager.PlayTorrentStream(ctx, directstream.PlayTorrentStreamOptions{
				ClientId:      opts.ClientId,
				EpisodeNumber: opts.EpisodeNumber,
				AnidbEpisode:  opts.AniDBEpisode,
				Media:         media.ToBaseAnime(),
				Torrent:       r.client.currentTorrent.MustGet(),
				File:          r.client.currentFile.MustGet(),
				DownloadDir:   r.GetDownloadDir(),
				OnTerminate: func() {
					_ = r.StopStream(true)
				},
			})
			if err != nil {
				r.logger.Error().Err(err).Msg("torrentstream: Failed to prepare new stream")
				r.sendStateEvent(eventLoadingFailed)
				return
			}
			if !r.directStreamManager.IsOpenActive(opts.ClientId) {
				closeReadyCh()
				return
			}

			if opts.PlaybackType == PlaybackTypeNativePlayer {
				if !r.directStreamManager.UpdateOpenStep(opts.ClientId, "Downloading metadata...") {
					closeReadyCh()
					return
				}
			}

			// Make sure the client is ready and the torrent is partially downloaded
			firstUsefulBytesTime, waitErr := r.waitForReadyToStream(context.Background(), opts.ClientId, true, startLaunchTime)
			if waitErr != nil {
				r.logger.Error().Err(waitErr).Msg("torrentstream: wait for ready failed")
				closeReadyCh()
				return
			}
			r.logDiagnostics(startLaunchTime, torrentSelectionDuration, metadataRetrievalDuration, firstUsefulBytesTime)
			readyChOnce.Do(func() {
				close(readyCh)
			})
		}
	}()

	if opts.PlaybackType == PlaybackTypeNativePlayer && !r.directStreamManager.IsOpenActive(opts.ClientId) {
		r.logger.Debug().Msg("torrentstream: Stream opening was cancelled before loaded event")
		return nil
	}

	r.sendStateEvent(eventTorrentLoaded)
	r.logger.Info().Msg("torrentstream: Stream started")

	return nil
}

// sendStreamToExternalPlayer sends the stream to the desktop player or external player link.
// It blocks until the some pieces have been downloaded before sending the stream for faster playback.
func (r *Repository) sendStreamToExternalPlayer(
	ctx context.Context,
	opts *StartStreamOptions,
	completeAnime *anilist.CompleteAnime,
	aniDbEpisode string,
	startLaunchTime time.Time,
	torrentSelectionDuration time.Duration,
	metadataRetrievalDuration time.Duration,
) {

	baseAnime := completeAnime.ToBaseAnime()

	r.wsEventManager.SendEvent(events.ShowIndefiniteLoader, "torrentstream")
	defer func() {
		r.wsEventManager.SendEvent(events.HideIndefiniteLoader, "torrentstream")
	}()

	// Make sure the client is ready and the torrent is partially downloaded
	firstUsefulBytesTime, waitErr := r.waitForReadyToStream(ctx, "", false, startLaunchTime)
	if waitErr != nil {
		r.logger.Error().Err(waitErr).Msg("torrentstream: wait for ready failed")
		return
	}
	r.logDiagnostics(startLaunchTime, torrentSelectionDuration, metadataRetrievalDuration, firstUsefulBytesTime)

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
	fromNative := len(fromNativePlayer) > 0 && fromNativePlayer[0]
	if !fromNative {
		r.cancelStartRequest()
	}

	r.logger.Info().Msg("torrentstream: Stopping stream")
	if r.directStreamManager != nil {
		r.directStreamManager.CloseOpen("")
	}

	r.playback.currentVideoDuration = 0

	r.streamActionMu.Lock()
	defer r.streamActionMu.Unlock()

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
			infoHash := currentTorrent.InfoHash()
			currentTorrent.Drop()
			r.client.removeTorrentFiles(infoHash)
		}
		r.client.repository.logger.Debug().Msg("torrentstream: Resetting current torrent and status")
	}
	r.client.currentTorrent = mo.None[*torrent.Torrent]()        // Reset the current torrent
	r.client.currentFile = mo.None[*torrent.File]()              // Reset the current file
	r.client.currentTorrentStatus = TorrentStatus{}              // Reset the torrent status
	r.client.repository.sendStateEvent(eventTorrentStopped, nil) // Send torrent stopped event
	r.client.repository.mediaPlayerRepository.Stop()             // Stop the media player gracefully if it's running
	r.client.mu.Unlock()

	if !fromNative {
		go func() {
			if session, ok := r.mediacoreCoordinator.GetActiveSession(); ok {
				playbackState, okState := r.mediacoreCoordinator.GetActivePlaybackState()
				if okState && playbackState.PlaybackInfo != nil && playbackState.PlaybackInfo.PlaybackType == player.PlaybackTypeTorrent {
					r.mediacoreCoordinator.Terminate(session)
				}
			}
		}()
	}

	r.logger.Info().Msg("torrentstream: Stream stopped")

	return nil
}

func (r *Repository) DropTorrent() error {
	r.streamActionMu.Lock()
	defer r.streamActionMu.Unlock()

	r.logger.Info().Msg("torrentstream: Dropping last torrent")

	if r.client.torrentClient.IsAbsent() {
		return nil
	}

	for _, t := range r.client.torrentClient.MustGet().Torrents() {
		infoHash := t.InfoHash()
		t.Drop()
		r.client.removeTorrentFiles(infoHash)
	}

	r.mediaPlayerRepository.Stop()

	r.logger.Info().Msg("torrentstream: Dropped last torrent")

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) GetMediaInfoFromOptions(ctx context.Context, opts *StartStreamOptions) (media *anilist.CompleteAnime, animeMetadata *metadata.AnimeMetadata, err error) {
	if opts != nil && opts.media != nil {
		return r.getMediaInfo(ctx, opts.media.GetID(), opts.media.ToCompleteAnime())
	}
	return r.GetMediaInfo(ctx, opts.MediaId)
}

func (r *Repository) GetMediaInfo(ctx context.Context, mediaId int) (media *anilist.CompleteAnime, animeMetadata *metadata.AnimeMetadata, err error) {
	return r.getMediaInfo(ctx, mediaId, nil)
}

func (r *Repository) getMediaInfo(ctx context.Context, mediaId int, media *anilist.CompleteAnime) (ret *anilist.CompleteAnime, animeMetadata *metadata.AnimeMetadata, err error) {
	// Get the media
	if media != nil {
		ret = media
	} else if cached, found := r.completeAnimeCache.Get(mediaId); found {
		ret = cached
	} else {
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
		ret = media
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
				AnilistId: ret.GetID(),
			},
		}
		animeMetadata.Titles["en"] = ret.GetTitleSafe()
		animeMetadata.Titles["x-jat"] = ret.GetRomajiTitleSafe()
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
	if a.MediaId != b.MediaId || a.EpisodeNumber != b.EpisodeNumber || a.AniDBEpisode != b.AniDBEpisode {
		return false
	}
	if a.FileIndex == nil || b.FileIndex == nil {
		return true
	}
	return *a.FileIndex == *b.FileIndex
}

func streamOptionFileIndex(opts *StartStreamOptions) int {
	if opts == nil || opts.FileIndex == nil {
		return -1
	}
	return *opts.FileIndex
}

func preparedStreamOptions(opts *StartStreamOptions, stream *playbackTorrent) *StartStreamOptions {
	if opts == nil || stream == nil || stream.Torrent == nil || stream.File == nil {
		return opts
	}

	normalized := *opts
	if normalized.FileIndex == nil {
		for i, file := range stream.Torrent.Files() {
			if file == stream.File || file.Path() == stream.File.Path() {
				fileIndex := i
				normalized.FileIndex = &fileIndex
				break
			}
		}
	}

	return &normalized
}

// PreloadStream starts pre-downloading a stream at reduced speed to avoid interfering with current playback
func (r *Repository) PreloadStream(ctx context.Context, opts *StartStreamOptions) (err error) {
	defer util.HandlePanicInModuleWithError("torrentstream/stream/PreloadStream", &err)
	r.streamActionMu.Lock()
	defer r.streamActionMu.Unlock()

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
	media, _, err := r.GetMediaInfoFromOptions(ctx, opts)
	if err != nil {
		return err
	}

	// Find best torrent
	var torrentToStream *playbackTorrent
	if opts.AutoSelect {
		torrentToStream, err = r.findBestTorrent(ctx, media, opts.AniDBEpisode, opts.EpisodeNumber)
		if err != nil {
			r.logger.Error().Err(err).Msg("torrentstream: Failed to find torrent for preloading")
			return err
		}
	} else {
		if opts.Torrent == nil {
			return fmt.Errorf("torrentstream: No torrent provided")
		}
		torrentToStream, err = r.findBestTorrentFromManualSelection(ctx, opts.Torrent, media, opts.AniDBEpisode, opts.FileIndex)
		if err != nil {
			r.logger.Error().Err(err).Msg("torrentstream: Failed to select torrent for preloading")
			return err
		}
	}

	if torrentToStream == nil {
		return fmt.Errorf("torrentstream: No torrent selected for preloading")
	}

	// Create a cancellable context for this prepared stream
	prepareCtx, cancelFunc := context.WithCancel(context.Background())

	r.logger.Info().
		Str("aniDBEpisode", opts.AniDBEpisode).
		Int("fileIndex", streamOptionFileIndex(preparedStreamOptions(opts, torrentToStream))).
		Str("torrent", torrentToStream.Torrent.Name()).
		Msg("torrentstream: Started preloading stream")

	// Store prepared stream info
	r.preloadedStream = mo.Some(&preloadedStream{
		Torrent:    torrentToStream.Torrent,
		File:       torrentToStream.File,
		Options:    preparedStreamOptions(opts, torrentToStream),
		CancelFunc: cancelFunc,
	})
	r.client.cleanupActiveTorrentFiles()

	// Start downloading in background
	go func() {
		<-prepareCtx.Done()
		r.logger.Debug().Msg("torrentstream: Prepared stream context cancelled")
	}()

	return nil
}

// CancelPreparedStream cancels any ongoing stream preloading
func (r *Repository) CancelPreparedStream() {
	r.streamActionMu.Lock()
	defer r.streamActionMu.Unlock()

	r.cancelPreparedStream()
}

func (r *Repository) cancelPreparedStream() {
	if prepared, ok := r.preloadedStream.Get(); ok {
		r.logger.Debug().Msg("torrentstream: Cancelling prepared stream")
		if prepared.CancelFunc != nil {
			prepared.CancelFunc()
		}
		// Drop the prepared torrent if it's not the current one
		if r.client.currentTorrent.IsAbsent() ||
			r.client.currentTorrent.MustGet().InfoHash() != prepared.Torrent.InfoHash() {
			infoHash := prepared.Torrent.InfoHash()
			prepared.Torrent.Drop()
			r.client.removeTorrentFiles(infoHash)
		}
		r.preloadedStream = mo.None[*preloadedStream]()
	}
}

// waitForReadyToStream blocks until the client is ready to stream or the context is cancelled.
// It uses piece state changes and a 250 ms ticker to poll readiness.
func (r *Repository) waitForReadyToStream(ctx context.Context, clientId string, checkOpenActive bool, startTime time.Time) (firstUsefulBytesTime time.Duration, err error) {
	if r.client.torrentClient.IsAbsent() || r.client.currentTorrent.IsAbsent() {
		return 0, errors.New("torrent is absent")
	}

	t := r.client.currentTorrent.MustGet()
	sub := t.SubscribePieceStateChanges()
	defer sub.Close()

	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	var firstUsefulBytesMeasured bool

	for {
		if checkOpenActive && !r.directStreamManager.IsOpenActive(clientId) {
			return 0, errors.New("stream opening was cancelled")
		}

		if r.client.readyToStream() {
			return firstUsefulBytesTime, nil
		}

		if r.client.torrentClient.IsAbsent() || r.client.currentTorrent.IsAbsent() || r.client.currentTorrent.MustGet().InfoHash() != t.InfoHash() {
			return 0, errors.New("torrent was dropped or changed")
		}

		stats := t.Stats()
		if !firstUsefulBytesMeasured && stats.BytesReadUsefulData.Int64() > 0 {
			firstUsefulBytesTime = time.Since(startTime)
			firstUsefulBytesMeasured = true
		}

		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-sub.Values:
			// Recheck immediately on piece state changes
		case <-ticker.C:
			// Fallback check
		}
	}
}

func (r *Repository) logDiagnostics(startLaunchTime time.Time, torrentSelection time.Duration, metadataRetrieval time.Duration, firstUsefulBytes time.Duration) {
	if r.client.currentTorrent.IsAbsent() {
		return
	}
	t := r.client.currentTorrent.MustGet()
	totalLaunchTime := time.Since(startLaunchTime)
	peerCountAtReady := len(t.PeerConns())

	r.logger.Info().
		Str("torrentSelectionTime", torrentSelection.String()).
		Str("metadataRetrievalTime", metadataRetrieval.String()).
		Str("firstUsefulBytesTime", firstUsefulBytes.String()).
		Str("playableReadinessTime", totalLaunchTime.String()).
		Int("peerCountAtReady", peerCountAtReady).
		Str("totalLaunchTime", totalLaunchTime.String()).
		Msg("torrentstream: Startup diagnostics completed")
}
