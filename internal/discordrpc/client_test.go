package discordrpc

import (
	"github.com/seanime-app/seanime/internal/constants"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	drpc, err := New(constants.DiscordApplicationId)
	if err != nil {
		t.Fatalf("failed to connect to discord ipc: %v", err)
	}
	defer drpc.Close()

	err = drpc.SetActivity(Activity{
		Details: "Foo",
		State:   "Bar",
		Assets: &Assets{
			SmallImage: "keyart_hero",
			LargeImage: "keyart_hero",
		},
		Timestamps: &Timestamps{
			Start: &Epoch{Time: time.Now()},
		},
		Buttons: []*Button{
			{
				Label: "Button 1",
				Url:   "https://www.google.com",
			},
			{
				Label: "Button 2",
				Url:   "https://www.youtube.com",
			},
		},
	})

	if err != nil {
		t.Fatalf("failed to set activity: %v", err)
	}

	time.Sleep(10 * time.Second)

}
