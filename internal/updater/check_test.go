package updater

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdater_getReleaseName(t *testing.T) {

	updater := Updater{}

	t.Log(updater.getReleaseName("0.2.2"))
}

func TestUpdater_FetchLatestRelease(t *testing.T) {

	updater := Updater{}
	release, err := updater.fetchLatestRelease()
	if err != nil {
		t.Fatal(err)
	}

	if assert.NotNil(t, release) {
		spew.Dump(release)
	}

}

func TestUpdater_CompareVersion(t *testing.T) {

	updater := Updater{}

	tests := []struct {
		currVersion   string
		latestVersion string
		shouldUpdate  bool
	}{
		{
			currVersion:   "0.2.2",
			latestVersion: "0.2.2",
			shouldUpdate:  false,
		},
		{
			currVersion:   "0.2.2",
			latestVersion: "0.2.3",
			shouldUpdate:  true,
		},
		{
			currVersion:   "0.2.2",
			latestVersion: "0.3.0",
			shouldUpdate:  true,
		},
		{
			currVersion:   "0.2.2",
			latestVersion: "1.0.0",
			shouldUpdate:  true,
		},
		{
			currVersion:   "0.2.2",
			latestVersion: "0.2.1",
			shouldUpdate:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.latestVersion, func(t *testing.T) {
			updateType, shouldUpdate := updater.compareVersion(tt.currVersion, tt.latestVersion)
			assert.Equal(t, tt.shouldUpdate, shouldUpdate)
			t.Log(tt.latestVersion, updateType)
		})
	}

}
