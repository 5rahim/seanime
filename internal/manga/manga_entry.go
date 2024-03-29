package manga

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/util/filecache"
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
		MediaId              int
		Logger               *zerolog.Logger
		FileCacher           *filecache.Cacher
		MangaCollection      *anilist.MangaCollection
		AnilistClientWrapper anilist.ClientWrapperInterface
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
		mediaF, err := opts.AnilistClientWrapper.BaseMangaByID(context.Background(), &opts.MediaId)
		if err != nil {
			return nil, err
		}
		entry.Media = mediaF.GetMedia()
	} else {
		// If the entry is found, we use the entry from the collection.
		entry.Media = anilistEntry.GetMedia()
		entry.EntryListData = &EntryListData{
			Progress:    *anilistEntry.Progress,
			Score:       *anilistEntry.Score,
			Status:      anilistEntry.Status,
			StartedAt:   anilist.ToEntryDate(anilistEntry.StartedAt),
			CompletedAt: anilist.ToEntryDate(anilistEntry.CompletedAt),
		}
	}

	return entry, nil
}
