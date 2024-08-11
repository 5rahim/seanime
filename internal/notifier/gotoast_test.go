//go:build windows

package notifier

import (
	"github.com/go-toast/toast"
	"path/filepath"
	"seanime/internal/test_utils"
	"testing"
)

func TestGoToast(t *testing.T) {
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t)

	notification := toast.Notification{
		AppID:   "Seanime",
		Title:   "Seanime",
		Icon:    filepath.Join(test_utils.ConfigData.Path.DataDir, "logo.png"),
		Message: "Auto Downloader has downloaded 1 episode",
	}
	err := notification.Push()
	if err != nil {
		t.Fatal(err)
	}

}
