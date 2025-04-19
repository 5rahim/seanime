package extension_repo_test

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
	"seanime/internal/extension"
	hibikemanga "seanime/internal/extension/hibike/manga"
	"seanime/internal/manga/providers"
	"seanime/internal/util"
	"testing"
)

// Tests the external manga provider extension loaded from the extension directory.
// This will load the extensions from ./testdir
func TestExternalGoMangaExtension(t *testing.T) {

	repo := getRepo(t)

	// Load all extensions
	// This should load all the extensions in the directory
	repo.ReloadExternalExtensions()

	ext, found := repo.GetMangaProviderExtensionByID("mangapill-external")
	require.True(t, found)

	t.Logf("\nExtension:\n\tID: %s \n\tName: %s", ext.GetID(), ext.GetName())

	// Test the extension
	so := hibikemanga.SearchOptions{
		Query: "Dandadan",
	}

	searchResults, err := ext.GetProvider().Search(so)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(searchResults), 1)

	chapters, err := ext.GetProvider().FindChapters(searchResults[0].ID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(chapters), 1)

	spew.Dump(chapters[0])

}

// Tests the built-in manga provider extension
func TestBuiltinMangaExtension(t *testing.T) {

	logger := util.NewLogger()
	repo := getRepo(t)

	// Load all extensions
	// This should load all the extensions in the directory
	repo.ReloadBuiltInExtension(extension.Extension{
		ID:          "seanime-builtin-mangapill",
		Type:        "manga-provider",
		Name:        "Mangapill",
		Version:     "0.0.0",
		Language:    "go",
		ManifestURI: "",
		Description: "",
		Author:      "",
		Payload:     "",
	}, manga_providers.NewMangapill(logger))

	ext, found := repo.GetMangaProviderExtensionByID("seanime-builtin-mangapill")
	require.True(t, found)

	t.Logf("\nExtension:\n\tID: %s \n\tName: %s", ext.GetID(), ext.GetName())

	// Test the extension
	so := hibikemanga.SearchOptions{
		Query: "Dandadan",
	}

	searchResults, err := ext.GetProvider().Search(so)
	require.NoError(t, err)

	spew.Dump(searchResults)

}
