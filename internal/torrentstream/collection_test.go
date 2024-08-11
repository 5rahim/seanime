package torrentstream

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"seanime/internal/api/anilist"
	"seanime/internal/api/anizip"
	"seanime/internal/api/metadata"
	"seanime/internal/events"
	"seanime/internal/library/anime"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"testing"
)

func TestStreamCollection(t *testing.T) {
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t, test_utils.Anilist())

	logger := util.NewLogger()
	filecacher, err := filecache.NewCacher(t.TempDir())
	require.NoError(t, err)
	metadataProvider := metadata.TestGetMockProvider(t)
	anilistClient := anilist.TestGetMockAnilistClient()
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilistClient, util.NewLogger())
	anizipCache := anizip.NewCache()
	anilistPlatform.SetUsername(test_utils.ConfigData.Provider.AnilistUsername)
	animeCollection, err := anilistPlatform.GetAnimeCollection(false)
	require.NoError(t, err)
	require.NotNil(t, animeCollection)

	repo := NewRepository(&NewRepositoryOptions{
		Logger:             logger,
		AnizipCache:        anizip.NewCache(),
		BaseAnimeCache:     anilist.NewBaseAnimeCache(),
		CompleteAnimeCache: anilist.NewCompleteAnimeCache(),
		Platform:           anilistPlatform,
		MetadataProvider: metadata.NewProvider(&metadata.NewProviderOptions{
			Logger:     logger,
			FileCacher: filecacher,
		}),
		PlaybackManager: nil,
		WSEventManager:  events.NewMockWSEventManager(logger),
	})

	// Mock Anilist collection and local files
	// User is currently watching Sousou no Frieren and One Piece
	lfs := make([]*anime.LocalFile, 0)

	// Sousou no Frieren
	// 7 episodes downloaded, 4 watched
	mediaId := 154587
	lfs = append(lfs, anime.MockHydratedLocalFiles(
		anime.MockGenerateHydratedLocalFileGroupOptions("E:/Anime", "E:\\Anime\\Sousou no Frieren\\[SubsPlease] Sousou no Frieren - %ep (1080p) [F02B9CEE].mkv", mediaId, []anime.MockHydratedLocalFileWrapperOptionsMetadata{
			{MetadataEpisode: 1, MetadataAniDbEpisode: "1", MetadataType: anime.LocalFileTypeMain},
			{MetadataEpisode: 2, MetadataAniDbEpisode: "2", MetadataType: anime.LocalFileTypeMain},
			{MetadataEpisode: 3, MetadataAniDbEpisode: "3", MetadataType: anime.LocalFileTypeMain},
			{MetadataEpisode: 4, MetadataAniDbEpisode: "4", MetadataType: anime.LocalFileTypeMain},
			{MetadataEpisode: 5, MetadataAniDbEpisode: "5", MetadataType: anime.LocalFileTypeMain},
			{MetadataEpisode: 6, MetadataAniDbEpisode: "6", MetadataType: anime.LocalFileTypeMain},
			{MetadataEpisode: 7, MetadataAniDbEpisode: "7", MetadataType: anime.LocalFileTypeMain},
		}),
	)...)
	anilist.TestModifyAnimeCollectionEntry(animeCollection, mediaId, anilist.TestModifyAnimeCollectionEntryInput{
		Status:   lo.ToPtr(anilist.MediaListStatusCurrent),
		Progress: lo.ToPtr(4), // Mock progress
	})

	libraryCollection, err := anime.NewLibraryCollection(&anime.NewLibraryCollectionOptions{
		AnimeCollection:  animeCollection,
		LocalFiles:       lfs,
		AnizipCache:      anizipCache,
		Platform:         anilistPlatform,
		MetadataProvider: metadataProvider,
	})
	require.NoError(t, err)

	// Create the stream collection
	repo.HydrateStreamCollection(&HydrateStreamCollectionOptions{
		AnimeCollection:   animeCollection,
		LibraryCollection: libraryCollection,
		AnizipCache:       anizipCache,
	})
	spew.Dump(libraryCollection)

}
