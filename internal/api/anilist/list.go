package anilist

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
)

func ListMediaM(
	Page *int,
	Search *string,
	PerPage *int,
	Sort []*MediaSort,
	Status []*MediaStatus,
	Genres []*string,
	AverageScoreGreater *int,
	Season *MediaSeason,
	SeasonYear *int,
	Format *MediaFormat,
	logger *zerolog.Logger,
) (*ListMedia, error) {

	variables := map[string]interface{}{}
	if Page != nil {
		variables["page"] = *Page
	}
	if Search != nil {
		variables["search"] = *Search
	}
	if PerPage != nil {
		variables["perPage"] = *PerPage
	}
	if Sort != nil {
		variables["sort"] = Sort
	}
	if Status != nil {
		variables["status"] = Status
	}
	if Genres != nil {
		variables["genres"] = Genres
	}
	if AverageScoreGreater != nil {
		variables["averageScore_greater"] = *AverageScoreGreater
	}
	if Season != nil {
		variables["season"] = *Season
	}
	if SeasonYear != nil {
		variables["seasonYear"] = *SeasonYear
	}
	if Format != nil {
		variables["format"] = *Format
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"query":     ListMediaQuery,
		"variables": variables,
	})
	if err != nil {
		return nil, err
	}

	data, err := customQuery(requestBody, logger)
	if err != nil {
		return nil, err
	}

	var listMediaF ListMedia
	m, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(m, &listMediaF); err != nil {
		return nil, err
	}

	return &listMediaF, nil
}
func ListMangaM(
	Page *int,
	Search *string,
	PerPage *int,
	Sort []*MediaSort,
	Status []*MediaStatus,
	Genres []*string,
	AverageScoreGreater *int,
	Season *MediaSeason,
	SeasonYear *int,
	Format *MediaFormat,
	logger *zerolog.Logger,
) (*ListManga, error) {

	variables := map[string]interface{}{}
	if Page != nil {
		variables["page"] = *Page
	}
	if Search != nil {
		variables["search"] = *Search
	}
	if PerPage != nil {
		variables["perPage"] = *PerPage
	}
	if Sort != nil {
		variables["sort"] = Sort
	}
	if Status != nil {
		variables["status"] = Status
	}
	if Genres != nil {
		variables["genres"] = Genres
	}
	if AverageScoreGreater != nil {
		variables["averageScore_greater"] = *AverageScoreGreater
	}
	if Season != nil {
		variables["season"] = *Season
	}
	if SeasonYear != nil {
		variables["seasonYear"] = *SeasonYear
	}
	if Format != nil {
		variables["format"] = *Format
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"query":     ListMangaQuery,
		"variables": variables,
	})
	if err != nil {
		return nil, err
	}

	data, err := customQuery(requestBody, logger)
	if err != nil {
		return nil, err
	}

	var listMediaF ListManga
	m, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(m, &listMediaF); err != nil {
		return nil, err
	}

	return &listMediaF, nil
}

func ListRecentAiringMediaM(
	Page *int,
	Search *string,
	PerPage *int,
	AiringAtGreater *int,
	AiringAtLesser *int,
	logger *zerolog.Logger,
) (*ListRecentMedia, error) {

	variables := map[string]interface{}{}
	if Page != nil {
		variables["page"] = *Page
	}
	if Search != nil {
		variables["search"] = *Search
	}
	if PerPage != nil {
		variables["perPage"] = *PerPage
	}
	if AiringAtGreater != nil {
		variables["airingAt_greater"] = *AiringAtGreater
	}
	if AiringAtLesser != nil {
		variables["airingAt_lesser"] = *AiringAtLesser
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"query":     ListRecentAiringMediaQuery,
		"variables": variables,
	})
	if err != nil {
		return nil, err
	}

	data, err := customQuery(requestBody, logger)
	if err != nil {
		return nil, err
	}

	var listMediaF ListRecentMedia
	m, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(m, &listMediaF); err != nil {
		return nil, err
	}

	return &listMediaF, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func ListMediaCacheKey(
	Page *int,
	Search *string,
	PerPage *int,
	Sort []*MediaSort,
	Status []*MediaStatus,
	Genres []*string,
	AverageScoreGreater *int,
	Season *MediaSeason,
	SeasonYear *int,
	Format *MediaFormat,
) string {

	key := "ListMedia"
	if Page != nil {
		key += fmt.Sprintf("_%d", *Page)
	}
	if Search != nil {
		key += fmt.Sprintf("_%s", *Search)
	}
	if PerPage != nil {
		key += fmt.Sprintf("_%d", *PerPage)
	}
	if Sort != nil {
		key += fmt.Sprintf("_%v", Sort)
	}
	if Status != nil {
		key += fmt.Sprintf("_%v", Status)
	}
	if Genres != nil {
		key += fmt.Sprintf("_%v", Genres)
	}
	if AverageScoreGreater != nil {
		key += fmt.Sprintf("_%d", *AverageScoreGreater)
	}
	if Season != nil {
		key += fmt.Sprintf("_%s", *Season)
	}
	if SeasonYear != nil {
		key += fmt.Sprintf("_%d", *SeasonYear)
	}
	if Format != nil {
		key += fmt.Sprintf("_%s", *Format)
	}

	return key

}

