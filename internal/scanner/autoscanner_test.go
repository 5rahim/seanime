package scanner

import (
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/util"
	"testing"
	"time"
)

func TestAutoScanner(t *testing.T) {

	doneCh := make(chan struct{})

	logger := util.NewLogger()
	anilistClientWrapper, _ := anilist.TestGetAnilistClientWrapperAndInfo()

	as := NewAutoScanner(&NewAutoScannerOptions{
		Database:             nil,
		Enabled:              false,
		AutoDownloader:       nil,
		AnilistClientWrapper: anilistClientWrapper,
		Logger:               logger,
		WSEventManager:       events.NewMockWSEventManager(logger),
		WaitTime:             5 * time.Second, // Set to 5 seconds for testing
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
