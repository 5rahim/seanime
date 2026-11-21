package manga_providers

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	hibikemanga "seanime/internal/extension/hibike/manga"
)

// Suwayomi represents one source installed on a Suwayomi server.
type Suwayomi struct {
	baseURL    *url.URL
	sourceID   string
	providerID string
	client     *http.Client
}

type SuwayomiSource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Lang string `json:"lang"`
}

type suwayomiChapter struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Number      float64 `json:"chapterNumber"`
	Scanlator   *string `json:"scanlator"`
	SourceOrder int     `json:"sourceOrder"`
}

type suwayomiManga struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func NewSuwayomi(rawURL, sourceID string) (*Suwayomi, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u == nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, fmt.Errorf("invalid Suwayomi server URL")
	}
	u.Path = strings.TrimRight(u.Path, "/")
	u.RawPath = ""
	sum := sha256.Sum256([]byte(u.String()))
	id := "suwayomi-" + hex.EncodeToString(sum[:6]) + "-" + sourceID
	return &Suwayomi{baseURL: u, sourceID: sourceID, providerID: id, client: &http.Client{Timeout: 45 * time.Second}}, nil
}

func (p *Suwayomi) ID() string { return p.providerID }

func (p *Suwayomi) endpoint(path string) string {
	u := *p.baseURL
	u.Path = strings.TrimRight(u.Path, "/") + "/" + strings.TrimLeft(path, "/")
	return u.String()
}

func (p *Suwayomi) graphql(query string, variables any, data any) error {
	body, err := json.Marshal(map[string]any{"query": query, "variables": variables})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, p.endpoint("api/graphql"), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("Suwayomi request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Suwayomi returned HTTP %d", resp.StatusCode)
	}
	var result struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, 16<<20)).Decode(&result); err != nil {
		return fmt.Errorf("invalid Suwayomi response: %w", err)
	}
	if len(result.Errors) != 0 {
		return fmt.Errorf("Suwayomi: %s", result.Errors[0].Message)
	}
	if len(result.Data) == 0 || string(result.Data) == "null" {
		return fmt.Errorf("Suwayomi returned no data")
	}
	return json.Unmarshal(result.Data, data)
}

func (p *Suwayomi) Sources() ([]SuwayomiSource, error) {
	var data struct {
		Sources struct {
			Nodes []SuwayomiSource `json:"nodes"`
		} `json:"sources"`
	}
	err := p.graphql(`query { sources { nodes { id name lang } } }`, nil, &data)
	return data.Sources.Nodes, err
}

func (p *Suwayomi) GetSettings() hibikemanga.Settings {
	return hibikemanga.Settings{SupportsMultiScanlator: true}
}

func (p *Suwayomi) Search(opts hibikemanga.SearchOptions) ([]*hibikemanga.SearchResult, error) {
	var saved struct {
		Mangas struct {
			Nodes []suwayomiManga `json:"nodes"`
		} `json:"mangas"`
	}
	savedErr := p.graphql(`query ($source: LongString!, $query: String!) { mangas(condition: { sourceId: $source, inLibrary: true }, filter: { title: { includesInsensitive: $query } }) { nodes { id title } } }`, map[string]any{"source": p.sourceID, "query": opts.Query}, &saved)
	var data struct {
		Fetch *struct {
			Mangas []suwayomiManga `json:"mangas"`
		} `json:"fetchSourceManga"`
	}
	err := p.graphql(`mutation ($input: FetchSourceMangaInput!) { fetchSourceManga(input: $input) { mangas { id title } } }`, map[string]any{"input": map[string]any{"source": p.sourceID, "type": "SEARCH", "page": 1, "query": opts.Query}}, &data)
	if err != nil && (savedErr != nil || len(saved.Mangas.Nodes) == 0) {
		return nil, err
	}
	if data.Fetch == nil && len(saved.Mangas.Nodes) == 0 {
		return nil, fmt.Errorf("Suwayomi returned no search results")
	}
	mangas := saved.Mangas.Nodes
	if data.Fetch != nil {
		mangas = append(mangas, data.Fetch.Mangas...)
	}
	results := make([]*hibikemanga.SearchResult, 0, len(mangas))
	seen := make(map[int]bool, len(mangas))
	for _, m := range mangas {
		if seen[m.ID] {
			continue
		}
		seen[m.ID] = true
		results = append(results, &hibikemanga.SearchResult{Provider: p.ID(), ID: strconv.Itoa(m.ID), Title: m.Title, Image: p.endpoint(fmt.Sprintf("api/v1/manga/%d/thumbnail", m.ID)), ImageHeaders: map[string]string{"X-Seanime-Suwayomi": "1"}})
	}
	return results, nil
}

