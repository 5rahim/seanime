package updater

import (
	"seanime/internal/constants"
	"seanime/internal/events"
	"seanime/internal/util"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdater_getReleaseName(t *testing.T) {

	updater := Updater{}

	t.Log(updater.GetReleaseName(constants.Version))
}

func TestUpdater_FetchLatestRelease(t *testing.T) {

	fallbackGithubUrl = "https://seanimedud.app/api/releases" // simulate dead endpoint
	//githubUrl = "https://api.github.com/repos/zbonfo/seanime-desktop/releases/latest"

	updater := New(constants.Version, util.NewLogger(), events.NewMockWSEventManager(util.NewLogger()))
	release, err := updater.fetchLatestRelease()
	if err != nil {
		t.Fatal(err)
	}

	if assert.NotNil(t, release) {
		spew.Dump(release)
	}
}

func TestUpdater_FetchLatestReleaseFromDocs(t *testing.T) {

	updater := New(constants.Version, util.NewLogger(), events.NewMockWSEventManager(util.NewLogger()))
	release, err := updater.fetchLatestReleaseFromDocs()
	if err != nil {
		t.Fatal(err)
	}

	if assert.NotNil(t, release) {
		spew.Dump(release)
	}
}

func TestUpdater_FetchLatestReleaseFromGitHub(t *testing.T) {

	updater := New(constants.Version, util.NewLogger(), events.NewMockWSEventManager(util.NewLogger()))
	release, err := updater.fetchLatestReleaseFromGitHub()
	if err != nil {
		t.Fatal(err)
	}

	if assert.NotNil(t, release) {
		spew.Dump(release)
	}
}

func TestUpdater_CompareVersion(t *testing.T) {

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
			currVersion:   "2.2.0-prerelease",
			latestVersion: "2.2.0",
			shouldUpdate:  true,
		},
		{
			currVersion:   "2.2.0",
			latestVersion: "2.2.0-prerelease",
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
		{
			currVersion:   "1.0.0",
			latestVersion: "0.2.1",
			shouldUpdate:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.latestVersion, func(t *testing.T) {
			updateType, shouldUpdate := util.CompareVersion(tt.currVersion, tt.latestVersion)
			assert.Equal(t, tt.shouldUpdate, shouldUpdate)
			t.Log(tt.latestVersion, updateType)
		})
	}

}

func TestUpdater(t *testing.T) {

	u := New(constants.Version, util.NewLogger(), events.NewMockWSEventManager(util.NewLogger()))

	rl, err := u.GetLatestRelease()
	require.NoError(t, err)

	rl.TagName = "v2.2.1"
	newV := strings.TrimPrefix(rl.TagName, "v")
	updateTypeI, shouldUpdate := util.CompareVersion(u.CurrentVersion, newV)
	isOlder := util.VersionIsOlderThan(u.CurrentVersion, newV)

	util.Spew(isOlder)
	util.Spew(shouldUpdate)
	util.Spew(updateTypeI)
}
