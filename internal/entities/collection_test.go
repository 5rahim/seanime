package entities

import (
	"context"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/anizip"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewLibraryCollection(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()

	anilistCollection, err := anilistClientWrapper.AnimeCollection(context.Background(), nil)

	if assert.NoError(t, err) {

		// Mock Anilist collection and local files
		// User is currently watching Sousou no Frieren and One Piece
		lfs := make([]*LocalFile, 0)

		// Sousou no Frieren
		// 7 episodes downloaded, 4 watched
		mediaId := 154587
		lfs = append(lfs, MockHydratedLocalFiles(
			MockGenerateHydratedLocalFileGroupOptions("E:/Anime", "E:\\Anime\\Sousou no Frieren\\[SubsPlease] Sousou no Frieren - %ep (1080p) [F02B9CEE].mkv", mediaId, []MockHydratedLocalFileWrapperOptionsMetadata{
				{metadataEpisode: 1, metadataAniDbEpisode: "1", metadataType: LocalFileTypeMain},
				{metadataEpisode: 2, metadataAniDbEpisode: "2", metadataType: LocalFileTypeMain},
				{metadataEpisode: 3, metadataAniDbEpisode: "3", metadataType: LocalFileTypeMain},
				{metadataEpisode: 4, metadataAniDbEpisode: "4", metadataType: LocalFileTypeMain},
				{metadataEpisode: 5, metadataAniDbEpisode: "5", metadataType: LocalFileTypeMain},
				{metadataEpisode: 6, metadataAniDbEpisode: "6", metadataType: LocalFileTypeMain},
				{metadataEpisode: 7, metadataAniDbEpisode: "7", metadataType: LocalFileTypeMain},
			}),
		)...)
		anilist.TestModifyAnimeCollectionEntry(anilistCollection, mediaId, anilist.TestModifyAnimeCollectionEntryInput{
			Status:   lo.ToPtr(anilist.MediaListStatusCurrent),
			Progress: lo.ToPtr(4), // Mock progress
		})

		// One Piece
		// Downloaded 1070-1075 but only watched up until 1060
		mediaId = 21
		lfs = append(lfs, MockHydratedLocalFiles(
			MockGenerateHydratedLocalFileGroupOptions("E:/Anime", "E:\\Anime\\One Piece\\[SubsPlease] One Piece - %ep (1080p) [F02B9CEE].mkv", mediaId, []MockHydratedLocalFileWrapperOptionsMetadata{
				{metadataEpisode: 1070, metadataAniDbEpisode: "1070", metadataType: LocalFileTypeMain},
				{metadataEpisode: 1071, metadataAniDbEpisode: "1071", metadataType: LocalFileTypeMain},
				{metadataEpisode: 1072, metadataAniDbEpisode: "1072", metadataType: LocalFileTypeMain},
				{metadataEpisode: 1073, metadataAniDbEpisode: "1073", metadataType: LocalFileTypeMain},
				{metadataEpisode: 1074, metadataAniDbEpisode: "1074", metadataType: LocalFileTypeMain},
				{metadataEpisode: 1075, metadataAniDbEpisode: "1075", metadataType: LocalFileTypeMain},
			}),
		)...)
		anilist.TestModifyAnimeCollectionEntry(anilistCollection, mediaId, anilist.TestModifyAnimeCollectionEntryInput{
			Status:   lo.ToPtr(anilist.MediaListStatusCurrent),
			Progress: lo.ToPtr(1060), // Mock progress
		})

		// Add unmatched local files
		mediaId = 0
		lfs = append(lfs, MockHydratedLocalFiles(
			MockGenerateHydratedLocalFileGroupOptions("E:/Anime", "E:\\Anime\\Unmatched\\[SubsPlease] Unmatched - %ep (1080p) [F02B9CEE].mkv", mediaId, []MockHydratedLocalFileWrapperOptionsMetadata{
				{metadataEpisode: 1, metadataAniDbEpisode: "1", metadataType: LocalFileTypeMain},
				{metadataEpisode: 2, metadataAniDbEpisode: "2", metadataType: LocalFileTypeMain},
				{metadataEpisode: 3, metadataAniDbEpisode: "3", metadataType: LocalFileTypeMain},
				{metadataEpisode: 4, metadataAniDbEpisode: "4", metadataType: LocalFileTypeMain},
			}),
		)...)

		libraryCollection, err := NewLibraryCollection(&NewLibraryCollectionOptions{
			AnilistCollection:    anilistCollection,
			LocalFiles:           lfs,
			AnizipCache:          anizip.NewCache(),
			AnilistClientWrapper: anilistClientWrapper,
		})

		if assert.NoError(t, err) {

			assert.Equal(t, 1, len(libraryCollection.ContinueWatchingList)) // Only Sousou no Frieren is in the continue watching list
			assert.Equal(t, 4, len(libraryCollection.UnmatchedLocalFiles))  // 4 unmatched local files

		}
	}

}
