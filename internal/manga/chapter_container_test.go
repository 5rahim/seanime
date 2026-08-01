package manga

import (
	"testing"

	hibikemanga "seanime/internal/extension/hibike/manga"

	"github.com/stretchr/testify/require"
)

func TestGetLatestMangaChapterNumberUsesChapterValue(t *testing.T) {
	chapters := []*hibikemanga.ChapterDetails{
		{Chapter: "205", Index: 0},
		{Chapter: "204.5", Index: 1},
		{Chapter: "1", Index: 205},
	}

	require.Equal(t, 205, getLatestMangaChapterNumber(chapters))
}
