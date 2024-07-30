package platform

import (
	"seanime/internal/library/anime"
	"seanime/internal/manga"
)

type (
	MediaSnapshot struct {
		MediaID       string         `json:"mediaId"`
		AnimeSnapshot *AnimeSnapshot `json:"animeSnapshot,omitempty"`
		MangaSnapshot *MangaSnapshot `json:"mangaSnapshot,omitempty"`
	}

	ImageSnapshot struct {
		CoverImage    string            `json:"coverImage"`
		BannerImage   string            `json:"bannerImage"`
		EpisodeImages map[string]string `json:"episodeImages"` // Local filepath to image file
	}

	AnimeSnapshot struct {
		Entry  *anime.AnimeEntry `json:"entry"`
		Images *ImageSnapshot    `json:"images"`
	}

	MangaSnapshot struct {
		Entry  *manga.Entry   `json:"entry"`
		Images *ImageSnapshot `json:"images"`
	}
)
