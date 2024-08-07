package anime

import (
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/api/metadata"
	"github.com/seanime-app/seanime/internal/util"
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
	// LibraryCollection holds the main data for the library collection.
	// It consists of:
	//  - ContinueWatchingList: a list of MediaEntryEpisode for the "continue watching" feature.
	//  - Lists: a list of LibraryCollectionList (one for each status).
	//  - UnmatchedLocalFiles: a list of unmatched local files (Media id == 0). "Resolve unmatched" feature.
	//  - UnmatchedGroups: a list of UnmatchedGroup instances. Like UnmatchedLocalFiles, but grouped by directory. "Resolve unmatched" feature.
	//  - IgnoredLocalFiles: a list of ignored local files. (DEVNOTE: Unused for now)
	//  - UnknownGroups: a list of UnknownGroup instances. Group of files whose media is not in the user's AniList "Resolve unknown media" feature.
	LibraryCollection struct {
		ContinueWatchingList []*MediaEntryEpisode     `json:"continueWatchingList"`
		Lists                []*LibraryCollectionList `json:"lists"`
		UnmatchedLocalFiles  []*LocalFile             `json:"unmatchedLocalFiles"`
		UnmatchedGroups      []*UnmatchedGroup        `json:"unmatchedGroups"`
		IgnoredLocalFiles    []*LocalFile             `json:"ignoredLocalFiles"`
		UnknownGroups        []*UnknownGroup          `json:"unknownGroups"`
		Stats                *LibraryCollectionStats  `json:"stats"`
	}

	LibraryCollectionListType string

	LibraryCollectionStats struct {
		TotalEntries  int    `json:"totalEntries"`
		TotalFiles    int    `json:"totalFiles"`
		TotalShows    int    `json:"totalShows"`
		TotalMovies   int    `json:"totalMovies"`
		TotalSpecials int    `json:"totalSpecials"`
		TotalSize     string `json:"totalSize"`
	}

	LibraryCollectionList struct {
		Type    LibraryCollectionListType `json:"type"`
		Status  anilist.MediaListStatus   `json:"status"`
		Entries []*LibraryCollectionEntry `json:"entries"`
	}

	// LibraryCollectionEntry holds the data for a single entry in a LibraryCollectionList.
	// It is a slimmed down version of MediaEntry. It holds the media, media id, library data, and list data.
	LibraryCollectionEntry struct {
		Media                 *anilist.BaseMedia     `json:"media"`
		MediaId               int                    `json:"mediaId"`
		MediaEntryLibraryData *MediaEntryLibraryData `json:"libraryData"` // Library data
		MediaEntryListData    *MediaEntryListData    `json:"listData"`    // AniList list data
	}

	// UnmatchedGroup holds the data for a group of unmatched local files.
	UnmatchedGroup struct {
		Dir         string               `json:"dir"`
		LocalFiles  []*LocalFile         `json:"localFiles"`
		Suggestions []*anilist.BaseMedia `json:"suggestions"`
	}
	// UnknownGroup holds the data for a group of local files whose media is not in the user's AniList.
	// The client will use this data to suggest media to the user, so they can add it to their AniList.
	UnknownGroup struct {
		MediaId    int          `json:"mediaId"`
		LocalFiles []*LocalFile `json:"localFiles"`
	}
)

type (
	// NewLibraryCollectionOptions is a struct that holds the data needed for creating a new LibraryCollection.
	NewLibraryCollectionOptions struct {
		AnimeCollection      *anilist.AnimeCollection
		LocalFiles           []*LocalFile
		AnizipCache          *anizip.Cache
		AnilistClientWrapper anilist.ClientWrapperInterface
		MetadataProvider     *metadata.Provider
	}
)

