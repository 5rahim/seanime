package offline

import (
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/library/entities"
	"github.com/seanime-app/seanime/internal/manga"
)

const (
	EntryStatusCurrent   EntryStatus = "current"
	EntryStatusPlanned   EntryStatus = "planned"
	EntryStatusCompleted EntryStatus = "completed"
	EntryStatusDropped   EntryStatus = "dropped"
	EntryStatusPaused    EntryStatus = "paused"
)

const (
	EntryTypeAnime EntryType = "anime"
	EntryTypeManga EntryType = "manga"
)

type (
	EntryStatus string
	EntryType   string

	Snapshot struct {
		DbId        uint              `json:"dbId"`
		User        *entities.User    `json:"user"`
		Entries     *Entries          `json:"entries"`
		Collections *Collections      `json:"libraryCollections"`
		AssetMap    map[string]string `json:"assetMap"` // Key: URL, Value: Local path
	}

	// Collections is a snapshot of the user's AniList collections.
	// This is created by [Snapshot] and is stored for offline use.
	//  - Used to download images for offline use.
	//  - Used as a metadata source for offline use.
	Collections struct {
		AnimeCollection *anilist.AnimeCollection `json:"animeCollection"`
		MangaCollection *anilist.MangaCollection `json:"mangaCollection"`
	}

	// Entries is a snapshot of the user's list entries.
	// This is created by [Snapshot] and is stored for offline use.
	//  - AssetsHandler walks through this struct to download assets when it is created.
	//  - Used to compare and update the user's list entries when they come online.
	Entries struct {
		AnimeEntries []*AnimeEntry `json:"animeEntries"` // All anime entries in the local library
		MangaEntries []*MangaEntry `json:"mangaEntries"` // Will only contain manga entries with downloaded chapters
	}

	// AnimeEntry is a snapshot of an anime list entry.
	//  - Updates are made to this struct, then saved to the database.
	AnimeEntry struct {
		MediaId          int                           `json:"mediaId"`
		ListData         *ListData                     `json:"listData"`
		Media            *anilist.BaseMedia            `json:"media"`
		Episodes         []*entities.MediaEntryEpisode `json:"episodes"`
		DownloadedAssets bool                          `json:"downloadedAssets"`
	}

	// MangaEntry is a snapshot of a manga list entry.
	//  - Updates are made to this struct, then saved to the database.
	MangaEntry struct {
		MediaId          int                     `json:"mediaId"`
		ListData         *ListData               `json:"listData"`
		Media            *anilist.BaseManga      `json:"media"`
		ChapterContainer *manga.ChapterContainer `json:"chapterContainer"`
		DownloadedAssets bool                    `json:"downloadedAssets"`
	}

	ListData struct {
		Score       float64     `json:"score"`
		Status      EntryStatus `json:"status"`
		Progress    int         `json:"progress"`
		StartedAt   string      `json:"startedAt"`
		CompletedAt string      `json:"completedAt"`
	}
)

func anilistStatusToEntryStatus(as *anilist.MediaListStatus) EntryStatus {
	switch *as {
	case anilist.MediaListStatusCurrent:
		return EntryStatusCurrent
	case anilist.MediaListStatusPlanning:
		return EntryStatusPlanned
	case anilist.MediaListStatusCompleted:
		return EntryStatusCompleted
	case anilist.MediaListStatusDropped:
		return EntryStatusDropped
	case anilist.MediaListStatusPaused:
		return EntryStatusPaused
	default:
		return EntryStatusPlanned
	}
}
