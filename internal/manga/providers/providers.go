package manga_providers

const (
	MangaseeProvider Provider = "mangasee"
	MangadexProvider Provider = "mangadex"
	ComickProvider   Provider = "comick"
)

type (
	Provider string

	MangaProvider interface {
		Search(opts SearchOptions) ([]*SearchResult, error)
		FindChapters(id string) ([]*ChapterDetails, error)
		FindChapterPages(info *ChapterDetails) ([]*ChapterPage, error)
	}

	SearchOptions struct {
		Query string
		Year  int
	}

	SearchResult struct {
		ID           string
		Title        string
		Synonyms     []string
		Year         int
		Image        string
		Provider     Provider
		SearchRating float64
	}

	ChapterDetails struct {
		Provider  Provider
		ID        string
		URL       string
		Title     string
		Chapter   string // e.g., "1", "1.5", "2", "3"
		Index     uint   // Index of the chapter in the manga
		Rating    int
		UpdatedAt string
	}

	ChapterPage struct {
		Provider Provider
		URL      string
		Index    int
		Headers  map[string]string
	}
)
