package scanner

import (
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/database/db"
	"seanime/internal/extension"
	"seanime/internal/library/anime"
	"seanime/internal/library/summary"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"seanime/internal/util/limiter"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileHydrator_HydrateMetadata(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	completeAnimeCache := anilist.NewCompleteAnimeCache()
	anilistRateLimiter := limiter.NewAnilistLimiter()
	logger := util.NewLogger()
	database, err := db.NewDatabase(test_utils.ConfigData.Path.DataDir, test_utils.ConfigData.Database.Name, logger)
	require.NoError(t, err)
	metadataProvider := metadata_provider.GetFakeProvider(t, database)
	anilistClient := anilist.NewAnilistClient(test_utils.ConfigData.Provider.AnilistJwt, "")
	anilistClientRef := util.NewRef[anilist.AnilistClient](anilistClient)
	extensionBankRef := util.NewRef(extension.NewUnifiedBank())
	//wsEventManager := events.NewMockWSEventManager(logger)
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilistClientRef, extensionBankRef, logger, database)
	anilistPlatform.SetUsername(test_utils.ConfigData.Provider.AnilistUsername)
	animeCollection, err := anilistPlatform.GetAnimeCollectionWithRelations(t.Context())
	require.NoError(t, err)
	require.NotNil(t, animeCollection)

	allMedia := animeCollection.GetAllAnime()

	tests := []struct {
		name            string
		paths           []string
		expectedMediaId int
		expectedType    anime.LocalFileType
		config          string
	}{
		{
			name: "should be hydrated with id 131586",
			paths: []string{
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 20v2 (1080p) [30072859].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 21v2 (1080p) [4B1616A5].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 22v2 (1080p) [58BF43B4].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 23v2 - Never-Ending (1080p) [D94B4894].mkv",
			},
			expectedMediaId: 131586, // 86 - Eighty Six Part 2
			expectedType:    anime.LocalFileTypeMain,
		},
		{
			name: "Danmachi S04P02 E12 - part-relative absolute, first ep of Part 2",
			paths: []string{
				"E:/Anime/Danmachi S04P02 1080p Dual Audio BDRip 10 bits DD x265-EMBER/S04E12-Amphisbaena A Song of Despair [080E734C].mkv",
			},
			expectedMediaId: 155211, // Danmachi IV Part 2
			expectedType:    anime.LocalFileTypeMain,
		},
		{
			name: "Danmachi S04P02 E22 - part-relative absolute, last ep of Part 2",
			paths: []string{
				"E:/Anime/Danmachi S04P02 1080p Dual Audio BDRip 10 bits DD x265-EMBER/S04E22-Amphisbaena A Song of Despair [080E734C].mkv",
			},
			expectedMediaId: 155211, // Danmachi IV Part 2
			expectedType:    anime.LocalFileTypeMain,
		},
		{
			// anidb lists 1 main episode but anilist lists 5
			// the media tree analysis should use the anilist episode count so normalize the episode numbers
			name: "Muri ja Nakatta!",
			paths: []string{
				"E:/Anime/Watashi ga Koibito ni Nareru Wake Naijan, Murimuri! (※Muri ja Nakatta! ) Next Shine!/Theres.No.Freaking.Way.Ill.Be.Your.Lover.Unless.S01E13.1080p.AMZN.WEB-DL.JPN.DDP2.0.H.264.MSubs-ToonsHub.mkv",
				"E:/Anime/Watashi ga Koibito ni Nareru Wake Naijan, Murimuri! (※Muri ja Nakatta! ) Next Shine!/Theres.No.Freaking.Way.Ill.Be.Your.Lover.Unless.S01E14.1080p.AMZN.WEB-DL.JPN.DDP2.0.H.264.MSubs-ToonsHub.mkv",
			},
			expectedMediaId: 199112,
			expectedType:    anime.LocalFileTypeMain,
		},
		{
			name: "Revue Starlight Movie",
			paths: []string{
				"E:/Anime/Shoujo Kageki Revue Starlight/[neoHEVC] Revue Starlight [Season 1 + Specials + Movie] [BD 1080p x265 HEVC AAC] [Dual Audio]/Specials/Revue Starlight - S00E05 - (S1M1 - Movie).mkv",
			},
			expectedMediaId: 113024,
			expectedType:    anime.LocalFileTypeMain,
		},
		{
			name: "Kimetsu no Yaiba Hashira Geiko-hen",
			paths: []string{
				"E:/Anime/Kimetsu no Yaiba Hashira Geiko-hen/[Judas] Kimetsu no Yaiba (Demon Slayer) (Season 05) [1080p][HEVC x265 10bit][Dual-Audio][Multi-Subs]/[Judas] Kimetsu no Yaiba - S05E01v2.mkv",
			},
			expectedMediaId: 166240,
			expectedType:    anime.LocalFileTypeMain,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			scanLogger, err := NewConsoleScanLogger()
			if err != nil {
				t.Fatal("expected result, got error:", err.Error())
			}

			// +---------------------+
			// |   Local Files       |
			// +---------------------+

			var lfs []*anime.LocalFile
			for _, path := range tt.paths {
				lf := anime.NewLocalFile(path, "E:/Anime")
				lfs = append(lfs, lf)
			}

			// +---------------------+
			// |   MediaContainer    |
			// +---------------------+

			mc := NewMediaContainer(&MediaContainerOptions{
				AllMedia:   NormalizedMediaFromAnilistComplete(allMedia),
				ScanLogger: scanLogger,
			})

			for _, nm := range mc.NormalizedMedia {
				t.Logf("media id: %d, title: %s", nm.ID, nm.GetTitleSafe())
			}

			// +---------------------+
			// |      Matcher        |
			// +---------------------+

			matcher := &Matcher{
				LocalFiles:     lfs,
				MediaContainer: mc,
				Logger:         util.NewLogger(),
				ScanLogger:     scanLogger,
			}

			err = matcher.MatchLocalFilesWithMedia()
			if err != nil {
				t.Fatal("expected result, got error:", err.Error())
			}

			// +---------------------+
			// |    FileHydrator     |
			// +---------------------+

			config, _ := ToConfig(tt.config)

			fh := &FileHydrator{
				LocalFiles:          lfs,
				AllMedia:            mc.NormalizedMedia,
				CompleteAnimeCache:  completeAnimeCache,
				PlatformRef:         util.NewRef(anilistPlatform),
				AnilistRateLimiter:  anilistRateLimiter,
				MetadataProviderRef: util.NewRef(metadataProvider),
				Logger:              logger,
				ScanLogger:          scanLogger,
				Config:              config,
			}

			fh.HydrateMetadata()

			for _, lf := range fh.LocalFiles {
				t.Logf("local file: %s, media id: %d, type: %s, episode: %d, aniDbEpisode: %s\n", lf.Name, lf.MediaId, lf.GetType(), lf.Metadata.Episode, lf.Metadata.AniDBEpisode)
				assert.NotNil(t, lf.MediaId, "expected media id to be set")
				assert.Equal(t, tt.expectedMediaId, lf.MediaId, "expected media id %d, got %d", tt.expectedMediaId, lf.MediaId)
				assert.Equal(t, tt.expectedType, lf.GetType(), "expected file type %s, got %s", tt.expectedType, lf.GetType())
			}

		})
	}

}

