package updater

import (
	"strconv"
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
		CheckForUpdate      bool
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

func New(currVersion string) *Updater {
	return &Updater{
		CurrentVersion:      currVersion,
		hasCheckedForUpdate: false,
		CheckForUpdate:      true,
	}
}

func (u *Updater) GetLatestUpdate() (*Update, error) {
	if !u.CheckForUpdate {
		return nil, nil
	}

	rl, err := u.getLatestRelease()
	if err != nil {
		return nil, err
	}

	updateType, shouldUpdate := u.compareVersion(u.CurrentVersion, rl.TagName)
	if !shouldUpdate {
		return nil, nil
	}

	return &Update{
		Release: rl,
		Type:    updateType,
	}, nil
}

func (u *Updater) ShouldRefetchReleases() {
	u.hasCheckedForUpdate = false
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// GetLatestRelease returns the latest release from the GitHub repository.
func (u *Updater) getLatestRelease() (*Release, error) {
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

// compareVersion compares current and latest version is returns true if the latest version is newer than the current version.
// It also returns the update type (patch, minor, major) if the latest version is newer than the current version.
func (u *Updater) compareVersion(currVersion string, tagName string) (string, bool) {
	tagName = strings.TrimPrefix(tagName, "v")

	currVParts := strings.Split(currVersion, ".")
	latestVParts := strings.Split(tagName, ".")

	if len(currVParts) != 3 || len(latestVParts) != 3 {
		return "", false
	}

	currMajor, _ := strconv.Atoi(currVParts[0])
	currMinor, _ := strconv.Atoi(currVParts[1])
	currPatch, _ := strconv.Atoi(currVParts[2])

	latestMajor, _ := strconv.Atoi(latestVParts[0])
	latestMinor, _ := strconv.Atoi(latestVParts[1])
	latestPatch, _ := strconv.Atoi(latestVParts[2])

	if currMajor > latestMajor {
		return "", false
	}

	if currMajor < latestMajor {
		return MajorRelease, true
	}

	if currMinor > latestMinor {
		return "", false
	}

	if currMinor < latestMinor {
		return MinorRelease, true
	}

	if currPatch > latestPatch {
		return "", false
	}

	if currPatch < latestPatch {
		return PatchRelease, true
	}

	return "", false
}
