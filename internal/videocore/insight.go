package videocore

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"seanime/internal/mkvparser"
	"seanime/internal/util/limiter"
	"seanime/internal/util/result"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/5rahim/go-astisub"
	"github.com/imroc/req/v3"
	"github.com/rs/zerolog"
)

// InSight returns the characters for any given anime.
// TODO: map specific characters to a window of time where they appear in the subtitles (ASS character) or are mentioned by other characters.
type InSight struct {
	logger *zerolog.Logger
	vc     *VideoCore

	currentPosition       int64
	inSightData           *InSightData
	mu                    sync.RWMutex
	characters            []*InSightCharacter
	searchCache           []insightSearchEntry
	processedCounts       map[int]int
	cancelPolling         context.CancelFunc
	rateLimiter           *limiter.Limiter
	characterDetailsCache *result.BoundedCache[int, *InSightCharacterDetails]
}

type insightSearchEntry struct {
	MalID  int
	Tokens []string
}

type InSightData struct {
	Characters  []*InSightCharacter `json:"characters"`
	Suggestions []*InSightSegment   `json:"suggestions"`
}

type InSightSegment struct {
	CharacterId int     `json:"characterId"`
	StartTime   float64 `json:"startTime"`
	EndTime     float64 `json:"endTime"`
}

