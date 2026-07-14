package manga

import (
	"seanime/internal/events"
	"seanime/internal/testutil"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMangaPreferencesImportAndPatch(t *testing.T) {
	env := testutil.NewTestEnv(t)
	repository := NewTestRepositoryWithEnv(env, env.NewDatabase("manga_preferences"))

	provider := "provider-a"
	_, err := repository.PatchPreference(1, &MangaPreferencePatch{Provider: &provider}, true)
	require.NoError(t, err)

	preferences, err := repository.ImportPreferences(&MangaPreferences{Entries: map[int]MangaEntryPreference{
		1: {
			Provider: "client-provider",
			Filters: map[string]MangaProviderFilter{
				"provider-a": {Scanlators: []string{"Group A"}, Language: "en"},
			},
		},
		2: {Provider: "provider-b"},
	}})
	require.NoError(t, err)
	require.Equal(t, "provider-a", preferences.Entries[1].Provider)
	require.Equal(t, "provider-b", preferences.Entries[2].Provider)
	require.Equal(t, []string{"Group A"}, preferences.Entries[1].Filters["provider-a"].Scanlators)

	_, err = repository.PatchPreference(1, &MangaPreferencePatch{Filter: &MangaProviderFilterPatch{
		Provider: "provider-a", Scanlators: new([]string{}), Language: new("fr"),
	}}, true)
	require.NoError(t, err)

	stored, err := repository.GetMangaPreferences()
	require.NoError(t, err)
	require.Equal(t, "provider-a", stored.Entries[1].Provider)
	require.Empty(t, stored.Entries[1].Filters["provider-a"].Scanlators)
	require.Equal(t, "fr", stored.Entries[1].Filters["provider-a"].Language)

	ws := repository.wsEventManager.(*events.MockWSEventManager)
	require.NotEmpty(t, ws.Events())
	require.Equal(t, events.MangaPreferencesUpdated, ws.Events()[len(ws.Events())-1].Type)
}

func TestMangaPreferenceValidation(t *testing.T) {
	env := testutil.NewTestEnv(t)
	repository := NewTestRepositoryWithEnv(env, env.NewDatabase("manga_preferences_validation"))

	_, err := repository.PatchPreference(0, &MangaPreferencePatch{}, true)
	require.Error(t, err)
	_, err = repository.PatchPreference(1, &MangaPreferencePatch{}, true)
	require.Error(t, err)

	empty := ""
	_, err = repository.PatchPreference(1, &MangaPreferencePatch{Provider: &empty}, true)
	require.Error(t, err)
	_, err = repository.PatchPreference(1, &MangaPreferencePatch{Filter: &MangaProviderFilterPatch{Provider: "provider-a"}}, true)
	require.Error(t, err)
}

func TestMangaPreferencesPersistAndMergeConcurrentPatches(t *testing.T) {
	env := testutil.NewTestEnv(t)
	database := env.NewDatabase("manga_preferences_persistence")
	repository := NewTestRepositoryWithEnv(env, database)

	provider := "provider-a"
	var waitGroup sync.WaitGroup
	errors := make(chan error, 3)
	waitGroup.Add(3)
	go func() {
		defer waitGroup.Done()
		_, err := repository.PatchPreference(1, &MangaPreferencePatch{Provider: &provider}, false)
		errors <- err
	}()
	go func() {
		defer waitGroup.Done()
		_, err := repository.PatchPreference(1, &MangaPreferencePatch{Filter: &MangaProviderFilterPatch{
			Provider: "provider-b", Scanlators: new([]string{"Group B"}),
		}}, false)
		errors <- err
	}()
	go func() {
		defer waitGroup.Done()
		_, err := repository.PatchPreference(1, &MangaPreferencePatch{Filter: &MangaProviderFilterPatch{
			Provider: "provider-b", Language: new("fr"),
		}}, false)
		errors <- err
	}()
	waitGroup.Wait()
	close(errors)
	for err := range errors {
		require.NoError(t, err)
	}

	restartedRepository := NewTestRepositoryWithEnv(env, database)
	preferences, err := restartedRepository.GetMangaPreferences()
	require.NoError(t, err)
	require.Equal(t, "provider-a", preferences.Entries[1].Provider)
	require.Equal(t, MangaProviderFilter{
		Scanlators: []string{"Group B"}, Language: "fr",
	}, preferences.Entries[1].Filters["provider-b"])
}
