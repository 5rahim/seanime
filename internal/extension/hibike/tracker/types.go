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
	}

	UserInfo struct {
		Username  string `json:"username"`
		AvatarURL string `json:"avatarUrl,omitempty"`
	}

	MediaEntry struct {
		// Source is "anilist" for AniList entries, or the custom source name for custom sources
		Source string `json:"source"`
		// MediaId is the AniList or custom source media ID (internal ID)
		MediaId int `json:"mediaId"`
		// ExternalId is the tracker's own media ID (e.g., MAL ID: "12345", Kitsu ID: "54321")
		// This should be populated by the extension's ID mapping logic
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

		PushEntry(ctx context.Context, entry *MediaEntry) error

		PushEntries(ctx context.Context, entries []*MediaEntry) (map[int]error, error)

		PullEntry(ctx context.Context, mediaId int) (*MediaEntry, error)

		PullEntries(ctx context.Context) ([]*MediaEntry, error)

		DeleteEntry(ctx context.Context, mediaId int) error

		IsLoggedIn(ctx context.Context) bool

		GetUserInfo(ctx context.Context) (*UserInfo, error)

		TestConnection(ctx context.Context) error

		ResolveMediaId(ctx context.Context, mediaId int, mediaType string) (string, error)
	}
)
