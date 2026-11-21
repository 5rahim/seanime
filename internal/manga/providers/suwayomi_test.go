package manga_providers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	hibikemanga "seanime/internal/extension/hibike/manga"
)

func TestSuwayomiProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/prefix/api/graphql" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(404)
			return
		}
		var request struct {
			Query     string                     `json:"query"`
			Variables map[string]json.RawMessage `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(request.Query, "sources {"):
			_, _ = w.Write([]byte(`{"data":{"sources":{"nodes":[{"id":"123","name":"Test","lang":"en"}]}}}`))
		case strings.Contains(request.Query, "mangas(condition"):
			_, _ = w.Write([]byte(`{"data":{"mangas":{"nodes":[{"id":7,"title":"Example"}]}}}`))
		case strings.Contains(request.Query, "fetchSourceManga"):
			var input struct {
				Source string `json:"source"`
				Query  string `json:"query"`
				Page   int    `json:"page"`
			}
			_ = json.Unmarshal(request.Variables["input"], &input)
			if input.Source != "123" || input.Query != "Example" || input.Page != 1 {
				t.Errorf("unexpected search input: %+v", input)
			}
			_, _ = w.Write([]byte(`{"data":{"fetchSourceManga":{"mangas":[{"id":7,"title":"Example"}]}}}`))
		case strings.Contains(request.Query, "fetchMangaAndChapters"):
			_, _ = w.Write([]byte(`{"data":{"fetchMangaAndChapters":{"chapters":[{"id":4,"name":"Chapter 2","chapterNumber":2,"scanlator":"Group B","sourceOrder":0},{"id":3,"name":"Chapter 1","chapterNumber":1,"scanlator":"Group A","sourceOrder":1}]}}}`))
		case strings.Contains(request.Query, "fetchChapterPages"):
			_, _ = w.Write([]byte(`{"data":{"fetchChapterPages":{"pages":["/api/v1/manga/7/chapter/0/page/0","/api/v1/manga/7/chapter/0/page/1"]}}}`))
		default:
			t.Error("unknown query")
		}
	}))
	defer server.Close()
	p, err := NewSuwayomi(server.URL+"/prefix/", "123")
	if err != nil {
		t.Fatal(err)
	}
	sources, err := p.Sources()
	if err != nil || len(sources) != 1 || sources[0].Name != "Test" {
		t.Fatalf("sources: %+v, %v", sources, err)
	}
	results, err := p.Search(hibikemanga.SearchOptions{Query: "Example"})
	if err != nil || len(results) != 1 || results[0].ID != "7" {
		t.Fatalf("results: %+v, %v", results, err)
	}
	if results[0].Image != server.URL+"/prefix/api/v1/manga/7/thumbnail" {
		t.Fatal(results[0].Image)
	}
	chapters, err := p.FindChapters("7")
	if err != nil || len(chapters) != 2 || chapters[0].ID != "3" || chapters[0].Scanlator != "Group A" || chapters[1].Chapter != "2" {
		t.Fatalf("chapters: %+v, %v", chapters, err)
	}
	pages, err := p.FindChapterPages("4")
	if err != nil || len(pages) != 2 || pages[0].URL != server.URL+"/prefix/api/v1/manga/7/chapter/0/page/0" || pages[1].Index != 1 {
		t.Fatalf("pages: %+v, %v", pages, err)
	}
	if _, err := p.FindChapters("../7"); err == nil {
		t.Fatal("invalid manga ID accepted")
	}
}

func TestSuwayomiUsesSavedMangaWhenSourceIsUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&request)
		switch {
		case strings.Contains(request.Query, "mangas(condition"):
			_, _ = w.Write([]byte(`{"data":{"mangas":{"nodes":[{"id":7,"title":"Saved"}]}}}`))
		case strings.Contains(request.Query, "manga(id:"):
			_, _ = w.Write([]byte(`{"data":{"manga":{"chapters":{"nodes":[{"id":3,"name":"Chapter 1","chapterNumber":1,"scanlator":"Group A","sourceOrder":0}]}}}}`))
		default:
			_, _ = w.Write([]byte(`{"errors":[{"message":"source unavailable"}]}`))
		}
	}))
	defer server.Close()
	p, err := NewSuwayomi(server.URL, "123")
	if err != nil {
		t.Fatal(err)
	}
	results, err := p.Search(hibikemanga.SearchOptions{Query: "Saved"})
	if err != nil || len(results) != 1 || results[0].ID != "7" {
		t.Fatalf("saved search: %+v, %v", results, err)
	}
	chapters, err := p.FindChapters("7")
	if err != nil || len(chapters) != 1 || chapters[0].Scanlator != "Group A" {
		t.Fatalf("saved chapters: %+v, %v", chapters, err)
	}
}

func TestSuwayomiRejectsBadResponses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errors":[{"message":"source unavailable"}]}`))
	}))
	defer server.Close()
	p, err := NewSuwayomi(server.URL, "123")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Sources(); err == nil || !strings.Contains(err.Error(), "source unavailable") {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, raw := range []string{"file:///tmp/a", "http://user:password@localhost", "https://localhost/?key=x"} {
		if _, err := NewSuwayomi(raw, "123"); err == nil {
			t.Fatalf("accepted %q", raw)
		}
	}
}
