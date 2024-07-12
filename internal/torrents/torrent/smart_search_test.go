package torrent

import (
	"context"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/api/metadata"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/torrents/animetosho"
	"github.com/seanime-app/seanime/internal/torrents/nyaa"
	"github.com/seanime-app/seanime/internal/util"
	"testing"
)

func TestSmartTest(t *testing.T) {
	test_utils.InitTestProvider(t)

	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()

	metadataProvider := metadata.TestGetMockProvider(t)

	tests := []struct {
		smartSearch    bool
		query          string
		episodeNumber  int
		batch          bool
		mediaId        int
		absoluteOffset int
		resolution     string
		provider       string
	}{
		{
			smartSearch:    true,
			query:          "",
			episodeNumber:  5,
			batch:          false,
			mediaId:        162670, // Dr. Stone S3
			absoluteOffset: 48,
			resolution:     "1080",
			provider:       "animetosho",
		},
		{
			smartSearch:    true,
			query:          "",
			episodeNumber:  1,
			batch:          true,
			mediaId:        77,
			absoluteOffset: 0,
			resolution:     "1080",
			provider:       "animetosho",
		},
		{
			smartSearch:    true,
			query:          "",
			episodeNumber:  1,
			batch:          true,
			mediaId:        109731, // Hibike Season 3
			absoluteOffset: 0,
			resolution:     "1080",
			provider:       "animetosho",
		},
		{
			smartSearch:    true,
			query:          "",
			episodeNumber:  1,
			batch:          true,
			mediaId:        1915, // Magical Girl Lyrical Nanoha StrikerS
			absoluteOffset: 0,
			resolution:     "",
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
					SmartSearch:    lo.ToPtr(tt.smartSearch),
					Query:          lo.ToPtr(tt.query),
					EpisodeNumber:  lo.ToPtr(tt.episodeNumber),
					Batch:          lo.ToPtr(tt.batch),
					Media:          media,
					Best:           lo.ToPtr(false),
					AbsoluteOffset: lo.ToPtr(tt.absoluteOffset),
					Resolution:     lo.ToPtr(tt.resolution),
					Provider:       tt.provider,
				},
				NyaaSearchCache:       nyaa.NewSearchCache(),
				AnimeToshoSearchCache: animetosho.NewSearchCache(),
				AnizipCache:           anizip.NewCache(),
				Logger:                util.NewLogger(),
				MetadataProvider:      metadataProvider,
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
