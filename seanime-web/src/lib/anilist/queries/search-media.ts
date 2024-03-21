import { ListMediaQuery, MediaFormat, MediaSeason, MediaSort, MediaStatus } from "@/lib/anilist/gql/graphql"
import axios from "axios"

type SearchAnilistMediaListOptions = {
    page?: number | null
    search?: string | null
    perPage?: number | null
    sort?: MediaSort[] | null
    status?: MediaStatus[] | null
    genres?: string[] | null
    averageScoreGreater?: number | null
    season?: MediaSeason | null
    seasonYear?: number | null
    format?: MediaFormat | null
}

export async function searchAnilistMediaList(options: SearchAnilistMediaListOptions): Promise<ListMediaQuery | undefined> {
    const {
        page,
        search,
        perPage,
        sort,
        status,
        genres,
        averageScoreGreater,
        season,
        seasonYear,
        format,
    } = options

    const query = `
    query ListMedia(
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

    try {
        let response = await axios
            .post("https://graphql.anilist.co", {
                query,
                variables: {
                    page,
                    search,
                    perPage,
                    sort,
                    status,
                    genres,
                    averageScore_greater: averageScoreGreater,
                    season,
                    seasonYear,
                    format,
                },
            })
        return response.data?.data
    }
    catch (error) {
        console.error(error)
        return undefined

    }
}
