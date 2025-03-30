package torrent_analyzer

import (
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSelectFilesFromSeason tests the selection of the accurate season files from a list of files from all seasons.
func TestSelectFilesFromSeason(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	logger := util.NewLogger()
	anilistClient := anilist.TestGetMockAnilistClient()
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilistClient, logger)
	metadataProvider := metadata.GetMockProvider(t)

	tests := []struct {
		name            string
		mediaId         int      // The media ID of the season
		filepaths       []string // All filepaths from all seasons
		expectedIndices []int    // The indices of the selected files
	}{
		{
			name: "Kakegurui xx",
			filepaths: []string{
				"Kakegurui [BD][1080p][HEVC 10bit x265][Dual Audio][Tenrai-Sensei]/Season 1/Kakegurui - S01E01 - The Woman Called Yumeko Jabami.mkv", // should be selected
				"Kakegurui [BD][1080p][HEVC 10bit x265][Dual Audio][Tenrai-Sensei]/Season 2/Kakegurui xx - S02E01 - The Woman Called Yumeko Jabami.mkv",
			},
			mediaId:         98314,
			expectedIndices: []int{0},
		},
		{
			name: "Kimi ni Todoke Season 2",
			filepaths: []string{
				"[Judas] Kimi ni Todoke (Seasons 1-2) [BD 1080p][HEVC x265 10bit][Eng-Subs]/[Judas] Kimi ni Todoke S1/[Judas] Kimi ni Todoke - S01E01.mkv",
				"[Judas] Kimi ni Todoke (Seasons 1-2) [BD 1080p][HEVC x265 10bit][Eng-Subs]/[Judas] Kimi ni Todoke S1/[Judas] Kimi ni Todoke - S01E02.mkv",
				"[Judas] Kimi ni Todoke (Seasons 1-2) [BD 1080p][HEVC x265 10bit][Eng-Subs]/[Judas] Kimi ni Todoke S2/[Judas] Kimi ni Todoke - S02E01.mkv", // should be selected
				"[Judas] Kimi ni Todoke (Seasons 1-2) [BD 1080p][HEVC x265 10bit][Eng-Subs]/[Judas] Kimi ni Todoke S2/[Judas] Kimi ni Todoke - S02E02.mkv", // should be selected
			},
			mediaId:         9656,
			expectedIndices: []int{2, 3},
		},
		{
			name: "Spy x Family Part 2",
			filepaths: []string{
				"[SubsPlease] Spy x Family (01-25) (1080p) [Batch]/[SubsPlease] Spy x Family - 10v2 (1080p) [F9F5C62B].mkv",
				"[SubsPlease] Spy x Family (01-25) (1080p) [Batch]/[SubsPlease] Spy x Family - 11v2 (1080p) [F9F5C62B].mkv",
				"[SubsPlease] Spy x Family (01-25) (1080p) [Batch]/[SubsPlease] Spy x Family - 12v2 (1080p) [F9F5C62B].mkv",
				"[SubsPlease] Spy x Family (01-25) (1080p) [Batch]/[SubsPlease] Spy x Family - 13v2 (1080p) [F9F5C62B].mkv", // should be selected
				"[SubsPlease] Spy x Family (01-25) (1080p) [Batch]/[SubsPlease] Spy x Family - 14v2 (1080p) [F9F5C62B].mkv", // should be selected
				"[SubsPlease] Spy x Family (01-25) (1080p) [Batch]/[SubsPlease] Spy x Family - 15v2 (1080p) [F9F5C62B].mkv", // should be selected
			},
			mediaId:         142838,
			expectedIndices: []int{3, 4, 5},
		},
		{
			name: "Mushoku Tensei: Jobless Reincarnation Season 2 Part 2",
			filepaths: []string{
				"[EMBER] Mushoku Tensei S2 - 13.mkv", // should be selected
				"[EMBER] Mushoku Tensei S2 - 14.mkv", // should be selected
				"[EMBER] Mushoku Tensei S2 - 15.mkv", // should be selected
				"[EMBER] Mushoku Tensei S2 - 16.mkv", // should be selected
			},
			mediaId:         166873,
			expectedIndices: []int{0, 1, 2, 3},
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			// Get media
			media, err := anilistPlatform.GetAnimeWithRelations(tt.mediaId)
			if err != nil {
				t.Fatal("expected result, got error:", err.Error())
			}

			analyzer := NewAnalyzer(&NewAnalyzerOptions{
				Logger:           logger,
				Filepaths:        tt.filepaths,
				Media:            media,
				Platform:         anilistPlatform,
				MetadataProvider: metadataProvider,
				ForceMatch:       false,
			})

			// AnalyzeTorrentFiles
			analysis, err := analyzer.AnalyzeTorrentFiles()
			if assert.NoError(t, err) {

				selectedFilesMap := analysis.GetCorrespondingMainFiles()
				selectedIndices := analysis.GetIndices(selectedFilesMap)

				// Check selected files
				assert.ElementsMatch(t, tt.expectedIndices, selectedIndices)

			}

		})

	}

}
