package entities

import (
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
)

type (
	MediaEntry struct {
		MediaId int                `json:"mediaId"`
		Media   *anilist.BaseMedia `json:"media"`

		LocalFiles  []*LocalFile `json:"localFiles"`
		EpisodeList struct {
			Spotlight []*MediaEntryEpisode `json:"spotlight"`
			All       []*MediaEntryEpisode `json:"all"`
		} `json:"episodeList"`

		AnizipData *anizip.Media `json:"anizipData"`

		Progress int                      `json:"progress,omitempty"`
		Score    float64                  `json:"score,omitempty"`
		Status   *anilist.MediaListStatus `json:"status,omitempty"`
	}

	MediaEntryEpisode struct {
		Title        string `json:"title"`
		EpisodeTitle string `json:"episodeTitle"`
		Number       int    `json:"number"`
	}

	NewMediaEntryOptions struct {
		MediaId     int
		LocalFiles  []*LocalFile
		AnizipCache *anizip.Cache
	}
)

func NewMediaEntry(opts *NewMediaEntryOptions) (*MediaEntry, error) {

	entry := new(MediaEntry)
	entry.MediaId = opts.MediaId

	// Get the entry's local files
	lfs := GetLocalFilesFromMediaId(opts.LocalFiles, opts.MediaId)
	entry.LocalFiles = lfs

	// Fetch AniDB data and cache it for 10 minutes
	anidb, err := anizip.FetchAniZipMediaC("anilist", opts.MediaId, opts.AnizipCache)
	if err != nil {
		return nil, err
	}
	entry.AnizipData = anidb

	return entry, nil

}
