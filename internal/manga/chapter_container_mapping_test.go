package manga

import (
	"seanime/internal/extension"
	hibikemanga "seanime/internal/extension/hibike/manga"
	"seanime/internal/testmocks"
	"seanime/internal/testutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPreviewMapping(t *testing.T) {
	env := testutil.NewTestEnv(t)
	repository := NewTestRepositoryWithEnv(env, env.NewDatabase("manga_mapping_preview"))
	provider := testmocks.NewFakeMangaProviderBuilder().
		WithChapters("manga-1",
			&hibikemanga.ChapterDetails{Chapter: "01", Language: "en", Scanlator: "B"},
			&hibikemanga.ChapterDetails{Chapter: "1", Language: "fr", Scanlator: "A"},
			&hibikemanga.ChapterDetails{Chapter: "3.5", Language: "en", Scanlator: "A"},
		).
		Build()
	repository.extensionBankRef.Get().Set("provider-a", extension.NewMangaProviderExtension(&extension.Extension{
		ID: "provider-a", Name: "Provider A", Type: extension.TypeMangaProvider,
	}, provider))

	preview, err := repository.PreviewMapping("provider-a", "manga-1")
	require.NoError(t, err)
	require.Equal(t, 2, preview.ChapterCount)
	require.Equal(t, "3.5", preview.Latest)
	require.Equal(t, []string{"en", "fr"}, preview.Languages)
	require.Equal(t, []string{"A", "B"}, preview.Scanlators)
}
