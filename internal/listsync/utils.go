package listsync

import (
	"fmt"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/mal"
)

// NewAnimeEntryFromAnilistBaseMedia converts an anilist.BaseMedia to an AnimeEntry
// "Progress", "Score" are set to 0, "Status" is set to AnimeStatusUnknown
func NewAnimeEntryFromAnilistBaseMedia(media *anilist.BaseMedia) *AnimeEntry {
	return &AnimeEntry{
		Source:       SourceAniList,
		ID:           media.ID,
		DisplayTitle: media.GetTitleSafe(),
		Url:          fmt.Sprintf("https://anilist.co/anime/%d", media.ID),
		TotalEpisode: media.GetTotalEpisodeCount(),
		Image:        *media.GetBannerImage(),
		Status:       AnimeStatusUnknown,
		Progress:     0,
		Score:        0,
	}
}

// NewAnimeEntryFromMALBasicAnime converts a mal.BasicAnime to an AnimeEntry
// "Progress", "Score" are set to 0, "Status" is set to AnimeStatusUnknown
func NewAnimeEntryFromMALBasicAnime(media *mal.BasicAnime) *AnimeEntry {
	return &AnimeEntry{
		Source:       SourceMAL,
		ID:           media.ID,
		DisplayTitle: media.Title,
		Url:          fmt.Sprintf("https://myanimelist.net/anime/%d", media.ID),
		TotalEpisode: media.NumEpisodes,
		Image:        media.MainPicture.Large,
		Status:       AnimeStatusUnknown,
		Progress:     0,
		Score:        0,
	}
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
