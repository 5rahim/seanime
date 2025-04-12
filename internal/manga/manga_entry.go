package manga

import (
	"errors"
	"seanime/internal/api/anilist"
	"seanime/internal/hook"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/platforms/platform"
	"seanime/internal/util/filecache"

	"github.com/rs/zerolog"
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
		Repeat      int                      `json:"repeat,omitempty"`
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

	reqEvent := new(MangaEntryRequestedEvent)
	reqEvent.MediaId = opts.MediaId
	reqEvent.MangaCollection = opts.MangaCollection
	reqEvent.Entry = entry

	err = hook.GlobalHookManager.OnMangaEntryRequested().Trigger(reqEvent)
	if err != nil {
		return nil, err
	}
	opts.MediaId = reqEvent.MediaId                 // Override the media ID
	opts.MangaCollection = reqEvent.MangaCollection // Override the manga collection
	entry = reqEvent.Entry                          // Override the entry

	if reqEvent.DefaultPrevented {
		mangaEvent := new(MangaEntryEvent)
		mangaEvent.Entry = reqEvent.Entry
		err = hook.GlobalHookManager.OnMangaEntry().Trigger(mangaEvent)
		if err != nil {
			return nil, err
		}

		if mangaEvent.Entry == nil {
			return nil, errors.New("no entry was returned")
		}
		return mangaEvent.Entry, nil
	}

	anilistEntry, found := opts.MangaCollection.GetListEntryFromMangaId(opts.MediaId)

	// If the entry is not found, we fetch the manga from the Anilist API.
	if !found {
		media, err := opts.Platform.GetManga(opts.MediaId)
		if err != nil {
			return nil, err
		}
		entry.Media = media

	} else {
		// If the entry is found, we use the entry from the collection.
		mangaEvent := new(anilist_platform.GetMangaEvent)
		mangaEvent.Manga = anilistEntry.GetMedia()
		err := hook.GlobalHookManager.OnGetManga().Trigger(mangaEvent)
		if err != nil {
			return nil, err
		}
		entry.Media = mangaEvent.Manga
		entry.EntryListData = &EntryListData{
			Progress:    *anilistEntry.Progress,
			Score:       *anilistEntry.Score,
			Status:      anilistEntry.Status,
			Repeat:      anilistEntry.GetRepeatSafe(),
			StartedAt:   anilist.FuzzyDateToString(anilistEntry.StartedAt),
			CompletedAt: anilist.FuzzyDateToString(anilistEntry.CompletedAt),
		}
	}

	mangaEvent := new(MangaEntryEvent)
	mangaEvent.Entry = entry
	err = hook.GlobalHookManager.OnMangaEntry().Trigger(mangaEvent)
	if err != nil {
		return nil, err
	}

	return mangaEvent.Entry, nil
}
