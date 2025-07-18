package anime_test

import (
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/library/anime"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestNewLibraryCollection(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())
	logger := util.NewLogger()
	metadataProvider := metadata.GetMockProvider(t)

	anilistClient := anilist.TestGetMockAnilistClient()
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilistClient, logger)

	animeCollection, err := anilistPlatform.GetAnimeCollection(t.Context(), false)

	if assert.NoError(t, err) {

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

		// One Piece
		// Downloaded 1070-1075 but only watched up until 1060
		mediaId = 21
		lfs = append(lfs, anime.MockHydratedLocalFiles(
			anime.MockGenerateHydratedLocalFileGroupOptions("E:/Anime", "E:\\Anime\\One Piece\\[SubsPlease] One Piece - %ep (1080p) [F02B9CEE].mkv", mediaId, []anime.MockHydratedLocalFileWrapperOptionsMetadata{
				{MetadataEpisode: 1070, MetadataAniDbEpisode: "1070", MetadataType: anime.LocalFileTypeMain},
				{MetadataEpisode: 1071, MetadataAniDbEpisode: "1071", MetadataType: anime.LocalFileTypeMain},
				{MetadataEpisode: 1072, MetadataAniDbEpisode: "1072", MetadataType: anime.LocalFileTypeMain},
				{MetadataEpisode: 1073, MetadataAniDbEpisode: "1073", MetadataType: anime.LocalFileTypeMain},
				{MetadataEpisode: 1074, MetadataAniDbEpisode: "1074", MetadataType: anime.LocalFileTypeMain},
				{MetadataEpisode: 1075, MetadataAniDbEpisode: "1075", MetadataType: anime.LocalFileTypeMain},
			}),
		)...)
		anilist.TestModifyAnimeCollectionEntry(animeCollection, mediaId, anilist.TestModifyAnimeCollectionEntryInput{
			Status:   lo.ToPtr(anilist.MediaListStatusCurrent),
			Progress: lo.ToPtr(1060), // Mock progress
		})

		// Add unmatched local files
		mediaId = 0
		lfs = append(lfs, anime.MockHydratedLocalFiles(
			anime.MockGenerateHydratedLocalFileGroupOptions("E:/Anime", "E:\\Anime\\Unmatched\\[SubsPlease] Unmatched - %ep (1080p) [F02B9CEE].mkv", mediaId, []anime.MockHydratedLocalFileWrapperOptionsMetadata{
				{MetadataEpisode: 1, MetadataAniDbEpisode: "1", MetadataType: anime.LocalFileTypeMain},
				{MetadataEpisode: 2, MetadataAniDbEpisode: "2", MetadataType: anime.LocalFileTypeMain},
				{MetadataEpisode: 3, MetadataAniDbEpisode: "3", MetadataType: anime.LocalFileTypeMain},
				{MetadataEpisode: 4, MetadataAniDbEpisode: "4", MetadataType: anime.LocalFileTypeMain},
			}),
		)...)

		libraryCollection, err := anime.NewLibraryCollection(t.Context(), &anime.NewLibraryCollectionOptions{
			AnimeCollection:  animeCollection,
			LocalFiles:       lfs,
			Platform:         anilistPlatform,
			MetadataProvider: metadataProvider,
		})

		if assert.NoError(t, err) {

			assert.Equal(t, 1, len(libraryCollection.ContinueWatchingList)) // Only Sousou no Frieren is in the continue watching list
			assert.Equal(t, 4, len(libraryCollection.UnmatchedLocalFiles))  // 4 unmatched local files

		}
	}

}