type InSightCharacter struct {
	MalID  int    `json:"mal_id"`
	URL    string `json:"url"`
	Images struct {
		Jpg struct {
			ImageUrl string `json:"image_url"`
		} `json:"jpg"`
		Webp struct {
			ImageUrl      string `json:"image_url"`
			SmallImageUrl string `json:"small_image_url"`
		} `json:"webp"`
	} `json:"images"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Favorites int    `json:"favorites"`
}

type InSightCharacterDetails struct {
	MalID  int    `json:"mal_id"`
	URL    string `json:"url"`
	Images struct {
		Jpg struct {
			ImageUrl string `json:"image_url"`
		} `json:"jpg"`
		Webp struct {
			ImageUrl      string `json:"image_url"`
			SmallImageUrl string `json:"small_image_url"`
		} `json:"webp"`
	} `json:"images"`
	Name      string   `json:"name"`
	NameKanji string   `json:"name_kanji"`
	Nicknames []string `json:"nicknames"`
	Favorites int      `json:"favorites"`
	About     string   `json:"about"`
}

type jikanCharacterFullResponse struct {
	Data InSightCharacterDetails `json:"data"`
}

var (
	JikanSeriesCharactersUrl = "https://api.jikan.moe/v4/anime/%d/characters"
	JikanCharacterUrl        = "https://api.jikan.moe/v4/characters/%d/full"
	reBraces                 = regexp.MustCompile(`\{.*?\}`)
	reTags                   = regexp.MustCompile(`<.*?>`)
)

func NewInSight(logger *zerolog.Logger, vc *VideoCore) *InSight {
	return &InSight{
		logger:                logger,
		vc:                    vc,
		characters:            make([]*InSightCharacter, 0),
		processedCounts:       make(map[int]int),
		rateLimiter:           limiter.NewLimiter(1*time.Second, 3), // max 3 requests per second
		characterDetailsCache: result.NewBoundedCache[int, *InSightCharacterDetails](100),
	}
}

func (vc *VideoCore) InSight() *InSight {
	return vc.inSight
}

func (is *InSight) sendToPlayer() {
	is.vc.SendInSightData(is.inSightData)
}

func (is *InSight) Start() {
	sub := is.vc.Subscribe("insight")
	go func() {
		for event := range sub.Events() {
			switch e := event.(type) {
			case *VideoLoadedEvent:
				go func(ev *VideoLoadedEvent) {
					if ev.State.PlaybackInfo != nil && ev.State.PlaybackInfo.Media != nil && ev.State.PlaybackInfo.Media.IDMal != nil {
						is.fetchCharacters(*ev.State.PlaybackInfo.Media.IDMal)
						// send to player
						is.sendToPlayer()
						//is.startPolling() todo
					}
				}(e)
			case *VideoSubtitleTrackEvent:
				// todo
				//go func(ev *VideoSubtitleTrackEvent) {
				//	if ev.Kind == "file" {
				//		is.vc.SendGetSubtitleTrackContent()
				//	}
				//}(e)
			case *VideoSubtitleTrackContentEvent:
				// todo
				//go func(ev *VideoSubtitleTrackContentEvent) {
				//	// Parse content
				//	events, err := is.ParseSubtitleContent(ev.Content, ev.Type)
				//	if err != nil {
				//		is.logger.Error().Err(err).Msg("insight: Failed to parse subtitle content")
				//		return
				//	}
				//	is.Analyze(events)
				//}(e)
			case *VideoTerminatedEvent:
				is.stopPolling()
				is.Clear()
			}
		}
	}()
}

func (is *InSight) startPolling() {
	is.stopPolling()

	ctx, cancel := context.WithCancel(context.Background())
	is.mu.Lock()
	is.cancelPolling = cancel
	is.mu.Unlock()

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		idleCount := 0
		maxIdleChecks := 10 // Stop after 20 seconds of no new events

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				foundNew := is.checkStreamingEvents()
				if foundNew {
					idleCount = 0
				} else {
					idleCount++
				}

				if idleCount >= maxIdleChecks {
					is.logger.Debug().Msg("insight: Polling stopped due to inactivity")
					return
				}
			}
		}
	}()
}

func (is *InSight) stopPolling() {
	if is.cancelPolling != nil {
		is.cancelPolling()
		is.cancelPolling = nil
	}
}

func (is *InSight) checkStreamingEvents() bool {
	if is.vc.playbackMkvEvents == nil {
		return false
	}

	foundNew := false

	is.vc.playbackMkvEvents.Range(func(trackNumber uint64, events []*mkvparser.SubtitleEvent) bool {
		is.mu.Lock()
		lastCount, ok := is.processedCounts[int(trackNumber)]
		if !ok {
			is.processedCounts[int(trackNumber)] = 0
			lastCount = 0
		}
		is.mu.Unlock()

		currentCount := len(events)
		if currentCount > lastCount {
			foundNew = true
			if len(events) > lastCount {
				newEvents := events[lastCount:]
				is.Analyze(newEvents)

				is.mu.Lock()
				is.processedCounts[int(trackNumber)] = currentCount
				is.mu.Unlock()
			}
		}
		return true
	})

	return foundNew
}

type jikanAnimeCharactersResponse struct {
	Data []*struct {
		Character struct {
			MalID  int    `json:"mal_id"`
			URL    string `json:"url"`
			Images struct {
				Jpg struct {
					ImageUrl string `json:"image_url"`
				} `json:"jpg"`
				Webp struct {
					ImageUrl      string `json:"image_url"`
					SmallImageUrl string `json:"small_image_url"`
				} `json:"webp"`
			} `json:"images"`
			Name string `json:"name"`
		} `json:"character"`
		Role      string `json:"role"`
		Favorites int    `json:"favorites"`
	} `json:"data"`
}

func (is *InSight) fetchCharacters(malId int) {
	if malId == 0 {
		return
	}
	is.logger.Debug().Int("malId", malId).Msg("insight: Fetching characters")

	is.rateLimiter.Wait()
	resp, err := req.C().R().Get(fmt.Sprintf(JikanSeriesCharactersUrl, malId))
	if err != nil {
		is.logger.Error().Err(err).Msg("insight: Failed to fetch characters")
		return
	}

	if resp.IsErrorState() {
		is.logger.Error().Msgf("insight: Failed to fetch characters: %s", resp.Status)
		return
	}

	var data jikanAnimeCharactersResponse
	if err := resp.UnmarshalJson(&data); err != nil {
		is.logger.Error().Err(err).Msg("insight: Failed to parse characters response")
		return
	}

	is.mu.Lock()
	defer is.mu.Unlock()

	is.characters = make([]*InSightCharacter, 0)
	is.searchCache = make([]insightSearchEntry, 0)
	for _, ch := range data.Data {
		is.characters = append(is.characters, &InSightCharacter{
			MalID:     ch.Character.MalID,
			URL:       ch.Character.URL,
			Images:    ch.Character.Images,
			Name:      ch.Character.Name,
			Role:      ch.Role,
			Favorites: ch.Favorites,
		})

		// Pre-compute tokens
		tokens := make([]string, 0)
		tokens = append(tokens, strings.ToLower(ch.Character.Name))
		parts := strings.FieldsFunc(ch.Character.Name, func(r rune) bool {
			return r == ',' || r == ' '
		})
		if len(parts) > 1 {
			for _, part := range parts {
				if len(part) > 2 {
					tokens = append(tokens, strings.ToLower(part))
				}
			}
		}
		is.searchCache = append(is.searchCache, insightSearchEntry{
			MalID:  ch.Character.MalID,
			Tokens: tokens,
		})
	}

	is.logger.Info().Int("count", len(is.characters)).Msg("insight: Characters fetched")
	if is.inSightData == nil {
		is.inSightData = &InSightData{
			Characters: is.characters,
		}
	} else {
		is.inSightData.Characters = is.characters
	}
}

// Analyze processes a batch of subtitle events and updates the InSight data.
// It is called when new subtitle events are available.
func (is *InSight) Analyze(events []*mkvparser.SubtitleEvent) {
	is.mu.RLock()
	if len(is.characters) == 0 {
		is.mu.RUnlock()
		return
	}

	cache := make([]insightSearchEntry, len(is.searchCache))
	copy(cache, is.searchCache)
	is.mu.RUnlock()

	suggestions := make([]*InSightSegment, 0)

	for _, event := range events {
		// Clean text
		text := is.CleanSubtitle(event.Text)
		if text == "" {
			continue
		}

		// Find matches
		matches := is.FindMatches(text, cache)
		if len(matches) == 0 {
			continue
		}

		// Create segments
		segments := is.CreateSegments(matches, event.StartTime, event.Duration)
		suggestions = append(suggestions, segments...)
	}

	//
	//
	//Merge overlapping segments for the same character
	suggestions = is.mergeOverlappingSegments(suggestions)

	if len(suggestions) == 0 {
		return
	}

	// Update data
	is.mu.Lock()
	defer is.mu.Unlock()

	if is.inSightData == nil {
		is.inSightData = &InSightData{}
	}
	is.inSightData.Suggestions = append(is.inSightData.Suggestions, suggestions...)
}

func (is *InSight) CleanSubtitle(text string) string {
	// femove formatting tags like {\...}
	text = reBraces.ReplaceAllString(text, "")
	// remove html tags
	text = reTags.ReplaceAllString(text, "")
	text = strings.TrimSpace(text)
	return text
}

func (is *InSight) FindMatches(text string, cache []insightSearchEntry) []int {
	matches := make([]int, 0)
	textLower := strings.ToLower(text)

	for _, entry := range cache {
		matched := false
		for _, token := range entry.Tokens {
			if strings.Contains(textLower, token) {
				matched = true
				break
			}
		}

		if matched {
			matches = append(matches, entry.MalID)
		}
	}
	return matches
}

func (is *InSight) CreateSegments(matches []int, startTime, duration float64) []*InSightSegment {
	segments := make([]*InSightSegment, 0)

	// it should linger for at least 3 seconds after the subtitle ends.
	// the total duration should be at least 6 seconds.

	const MinDuration = 6.0
	const Linger = 3.0

	endTime := startTime + duration + Linger
	if endTime-startTime < MinDuration {
		endTime = startTime + MinDuration
	}

	for _, id := range matches {
		segments = append(segments, &InSightSegment{
			CharacterId: id,
			StartTime:   startTime,
			EndTime:     endTime,
		})
	}
	return segments
}

func (is *InSight) mergeOverlappingSegments(segments []*InSightSegment) []*InSightSegment {
	if len(segments) <= 1 {
		return segments
	}

	// Group by character
	byChar := make(map[int][]*InSightSegment)
	for _, s := range segments {
		byChar[s.CharacterId] = append(byChar[s.CharacterId], s)
	}

	var merged []*InSightSegment

	for _, charSegments := range byChar {
		// Sort by start time
		sort.Slice(charSegments, func(i, j int) bool {
			return charSegments[i].StartTime < charSegments[j].StartTime
		})

		if len(charSegments) == 0 {
			continue
		}

		current := charSegments[0]

		for i := 1; i < len(charSegments); i++ {
			next := charSegments[i]

			// If overlap or close enough (gap < 2s)
			if next.StartTime <= current.EndTime+2.0 {
				// Merge
				if next.EndTime > current.EndTime {
					current.EndTime = next.EndTime
				}
			} else {
				merged = append(merged, current)
				current = next
			}
		}
		merged = append(merged, current)
	}

	// Sort final result by start time
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].StartTime < merged[j].StartTime
	})

	return merged
}

func (is *InSight) ParseSubtitleContent(content, subType string) ([]*mkvparser.SubtitleEvent, error) {
	events := make([]*mkvparser.SubtitleEvent, 0)

	// normalize line endings
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	var sub *astisub.Subtitles
	var err error

	if subType == "srt" {
		sub, err = astisub.ReadFromSRT(bytes.NewReader([]byte(content)))
	} else if subType == "vtt" {
		sub, err = astisub.ReadFromWebVTT(bytes.NewReader([]byte(content)))
	} else if subType == "ass" || subType == "ssa" {
		sub, err = astisub.ReadFromSSA(bytes.NewReader([]byte(content)))
	} else {
		return nil, fmt.Errorf("unsupported subtitle format: %s", subType)
	}

	if err != nil {
		return nil, err
	}

	for _, item := range sub.Items {
		var text []string
		for _, line := range item.Lines {
			var lineText string
			for _, lineItem := range line.Items {
				lineText += lineItem.Text
			}
			text = append(text, lineText)
		}

		events = append(events, &mkvparser.SubtitleEvent{
			StartTime: item.StartAt.Seconds(),
			Duration:  (item.EndAt - item.StartAt).Seconds(),
			Text:      strings.Join(text, "\n"),
		})
	}

	return events, nil
}

// Clear is called when the playback stops
func (is *InSight) Clear() {
	if is == nil {
		return
	}
	is.mu.Lock()
	defer is.mu.Unlock()
	is.inSightData = nil
	is.characters = make([]*InSightCharacter, 0)
	is.searchCache = make([]insightSearchEntry, 0)
	is.processedCounts = make(map[int]int)
	is.characterDetailsCache.Clear()
	is.stopPolling()
}

func (is *InSight) GetCharacterInfo(malId int) (*InSightCharacterDetails, error) {
	if cached, ok := is.characterDetailsCache.Get(malId); ok {
		return cached, nil
	}

	is.logger.Debug().Int("malId", malId).Msg("insight: Fetching character info")

	is.rateLimiter.Wait()

	resp, err := req.C().R().Get(fmt.Sprintf(JikanCharacterUrl, malId))
	if err != nil {
		return nil, err
	}

	if resp.IsErrorState() {
		return nil, fmt.Errorf("failed to fetch character info: %s", resp.Status)
	}

	var data jikanCharacterFullResponse
	if err := resp.UnmarshalJson(&data); err != nil {
		return nil, err
	}

	is.characterDetailsCache.Set(malId, &data.Data)

	return &data.Data, nil
}
