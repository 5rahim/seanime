package discordrpc_presence

import (
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
	lastSet  time.Time
	hasSent  bool
	username string
	mu       sync.Mutex
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

	return &Presence{
		client:   client,
		settings: settings,
		logger:   logger,
		lastSet:  time.Now(),
		hasSent:  false,
	}
}

// Close closes the discord rpc client.
// If the client is nil, it does nothing.
func (p *Presence) Close() {
	defer util.HandlePanicInModuleThen("discordrpc/presence/Close", func() {})

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

	// Close the current client
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
			p.logger.Error().Err(err).Msg("discordrpc: rich presence enabled but failed to create discord rpc client")
			return
		}
		p.client = client
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

	// If the last set time is less than 5 seconds ago, return false
	if time.Since(p.lastSet) < 5*time.Second {
		rest := 5*time.Second - time.Since(p.lastSet)
		time.Sleep(rest)
	}

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
}

// SetAnimeActivity sets the presence to watching anime.
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

	p.logger.Debug().Msgf("discordrpc: setting anime activity: %s", a.Title)
	_ = p.client.SetActivity(activity)
	p.lastSet = time.Now()
}

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

	p.logger.Debug().Msgf("discordrpc: setting manga activity: %s", a.Title)
	_ = p.client.SetActivity(activity)
	p.lastSet = time.Now()
}
