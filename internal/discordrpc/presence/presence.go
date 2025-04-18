package discordrpc_presence

import (
	"context"
	"fmt"
	"seanime/internal/constants"
	"seanime/internal/database/models"
	discordrpc_client "seanime/internal/discordrpc/client"
	"seanime/internal/util"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type Presence struct {
	client   *discordrpc_client.Client
	settings *models.DiscordSettings
	logger   *zerolog.Logger
	hasSent  bool
	username string
	mu       sync.RWMutex

	animeActivity               *AnimeActivity
	lastAnimeActivityUpdateSent time.Time

	lastSent   time.Time
	eventQueue chan func()
	cancelFunc context.CancelFunc // Cancel function for the event loop context
}

// New creates a new Presence instance.
// If rich presence is enabled, it sets up a new discord rpc client.
func New(settings *models.DiscordSettings, logger *zerolog.Logger) *Presence {
	var client *discordrpc_client.Client

	if settings != nil && settings.EnableRichPresence {
		var err error
		client, err = discordrpc_client.New(constants.DiscordApplicationId)
		if err != nil {
			logger.Error().Err(err).Msg("discordrpc: rich presence enabled but failed to create discord rpc client")
		}
	}

	p := &Presence{
		client:                      client,
		settings:                    settings,
		logger:                      logger,
		lastAnimeActivityUpdateSent: time.Now().Add(5 * time.Second),
		lastSent:                    time.Now().Add(-5 * time.Second),
		hasSent:                     false,
		eventQueue:                  make(chan func(), 100),
	}

	if settings != nil && settings.EnableRichPresence {
		p.startEventLoop()
	}

	return p
}

func (p *Presence) startEventLoop() {
	// Cancel any existing goroutine
	if p.cancelFunc != nil {
		p.cancelFunc()
	}

	// Create new context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	p.cancelFunc = cancel

	ticker := time.NewTicker(5 * time.Second)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				p.logger.Debug().Msg("discordrpc: Event loop stopped")
				return
			case <-ticker.C:
				select {
				case job := <-p.eventQueue:
					p.mu.RLock()
					if p.client == nil {
						p.mu.RUnlock()
						continue
					}
					job()
					p.lastSent = time.Now()
					p.mu.RUnlock()
				default:
				}
			}
		}
	}()
}

// Close closes the discord rpc client.
// If the client is nil, it does nothing.
func (p *Presence) Close() {
	defer util.HandlePanicInModuleThen("discordrpc/presence/Close", func() {})

	p.clearEventQueue()

	// Cancel the event loop goroutine
	if p.cancelFunc != nil {
		p.cancelFunc()
		p.cancelFunc = nil
	}

	if p.client == nil {
		return
	}
	p.client.Close()
	p.client = nil
}

func (p *Presence) SetSettings(settings *models.DiscordSettings) {
	p.mu.Lock()
	defer p.mu.Unlock()

	defer util.HandlePanicInModuleThen("discordrpc/presence/SetSettings", func() {})

	// Close the current client and stop event loop
	p.Close()

	p.settings = settings

	// Create a new client if rich presence is enabled
	if settings.EnableRichPresence {
		p.logger.Info().Msg("discordrpc: Discord Rich Presence enabled")
		p.setClient()
	} else {
		p.client = nil
	}
}

func (p *Presence) SetUsername(username string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.username = username
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (p *Presence) setClient() {
	defer util.HandlePanicInModuleThen("discordrpc/presence/setClient", func() {})

	if p.client == nil {
		client, err := discordrpc_client.New(constants.DiscordApplicationId)
		if err != nil {
			p.logger.Error().Err(err).Msg("discordrpc: Rich presence enabled but failed to create discord rpc client")
			return
		}
		p.client = client
		p.startEventLoop()
		p.logger.Debug().Msg("discordrpc: RPC client initialized and event loop started")
	}
}

var isChecking bool

// check executes multiple checks to determine if the presence should be set.
// It returns true if the presence should be set.
func (p *Presence) check() (proceed bool) {
	defer util.HandlePanicInModuleThen("discordrpc/presence/check", func() {
		proceed = false
	})

	if isChecking {
		return false
	}
	isChecking = true
	defer func() {
		isChecking = false
	}()

	// If the client is nil, return false
	if p.settings == nil {
		return false
	}

	// If rich presence is disabled, return false
	if !p.settings.EnableRichPresence {
		return false
	}

	// If the client is nil, create a new client
	if p.client == nil {
		p.setClient()
	}

	// If the client is still nil, return false
	if p.client == nil {
		return false
	}

	// If this is the first time setting the presence, return true
	if !p.hasSent {
		p.hasSent = true
		return true
	}

	// // If the last sent time is less than 5 seconds ago, return false
	// if time.Since(p.lastSent) < 5*time.Second {
	// 	rest := 5*time.Second - time.Since(p.lastSent)
	// 	time.Sleep(rest)
	// }

	return true
}

var (
	defaultActivity = discordrpc_client.Activity{
		Details: "",
		State:   "",
		Assets: &discordrpc_client.Assets{
			LargeImage: "",
			LargeText:  "",
			SmallImage: "https://seanime.rahim.app/images/circular-logo.png",
			SmallText:  "Seanime v" + constants.Version,
		},
		Timestamps: &discordrpc_client.Timestamps{
			Start: &discordrpc_client.Epoch{
				Time: time.Now(),
			},
			End: &discordrpc_client.Epoch{
				Time: time.Now(),
			},
		},
		Buttons: []*discordrpc_client.Button{
			{
				Label: "Seanime",
				Url:   "https://github.com/5rahim/seanime",
			},
		},
		Instance: true,
		Type:     3,
	}
)

type AnimeActivity struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	Image         string `json:"image"`
	IsMovie       bool   `json:"isMovie"`
	EpisodeNumber int    `json:"episodeNumber"`
	Paused        bool   `json:"paused"`
	Progress      int    `json:"progress"`
	Duration      int    `json:"duration"`
}

