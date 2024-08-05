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

	t.Log(test_utils.ConfigData.Path.DataDir + "logo.png")

	notification := toast.Notification{
		AppID:   "Seanime",
		Title:   "Downloaded 1 episode",
		Icon:    filepath.Join(test_utils.ConfigData.Path.DataDir, "logo.png"),
		Message: "Some message about how important something is...",
	}
	err := notification.Push()
	if err != nil {
		t.Fatal(err)
	}

}
