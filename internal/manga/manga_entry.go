package manga

import (
	"github.com/rs/zerolog"
	"seanime/internal/api/anilist"
	"seanime/internal/platform"
	"seanime/internal/util/filecache"
)

type (
	// Entry is fetched when the user goes to the manga entry page.
	Entry struct {
		MediaId       int                `json:"mediaId"`
		Media         *anilist.BaseManga `json:"media"`
		EntryListData *EntryListData     `json:"listData,omitempty"`
	}

	EntryListData struct {
		Progress    int                      `json:"progress,omitempty"`
		Score       float64                  `json:"score,omitempty"`
		Status      *anilist.MediaListStatus `json:"status,omitempty"`
		StartedAt   string                   `json:"startedAt,omitempty"`
		CompletedAt string                   `json:"completedAt,omitempty"`
	}
)

type (
	// NewEntryOptions is the options for creating a new manga entry.
	NewEntryOptions struct {
		MediaId         int
		Logger          *zerolog.Logger
		FileCacher      *filecache.Cacher
		MangaCollection *anilist.MangaCollection
		Platform        platform.Platform
	}
)

// NewEntry creates a new manga entry.
func NewEntry(opts *NewEntryOptions) (entry *Entry, err error) {
	entry = &Entry{
		MediaId: opts.MediaId,
	}

	anilistEntry, found := opts.MangaCollection.GetListEntryFromMediaId(opts.MediaId)

	// If the entry is not found, we fetch the manga from the Anilist API.
	if !found {
		media, err := opts.Platform.GetManga(opts.MediaId)
		if err != nil {
			return nil, err
		}
		entry.Media = media
	} else {
		// If the entry is found, we use the entry from the collection.
		entry.Media = anilistEntry.GetMedia()
		entry.EntryListData = &EntryListData{
			Progress:    *anilistEntry.Progress,
			Score:       *anilistEntry.Score,
			Status:      anilistEntry.Status,
			StartedAt:   anilist.FuzzyDateToString(anilistEntry.StartedAt),
			CompletedAt: anilist.FuzzyDateToString(anilistEntry.CompletedAt),
		}
	}

	return entry, nil
}
