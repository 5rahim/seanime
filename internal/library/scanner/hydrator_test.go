package scanner

import (
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/database/db"
	"seanime/internal/extension"
	"seanime/internal/library/anime"
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
	anilistClient := anilist.TestGetMockAnilistClient()
	anilistClientRef := util.NewRef(anilistClient)
	extensionBankRef := util.NewRef(extension.NewUnifiedBank())
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
				AllMedia:   allMedia,
				ScanLogger: scanLogger,
			})

			for _, nm := range mc.NormalizedMedia {
				t.Logf("media id: %d, title: %s", nm.ID, nm.GetTitleSafe())
			}

			// +---------------------+
			// |      Matcher        |
			// +---------------------+

			matcher := &Matcher{
				LocalFiles:         lfs,
				MediaContainer:     mc,
				CompleteAnimeCache: nil,
				Logger:             util.NewLogger(),
				ScanLogger:         scanLogger,
			}

			err = matcher.MatchLocalFilesWithMedia()
			if err != nil {
				t.Fatal("expected result, got error:", err.Error())
			}

			// +---------------------+
			// |    FileHydrator     |
			// +---------------------+

			fh := &FileHydrator{
				LocalFiles:          lfs,
				AllMedia:            mc.NormalizedMedia,
				CompleteAnimeCache:  completeAnimeCache,
				PlatformRef:         util.NewRef(anilistPlatform),
				AnilistRateLimiter:  anilistRateLimiter,
				MetadataProviderRef: util.NewRef(metadataProvider),
				Logger:              logger,
				ScanLogger:          scanLogger,
			}

			fh.HydrateMetadata()

			for _, lf := range fh.LocalFiles {
				t.Logf("local file: %s,\nmedia id: %d, type: %s\n", lf.Name, lf.MediaId, lf.GetType())
				assert.NotNil(t, lf.MediaId, "expected media id to be set")
				assert.Equal(t, tt.expectedMediaId, lf.MediaId, "expected media id %d, got %d", tt.expectedMediaId, lf.MediaId)
				assert.Equal(t, tt.expectedType, lf.GetType(), "expected file type %s, got %s", tt.expectedType, lf.GetType())

			}

		})
	}

}
