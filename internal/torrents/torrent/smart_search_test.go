package torrent

import (
	"context"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/torrents/animetosho"
	"github.com/seanime-app/seanime/internal/torrents/nyaa"
	"github.com/seanime-app/seanime/internal/util"
	"testing"
)

func TestSmartTest(t *testing.T) {
	test_utils.InitTestProvider(t)

	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()

	tests := []struct {
		quickSearch    bool
		query          string
		episodeNumber  int
		batch          bool
		mediaId        int
		absoluteOffset int
		resolution     string
		provider       string
	}{
		{
			quickSearch:    true,
			query:          "",
			episodeNumber:  1,
			batch:          false,
			mediaId:        162670,
			absoluteOffset: 48,
			resolution:     "1080",
			provider:       "animetosho",
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {

			mediaF, err := anilistClientWrapper.BaseMediaByID(context.Background(), &tt.mediaId)
			if err != nil {
				t.Fatalf("could not fetch media id %d", tt.mediaId)
			}

			media := mediaF.GetMedia()

			data, err := NewSmartSearch(&SmartSearchOptions{
				SmartSearchQueryOptions: SmartSearchQueryOptions{
					QuickSearch:    lo.ToPtr(tt.quickSearch),
					Query:          lo.ToPtr(tt.query),
					EpisodeNumber:  lo.ToPtr(tt.episodeNumber),
					Batch:          lo.ToPtr(tt.batch),
					Media:          media,
					AbsoluteOffset: lo.ToPtr(tt.absoluteOffset),
					Resolution:     lo.ToPtr(tt.resolution),
					Provider:       tt.provider,
				},
				NyaaSearchCache:       nyaa.NewSearchCache(),
				AnimeToshoSearchCache: animetosho.NewSearchCache(),
				AnizipCache:           anizip.NewCache(),
				Logger:                util.NewLogger(),
			})
			if err != nil {
				t.Errorf("NewSmartSearch() failed: %v", err)
			}

			t.Log("----------------------- Previews --------------------------")
			for _, preview := range data.Previews {
				t.Logf("> %s", preview.Torrent.Name)
				if preview.Episode != nil {
					t.Logf("\t\t %s", preview.Episode.DisplayTitle)
				} else {
					t.Logf("\t\t Batch")
				}
			}
			t.Log("----------------------- Torrents --------------------------")
			for _, torrent := range data.Torrents {
				t.Logf("> %s", torrent.Name)
			}

		})
	}
}
