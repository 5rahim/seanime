package entities

import (
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/sourcegraph/conc/pool"
	"slices"
)

type LibraryEntryType string

const (
	LibraryEntryTypeCurrent   LibraryEntryType = "current"
	LibraryEntryTypePlanned   LibraryEntryType = "planned"
	LibraryEntryTypeCompleted LibraryEntryType = "completed"
	LibraryEntryTypePaused    LibraryEntryType = "paused"
	LibraryEntryTypeDropped   LibraryEntryType = "dropped"
)

type LibraryEntry struct {
	Type    LibraryEntryType        `json:"type"`
	Status  anilist.MediaListStatus `json:"status"`
	Entries []*Entry                `json:"current"`
}

type Entry struct {
	Media          *anilist.BaseMedia `json:"media"`
	MediaId        int                `json:"mediaId"`
	Progress       int                `json:"progress,omitempty"`
	Score          float64            `json:"score,omitempty"`
	AllFilesLocked bool               `json:"allFilesLocked"`
}

type NewLibraryEntriesOptions struct {
	Collection *anilist.AnimeCollection
	LocalFiles []*LocalFile
}

func NewLibraryEntries(opts *NewLibraryEntriesOptions) []*LibraryEntry {

	// Group local files by media id
	groupedLfs := lop.GroupBy(opts.LocalFiles, func(item *LocalFile) int {
		return item.MediaId
	})

	// Get slice of media ids from local files
	mIds := make([]int, len(groupedLfs))
	for key := range groupedLfs {
		if !slices.Contains(mIds, key) {
			mIds = append(mIds, key)
		}
	}

	// Get lists from collection
	lists := opts.Collection.GetMediaListCollection().GetLists()

	// We will create a new LibraryEntries struct containing the entries for each list
	// This is done in parallel
	p := pool.NewWithResults[*LibraryEntry]()
	for _, list := range lists {
		list := list
		p.Go(func() *LibraryEntry {

			// For each list, get the entries
			entries := list.GetEntries()

			p2 := pool.NewWithResults[*Entry]()
			// For each entry, check if the media id is in the local files
			// If it is, create a new Entry
			for _, entry := range entries {
				entry := entry
				p2.Go(func() *Entry {
					if slices.Contains(mIds, entry.Media.ID) {
						lfs := groupedLfs[entry.Media.ID]
						return &Entry{
							MediaId:        entry.Media.ID,
							Media:          entry.Media,
							Progress:       *entry.Progress,
							Score:          *entry.Score,
							AllFilesLocked: lo.EveryBy(lfs, func(item *LocalFile) bool { return item.Locked }),
						}
					} else {
						return nil
					}
				})
			}

			r := p2.Wait()
			// Filter out nil entries
			r = lo.Filter(r, func(item *Entry, index int) bool {
				return item != nil
			})

			// Return a new LibraryEntries struct
			return &LibraryEntry{
				Type:    getLibraryEntryTypeFromListStatus(*list.Status),
				Status:  *list.Status,
				Entries: r,
			}

		})
	}

	res := p.Wait()

	// Merge repeating to current
	repeat, ok := lo.Find(res, func(item *LibraryEntry) bool {
		return item.Status == anilist.MediaListStatusRepeating
	})
	if ok {
		current, ok := lo.Find(res, func(item *LibraryEntry) bool {
			return item.Status == anilist.MediaListStatusCurrent
		})
		if len(repeat.Entries) > 0 && ok {
			current.Entries = append(current.Entries, repeat.Entries...)
		}
		// Remove repeating from res
		res = lo.Filter(res, func(item *LibraryEntry, index int) bool {
			return item.Status != anilist.MediaListStatusRepeating
		})
	}

	return res

}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func getLibraryEntryTypeFromListStatus(st anilist.MediaListStatus) LibraryEntryType {
	switch st {
	case anilist.MediaListStatusCurrent:
		return LibraryEntryTypeCurrent
	case anilist.MediaListStatusRepeating:
		return LibraryEntryTypeCurrent
	case anilist.MediaListStatusPlanning:
		return LibraryEntryTypePlanned
	case anilist.MediaListStatusCompleted:
		return LibraryEntryTypeCompleted
	case anilist.MediaListStatusPaused:
		return LibraryEntryTypePaused
	case anilist.MediaListStatusDropped:
		return LibraryEntryTypeDropped
	default:
		return LibraryEntryTypeCurrent
	}
}
