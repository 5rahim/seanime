package entities

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocalFileWrapperEntry(t *testing.T) {

	lfs := MockHydratedLocalFiles(
		MockGenerateHydratedLocalFileGroupOptions("/mnt/anime/", "/mnt/anime/One Piece/One Piece - %ep.mkv", 21, []MockHydratedLocalFileWrapperOptionsMetadata{
			{metadataEpisode: 1070, metadataAniDbEpisode: "1070", metadataType: LocalFileTypeMain},
			{metadataEpisode: 1071, metadataAniDbEpisode: "1071", metadataType: LocalFileTypeMain},
			{metadataEpisode: 1072, metadataAniDbEpisode: "1072", metadataType: LocalFileTypeMain},
			{metadataEpisode: 1073, metadataAniDbEpisode: "1073", metadataType: LocalFileTypeMain},
			{metadataEpisode: 1074, metadataAniDbEpisode: "1074", metadataType: LocalFileTypeMain},
		}),
		MockGenerateHydratedLocalFileGroupOptions("/mnt/anime/", "/mnt/anime/Blue Lock/Blue Lock - %ep.mkv", 22222, []MockHydratedLocalFileWrapperOptionsMetadata{
			{metadataEpisode: 1, metadataAniDbEpisode: "1", metadataType: LocalFileTypeMain},
			{metadataEpisode: 2, metadataAniDbEpisode: "2", metadataType: LocalFileTypeMain},
			{metadataEpisode: 3, metadataAniDbEpisode: "3", metadataType: LocalFileTypeMain},
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

	lfw := NewLocalFileWrapper(lfs)

	// Not empty
	if assert.Greater(t, len(lfw.GetLocalEntries()), 0) {

		for _, tt := range tests {

			// Can get by id
			entry, ok := lfw.GetLocalEntryById(tt.mediaId)
			if assert.Truef(t, ok, "could not find entry for %s", tt.name) {

				assert.Equalf(t, tt.mediaId, entry.MediaId, "media id does not match for %s", tt.name)

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
