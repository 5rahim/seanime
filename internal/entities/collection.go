package entities

import (
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/anizip"
	"github.com/sourcegraph/conc/pool"
	"path/filepath"
	"slices"
	"sort"
)

const (
	LibraryCollectionEntryCurrent   LibraryCollectionListType = "current"
	LibraryCollectionEntryPlanned   LibraryCollectionListType = "planned"
	LibraryCollectionEntryCompleted LibraryCollectionListType = "completed"
	LibraryCollectionEntryPaused    LibraryCollectionListType = "paused"
	LibraryCollectionEntryDropped   LibraryCollectionListType = "dropped"
)

type (
	LibraryCollection struct {
		ContinueWatchingList []*MediaEntryEpisode     `json:"continueWatchingList"`
		Lists                []*LibraryCollectionList `json:"lists"`
		UnmatchedLocalFiles  []*LocalFile             `json:"unmatchedLocalFiles"`
		IgnoredLocalFiles    []*LocalFile             `json:"ignoredLocalFiles"`
		UnmatchedGroups      []*UnmatchedGroup        `json:"unmatchedGroups"`
	}
	LibraryCollectionListType string

	LibraryCollectionList struct {
		Type    LibraryCollectionListType `json:"type"`
		Status  anilist.MediaListStatus   `json:"status"`
		Entries []*LibraryCollectionEntry `json:"entries"`
	}

	LibraryCollectionEntry struct {
		Media                 *anilist.BaseMedia     `json:"media"`
		MediaId               int                    `json:"mediaId"`
		MediaEntryLibraryData *MediaEntryLibraryData `json:"libraryData"` // Library data
		MediaEntryListData    *MediaEntryListData    `json:"listData"`    // AniList list data
	}

	UnmatchedGroup struct {
		Dir         string                `json:"dir"`
		LocalFiles  []*LocalFile          `json:"localFiles"`
		Suggestions []*anilist.BasicMedia `json:"suggestions"`
	}
	NewLibraryCollectionOptions struct {
		AnilistCollection *anilist.AnimeCollection
		LocalFiles        []*LocalFile
		AnizipCache       *anizip.Cache
		AnilistClient     *anilist.Client
	}
)

// NewLibraryCollection creates a new LibraryCollection.
// A LibraryCollection consists of a list of LibraryCollectionList (one for each status).
func NewLibraryCollection(opts *NewLibraryCollectionOptions) *LibraryCollection {

	// Get lists from collection
	aniLists := opts.AnilistCollection.GetMediaListCollection().GetLists()

	lc := new(LibraryCollection)

	// Create lists
	lc.hydrateCollectionLists(
		opts.LocalFiles,
		aniLists,
	)

	// Add Continue Watching list
	lc.hydrateContinueWatchingList(
		opts.LocalFiles,
		opts.AnilistCollection,
		opts.AnizipCache,
		opts.AnilistClient,
	)

	lc.hydrateRest(opts.LocalFiles)

	lc.hydrateUnmatchedGroups()

	return lc

}

//----------------------------------------------------------------------------------------------------------------------

func (lc *LibraryCollection) hydrateCollectionLists(
	localFiles []*LocalFile,
	aniLists []*anilist.AnimeCollection_MediaListCollection_Lists,
) {

	// Group local files by media id
	groupedLfs := GroupLocalFilesByMediaID(localFiles)
	// Get slice of media ids from local files
	mIds := GetMediaIdsFromLocalFiles(localFiles)

	// Create a new LibraryCollectionList for each list
	// This is done in parallel
	p := pool.NewWithResults[*LibraryCollectionList]()
	for _, list := range aniLists {
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

						entryLfs, _ := groupedLfs[entry.Media.ID]
						libraryData, _ := NewMediaEntryLibraryData(&NewMediaEntryLibraryDataOptions{
							entryLocalFiles: entryLfs,
							mediaId:         entry.Media.ID,
						})

						return &LibraryCollectionEntry{
							MediaId:               entry.Media.ID,
							Media:                 entry.Media,
							MediaEntryLibraryData: libraryData,
							MediaEntryListData: &MediaEntryListData{
								Progress:    *entry.Progress,
								Score:       *entry.Score,
								Status:      entry.Status,
								StartedAt:   anilist.ToEntryStartDate(entry.StartedAt),
								CompletedAt: anilist.ToEntryCompletionDate(entry.CompletedAt),
							},
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
			sort.Slice(r, func(i, j int) bool {
				return r[i].Media.GetTitleSafe() < r[j].Media.GetTitleSafe()
			})

			// Return a new LibraryEntries struct
			return &LibraryCollectionList{
				Type:    getLibraryCollectionEntryFromListStatus(*list.Status),
				Status:  *list.Status,
				Entries: r,
			}

		})
	}

	lists := p.Wait()

	// Merge repeating to current
	repeat, ok := lo.Find(lists, func(item *LibraryCollectionList) bool {
		return item.Status == anilist.MediaListStatusRepeating
	})
	if ok {
		current, ok := lo.Find(lists, func(item *LibraryCollectionList) bool {
			return item.Status == anilist.MediaListStatusCurrent
		})
		if len(repeat.Entries) > 0 && ok {
			current.Entries = append(current.Entries, repeat.Entries...)
		}
		// Remove repeating from lists
		lists = lo.Filter(lists, func(item *LibraryCollectionList, index int) bool {
			return item.Status != anilist.MediaListStatusRepeating
		})
	}

	// Lists
	lc.Lists = lists
}

