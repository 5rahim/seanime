package customsource

import (
	"context"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/extension"
	hibikecustomsource "seanime/internal/extension/hibike/customsource"
	"seanime/internal/testutil"
	"seanime/internal/util"
	"testing"

	"github.com/stretchr/testify/require"
)

// Verifies that custom source IDs round-trip through the pack/unpack helpers
// and that plain AniList IDs are not mistaken for custom source IDs.
func TestMediaIDHelpers(t *testing.T) {
	// Custom source IDs encode the extension identifier and the provider-local media ID into one runtime ID.
	mediaID := GenerateMediaId(321, 987654)
	require.True(t, IsExtensionId(mediaID))

	extensionIdentifier, localID := ExtractExtensionData(mediaID)
	require.Equal(t, 321, extensionIdentifier)
	require.Equal(t, 987654, localID)

	require.False(t, IsExtensionId(12345))
	plainExt, plainLocal := ExtractExtensionData(12345)
	require.Zero(t, plainExt)
	require.Zero(t, plainLocal)
}

// Verifies how custom source URLs are tagged and recovered.
// This protects the split between extension-owned URLs and normal AniList URLs.
func TestSiteURLHelpers(t *testing.T) {
	t.Run("formats nil and custom urls", func(t *testing.T) {
		// This covers the two custom-source formats we care about:
		// a synthetic source-only URL when there is no site URL,
		// and a tagged URL when the provider returns its own link.
		formattedNil := formatSiteUrl("demo", nil)
		require.NotNil(t, formattedNil)
		require.Equal(t, "ext_custom_source_demo", *formattedNil)

		extIDNil, okNil := GetCustomSourceExtensionIdFromSiteUrl(formattedNil)
		require.True(t, okNil)
		require.Equal(t, "demo", extIDNil)

		customURL := "https://example.com/item"
		formattedCustom := formatSiteUrl("demo", &customURL)
		require.NotNil(t, formattedCustom)
		require.Equal(t, "ext_custom_source_demo|END|https://example.com/item", *formattedCustom)

		extID, ok := GetCustomSourceExtensionIdFromSiteUrl(formattedCustom)
		require.True(t, ok)
		require.Equal(t, "demo", extID)
	})

	t.Run("keeps AniList urls untouched", func(t *testing.T) {
		// AniList URLs are intentionally left alone so downstream code can still treat them as native AniList media.
		aniListURL := "https://anilist.co/anime/1"
		formatted := formatSiteUrl("demo", &aniListURL)
		require.Same(t, &aniListURL, formatted)

		extID, ok := GetCustomSourceExtensionIdFromSiteUrl(formatted)
		require.False(t, ok)
		require.Empty(t, extID)
	})
}

// Verifies that normalization rewrites provider media into the runtime shape
// the rest of the app expects, including IDs, tagged URLs, and title fallbacks.
func TestNormalizeMedia(t *testing.T) {
	t.Run("normalizes anime ids urls and title fallback", func(t *testing.T) {
		// Normalization rewrites both the ID and the site URL so the rest of the app can tell this apart from AniList media.
		anime := &anilist.BaseAnime{
			ID:      25,
			SiteURL: new("https://example.com/anime/25"),
			Title: &anilist.BaseAnime_Title{
				English: new("Fresh Anime"),
			},
		}

		NormalizeMedia(17, "demo", anime)

		require.Equal(t, GenerateMediaId(17, 25), anime.ID)
		require.NotNil(t, anime.SiteURL)
		require.Equal(t, "ext_custom_source_demo|END|https://example.com/anime/25", *anime.SiteURL)
		require.NotNil(t, anime.Title)
		require.Equal(t, "Fresh Anime", *anime.Title.UserPreferred)
		_, ok := GetCustomSourceExtensionIdFromSiteUrl(anime.SiteURL)
		require.True(t, ok)
	})

	t.Run("fills missing manga title", func(t *testing.T) {
		// Providers are allowed to omit titles, but the app expects something printable.
		manga := &anilist.BaseManga{ID: 30}

		NormalizeMedia(21, "reader", manga)

		require.Equal(t, GenerateMediaId(21, 30), manga.ID)
		require.NotNil(t, manga.SiteURL)
		require.Equal(t, "ext_custom_source_reader", *manga.SiteURL)
		require.NotNil(t, manga.Title)
		require.Equal(t, "???", *manga.Title.UserPreferred)
		require.Equal(t, "???", *manga.Title.English)
	})
}

