import { AL_MangaDetailsById_Media, Manga_Entry, Nullish } from "@/api/generated/types"
import { MediaCardGrid } from "@/app/(main)/_features/media/_components/media-card-grid"
import { MediaEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"
import capitalize from "lodash/capitalize"
import React from "react"

type MangaRecommendationsProps = {
    entry: Nullish<Manga_Entry>
    details: Nullish<AL_MangaDetailsById_Media>
    maxCol?: number
}

export function MangaRecommendations(props: MangaRecommendationsProps) {

    const {
        entry,
        details,
        maxCol,
        ...rest
    } = props

    const anime = details?.relations?.edges?.filter(Boolean)?.filter(edge => edge?.node?.type === "ANIME" &&
        (edge?.node?.format === "TV" || edge?.node?.format === "MOVIE" || edge?.node?.format === "TV_SHORT"))?.slice(0, 3)

    const recommendations = details?.recommendations?.edges?.map(edge => edge?.node?.mediaRecommendation)?.filter(Boolean)?.slice(0, 6) || []

    if (!entry || !details) return null

    return (
        <div className="space-y-4" data-manga-recommendations-container>
            {!!anime?.length && (
                <>
                    <h2>Relations</h2>
                    <MediaCardGrid maxCol={maxCol}>
                        {anime?.toSorted((a, b) => (a.node?.format === "TV" && b.node?.format !== "TV")
                            ? -1
                            : (a.node?.format !== "TV" && b.node?.format === "TV") ? 1 : 0).map(edge => {
                            return <div key={edge?.node?.id!} className="col-span-1">
                                <MediaEntryCard
                                    media={edge?.node!}
                                    showLibraryBadge
                                    showTrailer
                                    overlay={<p
                                        className="font-semibold text-white bg-gray-950 z-[-1] absolute right-0 w-fit px-4 py-1.5 text-center !bg-opacity-90 text-sm lg:text-base rounded-none rounded-bl-lg"
                                    >{capitalize(edge.relationType || "").replace("_", " ")}{edge?.node?.format === "MOVIE" ? " (Movie)" : ""}</p>}
                                    type="anime"
                                />
                            </div>
                        })}
                    </MediaCardGrid>
                </>
            )}
            {recommendations.length > 0 && <>
                <h2>Recommendations</h2>
                <MediaCardGrid maxCol={maxCol}>
                    {recommendations.map(media => {
                        return <div key={media.id} className="col-span-1">
                            <MediaEntryCard
                                media={media!}
                                type="manga"
                            />
                        </div>
                    })}
                </MediaCardGrid>
            </>}
        </div>
    )
}

