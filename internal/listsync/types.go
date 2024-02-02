package listsync

const (
	SourceAniList                 Source                = "anilist"
	SourceMAL                     Source                = "mal"
	AnimeStatusWatching           AnimeListStatus       = "watching"
	AnimeStatusPlanning           AnimeListStatus       = "planning"
	AnimeStatusDropped            AnimeListStatus       = "dropped"
	AnimeStatusCompleted          AnimeListStatus       = "completed"
	AnimeStatusPaused             AnimeListStatus       = "paused"
	AnimeStatusUnknown            AnimeListStatus       = "unknown"
	AnimeDiffKindMissingOrigin    AnimeDiffKind         = "missing_in_origin" // Anime is missing in the origin (i.e. Delete from target)
	AnimeDiffKindMissingTarget    AnimeDiffKind         = "missing_in_target" // Anime is missing in the target (i.e. Add to target)
	AnimeDiffKindMetadata         AnimeDiffKind         = "metadata"          // Anime metadata is different in the origin and the target (i.e. Update in target)
	AnimeMetadataDiffKindScore    AnimeMetadataDiffKind = "score"
	AnimeMetadataDiffKindProgress AnimeMetadataDiffKind = "progress"
	AnimeMetadataDiffStatus       AnimeMetadataDiffKind = "status"
)

type (
	AnimeDiffKind         string
	AnimeMetadataDiffKind string
	ListSync              struct {
		Origin  *Provider
		Targets []*Provider
	}
	MissingAnime struct {
		Provider      *Provider
		OriginEntries []*AnimeEntry // Entries that are present in the origin but not in the target
	}
	AnimeDiff struct {
		TargetSource      Source                  `json:"targetSource"`
		OriginEntry       *AnimeEntry             `json:"originEntry"`
		TargetEntry       *AnimeEntry             `json:"targetEntry"` // Entry that will be updated
		Kind              AnimeDiffKind           `json:"kind"`
		MetadataDiffKinds []AnimeMetadataDiffKind `json:"metadataDiffKinds"`
	}
)

type (
	Source          string
	AnimeListStatus string
	AnimeEntry      struct {
		Source       Source          `json:"source"`
		SourceID     int             `json:"sourceID"`
		MalID        int             `json:"malID"` // Used for matching
		DisplayTitle string          `json:"displayTitle"`
		Url          string          `json:"url"`
		Progress     int             `json:"progress"`
		TotalEpisode int             `json:"totalEpisode"`
		Status       AnimeListStatus `json:"status"`
		Image        string          `json:"image"`
		Score        int             `json:"score"`
	}
)
