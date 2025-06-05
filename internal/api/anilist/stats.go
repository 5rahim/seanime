package anilist

import (
	"context"
	"seanime/internal/util"
)

type (
	Stats struct {
		AnimeStats *AnimeStats `json:"animeStats"`
		MangaStats *MangaStats `json:"mangaStats"`
	}

	AnimeStats struct {
		Count           int                     `json:"count"`
		MinutesWatched  int                     `json:"minutesWatched"`
		EpisodesWatched int                     `json:"episodesWatched"`
		MeanScore       float64                 `json:"meanScore"`
		Genres          []*UserGenreStats       `json:"genres"`
		Formats         []*UserFormatStats      `json:"formats"`
		Statuses        []*UserStatusStats      `json:"statuses"`
		Studios         []*UserStudioStats      `json:"studios"`
		Scores          []*UserScoreStats       `json:"scores"`
		StartYears      []*UserStartYearStats   `json:"startYears"`
		ReleaseYears    []*UserReleaseYearStats `json:"releaseYears"`
	}

	MangaStats struct {
		Count        int                     `json:"count"`
		ChaptersRead int                     `json:"chaptersRead"`
		MeanScore    float64                 `json:"meanScore"`
		Genres       []*UserGenreStats       `json:"genres"`
		Statuses     []*UserStatusStats      `json:"statuses"`
		Scores       []*UserScoreStats       `json:"scores"`
		StartYears   []*UserStartYearStats   `json:"startYears"`
		ReleaseYears []*UserReleaseYearStats `json:"releaseYears"`
	}
)

func GetStats(ctx context.Context, stats *ViewerStats) (ret *Stats, err error) {
	defer util.HandlePanicInModuleWithError("api/anilist/GetStats", &err)

	allStats := stats.GetViewer().GetStatistics()

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
			StartYears:      allStats.GetAnime().GetStartYears(),
			ReleaseYears:    allStats.GetAnime().GetReleaseYears(),
		},
		MangaStats: &MangaStats{
			Count:        allStats.GetManga().GetCount(),
			ChaptersRead: allStats.GetManga().GetChaptersRead(),
			MeanScore:    allStats.GetManga().GetMeanScore(),
			Genres:       allStats.GetManga().GetGenres(),
			Statuses:     allStats.GetManga().GetStatuses(),
			Scores:       allStats.GetManga().GetScores(),
			StartYears:   allStats.GetManga().GetStartYears(),
			ReleaseYears: allStats.GetManga().GetReleaseYears(),
		},
	}

	return ret, nil
}