func TestFileHydrator_applyHydrationRule(t *testing.T) {
	logger := util.NewLogger()
	scanSummaryLogger := summary.NewScanSummaryLogger()

	tests := []struct {
		name           string
		rules          []*HydrationRule
		localFile      *anime.LocalFile
		expectedResult bool
		expectedMeta   *anime.LocalFileMetadata
	}{
		{
			name: "should apply rule with exact filename match",
			rules: []*HydrationRule{
				{
					Pattern: "",
					MediaID: 123,
					Files: []*HydrationFileRule{
						{
							Filename:     "episode_01.mkv",
							IsRegex:      false,
							Episode:      "1",
							AniDbEpisode: "S1",
							Type:         anime.LocalFileTypeMain,
						},
					},
				},
			},
			localFile: &anime.LocalFile{
				Name:    "episode_01.mkv",
				MediaId: 123,
				Metadata: &anime.LocalFileMetadata{
					Episode:      0,
					AniDBEpisode: "",
					Type:         "",
				},
			},
			expectedResult: true,
			expectedMeta: &anime.LocalFileMetadata{
				Episode:      1,
				AniDBEpisode: "S1",
				Type:         anime.LocalFileTypeMain,
			},
		},
		{
			name: "should apply rule with regex pattern match",
			rules: []*HydrationRule{
				{
					Pattern: ".*Episode.*",
					MediaID: 0,
					Files: []*HydrationFileRule{
						{
							Filename:     "Episode_(\\d+)\\.mkv",
							IsRegex:      true,
							Episode:      "5",
							AniDbEpisode: "",
							Type:         anime.LocalFileTypeSpecial,
						},
					},
				},
			},
			localFile: &anime.LocalFile{
				Name:    "Episode_03.mkv",
				MediaId: 456,
				Metadata: &anime.LocalFileMetadata{
					Episode:      0,
					AniDBEpisode: "",
					Type:         "",
				},
			},
			expectedResult: true,
			expectedMeta: &anime.LocalFileMetadata{
				Episode:      5,
				AniDBEpisode: "",
				Type:         anime.LocalFileTypeSpecial,
			},
		},
		{
			name: "should apply rule with regex substitution, single capture group",
			rules: []*HydrationRule{
				{
					Pattern: "",
					MediaID: 789,
					Files: []*HydrationFileRule{
						{
							Filename:     "EP_(\\d+)\\.mkv",
							IsRegex:      true,
							Episode:      "$1",
							AniDbEpisode: "$1",
							Type:         anime.LocalFileTypeMain,
						},
					},
				},
			},
			localFile: &anime.LocalFile{
				Name:    "EP_12.mkv",
				MediaId: 789,
				Metadata: &anime.LocalFileMetadata{
					Episode:      0,
					AniDBEpisode: "",
					Type:         "",
				},
			},
			expectedResult: true,
			expectedMeta: &anime.LocalFileMetadata{
				Episode:      12,
				AniDBEpisode: "12",
				Type:         anime.LocalFileTypeMain,
			},
		},
		{
			name: "should apply rule with regex substitution, single capture group, calc",
			rules: []*HydrationRule{
				{
					Pattern: "",
					MediaID: 789,
					Files: []*HydrationFileRule{
						{
							Filename:     "EP_(\\d+)\\.mkv",
							IsRegex:      true,
							Episode:      "calc($1-11)",
							AniDbEpisode: "S{calc($1-11)}",
							Type:         anime.LocalFileTypeMain,
						},
					},
				},
			},
			localFile: &anime.LocalFile{
				Name:    "EP_12.mkv",
				MediaId: 789,
				Metadata: &anime.LocalFileMetadata{
					Episode:      0,
					AniDBEpisode: "",
					Type:         "",
				},
			},
			expectedResult: true,
			expectedMeta: &anime.LocalFileMetadata{
				Episode:      1,
				AniDBEpisode: "S1",
				Type:         anime.LocalFileTypeMain,
			},
		},
		{
			name: "should apply rule with regex substitution, multiple capture groups",
			rules: []*HydrationRule{
				{
					Pattern: "",
					MediaID: 100,
					Files: []*HydrationFileRule{
						{
							Filename:     "Season (\\d+) Episode (\\d+)\\.mkv",
							IsRegex:      true,
							Episode:      "$2",
							AniDbEpisode: "$2",
							Type:         anime.LocalFileTypeMain,
						},
					},
				},
			},
			localFile: &anime.LocalFile{
				Name:    "Season 2 Episode 5.mkv",
				MediaId: 100,
				Metadata: &anime.LocalFileMetadata{
					Episode:      0,
					AniDBEpisode: "",
					Type:         "",
				},
			},
			expectedResult: true,
			expectedMeta: &anime.LocalFileMetadata{
				Episode:      5,
				AniDBEpisode: "5",
				Type:         anime.LocalFileTypeMain,
			},
		},
		{
			name: "should not apply rule when media ID doesn't match",
			rules: []*HydrationRule{
				{
					Pattern: "",
					MediaID: 999,
					Files: []*HydrationFileRule{
						{
							Filename:     "episode_01.mkv",
							IsRegex:      false,
							Episode:      "1",
							AniDbEpisode: "",
							Type:         anime.LocalFileTypeMain,
						},
					},
				},
			},
			localFile: &anime.LocalFile{
				Name:    "episode_01.mkv",
				MediaId: 123,
				Metadata: &anime.LocalFileMetadata{
					Episode:      0,
					AniDBEpisode: "",
					Type:         "",
				},
			},
			expectedResult: false,
			expectedMeta: &anime.LocalFileMetadata{
				Episode:      0,
				AniDBEpisode: "",
				Type:         "",
			},
		},
		{
			name: "should not apply rule when pattern doesn't match",
			rules: []*HydrationRule{
				{
					Pattern: ".*Special.*",
					MediaID: 0,
					Files: []*HydrationFileRule{
						{
							Filename:     "episode_01.mkv",
							IsRegex:      false,
							Episode:      "1",
							AniDbEpisode: "",
							Type:         anime.LocalFileTypeMain,
						},
					},
				},
			},
			localFile: &anime.LocalFile{
				Name:    "episode_01.mkv",
				MediaId: 123,
				Metadata: &anime.LocalFileMetadata{
					Episode:      0,
					AniDBEpisode: "",
					Type:         "",
				},
			},
			expectedResult: false,
			expectedMeta: &anime.LocalFileMetadata{
				Episode:      0,
				AniDBEpisode: "",
				Type:         "",
			},
		},
		{
			name: "should not apply rule when filename doesn't match",
			rules: []*HydrationRule{
				{
					Pattern: "",
					MediaID: 123,
					Files: []*HydrationFileRule{
						{
							Filename:     "episode_02.mkv",
							IsRegex:      false,
							Episode:      "2",
							AniDbEpisode: "",
							Type:         anime.LocalFileTypeMain,
						},
					},
				},
			},
			localFile: &anime.LocalFile{
				Name:    "episode_01.mkv",
				MediaId: 123,
				Metadata: &anime.LocalFileMetadata{
					Episode:      0,
					AniDBEpisode: "",
					Type:         "",
				},
			},
			expectedResult: false,
			expectedMeta: &anime.LocalFileMetadata{
				Episode:      0,
				AniDBEpisode: "",
				Type:         "",
			},
		},
		{
			name: "should not apply rule when filename regex doesn't match",
			rules: []*HydrationRule{
				{
					Pattern: "",
					MediaID: 123,
					Files: []*HydrationFileRule{
						{
							Filename:     "Special_(\\d+)\\.mkv",
							IsRegex:      true,
							Episode:      "$1",
							AniDbEpisode: "",
							Type:         anime.LocalFileTypeSpecial,
						},
					},
				},
			},
			localFile: &anime.LocalFile{
				Name:    "episode_01.mkv",
				MediaId: 123,
				Metadata: &anime.LocalFileMetadata{
					Episode:      0,
					AniDBEpisode: "",
					Type:         "",
				},
			},
			expectedResult: false,
			expectedMeta: &anime.LocalFileMetadata{
				Episode:      0,
				AniDBEpisode: "",
				Type:         "",
			},
		},
		{
			name: "should skip rule with no pattern and no media ID",
			rules: []*HydrationRule{
				{
					Pattern: "",
					MediaID: 0,
					Files: []*HydrationFileRule{
						{
							Filename:     "episode_01.mkv",
							IsRegex:      false,
							Episode:      "1",
							AniDbEpisode: "",
							Type:         anime.LocalFileTypeMain,
						},
					},
				},
			},
			localFile: &anime.LocalFile{
				Name:    "episode_01.mkv",
				MediaId: 123,
				Metadata: &anime.LocalFileMetadata{
					Episode:      0,
					AniDBEpisode: "",
					Type:         "",
				},
			},
			expectedResult: false,
			expectedMeta: &anime.LocalFileMetadata{
				Episode:      0,
				AniDBEpisode: "",
				Type:         "",
			},
		},
		{
			name: "should apply rule with pattern and media ID both matching",
			rules: []*HydrationRule{
				{
					Pattern: ".*episode.*",
					MediaID: 123,
					Files: []*HydrationFileRule{
						{
							Filename:     "episode_01.mkv",
							IsRegex:      false,
							Episode:      "10",
							AniDbEpisode: "10",
							Type:         anime.LocalFileTypeMain,
						},
					},
				},
			},
			localFile: &anime.LocalFile{
				Name:    "episode_01.mkv",
				MediaId: 123,
				Metadata: &anime.LocalFileMetadata{
					Episode:      0,
					AniDBEpisode: "",
					Type:         "",
				},
			},
			expectedResult: true,
			expectedMeta: &anime.LocalFileMetadata{
				Episode:      10,
				AniDBEpisode: "10",
				Type:         anime.LocalFileTypeMain,
			},
		},
		{
			name: "should apply rule with only pattern (no media ID)",
			rules: []*HydrationRule{
				{
					Pattern: ".*NC.*",
					MediaID: 0,
					Files: []*HydrationFileRule{
						{
							Filename:     "NC_OP.mkv",
							IsRegex:      false,
							Episode:      "0",
							AniDbEpisode: "NC",
							Type:         anime.LocalFileTypeNC,
						},
					},
				},
			},
			localFile: &anime.LocalFile{
				Name:    "NC_OP.mkv",
				MediaId: 456,
				Metadata: &anime.LocalFileMetadata{
					Episode:      0,
					AniDBEpisode: "",
					Type:         "",
				},
			},
			expectedResult: true,
			expectedMeta: &anime.LocalFileMetadata{
				Episode:      0,
				AniDBEpisode: "NC",
				Type:         anime.LocalFileTypeNC,
			},
		},
		{
			name: "should apply only episode when other fields are empty",
			rules: []*HydrationRule{
				{
					Pattern: "",
					MediaID: 111,
					Files: []*HydrationFileRule{
						{
							Filename:     "test.mkv",
							IsRegex:      false,
							Episode:      "7",
							AniDbEpisode: "",
							Type:         "",
						},
					},
				},
			},
			localFile: &anime.LocalFile{
				Name:    "test.mkv",
				MediaId: 111,
				Metadata: &anime.LocalFileMetadata{
					Episode:      0,
					AniDBEpisode: "",
					Type:         "",
				},
			},
			expectedResult: true,
			expectedMeta: &anime.LocalFileMetadata{
				Episode:      7,
				AniDBEpisode: "",
				Type:         "",
			},
		},
		{
			name: "should handle invalid episode number gracefully",
			rules: []*HydrationRule{
				{
					Pattern: "",
					MediaID: 333,
					Files: []*HydrationFileRule{
						{
							Filename:     "test.mkv",
							IsRegex:      false,
							Episode:      "invalid",
							AniDbEpisode: "A1",
							Type:         anime.LocalFileTypeMain,
						},
					},
				},
			},
			localFile: &anime.LocalFile{
				Name:    "test.mkv",
				MediaId: 333,
				Metadata: &anime.LocalFileMetadata{
					Episode:      0,
					AniDBEpisode: "",
					Type:         "",
				},
			},
			expectedResult: true,
			expectedMeta: &anime.LocalFileMetadata{
				Episode:      0, // Should remain 0 if episode is invalid
				AniDBEpisode: "A1",
				Type:         anime.LocalFileTypeMain,
			},
		},
		{
			name: "should match different files with different rules",
			rules: []*HydrationRule{
				{
					Pattern: "",
					MediaID: 444,
					Files: []*HydrationFileRule{
						{
							Filename:     "episode_01.mkv",
							IsRegex:      false,
							Episode:      "1",
							AniDbEpisode: "First",
							Type:         anime.LocalFileTypeMain,
						},
						{
							Filename:     "special_01.mkv",
							IsRegex:      false,
							Episode:      "2",
							AniDbEpisode: "Second",
							Type:         anime.LocalFileTypeSpecial,
						},
					},
				},
			},
			localFile: &anime.LocalFile{
				Name:    "episode_01.mkv",
				MediaId: 444,
				Metadata: &anime.LocalFileMetadata{
					Episode:      0,
					AniDBEpisode: "",
					Type:         "",
				},
			},
			expectedResult: true,
			expectedMeta: &anime.LocalFileMetadata{
				Episode:      1,
				AniDBEpisode: "First",
				Type:         anime.LocalFileTypeMain,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Hydration: HydrationConfig{
					Rules: tt.rules,
				},
			}

			fh := &FileHydrator{
				Logger:            logger,
				ScanSummaryLogger: scanSummaryLogger,
				Config:            config,
				hydrationRules:    make(map[string]*compiledHydrationRule),
			}

			// Precompile the rules
			fh.precompileRules()

			// Apply the hydration rule
			result := fh.applyHydrationRule(tt.localFile)

			assert.Equal(t, tt.expectedResult, result, "Expected result mismatch")

			assert.Equal(t, tt.expectedMeta.Episode, tt.localFile.Metadata.Episode, "Episode mismatch")
			assert.Equal(t, tt.expectedMeta.AniDBEpisode, tt.localFile.Metadata.AniDBEpisode, "AniDBEpisode mismatch")
			assert.Equal(t, tt.expectedMeta.Type, tt.localFile.Metadata.Type, "Type mismatch")
		})
	}
}