func animeActivityKey(a *AnimeActivity) string {
	return fmt.Sprintf("%d:%d", a.ID, a.EpisodeNumber)
}

func (p *Presence) SetAnimeActivity(a *AnimeActivity) {
	p.mu.Lock()
	defer p.mu.Unlock()

	defer util.HandlePanicInModuleThen("discordrpc/presence/SetAnimeActivity", func() {})

	if !p.check() {
		return
	}

	if !p.settings.EnableAnimeRichPresence {
		return
	}

	// Clear the queue if the anime activity is different
	if p.animeActivity != nil && animeActivityKey(a) != animeActivityKey(p.animeActivity) {
		p.clearEventQueue()
	}

	state := fmt.Sprintf("Watching Episode %d", a.EpisodeNumber)
	if a.IsMovie {
		state = "Watching Movie"
	}

	activity := defaultActivity
	activity.Details = a.Title
	activity.State = state
	activity.Assets.LargeImage = a.Image
	activity.Assets.LargeText = a.Title

	// Calculate the start time
	startTime := time.Now()
	if a.Progress > 0 {
		startTime = startTime.Add(-time.Duration(a.Progress) * time.Second)
	}

	activity.Timestamps.Start.Time = startTime
	activity.Timestamps.End = &discordrpc_client.Epoch{
		Time: startTime.Add(time.Duration(a.Duration) * time.Second),
	}
	activity.Buttons = make([]*discordrpc_client.Button, 0)

	if p.settings.RichPresenceShowAniListMediaButton && a.ID != 0 {
		activity.Buttons = append(activity.Buttons, &discordrpc_client.Button{
			Label: "View Anime",
			Url:   fmt.Sprintf("https://anilist.co/anime/%d", a.ID),
		})
	}

	if p.settings.RichPresenceShowAniListProfileButton {
		activity.Buttons = append(activity.Buttons, &discordrpc_client.Button{
			Label: "View Profile",
			Url:   fmt.Sprintf("https://anilist.co/user/%s", p.username),
		})
	}

	if !(p.settings.RichPresenceHideSeanimeRepositoryButton || len(activity.Buttons) > 1) {
		activity.Buttons = append(activity.Buttons, &discordrpc_client.Button{
			Label: "Seanime",
			Url:   "https://github.com/5rahim/seanime",
		})
	}

	// p.logger.Debug().Msgf("discordrpc: Setting anime activity: %s", a.Title)

	p.animeActivity = a

	select {
	case p.eventQueue <- func() {
		_ = p.client.SetActivity(activity)
		// p.logger.Debug().Int("progress", a.Progress).Int("duration", a.Duration).Msgf("discordrpc: Anime activity set")
	}:
	default:
		// p.logger.Error().Msg("discordrpc: event queue is full")
	}
}

// clearEventQueue drains the event queue channel
func (p *Presence) clearEventQueue() {
	//p.logger.Debug().Msg("discordrpc: Clearing event queue")
	for {
		select {
		case <-p.eventQueue:
		default:
			return
		}
	}
}

