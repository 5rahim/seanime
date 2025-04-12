package extension_repo_test

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
	hibikeonlinestream "seanime/internal/extension/hibike/onlinestream"
	"testing"
)

func TestExternalGoOnlinestreamProviderExtension(t *testing.T) {

	repo := getRepo(t)

	// Load all extensions
	// This should load all the extensions in the directory
	repo.ReloadExternalExtensions()

	ext, found := repo.GetOnlinestreamProviderExtensionByID("gogoanime-external")
	require.True(t, found)

	t.Logf("\nExtension:\n\tID: %s \n\tName: %s", ext.GetID(), ext.GetName())

	searchResults, err := ext.GetProvider().Search(hibikeonlinestream.SearchOptions{
		Query: "Blue Lock",
		Dub:   false,
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(searchResults), 1)

	episodes, err := ext.GetProvider().FindEpisodes(searchResults[0].ID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(episodes), 1)

	server, err := ext.GetProvider().FindEpisodeServer(episodes[0], ext.GetProvider().GetSettings().EpisodeServers[0])
	require.NoError(t, err)

	spew.Dump(server)

}
