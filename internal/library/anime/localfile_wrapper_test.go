package anime_test

import (
	"cmp"
	"github.com/stretchr/testify/assert"
	"seanime/internal/library/anime"
	"slices"
	"testing"
)

func TestLocalFileWrapperEntry(t *testing.T) {

	lfs := anime.MockHydratedLocalFiles(
		anime.MockGenerateHydratedLocalFileGroupOptions("/mnt/anime/", "/mnt/anime/One Piece/One Piece - %ep.mkv", 21, []anime.MockHydratedLocalFileWrapperOptionsMetadata{
			{MetadataEpisode: 1070, MetadataAniDbEpisode: "1070", MetadataType: anime.LocalFileTypeMain},
			{MetadataEpisode: 1071, MetadataAniDbEpisode: "1071", MetadataType: anime.LocalFileTypeMain},
			{MetadataEpisode: 1072, MetadataAniDbEpisode: "1072", MetadataType: anime.LocalFileTypeMain},
			{MetadataEpisode: 1073, MetadataAniDbEpisode: "1073", MetadataType: anime.LocalFileTypeMain},
			{MetadataEpisode: 1074, MetadataAniDbEpisode: "1074", MetadataType: anime.LocalFileTypeMain},
		}),
		anime.MockGenerateHydratedLocalFileGroupOptions("/mnt/anime/", "/mnt/anime/Blue Lock/Blue Lock - %ep.mkv", 22222, []anime.MockHydratedLocalFileWrapperOptionsMetadata{
			{MetadataEpisode: 1, MetadataAniDbEpisode: "1", MetadataType: anime.LocalFileTypeMain},
			{MetadataEpisode: 2, MetadataAniDbEpisode: "2", MetadataType: anime.LocalFileTypeMain},
			{MetadataEpisode: 3, MetadataAniDbEpisode: "3", MetadataType: anime.LocalFileTypeMain},
		}),
		anime.MockGenerateHydratedLocalFileGroupOptions("/mnt/anime/", "/mnt/anime/Kimi ni Todoke/Kimi ni Todoke - %ep.mkv", 9656, []anime.MockHydratedLocalFileWrapperOptionsMetadata{
			{MetadataEpisode: 0, MetadataAniDbEpisode: "S1", MetadataType: anime.LocalFileTypeMain},
			{MetadataEpisode: 1, MetadataAniDbEpisode: "1", MetadataType: anime.LocalFileTypeMain},
			{MetadataEpisode: 2, MetadataAniDbEpisode: "2", MetadataType: anime.LocalFileTypeMain},
		}),
	)

	tests := []struct {
		name                              string
		mediaId                           int
		expectedNbMainLocalFiles          int
		expectedLatestEpisode             int
		expectedEpisodeNumberAfterEpisode []int
	}{
		{
			name:                              "One Piece",
			mediaId:                           21,
			expectedNbMainLocalFiles:          5,
			expectedLatestEpisode:             1074,
			expectedEpisodeNumberAfterEpisode: []int{1071, 1072},
		},
		{
			name:                              "Blue Lock",
			mediaId:                           22222,
			expectedNbMainLocalFiles:          3,
			expectedLatestEpisode:             3,
			expectedEpisodeNumberAfterEpisode: []int{2, 3},
		},
	}

	lfw := anime.NewLocalFileWrapper(lfs)

	// Not empty
	if assert.Greater(t, len(lfw.GetLocalEntries()), 0) {

		for _, tt := range tests {

			// Can get by id
			entry, ok := lfw.GetLocalEntryById(tt.mediaId)
			if assert.Truef(t, ok, "could not find entry for %s", tt.name) {

				assert.Equalf(t, tt.mediaId, entry.GetMediaId(), "media id does not match for %s", tt.name)

				// Can get main local files
				mainLfs, ok := entry.GetMainLocalFiles()
				if assert.Truef(t, ok, "could not find main local files for %s", tt.name) {

					// Number of main local files matches
					assert.Equalf(t, tt.expectedNbMainLocalFiles, len(mainLfs), "number of main local files does not match for %s", tt.name)

					// Can find latest episode
					latest, ok := entry.FindLatestLocalFile()
					if assert.Truef(t, ok, "could not find latest local file for %s", tt.name) {
						assert.Equalf(t, tt.expectedLatestEpisode, latest.GetEpisodeNumber(), "latest episode does not match for %s", tt.name)
					}

					// Can find successive episodes
					firstEp, ok := entry.FindLocalFileWithEpisodeNumber(tt.expectedEpisodeNumberAfterEpisode[0])
					if assert.True(t, ok) {
						secondEp, ok := entry.FindNextEpisode(firstEp)
						if assert.True(t, ok) {
							assert.Equal(t, tt.expectedEpisodeNumberAfterEpisode[1], secondEp.GetEpisodeNumber(), "second episode does not match for %s", tt.name)
						}
					}

				}

			}

		}

	}

}

