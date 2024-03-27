package manga_providers

const (
	MangaseeProvider Provider = "mangasee"
	MangadexProvider Provider = "mangadex"
	ComickProvider   Provider = "comick"
)

type (
	Provider string

	MangaProvider interface {
		Search(query string) ([]*SearchResult, error)
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
		Slug      string
		URL       string
		Title     string
		Number    int
		Rating    int
		UpdatedAt int
	}

	ChapterPage struct {
		Provider Provider
		Url      string
		Index    int
		Headers  map[string]string
	}
)
