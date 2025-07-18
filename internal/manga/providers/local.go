package manga_providers

import (
	"archive/zip"
	"bytes"
	"fmt"
	// "image/jpeg"
	"io"
	"os"
	"path/filepath"
	hibikemanga "seanime/internal/extension/hibike/manga"
	"seanime/internal/util/comparison"
	"slices"
	"strconv"
	"strings"
	"sync"

	// "github.com/gen2brain/go-fitz"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

const (
	LocalServePath = "{{manga-local-assets}}"
)

type Local struct {
	dir    string // Directory to scan for manga
	logger *zerolog.Logger

	mu                 sync.Mutex
	currentChapterPath string
	currentZipCloser   io.Closer
	currentPages       map[string]*loadedPage
}

type loadedPage struct {
	buf  []byte
	page *hibikemanga.ChapterPage
}

// chapterEntry represents a potential chapter file or directory found during scanning
type chapterEntry struct {
	RelativePath string // Path relative to manga root (e.g., "mangaID/chapter1.cbz" or "mangaID/vol1/ch1.cbz")
	IsDir        bool   // Whether this entry is a directory
}

func NewLocal(dir string, logger *zerolog.Logger) hibikemanga.Provider {
	_ = os.MkdirAll(dir, 0755)

	return &Local{
		dir:          dir,
		logger:       logger,
		currentPages: make(map[string]*loadedPage),
	}
}

func (p *Local) GetSettings() hibikemanga.Settings {
	return hibikemanga.Settings{
		SupportsMultiScanlator: false,
		SupportsMultiLanguage:  false,
	}
}

func (p *Local) SetSourceDirectory(dir string) {
	if dir != "" {
		p.dir = dir
	}
}

func (p *Local) getAllManga() (res []*hibikemanga.SearchResult, err error) {
	if p.dir == "" {
		return make([]*hibikemanga.SearchResult, 0), nil
	}

	entries, err := os.ReadDir(p.dir)
	if err != nil {
		return nil, err
	}

	res = make([]*hibikemanga.SearchResult, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			res = append(res, &hibikemanga.SearchResult{
				ID:       entry.Name(),
				Title:    entry.Name(),
				Provider: LocalProvider,
			})
		}
	}

	return res, nil
}

func (p *Local) Search(opts hibikemanga.SearchOptions) (res []*hibikemanga.SearchResult, err error) {
	res = make([]*hibikemanga.SearchResult, 0)
	all, err := p.getAllManga()
	if err != nil {
		return nil, err
	}

	if opts.Query == "" {
		return all, nil
	}

	allTitles := make([]*string, len(all))
	for i, manga := range all {
		allTitles[i] = &manga.Title
	}
	compRes := comparison.CompareWithLevenshteinCleanFunc(&opts.Query, allTitles, cleanMangaTitle)

	var bestMatch *comparison.LevenshteinResult
	for _, res := range compRes {
		if bestMatch == nil || res.Distance < bestMatch.Distance {
			bestMatch = res
		}
	}

	if bestMatch == nil {
		return res, nil
	}

	if bestMatch.Distance > 3 {
		// If the best match is too far away, return no results
		return res, nil
	}

	manga, ok := lo.Find(all, func(manga *hibikemanga.SearchResult) bool {
		return manga.Title == *bestMatch.Value
	})

	if !ok {
		return res, nil
	}

	res = append(res, manga)

	return res, nil
}

func cleanMangaTitle(title string) string {
	title = strings.TrimSpace(title)

	// Remove some characters to make comparison easier
	title = strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '!' || r == '"' || r == '<' || r == '>' || r == '|' || r == ',' {
			return rune(0)
		}
		return r
	}, title)

	return title
}

