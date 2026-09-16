package debrid_client

import (
	"context"
	"errors"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/debrid/debrid"
	"seanime/internal/directstream"
	"seanime/internal/events"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/mediacore"
	"seanime/internal/player"
	"seanime/internal/testmocks"
	"seanime/internal/util"
	"testing"
	"testing/synctest"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"github.com/stretchr/testify/require"
)

type failedStreamProvider struct {
	fakeDebridProvider
	progressEntered <-chan struct{}
	failure         error
}

func (p *failedStreamProvider) GetTorrentStreamUrl(_ context.Context, _ debrid.StreamTorrentOptions, items chan debrid.TorrentItem) (string, error) {
	items <- debrid.TorrentItem{Name: "Batch", CompletionPercentage: 100}
	<-p.progressEntered
	return "", p.failure
}

type blockedStreamProgressEvents struct {
	*events.MockWSEventManager
	progressEntered chan struct{}
	releaseProgress chan struct{}
	finished        chan struct{}
}

func (w *blockedStreamProgressEvents) SendEvent(event string, payload interface{}) {
	if state, ok := payload.(StreamState); event == events.DebridStreamState && ok &&
		state.Message == "Downloading torrent: 100%" {
		close(w.progressEntered)
		<-w.releaseProgress
	}
	w.MockWSEventManager.SendEvent(event, payload)
	if event == events.HideIndefiniteLoader {
		close(w.finished)
	}
}

type streamAbortBackend struct {
	mediacore.Backend
	ws     *blockedStreamProgressEvents
	events chan player.Event
}

func (b *streamAbortBackend) OpenAndAwait(string, string) {}
func (b *streamAbortBackend) AbortOpen(_ string, reason string) {
	b.ws.SendEvent("test-stream-abort", reason)
}
func (b *streamAbortBackend) Events() <-chan player.Event { return b.events }
func (b *streamAbortBackend) Close() error {
	close(b.events)
	return nil
}

func TestStartStreamFinishesProgressBeforeFailure(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		logger := zerolog.Nop()
		ws := &blockedStreamProgressEvents{
			MockWSEventManager: events.NewMockWSEventManager(&logger),
			progressEntered:    make(chan struct{}),
			releaseProgress:    make(chan struct{}),
			finished:           make(chan struct{}),
		}
		released := false
		defer func() {
			if !released {
				close(ws.releaseProgress)
			}
		}()
		backend := &streamAbortBackend{ws: ws, events: make(chan player.Event)}
		coordinator := mediacore.NewCoordinator(mediacore.NewCoordinatorOptions{
			Logger: &logger,
			Backends: map[player.Target]mediacore.Backend{
				player.TargetVideoCore: backend,
			},
		})
		defer coordinator.Close()
		defer coordinator.Unsubscribe("directstream")
		direct := directstream.NewManager(directstream.NewManagerOptions{
			Logger:               &logger,
			WSEventManager:       ws,
			MediacoreCoordinator: coordinator,
		})
		failure := errors.New("selected file not found")
		cache := anilist.NewCompleteAnimeCache()
		cache.SetT(21, &anilist.CompleteAnime{ID: 21}, time.Second)
		defer time.Sleep(time.Second)
		repo := &Repository{
			logger: &logger,
			provider: mo.Some[debrid.Provider](&failedStreamProvider{
				progressEntered: ws.progressEntered,
				failure:         failure,
			}),
			wsEventManager:      ws,
			directStreamManager: direct,
			completeAnimeCache:  cache,
			metadataProviderRef: util.NewRef[metadata_provider.Provider](testmocks.NewFakeMetadataProviderBuilder().
				WithAnimeMetadata(21, &metadata.AnimeMetadata{}).Build()),
		}
		manager := NewStreamManager(repo)
		err := manager.startStream(context.Background(), &StartStreamOptions{
			MediaId:       21,
			EpisodeNumber: 1,
			AniDBEpisode:  "1",
			Torrent:       &hibiketorrent.AnimeTorrent{Name: "Batch"},
			FileId:        "Episode.mkv",
			ClientId:      "test-client",
			PlaybackType:  PlaybackTypeNativePlayer,
		})
		require.NoError(t, err)
		<-ws.progressEntered
		synctest.Wait()

		for _, event := range ws.Events() {
			require.NotEqual(t, "test-stream-abort", event.Type, "player aborted before progress finished")
			if state, ok := event.Payload.(StreamState); ok {
				require.NotEqual(t, StreamStatusFailed, state.Status, "failure published before progress finished")
			}
		}

		released = true
		close(ws.releaseProgress)
		<-ws.finished
		synctest.Wait()

		var order []string
		for _, event := range ws.Events() {
			if event.Type == "test-stream-abort" {
				require.Equal(t, failure.Error(), event.Payload)
				order = append(order, "abort")
			}
			if state, ok := event.Payload.(StreamState); ok {
				if state.Message == "Downloading torrent: 100%" {
					order = append(order, "progress")
				}
				if state.Status == StreamStatusFailed {
					require.Contains(t, state.Message, failure.Error())
					order = append(order, "failed")
				}
			}
		}
		require.Equal(t, []string{"progress", "abort", "failed"}, order)
	})
}
