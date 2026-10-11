package anime_test

import (
	"errors"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/database/db"
	"seanime/internal/extension"
	"seanime/internal/library/anime"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/util"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntryWithoutMetadata(t *testing.T) {
	tests := []struct {
		name     string
		count    *int
		progress int
		local    []anime.TestLocalFileEpisode
		airing   int
		want     []int
	}{
		{name: "known count", count: new(4), want: []int{1, 2, 3, 4}},
		{
			name:     "watched and downloaded episodes",
			count:    new(4),
			progress: 1,
			local:    []anime.TestLocalFileEpisode{{Episode: 3, AniDBEpisode: "3", Type: anime.LocalFileTypeMain}},
			want:     []int{2, 4},
		},
		{name: "unknown count"},
		{name: "airing episodes", count: new(12), airing: 4, want: []int{1, 2, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newAnimeTestWrapper(t)
			const mediaID = 154587
			listEntry := patchAnimeCollectionEntry(t, h.animeCollection, mediaID, anilist.AnimeCollectionEntryPatch{
				Progress: new(tt.progress),
				Status:   new(anilist.MediaListStatusCurrent),
			})
			listEntry.Media.Episodes = tt.count
			listEntry.Media.NextAiringEpisode = nil
			listEntry.Media.Status = new(anilist.MediaStatusFinished)
			if tt.airing > 0 {
				listEntry.Media.Status = new(anilist.MediaStatusReleasing)
				listEntry.Media.NextAiringEpisode = &anilist.BaseAnime_NextAiringEpisode{Episode: tt.airing}
			}
			h.metadataProvider.errors[mediaID] = errors.New("metadata not found")
			localFiles := anime.NewTestLocalFiles(anime.TestLocalFileGroup{
				LibraryPath:      "/anime",
				FilePathTemplate: "/anime/Frieren/episode-%ep.mkv",
				MediaID:          mediaID,
				Episodes:         tt.local,
			})

			entry, err := anime.NewEntry(t.Context(), &anime.NewEntryOptions{
				MediaId:             mediaID,
				LocalFiles:          localFiles,
				AnimeCollection:     h.animeCollection,
				PlatformRef:         h.platformRef,
				MetadataProviderRef: h.metadataProviderRef,
			})
			require.NoError(t, err)
			require.Len(t, entry.Episodes, len(tt.local))
			if tt.want == nil {
				require.Nil(t, entry.EntryDownloadInfo)
				return
			}
			require.NotNil(t, entry.EntryDownloadInfo)
			var got []int
			for _, ep := range entry.EntryDownloadInfo.EpisodesToDownload {
				got = append(got, ep.EpisodeNumber)
				require.Equal(t, ep.EpisodeNumber, ep.Episode.ProgressNumber)
				require.False(t, ep.Episode.IsDownloaded)
				require.Nil(t, ep.Episode.LocalFile)
			}
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

// TestNewAnimeEntry tests /library/entry endpoint.
// /!\ MAKE SURE TO HAVE THE MEDIA ADDED TO YOUR LIST TEST ACCOUNT LISTS
func TestNewAnimeEntry(t *testing.T) {
	logger := util.NewLogger()

	database, err := db.NewDatabase(t.TempDir(), "test", logger)
	assert.NoError(t, err)

	metadataProvider := metadata_provider.NewTestProvider(t, database)

	tests := []struct {
		name                              string
		mediaId                           int
		localFiles                        []*anime.LocalFile
		currentProgress                   int
		expectedNextEpisodeNumber         int
		expectedNextEpisodeProgressNumber int
	}{
		{
			name:    "Sousou no Frieren",
			mediaId: 154587,
			localFiles: anime.NewTestLocalFiles(
				anime.TestLocalFileGroup{
					LibraryPath:      "E:/Anime",
					FilePathTemplate: "E:\\Anime\\Sousou no Frieren\\[SubsPlease] Sousou no Frieren - %ep (1080p) [F02B9CEE].mkv",
					MediaID:          154587,
					Episodes: []anime.TestLocalFileEpisode{
						{Episode: 1, AniDBEpisode: "1", Type: anime.LocalFileTypeMain},
						{Episode: 2, AniDBEpisode: "2", Type: anime.LocalFileTypeMain},
						{Episode: 3, AniDBEpisode: "3", Type: anime.LocalFileTypeMain},
						{Episode: 4, AniDBEpisode: "4", Type: anime.LocalFileTypeMain},
						{Episode: 5, AniDBEpisode: "5", Type: anime.LocalFileTypeMain},
					},
				},
			),
			currentProgress:                   4,
			expectedNextEpisodeNumber:         5,
			expectedNextEpisodeProgressNumber: 5,
		},
		{
			name:    "Mushoku Tensei II Isekai Ittara Honki Dasu",
			mediaId: 146065,
			localFiles: anime.NewTestLocalFiles(
				anime.TestLocalFileGroup{
					LibraryPath:      "E:/Anime",
					FilePathTemplate: "E:/Anime/Mushoku Tensei II Isekai Ittara Honki Dasu/[SubsPlease] Mushoku Tensei S2 - 00 (1080p) [9C362DC3].mkv",
					MediaID:          146065,
					Episodes: []anime.TestLocalFileEpisode{
						{Episode: 0, AniDBEpisode: "S1", Type: anime.LocalFileTypeMain},
						{Episode: 1, AniDBEpisode: "1", Type: anime.LocalFileTypeMain},
						{Episode: 2, AniDBEpisode: "2", Type: anime.LocalFileTypeMain},
						{Episode: 3, AniDBEpisode: "3", Type: anime.LocalFileTypeMain},
						{Episode: 4, AniDBEpisode: "4", Type: anime.LocalFileTypeMain},
						{Episode: 5, AniDBEpisode: "5", Type: anime.LocalFileTypeMain},
						{Episode: 6, AniDBEpisode: "6", Type: anime.LocalFileTypeMain},
						{Episode: 7, AniDBEpisode: "7", Type: anime.LocalFileTypeMain},
						{Episode: 8, AniDBEpisode: "8", Type: anime.LocalFileTypeMain},
						{Episode: 9, AniDBEpisode: "9", Type: anime.LocalFileTypeMain},
						{Episode: 10, AniDBEpisode: "10", Type: anime.LocalFileTypeMain},
						{Episode: 11, AniDBEpisode: "11", Type: anime.LocalFileTypeMain},
						{Episode: 12, AniDBEpisode: "12", Type: anime.LocalFileTypeMain},
					},
				},
			),
			currentProgress:                   0,
			expectedNextEpisodeNumber:         0,
			expectedNextEpisodeProgressNumber: 1,
		},
	}

	anilistClient := anilist.NewTestAnilistClient()
	anilistPlatform := anilist_platform.NewAnilistPlatform(util.NewRef(anilistClient), util.NewRef(extension.NewUnifiedBank()), logger, database)
	animeCollection, err := anilistPlatform.GetAnimeCollection(t.Context(), false)
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			anilist.PatchAnimeCollectionEntry(animeCollection, tt.mediaId, anilist.AnimeCollectionEntryPatch{
				Progress: new(tt.currentProgress), // Mock progress
			})

			entry, err := anime.NewEntry(t.Context(), &anime.NewEntryOptions{
				MediaId:             tt.mediaId,
				LocalFiles:          tt.localFiles,
				AnimeCollection:     animeCollection,
				PlatformRef:         util.NewRef(anilistPlatform),
				MetadataProviderRef: util.NewRef(metadataProvider),
			})

			if assert.NoErrorf(t, err, "Failed to get mock data") {

				if assert.NoError(t, err) {

					// Mock progress is 4
					nextEp, found := entry.FindNextEpisode()
					if assert.True(t, found, "did not find next episode") {
						assert.Equal(t, tt.expectedNextEpisodeNumber, nextEp.EpisodeNumber, "next episode number mismatch")
						assert.Equal(t, tt.expectedNextEpisodeProgressNumber, nextEp.ProgressNumber, "next episode progress number mismatch")
					}

					t.Logf("Found %v episodes", len(entry.Episodes))

				}

			}

		})

	}
}
