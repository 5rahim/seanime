package playbackmanager

import "seanime/internal/library/anime"

type (
	StreamMagnetRequestOptions struct {
		MagnetLink      string `json:"magnet_link"`               // magnet link to stream
		OptionalMediaId int    `json:"optionalMediaId,omitempty"` // optional media ID to associate with the magnet link
		Untracked       bool   `json:"untracked"`
	}

	// TrackedStreamMagnetRequestResponse is returned after analysis of the magnet link
	TrackedStreamMagnetRequestResponse struct {
		EpisodeNumber     int                      `json:"episodeNumber"` // episode number of the magnet link
		EpisodeCollection *anime.EpisodeCollection `json:"episodeCollection"`
	}

	TrackedStreamMagnetOptions struct {
		EpisodeNumber int    `json:"episodeNumber"`
		AniDBEpisode  string `json:"anidbEpisode"`
	}
)
