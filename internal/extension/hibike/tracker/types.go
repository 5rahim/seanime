package hibiketracker

import (
	"context"
	"seanime/internal/api/anilist"
	"time"
)

type (
	Settings struct {
		// SupportsAnime indicates if this tracker supports anime tracking
		SupportsAnime bool `json:"supportsAnime"`
		// SupportsManga indicates if this tracker supports manga tracking
		SupportsManga bool `json:"supportsManga"`
		// SupportsBidirectionalSync indicates if Seanime can pull data from the tracker
		SupportsBidirectionalSync bool `json:"supportsBidirectionalSync"`
		MaxRequestsPerSecond      int  `json:"maxRequestsPerSecond,omitempty"`
		CacheVersion              int  `json:"cacheVersion,omitempty"`
	}

	UserInfo struct {
		Username  string `json:"username"`
		AvatarURL string `json:"avatarUrl,omitempty"`
	}

	MediaEntry struct {
		// Source is "anilist" for AniList entries, or the custom source ID for custom sources.
		//	e.g. For "One Piece" on AniList, this would be "anilist".
		//		For "Plur1bus" on the SIMKL custom source, this would be "simkl".
		// When Pulling: The extension MUST populate this.
		Source string `json:"source"`
		// MediaId is the AniList or custom source media ID (Seanime ID).
		//	When Pulling: The extension could leave this empty.
		MediaId int `json:"mediaId"`
		// MalId is the MyAnimeList ID (if available)
		MalId *int `json:"malId,omitempty"`
		// ExternalId
		//	When Pulling: The extension MUST populate this.
		//	When Pushing: Seanime will populate this (using ResolveExternalId).
		ExternalId string `json:"externalId"`
		// MediaType is either "ANIME" or "MANGA"
		MediaType string `json:"mediaType"`
		// Status is the watching/reading status
		Status *anilist.MediaListStatus `json:"status,omitempty"`
		// Score is 0-100 scale (extensions should convert to their service's scale)
		Score *int `json:"score,omitempty"`
		// Progress is the number of episodes/chapters consumed
		Progress *int `json:"progress,omitempty"`
		// Repeat is the number of times rewatched/reread (applies to both anime and manga)
		Repeat *int `json:"repeat,omitempty"`
		// StartedAt is when the user started watching/reading
		StartedAt *anilist.FuzzyDateInput `json:"startedAt,omitempty"`
		// CompletedAt is when the user completed the entry
		CompletedAt *anilist.FuzzyDateInput `json:"completedAt,omitempty"`
		// UpdatedAt is used for conflict resolution in bidirectional sync
		UpdatedAt *time.Time `json:"updatedAt,omitempty"`
	}

	Provider interface {
		GetSettings() Settings

		// PushEntry updates the given entry on the tracker.
		PushEntry(ctx context.Context, entry *MediaEntry) error

		// PullEntries returns all entries from the tracker.
		PullEntries(ctx context.Context) ([]*MediaEntry, error)

		// DeleteEntry deletes the entry with from the tracker.
		DeleteEntry(ctx context.Context, mediaId int) error

		IsLoggedIn(ctx context.Context) bool

		GetUserInfo(ctx context.Context) (*UserInfo, bool)

		TestConnection(ctx context.Context) error

		// ResolveExternalId finds the tracker-specific ID for a given Seanime media entry.
		// Seanime calls this BEFORE calling PushEntry to ensure MediaEntry.ExternalId is populated.
		// The ID is cached for future calls, in case it were to change, you can invalidate the cache by changing the value of Settings.CacheVersion.
		ResolveExternalId(ctx context.Context, entry *MediaEntry) (string, error)

		//
		ResolveReverseMapping(ctx context.Context, externalId string) (*MediaEntry, error)
	}
)

type DiffType string

const (
	DiffTypeNone         DiffType = "none"
	DiffTypeLocalOnly    DiffType = "local-only"    // Exists in Seanime, missing in Tracker
	DiffTypeRemoteOnly   DiffType = "remote-only"   // Exists in Tracker, missing in Seanime
	DiffTypeDivergent    DiffType = "divergent"     // Exists in both, but values differ
	DiffTypeMappingError DiffType = "mapping-error" // Cannot resolve ID
)

type Action string

const (
	ActionPush   Action = "push"
	ActionPull   Action = "pull"
	ActionIgnore Action = "ignore"
)

type SyncDiff struct {
	MediaID        int    // Seanime ID
	ExternalID     string // Tracker ID
	Type           DiffType
	Local          *MediaEntry
	Remote         *MediaEntry
	ProposedAction Action
}
