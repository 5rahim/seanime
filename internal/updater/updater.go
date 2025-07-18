package updater

import (
	"net/http"
	"seanime/internal/events"
	"seanime/internal/util"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
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
		client              *http.Client
		wsEventManager      mo.Option[events.WSEventManagerInterface]
		announcements       []Announcement
	}

	Update struct {
		Release        *Release `json:"release,omitempty"`
		CurrentVersion string   `json:"current_version,omitempty"`
		Type           string   `json:"type"`
	}
)

func New(currVersion string, logger *zerolog.Logger, wsEventManager events.WSEventManagerInterface) *Updater {
	ret := &Updater{
		CurrentVersion:      currVersion,
		hasCheckedForUpdate: false,
		checkForUpdate:      true,
		logger:              logger,
		client: &http.Client{
			Timeout: time.Second * 10,
		},
		wsEventManager: mo.None[events.WSEventManagerInterface](),
	}

	if wsEventManager != nil {
		ret.wsEventManager = mo.Some[events.WSEventManagerInterface](wsEventManager)
	}

	return ret
}

func (u *Updater) GetLatestUpdate() (*Update, error) {
	if !u.checkForUpdate {
		return nil, nil
	}

	rl, err := u.GetLatestRelease()
	if err != nil {
		return nil, err
	}

	if rl == nil || rl.TagName == "" {
		return nil, nil
	}

	if !rl.Released {
		return nil, nil
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
		Release:        rl,
		CurrentVersion: u.CurrentVersion,
		Type:           updateType,
	}, nil
}

func (u *Updater) ShouldRefetchReleases() {
	u.hasCheckedForUpdate = false

	if u.wsEventManager.IsPresent() {
		// Tell the client to send a request to fetch the latest release
		u.wsEventManager.MustGet().SendEvent(events.CheckForUpdates, nil)
	}
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
