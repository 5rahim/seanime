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
		Provider  Provider `json:"provider"`
		ID        string   `json:"id"`
		URL       string   `json:"url"`
		Title     string   `json:"title"`
		Chapter   string   `json:"chapter"` // e.g., "1", "1.5", "2", "3"
		Index     uint     `json:"index"`   // Index of the chapter in the manga
		Rating    int      `json:"rating"`
		UpdatedAt string   `json:"updatedAt"`
	}

	ChapterPage struct {
		Provider Provider          `json:"provider"`
		URL      string            `json:"url"`
		Index    int               `json:"index"`
		Headers  map[string]string `json:"headers"`
	}
)