// NewLibraryCollection creates a new LibraryCollection.
func NewLibraryCollection(opts *NewLibraryCollectionOptions) (lc *LibraryCollection, err error) {

	defer util.HandlePanicInModuleWithError("entities/collection/NewLibraryCollection", &err)

	// Get lists from collection
	aniLists := opts.AnimeCollection.GetMediaListCollection().GetLists()

	lc = new(LibraryCollection)

	// Create lists
	lc.hydrateCollectionLists(
		opts.LocalFiles,
		aniLists,
	)

	lc.hydrateStats(opts.LocalFiles)

	// Add Continue Watching list
	lc.hydrateContinueWatchingList(
		opts.LocalFiles,
		opts.AnimeCollection,
		opts.AnizipCache,
		opts.AnilistClientWrapper,
		opts.MetadataProvider,
	)

	lc.UnmatchedLocalFiles = lo.Filter(opts.LocalFiles, func(lf *LocalFile, index int) bool {
		return lf.MediaId == 0
	})

	lc.IgnoredLocalFiles = lo.Filter(opts.LocalFiles, func(lf *LocalFile, index int) bool {
		return lf.Ignored == true
	})

	lc.hydrateUnmatchedGroups()

	return
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
	foundIds := make([]int, 0)

	for _, list := range aniLists {
		entries := list.GetEntries()
		for _, entry := range entries {
			foundIds = append(foundIds, entry.Media.ID)
		}
	}

	// Create a new LibraryCollectionList for each list
	// This is done in parallel
	p := pool.NewWithResults[*LibraryCollectionList]()
	for _, list := range aniLists {
		p.Go(func() *LibraryCollectionList {
			// If the list has no status, return nil
			// This occurs when there are custom lists (DEVNOTE: This shouldn't occur because we remove custom lists when the collection is fetched)
			if list.Status == nil {
				return nil
			}

			// For each list, get the entries
			entries := list.GetEntries()

			// For each entry, check if the media id is in the local files
			// If it is, create a new LibraryCollectionEntry with the associated local files
			p2 := pool.NewWithResults[*LibraryCollectionEntry]()
			for _, entry := range entries {
				p2.Go(func() *LibraryCollectionEntry {
					if slices.Contains(mIds, entry.Media.ID) {

						entryLfs, _ := groupedLfs[entry.Media.ID]
						libraryData, _ := NewMediaEntryLibraryData(&NewMediaEntryLibraryDataOptions{
							EntryLocalFiles: entryLfs,
							MediaId:         entry.Media.ID,
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
			// Sort by title
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

	// Get the lists from the pool
	lists := p.Wait()
	// Filter out nil entries
	lists = lo.Filter(lists, func(item *LibraryCollectionList, index int) bool {
		return item != nil
	})

	// Merge repeating to current (no need to show repeating as a separate list)
	repeatingList, ok := lo.Find(lists, func(item *LibraryCollectionList) bool {
		return item.Status == anilist.MediaListStatusRepeating
	})
	if ok {
		currentList, ok := lo.Find(lists, func(item *LibraryCollectionList) bool {
			return item.Status == anilist.MediaListStatusCurrent
		})
		if len(repeatingList.Entries) > 0 && ok {
			currentList.Entries = append(currentList.Entries, repeatingList.Entries...)
		} else if len(repeatingList.Entries) > 0 {
			newCurrentList := repeatingList
			newCurrentList.Type = LibraryCollectionEntryCurrent
			lists = append(lists, newCurrentList)
		}
		// Remove repeating from lists
		lists = lo.Filter(lists, func(item *LibraryCollectionList, index int) bool {
			return item.Status != anilist.MediaListStatusRepeating
		})
	}

	// Lists
	lc.Lists = lists

	if lc.Lists == nil {
		lc.Lists = make([]*LibraryCollectionList, 0)
	}

	// +---------------------+
	// |  Unknown media ids  |
	// +---------------------+

	unknownIds := make([]int, 0)
	for _, id := range mIds {
		if id != 0 && !slices.Contains(foundIds, id) {
			unknownIds = append(unknownIds, id)
		}
	}

	lc.UnknownGroups = make([]*UnknownGroup, 0)
	for _, id := range unknownIds {
		lc.UnknownGroups = append(lc.UnknownGroups, &UnknownGroup{
			MediaId:    id,
			LocalFiles: groupedLfs[id],
		})
	}

	return
}

//----------------------------------------------------------------------------------------------------------------------

func (lc *LibraryCollection) hydrateStats(lfs []*LocalFile) {
	stats := &LibraryCollectionStats{
		TotalFiles:    len(lfs),
		TotalEntries:  0,
		TotalShows:    0,
		TotalMovies:   0,
		TotalSpecials: 0,
		TotalSize:     "",
	}

	for _, list := range lc.Lists {
		for _, entry := range list.Entries {
			stats.TotalEntries++
			if entry.Media.Format != nil {
				if *entry.Media.Format == anilist.MediaFormatMovie {
					stats.TotalMovies++
				} else if *entry.Media.Format == anilist.MediaFormatSpecial || *entry.Media.Format == anilist.MediaFormatOva {
					stats.TotalSpecials++
				} else {
					stats.TotalShows++
				}
			}
		}
	}

	lc.Stats = stats
}

//----------------------------------------------------------------------------------------------------------------------

// hydrateContinueWatchingList creates a list of MediaEntryEpisode for the "continue watching" feature.
// This should be called after the LibraryCollectionList's have been created.
func (lc *LibraryCollection) hydrateContinueWatchingList(
	localFiles []*LocalFile,
	animeCollection *anilist.AnimeCollection,
	anizipCache *anizip.Cache,
	anilistClientWrapper anilist.ClientWrapperInterface,
	metadataProvider *metadata.Provider,
) {

	// Get currently watching list
	current, found := lo.Find(lc.Lists, func(item *LibraryCollectionList) bool {
		return item.Status == anilist.MediaListStatusCurrent
	})

	// If no currently watching list is found, return an empty slice
	if !found {
		lc.ContinueWatchingList = make([]*MediaEntryEpisode, 0) // Set empty slice
		return
	}
	// Get media ids from current list
	mIds := make([]int, len(current.Entries))
	for i, entry := range current.Entries {
		mIds[i] = entry.MediaId
	}

	// Create a new MediaEntry for each media id
	mEntryPool := pool.NewWithResults[*MediaEntry]()
	for _, mId := range mIds {
		mEntryPool.Go(func() *MediaEntry {
			me, _ := NewMediaEntry(&NewMediaEntryOptions{
				MediaId:              mId,
				LocalFiles:           localFiles,
				AnimeCollection:      animeCollection,
				AnizipCache:          anizipCache,
				AnilistClientWrapper: anilistClientWrapper,
				MetadataProvider:     metadataProvider,
			})
			return me
		})
	}
	mEntries := mEntryPool.Wait()
	mEntries = lo.Filter(mEntries, func(item *MediaEntry, index int) bool {
		return item != nil
	}) // Filter out nil entries

	// If there are no entries, return an empty slice
	if len(mEntries) == 0 {
		lc.ContinueWatchingList = make([]*MediaEntryEpisode, 0) // Return empty slice
		return
	}

	// Sort by progress
	sort.Slice(mEntries, func(i, j int) bool {
		return mEntries[i].MediaEntryListData.Progress > mEntries[j].MediaEntryListData.Progress
	})

	// Remove entries the user has watched all episodes of
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

// hydrateUnmatchedGroups is a method of the LibraryCollection struct.
// It is responsible for grouping unmatched local files by their directory and creating UnmatchedGroup instances for each group.
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
			Suggestions: make([]*anilist.BaseMedia, 0),
		})
	}

	// Assign the created groups
	lc.UnmatchedGroups = groups

	return
}

//----------------------------------------------------------------------------------------------------------------------

// getLibraryCollectionEntryFromListStatus maps anilist.MediaListStatus to LibraryCollectionListType.
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
