package anime_test

import (
	"seanime/internal/api/anilist"
	"seanime/internal/library/anime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewMissingEpisodes(t *testing.T) {
	// missing episodes now collapse each show down to the next thing you need,
	// and anything silenced should be split into its own list.
	h := newAnimeTestWrapper(t)

	localFiles := anime.NewTestLocalFiles(
		anime.TestLocalFileGroup{
			LibraryPath:      "/Anime",
			FilePathTemplate: "/Anime/Frieren/%ep.mkv",
			MediaID:          154587,
			Episodes: []anime.TestLocalFileEpisode{
				{Episode: 1, AniDBEpisode: "1", Type: anime.LocalFileTypeMain},
				{Episode: 2, AniDBEpisode: "2", Type: anime.LocalFileTypeMain},
				{Episode: 3, AniDBEpisode: "3", Type: anime.LocalFileTypeMain},
				{Episode: 4, AniDBEpisode: "4", Type: anime.LocalFileTypeMain},
				{Episode: 5, AniDBEpisode: "5", Type: anime.LocalFileTypeMain},
			},
		},
		anime.TestLocalFileGroup{
			LibraryPath:      "/Anime",
			FilePathTemplate: "/Anime/Mushoku/%ep.mkv",
			MediaID:          146065,
			Episodes: []anime.TestLocalFileEpisode{
				{Episode: 0, AniDBEpisode: "S1", Type: anime.LocalFileTypeMain},
				{Episode: 1, AniDBEpisode: "1", Type: anime.LocalFileTypeMain},
				{Episode: 2, AniDBEpisode: "2", Type: anime.LocalFileTypeMain},
			},
		},
		anime.TestLocalFileGroup{
			LibraryPath:      "/Anime",
			FilePathTemplate: "/Anime/OnePiece/%ep.mkv",
			MediaID:          21,
			Episodes: []anime.TestLocalFileEpisode{
				{Episode: 1069, AniDBEpisode: "1069", Type: anime.LocalFileTypeMain},
			},
		},
	)

	// frieren should surface as a normal missing-episodes card.
	patchAnimeCollectionEntry(t, h.animeCollection, 154587, anilist.AnimeCollectionEntryPatch{
		Status:            new(anilist.MediaListStatusCurrent),
		Progress:          new(4),
		AiredEpisodes:     new(10),
		NextAiringEpisode: &anilist.BaseAnime_NextAiringEpisode{Episode: 11},
	})
	h.setEpisodeMetadata(t, 154587, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil)

	// mushoku follows the episode-zero discrepancy path, but this one is silenced.
	patchAnimeCollectionEntry(t, h.animeCollection, 146065, anilist.AnimeCollectionEntryPatch{
		Status:            new(anilist.MediaListStatusCurrent),
		Progress:          new(1),
		AiredEpisodes:     new(6),
		NextAiringEpisode: &anilist.BaseAnime_NextAiringEpisode{Episode: 7},
	})
	h.setEpisodeMetadata(t, 146065, []int{1, 2, 3, 4, 5}, map[string]int{"S1": 1})

	// dropped entries should never show up here.
	patchAnimeCollectionEntry(t, h.animeCollection, 21, anilist.AnimeCollectionEntryPatch{
		Status:            new(anilist.MediaListStatusDropped),
		Progress:          new(1060),
		AiredEpisodes:     new(1100),
		NextAiringEpisode: &anilist.BaseAnime_NextAiringEpisode{Episode: 1101},
	})

	missing := h.newMissingEpisodes(t, localFiles, []int{146065})

	require.Len(t, missing.Episodes, 1)
	require.Equal(t, 154587, missing.Episodes[0].BaseAnime.ID)
	require.Equal(t, 6, missing.Episodes[0].EpisodeNumber)
	require.Equal(t, "Episode 6 & 4 more", missing.Episodes[0].DisplayTitle)
	require.True(t, missing.Episodes[0].IsMissingGroup)

	require.Len(t, missing.SilencedEpisodes, 1)
	require.Equal(t, 146065, missing.SilencedEpisodes[0].BaseAnime.ID)
	require.Equal(t, 3, missing.SilencedEpisodes[0].EpisodeNumber)
	require.Equal(t, "Episode 3 & 2 more", missing.SilencedEpisodes[0].DisplayTitle)
	require.True(t, missing.SilencedEpisodes[0].IsMissingGroup)
}