//----------------------------------------------------------------------------------------------------------------------

// hydrateContinueWatchingList creates a list for "continue watching".
// This should be called after the lists have been created.
func (lc *LibraryCollection) hydrateContinueWatchingList(
	localFiles []*LocalFile,
	anilistCollection *anilist.AnimeCollection,
	anizipCache *anizip.Cache,
	anilistClient *anilist.Client,
) {

	// Create media entry for media in "Current" list
	current, found := lo.Find(lc.Lists, func(item *LibraryCollectionList) bool {
		return item.Status == anilist.MediaListStatusCurrent
	})
	if !found {
		lc.ContinueWatchingList = make([]*MediaEntryEpisode, 0) // Return empty slice
		return
	}
	mIds := make([]int, len(current.Entries))
	for i, entry := range current.Entries {
		mIds[i] = entry.MediaId
	}

	mEntryPool := pool.NewWithResults[*MediaEntry]()
	for _, mId := range mIds {
		mId := mId
		mEntryPool.Go(func() *MediaEntry {
			me, _ := NewMediaEntry(&NewMediaEntryOptions{
				MediaId:           mId,
				LocalFiles:        localFiles,
				AnilistCollection: anilistCollection,
				AnizipCache:       anizipCache,
				AnilistClient:     anilistClient,
			})
			return me
		})
	}
	mEntries := mEntryPool.Wait()
	mEntries = lo.Filter(mEntries, func(item *MediaEntry, index int) bool {
		return item != nil
	})

	if len(mEntries) == 0 {
		lc.ContinueWatchingList = make([]*MediaEntryEpisode, 0) // Return empty slice
		return
	}

	// Sort by progress
	sort.Slice(mEntries, func(i, j int) bool {
		return mEntries[i].MediaEntryListData.Progress > mEntries[j].MediaEntryListData.Progress
	})

	// Remove entries whose user's progress is equal to the latest episode's progress number, meaning the user has watched the latest episode
	mEntries = lop.Map(mEntries, func(mEntry *MediaEntry, index int) *MediaEntry {
		if !mEntry.HasWatchedAll() {
			return mEntry
		}
		return nil
	})
	mEntries = lo.Filter(mEntries, func(item *MediaEntry, index int) bool {
		return item != nil
	})

	// Get the next episode for each media entry
	mEpisodes := lop.Map(mEntries, func(mEntry *MediaEntry, index int) *MediaEntryEpisode {
		ep, ok := mEntry.FindNextEpisode()
		if ok {
			return ep
		}
		return nil
	})
	mEpisodes = lo.Filter(mEpisodes, func(item *MediaEntryEpisode, index int) bool {
		return item != nil
	})

	lc.ContinueWatchingList = mEpisodes

	return

}

//----------------------------------------------------------------------------------------------------------------------

func (lc *LibraryCollection) hydrateUnmatchedGroups() {

	groups := make([]*UnmatchedGroup, 0)

	// Group by directory
	groupedLfs := lop.GroupBy(lc.UnmatchedLocalFiles, func(lf *LocalFile) string {
		return filepath.Dir(lf.GetPath())
	})

	for key, value := range groupedLfs {
		groups = append(groups, &UnmatchedGroup{
			Dir:         key,
			LocalFiles:  value,
			Suggestions: make([]*anilist.BasicMedia, 0),
		})
	}

	lc.UnmatchedGroups = groups
}

//----------------------------------------------------------------------------------------------------------------------

func (lc *LibraryCollection) hydrateRest(localFiles []*LocalFile) {

	lc.UnmatchedLocalFiles = lo.Filter(localFiles, func(lf *LocalFile, index int) bool {
		return lf.MediaId == 0
	})

	lc.IgnoredLocalFiles = lo.Filter(localFiles, func(lf *LocalFile, index int) bool {
		return lf.Ignored == true
	})

}

//----------------------------------------------------------------------------------------------------------------------

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