// FindChapters scans the manga series directory and returns the chapters.
// Supports nested folder structures up to 2 levels deep.
//
// Example:
//
//	Series title/
//	├── Chapter 1/
//	│   ├── image_1.ext
//	│   └── image_n.ext
//	├── Chapter 2.pdf
//	└── Ch 1-10/
//	    ├── Ch 1/
//	    └── Ch 2/
func (p *Local) FindChapters(mangaID string) (res []*hibikemanga.ChapterDetails, err error) {
	if p.dir == "" {
		return make([]*hibikemanga.ChapterDetails, 0), nil
	}

	mangaPath := filepath.Join(p.dir, mangaID)

	p.logger.Trace().Str("mangaPath", mangaPath).Msg("manga: Finding local chapters")

	// Collect all potential chapter entries up to 2 levels deep
	chapterEntries, err := p.collectChapterEntries(mangaPath, mangaID, 0)
	if err != nil {
		return nil, err
	}

	res = make([]*hibikemanga.ChapterDetails, 0)
	// Go through all collected entries.
	for _, entry := range chapterEntries {
		scannedEntry, ok := scanChapterFilename(filepath.Base(entry.RelativePath))
		if !ok {
			continue
		}

		if len(scannedEntry.Chapter) != 1 {
			// Handle one-shots (no chapter number and only one entry)
			if len(scannedEntry.Chapter) == 0 && len(chapterEntries) == 1 {
				chapterTitle := "Chapter 1"
				if scannedEntry.ChapterTitle != "" {
					chapterTitle += " - " + scannedEntry.ChapterTitle
				}
				res = append(res, &hibikemanga.ChapterDetails{
					Provider:   LocalProvider,
					ID:         filepath.ToSlash(entry.RelativePath), // ID is the relative filepath, e.g. "/series/chapter_1.cbz" or "/series/vol1/ch1.cbz"
					URL:        "",
					Title:      chapterTitle,
					Chapter:    "1",
					Index:      0, // placeholder, will be set later
					LocalIsPDF: scannedEntry.IsPDF,
				})
			} else if len(scannedEntry.Chapter) == 2 {
				// Handle combined chapters (e.g. "Chapter 1-2")
				chapterTitle := "Chapter " + cleanChapter(scannedEntry.Chapter[0]) + "-" + cleanChapter(scannedEntry.Chapter[1])
				if scannedEntry.ChapterTitle != "" {
					chapterTitle += " - " + scannedEntry.ChapterTitle
				}
				res = append(res, &hibikemanga.ChapterDetails{
					Provider: LocalProvider,
					ID:       filepath.ToSlash(entry.RelativePath), // ID is the relative filepath, e.g. "/series/chapter_1.cbz" or "/series/vol1/ch1.cbz"
					URL:      "",
					Title:    chapterTitle,
					// Use the last chapter number as the chapter for progress tracking
					Chapter:    cleanChapter(scannedEntry.Chapter[1]),
					Index:      0, // placeholder, will be set later
					LocalIsPDF: scannedEntry.IsPDF,
				})
			}
			continue
		}

		ch := cleanChapter(scannedEntry.Chapter[0])
		chapterTitle := "Chapter " + ch
		if scannedEntry.ChapterTitle != "" {
			chapterTitle += " - " + scannedEntry.ChapterTitle
		}

		res = append(res, &hibikemanga.ChapterDetails{
			Provider:   LocalProvider,
			ID:         filepath.ToSlash(entry.RelativePath), // ID is the relative filepath, e.g. "/series/chapter_1.cbz" or "/series/vol1/ch1.cbz"
			URL:        "",
			Title:      chapterTitle,
			Chapter:    ch,
			Index:      0, // placeholder, will be set later
			LocalIsPDF: scannedEntry.IsPDF,
		})
	}

	// sort by chapter number (ascending)
	slices.SortFunc(res, func(a, b *hibikemanga.ChapterDetails) int {
		chA, _ := strconv.ParseFloat(a.Chapter, 64)
		chB, _ := strconv.ParseFloat(b.Chapter, 64)
		return int(chA - chB)
	})

	// set the indexes
	for i, chapter := range res {
		chapter.Index = uint(i)
	}

	return res, nil
}

