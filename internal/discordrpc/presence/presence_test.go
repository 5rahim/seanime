package discordrpc_presence

import (
	"testing"

	"seanime/internal/api/anilist"

	"github.com/stretchr/testify/require"
)

func TestNewAnimeActivityIncludesEpisodeDetails(t *testing.T) {
	title := "Test Anime"
	image := "https://example.com/cover.jpg"
	totalEpisodes := 12
	nextEpisode := 9

	activity := NewAnimeActivity(&anilist.BaseAnime{
		ID:       123,
		Episodes: &totalEpisodes,
		Title: &anilist.BaseAnime_Title{
			UserPreferred: &title,
		},
		CoverImage: &anilist.BaseAnime_CoverImage{
			ExtraLarge: &image,
		},
		NextAiringEpisode: &anilist.BaseAnime_NextAiringEpisode{
			Episode: nextEpisode,
		},
	}, 8, "The Turning Point", 30, 1440)

	require.Equal(t, 123, activity.ID)
	require.Equal(t, title, activity.Title)
	require.Equal(t, image, activity.Image)
	require.Equal(t, 8, activity.EpisodeNumber)
	require.Equal(t, 30, activity.Progress)
	require.Equal(t, 1440, activity.Duration)
	require.Equal(t, totalEpisodes, *activity.TotalEpisodes)
	require.Equal(t, nextEpisode-1, *activity.CurrentEpisodeCount)
	require.Equal(t, "The Turning Point", *activity.EpisodeTitle)
}

func TestNewAnimeActivityOmitsEmptyEpisodeTitle(t *testing.T) {
	activity := NewAnimeActivity(&anilist.BaseAnime{}, 1, "", 0, 0)

	require.Nil(t, activity.EpisodeTitle)
}
