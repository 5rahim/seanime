package listsync

const (
	SourceAniList        Source          = "anilist"
	SourceMAL            Source          = "mal"
	AnimeStatusWatching  AnimeListStatus = "watching"
	AnimeStatusPlanning  AnimeListStatus = "planning"
	AnimeStatusDropped   AnimeListStatus = "dropped"
	AnimeStatusCompleted AnimeListStatus = "completed"
	AnimeStatusPaused    AnimeListStatus = "paused"
	AnimeStatusUnknown   AnimeListStatus = "unknown"
)

type (
	Source          string
	AnimeListStatus string
	AnimeEntry      struct {
		Source       Source
		SourceID     int
		MalID        int // Used for matching
		DisplayTitle string
		Url          string
		Progress     int
		TotalEpisode int
		Status       AnimeListStatus
		Image        string
		Score        int
	}
)
