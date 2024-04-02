import { MediaEntry } from "@/app/(main)/(library)/_lib/anime-library.types"
import { serverStatusAtom } from "@/atoms/server-status"
import { AnimeListItem } from "@/components/shared/anime-list-item"
import { Badge } from "@/components/ui/badge"
import { Separator } from "@/components/ui/separator"
import { MediaDetailsByIdQuery } from "@/lib/anilist/gql/graphql"
import { useAtomValue } from "jotai/react"
import capitalize from "lodash/capitalize"
import React from "react"

type RelationsRecommendationsSectionProps = {
    entry: MediaEntry | undefined
    details: MediaDetailsByIdQuery["Media"] | undefined
}

export function RelationsRecommendationsSection(props: RelationsRecommendationsSectionProps) {

    const {
        entry,
        details,
        ...rest
    } = props

    const serverStatus = useAtomValue(serverStatusAtom)

    const sourceManga = serverStatus?.settings?.library?.enableManga
        ? entry?.media?.relations?.edges?.find(edge => (edge?.relationType === "SOURCE" || edge?.relationType === "ADAPTATION") && edge?.node?.format === "MANGA")?.node
        : undefined

    const relations = (entry?.media?.relations?.edges?.map(edge => edge) || [])
        .filter(Boolean)
        .filter(n => (n.node?.format === "TV" || n.node?.format === "OVA" || n.node?.format === "MOVIE" || n.node?.format === "SPECIAL") && (n.relationType === "PREQUEL" || n.relationType === "SEQUEL" || n.relationType === "PARENT" || n.relationType === "SIDE_STORY" || n.relationType === "ALTERNATIVE" || n.relationType === "ADAPTATION"))

    const recommendations = details?.recommendations?.edges?.map(edge => edge?.node?.mediaRecommendation)?.filter(Boolean) || []

    if (!entry || !details) return null

    return (
        <>
            {(!!sourceManga || relations.length > 0 || recommendations.length > 0) && <Separator />}
            {(!!sourceManga || relations.length > 0) && (
                <>
                    <h2>Relations</h2>
                    <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-7 min-[2000px]:grid-cols-8 gap-4">
                        {!!sourceManga && <div className="col-span-1">
                            <AnimeListItem
                                media={sourceManga!}
                                overlay={<Badge
                                    className="font-semibold text-white bg-gray-950 !bg-opacity-90 rounded-md text-base rounded-bl-none rounded-tr-none"
                                    intent="gray"
                                    size="lg"
                                >Manga</Badge>}
                                isManga
                            /></div>}
                        {relations.slice(0, 4).map(edge => {
                            return <div key={edge.node?.id} className="col-span-1">
                                <AnimeListItem
                                    media={edge.node!}
                                    overlay={<Badge
                                        className="font-semibold text-white bg-gray-950 !bg-opacity-90 rounded-md text-base rounded-bl-none rounded-tr-none"
                                        intent="gray"
                                        size="lg"
                                    >{edge.node?.format === "MOVIE"
                                        ? capitalize(edge.relationType || "").replace("_", " ") + " (Movie)"
                                        : capitalize(edge.relationType || "").replace("_", " ")}</Badge>}
                                    showLibraryBadge
                                    showTrailer
                                />
                            </div>
                        })}
                    </div>
                </>
            )}
            {recommendations.length > 0 && <>
                <h2>Recommendations</h2>
                <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-7 min-[2000px]:grid-cols-8 gap-4">
                    {recommendations.map(media => {
                        return <div key={media.id} className="col-span-1">
                            <AnimeListItem
                                media={media!}
                                showLibraryBadge
                                showTrailer
                            />
                        </div>
                    })}
                </div>
            </>}
        </>
    )
}
