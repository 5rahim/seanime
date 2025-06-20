package manga

import (
	"cmp"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/hook"
	"seanime/internal/platforms/platform"
	"slices"

	"github.com/samber/lo"
	"github.com/sourcegraph/conc/pool"
)

type (
	CollectionStatusType string

	Collection struct {
		Lists []*CollectionList `json:"lists"`
	}

	CollectionList struct {
		Type    anilist.MediaListStatus `json:"type"`
		Status  anilist.MediaListStatus `json:"status"`
		Entries []*CollectionEntry      `json:"entries"`
	}

	CollectionEntry struct {
		Media         *anilist.BaseManga `json:"media"`
		MediaId       int                `json:"mediaId"`
		EntryListData *EntryListData     `json:"listData"` // AniList list data
	}
)

type (
	NewCollectionOptions struct {
		MangaCollection *anilist.MangaCollection
		Platform        platform.Platform
	}
)

func NewCollection(opts *NewCollectionOptions) (collection *Collection, err error) {
	coll := &Collection{}
	if opts.MangaCollection == nil {
		return nil, nil
	}
	if opts.Platform == nil {
		return nil, fmt.Errorf("platform is nil")
	}

	optsEvent := new(MangaLibraryCollectionRequestedEvent)
	optsEvent.MangaCollection = opts.MangaCollection
	err = hook.GlobalHookManager.OnMangaLibraryCollectionRequested().Trigger(optsEvent)
	if err != nil {
		return nil, err
	}
	opts.MangaCollection = optsEvent.MangaCollection

	aniLists := opts.MangaCollection.GetMediaListCollection().GetLists()

	aniLists = lo.Filter(aniLists, func(list *anilist.MangaList, _ int) bool {
		return list.Status != nil
	})

	p := pool.NewWithResults[*CollectionList]()
	for _, list := range aniLists {
		p.Go(func() *CollectionList {

			if list.Status == nil {
				return nil
			}

			entries := list.GetEntries()

			p2 := pool.NewWithResults[*CollectionEntry]()
			for _, entry := range entries {
				p2.Go(func() *CollectionEntry {

					return &CollectionEntry{
						Media:   entry.GetMedia(),
						MediaId: entry.GetMedia().GetID(),
						EntryListData: &EntryListData{
							Progress:    *entry.Progress,
							Score:       *entry.Score,
							Status:      entry.Status,
							Repeat:      entry.GetRepeatSafe(),
							StartedAt:   anilist.FuzzyDateToString(entry.StartedAt),
							CompletedAt: anilist.FuzzyDateToString(entry.CompletedAt),
						},
					}
				})
			}

			collectionEntries := p2.Wait()

			slices.SortFunc(collectionEntries, func(i, j *CollectionEntry) int {
				return cmp.Compare(i.Media.GetTitleSafe(), j.Media.GetTitleSafe())
			})

			return &CollectionList{
				Type:    getCollectionEntryFromListStatus(*list.Status),
				Status:  *list.Status,
				Entries: collectionEntries,
			}

		})
	}
	lists := p.Wait()

	lists = lo.Filter(lists, func(l *CollectionList, _ int) bool {
		return l != nil
	})

	// Merge repeating to current (no need to show repeating as a separate list)
	repeat, ok := lo.Find(lists, func(item *CollectionList) bool {
		return item.Status == anilist.MediaListStatusRepeating
	})
	if ok {
		current, ok := lo.Find(lists, func(item *CollectionList) bool {
			return item.Status == anilist.MediaListStatusCurrent
		})
		if len(repeat.Entries) > 0 && ok {
			current.Entries = append(current.Entries, repeat.Entries...)
		}
		// Remove repeating from lists
		lists = lo.Filter(lists, func(item *CollectionList, index int) bool {
			return item.Status != anilist.MediaListStatusRepeating
		})
	}

	coll.Lists = lists

	event := new(MangaLibraryCollectionEvent)
	event.LibraryCollection = coll
	_ = hook.GlobalHookManager.OnMangaLibraryCollection().Trigger(event)
	coll = event.LibraryCollection

	return coll, nil
}

func getCollectionEntryFromListStatus(st anilist.MediaListStatus) anilist.MediaListStatus {
	if st == anilist.MediaListStatusRepeating {
		return anilist.MediaListStatusCurrent
	}

	return st
}
