package torrentstream

import (
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/database/db"
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/library/anime"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/test_utils"
	"seanime/internal/torrents/torrent"
	"seanime/internal/util"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

func TestStreamCollection(t *testing.T) {
	t.Skip("Incomplete")
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t, test_utils.Anilist())

	database, err := db.NewDatabase(test_utils.ConfigData.Path.DataDir, test_utils.ConfigData.Database.Name, util.NewLogger())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	logger := util.NewLogger()
	metadataProvider := metadata_provider.GetFakeProvider(t, database)
	anilistClient := anilist.TestGetMockAnilistClient()
	anilistPlatform := anilist_platform.NewAnilistPlatform(util.NewRef(anilistClient), util.NewRef(extension.NewUnifiedBank()), logger, database)
	anilistPlatform.SetUsername(test_utils.ConfigData.Provider.AnilistUsername)
	animeCollection, err := anilistPlatform.GetAnimeCollection(t.Context(), false)
	require.NoError(t, err)
	require.NotNil(t, animeCollection)

	repo := NewRepository(&NewRepositoryOptions{
		Logger:              logger,
		BaseAnimeCache:      anilist.NewBaseAnimeCache(),
		CompleteAnimeCache:  anilist.NewCompleteAnimeCache(),
		PlatformRef:         util.NewRef(anilistPlatform),
		MetadataProviderRef: util.NewRef(metadataProvider),
		WSEventManager:      events.NewMockWSEventManager(logger),
		TorrentRepository:   &torrent.Repository{},
		PlaybackManager:     &playbackmanager.PlaybackManager{},
		Database:            database,
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
		Status:   new(anilist.MediaListStatusCurrent),
		Progress: new(4), // Mock progress
	})

	libraryCollection, err := anime.NewLibraryCollection(t.Context(), &anime.NewLibraryCollectionOptions{
		AnimeCollection:     animeCollection,
		LocalFiles:          lfs,
		PlatformRef:         util.NewRef(anilistPlatform),
		MetadataProviderRef: util.NewRef(metadataProvider),
	})
	require.NoError(t, err)

	// Create the stream collection
	repo.HydrateStreamCollection(&HydrateStreamCollectionOptions{
		AnimeCollection:   animeCollection,
		LibraryCollection: libraryCollection,
	})
	spew.Dump(libraryCollection)

}