func TestLocalFileWrapperEntryProgressNumber(t *testing.T) {

	lfs := anime.MockHydratedLocalFiles(
		anime.MockGenerateHydratedLocalFileGroupOptions("/mnt/anime/", "/mnt/anime/Kimi ni Todoke/Kimi ni Todoke - %ep.mkv", 9656, []anime.MockHydratedLocalFileWrapperOptionsMetadata{
			{MetadataEpisode: 0, MetadataAniDbEpisode: "S1", MetadataType: anime.LocalFileTypeMain},
			{MetadataEpisode: 1, MetadataAniDbEpisode: "1", MetadataType: anime.LocalFileTypeMain},
			{MetadataEpisode: 2, MetadataAniDbEpisode: "2", MetadataType: anime.LocalFileTypeMain},
		}),
		anime.MockGenerateHydratedLocalFileGroupOptions("/mnt/anime/", "/mnt/anime/Kimi ni Todoke/Kimi ni Todoke - %ep.mkv", 9656_2, []anime.MockHydratedLocalFileWrapperOptionsMetadata{
			{MetadataEpisode: 1, MetadataAniDbEpisode: "S1", MetadataType: anime.LocalFileTypeMain},
			{MetadataEpisode: 2, MetadataAniDbEpisode: "1", MetadataType: anime.LocalFileTypeMain},
			{MetadataEpisode: 3, MetadataAniDbEpisode: "2", MetadataType: anime.LocalFileTypeMain},
		}),
	)

	tests := []struct {
		name                              string
		mediaId                           int
		expectedNbMainLocalFiles          int
		expectedLatestEpisode             int
		expectedEpisodeNumberAfterEpisode []int
		expectedProgressNumbers           []int
	}{
		{
			name:                              "Kimi ni Todoke",
			mediaId:                           9656,
			expectedNbMainLocalFiles:          3,
			expectedLatestEpisode:             2,
			expectedEpisodeNumberAfterEpisode: []int{1, 2},
			expectedProgressNumbers:           []int{1, 2, 3}, // S1 -> 1, 1 -> 2, 2 -> 3
		},
		{
			name:                              "Kimi ni Todoke 2",
			mediaId:                           9656_2,
			expectedNbMainLocalFiles:          3,
			expectedLatestEpisode:             3,
			expectedEpisodeNumberAfterEpisode: []int{2, 3},
			expectedProgressNumbers:           []int{1, 2, 3}, // S1 -> 1, 1 -> 2, 2 -> 3
		},
	}

	lfw := anime.NewLocalFileWrapper(lfs)

	// Not empty
	if assert.Greater(t, len(lfw.GetLocalEntries()), 0) {

		for _, tt := range tests {

			// Can get by id
			entry, ok := lfw.GetLocalEntryById(tt.mediaId)
			if assert.Truef(t, ok, "could not find entry for %s", tt.name) {

				assert.Equalf(t, tt.mediaId, entry.GetMediaId(), "media id does not match for %s", tt.name)

				// Can get main local files
				mainLfs, ok := entry.GetMainLocalFiles()
				if assert.Truef(t, ok, "could not find main local files for %s", tt.name) {

					// Number of main local files matches
					assert.Equalf(t, tt.expectedNbMainLocalFiles, len(mainLfs), "number of main local files does not match for %s", tt.name)

					// Can find latest episode
					latest, ok := entry.FindLatestLocalFile()
					if assert.Truef(t, ok, "could not find latest local file for %s", tt.name) {
						assert.Equalf(t, tt.expectedLatestEpisode, latest.GetEpisodeNumber(), "latest episode does not match for %s", tt.name)
					}

					// Can find successive episodes
					firstEp, ok := entry.FindLocalFileWithEpisodeNumber(tt.expectedEpisodeNumberAfterEpisode[0])
					if assert.True(t, ok) {
						secondEp, ok := entry.FindNextEpisode(firstEp)
						if assert.True(t, ok) {
							assert.Equal(t, tt.expectedEpisodeNumberAfterEpisode[1], secondEp.GetEpisodeNumber(), "second episode does not match for %s", tt.name)
						}
					}

					slices.SortStableFunc(mainLfs, func(i *anime.LocalFile, j *anime.LocalFile) int {
						return cmp.Compare(i.GetEpisodeNumber(), j.GetEpisodeNumber())
					})
					for idx, lf := range mainLfs {
						progressNum := entry.GetProgressNumber(lf)

						assert.Equalf(t, tt.expectedProgressNumbers[idx], progressNum, "progress number does not match for %s", tt.name)
					}

				}

			}

		}

	}

}
