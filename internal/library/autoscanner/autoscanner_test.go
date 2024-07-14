package autoscanner

import (
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/platform"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/util"
	"testing"
	"time"
)

func TestAutoScanner(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	doneCh := make(chan struct{})

	logger := util.NewLogger()
	anilistClient := anilist.TestGetMockAnilistClient()
	anilistPlatform := platform.NewAnilistPlatform(anilistClient, logger)

	as := New(&NewAutoScannerOptions{
		Database:       nil,
		Enabled:        false,
		AutoDownloader: nil,
		Platform:       anilistPlatform,
		Logger:         logger,
		WSEventManager: events.NewMockWSEventManager(logger),
		WaitTime:       5 * time.Second, // Set to 5 seconds for testing
	})

	go as.SetEnabled(true)

	as.Start()

	go func() {
		as.Notify()
		<-time.After(2 * time.Second)
		as.Notify()
	}()

	go func() {
		<-time.After(30 * time.Second)
		close(doneCh)
	}()

	select {
	case <-doneCh:
		t.Log("AutoScanner test done")
		break
	}
}
