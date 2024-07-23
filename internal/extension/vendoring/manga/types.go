package vendor_hibike_manga

type (
	Provider interface {
		Search(opts SearchOptions) ([]*SearchResult, error)
		FindChapters(id string) ([]*ChapterDetails, error)
		FindChapterPages(id string) ([]*ChapterPage, error)
	}

	SearchOptions struct {
		Query string
		Year  int
	}

	SearchResult struct {
		// Provider is the ID of the provider.
		// This should be the same as the extension ID and follow the same format.
		Provider string `json:"provider"`
		// ID is the manga slug.
		// It is used to fetch the chapter details.
		// It can be a combination of keys separated by the $ delimiter.
		ID string `json:"id"`
		// The title of the manga.
		Title string `json:"title"`
		// Synonyms are alternative titles for the manga.
		Synonyms []string `json:"synonyms,omitempty"`
		// Year is the year the manga was released.
		Year int `json:"year,omitempty"`
		// Image is the URL of the manga cover image.
		Image string `json:"image,omitempty"`
		// SearchRating shows how well the chapter title matches the search query.
		// It is a number from 0 to 1.
		SearchRating float64 `json:"searchRating,omitempty"`
	}

	ChapterDetails struct {
		// ID of the provider.
		// This should be the same as the extension ID and follow the same format.
		Provider string `json:"provider"`
		// ID is the chapter slug.
		// It is used to fetch the chapter pages.
		// It can be a combination of keys separated by the $ delimiter.
		// e.g., "10010$one-piece-1", where "10010" is the manga ID and "one-piece-1" is the chapter slug that is reconstructed to "%url/10010/one-piece-1".
		ID string `json:"id"`
		// URL is the chapter page URL.
		URL string `json:"url"`
		// Title is the chapter title.
		// It should start with "Chapter X" or "Chapter X.Y" where X is the chapter number and Y is the subchapter number.
		Title string `json:"title"`
		// e.g., "1", "1.5", "2", "3"
		Chapter string `json:"chapter"`
		// From 0 to n
		Index uint `json:"index"`
		// Rating is the rating of the chapter. It is a number from 0 to 100.
		// Leave it empty if the rating is not available.
		Rating int `json:"rating,omitempty"`
		// UpdatedAt is the date when the chapter was last updated.
		// It should be in the format "YYYY-MM-DD".
		// Leave it empty if the date is not available.
		UpdatedAt string `json:"updatedAt,omitempty"`
	}

	ChapterPage struct {
		// ID of the provider.
		// This should be the same as the extension ID and follow the same format.
		Provider string `json:"provider"`
		// URL of the chapter page.
		URL string `json:"url"`
		// Index of the page in the chapter.
		// From 0 to n.
		Index int `json:"index"`
		// Request headers for the page if proxying is required.
		Headers map[string]string `json:"headers"`
	}
)
