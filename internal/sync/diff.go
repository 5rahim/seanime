package sync

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"seanime/internal/api/anilist"
	hibikemanga "seanime/internal/extension/hibike/manga"
	"seanime/internal/library/anime"
	"seanime/internal/manga"
	"slices"
	"strings"
)

// DEVNOTE: Here we compare the media data from the current up-to-date collections with the local data.
// Outdated media are added to the Syncer to be updated.
// If the media doesn't have a snapshot -> a new snapshot is created.
// If the reference key is different -> the metadata is re-fetched and the snapshot is updated.
// If the list data is different -> the list data is updated.

const (
	DiffTypeMissing  DiffType = iota // We need to add a new snapshot
	DiffTypeMetadata                 // We need to re-fetch the snapshot metadata (episode metadata / chapter containers), list data will be updated as well
	DiffTypeListData                 // We need to update the list data
)

type (
	Diff struct {
		Logger *zerolog.Logger
	}

	DiffType int
)

//----------------------------------------------------------------------------------------------------------------------------------------------------

type GetAnimeDiffOptions struct {
	Collection      *anilist.AnimeCollection
	LocalCollection mo.Option[*anilist.AnimeCollection]
	LocalFiles      []*anime.LocalFile
	TrackedAnime    map[int]*TrackedMedia
	Snapshots       map[int]*AnimeSnapshot
}

type AnimeDiffResult struct {
	AnimeEntry    *anilist.AnimeListEntry
	AnimeSnapshot *AnimeSnapshot
	DiffType      DiffType
}

// GetAnimeDiffs returns the anime that have changed.
// The anime is considered changed if:
// - It doesn't have a snapshot
// - The reference key is different (e.g. the number of local files has changed), meaning we need to update the snapshot.
func (d *Diff) GetAnimeDiffs(opts GetAnimeDiffOptions) map[int]*AnimeDiffResult {

	collection := opts.Collection
	localCollection := opts.LocalCollection
	trackedAnimeMap := opts.TrackedAnime
	snapshotMap := opts.Snapshots

	changedMap := make(map[int]*AnimeDiffResult)

	if len(collection.MediaListCollection.Lists) == 0 || len(trackedAnimeMap) == 0 {
		return changedMap
	}

	for _, _list := range collection.MediaListCollection.Lists {
		if _list.GetStatus() == nil || _list.GetEntries() == nil {
			continue
		}
		for _, _entry := range _list.GetEntries() {
			// Check if the anime is tracked
			_, isTracked := trackedAnimeMap[_entry.GetMedia().GetID()]
			if !isTracked {
				continue
			}

			if localCollection.IsAbsent() {
				d.Logger.Trace().Msgf("sync: Diff > Anime %d, local collection is missing", _entry.GetMedia().GetID())
				changedMap[_entry.GetMedia().GetID()] = &AnimeDiffResult{
					AnimeEntry: _entry,
					DiffType:   DiffTypeMissing,
				}
				continue // Go to the next anime
			}

			// Check if the anime has a snapshot
			snapshot, hasSnapshot := snapshotMap[_entry.GetMedia().GetID()]
			if !hasSnapshot {
				d.Logger.Trace().Msgf("sync: Diff > Anime %d is missing a snapshot", _entry.GetMedia().GetID())
				changedMap[_entry.GetMedia().GetID()] = &AnimeDiffResult{
					AnimeEntry: _entry,
					DiffType:   DiffTypeMissing,
				}
				continue // Go to the next anime
			}

			_lfs := lo.Filter(opts.LocalFiles, func(lf *anime.LocalFile, _ int) bool {
				return lf.MediaId == _entry.GetMedia().GetID()
			})

			// Check if the anime has changed
			_referenceKey := GetAnimeReferenceKey(_entry.Media, _lfs)

			// Check if the reference key is different
			if snapshotMap[_entry.GetMedia().GetID()].ReferenceKey != _referenceKey {
				d.Logger.Trace().Str("localReferenceKey", snapshotMap[_entry.GetMedia().GetID()].ReferenceKey).Str("currentReferenceKey", _referenceKey).Msgf("sync: Diff > Anime %d has an outdated snapshot", _entry.GetMedia().GetID())
				changedMap[_entry.GetMedia().GetID()] = &AnimeDiffResult{
					AnimeEntry:    _entry,
					AnimeSnapshot: snapshot,
					DiffType:      DiffTypeMetadata,
				}
				continue // Go to the next anime
			}

			localEntry, found := localCollection.MustGet().GetListEntryFromAnimeId(_entry.GetMedia().GetID())
			if !found {
				d.Logger.Trace().Msgf("sync: Diff > Anime %d is missing from the local collection", _entry.GetMedia().GetID())
				changedMap[_entry.GetMedia().GetID()] = &AnimeDiffResult{
					AnimeEntry:    _entry,
					AnimeSnapshot: snapshot,
					DiffType:      DiffTypeMissing,
				}
				continue // Go to the next anime
			}

			// Check if the list data has changed
			_listDataKey := GetAnimeListDataKey(_entry)
			localListDataKey := GetAnimeListDataKey(localEntry)

			if _listDataKey != localListDataKey {
				d.Logger.Trace().Str("localListDataKey", localListDataKey).Str("currentListDataKey", _listDataKey).Msgf("sync: Diff > Anime %d has changed list data", _entry.GetMedia().GetID())
				changedMap[_entry.GetMedia().GetID()] = &AnimeDiffResult{
					AnimeEntry:    _entry,
					AnimeSnapshot: snapshot,
					DiffType:      DiffTypeListData,
				}
				continue // Go to the next anime
			}

		}
	}

	return changedMap
}

