package discordrpc_presence

import (
	"seanime/internal/database/models"
	"seanime/internal/util"
	"testing"
	"time"
)

func TestPresence(t *testing.T) {

	settings := &models.DiscordSettings{
		EnableRichPresence:      true,
		EnableAnimeRichPresence: true,
		EnableMangaRichPresence: true,
	}

	presence := New(nil, util.NewLogger())
	presence.SetSettings(settings)
	presence.SetUsername("test")
	defer presence.Close()

	presence.SetMangaActivity(&MangaActivity{
		Title:   "Boku no Kokoro no Yabai Yatsu",
		Image:   "https://s4.anilist.co/file/anilistcdn/media/manga/cover/medium/bx101557-bEJu54cmVYxx.jpg",
		Chapter: "30",
	})

	time.Sleep(10 * time.Second)

	// Simulate settings being updated

	settings.EnableMangaRichPresence = false
	presence.SetSettings(settings)
	presence.SetUsername("test")

	time.Sleep(5 * time.Second)

	presence.SetMangaActivity(&MangaActivity{
		Title:   "Boku no Kokoro no Yabai Yatsu",
		Image:   "https://s4.anilist.co/file/anilistcdn/media/manga/cover/medium/bx101557-bEJu54cmVYxx.jpg",
		Chapter: "31",
	})

	// Simulate settings being updated

	settings.EnableMangaRichPresence = true
	presence.SetSettings(settings)
	presence.SetUsername("test")

	time.Sleep(5 * time.Second)

	presence.SetMangaActivity(&MangaActivity{
		Title:   "Boku no Kokoro no Yabai Yatsu",
		Image:   "https://s4.anilist.co/file/anilistcdn/media/manga/cover/medium/bx101557-bEJu54cmVYxx.jpg",
		Chapter: "31",
	})

	time.Sleep(10 * time.Second)
}
