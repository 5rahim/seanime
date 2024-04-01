import { MangaEntry } from "@/app/(main)/manga/_lib/types"
import { AnimeListItem } from "@/components/shared/anime-list-item"
import { MangaDetailsByIdQuery } from "@/lib/anilist/gql/graphql"
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

    const anime = entry?.media?.relations?.edges?.map(edge => edge?.node).find(node => node?.type === "ANIME" && node?.format === "TV")

    const recommendations = details?.recommendations?.edges?.map(edge => edge?.node?.mediaRecommendation)?.filter(Boolean)?.slice(0, 6) || []

    if (!entry || !details) return null

    return (
        <div className="space-y-4">
            {!!anime && (
                <>
                    <h3>Relations</h3>
                    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-2 2xl:grid-cols-3 gap-4">
                        <div className="col-span-1">
                            <AnimeListItem
                                media={anime!}
                                showLibraryBadge
                                showTrailer
                            />
                        </div>
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

