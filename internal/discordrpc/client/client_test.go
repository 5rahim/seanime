package discordrpc_client

import (
	"seanime/internal/constants"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	drpc, err := New(constants.DiscordApplicationId)
	if err != nil {
		t.Fatalf("failed to connect to discord ipc: %v", err)
	}
	defer drpc.Close()

	mangaActivity := Activity{
		Details: "Boku no Kokoro no Yabai Yatsu",
		State:   "Reading Chapter 30",
		Assets: &Assets{
			LargeImage: "https://s4.anilist.co/file/anilistcdn/media/manga/cover/medium/bx101557-bEJu54cmVYxx.jpg",
			LargeText:  "Boku no Kokoro no Yabai Yatsu",
			SmallImage: "logo",
			SmallText:  "Seanime",
		},
		Timestamps: &Timestamps{
			Start: &Epoch{
				Time: time.Now(),
			},
		},
		Instance: true,
		Type:     3,
	}

	go func() {
		_ = drpc.SetActivity(mangaActivity)
		time.Sleep(10 * time.Second)
		mangaActivity2 := mangaActivity
		mangaActivity2.Timestamps.Start.Time = time.Now()
		mangaActivity2.State = "Reading Chapter 31"
		_ = drpc.SetActivity(mangaActivity2)
		return
	}()

	//if err != nil {
	//	t.Fatalf("failed to set activity: %v", err)
	//}

	time.Sleep(30 * time.Second)
}
