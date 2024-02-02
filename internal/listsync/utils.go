package listsync

import (
	"fmt"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/mal"
	"github.com/sourcegraph/conc/pool"
)

func (e *AnimeEntry) FindMetadataDiffs(other *AnimeEntry) ([]AnimeMetadataDiffType, bool) {
	if e.IsEqual(other) {
		return nil, false
	}
	var diffs []AnimeMetadataDiffType
	if e.Status != other.Status {
		diffs = append(diffs, AnimeMetadataDiffTypeStatus)
	}
	if e.Progress != other.Progress {
		diffs = append(diffs, AnimeMetadataDiffTypeProgress)
	}
	if e.Score != other.Score {
		diffs = append(diffs, AnimeMetadataDiffTypeScore)
	}
	return diffs, true
}

func (e *AnimeEntry) IsEqual(other *AnimeEntry) bool {
	if e.Status != other.Status {
		return false
	}
	if e.Progress != other.Progress {
		return false
	}
	if e.Score != other.Score {
		return false
	}
	return true
}

// FromAnilistCollection converts an AniList anime collection to a list of AnimeEntry
func FromAnilistCollection(collection *anilist.AnimeCollection) []*AnimeEntry {
	p := pool.NewWithResults[*AnimeEntry]()
	for _, list := range collection.GetMediaListCollection().GetLists() {
		list := list
		for _, entry := range list.GetEntries() {
			entry := entry
			p.Go(func() *AnimeEntry {
				media := entry.GetMedia()
				animeEntry, ok := NewAnimeEntryFromAnilistBaseMedia(media)
				if !ok {
					return nil
				}
				animeEntry.Status = FromAnilistListStatus(*entry.GetStatus()) // Update status
				if entry.GetProgress() != nil {
					animeEntry.Progress = *entry.GetProgress() // Update progress
				}
				if entry.GetScore() != nil {
					animeEntry.Score = FromAnilistFloatScore(entry.GetScore()) // Update score
				}
				return animeEntry
			})
		}
	}
	ret := p.Wait()
	ret = lo.Filter(ret, func(entry *AnimeEntry, _ int) bool {
		return entry != nil
	})

	return ret
}

// FromMALCollection converts a MAL anime collection to a list of AnimeEntry
func FromMALCollection(collection []*mal.AnimeListEntry) []*AnimeEntry {

	p := pool.NewWithResults[*AnimeEntry]()
	for _, entry := range collection {
		entry := entry
		p.Go(func() *AnimeEntry {
			return NewAnimeEntryFromMALBasicAnime(entry)
		})
	}
	ret := p.Wait()

	return ret
}

// NewAnimeEntryFromAnilistBaseMedia converts an anilist.BaseMedia to an AnimeEntry
// "Progress", "Score" are set to 0, "Status" is set to AnimeStatusUnknown
func NewAnimeEntryFromAnilistBaseMedia(media *anilist.BaseMedia) (*AnimeEntry, bool) {
	if media.IDMal == nil {
		return nil, false
	}

	return &AnimeEntry{
		Source:       SourceAniList,
		SourceID:     media.ID,
		MalID:        *media.IDMal,
		DisplayTitle: media.GetTitleSafe(),
		Url:          fmt.Sprintf("https://anilist.co/anime/%d", media.ID),
		TotalEpisode: media.GetTotalEpisodeCount(),
		Image:        *media.GetBannerImage(),
		Status:       AnimeStatusUnknown,
		Progress:     0,
		Score:        0,
	}, true
}

// NewAnimeEntryFromMALBasicAnime converts a mal.BasicAnime to an AnimeEntry
// "Progress", "Score" are set to 0, "Status" is set to AnimeStatusUnknown
func NewAnimeEntryFromMALBasicAnime(entry *mal.AnimeListEntry) *AnimeEntry {
	return &AnimeEntry{
		Source:       SourceMAL,
		SourceID:     entry.Node.ID,
		MalID:        entry.Node.ID,
		DisplayTitle: entry.Node.Title,
		Url:          fmt.Sprintf("https://myanimelist.net/anime/%d", entry.Node.ID),
		TotalEpisode: 0, // DEVNOTE: MAL does not provide total episode count
		Image:        entry.Node.MainPicture.Large,
		Status:       FromMALStatusToAnimeStatus(entry.ListStatus.Status),
		Progress:     entry.ListStatus.NumWatchedEpisodes,
		Score:        entry.ListStatus.Score,
	}
}

func FromAnilistFloatScore(score *float64) int {
	if score == nil {
		return 0
	}
	return int(*score * 10)
}

func FromAnilistListStatus(status anilist.MediaListStatus) AnimeListStatus {
	switch status {
	case anilist.MediaListStatusCurrent:
		return AnimeStatusWatching
	case anilist.MediaListStatusPlanning:
		return AnimeStatusPlanning
	case anilist.MediaListStatusDropped:
		return AnimeStatusDropped
	case anilist.MediaListStatusCompleted:
		return AnimeStatusCompleted
	case anilist.MediaListStatusPaused:
		return AnimeStatusPaused
	case anilist.MediaListStatusRepeating:
		return AnimeStatusWatching
	default:
		return AnimeStatusUnknown
	}
}

func ToAnilistListStatus(status AnimeListStatus) anilist.MediaListStatus {
	switch status {
	case AnimeStatusWatching:
		return anilist.MediaListStatusCurrent
	case AnimeStatusPlanning:
		return anilist.MediaListStatusPlanning
	case AnimeStatusDropped:
		return anilist.MediaListStatusDropped
	case AnimeStatusCompleted:
		return anilist.MediaListStatusCompleted
	case AnimeStatusPaused:
		return anilist.MediaListStatusPaused
	default:
		return anilist.MediaListStatusPlanning // default to planning
	}
}

func FromMALStatusToAnimeStatus(status mal.MediaListStatus) AnimeListStatus {
	switch status {
	case mal.MediaListStatusWatching:
		return AnimeStatusWatching
	case mal.MediaListStatusPlanToWatch:
		return AnimeStatusPlanning
	case mal.MediaListStatusDropped:
		return AnimeStatusDropped
	case mal.MediaListStatusCompleted:
		return AnimeStatusCompleted
	case mal.MediaListStatusOnHold:
		return AnimeStatusPaused
	default:
		return AnimeStatusUnknown
	}
}

func ToMALStatusFromAnimeStatus(status AnimeListStatus) mal.MediaListStatus {
	switch status {
	case AnimeStatusWatching:
		return mal.MediaListStatusWatching
	case AnimeStatusPlanning:
		return mal.MediaListStatusPlanToWatch
	case AnimeStatusDropped:
		return mal.MediaListStatusDropped
	case AnimeStatusCompleted:
		return mal.MediaListStatusCompleted
	case AnimeStatusPaused:
		return mal.MediaListStatusOnHold
	default:
		return mal.MediaListStatusPlanToWatch // default to planning
	}
}
