package updater

import (
	"io"
	"net/http"
	"runtime"
	"seanime/internal/constants"
	"seanime/internal/database/models"
	"seanime/internal/events"
	"slices"

	"github.com/Masterminds/semver/v3"
	"github.com/goccy/go-json"
)

type AnnouncementType string

const (
	AnnouncementTypeToast  AnnouncementType = "toast"
	AnnouncementTypeDialog AnnouncementType = "dialog"
	AnnouncementTypeBanner AnnouncementType = "banner"
)

type AnnouncementSeverity string

const (
	AnnouncementSeverityInfo     AnnouncementSeverity = "info"
	AnnouncementSeverityWarning  AnnouncementSeverity = "warning"
	AnnouncementSeverityError    AnnouncementSeverity = "error"
	AnnouncementSeverityCritical AnnouncementSeverity = "critical"
)

type AnnouncementAction struct {
	Label string `json:"label"`
	URL   string `json:"url"`
	Type  string `json:"type"`
}

type AnnouncementConditions struct {
	OS       []string `json:"os,omitempty"`       // ["windows", "darwin", "linux"]
	Platform []string `json:"platform,omitempty"` // ["tauri", "web", "denshi"]
	// FeatureFlags      []string `json:"featureFlags,omitempty"`      // Required feature flags
	VersionConstraint string   `json:"versionConstraint,omitempty"` // e.g. "<= 2.9.0", "2.9.0"
	UserSettingsPath  string   `json:"userSettingsPath,omitempty"`  // JSON path to check in user settings
	UserSettingsValue []string `json:"userSettingsValue,omitempty"` // Expected values at that path
}

type Announcement struct {
	ID       string               `json:"id"`              // Unique identifier for tracking
	Title    string               `json:"title,omitempty"` // Title for dialogs/banners
	Message  string               `json:"message"`         // The message to display
	Type     AnnouncementType     `json:"type"`            // The type of announcement
	Severity AnnouncementSeverity `json:"severity"`        // Severity level
	Date     interface{}          `json:"date"`            // Date of the announcement

	NotDismissible bool `json:"notDismissible"` // Can user dismiss it

	Conditions *AnnouncementConditions `json:"conditions,omitempty"` // Advanced targeting

	Actions []AnnouncementAction `json:"actions,omitempty"` // Action buttons

	Priority int `json:"priority"`
}

func (u *Updater) GetAnnouncements(version string, platform string, settings *models.Settings) []Announcement {
	var filteredAnnouncements []Announcement
	if !u.checkForUpdate {
		return filteredAnnouncements
	}
	// filter out
	for _, announcement := range u.announcements {
		if announcement.Conditions == nil {
			filteredAnnouncements = append(filteredAnnouncements, announcement)
			continue
		}

		conditions := announcement.Conditions

		if len(conditions.OS) > 0 && !slices.Contains(conditions.OS, runtime.GOOS) {
			continue
		}

		if conditions.Platform != nil && !slices.Contains(conditions.Platform, platform) {
			continue
		}

		if conditions.VersionConstraint != "" {
			versionConstraint, err := semver.NewConstraint(conditions.VersionConstraint)
			if err != nil {
				u.logger.Error().Err(err).Msgf("updater: Failed to parse version constraint")
				continue
			}

			currVersion, err := semver.NewVersion(version)
			if err != nil {
				u.logger.Error().Err(err).Msgf("updater: Failed to parse current version")
				continue
			}

			if !versionConstraint.Check(currVersion) {
				continue
			}

		}

		filteredAnnouncements = append(filteredAnnouncements, announcement)
	}

	u.announcements = filteredAnnouncements

	return u.announcements
}

func (u *Updater) FetchAnnouncements() []Announcement {
	var announcements []Announcement

	response, err := http.Get(constants.AnnouncementURL)
	if err != nil {
		u.logger.Error().Err(err).Msgf("updater: Failed to get announcements")
		return announcements
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		u.logger.Error().Err(err).Msgf("updater: Failed to read announcements")
		return announcements
	}

	err = json.Unmarshal(body, &announcements)
	if err != nil {
		u.logger.Error().Err(err).Msgf("updater: Failed to unmarshal announcements")
		return announcements
	}

	// Filter out announcements
	var filteredAnnouncements []Announcement
	for _, announcement := range announcements {
		if announcement.Conditions == nil {
			filteredAnnouncements = append(filteredAnnouncements, announcement)
			continue
		}

		conditions := announcement.Conditions

		if len(conditions.OS) > 0 && !slices.Contains(conditions.OS, runtime.GOOS) {
			continue
		}

		filteredAnnouncements = append(filteredAnnouncements, announcement)
	}

	u.announcements = announcements

	if u.wsEventManager.IsPresent() {
		// Tell the client to send a request to fetch the latest announcements
		u.wsEventManager.MustGet().SendEvent(events.CheckForAnnouncements, nil)
	}

	return announcements
}
