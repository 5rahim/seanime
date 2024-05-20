package tracker

const (
	AnimeKind         = "anime"
	MangaKind         = "manga"
	AnimeAndMangaKind = "anime_and_manga"
)

type (
	Kind string // Kind is the type of tracker

	BaseTracker interface {
		// GetName returns the name of the tracker
		GetName() string
		GetLogo() string
		GetUsername() string
		GetPassword() string
		GetKind() Kind
	}
)