func (p *Presence) UpdateAnimeActivity(progress int, duration int, paused bool) {
	// do not lock, we call SetAnimeActivity

	defer util.HandlePanicInModuleThen("discordrpc/presence/UpdateWatching", func() {})

	if p.animeActivity == nil {
		return
	}

	p.animeActivity.Progress = progress
	p.animeActivity.Duration = duration

	// Pause status	changed
	if p.animeActivity.Paused != paused {
		// p.logger.Debug().Msgf("discordrpc: Pause status changed to %t", paused)
		p.animeActivity.Paused = paused
		p.lastAnimeActivityUpdateSent = time.Now()

		// Clear the event queue to ensure pause/unpause takes precedence
		p.clearEventQueue()

		if paused {
			// p.logger.Debug().Msg("discordrpc: Stopping activity")
			// Stop the current activity if paused
			p.Close()
		} else {
			// p.logger.Debug().Msg("discordrpc: Restarting activity")
			// Restart the current activity if unpaused
			p.SetAnimeActivity(p.animeActivity)
		}
		return
	}

	// Handles seeking
	if !p.animeActivity.Paused {
		// If the last update was more than 5 seconds ago, update the activity
		if time.Since(p.lastAnimeActivityUpdateSent) > 6*time.Second {
			// p.logger.Debug().Msg("discordrpc: Updating activity")
			p.lastAnimeActivityUpdateSent = time.Now()
			p.SetAnimeActivity(p.animeActivity)
		}
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type LegacyAnimeActivity struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	Image         string `json:"image"`
	IsMovie       bool   `json:"isMovie"`
	EpisodeNumber int    `json:"episodeNumber"`
}

// LegacySetAnimeActivity sets the presence to watching anime.
func (p *Presence) LegacySetAnimeActivity(a *LegacyAnimeActivity) {
	p.mu.Lock()
	defer p.mu.Unlock()

	defer util.HandlePanicInModuleThen("discordrpc/presence/SetAnimeActivity", func() {})

	if !p.check() {
		return
	}

	if !p.settings.EnableAnimeRichPresence {
		return
	}

	state := fmt.Sprintf("Watching Episode %d", a.EpisodeNumber)
	if a.IsMovie {
		state = "Watching Movie"
	}

	activity := defaultActivity
	activity.Details = a.Title
	activity.State = state
	activity.Assets.LargeImage = a.Image
	activity.Assets.LargeText = a.Title
	activity.Timestamps.Start.Time = time.Now()
	activity.Timestamps.End = nil
	activity.Buttons = make([]*discordrpc_client.Button, 0)

	if p.settings.RichPresenceShowAniListMediaButton && a.ID != 0 {
		activity.Buttons = append(activity.Buttons, &discordrpc_client.Button{
			Label: "View Anime",
			Url:   fmt.Sprintf("https://anilist.co/anime/%d", a.ID),
		})
	}

	if p.settings.RichPresenceShowAniListProfileButton {
		activity.Buttons = append(activity.Buttons, &discordrpc_client.Button{
			Label: "View Profile",
			Url:   fmt.Sprintf("https://anilist.co/user/%s", p.username),
		})
	}

	if !(p.settings.RichPresenceHideSeanimeRepositoryButton || len(activity.Buttons) > 1) {
		activity.Buttons = append(activity.Buttons, &discordrpc_client.Button{
			Label: "Seanime",
			Url:   "https://github.com/5rahim/seanime",
		})
	}

	// p.logger.Debug().Msgf("discordrpc: Setting anime activity: %s", a.Title)

	p.eventQueue <- func() {
		_ = p.client.SetActivity(activity)
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type MangaActivity struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Image   string `json:"image"`
	Chapter string `json:"chapter"`
}

// SetMangaActivity sets the presence to watching anime.
func (p *Presence) SetMangaActivity(a *MangaActivity) {
	p.mu.Lock()
	defer p.mu.Unlock()

	defer util.HandlePanicInModuleThen("discordrpc/presence/SetMangaActivity", func() {})

	if !p.check() {
		return
	}

	if !p.settings.EnableMangaRichPresence {
		return
	}

	activity := defaultActivity
	activity.Details = a.Title
	activity.State = fmt.Sprintf("Reading Chapter %s", a.Chapter)
	activity.Assets.LargeImage = a.Image
	activity.Assets.LargeText = a.Title
	activity.Timestamps.Start.Time = time.Now()
	activity.Buttons = make([]*discordrpc_client.Button, 0)

	if p.settings.RichPresenceShowAniListMediaButton && a.ID != 0 {
		activity.Buttons = append(activity.Buttons, &discordrpc_client.Button{
			Label: "View Manga",
			Url:   fmt.Sprintf("https://anilist.co/manga/%d", a.ID),
		})
	}

	if p.settings.RichPresenceShowAniListProfileButton && p.username != "" {
		activity.Buttons = append(activity.Buttons, &discordrpc_client.Button{
			Label: "View Profile",
			Url:   fmt.Sprintf("https://anilist.co/user/%s", p.username),
		})
	}

	if !(p.settings.RichPresenceHideSeanimeRepositoryButton || len(activity.Buttons) > 1) {
		activity.Buttons = append(activity.Buttons, &discordrpc_client.Button{
			Label: "Seanime",
			Url:   "https://github.com/5rahim/seanime",
		})
	}

	p.logger.Debug().Msgf("discordrpc: Setting manga activity: %s", a.Title)

	p.eventQueue <- func() {
		_ = p.client.SetActivity(activity)
	}
}
