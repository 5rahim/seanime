package entities

import (
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/sourcegraph/conc/pool"
	"slices"
)

type LibraryCollectionListType string

const (
	LibraryCollectionEntryCurrent   LibraryCollectionListType = "current"
	LibraryCollectionEntryPlanned   LibraryCollectionListType = "planned"
	LibraryCollectionEntryCompleted LibraryCollectionListType = "completed"
	LibraryCollectionEntryPaused    LibraryCollectionListType = "paused"
	LibraryCollectionEntryDropped   LibraryCollectionListType = "dropped"
)

type (
	LibraryCollectionList struct {
		Type    LibraryCollectionListType `json:"type"`
		Status  anilist.MediaListStatus   `json:"status"`
		Entries []*LibraryCollectionEntry `json:"current"`
	}
	LibraryCollectionEntry struct {
		Media          *anilist.BaseMedia `json:"media"`
		MediaId        int                `json:"mediaId"`
		Progress       int                `json:"progress,omitempty"`
		Score          float64            `json:"score,omitempty"`
		AllFilesLocked bool               `json:"allFilesLocked"`
	}
	NewLibraryCollectionOptions struct {
		Collection *anilist.AnimeCollection
		LocalFiles []*LocalFile
	}
)

func NewLibraryCollection(opts *NewLibraryCollectionOptions) []*LibraryCollectionList {

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

	// Create a new LibraryCollectionList for each list
	// This is done in parallel
	p := pool.NewWithResults[*LibraryCollectionList]()
	for _, list := range lists {
		list := list
		p.Go(func() *LibraryCollectionList {

			// For each list, get the entries
			entries := list.GetEntries()

			p2 := pool.NewWithResults[*LibraryCollectionEntry]()
			// For each entry, check if the media id is in the local files
			// If it is, create a new LibraryCollectionEntry
			for _, entry := range entries {
				entry := entry
				p2.Go(func() *LibraryCollectionEntry {
					if slices.Contains(mIds, entry.Media.ID) {
						lfs := groupedLfs[entry.Media.ID]
						return &LibraryCollectionEntry{
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
			r = lo.Filter(r, func(item *LibraryCollectionEntry, index int) bool {
				return item != nil
			})

			// Return a new LibraryEntries struct
			return &LibraryCollectionList{
				Type:    getLibraryCollectionEntryFromListStatus(*list.Status),
				Status:  *list.Status,
				Entries: r,
			}

		})
	}

	res := p.Wait()

	// Merge repeating to current
	repeat, ok := lo.Find(res, func(item *LibraryCollectionList) bool {
		return item.Status == anilist.MediaListStatusRepeating
	})
	if ok {
		current, ok := lo.Find(res, func(item *LibraryCollectionList) bool {
			return item.Status == anilist.MediaListStatusCurrent
		})
		if len(repeat.Entries) > 0 && ok {
			current.Entries = append(current.Entries, repeat.Entries...)
		}
		// Remove repeating from res
		res = lo.Filter(res, func(item *LibraryCollectionList, index int) bool {
			return item.Status != anilist.MediaListStatusRepeating
		})
	}

	return res

}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func getLibraryCollectionEntryFromListStatus(st anilist.MediaListStatus) LibraryCollectionListType {
	switch st {
	case anilist.MediaListStatusCurrent:
		return LibraryCollectionEntryCurrent
	case anilist.MediaListStatusRepeating:
		return LibraryCollectionEntryCurrent
	case anilist.MediaListStatusPlanning:
		return LibraryCollectionEntryPlanned
	case anilist.MediaListStatusCompleted:
		return LibraryCollectionEntryCompleted
	case anilist.MediaListStatusPaused:
		return LibraryCollectionEntryPaused
	case anilist.MediaListStatusDropped:
		return LibraryCollectionEntryDropped
	default:
		return LibraryCollectionEntryCurrent
	}
}
