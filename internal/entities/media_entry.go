package entities

import (
	"errors"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
)

type (
	MediaEntry struct {
		MediaId           int                `json:"mediaId"`
		Media             *anilist.BaseMedia `json:"media"`
		MediaEntryDetails *MediaEntryDetails `json:"listEntry"`

		LocalFiles  []*LocalFile `json:"localFiles"`
		EpisodeList struct {
			Spotlight []*MediaEntryEpisode `json:"spotlight"`
			All       []*MediaEntryEpisode `json:"all"`
		} `json:"episodeList"`

		AnizipData *anizip.Media `json:"anizipData"`
	}

	MediaEntryDetails struct {
		Progress    int                      `json:"progress,omitempty"`
		Score       float64                  `json:"score,omitempty"`
		Status      *anilist.MediaListStatus `json:"status,omitempty"`
		StartedAt   string                   `json:"startedAt,omitempty"`
		CompletedAt string                   `json:"completedAt,omitempty"`
	}

	MediaEntryEpisode struct {
		Title        string `json:"title"`
		EpisodeTitle string `json:"episodeTitle"`
		Number       int    `json:"number"`
	}

	NewMediaEntryOptions struct {
		MediaId           int
		LocalFiles        []*LocalFile
		AnizipCache       *anizip.Cache
		AnilistCollection *anilist.AnimeCollection
	}
)

func NewMediaEntry(opts *NewMediaEntryOptions) (*MediaEntry, error) {

	if opts.AnilistCollection == nil ||
		opts.LocalFiles == nil ||
		opts.AnizipCache == nil {
		return nil, errors.New("missing arguments when creating media entry")
	}

	anilistEntry, found := opts.AnilistCollection.GetListEntryFromMediaId(opts.MediaId)

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

	// If the media exist in the user's anime list, add the details
	if found {
		entry.MediaEntryDetails = &MediaEntryDetails{
			Progress:    *anilistEntry.Progress,
			Score:       *anilistEntry.Score,
			Status:      anilistEntry.Status,
			StartedAt:   anilist.ToEntryStartDate(anilistEntry.StartedAt),
			CompletedAt: anilist.ToEntryCompletionDate(anilistEntry.CompletedAt),
		}
	}

	return entry, nil

}
