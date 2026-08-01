package availability

import (
	"context"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/library/anime"
	"strings"
	"sync"
	"time"
)

const (
	checkInterval        = 5 * time.Minute
	recentEpisodeWindow  = 24 * time.Hour
	weeklyAiringInterval = 7 * 24 * time.Hour
)

type (
	monitor struct {
		ctx           context.Context
		cancel        context.CancelFunc
		search        func(context.Context, string, *anilist.BaseAnime, int) (bool, error)
		getProviderID func() (string, bool)
		onUpdated     func()
		mu            sync.Mutex
		items         map[string]*item
		results       map[string]result
		wake          chan struct{}
		running       bool
		interval      time.Duration
	}

	item struct {
		media         *anilist.BaseAnime
		episodeNumber int
		providerID    string
		expiresAt     time.Time
		nextCheck     time.Time
		available     bool
	}

	result struct {
		status    anime.EpisodeTorrentAvailability
		expiresAt time.Time
	}

	checkResult struct {
		key    string
		item   *item
		status anime.EpisodeTorrentAvailability
	}
)

func NewMonitor(
	search func(context.Context, string, *anilist.BaseAnime, int) (bool, error),
	getProviderID func() (string, bool),
	onUpdated func(),
) *monitor {
	ctx, cancel := context.WithCancel(context.Background())
	return &monitor{
		ctx:           ctx,
		cancel:        cancel,
		search:        search,
		getProviderID: getProviderID,
		onUpdated:     onUpdated,
		items:         make(map[string]*item),
		results:       make(map[string]result),
		wake:          make(chan struct{}, 1),
		interval:      checkInterval,
	}
}

func (m *monitor) Stop() {
	m.cancel()
	m.wakeWorker()
}

func (m *monitor) WithEpisodes(episodes []*anime.Episode) []*anime.Episode {
	return m.withEpisodes(episodes, time.Now())
}

func (m *monitor) withEpisodes(episodes []*anime.Episode, now time.Time) []*anime.Episode {
	ret := make([]*anime.Episode, len(episodes))
	providerID, providerFound := m.getProviderID()
	m.mu.Lock()
	for key, cached := range m.results {
		if !now.Before(cached.expiresAt) {
			delete(m.results, key)
		}
	}
	m.mu.Unlock()

	for i, episode := range episodes {
		if episode == nil {
			continue
		}

		clone := *episode
		clone.TorrentAvailability = ""
		ret[i] = &clone
		if clone.IsMissingGroup {
			continue
		}

		airedAt, found := episodeAiredAt(&clone)
		if !found || !isRecentAiring(airedAt, now) {
			continue
		}
		if !providerFound {
			clone.TorrentAvailability = anime.EpisodeTorrentAvailabilityUnknown
			continue
		}

		clone.TorrentAvailability = m.track(
			providerID,
			clone.BaseAnime,
			clone.GetEpisodeNumber(),
			airedAt.Add(recentEpisodeWindow),
			now,
		)
	}

	return ret
}

func (m *monitor) track(providerID string, media *anilist.BaseAnime, episodeNumber int, expiresAt, now time.Time) anime.EpisodeTorrentAvailability {
	key := resultKey(providerID, media.GetID(), episodeNumber)
	startWorker := false
	wakeWorker := false

	m.mu.Lock()
	cached, found := m.results[key]
	if found && !now.Before(cached.expiresAt) {
		delete(m.results, key)
		cached = result{}
		found = false
	}

	status := anime.EpisodeTorrentAvailabilityChecking
	if found {
		status = cached.status
	}

	if status != anime.EpisodeTorrentAvailabilityAvailable {
		if _, tracked := m.items[key]; !tracked {
			m.items[key] = &item{
				media:         media,
				episodeNumber: episodeNumber,
				providerID:    providerID,
				expiresAt:     expiresAt,
				nextCheck:     now,
			}
			wakeWorker = true
		}
		if !m.running {
			m.running = true
			startWorker = true
		}
	}
	m.mu.Unlock()

	if startWorker {
		go m.run()
	}
	if wakeWorker {
		m.wakeWorker()
	}

	return status
}

func (m *monitor) run() {
	for {
		m.check(time.Now())

		m.mu.Lock()
		if len(m.items) == 0 {
			m.running = false
			m.mu.Unlock()
			return
		}

		nextCheck := time.Time{}
		for _, item := range m.items {
			if nextCheck.IsZero() || item.nextCheck.Before(nextCheck) {
				nextCheck = item.nextCheck
			}
		}
		m.mu.Unlock()

		wait := time.Until(nextCheck)
		if wait < 0 {
			wait = 0
		}
		timer := time.NewTimer(wait)
		select {
		case <-timer.C:
		case <-m.wake:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
		case <-m.ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			m.mu.Lock()
			m.running = false
			m.mu.Unlock()
			return
		}
	}
}

