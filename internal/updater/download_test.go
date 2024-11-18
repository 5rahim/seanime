package updater

import (
	"github.com/samber/lo"
	"os"
	"seanime/internal/util"
	"strings"
	"testing"
)

func TestUpdater_DownloadLatestRelease(t *testing.T) {

	updater := New("0.2.0", util.NewLogger(), nil)

	//tempDir := "E:\\SEANIME-REPO-TEST"
	tempDir := t.TempDir()

	// Get the latest release
	release, err := updater.GetLatestRelease()
	if err != nil {
		t.Fatal(err)
	}

	// Find the asset (zip file)
	asset, ok := lo.Find(release.Assets, func(asset ReleaseAsset) bool {
		return strings.HasSuffix(asset.BrowserDownloadUrl, "Windows_x86_64.zip")
	})
	if !ok {
		t.Fatal("could not find release asset")
	}

	// Download the asset
	folderPath, err := updater.DownloadLatestRelease(asset.BrowserDownloadUrl, tempDir)
	if err != nil {
		t.Log("Downloaded to:", folderPath)
		t.Fatal(err)
	}

	t.Log("Downloaded to:", folderPath)

	// Check if the folder is not empty
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) == 0 {
		t.Fatal("folder is empty")
	}

	for _, entry := range entries {
		t.Log(entry.Name())
	}

	// Delete the folder
	if err := os.RemoveAll(folderPath); err != nil {
		t.Fatal(err)
	}

	// Find the asset (.tar.gz file)
	asset2, ok := lo.Find(release.Assets, func(asset ReleaseAsset) bool {
		return strings.HasSuffix(asset.BrowserDownloadUrl, "MacOS_arm64.tar.gz")
	})
	if !ok {
		t.Fatal("could not find release asset")
	}

	// Download the asset
	folderPath2, err := updater.DownloadLatestRelease(asset2.BrowserDownloadUrl, tempDir)
	if err != nil {
		t.Log("Downloaded to:", folderPath2)
		t.Fatal(err)
	}

	t.Log("Downloaded to:", folderPath2)

	// Check if the folder is not empty
	entries2, err := os.ReadDir(folderPath2)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries2) == 0 {
		t.Fatal("folder is empty")
	}

	for _, entry := range entries2 {
		t.Log(entry.Name())
	}
}