// Verifies that the manager can resolve a provider from all supported entry points:
// direct media ID lookup, BaseAnime lookup, and the missing-provider path.
func TestManagerProviderResolution(t *testing.T) {
	provider := &fakeCustomSourceProvider{extensionIdentifier: 7}
	manager := newCustomSourceTestManager(t, customSourceTestExtension{
		id:         "demo-ext",
		identifier: 7,
		provider:   provider,
	})

	customID := GenerateMediaId(7, 55)
	ext, localID, isCustom, exists := manager.GetProviderFromId(customID)
	require.True(t, isCustom)
	require.True(t, exists)
	require.NotNil(t, ext)
	require.Equal(t, 55, localID)
	require.Equal(t, "demo-ext", ext.GetID())

	missingExt, missingLocalID, missingIsCustom, missingExists := manager.GetProviderFromId(GenerateMediaId(9, 11))
	require.True(t, missingIsCustom)
	require.False(t, missingExists)
	require.Nil(t, missingExt)
	require.Zero(t, missingLocalID)

	baseAnimeExt, animeLocalID, animeIsCustom, animeExists := manager.GetProviderFromBaseAnime(&anilist.BaseAnime{ID: customID})
	require.True(t, animeIsCustom)
	require.True(t, animeExists)
	require.NotNil(t, baseAnimeExt)
	require.Equal(t, 55, animeLocalID)

	baseMangaExt, mangaLocalID, mangaIsCustom, mangaExists := manager.GetProviderFromBaseManga(&anilist.BaseManga{ID: 123})
	require.False(t, mangaIsCustom)
	require.False(t, mangaExists)
	require.Nil(t, baseMangaExt)
	require.Zero(t, mangaLocalID)
}

// Verifies that stored anime entries are refreshed against the live provider on read
// and that entries for extensions that are no longer loaded are filtered out.
func TestGetCustomSourceAnimeEntriesRefreshesMedia(t *testing.T) {
	provider := &fakeCustomSourceProvider{
		extensionIdentifier: 13,
		animeByID: map[int]*anilist.BaseAnime{
			101: newBaseAnime(101, "Fresh Title", "https://example.com/fresh"),
		},
	}
	manager := newCustomSourceTestManager(t, customSourceTestExtension{
		id:         "anime-ext",
		identifier: 13,
		provider:   provider,
	})

	require.NoError(t, manager.SaveCustomSourceAnimeEntries("anime-ext", map[int]*anilist.AnimeListEntry{
		101: {
			ID:     101,
			Status: new(anilist.MediaListStatusCurrent),
			Media:  newBaseAnime(101, "Stale Title", "https://example.com/stale"),
		},
	}))
	require.NoError(t, manager.SaveCustomSourceAnimeEntries("missing-ext", map[int]*anilist.AnimeListEntry{
		9: {ID: 9, Media: newBaseAnime(9, "Ghost", "https://example.com/ghost")},
	}))

	// Stored entries are refreshed from the extension on read so stale media data does not leak back into the collection.
	entries, ok := manager.GetCustomSourceAnimeEntries()
	require.True(t, ok)
	require.Contains(t, entries, "anime-ext")
	require.NotContains(t, entries, "missing-ext")
	require.Equal(t, "Fresh Title", *entries["anime-ext"][101].Media.Title.English)
	require.Equal(t, "https://example.com/fresh", *entries["anime-ext"][101].Media.SiteURL)
}

// Verifies the main anime mutation lifecycle end to end:
// create an entry, update progress, auto-complete at the total count, update repeat, then delete it.
func TestUpdateEntryAnimeLifecycle(t *testing.T) {
	provider := &fakeCustomSourceProvider{
		extensionIdentifier: 3,
		animeByID: map[int]*anilist.BaseAnime{
			77: newBaseAnime(77, "Tracked Anime", "https://example.com/anime/77"),
		},
	}
	manager := newCustomSourceTestManager(t, customSourceTestExtension{
		id:         "tracker",
		identifier: 3,
		provider:   provider,
	})

	mediaID := GenerateMediaId(3, 77)
	status := anilist.MediaListStatusPlanning
	score := 84
	progress := 6
	startedAt := &anilist.FuzzyDateInput{Year: new(2024), Month: new(2), Day: new(10)}
	completedAt := &anilist.FuzzyDateInput{Year: new(2024), Month: new(3), Day: new(1)}

	// This walks the main mutation flow: create the entry, advance progress, bump repeat count, then remove it.
	require.NoError(t, manager.UpdateEntry(context.Background(), mediaID, &status, &score, &progress, startedAt, completedAt))

	entries, ok := manager.GetCustomSourceAnimeEntries()
	require.True(t, ok)
	entry, found := entries["tracker"][77]
	require.True(t, found)
	require.Equal(t, status, *entry.Status)
	require.Equal(t, 84.0, *entry.Score)
	require.Equal(t, 6, *entry.Progress)
	require.Equal(t, 2024, *entry.StartedAt.Year)
	require.Equal(t, 1, *entry.CompletedAt.Day)

	require.NoError(t, manager.UpdateEntryProgress(context.Background(), mediaID, 12, new(12)))
	entries, ok = manager.GetCustomSourceAnimeEntries()
	require.True(t, ok)
	entry = entries["tracker"][77]
	require.Equal(t, anilist.MediaListStatusCompleted, *entry.Status)
	require.Equal(t, 12, *entry.Progress)

	require.NoError(t, manager.UpdateEntryRepeat(context.Background(), mediaID, 2))
	entries, ok = manager.GetCustomSourceAnimeEntries()
	require.True(t, ok)
	entry = entries["tracker"][77]
	require.Equal(t, 2, *entry.Repeat)

	require.NoError(t, manager.DeleteEntry(context.Background(), mediaID, 0))
	entries, ok = manager.GetCustomSourceAnimeEntries()
	require.True(t, ok)
	_, found = entries["tracker"]
	require.False(t, found)
}

