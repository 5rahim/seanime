package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"seanime/internal/api/anilist"
	chapter_downloader "seanime/internal/manga/downloader"

	"github.com/labstack/echo/v4"
)

type mihonManga struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Author       string `json:"author"`
	Artist       string `json:"artist"`
	Description  string `json:"description"`
	CoverURL     string `json:"cover_url"`
	Status       int    `json:"status"`
	ChapterCount int    `json:"chapter_count"`
	Genres       string `json:"genres"`
}

type mihonChapter struct {
	Dir       string  `json:"dir"`
	Number    float64 `json:"number"`
	Title     string  `json:"title"`
	PageCount int     `json:"page_count"`
}

type mihonPage struct {
	Index int    `json:"index"`
	URL   string `json:"url"`
}

// HandleMihonLibrary returns all manga that have downloaded chapters.
//
//	@summary returns the manga library for Mihon consumption.
//	@route /api/v1/mihon/library [GET]
//	@returns []mihonManga
func (h *Handler) HandleMihonLibrary(c echo.Context) error {
	downloadDir := h.App.Config.Manga.DownloadDir
	if downloadDir == "" {
		return h.RespondWithData(c, []mihonManga{})
	}

	entries, err := os.ReadDir(downloadDir)
	if err != nil {
		return h.RespondWithData(c, []mihonManga{})
	}

	type mangaInfo struct {
		mediaId      int
		chapterCount int
	}
	mangaMap := make(map[int]*mangaInfo)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		did, ok := chapter_downloader.ParseChapterDirName(entry.Name())
		if !ok {
			continue
		}
		if m, exists := mangaMap[did.MediaId]; exists {
			m.chapterCount++
		} else {
			mangaMap[did.MediaId] = &mangaInfo{
				mediaId:      did.MediaId,
				chapterCount: 1,
			}
		}
	}

	if len(mangaMap) == 0 {
		return h.RespondWithData(c, []mihonManga{})
	}

	metadataMap := h.buildAniListMetadataMap()

	query := strings.ToLower(c.QueryParam("q"))

	mangas := make([]mihonManga, 0, len(mangaMap))
	for mediaId, info := range mangaMap {
		m := mihonManga{
			ID:           mediaId,
			Title:        fmt.Sprintf("Manga #%d", mediaId),
			ChapterCount: info.chapterCount,
		}

		if meta, ok := metadataMap[mediaId]; ok {
			m.Title = meta.title
			m.Author = meta.author
			m.Artist = meta.artist
			m.Description = meta.description
			m.CoverURL = meta.coverURL
			m.Status = meta.status
			m.Genres = meta.genres
		}

		if query != "" && !strings.Contains(strings.ToLower(m.Title), query) {
			continue
		}

		mangas = append(mangas, m)
	}

	sort.Slice(mangas, func(i, j int) bool {
		return mangas[i].Title < mangas[j].Title
	})

	return h.RespondWithData(c, mangas)
}

// HandleMihonMangaDetails returns details for a single manga.
//
//	@summary returns manga details for Mihon.
//	@route /api/v1/mihon/manga/:id [GET]
//	@returns mihonManga
func (h *Handler) HandleMihonMangaDetails(c echo.Context) error {
	mediaId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return h.RespondWithError(c, fmt.Errorf("invalid manga id"))
	}

	downloadDir := h.App.Config.Manga.DownloadDir
	chapterCount := 0

	if downloadDir != "" {
		entries, _ := os.ReadDir(downloadDir)
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			did, ok := chapter_downloader.ParseChapterDirName(entry.Name())
			if !ok {
				continue
			}
			if did.MediaId == mediaId {
				chapterCount++
			}
		}
	}

	m := mihonManga{
		ID:           mediaId,
		Title:        fmt.Sprintf("Manga #%d", mediaId),
		ChapterCount: chapterCount,
	}

	metadataMap := h.buildAniListMetadataMap()
	if meta, ok := metadataMap[mediaId]; ok {
		m.Title = meta.title
		m.Author = meta.author
		m.Artist = meta.artist
		m.Description = meta.description
		m.CoverURL = meta.coverURL
		m.Status = meta.status
		m.Genres = meta.genres
	}

	return h.RespondWithData(c, m)
}

// HandleMihonMangaChapters returns downloaded chapters for a manga.
//
//	@summary returns the chapter list for Mihon.
//	@route /api/v1/mihon/manga/:id/chapters [GET]
//	@returns []mihonChapter
func (h *Handler) HandleMihonMangaChapters(c echo.Context) error {
	mediaId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return h.RespondWithError(c, fmt.Errorf("invalid manga id"))
	}

	downloadDir := h.App.Config.Manga.DownloadDir
	if downloadDir == "" {
		return h.RespondWithData(c, []mihonChapter{})
	}

	entries, err := os.ReadDir(downloadDir)
	if err != nil {
		return h.RespondWithData(c, []mihonChapter{})
	}

	chapters := make([]mihonChapter, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		did, ok := chapter_downloader.ParseChapterDirName(entry.Name())
		if !ok || did.MediaId != mediaId {
			continue
		}

		chapterNum, _ := strconv.ParseFloat(did.ChapterNumber, 64)
		pageCount := countPagesInDir(filepath.Join(downloadDir, entry.Name()))

		chapters = append(chapters, mihonChapter{
			Dir:       entry.Name(),
			Number:    chapterNum,
			Title:     fmt.Sprintf("Chapter %s", did.ChapterNumber),
			PageCount: pageCount,
		})
	}

	sort.Slice(chapters, func(i, j int) bool {
		return chapters[i].Number < chapters[j].Number
	})

	return h.RespondWithData(c, chapters)
}

