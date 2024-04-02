package discordrpc_presence

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/constants"
	"github.com/seanime-app/seanime/internal/database/models"
	"github.com/seanime-app/seanime/internal/discordrpc/client"
	"time"
)

type Presence struct {
	client   *discordrpc_client.Client
	settings *models.DiscordSettings
	logger   *zerolog.Logger
	lastSet  time.Time
	hasSent  bool
}

// New creates a new Presence instance.
// If rich presence is enabled, it sets up a new discord rpc client.
func New(settings *models.DiscordSettings, logger *zerolog.Logger) *Presence {
	var client *discordrpc_client.Client

	if settings.EnableRichPresence {
		client, _ = discordrpc_client.New(constants.DiscordApplicationId)
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
	if p.client == nil {
		return
	}
	p.client.Close()
}

func (p *Presence) SetSettings(settings *models.DiscordSettings) {
	// Close the current client
	p.Close()

	// Create a new client if rich presence is enabled
	if settings.EnableRichPresence {
		client, _ := discordrpc_client.New(constants.DiscordApplicationId)
		p.client = client
	} else {
		p.client = nil
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// checkLastSet checks if the last set was less than 5 seconds ago.
// If it was, it stops the function from setting the presence.
func (p *Presence) checkLastSet() (proceed bool) {
	if !p.hasSent {
		p.hasSent = true
		return true
	}
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
	if p.client == nil || !p.settings.EnableRichPresence {
		return
	}

	if !p.settings.EnableAnimeRichPresence {
		return
	}

	if !p.checkLastSet() {
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
	if p.client == nil || !p.settings.EnableRichPresence {
		return
	}

	if !p.settings.EnableMangaRichPresence {
		return
	}

	if !p.checkLastSet() {
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