// Verifies the type-detection fallback inside UpdateEntry.
// If anime lookup does not find anything, the manager should create a manga entry instead of failing.
func TestUpdateEntryCreatesMangaEntryWhenAnimeLookupMisses(t *testing.T) {
	provider := &fakeCustomSourceProvider{
		extensionIdentifier: 5,
		mangaByID: map[int]*anilist.BaseManga{
			88: newBaseManga(88, "Tracked Manga", "https://example.com/manga/88"),
		},
	}
	manager := newCustomSourceTestManager(t, customSourceTestExtension{
		id:         "reader",
		identifier: 5,
		provider:   provider,
	})

	mediaID := GenerateMediaId(5, 88)
	status := anilist.MediaListStatusCurrent
	progress := 14

	// UpdateEntry tries anime first, then falls back to manga when the anime lookup does not return anything.
	require.NoError(t, manager.UpdateEntry(context.Background(), mediaID, &status, nil, &progress, nil, nil))

	entries, ok := manager.GetCustomSourceMangaCollection()
	require.True(t, ok)
	entry, found := entries["reader"][88]
	require.True(t, found)
	require.Equal(t, status, *entry.Status)
	require.Equal(t, 14, *entry.Progress)
	require.Equal(t, "Tracked Manga", *entry.Media.Title.English)
}

// Verifies that custom source anime entries are merged into an existing AniList collection
// under the right status list and with normalized runtime media data.
func TestMergeAnimeEntries(t *testing.T) {
	provider := &fakeCustomSourceProvider{
		extensionIdentifier: 11,
		animeByID: map[int]*anilist.BaseAnime{
			41: newBaseAnime(41, "Merged Anime", "https://example.com/merged"),
		},
	}
	manager := newCustomSourceTestManager(t, customSourceTestExtension{
		id:         "merge-ext",
		identifier: 11,
		provider:   provider,
	})

	require.NoError(t, manager.SaveCustomSourceAnimeEntries("merge-ext", map[int]*anilist.AnimeListEntry{
		41: {
			ID:       41,
			Status:   new(anilist.MediaListStatusCurrent),
			Progress: new(4),
			Media:    newBaseAnime(41, "Stored Anime", "https://example.com/stored"),
		},
	}))

	collection := &anilist.AnimeCollection{
		MediaListCollection: &anilist.AnimeCollection_MediaListCollection{
			Lists: []*anilist.AnimeCollection_MediaListCollection_Lists{{
				Status:  new(anilist.MediaListStatusPlanning),
				Entries: []*anilist.AnimeCollection_MediaListCollection_Lists_Entries{},
			}},
		},
	}

	// Merge keeps the existing AniList lists and appends a generated list for the custom source status bucket.
	manager.MergeAnimeEntries(collection)

	require.Len(t, collection.MediaListCollection.Lists, 2)
	currentList := findAnimeListByStatus(t, collection, anilist.MediaListStatusCurrent)
	require.Len(t, currentList.Entries, 1)
	require.Equal(t, GenerateMediaId(11, 41), currentList.Entries[0].ID)
	require.Equal(t, GenerateMediaId(11, 41), currentList.Entries[0].Media.ID)
	require.Equal(t, "ext_custom_source_merge-ext|END|https://example.com/merged", *currentList.Entries[0].Media.SiteURL)
	require.Equal(t, "Merged Anime", *currentList.Entries[0].Media.Title.UserPreferred)
}