// HandleMihonChapterPages returns pages for a downloaded chapter.
//
//	@summary returns the page list for a chapter for Mihon.
//	@route /api/v1/mihon/chapter/:dir/pages [GET]
//	@returns []mihonPage
func (h *Handler) HandleMihonChapterPages(c echo.Context) error {
	dirName := c.Param("dir")

	if _, ok := chapter_downloader.ParseChapterDirName(dirName); !ok {
		return h.RespondWithError(c, fmt.Errorf("invalid chapter directory"))
	}

	downloadDir := h.App.Config.Manga.DownloadDir
	if downloadDir == "" {
		return h.RespondWithError(c, fmt.Errorf("manga downloads not configured"))
	}

	chapterPath := filepath.Join(downloadDir, dirName)

	rel, err := filepath.Rel(downloadDir, chapterPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return h.RespondWithError(c, fmt.Errorf("invalid chapter directory"))
	}

	if _, err := os.Stat(chapterPath); os.IsNotExist(err) {
		return h.RespondWithError(c, fmt.Errorf("chapter not found"))
	}

	registryPath := filepath.Join(chapterPath, "registry.json")
	pages, err := readRegistryPages(registryPath, dirName)
	if err != nil {
		pages = buildPagesFromFiles(chapterPath, dirName)
	}

	return h.RespondWithData(c, pages)
}

type anilistMeta struct {
	title       string
	author      string
	artist      string
	description string
	coverURL    string
	status      int
	genres      string
}

func (h *Handler) buildAniListMetadataMap() map[int]*anilistMeta {
	metadataMap := make(map[int]*anilistMeta)

	collection, err := h.App.GetMangaCollection(false)
	if err != nil || collection == nil {
		return metadataMap
	}

	lists := collection.GetMediaListCollection()
	if lists == nil {
		return metadataMap
	}

	for _, list := range lists.GetLists() {
		for _, entry := range list.GetEntries() {
			media := entry.GetMedia()
			if media == nil {
				continue
			}

			meta := &anilistMeta{
				title: fmt.Sprintf("Manga #%d", media.GetID()),
			}

			if t := media.GetTitle(); t != nil {
				if p := t.GetUserPreferred(); p != nil {
					meta.title = *p
				}
			}

			if ci := media.GetCoverImage(); ci != nil {
				if xl := ci.GetExtraLarge(); xl != nil {
					meta.coverURL = *xl
				}
			}

			if d := media.GetDescription(); d != nil {
				meta.description = *d
			}

			meta.status = mapAniListStatus(media.GetStatus())

			if genres := media.Genres; len(genres) > 0 {
				parts := make([]string, 0, len(genres))
				for _, g := range genres {
					if g != nil {
						parts = append(parts, *g)
					}
				}
				meta.genres = strings.Join(parts, ", ")
			}

			metadataMap[media.GetID()] = meta
		}
	}

	return metadataMap
}

func mapAniListStatus(status *anilist.MediaStatus) int {
	if status == nil {
		return 0
	}
	switch *status {
	case anilist.MediaStatusReleasing:
		return 1 // ONGOING
	case anilist.MediaStatusFinished:
		return 2 // COMPLETED
	case anilist.MediaStatusCancelled:
		return 5 // CANCELLED
	case anilist.MediaStatusHiatus:
		return 6 // ON_HIATUS
	default:
		return 0 // UNKNOWN
	}
}

type registryEntry struct {
	Index    int    `json:"index"`
	Filename string `json:"filename"`
}

func readRegistryPages(registryPath string, dirName string) ([]mihonPage, error) {
	data, err := os.ReadFile(registryPath)
	if err != nil {
		return nil, err
	}

	var registry map[string]registryEntry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, err
	}

	pages := make([]mihonPage, 0, len(registry))
	for _, entry := range registry {
		pages = append(pages, mihonPage{
			Index: entry.Index,
			URL:   fmt.Sprintf("/manga-downloads/%s/%s", dirName, entry.Filename),
		})
	}

	sort.Slice(pages, func(i, j int) bool {
		return pages[i].Index < pages[j].Index
	})

	return pages, nil
}

func buildPagesFromFiles(chapterPath string, dirName string) []mihonPage {
	files, err := os.ReadDir(chapterPath)
	if err != nil {
		return []mihonPage{}
	}

	pages := make([]mihonPage, 0)
	idx := 0
	for _, f := range files {
		if f.IsDir() || f.Name() == "registry.json" {
			continue
		}
		ext := strings.ToLower(filepath.Ext(f.Name()))
		if ext == ".webp" || ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".avif" {
			pages = append(pages, mihonPage{
				Index: idx,
				URL:   fmt.Sprintf("/manga-downloads/%s/%s", dirName, f.Name()),
			})
			idx++
		}
	}

	sort.Slice(pages, func(i, j int) bool {
		return pages[i].Index < pages[j].Index
	})

	return pages
}

func countPagesInDir(dirPath string) int {
	registryPath := filepath.Join(dirPath, "registry.json")
	data, err := os.ReadFile(registryPath)
	if err != nil {
		files, err := os.ReadDir(dirPath)
		if err != nil {
			return 0
		}
		count := 0
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(f.Name()))
			if ext == ".webp" || ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".avif" {
				count++
			}
		}
		return count
	}

	var registry map[string]registryEntry
	if err := json.Unmarshal(data, &registry); err != nil {
		return 0
	}
	return len(registry)
}
