import { MediaDetailsByIdQuery } from "@/lib/anilist/gql/graphql"

export function getMediaEntryRankings(details: MediaDetailsByIdQuery["Media"] | undefined) {

    const seasonMostPopular = details?.rankings?.find(r => (!!r?.season || !!r?.year) && r?.type === "POPULAR" && r.rank <= 10)
    const allTimeHighestRated = details?.rankings?.find(r => !!r?.allTime && r?.type === "RATED" && r.rank <= 100)
    const seasonHighestRated = details?.rankings?.find(r => (!!r?.season || !!r?.year) && r?.type === "RATED" && r.rank <= 5)
    const allTimeMostPopular = details?.rankings?.find(r => !!r?.allTime && r?.type === "POPULAR" && r.rank <= 100)

    return {
        seasonHighestRated,
        seasonMostPopular,
        allTimeHighestRated,
        allTimeMostPopular,
    }

}