type customSourceTestExtension struct {
	id         string
	identifier int
	provider   *fakeCustomSourceProvider
}

func newCustomSourceTestManager(t *testing.T, exts ...customSourceTestExtension) *Manager {
	t.Helper()

	env := testutil.NewTestEnv(t)
	bank := extension.NewUnifiedBank()
	// The real manager listens to the unified extension bank, so the test wrapper builds the same wiring with a temp DB.
	for _, spec := range exts {
		ext := extension.NewCustomSourceExtension(&extension.Extension{
			ID:          spec.id,
			Name:        spec.id,
			Version:     "1.0.0",
			ManifestURI: "builtin",
			Language:    extension.LanguageGo,
			Type:        extension.TypeCustomSource,
		}, spec.provider)
		ext.SetExtensionIdentifier(spec.identifier)
		bank.Set(spec.id, ext)
	}

	manager := NewManager(util.NewRef(bank), env.NewDatabase(""), env.Logger())
	t.Cleanup(manager.Close)
	return manager
}

func findAnimeListByStatus(t *testing.T, collection *anilist.AnimeCollection, status anilist.MediaListStatus) *anilist.AnimeCollection_MediaListCollection_Lists {
	t.Helper()

	for _, list := range collection.MediaListCollection.Lists {
		if list.Status != nil && *list.Status == status {
			return list
		}
	}

	t.Fatalf("anime list with status %s not found", status)
	return nil
}

func newBaseAnime(id int, title string, siteURL string) *anilist.BaseAnime {
	return &anilist.BaseAnime{
		ID:      id,
		SiteURL: new(siteURL),
		Title: &anilist.BaseAnime_Title{
			English: new(title),
		},
	}
}

func newBaseManga(id int, title string, siteURL string) *anilist.BaseManga {
	return &anilist.BaseManga{
		ID:      id,
		SiteURL: new(siteURL),
		Title: &anilist.BaseManga_Title{
			English: new(title),
		},
	}
}

type fakeCustomSourceProvider struct {
	extensionIdentifier int
	animeByID           map[int]*anilist.BaseAnime
	mangaByID           map[int]*anilist.BaseManga
	animeErr            error
	mangaErr            error
}

// The fake provider only implements the read paths these tests need; everything else can stay nil.

func (f *fakeCustomSourceProvider) GetExtensionIdentifier() int {
	return f.extensionIdentifier
}

func (f *fakeCustomSourceProvider) GetSettings() hibikecustomsource.Settings {
	return hibikecustomsource.Settings{
		SupportsAnime: true,
		SupportsManga: true,
	}
}

func (f *fakeCustomSourceProvider) GetAnime(_ context.Context, ids []int) ([]*anilist.BaseAnime, error) {
	if f.animeErr != nil {
		return nil, f.animeErr
	}

	ret := make([]*anilist.BaseAnime, 0, len(ids))
	for _, id := range ids {
		if media, ok := f.animeByID[id]; ok {
			ret = append(ret, media)
		}
	}
	return ret, nil
}

func (f *fakeCustomSourceProvider) ListAnime(_ context.Context, _ string, _ int, _ int) (*hibikecustomsource.ListAnimeResponse, error) {
	return nil, nil
}

func (f *fakeCustomSourceProvider) GetAnimeWithRelations(_ context.Context, _ int) (*anilist.CompleteAnime, error) {
	return nil, nil
}

func (f *fakeCustomSourceProvider) GetAnimeMetadata(_ context.Context, _ int) (*metadata.AnimeMetadata, error) {
	return nil, nil
}

func (f *fakeCustomSourceProvider) GetAnimeDetails(_ context.Context, _ int) (*anilist.AnimeDetailsById_Media, error) {
	return nil, nil
}

func (f *fakeCustomSourceProvider) GetManga(_ context.Context, ids []int) ([]*anilist.BaseManga, error) {
	if f.mangaErr != nil {
		return nil, f.mangaErr
	}

	ret := make([]*anilist.BaseManga, 0, len(ids))
	for _, id := range ids {
		if media, ok := f.mangaByID[id]; ok {
			ret = append(ret, media)
		}
	}
	return ret, nil
}

func (f *fakeCustomSourceProvider) ListManga(_ context.Context, _ string, _ int, _ int) (*hibikecustomsource.ListMangaResponse, error) {
	return nil, nil
}

func (f *fakeCustomSourceProvider) GetMangaDetails(_ context.Context, _ int) (*anilist.MangaDetailsById_Media, error) {
	return nil, nil
}