// collectChapterEntries walks the directory tree up to maxDepth levels deep and collects
// all potential chapter files and directories.
func (p *Local) collectChapterEntries(currentPath, mangaID string, currentDepth int) (entries []*chapterEntry, err error) {
	const maxDepth = 2

	if currentDepth > maxDepth {
		return entries, nil
	}

	dirEntries, err := os.ReadDir(currentPath)
	if err != nil {
		return nil, err
	}

	entries = make([]*chapterEntry, 0)

	for _, entry := range dirEntries {
		entryPath := filepath.Join(currentPath, entry.Name())

		// Calculate relative path from manga root
		var relativePath string
		if currentDepth == 0 {
			// At manga root level
			relativePath = filepath.Join(mangaID, entry.Name())
		} else {
			// Get the relative part from current path
			relativeFromManga, err := filepath.Rel(filepath.Join(p.dir, mangaID), entryPath)
			if err != nil {
				continue
			}
			relativePath = filepath.Join(mangaID, relativeFromManga)
		}

		if entry.IsDir() {
			// Check if this directory contains only images (making it a chapter directory)
			isImageDirectory, _ := p.isImageOnlyDirectory(entryPath)

			if isImageDirectory {
				// Directory contains only images, treat it as a chapter
				entries = append(entries, &chapterEntry{
					RelativePath: relativePath,
					IsDir:        true,
				})
			} else if currentDepth < maxDepth {
				// Directory doesn't contain only images, recursively scan subdirectories
				subEntries, err := p.collectChapterEntries(entryPath, mangaID, currentDepth+1)
				if err != nil {
					continue
				}

				// If subdirectory contains chapters, add them
				if len(subEntries) > 0 {
					entries = append(entries, subEntries...)
				} else {
					// If no sub-chapters found, treat directory itself as potential chapter
					entries = append(entries, &chapterEntry{
						RelativePath: relativePath,
						IsDir:        true,
					})
				}
			} else {
				// At max depth, treat directory as potential chapter
				entries = append(entries, &chapterEntry{
					RelativePath: relativePath,
					IsDir:        true,
				})
			}
		} else {
			// File entry - check if it's a potential chapter file
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if ext == ".cbz" || ext == ".cbr" || ext == ".pdf" || ext == ".zip" {
				entries = append(entries, &chapterEntry{
					RelativePath: relativePath,
					IsDir:        false,
				})
			}
		}
	}

	return entries, nil
}

// isImageOnlyDirectory checks if a directory contains only image files (no subdirectories or other files)
func (p *Local) isImageOnlyDirectory(dirPath string) (bool, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false, err
	}

	if len(entries) == 0 {
		return false, nil
	}

	hasImages := false
	for _, entry := range entries {
		if entry.IsDir() {
			return false, nil
		}

		if isFileImage(entry.Name()) {
			hasImages = true
		} else {
			return false, nil
		}
	}

	return hasImages, nil
}

// "0001" -> "1", "0" -> "0"
func cleanChapter(ch string) string {
	if ch == "" {
		return ""
	}
	if ch == "0" {
		return "0"
	}
	if strings.HasPrefix(ch, "0") {
		return strings.TrimLeft(ch, "0")
	}
	return ch
}

