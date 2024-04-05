import { MangaEntry } from "@/app/(main)/manga/_lib/manga.types"
import { AnimeListItem } from "@/components/shared/anime-list-item"
import { Badge } from "@/components/ui/badge"
import { MangaDetailsByIdQuery } from "@/lib/anilist/gql/graphql"
import capitalize from "lodash/capitalize"
import React from "react"

type MangaRecommendationsProps = {
    entry: MangaEntry | undefined
    details: MangaDetailsByIdQuery["Media"] | undefined
}

export function MangaRecommendations(props: MangaRecommendationsProps) {

    const {
        entry,
        details,
        ...rest
    } = props

    const anime = entry?.media?.relations?.edges?.filter(Boolean)?.filter(edge => edge?.node?.type === "ANIME" &&
        (edge?.node?.format === "TV" || edge?.node?.format === "MOVIE" || edge?.node?.format === "TV_SHORT"))?.slice(0, 3)

    const recommendations = details?.recommendations?.edges?.map(edge => edge?.node?.mediaRecommendation)?.filter(Boolean)?.slice(0, 6) || []

    if (!entry || !details) return null

    return (
        <div className="space-y-4">
            {!!anime?.length && (
                <>
                    <h3>Relations</h3>
                    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-2 2xl:grid-cols-3 gap-4">
                        {anime?.toSorted((a, b) => (a.node?.format === "TV" && b.node?.format !== "TV")
                            ? -1
                            : (a.node?.format !== "TV" && b.node?.format === "TV") ? 1 : 0).map(edge => {
                            return <div key={edge?.node?.id!} className="col-span-1">
                                <AnimeListItem
                                    media={edge?.node!}
                                    showLibraryBadge
                                    showTrailer
                                    overlay={<Badge
                                        className="font-semibold text-white bg-gray-950 !bg-opacity-90 rounded-md text-base rounded-bl-none rounded-tr-none"
                                        intent="gray"
                                        size="lg"
                                    >{edge?.node?.format === "MOVIE"
                                        ? capitalize(edge.relationType || "").replace("_", " ") + " (Movie)"
                                        : capitalize(edge.relationType || "").replace("_", " ")}</Badge>}
                                />
                            </div>
                        })}
                    </div>
                </>
            )}
            <h3>Recommendations</h3>
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-2 2xl:grid-cols-3 gap-4">
                {recommendations.map(media => {
                    return <div key={media.id} className="col-span-1">
                        <AnimeListItem
                            media={media!}
                            isManga
                        />
                    </div>
                })}
            </div>
        </div>
    )
}

