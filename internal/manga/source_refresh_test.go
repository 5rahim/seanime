package manga

import (
	"context"
	"errors"
	"seanime/internal/api/anilist"
	"seanime/internal/database/models"
	"seanime/internal/events"
	"seanime/internal/extension"
	hibikemanga "seanime/internal/extension/hibike/manga"
	"seanime/internal/testmocks"
	"seanime/internal/testutil"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestMangaSourceRefreshChapterScore(t *testing.T) {
	container := &ChapterContainer{Chapters: []*hibikemanga.ChapterDetails{
		{Chapter: "01", Language: "en"},
		{Chapter: "1", Language: "fr"},
		{Chapter: "01.0", Scanlator: "A"},
		{Chapter: "1.0", Scanlator: "B"},
	}}
	require.Equal(t, 2, mangaSourceRefreshChapterScore(container))
}

func TestChooseMangaSourceRefreshCandidate(t *testing.T) {
	chapters := func(count int) *ChapterContainer {
		container := &ChapterContainer{}
		for i := 1; i <= count; i++ {
			container.Chapters = append(container.Chapters, &hibikemanga.ChapterDetails{Chapter: string(rune('0' + i))})
		}
		return container
	}
	candidates := []mangaSourceRefreshTaskResult{
		{providerId: "provider-b", container: chapters(2)},
		{providerId: "provider-a", container: chapters(2)},
	}
	require.Equal(t, "provider-b", chooseMangaSourceRefreshCandidate(candidates, "provider-b", "provider-a").providerId)
	require.Equal(t, "provider-a", chooseMangaSourceRefreshCandidate(candidates, "", "provider-a").providerId)
	require.Equal(t, "provider-a", chooseMangaSourceRefreshCandidate(candidates, "", "").providerId)
}

func TestBuildMangaSourceRefreshPhases(t *testing.T) {
	collection := newMangaSourceRefreshCollection(
		newMangaSourceRefreshEntry(1, anilist.MediaListStatusCurrent),
		newMangaSourceRefreshEntry(2, anilist.MediaListStatusRepeating),
		newMangaSourceRefreshEntry(3, anilist.MediaListStatusCompleted),
	)
	preferences := &MangaPreferences{Entries: map[int]MangaEntryPreference{
		1: {Provider: "provider-a"},
	}}
	providers := []string{"local-manga", "provider-a"}

	selected := buildMangaSourceRefreshPhases(collection, preferences, providers, MangaSourceRefreshSelected)
	require.Len(t, selected, 1)
	require.Len(t, selected[0].plans, 1)
	require.Equal(t, 1, selected[0].plans[0].mediaId)

	missing := buildMangaSourceRefreshPhases(collection, preferences, providers, MangaSourceRefreshMissing)
	require.Len(t, missing, 1)
	require.Len(t, missing[0].plans, 1)
	require.Equal(t, 2, missing[0].plans[0].mediaId)

	phases := buildMangaSourceRefreshPhases(collection, preferences, providers, MangaSourceRefreshSelectedMissing)
	require.Len(t, phases, 2)
	require.Equal(t, []string{"provider-a"}, phases[0].plans[0].providers)
	require.Equal(t, 2, phases[1].plans[0].mediaId)
	require.Equal(t, providers, phases[1].plans[0].providers)

	all := buildMangaSourceRefreshPhases(collection, preferences, providers, MangaSourceRefreshAll)
	require.Len(t, all[0].plans, 2)

	targeted := buildMangaSourceRefreshPhases(collection, preferences, providers, MangaSourceRefreshAll, 2, 3)
	require.Len(t, targeted[0].plans, 1)
	require.Equal(t, 2, targeted[0].plans[0].mediaId)
}

func TestMangaSourceRefreshFindsAndPersistsProvider(t *testing.T) {
	env := testutil.NewTestEnv(t)
	database := env.NewDatabase("manga_source_refresh")
	repository := NewTestRepositoryWithEnv(env, database)
	muteMangaSourceRefreshLogs(repository)
	repository.SetSettings(&models.Settings{Manga: &models.MangaSettings{DefaultProvider: "local-manga"}})

	provider := testmocks.NewFakeMangaProviderBuilder().
		WithChapters("manga-1",
			&hibikemanga.ChapterDetails{ID: "1", Chapter: "1"},
			&hibikemanga.ChapterDetails{ID: "2", Chapter: "2"},
		).
		Build()
	repository.extensionBankRef.Get().Set("local-manga", extension.NewMangaProviderExtension(&extension.Extension{
		ID: "local-manga", Name: "Local manga", Type: extension.TypeMangaProvider,
	}, provider))
	require.NoError(t, database.InsertMangaMapping("local-manga", 1, "manga-1"))
	pageBucket := repository.getFcProviderBucket("local-manga", 1, bucketTypePage)
	dimensionBucket := repository.getFcProviderBucket("local-manga", 1, bucketTypePageDimensions)
	otherPageBucket := repository.getFcProviderBucket("local-manga", 2, bucketTypePage)
	require.NoError(t, repository.fileCacher.Set(pageBucket, "page", "old"))
	require.NoError(t, repository.fileCacher.Set(dimensionBucket, "dimensions", "old"))
	require.NoError(t, repository.fileCacher.Set(otherPageBucket, "page", "keep"))

	job, err := repository.StartMangaSourceRefresh("client-1", MangaSourceRefreshMissing,
		newMangaSourceRefreshCollection(newMangaSourceRefreshEntry(1, anilist.MediaListStatusCurrent)))
	require.NoError(t, err)
	require.Equal(t, 1, job.Total)

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		job = repository.GetMangaSourceRefresh("client-1")
		if job != nil && job.Status == MangaSourceRefreshCompleted {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.NotNil(t, job)
	require.Equal(t, MangaSourceRefreshCompleted, job.Status)
	require.Equal(t, 1, job.Result.Found)

	preferences, err := repository.GetMangaPreferences()
	require.NoError(t, err)
	require.Equal(t, "local-manga", preferences.Entries[1].Provider)
	require.Equal(t, 1, provider.ChapterCalls())
	var value string
	found, err := repository.fileCacher.Get(pageBucket, "page", &value)
	require.NoError(t, err)
	require.False(t, found)
	found, err = repository.fileCacher.Get(dimensionBucket, "dimensions", &value)
	require.NoError(t, err)
	require.False(t, found)
	found, err = repository.fileCacher.Get(otherPageBucket, "page", &value)
	require.NoError(t, err)
	require.True(t, found)

	preferenceEvents := 0
	for _, event := range repository.wsEventManager.(*events.MockWSEventManager).Events() {
		if event.Type == events.MangaPreferencesUpdated {
			preferenceEvents++
		}
	}
	require.Equal(t, 1, preferenceEvents)

	_, err = repository.StopMangaSourceRefresh("client-1")
	require.NoError(t, err)
	require.Nil(t, repository.GetMangaSourceRefresh("client-1"))
}

func TestMangaSourceRefreshOnlyReplacesDuringReevaluation(t *testing.T) {
	env := testutil.NewTestEnv(t)
	database := env.NewDatabase("manga_source_refresh_replacement")
	repository := NewTestRepositoryWithEnv(env, database)
	muteMangaSourceRefreshLogs(repository)
	repository.SetSettings(&models.Settings{Manga: &models.MangaSettings{DefaultProvider: "refresh-provider-b"}})

	providerA := testmocks.NewFakeMangaProviderBuilder().
		WithChapters("manga-a", &hibikemanga.ChapterDetails{ID: "1", Chapter: "1"}).
		Build()
	providerB := testmocks.NewFakeMangaProviderBuilder().
		WithChapters("manga-b",
			&hibikemanga.ChapterDetails{ID: "1", Chapter: "1"},
			&hibikemanga.ChapterDetails{ID: "2", Chapter: "2"},
		).
		Build()
	for id, provider := range map[string]*testmocks.FakeMangaProvider{
		"refresh-provider-a": providerA,
		"refresh-provider-b": providerB,
	} {
		repository.extensionBankRef.Get().Set(id, extension.NewMangaProviderExtension(&extension.Extension{
			ID: id, Name: id, Type: extension.TypeMangaProvider,
		}, provider))
	}
	require.NoError(t, database.InsertMangaMapping("refresh-provider-a", 77, "manga-a"))
	require.NoError(t, database.InsertMangaMapping("refresh-provider-b", 77, "manga-b"))

	_, err := repository.PatchPreference(77, &MangaPreferencePatch{
		Provider: new("refresh-provider-a"),
		Filter: &MangaProviderFilterPatch{
			Provider: "refresh-provider-b", Scanlators: new([]string{"Group B"}), Language: new("fr"),
		},
	}, false)
	require.NoError(t, err)
	collection := newMangaSourceRefreshCollection(newMangaSourceRefreshEntry(77, anilist.MediaListStatusCurrent))

	_, err = repository.StartMangaSourceRefresh("client-1", MangaSourceRefreshSelected, collection)
	require.NoError(t, err)
	job := waitForMangaSourceRefresh(t, repository, "client-1")
	require.Equal(t, 1, job.Result.Refreshed)
	require.Equal(t, 0, job.Result.Replaced)
	preferences, err := repository.GetMangaPreferences()
	require.NoError(t, err)
	require.Equal(t, "refresh-provider-a", preferences.Entries[77].Provider)
	_, err = repository.StopMangaSourceRefresh("client-1")
	require.NoError(t, err)

	_, err = repository.StartMangaSourceRefresh("client-1", MangaSourceRefreshAll, collection)
	require.NoError(t, err)
	job = waitForMangaSourceRefresh(t, repository, "client-1")
	require.Equal(t, 1, job.Result.Replaced)
	preferences, err = repository.GetMangaPreferences()
	require.NoError(t, err)
	require.Equal(t, "refresh-provider-b", preferences.Entries[77].Provider)
	require.Equal(t, MangaProviderFilter{
		Scanlators: []string{"Group B"}, Language: "fr",
	}, preferences.Entries[77].Filters["refresh-provider-b"])
}

func TestMangaSourceRefreshOwnershipAndDismissal(t *testing.T) {
	env := testutil.NewTestEnv(t)
	repository := NewTestRepositoryWithEnv(env, env.NewDatabase("manga_source_refresh_ownership"))
	ctx, cancel := context.WithCancel(context.Background())
	repository.sourceRefresh = &mangaSourceRefreshState{
		owner:  "owner",
		cancel: cancel,
		job: MangaSourceRefreshJob{
			Id: "active", Mode: MangaSourceRefreshSelected, Status: MangaSourceRefreshRunning,
		},
	}
	repository.sourceRefreshLog["viewer"] = mangaSourceRefreshCompleted{
		job:       MangaSourceRefreshJob{Id: "completed", Status: MangaSourceRefreshCompleted},
		expiresAt: time.Now().Add(time.Hour),
	}

	job, err := repository.StartMangaSourceRefresh("owner", MangaSourceRefreshSelected, nil)
	require.NoError(t, err)
	require.Equal(t, "active", job.Id)

	_, err = repository.GetActiveMangaSourceRefresh("other")
	require.ErrorIs(t, err, ErrMangaSourceRefreshConflict)

	job, err = repository.StopMangaSourceRefresh("viewer")
	require.NoError(t, err)
	require.Nil(t, job)
	require.Nil(t, repository.GetMangaSourceRefresh("viewer"))

	job, err = repository.StopMangaSourceRefresh("owner")
	require.NoError(t, err)
	require.Equal(t, MangaSourceRefreshStopping, job.Status)
	require.ErrorIs(t, ctx.Err(), context.Canceled)
}

func TestGetMangaChapterContainerRefreshOptions(t *testing.T) {
	env := testutil.NewTestEnv(t)
	database := env.NewDatabase("manga_chapter_refresh_options")
	repository := NewTestRepositoryWithEnv(env, database)
	repository.SetSettings(&models.Settings{Manga: &models.MangaSettings{}})

	provider := testmocks.NewFakeMangaProviderBuilder().
		WithChapters("manga-1", &hibikemanga.ChapterDetails{ID: "1", Chapter: "1"}).
		Build()
	repository.extensionBankRef.Get().Set("provider-a", extension.NewMangaProviderExtension(&extension.Extension{
		ID: "provider-a", Name: "Provider A", Type: extension.TypeMangaProvider,
	}, provider))
	require.NoError(t, database.InsertMangaMapping("provider-a", 1, "manga-1"))

	opts := &GetMangaChapterContainerOptions{Provider: "provider-a", MediaId: 1}
	container, err := repository.GetMangaChapterContainer(opts)
	require.NoError(t, err)
	require.Equal(t, "1", container.Chapters[0].Chapter)

	provider.SetChapterError(errors.New("provider failed"))
	container, err = repository.GetMangaChapterContainer(opts)
	require.NoError(t, err)
	require.Equal(t, "1", container.Chapters[0].Chapter)
	require.Equal(t, 1, provider.ChapterCalls())

	waits := 0
	_, err = repository.GetMangaChapterContainer(&GetMangaChapterContainerOptions{
		Provider: "provider-a", MediaId: 1, skipCache: true,
		beforeProviderCall: func() error { waits++; return nil },
	})
	require.Error(t, err)
	require.Equal(t, 1, waits)

	provider.SetChapterError(nil)
	provider.SetChapters("manga-1", &hibikemanga.ChapterDetails{ID: "2", Chapter: "2"})
	container, err = repository.GetMangaChapterContainer(&GetMangaChapterContainerOptions{
		Provider: "provider-a", MediaId: 1, skipCache: true,
	})
	require.NoError(t, err)
	require.Equal(t, "2", container.Chapters[0].Chapter)

	container, err = repository.GetMangaChapterContainer(opts)
	require.NoError(t, err)
	require.Equal(t, "2", container.Chapters[0].Chapter)
}

func TestGetMangaChapterContainerCallsLimiterBeforeSearchAndChapters(t *testing.T) {
	env := testutil.NewTestEnv(t)
	repository := NewTestRepositoryWithEnv(env, env.NewDatabase("manga_chapter_refresh_limiter"))
	repository.SetSettings(&models.Settings{Manga: &models.MangaSettings{}})

	provider := testmocks.NewFakeMangaProviderBuilder().
		WithSearchResults(&hibikemanga.SearchResult{ID: "manga-1", Title: "Manga"}).
		WithChapters("manga-1", &hibikemanga.ChapterDetails{ID: "1", Chapter: "1"}).
		Build()
	repository.extensionBankRef.Get().Set("provider-a", extension.NewMangaProviderExtension(&extension.Extension{
		ID: "provider-a", Name: "Provider A", Type: extension.TypeMangaProvider,
	}, provider))

	title := "Manga"
	waits := 0
	container, err := repository.GetMangaChapterContainer(&GetMangaChapterContainerOptions{
		Provider: "provider-a", MediaId: 999, Titles: []*string{&title}, skipCache: true,
		beforeProviderCall: func() error { waits++; return nil },
	})
	require.NoError(t, err)
	require.Len(t, container.Chapters, 1)
	require.Equal(t, 2, waits)
	require.Equal(t, 1, provider.SearchCalls())
	require.Equal(t, 1, provider.ChapterCalls())
}

func TestMangaSourceRefreshErrorCategories(t *testing.T) {
	require.False(t, isSourceRefreshProviderError(ErrNoResults))
	require.False(t, isSourceRefreshProviderError(ErrNoChapters))
	require.False(t, isSourceRefreshProviderError(context.Canceled))
	require.True(t, isSourceRefreshProviderError(errors.New("provider failed")))
	require.True(t, isSourceRefreshProviderError(errors.Join(ErrNoChapters, errors.New("provider failed"))))
}

func newMangaSourceRefreshEntry(mediaId int, status anilist.MediaListStatus) *anilist.MangaListEntry {
	return &anilist.MangaListEntry{
		Media:  testmocks.NewBaseManga(mediaId, "Manga"),
		Status: new(status),
	}
}

func newMangaSourceRefreshCollection(entries ...*anilist.MangaListEntry) *anilist.MangaCollection {
	return &anilist.MangaCollection{MediaListCollection: &anilist.MangaCollection_MediaListCollection{
		Lists: []*anilist.MangaList{{Entries: entries}},
	}}
}

func waitForMangaSourceRefresh(t *testing.T, repository *Repository, clientId string) *MangaSourceRefreshJob {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		job := repository.GetMangaSourceRefresh(clientId)
		if job != nil && job.Status != MangaSourceRefreshRunning && job.Status != MangaSourceRefreshStopping {
			return job
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("manga source refresh did not finish")
	return nil
}

func muteMangaSourceRefreshLogs(repository *Repository) {
	logger := zerolog.Nop()
	repository.logger = &logger
	repository.wsEventManager = events.NewMockWSEventManager(&logger)
}
