import { AL_AnimeDetailsById_Media, Anime_Entry, Nullish } from "@/api/generated/types"
import { MediaCardGrid } from "@/app/(main)/_features/media/_components/media-card-grid"
import { MediaEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import capitalize from "lodash/capitalize"
import React from "react"

type RelationsRecommendationsSectionProps = {
    entry: Nullish<Anime_Entry>
    details: Nullish<AL_AnimeDetailsById_Media>
    containerRef?: React.RefObject<HTMLElement>
}

export function RelationsRecommendationsSection(props: RelationsRecommendationsSectionProps) {

    const {
        entry,
        details,
        containerRef,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    const sourceManga = React.useMemo(() => {
        return serverStatus?.settings?.library?.enableManga
            ? details?.relations?.edges?.find(edge => (edge?.relationType === "SOURCE" || edge?.relationType === "ADAPTATION") && edge?.node?.format === "MANGA")?.node
            : undefined
    }, [details?.relations?.edges, serverStatus?.settings?.library?.enableManga])

    const relations = React.useMemo(() => (details?.relations?.edges?.map(edge => edge) || [])
        .filter(Boolean)
            .filter(n => (n.node?.format === "TV" || n.node?.format === "OVA" || n.node?.format === "MOVIE" || n.node?.format === "SPECIAL") && (n.relationType === "PREQUEL" || n.relationType === "SEQUEL" || n.relationType === "PARENT" || n.relationType === "SIDE_STORY" || n.relationType === "ALTERNATIVE" || n.relationType === "ADAPTATION")),
        [details?.relations?.edges])

    const recommendations = React.useMemo(() => details?.recommendations?.edges?.map(edge => edge?.node?.mediaRecommendation)?.filter(Boolean) || [],
        [details?.recommendations?.edges])

    if (!entry || !details) return null

    return (
        <>
            {/*{(!!sourceManga || relations.length > 0 || recommendations.length > 0) && <Separator />}*/}
            {(!!sourceManga || relations.length > 0) && (
                <>
                    <h2>Relations</h2>
                    <MediaCardGrid>
                        {!!sourceManga && <div className="col-span-1">
                            <MediaEntryCard
                                media={sourceManga!}
                                overlay={<p
                                    className="font-semibold text-white bg-gray-950 z-[-1] absolute right-0 w-fit px-4 py-1.5 text-center !bg-opacity-90 text-sm lg:text-base rounded-none rounded-bl-lg border border-t-0 border-r-0"
                                >Manga</p>}
                                type="manga"
                            /></div>}
                        {relations.slice(0, 4).map(edge => {
                            return <div key={edge.node?.id} className="col-span-1">
                                <MediaEntryCard
                                    media={edge.node!}
                                    overlay={<p
                                        className="font-semibold text-white bg-gray-950 z-[-1] absolute right-0 w-fit px-4 py-1.5 text-center !bg-opacity-90 text-sm lg:text-base rounded-none rounded-bl-lg border border-t-0 border-r-0"
                                    >{edge.node?.format === "MOVIE"
                                        ? capitalize(edge.relationType || "").replace("_", " ") + " (Movie)"
                                        : capitalize(edge.relationType || "").replace("_", " ")}</p>}
                                    showLibraryBadge
                                    showTrailer
                                    type="anime"
                                />
                            </div>
                        })}
                    </MediaCardGrid>
                </>
            )}
            {recommendations.length > 0 && <>
                <h2>Recommendations</h2>
                <MediaCardGrid>
                    {recommendations.map(media => {
                        return <div key={media.id} className="col-span-1">
                            <MediaEntryCard
                                media={media!}
                                showLibraryBadge
                                showTrailer
                                type="anime"
                            />
                        </div>
                    })}
                </MediaCardGrid>
            </>}
        </>
    )
}