func TestFileHydrator_evaluateCalcExpressions(t *testing.T) {
	logger := util.NewLogger()
	scanSummaryLogger := summary.NewScanSummaryLogger()

	fh := &FileHydrator{
		Logger:            logger,
		ScanSummaryLogger: scanSummaryLogger,
		Config:            &Config{},
		hydrationRules:    make(map[string]*compiledHydrationRule),
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "should evaluate simple subtraction",
			input:    "S{calc(12-11)}",
			expected: "S1",
		},
		{
			name:     "should evaluate simple addition",
			input:    "S{calc(10+5)}",
			expected: "S15",
		},
		{
			name:     "should evaluate multiplication",
			input:    "E{calc(2*3)}",
			expected: "E6",
		},
		{
			name:     "should evaluate division",
			input:    "E{calc(10/2)}",
			expected: "E5",
		},
		{
			name:     "should handle multiple calc expressions",
			input:    "S{calc(12-11)}E{calc(5+5)}",
			expected: "S1E10",
		},
		{
			name:     "should preserve text without calc expressions",
			input:    "S1E5",
			expected: "S1E5",
		},
		{
			name:     "should handle invalid expressions gracefully",
			input:    "S{calc(invalid)}",
			expected: "S{calc(invalid)}",
		},
		{
			name:     "should handle division by zero gracefully",
			input:    "S{calc(10/0)}",
			expected: "S{calc(10/0)}",
		},
		{
			name:     "should handle empty calc expression",
			input:    "S{calc()}",
			expected: "S{calc()}",
		},
		{
			name:     "should handle nested parentheses",
			input:    "E{calc(10+5)}",
			expected: "E15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fh.evaluateCalcExpressions(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFileHydrator_applyHydrationRule_WithCalcExpressions(t *testing.T) {
	logger := util.NewLogger()
	scanSummaryLogger := summary.NewScanSummaryLogger()

	tests := []struct {
		name           string
		rules          []*HydrationRule
		localFile      *anime.LocalFile
		expectedResult bool
		expectedMeta   *anime.LocalFileMetadata
	}{
		{
			name: "should evaluate calc expression in aniDbEpisode",
			rules: []*HydrationRule{
				{
					Pattern: "",
					MediaID: 123,
					Files: []*HydrationFileRule{
						{
							Filename:     "Episode_(\\d+)\\.mkv",
							IsRegex:      true,
							Episode:      "$1",
							AniDbEpisode: "S{calc($1-11)}",
							Type:         anime.LocalFileTypeMain,
						},
					},
				},
			},
			localFile: &anime.LocalFile{
				Name:    "Episode_12.mkv",
				MediaId: 123,
				Metadata: &anime.LocalFileMetadata{
					Episode:      0,
					AniDBEpisode: "",
					Type:         "",
				},
			},
			expectedResult: true,
			expectedMeta: &anime.LocalFileMetadata{
				Episode:      12,
				AniDBEpisode: "S1",
				Type:         anime.LocalFileTypeMain,
			},
		},
		{
			name: "should handle calc with addition",
			rules: []*HydrationRule{
				{
					Pattern: "",
					MediaID: 456,
					Files: []*HydrationFileRule{
						{
							Filename:     "EP(\\d+)\\.mkv",
							IsRegex:      true,
							Episode:      "$1",
							AniDbEpisode: "S{calc($1+10)}",
							Type:         anime.LocalFileTypeMain,
						},
					},
				},
			},
			localFile: &anime.LocalFile{
				Name:    "EP5.mkv",
				MediaId: 456,
				Metadata: &anime.LocalFileMetadata{
					Episode:      0,
					AniDBEpisode: "",
					Type:         "",
				},
			},
			expectedResult: true,
			expectedMeta: &anime.LocalFileMetadata{
				Episode:      5,
				AniDBEpisode: "S15",
				Type:         anime.LocalFileTypeMain,
			},
		},
		{
			name: "should handle multiple capture groups with calc",
			rules: []*HydrationRule{
				{
					Pattern: "",
					MediaID: 789,
					Files: []*HydrationFileRule{
						{
							Filename:     "S(\\d+)E(\\d+)\\.mkv",
							IsRegex:      true,
							Episode:      "$2",
							AniDbEpisode: "S{calc($1-1)}E$2",
							Type:         anime.LocalFileTypeMain,
						},
					},
				},
			},
			localFile: &anime.LocalFile{
				Name:    "S2E10.mkv",
				MediaId: 789,
				Metadata: &anime.LocalFileMetadata{
					Episode:      0,
					AniDBEpisode: "",
					Type:         "",
				},
			},
			expectedResult: true,
			expectedMeta: &anime.LocalFileMetadata{
				Episode:      10,
				AniDBEpisode: "S1E10",
				Type:         anime.LocalFileTypeMain,
			},
		},
		{
			name: "should evaluate calc in episode field",
			rules: []*HydrationRule{
				{
					Pattern: "",
					MediaID: 999,
					Files: []*HydrationFileRule{
						{
							Filename:     "test.mkv",
							IsRegex:      false,
							Episode:      "calc(5+5)",
							AniDbEpisode: "E10",
							Type:         anime.LocalFileTypeMain,
						},
					},
				},
			},
			localFile: &anime.LocalFile{
				Name:    "test.mkv",
				MediaId: 999,
				Metadata: &anime.LocalFileMetadata{
					Episode:      0,
					AniDBEpisode: "",
					Type:         "",
				},
			},
			expectedResult: true,
			expectedMeta: &anime.LocalFileMetadata{
				Episode:      10,
				AniDBEpisode: "E10",
				Type:         anime.LocalFileTypeMain,
			},
		},
		{
			name: "should not evaluate malformed calc in episode field",
			rules: []*HydrationRule{
				{
					Pattern: "",
					MediaID: 999,
					Files: []*HydrationFileRule{
						{
							Filename:     "test.mkv",
							IsRegex:      false,
							Episode:      "{calc(5+5)}",
							AniDbEpisode: "E10",
							Type:         anime.LocalFileTypeMain,
						},
					},
				},
			},
			localFile: &anime.LocalFile{
				Name:    "test.mkv",
				MediaId: 999,
				Metadata: &anime.LocalFileMetadata{
					Episode:      0,
					AniDBEpisode: "",
					Type:         "",
				},
			},
			expectedResult: true,
			expectedMeta: &anime.LocalFileMetadata{
				Episode:      0, // stays 0
				AniDBEpisode: "E10",
				Type:         anime.LocalFileTypeMain,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Hydration: HydrationConfig{
					Rules: tt.rules,
				},
			}

			fh := &FileHydrator{
				Logger:            logger,
				ScanSummaryLogger: scanSummaryLogger,
				Config:            config,
				hydrationRules:    make(map[string]*compiledHydrationRule),
			}

			fh.precompileRules()
			result := fh.applyHydrationRule(tt.localFile)

			assert.Equal(t, tt.expectedResult, result, "Expected result mismatch")
			assert.Equal(t, tt.expectedMeta.Episode, tt.localFile.Metadata.Episode, "Episode mismatch")
			assert.Equal(t, tt.expectedMeta.AniDBEpisode, tt.localFile.Metadata.AniDBEpisode, "AniDBEpisode mismatch")
			assert.Equal(t, tt.expectedMeta.Type, tt.localFile.Metadata.Type, "Type mismatch")
		})
	}
}
