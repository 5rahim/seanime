package entities

import (
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/sourcegraph/conc/pool"
	"slices"
)

type LibraryEntries struct {
	Current   []*Entry `json:"current"` // "current" and "repeating"
	Paused    []*Entry `json:"paused"`
	Planned   []*Entry `json:"planned"`
	Completed []*Entry `json:"completed"`
	Dropped   []*Entry `json:"dropped"`
}

type Entry struct {
	LocalFiles []*LocalFile            `json:"localFiles"`
	Media      *anilist.BaseMedia      `json:"media"`
	Progress   int                     `json:"progress,omitempty"`
	Status     anilist.MediaListStatus `json:"status,omitempty"`
	Score      float64                 `json:"score,omitempty"`
}

type NewLibraryEntriesOptions struct {
	Collection *anilist.AnimeCollection
	LocalFiles []*LocalFile
}

func NewLibraryEntries(opts *NewLibraryEntriesOptions) []*LibraryEntries {

	lists := opts.Collection.GetMediaListCollection().GetLists()

	groupedLfs := lop.GroupBy(opts.LocalFiles, func(item *LocalFile) int {
		return item.MediaId
	})

	mIds := make([]int, len(groupedLfs))
	for key := range groupedLfs {
		if !slices.Contains(mIds, key) {
			mIds = append(mIds, key)
		}
	}
	p := pool.NewWithResults[*LibraryEntries]()
	for _, list := range lists {
		list := list
		p.Go(func() *LibraryEntries {

			entries := list.GetEntries()
			p2 := pool.NewWithResults[*Entry]()
			for _, entry := range entries {
				entry := entry
				p2.Go(func() *Entry {
					if slices.Contains(mIds, entry.Media.ID) {
						return &Entry{
							//LocalFiles: groupedLfs[entry.Media.ID],
							Media:    entry.Media,
							Progress: *entry.Progress,
							Status:   *entry.Status,
							Score:    *entry.Score,
						}
					} else {
						return nil
					}
				})
			}

			r := p2.Wait()
			r = lo.Filter(r, func(item *Entry, index int) bool {
				return item != nil
			})

			groupedEntries := lop.GroupBy(r, func(item *Entry) string {
				return item.Status.String()
			})

			return &LibraryEntries{
				Current:   lo.Interleave(groupedEntries[anilist.MediaListStatusCurrent.String()], groupedEntries[anilist.MediaListStatusRepeating.String()]),
				Paused:    groupedEntries[anilist.MediaListStatusPaused.String()],
				Planned:   groupedEntries[anilist.MediaListStatusPlanning.String()],
				Completed: groupedEntries[anilist.MediaListStatusCompleted.String()],
				Dropped:   groupedEntries[anilist.MediaListStatusDropped.String()],
			}

		})
	}

	res := p.Wait()

	return res

}
