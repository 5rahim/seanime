package anilist

import (
	"context"
	"github.com/seanime-app/seanime/internal/util"
)

type (
	Stats struct {
		AnimeStats *AnimeStats `json:"animeStats"`
		MangaStats *MangaStats `json:"mangaStats"`
	}

	AnimeStats struct {
		Count           int                `json:"count"`
		MinutesWatched  int                `json:"minutesWatched"`
		EpisodesWatched int                `json:"episodesWatched"`
		MeanScore       float64            `json:"meanScore"`
		Genres          []*UserGenreStats  `json:"genres"`   // Key is the genre name
		Formats         []*UserFormatStats `json:"formats"`  // Key is the format name
		Statuses        []*UserStatusStats `json:"statuses"` // Key is the status name
		Studios         []*UserStudioStats `json:"studios"`  // Key is the studio name
		Scores          []*UserScoreStats  `json:"scores"`
	}

	MangaStats struct {
		Count        int                `json:"count"`
		ChaptersRead int                `json:"chaptersRead"`
		MeanScore    float64            `json:"meanScore"`
		Genres       []*UserGenreStats  `json:"genres"`   // Key is the genre name
		Statuses     []*UserStatusStats `json:"statuses"` // Key is the status name
		Scores       []*UserScoreStats  `json:"scores"`
	}
)

func GetStats(ctx context.Context, client ClientWrapperInterface) (ret *Stats, err error) {
	defer util.HandlePanicInModuleWithError("api/anilist/GetStats", &err)

	resp, err := client.ViewerStats(ctx)
	if err != nil {
		return nil, err
	}

	allStats := resp.GetViewer().GetStatistics()

	ret = &Stats{
		AnimeStats: &AnimeStats{
			Count:           allStats.GetAnime().GetCount(),
			MinutesWatched:  allStats.GetAnime().GetMinutesWatched(),
			EpisodesWatched: allStats.GetAnime().GetEpisodesWatched(),
			MeanScore:       allStats.GetAnime().GetMeanScore(),
			Genres:          allStats.GetAnime().GetGenres(),
			Formats:         allStats.GetAnime().GetFormats(),
			Statuses:        allStats.GetAnime().GetStatuses(),
			Studios:         allStats.GetAnime().GetStudios(),
			Scores:          allStats.GetAnime().GetScores(),
		},
		MangaStats: &MangaStats{
			Count:        allStats.GetManga().GetCount(),
			ChaptersRead: allStats.GetManga().GetChaptersRead(),
			MeanScore:    allStats.GetManga().GetMeanScore(),
			Genres:       allStats.GetManga().GetGenres(),
			Statuses:     allStats.GetManga().GetStatuses(),
			Scores:       allStats.GetManga().GetScores(),
		},
	}

	return ret, nil
}
