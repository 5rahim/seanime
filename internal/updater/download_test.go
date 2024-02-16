package updater

import (
	"github.com/samber/lo"
	"strings"
	"testing"
)

var repoPath = "E:/SEANIME-REPO-TEST"

func TestUpdater_DownloadLatestRelease(t *testing.T) {

	updater := New("0.2.0")

	release, err := updater.getLatestRelease()
	if err != nil {
		t.Fatal(err)
	}

	asset, ok := lo.Find(release.Assets, func(asset ReleaseAsset) bool {
		return strings.HasSuffix(asset.BrowserDownloadUrl, "Windows_x86_64.zip")
	})
	if !ok {
		t.Fatal("could not find release asset")
	}

	folderPath, err := updater.DownloadLatestRelease(asset.BrowserDownloadUrl, repoPath)
	if err != nil {
		t.Log("Downloaded to:", folderPath)
		t.Fatal(err)
	}

	t.Log("Downloaded to:", folderPath)
}