//----------------------------------------------------------------------------------------------------------------------------------------------------

type GetMangaDiffOptions struct {
	Collection                  *anilist.MangaCollection
	LocalCollection             mo.Option[*anilist.MangaCollection]
	DownloadedChapterContainers []*manga.ChapterContainer
	TrackedManga                map[int]*TrackedMedia
	Snapshots                   map[int]*MangaSnapshot
}

type MangaDiffResult struct {
	MangaEntry    *anilist.MangaListEntry
	MangaSnapshot *MangaSnapshot
	DiffType      DiffType
}

// GetMangaDiffs returns the manga that have changed.
func (d *Diff) GetMangaDiffs(opts GetMangaDiffOptions) map[int]*MangaDiffResult {

	collection := opts.Collection
	localCollection := opts.LocalCollection
	trackedMangaMap := opts.TrackedManga
	snapshotMap := opts.Snapshots

	changedMap := make(map[int]*MangaDiffResult)

	if len(collection.MediaListCollection.Lists) == 0 || len(trackedMangaMap) == 0 {
		return changedMap
	}

	for _, _list := range collection.MediaListCollection.Lists {
		if _list.GetStatus() == nil || _list.GetEntries() == nil {
			continue
		}
		for _, _entry := range _list.GetEntries() {
			// Check if the manga is tracked
			_, isTracked := trackedMangaMap[_entry.GetMedia().GetID()]
			if !isTracked {
				continue
			}

			if localCollection.IsAbsent() {
				d.Logger.Trace().Msgf("sync: Diff > Manga %d, local collection is missing", _entry.GetMedia().GetID())
				changedMap[_entry.GetMedia().GetID()] = &MangaDiffResult{
					MangaEntry: _entry,
					DiffType:   DiffTypeMissing,
				}
				continue // Go to the next manga
			}

			// Check if the manga has a snapshot
			snapshot, hasSnapshot := snapshotMap[_entry.GetMedia().GetID()]
			if !hasSnapshot {
				d.Logger.Trace().Msgf("sync: Diff > Manga %d is missing a snapshot", _entry.GetMedia().GetID())
				changedMap[_entry.GetMedia().GetID()] = &MangaDiffResult{
					MangaEntry: _entry,
					DiffType:   DiffTypeMissing,
				}
				continue // Go to the next manga
			}

			// Check if the manga has changed
			_referenceKey := GetMangaReferenceKey(_entry.Media, opts.DownloadedChapterContainers)

			// Check if the reference key is different
			if snapshotMap[_entry.GetMedia().GetID()].ReferenceKey != _referenceKey {
				d.Logger.Trace().Str("localReferenceKey", snapshotMap[_entry.GetMedia().GetID()].ReferenceKey).Str("currentReferenceKey", _referenceKey).Msgf("sync: Diff > Manga %d has an outdated snapshot", _entry.GetMedia().GetID())
				changedMap[_entry.GetMedia().GetID()] = &MangaDiffResult{
					MangaEntry:    _entry,
					MangaSnapshot: snapshot,
					DiffType:      DiffTypeMetadata,
				}
				continue // Go to the next manga
			}

			localEntry, found := localCollection.MustGet().GetListEntryFromMangaId(_entry.GetMedia().GetID())
			if !found {
				d.Logger.Trace().Msgf("sync: Diff > Manga %d is missing from the local collection", _entry.GetMedia().GetID())
				changedMap[_entry.GetMedia().GetID()] = &MangaDiffResult{
					MangaEntry:    _entry,
					MangaSnapshot: snapshot,
					DiffType:      DiffTypeMissing,
				}
				continue // Go to the next manga
			}

			// Check if the list data has changed
			_listDataKey := GetMangaListDataKey(_entry)
			localListDataKey := GetMangaListDataKey(localEntry)

			if _listDataKey != localListDataKey {
				d.Logger.Trace().Str("localListDataKey", localListDataKey).Str("currentListDataKey", _listDataKey).Msgf("sync: Diff > Manga %d has changed list data", _entry.GetMedia().GetID())
				changedMap[_entry.GetMedia().GetID()] = &MangaDiffResult{
					MangaEntry:    _entry,
					MangaSnapshot: snapshot,
					DiffType:      DiffTypeListData,
				}
				continue // Go to the next manga
			}

		}
	}

	return changedMap
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func GetAnimeReferenceKey(bAnime *anilist.BaseAnime, lfs []*anime.LocalFile) string {
	// Reference key is used to compare the snapshot with the current data.
	// If the reference key is different, the snapshot is outdated.
	animeLfs := lo.Filter(lfs, func(lf *anime.LocalFile, _ int) bool {
		return lf.MediaId == bAnime.ID
	})

	// Extract the paths and sort them to maintain a consistent order.
	paths := lo.Map(animeLfs, func(lf *anime.LocalFile, _ int) string {
		return lf.Path
	})
	slices.Sort(paths)

	return fmt.Sprintf("%d-%s", bAnime.ID, strings.Join(paths, ","))
}

func GetMangaReferenceKey(bManga *anilist.BaseManga, dcc []*manga.ChapterContainer) string {
	// Reference key is used to compare the snapshot with the current data.
	// If the reference key is different, the snapshot is outdated.
	mangaDcc := lo.Filter(dcc, func(dc *manga.ChapterContainer, _ int) bool {
		return dc.MediaId == bManga.ID
	})

	slices.SortFunc(mangaDcc, func(i, j *manga.ChapterContainer) int {
		return strings.Compare(i.Provider, j.Provider)
	})
	var k string
	for _, dc := range mangaDcc {
		l := dc.Provider + "-"
		slices.SortFunc(dc.Chapters, func(i, j *hibikemanga.ChapterDetails) int {
			return strings.Compare(i.ID, j.ID)
		})
		for _, c := range dc.Chapters {
			l += c.ID + "-"
		}
		k += l
	}

	return fmt.Sprintf("%d-%s", bManga.ID, k)
}

func GetAnimeListDataKey(entry *anilist.AnimeListEntry) string {
	return fmt.Sprintf("%s-%d-%f-%d-%v-%v-%v-%v-%v-%v",
		MediaListStatusPointerValue(entry.GetStatus()),
		IntPointerValue(entry.GetProgress()),
		Float64PointerValue(entry.GetScore()),
		IntPointerValue(entry.GetRepeat()),
		IntPointerValue(entry.GetStartedAt().GetYear()),
		IntPointerValue(entry.GetStartedAt().GetMonth()),
		IntPointerValue(entry.GetStartedAt().GetDay()),
		IntPointerValue(entry.GetCompletedAt().GetYear()),
		IntPointerValue(entry.GetCompletedAt().GetMonth()),
		IntPointerValue(entry.GetCompletedAt().GetDay()),
	)
}

func GetMangaListDataKey(entry *anilist.MangaListEntry) string {
	return fmt.Sprintf("%s-%d-%f-%d-%v-%v-%v-%v-%v-%v",
		MediaListStatusPointerValue(entry.GetStatus()),
		IntPointerValue(entry.GetProgress()),
		Float64PointerValue(entry.GetScore()),
		IntPointerValue(entry.GetRepeat()),
		IntPointerValue(entry.GetStartedAt().GetYear()),
		IntPointerValue(entry.GetStartedAt().GetMonth()),
		IntPointerValue(entry.GetStartedAt().GetDay()),
		IntPointerValue(entry.GetCompletedAt().GetYear()),
		IntPointerValue(entry.GetCompletedAt().GetMonth()),
		IntPointerValue(entry.GetCompletedAt().GetDay()),
	)
}