// FindChapterPages will extract the images
func (p *Local) FindChapterPages(id string) (ret []*hibikemanga.ChapterPage, err error) {
	if p.dir == "" {
		return make([]*hibikemanga.ChapterPage, 0), nil
	}

	// id = filepath
	// e.g. "series/chapter_1.cbz"
	fullpath := filepath.Join(p.dir, id) // e.g. "/collection/series/chapter_1.cbz"

	// Prefix with {{manga-local-assets}} to signal the client that this is a local file
	// e.g. "{{manga-local-assets}}/series/chapter_1.cbz/image_1.jpg"
	formatUrl := func(fileName string) string {
		return filepath.ToSlash(filepath.Join(LocalServePath, id, fileName))
	}

	ext := filepath.Ext(fullpath)

	// Close the current pages
	if p.currentZipCloser != nil {
		_ = p.currentZipCloser.Close()
	}
	for _, loadedPage := range p.currentPages {
		loadedPage.buf = nil
	}
	p.currentPages = make(map[string]*loadedPage)
	p.currentZipCloser = nil
	p.currentChapterPath = fullpath

	switch ext {
	case ".zip", ".cbz":
		r, err := zip.OpenReader(fullpath)
		if err != nil {
			return nil, err
		}
		defer r.Close()

		for _, f := range r.File {
			if !isFileImage(f.Name) {
				continue
			}

			page, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open page: %w", err)
			}
			buf, err := io.ReadAll(page)
			if err != nil {
				return nil, fmt.Errorf("failed to read page: %w", err)
			}
			p.currentPages[strings.ToLower(f.Name)] = &loadedPage{
				buf: buf,
				page: &hibikemanga.ChapterPage{
					Provider: LocalProvider,
					URL:      formatUrl(f.Name),
					Index:    0, // placeholder, will be set later
					Buf:      buf,
				},
			}
		}
	case ".pdf":
		// doc, err := fitz.New(fullpath)
		// if err != nil {
		// 	return nil, fmt.Errorf("failed to open PDF file: %w", err)
		// }
		// defer doc.Close()

		// // Load images into memory
		// for n := 0; n < doc.NumPage(); n++ {
		// 	img, err := doc.Image(n)
		// 	if err != nil {
		// 		panic(err)
		// 	}

		// 	var buf bytes.Buffer
		// 	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
		// 	if err != nil {
		// 		panic(err)
		// 	}

		// 	p.currentPages[fmt.Sprintf("page_%d.jpg", n)] = &loadedPage{
		// 		buf: buf.Bytes(),
		// 		page: &hibikemanga.ChapterPage{
		// 			Provider: LocalProvider,
		// 			URL:      formatUrl(fmt.Sprintf("page_%d.jpg", n)),
		// 			Index:    n,
		// 		},
		// 	}
		// }
	default:
		// If it's a directory of images
		stat, err := os.Stat(fullpath)
		if err != nil {
			return nil, fmt.Errorf("failed to stat file: %w", err)
		}
		if !stat.IsDir() {
			return nil, fmt.Errorf("file is not a directory: %s", fullpath)
		}

		entries, err := os.ReadDir(fullpath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %w", err)
		}

		for _, entry := range entries {
			if !isFileImage(entry.Name()) {
				continue
			}

			page, err := os.Open(filepath.Join(fullpath, entry.Name()))
			if err != nil {
				return nil, fmt.Errorf("failed to open page: %w", err)
			}
			buf, err := io.ReadAll(page)
			if err != nil {
				return nil, fmt.Errorf("failed to read page: %w", err)
			}
			p.currentPages[strings.ToLower(entry.Name())] = &loadedPage{
				buf: buf,
				page: &hibikemanga.ChapterPage{
					Provider: LocalProvider,
					URL:      formatUrl(entry.Name()),
					Index:    0, // placeholder, will be set later
					Buf:      buf,
				},
			}
		}
	}

	type pageStruct struct {
		Number     float64
		LoadedPage *loadedPage
	}

	pages := make([]*pageStruct, 0)

	// Parse and order the pages
	for _, loadedPage := range p.currentPages {
		scannedPage, ok := parsePageFilename(filepath.Base(loadedPage.page.URL))
		if !ok {
			continue
		}
		pages = append(pages, &pageStruct{
			Number:     scannedPage.Number,
			LoadedPage: loadedPage,
		})
	}

	// Sort pages
	slices.SortFunc(pages, func(a, b *pageStruct) int {
		return strings.Compare(filepath.Base(a.LoadedPage.page.URL), filepath.Base(b.LoadedPage.page.URL))
	})

	ret = make([]*hibikemanga.ChapterPage, 0)
	for idx, pageStruct := range pages {
		pageStruct.LoadedPage.page.Index = idx
		ret = append(ret, pageStruct.LoadedPage.page)
	}

	return ret, nil
}

func (p *Local) ReadPage(path string) (ret io.ReadCloser, err error) {
	// e.g. path = "/series/chapter_1.cbz/image_1.jpg"

	// If the pages are already in memory, return them
	if len(p.currentPages) > 0 {
		page, ok := p.currentPages[strings.ToLower(filepath.Base(path))]
		if ok {
			return io.NopCloser(bytes.NewReader(page.buf)), nil // Return the page
		}
	}

	return nil, fmt.Errorf("page not found: %s", path)
}
