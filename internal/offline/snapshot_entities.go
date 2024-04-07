package offline

import "github.com/seanime-app/seanime/internal/api/anilist"

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

	// Data represents the offline data, sent to the client.
	Data struct {
		Entries     *Entries     `json:"entries"`
		Collections *Collections `json:"libraryCollections"`
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
	//  - Updates are made to this struct, then saved to the database.
	//  - Used to compare and update the user's list entries when they come online.
	Entries struct {
		AnimeEntries []*Entry `json:"animeEntries"` // All anime entries in the local library
		MangaEntries []*Entry `json:"mangaEntries"` // Will only contain manga entries with downloaded chapters
	}

	// Entry represents a single entry in the user's list.
	Entry struct {
		MediaId  int            `json:"mediaId"`
		Status   EntryStatus    `json:"status"`
		Progress int            `json:"progress"`
		Type     EntryType      `json:"type"`
		Metadata *EntryMetadata `json:"metadata,omitempty"`
	}

	// EntryMetadata is created during the snapshot process, if the user chooses to download additional metadata.
	EntryMetadata struct {
		CoverImage      string                          `json:"coverImage"`
		BannerImage     string                          `json:"bannerImage"`
		EpisodeMetadata []*EntryMetadataEpisodeMetadata `json:"episodeMetadata,omitempty"`
	}

	EntryMetadataEpisodeMetadata struct {
		EpisodeNumber int     `json:"episodeNumber"`
		Title         string  `json:"title"`
		Image         *string `json:"image,omitempty"`
		LocalFilePath string  `json:"localFilePath"`
	}
)
