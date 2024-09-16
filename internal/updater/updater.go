package updater

import (
	"github.com/rs/zerolog"
	"seanime/internal/util"
	"strings"
)

const (
	PatchRelease = "patch"
	MinorRelease = "minor"
	MajorRelease = "major"
)

type (
	Updater struct {
		CurrentVersion      string
		hasCheckedForUpdate bool
		LatestRelease       *Release
		checkForUpdate      bool
		logger              *zerolog.Logger
	}

	Update struct {
		Release *Release `json:"release,omitempty"`
		Type    string   `json:"type"`
	}

	LatestReleaseResponse struct {
		Release Release `json:"release"`
	}
	Release struct {
		Url         string         `json:"url"`
		HtmlUrl     string         `json:"html_url"`
		NodeId      string         `json:"node_id"`
		TagName     string         `json:"tag_name"`
		Name        string         `json:"name"`
		Body        string         `json:"body"`
		PublishedAt string         `json:"published_at"`
		Released    bool           `json:"released"`
		Version     string         `json:"version"`
		Assets      []ReleaseAsset `json:"assets"`
	}
	ReleaseAsset struct {
		Url                string `json:"url"`
		Id                 int    `json:"id"`
		NodeId             string `json:"node_id"`
		Name               string `json:"name"`
		ContentType        string `json:"content_type"`
		Uploaded           bool   `json:"uploaded"`
		Size               int    `json:"size"`
		BrowserDownloadUrl string `json:"browser_download_url"`
	}
)

func New(currVersion string, logger *zerolog.Logger) *Updater {
	return &Updater{
		CurrentVersion:      currVersion,
		hasCheckedForUpdate: false,
		checkForUpdate:      true,
		logger:              logger,
	}
}

func (u *Updater) GetLatestUpdate() (*Update, error) {
	if !u.checkForUpdate {
		return nil, nil
	}

	rl, err := u.GetLatestRelease()
	if err != nil {
		return nil, err
	}

	newV := strings.TrimPrefix(rl.TagName, "v")
	updateTypeI, shouldUpdate := util.CompareVersion(u.CurrentVersion, newV)
	if !shouldUpdate {
		return nil, nil
	}

	updateType := ""
	if updateTypeI == -1 {
		updateType = MinorRelease
	} else if updateTypeI == -2 {
		updateType = PatchRelease
	} else if updateTypeI == -3 {
		updateType = MajorRelease
	}

	return &Update{
		Release: rl,
		Type:    updateType,
	}, nil
}

func (u *Updater) ShouldRefetchReleases() {
	u.hasCheckedForUpdate = false
}

func (u *Updater) SetEnabled(checkForUpdate bool) {
	u.checkForUpdate = checkForUpdate
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// GetLatestRelease returns the latest release from the GitHub repository.
func (u *Updater) GetLatestRelease() (*Release, error) {
	if u.hasCheckedForUpdate {
		return u.LatestRelease, nil
	}

	release, err := u.fetchLatestRelease()
	if err != nil {
		return nil, err
	}

	u.hasCheckedForUpdate = true
	u.LatestRelease = release
	return release, nil
}