func ListRecentAiringMediaCacheKey(
	Page *int,
	Search *string,
	PerPage *int,
	AiringAtGreater *int,
	AiringAtLesser *int,
) string {

	key := "ListRecentMedia"
	if Page != nil {
		key += fmt.Sprintf("_%d", *Page)
	}
	if Search != nil {
		key += fmt.Sprintf("_%s", *Search)
	}
	if PerPage != nil {
		key += fmt.Sprintf("_%d", *PerPage)
	}
	if AiringAtGreater != nil {
		key += fmt.Sprintf("_%d", *AiringAtGreater)
	}
	if AiringAtLesser != nil {
		key += fmt.Sprintf("_%d", *AiringAtLesser)
	}

	return key

}

const ListMediaQuery = `query ListMedia(
      $page: Int
      $search: String
      $perPage: Int
      $sort: [MediaSort]
      $status: [MediaStatus]
      $genres: [String]
      $averageScore_greater: Int
      $season: MediaSeason
      $seasonYear: Int
      $format: MediaFormat
    ) {
      Page(page: $page, perPage: $perPage) {
        pageInfo {
          hasNextPage
          total
          perPage
          currentPage
          lastPage
        }
        media(
          type: ANIME
          search: $search
          sort: $sort
          status_in: $status
          isAdult: false
          format: $format
          genre_in: $genres
          averageScore_greater: $averageScore_greater
          season: $season
          seasonYear: $seasonYear
          format_not: MUSIC
        ) {
          ...basicMedia
        }
      }
    }
    fragment basicMedia on Media {
      id
      idMal
      siteUrl
      status(version: 2)
      season
      type
      format
      bannerImage
      episodes
      synonyms
      isAdult
      countryOfOrigin
      meanScore
      description
      trailer {
        id
        site
        thumbnail
      }
      title {
        userPreferred
        romaji
        english
        native
      }
      coverImage {
        extraLarge
        large
        medium
        color
      }
      startDate {
        year
        month
        day
      }
      endDate {
        year
        month
        day
      }
      nextAiringEpisode {
        airingAt
        timeUntilAiring
        episode
      }
    }`

const ListMangaQuery = `query ListManga(
      $page: Int
      $search: String
      $perPage: Int
      $sort: [MediaSort]
      $status: [MediaStatus]
      $genres: [String]
      $averageScore_greater: Int
      $season: MediaSeason
      $seasonYear: Int
      $format: MediaFormat
    ) {
        Page(page: $page, perPage: $perPage){
		pageInfo{
		  hasNextPage
		  total
		  perPage
		  currentPage
		  lastPage
		},
		media(type: MANGA, isAdult: false, search: $search, sort: $sort, status_in: $status, format: $format, genre_in: $genres, averageScore_greater: $averageScore_greater, season: $season, seasonYear: $seasonYear, format_not: MUSIC){
		  ...basicManga
		}
	  }
    }
    fragment basicManga on Media {
  id
  idMal
  siteUrl
  status(version: 2)
  season
  type
  format
  bannerImage
  chapters
  volumes
  synonyms
  isAdult
  countryOfOrigin
  meanScore
  description
  title {
    userPreferred
    romaji
    english
    native
  }
  coverImage {
    extraLarge
    large
    medium
    color
  }
  startDate {
    year
    month
    day
  }
  endDate {
    year
    month
    day
  }
}`

const ListRecentAiringMediaQuery = `
    query ListRecentMedia($page: Int, $perPage: Int, $airingAt_greater: Int, $airingAt_lesser: Int){
        Page(page: $page, perPage: $perPage){
            pageInfo{
                hasNextPage
                total
                perPage
                currentPage
                lastPage
            },
            airingSchedules(notYetAired: false, sort: TIME_DESC, airingAt_greater: $airingAt_greater, airingAt_lesser: $airingAt_lesser){
                id
                airingAt
                episode
                timeUntilAiring
                media {
                    isAdult
                    ...basicMedia
                }
            }
        }
    }
    fragment basicMedia on Media {
      id
      idMal
      siteUrl
      status(version: 2)
      season
      type
      format
      bannerImage
      episodes
      synonyms
      isAdult
      countryOfOrigin
      meanScore
      description
      trailer {
        id
        site
        thumbnail
      }
      title {
        userPreferred
        romaji
        english
        native
      }
      coverImage {
        extraLarge
        large
        medium
        color
      }
      startDate {
        year
        month
        day
      }
      endDate {
        year
        month
        day
      }
      nextAiringEpisode {
        airingAt
        timeUntilAiring
        episode
      }
    }
  `
