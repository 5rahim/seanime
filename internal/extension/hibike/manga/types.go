package hibikemanga

type (
	Provider interface {
		// Search returns the search results for the given query.
		Search(opts SearchOptions) ([]*SearchResult, error)
		// FindChapters returns the chapter details for the given manga ID.
		FindChapters(id string) ([]*ChapterDetails, error)
		// FindChapterPages returns the chapter pages for the given chapter ID.
		FindChapterPages(id string) ([]*ChapterPage, error)
		// GetSettings returns the provider settings.
		GetSettings() Settings
	}

	Settings struct {
		SupportsMultiScanlator bool `json:"supportsMultiScanlator"`
		SupportsMultiLanguage  bool `json:"supportsMultiLanguage"`
	}

	SearchOptions struct {
		Query string `json:"query"`
		// Year is the year the manga was released.
		// It will be 0 if the year is not available.
		Year int `json:"year"`
	}

	SearchResult struct {
		// "ID" of the extension.
		Provider string `json:"provider"`
		// ID of the manga, used to fetch the chapter details.
		// It can be a combination of keys separated by a delimiter. (Delimiters should not be slashes).
		ID string `json:"id"`
		// The title of the manga.
		Title string `json:"title"`
		// Synonyms are alternative titles for the manga.
		Synonyms []string `json:"synonyms,omitempty"`
		// Year the manga was released.
		Year int `json:"year,omitempty"`
		// URL of the manga cover image.
		Image string `json:"image,omitempty"`
		// Indicates how well the chapter title matches the search query.
		// It is a number from 0 to 1.
		// Leave it empty if the comparison should be done by Seanime.
		SearchRating float64 `json:"searchRating,omitempty"`
	}

	ChapterDetails struct {
		// "ID" of the extension.
		// This should be the same as the extension ID and follow the same format.
		Provider string `json:"provider"`
		// ID of the chapter, used to fetch the chapter pages.
		// It can be a combination of keys separated by a delimiter. (Delimiters should not be slashes).
		//	If the same ID has multiple languages, the language key should be included. (e.g., "one-piece-001$chapter-1$en").
		//	If the same ID has multiple scanlators, the group key should be included. (e.g., "one-piece-001$chapter-1$group-1").
		ID string `json:"id"`
		// The chapter page URL.
		URL string `json:"url"`
		// The chapter title.
		// It should be in this format: "Chapter X.Y - {title}" where X is the chapter number and Y is the subchapter number.
		Title string `json:"title"`
		// e.g., "1", "1.5", "2", "3"
		Chapter string `json:"chapter"`
		// From 0 to n
		Index uint `json:"index"`
		// The scanlator that translated the chapter.
		// Leave it empty if your extension does not support multiple scanlators.
		Scanlator string `json:"scanlator,omitempty"`
		// The language of the chapter.
		// Leave it empty if your extension does not support multiple languages.
		Language string `json:"language,omitempty"`
		// The rating of the chapter. It is a number from 0 to 100.
		// Leave it empty if the rating is not available.
		Rating int `json:"rating,omitempty"`
		// UpdatedAt is the date when the chapter was last updated.
		// It should be in the format "YYYY-MM-DD".
		// Leave it empty if the date is not available.
		UpdatedAt string `json:"updatedAt,omitempty"`

		// LocalIsPDF is true if the chapter is a single, readable PDF file.
		LocalIsPDF bool `json:"localIsPDF,omitempty"`
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

		Buf []byte `json:"-"`
	}
)