func (p *Suwayomi) FindChapters(id string) ([]*hibikemanga.ChapterDetails, error) {
	mangaID, err := strconv.Atoi(id)
	if err != nil || mangaID < 1 {
		return nil, fmt.Errorf("invalid Suwayomi manga ID")
	}
	var data struct {
		Fetch *struct {
			Chapters []suwayomiChapter `json:"chapters"`
		} `json:"fetchMangaAndChapters"`
	}
	err = p.graphql(`mutation ($input: FetchMangaAndChaptersInput!) { fetchMangaAndChapters(input: $input) { chapters { id name chapterNumber scanlator sourceOrder } } }`, map[string]any{"input": map[string]any{"id": mangaID, "fetchManga": false, "fetchChapters": true}}, &data)
	var sourceChapters []suwayomiChapter
	if err == nil && data.Fetch != nil {
		sourceChapters = data.Fetch.Chapters
	} else {
		// Downloaded chapters remain readable when an extension is temporarily unavailable.
		var cached struct {
			Manga *struct {
				Chapters struct {
					Nodes []suwayomiChapter `json:"nodes"`
				} `json:"chapters"`
			} `json:"manga"`
		}
		cacheErr := p.graphql(`query ($id: Int!) { manga(id: $id) { chapters { nodes { id name chapterNumber scanlator sourceOrder } } } }`, map[string]any{"id": mangaID}, &cached)
		if cacheErr != nil || cached.Manga == nil || len(cached.Manga.Chapters.Nodes) == 0 {
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("Suwayomi returned no chapters")
		}
		sourceChapters = cached.Manga.Chapters.Nodes
	}
	sort.SliceStable(sourceChapters, func(i, j int) bool { return sourceChapters[i].SourceOrder > sourceChapters[j].SourceOrder })
	chapters := make([]*hibikemanga.ChapterDetails, 0, len(sourceChapters))
	for i, c := range sourceChapters {
		number := ""
		if c.Number >= 0 {
			number = strconv.FormatFloat(c.Number, 'f', -1, 64)
		}
		scanlator := ""
		if c.Scanlator != nil {
			scanlator = *c.Scanlator
		}
		chapters = append(chapters, &hibikemanga.ChapterDetails{Provider: p.ID(), ID: strconv.Itoa(c.ID), Title: c.Name, Chapter: number, Index: uint(i), Scanlator: scanlator})
	}
	return chapters, nil
}

func (p *Suwayomi) FindChapterPages(id string) ([]*hibikemanga.ChapterPage, error) {
	chapterID, err := strconv.Atoi(id)
	if err != nil || chapterID < 1 {
		return nil, fmt.Errorf("invalid Suwayomi chapter ID")
	}
	var data struct {
		Fetch *struct {
			Pages []string `json:"pages"`
		} `json:"fetchChapterPages"`
	}
	err = p.graphql(`mutation ($input: FetchChapterPagesInput!) { fetchChapterPages(input: $input) { pages } }`, map[string]any{"input": map[string]any{"chapterId": chapterID}}, &data)
	if err != nil {
		return nil, err
	}
	if data.Fetch == nil {
		return nil, fmt.Errorf("Suwayomi returned no pages")
	}
	pages := make([]*hibikemanga.ChapterPage, 0, len(data.Fetch.Pages))
	for i, page := range data.Fetch.Pages {
		if !strings.HasPrefix(page, "/api/v1/manga/") || strings.Contains(page, "..") {
			return nil, fmt.Errorf("invalid Suwayomi page path")
		}
		pages = append(pages, &hibikemanga.ChapterPage{Provider: p.ID(), URL: p.endpoint(page), Index: i, Headers: map[string]string{"X-Seanime-Suwayomi": "1"}})
	}
	return pages, nil
}
