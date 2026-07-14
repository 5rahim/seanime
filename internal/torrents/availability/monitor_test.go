package availability

import (
	"context"
	"errors"
	"seanime/internal/api/anilist"
	"seanime/internal/library/anime"
	"seanime/internal/testmocks"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testSearcher struct {
	mu        sync.Mutex
	available bool
	err       error
	calls     int
	started   chan struct{}
	release   chan struct{}
}

func (s *testSearcher) search(ctx context.Context, _ string, _ *anilist.BaseAnime, _ int) (bool, error) {
	s.mu.Lock()
	s.calls++
	available := s.available
	err := s.err
	started := s.started
	release := s.release
	s.mu.Unlock()

	if started != nil {
		select {
		case started <- struct{}{}:
		default:
		}
	}
	if release != nil {
		select {
		case <-release:
		case <-ctx.Done():
			return false, ctx.Err()
		}
	}
	return available, err
}

func (s *testSearcher) setAvailable(available bool) {
	s.mu.Lock()
	s.available = available
	s.mu.Unlock()
}

func (s *testSearcher) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

func TestWithEpisodesChecksInBackground(t *testing.T) {
	now := time.Now()
	searchStarted := make(chan struct{}, 1)
	searchRelease := make(chan struct{})
	searcher := &testSearcher{
		available: true,
		started:   searchStarted,
		release:   searchRelease,
	}
	monitor, updated := newTestMonitor(t, searcher)
	episode := newTestEpisode(now, 5)

	ret := monitor.withEpisodes([]*anime.Episode{episode}, now)
	require.Equal(t, anime.EpisodeTorrentAvailabilityChecking, ret[0].TorrentAvailability)
	require.Empty(t, episode.TorrentAvailability)
	waitForSignal(t, searchStarted)

	close(searchRelease)
	waitForSignal(t, updated)
	ret = monitor.withEpisodes([]*anime.Episode{episode}, now)
	require.Equal(t, anime.EpisodeTorrentAvailabilityAvailable, ret[0].TorrentAvailability)
	require.Equal(t, 1, searcher.callCount())
}

func TestMonitorRetriesUnresolvedEpisodes(t *testing.T) {
	now := time.Now()
	searcher := new(testSearcher)
	monitor, updated := newTestMonitor(t, searcher)
	monitor.interval = 100 * time.Millisecond
	episode := newTestEpisode(now, 5)

	ret := monitor.withEpisodes([]*anime.Episode{episode}, now)
	require.Equal(t, anime.EpisodeTorrentAvailabilityChecking, ret[0].TorrentAvailability)
	waitForSignal(t, updated)

	ret = monitor.withEpisodes([]*anime.Episode{episode}, now)
	require.Equal(t, anime.EpisodeTorrentAvailabilityWaiting, ret[0].TorrentAvailability)
	searcher.setAvailable(true)

	waitForSignal(t, updated)
	ret = monitor.withEpisodes([]*anime.Episode{episode}, now)
	require.Equal(t, anime.EpisodeTorrentAvailabilityAvailable, ret[0].TorrentAvailability)
	require.Equal(t, 2, searcher.callCount())

	time.Sleep(2 * monitor.interval)
	require.Equal(t, 2, searcher.callCount())
}

func TestWithEpisodesMarksSearchErrorsUnknown(t *testing.T) {
	now := time.Now()
	searcher := &testSearcher{err: errors.New("provider unavailable")}
	monitor, updated := newTestMonitor(t, searcher)
	episode := newTestEpisode(now, 5)

	ret := monitor.withEpisodes([]*anime.Episode{episode}, now)
	require.Equal(t, anime.EpisodeTorrentAvailabilityChecking, ret[0].TorrentAvailability)
	waitForSignal(t, updated)
	ret = monitor.withEpisodes([]*anime.Episode{episode}, now)
	require.Equal(t, anime.EpisodeTorrentAvailabilityUnknown, ret[0].TorrentAvailability)
}

func TestWithEpisodesSkipsOlderEpisodes(t *testing.T) {
	now := time.Now()
	searcher := new(testSearcher)
	monitor, _ := newTestMonitor(t, searcher)

	ret := monitor.withEpisodes([]*anime.Episode{newTestEpisode(now.Add(-48*time.Hour), 5)}, now)
	require.Empty(t, ret[0].TorrentAvailability)
	require.Equal(t, 0, searcher.callCount())
}

func TestAvailableStatusExpires(t *testing.T) {
	now := time.Now()
	searcher := &testSearcher{available: true}
	monitor, updated := newTestMonitor(t, searcher)
	episode := newTestEpisode(now, 5)

	monitor.withEpisodes([]*anime.Episode{episode}, now)
	waitForSignal(t, updated)
	monitor.check(now.Add(recentEpisodeWindow))
	waitForSignal(t, updated)

	ret := monitor.withEpisodes([]*anime.Episode{episode}, now.Add(recentEpisodeWindow))
	require.Empty(t, ret[0].TorrentAvailability)
	require.Equal(t, 1, searcher.callCount())
}

func TestWithEpisodesSkipsMissingEpisodeGroups(t *testing.T) {
	now := time.Now()
	searcher := new(testSearcher)
	monitor, _ := newTestMonitor(t, searcher)
	episode := newTestEpisode(now, 5)
	episode.IsMissingGroup = true

	ret := monitor.withEpisodes([]*anime.Episode{episode}, now)
	require.Empty(t, ret[0].TorrentAvailability)
	require.Equal(t, 0, searcher.callCount())
}

func TestRecentEpisodeWindow(t *testing.T) {
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	require.True(t, isRecentEpisode(newTestEpisode(now.Add(-recentEpisodeWindow+time.Minute), 5), now))
	require.False(t, isRecentEpisode(newTestEpisode(now.Add(-recentEpisodeWindow), 5), now))
}

func TestWithEpisodesUsesNextAiringFallback(t *testing.T) {
	now := time.Now()
	searcher := new(testSearcher)
	monitor, updated := newTestMonitor(t, searcher)
	episode := newTestEpisode(now, 5)
	episode.EpisodeMetadata.AirDate = ""
	episode.BaseAnime.NextAiringEpisode.Episode = 6
	episode.BaseAnime.NextAiringEpisode.AiringAt = int(now.Add(weeklyAiringInterval - 6*time.Hour).Unix())

	ret := monitor.withEpisodes([]*anime.Episode{episode}, now)
	require.Equal(t, anime.EpisodeTorrentAvailabilityChecking, ret[0].TorrentAvailability)
	waitForSignal(t, updated)
}

func TestWithEpisodesRequiresPreviousAiringEpisode(t *testing.T) {
	now := time.Now()
	searcher := new(testSearcher)
	monitor, _ := newTestMonitor(t, searcher)
	episode := newTestEpisode(now, 4)
	episode.EpisodeMetadata.AirDate = ""
	episode.BaseAnime.NextAiringEpisode.Episode = 6
	episode.BaseAnime.NextAiringEpisode.AiringAt = int(now.Add(weeklyAiringInterval - 6*time.Hour).Unix())

	ret := monitor.withEpisodes([]*anime.Episode{episode}, now)
	require.Empty(t, ret[0].TorrentAvailability)
	require.Equal(t, 0, searcher.callCount())
}

func TestWithEpisodesUsesEndDateForFinale(t *testing.T) {
	now := time.Now().UTC()
	searcher := new(testSearcher)
	monitor, updated := newTestMonitor(t, searcher)
	episode := newTestEpisode(now, 12)
	episode.EpisodeMetadata.AirDate = ""
	episode.BaseAnime.NextAiringEpisode = nil
	episode.BaseAnime.Status = new(anilist.MediaStatusFinished)
	episode.BaseAnime.Episodes = new(12)
	episode.BaseAnime.EndDate = &anilist.BaseAnime_EndDate{
		Year:  new(now.Year()),
		Month: new(int(now.Month())),
		Day:   new(now.Day()),
	}

	ret := monitor.withEpisodes([]*anime.Episode{episode}, now)
	require.Equal(t, anime.EpisodeTorrentAvailabilityChecking, ret[0].TorrentAvailability)
	waitForSignal(t, updated)
}

func TestWithEpisodesDoesNotUseEndDateBeforeFinale(t *testing.T) {
	now := time.Now().UTC()
	searcher := new(testSearcher)
	monitor, _ := newTestMonitor(t, searcher)
	episode := newTestEpisode(now, 11)
	episode.EpisodeMetadata.AirDate = ""
	episode.BaseAnime.NextAiringEpisode = nil
	episode.BaseAnime.Episodes = new(12)
	episode.BaseAnime.EndDate = &anilist.BaseAnime_EndDate{
		Year:  new(now.Year()),
		Month: new(int(now.Month())),
		Day:   new(now.Day()),
	}

	ret := monitor.withEpisodes([]*anime.Episode{episode}, now)
	require.Empty(t, ret[0].TorrentAvailability)
	require.Equal(t, 0, searcher.callCount())
}

func newTestMonitor(t *testing.T, searcher *testSearcher) (*monitor, <-chan struct{}) {
	t.Helper()
	updated := make(chan struct{}, 4)
	monitor := NewMonitor(
		searcher.search,
		func() (string, bool) {
			return "fake-provider", true
		},
		func() {
			select {
			case updated <- struct{}{}:
			default:
			}
		},
	)
	t.Cleanup(monitor.Stop)
	return monitor, updated
}

func waitForSignal(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for torrent availability update")
	}
}

func newTestEpisode(airedAt time.Time, episodeNumber int) *anime.Episode {
	media := testmocks.NewBaseAnime(100, "Example Show")
	media.Status = new(anilist.MediaStatusReleasing)
	media.NextAiringEpisode = new(anilist.BaseAnime_NextAiringEpisode)
	return &anime.Episode{
		BaseAnime:      media,
		EpisodeNumber:  episodeNumber,
		ProgressNumber: episodeNumber,
		EpisodeMetadata: &anime.EpisodeMetadata{
			AirDate: airedAt.Format(time.RFC3339),
		},
	}
}
