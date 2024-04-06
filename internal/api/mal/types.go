package mal

import "time"

type (
	RequestOptions struct {
		AccessToken  string
		RefreshToken string
		ExpiresAt    time.Time
	}

	MediaType       string
	MediaStatus     string
	MediaListStatus string
)

const (
	MediaTypeTV                    MediaType       = "tv"      // Anime
	MediaTypeOVA                   MediaType       = "ova"     // Anime
	MediaTypeMovie                 MediaType       = "movie"   // Anime
	MediaTypeSpecial               MediaType       = "special" // Anime
	MediaTypeONA                   MediaType       = "ona"     // Anime
	MediaTypeMusic                 MediaType       = "music"
	MediaTypeManga                 MediaType       = "manga"                // Manga
	MediaTypeNovel                 MediaType       = "novel"                // Manga
	MediaTypeOneShot               MediaType       = "oneshot"              // Manga
	MediaStatusFinishedAiring      MediaStatus     = "finished_airing"      // Anime
	MediaStatusCurrentlyAiring     MediaStatus     = "currently_airing"     // Anime
	MediaStatusNotYetAired         MediaStatus     = "not_yet_aired"        // Anime
	MediaStatusFinished            MediaStatus     = "finished"             // Manga
	MediaStatusCurrentlyPublishing MediaStatus     = "currently_publishing" // Manga
	MediaStatusNotYetPublished     MediaStatus     = "not_yet_published"    // Manga
	MediaListStatusReading         MediaListStatus = "reading"              // Manga
	MediaListStatusWatching        MediaListStatus = "watching"             // Anime
	MediaListStatusCompleted       MediaListStatus = "completed"
	MediaListStatusOnHold          MediaListStatus = "on_hold"
	MediaListStatusDropped         MediaListStatus = "dropped"
	MediaListStatusPlanToWatch     MediaListStatus = "plan_to_watch" // Anime
	MediaListStatusPlanToRead      MediaListStatus = "plan_to_read"  // Manga
)
