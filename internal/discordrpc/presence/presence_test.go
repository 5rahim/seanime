package discordrpc_presence

import (
	"seanime/internal/api/anilist"
	"seanime/internal/database/models"
	discordrpc_client "seanime/internal/discordrpc/client"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestNewAnimeActivityIncludesEpisodeDetails(t *testing.T) {
	title := "Test Anime"
	image := "https://example.com/cover.jpg"
	totalEpisodes := 12
	nextEpisode := 9

	activity := NewAnimeActivity(&anilist.BaseAnime{
		ID:       123,
		Episodes: &totalEpisodes,
		Title: &anilist.BaseAnime_Title{
			UserPreferred: &title,
		},
		CoverImage: &anilist.BaseAnime_CoverImage{
			ExtraLarge: &image,
		},
		NextAiringEpisode: &anilist.BaseAnime_NextAiringEpisode{
			Episode: nextEpisode,
		},
	}, 8, "The Turning Point", 30, 1440)

	require.Equal(t, 123, activity.ID)
	require.Equal(t, title, activity.Title)
	require.Equal(t, image, activity.Image)
	require.Equal(t, 8, activity.EpisodeNumber)
	require.Equal(t, 30, activity.Progress)
	require.Equal(t, 1440, activity.Duration)
	require.Equal(t, totalEpisodes, *activity.TotalEpisodes)
	require.Equal(t, nextEpisode-1, *activity.CurrentEpisodeCount)
	require.Equal(t, "The Turning Point", *activity.EpisodeTitle)
}

func TestNewAnimeActivityOmitsEmptyEpisodeTitle(t *testing.T) {
	activity := NewAnimeActivity(&anilist.BaseAnime{}, 1, "", 0, 0)

	require.Nil(t, activity.EpisodeTitle)
}

func TestSetCustomActivity(t *testing.T) {
	settings := &models.DiscordSettings{
		EnableRichPresence: true,
	}
	p := &Presence{
		settings:   settings,
		logger:     new(zerolog.Nop()),
		eventQueue: make(chan func(), 100),
		hasSent:    true,
		client:     &discordrpc_client.Client{},
	}
	p.animeActivity = &AnimeActivity{ID: 123}

	customAct := &CustomActivity{
		Type:           new(2),
		Details:        "Testing Custom Activity",
		State:          "Playing around",
		LargeImageKey:  "large-img",
		LargeImageText: "Large Image",
		StartTimestamp: new(int64),
	}
	*customAct.StartTimestamp = 0

	p.SetCustomActivity(customAct)

	require.Nil(t, p.animeActivity)
	require.Len(t, p.eventQueue, 1)

	time.Sleep(5 * time.Second)
}

//func TestSetCustomActivityReal(t *testing.T) {
//	settings := &models.DiscordSettings{
//		EnableRichPresence: true,
//	}
//	p := New(settings, new(zerolog.New(zerolog.NewConsoleWriter())))
//	if p.client == nil {
//		t.Skip("Discord client not running or failed to connect")
//	}
//
//	p.animeActivity = &AnimeActivity{ID: 123}
//
//	customType := 3 // Watching
//	customAct := &CustomActivity{
//		Type:           &customType,
//		Details:        "Testing Custom Activity",
//		State:          "Watching stuff",
//		LargeImageKey:  "https://seanime.app/images/circular-logo.png",
//		LargeImageText: "Seanime Logo",
//	}
//
//	p.SetCustomActivity(customAct)
//
//	time.Sleep(15 * time.Second)
//
//	require.Nil(t, p.animeActivity)
//}
