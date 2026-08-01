package testmocks

import (
	"errors"
	"sync"

	hibikemanga "seanime/internal/extension/hibike/manga"
)

type FakeMangaProvider struct {
	mu            sync.Mutex
	searchResults []*hibikemanga.SearchResult
	chapters      map[string][]*hibikemanga.ChapterDetails
	searchErr     error
	chaptersErr   error
	searchCalls   int
	chapterCalls  int
	settings      hibikemanga.Settings
}

type FakeMangaProviderBuilder struct {
	provider *FakeMangaProvider
}

func NewFakeMangaProviderBuilder() *FakeMangaProviderBuilder {
	return &FakeMangaProviderBuilder{provider: &FakeMangaProvider{
		chapters: make(map[string][]*hibikemanga.ChapterDetails),
	}}
}

func (b *FakeMangaProviderBuilder) WithSearchResults(results ...*hibikemanga.SearchResult) *FakeMangaProviderBuilder {
	b.provider.searchResults = results
	return b
}

func (b *FakeMangaProviderBuilder) WithChapters(id string, chapters ...*hibikemanga.ChapterDetails) *FakeMangaProviderBuilder {
	b.provider.chapters[id] = chapters
	return b
}

func (b *FakeMangaProviderBuilder) WithSearchError(err error) *FakeMangaProviderBuilder {
	b.provider.searchErr = err
	return b
}

func (b *FakeMangaProviderBuilder) WithChapterError(err error) *FakeMangaProviderBuilder {
	b.provider.chaptersErr = err
	return b
}

func (b *FakeMangaProviderBuilder) WithSettings(settings hibikemanga.Settings) *FakeMangaProviderBuilder {
	b.provider.settings = settings
	return b
}

func (b *FakeMangaProviderBuilder) Build() *FakeMangaProvider {
	return b.provider
}

func (p *FakeMangaProvider) Search(hibikemanga.SearchOptions) ([]*hibikemanga.SearchResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.searchCalls++
	if p.searchErr != nil {
		return nil, p.searchErr
	}
	return append([]*hibikemanga.SearchResult(nil), p.searchResults...), nil
}

func (p *FakeMangaProvider) FindChapters(id string) ([]*hibikemanga.ChapterDetails, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.chapterCalls++
	if p.chaptersErr != nil {
		return nil, p.chaptersErr
	}
	chapters, found := p.chapters[id]
	if !found {
		return nil, errors.New("chapters not found")
	}
	return append([]*hibikemanga.ChapterDetails(nil), chapters...), nil
}

func (p *FakeMangaProvider) FindChapterPages(string) ([]*hibikemanga.ChapterPage, error) {
	return nil, nil
}

func (p *FakeMangaProvider) GetSettings() hibikemanga.Settings {
	return p.settings
}

func (p *FakeMangaProvider) SearchCalls() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.searchCalls
}

func (p *FakeMangaProvider) ChapterCalls() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.chapterCalls
}

func (p *FakeMangaProvider) SetChapters(id string, chapters ...*hibikemanga.ChapterDetails) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.chapters[id] = chapters
}

func (p *FakeMangaProvider) SetChapterError(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.chaptersErr = err
}
