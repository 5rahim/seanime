package manga

import (
	"cmp"
	"fmt"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/sourcegraph/conc/pool"
	"slices"
)

const (
	CollectionEntryCurrent   CollectionStatusType = "current"
	CollectionEntryPlanned   CollectionStatusType = "planned"
	CollectionEntryCompleted CollectionStatusType = "completed"
	CollectionEntryPaused    CollectionStatusType = "paused"
	CollectionEntryDropped   CollectionStatusType = "dropped"
)

type (
	CollectionStatusType string

	Collection struct {
		Lists []*CollectionList `json:"lists"`
	}

	CollectionList struct {
		Type    CollectionStatusType    `json:"type"`
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
		MangaCollection      *anilist.MangaCollection
		AnilistClientWrapper anilist.ClientWrapperInterface
	}
)

func NewCollection(opts *NewCollectionOptions) (collection *Collection, err error) {
	coll := &Collection{}
	if opts.MangaCollection == nil {
		return nil, fmt.Errorf("MangaCollection is nil")
	}
	if opts.AnilistClientWrapper == nil {
		return nil, fmt.Errorf("AnilistClientWrapper is nil")
	}

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
						MediaId: entry.GetID(),
						EntryListData: &EntryListData{
							Progress:    *entry.Progress,
							Score:       *entry.Score,
							Status:      entry.Status,
							StartedAt:   anilist.ToEntryDate(entry.StartedAt),
							CompletedAt: anilist.ToEntryDate(entry.CompletedAt),
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

	return coll, nil
}

func getCollectionEntryFromListStatus(st anilist.MediaListStatus) CollectionStatusType {
	switch st {
	case anilist.MediaListStatusCurrent:
		return CollectionEntryCurrent
	case anilist.MediaListStatusRepeating:
		return CollectionEntryCurrent
	case anilist.MediaListStatusPlanning:
		return CollectionEntryPlanned
	case anilist.MediaListStatusCompleted:
		return CollectionEntryCompleted
	case anilist.MediaListStatusPaused:
		return CollectionEntryPaused
	case anilist.MediaListStatusDropped:
		return CollectionEntryDropped
	default:
		return CollectionEntryCurrent
	}
}
