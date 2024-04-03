package discordrpc_presence

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/constants"
	"github.com/seanime-app/seanime/internal/database/models"
	"github.com/seanime-app/seanime/internal/discordrpc/client"
	"github.com/seanime-app/seanime/internal/util"
	"sync"
	"time"
)

type Presence struct {
	client   *discordrpc_client.Client
	settings *models.DiscordSettings
	logger   *zerolog.Logger
	lastSet  time.Time
	hasSent  bool
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
			logger.Error().Err(err).Msg("discord rpc: rich presence enabled but failed to create discord rpc client")
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
		p.logger.Debug().Msg("discord rpc: rich presence enabled")
		p.setClient()
	} else {
		p.client = nil
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (p *Presence) setClient() {
	defer util.HandlePanicInModuleThen("discordrpc/presence/setClient", func() {})

	if p.client == nil {
		client, err := discordrpc_client.New(constants.DiscordApplicationId)
		if err != nil {
			p.logger.Error().Err(err).Msg("discord rpc: rich presence enabled but failed to create discord rpc client")
			return
		}
		p.client = client
	}
}

// check executes multiple checks to determine if the presence should be set.
// It returns true if the presence should be set.
func (p *Presence) check() (proceed bool) {
	defer util.HandlePanicInModuleThen("discordrpc/presence/check", func() {
		proceed = false
	})

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
		return false
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
			SmallImage: "logo",
			SmallText:  "Seanime",
		},
		Timestamps: &discordrpc_client.Timestamps{
			Start: &discordrpc_client.Epoch{
				Time: time.Now(),
			},
		},
		Instance: true,
		Type:     3,
	}
)

type AnimeActivity struct {
	Title         string
	Image         string
	IsMovie       bool
	EpisodeNumber int
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

	p.logger.Debug().Msgf("discord rpc: setting anime activity: %s", a.Title)
	_ = p.client.SetActivity(activity)
	p.lastSet = time.Now()
}

type MangaActivity struct {
	Title         string
	Image         string
	ChapterNumber int
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
	activity.State = fmt.Sprintf("Reading Chapter %d", a.ChapterNumber)
	activity.Assets.LargeImage = a.Image
	activity.Assets.LargeText = a.Title
	activity.Timestamps.Start.Time = time.Now()

	p.logger.Debug().Msgf("discord rpc: setting manga activity: %s", a.Title)
	_ = p.client.SetActivity(activity)
	p.lastSet = time.Now()
}