func (m *monitor) wakeWorker() {
	select {
	case m.wake <- struct{}{}:
	default:
	}
}

func (m *monitor) check(now time.Time) {
	providerID, providerFound := m.getProviderID()
	if !providerFound {
		providerID = ""
	}

	items := make([]struct {
		key  string
		item *item
	}, 0)
	changed := false

	m.mu.Lock()
	for key, cached := range m.results {
		if !now.Before(cached.expiresAt) {
			delete(m.results, key)
		}
	}
	for key, trackedItem := range m.items {
		if !now.Before(trackedItem.expiresAt) || trackedItem.providerID != providerID {
			delete(m.items, key)
			delete(m.results, key)
			changed = true
			continue
		}
		if trackedItem.nextCheck.After(now) {
			continue
		}
		if trackedItem.available {
			continue
		}

		trackedItem.nextCheck = now.Add(m.interval)
		items = append(items, struct {
			key  string
			item *item
		}{key: key, item: trackedItem})
	}
	m.mu.Unlock()

	checks := make(chan checkResult, len(items))
	sem := make(chan struct{}, 4)
	var wg sync.WaitGroup
	for _, tracked := range items {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			checks <- checkResult{
				key:    tracked.key,
				item:   tracked.item,
				status: m.searchStatus(tracked.item),
			}
		}()
	}
	wg.Wait()
	close(checks)

	if m.ctx.Err() != nil {
		return
	}

	m.mu.Lock()
	for check := range checks {
		item, tracked := m.items[check.key]
		if !tracked || item != check.item {
			continue
		}

		previous, found := m.results[check.key]
		if !found || previous.status != check.status {
			changed = true
		}
		m.results[check.key] = result{
			status:    check.status,
			expiresAt: item.expiresAt,
		}
		if check.status == anime.EpisodeTorrentAvailabilityAvailable {
			item.available = true
			item.nextCheck = item.expiresAt
		}
	}
	m.mu.Unlock()

	if changed && m.onUpdated != nil {
		m.onUpdated()
	}
}

func (m *monitor) searchStatus(item *item) anime.EpisodeTorrentAvailability {
	available, err := m.search(m.ctx, item.providerID, item.media, item.episodeNumber)
	if err != nil {
		return anime.EpisodeTorrentAvailabilityUnknown
	}
	if available {
		return anime.EpisodeTorrentAvailabilityAvailable
	}
	return anime.EpisodeTorrentAvailabilityWaiting
}

func resultKey(providerID string, mediaID, episodeNumber int) string {
	return fmt.Sprintf("%s-%d-%d", providerID, mediaID, episodeNumber)
}

func isRecentEpisode(episode *anime.Episode, now time.Time) bool {
	airedAt, found := episodeAiredAt(episode)
	return found && isRecentAiring(airedAt, now)
}

func isRecentAiring(airedAt, now time.Time) bool {
	age := now.Sub(airedAt)
	return age >= 0 && age < recentEpisodeWindow
}

func episodeAiredAt(episode *anime.Episode) (time.Time, bool) {
	if episode == nil || episode.BaseAnime == nil {
		return time.Time{}, false
	}

	if episode.EpisodeMetadata != nil {
		airDate := strings.TrimSpace(episode.EpisodeMetadata.AirDate)
		if airDate != "" {
			if parsed, err := time.Parse(time.RFC3339, airDate); err == nil {
				return parsed, true
			}
			if parsed, err := time.Parse(time.DateOnly, airDate); err == nil {
				return parsed, true
			}
		}
	}

	next := episode.BaseAnime.GetNextAiringEpisode()
	progressNumber := episode.GetProgressNumber()
	if progressNumber <= 0 {
		progressNumber = episode.GetEpisodeNumber()
	}
	if next != nil && next.GetAiringAt() > 0 && next.GetEpisode() == progressNumber+1 {
		return time.Unix(int64(next.GetAiringAt()), 0).Add(-weeklyAiringInterval), true
	}

	if next != nil || progressNumber != episode.BaseAnime.GetTotalEpisodeCount() {
		return time.Time{}, false
	}

	endDate := episode.BaseAnime.GetEndDate()
	if endDate == nil || endDate.GetYear() == nil || endDate.GetMonth() == nil || endDate.GetDay() == nil {
		return time.Time{}, false
	}
	return time.Date(
		*endDate.GetYear(),
		time.Month(*endDate.GetMonth()),
		*endDate.GetDay(),
		0, 0, 0, 0,
		time.UTC,
	), true
}
