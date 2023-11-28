import { ListRecentMediaQuery } from "@/lib/anilist/gql/graphql"
import axios from "axios"

type Options = {
    page?: number | null
    perPage?: number | null
    airingAt_greater?: number | null
    airingAt_lesser?: number | null
}

export async function getRecentMediaAirings(options: Options): Promise<ListRecentMediaQuery | undefined> {
    const {
        page,
        perPage,
        airingAt_greater,
        airingAt_lesser,
    } = options

    const query = `
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
                    perPage,
                    airingAt_greater,
                    airingAt_lesser,
                },
            })
        return response.data?.data
    } catch (error) {
        console.error(error)
        return undefined

    }
}