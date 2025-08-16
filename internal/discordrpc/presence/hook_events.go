package discordrpc_presence

import (
	discordrpc_client "seanime/internal/discordrpc/client"
	"seanime/internal/hook_resolver"
)

// DiscordPresenceAnimeActivityRequestedEvent is triggered when anime activity is requested, after the [animeActivity] is processed, and right before the activity is sent to queue.
// There is no guarantee as to when or if the activity will be successfully sent to discord.
// Note that this event is triggered every 6 seconds or so, avoid heavy processing or perform it only when the activity is changed.
// Prevent default to stop the activity from being sent to discord.
type DiscordPresenceAnimeActivityRequestedEvent struct {
	hook_resolver.Event
	// Anime activity object used to generate the activity
	AnimeActivity *AnimeActivity `json:"animeActivity"`

	// Name of the activity
	Name string `json:"name"`
	// Details of the activity
	Details    string `json:"details"`
	DetailsURL string `json:"detailsUrl"`
	// State of the activity
	State string `json:"state"`
	// Timestamps of the activity
	StartTimestamp *int64 `json:"startTimestamp"`
	EndTimestamp   *int64 `json:"endTimestamp"`

	// Assets of the activity
	LargeImage string `json:"largeImage"`
	LargeText  string `json:"largeText"`
	LargeURL   string `json:"largeUrl,omitempty"` // URL to large image, if any
	SmallImage string `json:"smallImage"`
	SmallText  string `json:"smallText"`
	SmallURL   string `json:"smallUrl,omitempty"` // URL to small image, if any

	// Buttons of the activity
	Buttons []*discordrpc_client.Button `json:"buttons"`

	// Whether the activity is an instance
	Instance bool `json:"instance"`
	// Type of the activity
	Type int `json:"type"`
	// StatusDisplayType controls formatting
	StatusDisplayType int `json:"statusDisplayType,omitempty"`
}

// DiscordPresenceMangaActivityRequestedEvent is triggered when manga activity is requested, after the [mangaActivity] is processed, and right before the activity is sent to queue.
// There is no guarantee as to when or if the activity will be successfully sent to discord.
// Note that this event is triggered every 6 seconds or so, avoid heavy processing or perform it only when the activity is changed.
// Prevent default to stop the activity from being sent to discord.
type DiscordPresenceMangaActivityRequestedEvent struct {
	hook_resolver.Event
	// Manga activity object used to generate the activity
	MangaActivity *MangaActivity `json:"mangaActivity"`

	// Name of the activity
	Name string `json:"name"`
	// Details of the activity
	Details    string `json:"details"`
	DetailsURL string `json:"detailsUrl"`
	// State of the activity
	State string `json:"state"`
	// Timestamps of the activity
	StartTimestamp *int64 `json:"startTimestamp"`
	EndTimestamp   *int64 `json:"endTimestamp"`

	// Assets of the activity
	LargeImage string `json:"largeImage"`
	LargeText  string `json:"largeText"`
	LargeURL   string `json:"largeUrl,omitempty"` // URL to large image, if any
	SmallImage string `json:"smallImage"`
	SmallText  string `json:"smallText"`
	SmallURL   string `json:"smallUrl,omitempty"` // URL to small image, if any

	// Buttons of the activity
	Buttons []*discordrpc_client.Button `json:"buttons"`

	// Whether the activity is an instance
	Instance bool `json:"instance"`
	// Type of the activity
	Type int `json:"type"`
	// StatusDisplayType controls formatting
	StatusDisplayType int `json:"statusDisplayType,omitempty"`
}

// DiscordPresenceClientClosedEvent is triggered when the discord rpc client is closed.
type DiscordPresenceClientClosedEvent struct {
	hook_resolver.Event
}
