import { AL_MediaDetailsById_Media, Anime_MediaEntry } from "@/api/generated/types"
import { serverStatusAtom } from "@/app/(main)/_atoms/server-status.atoms"
import { MediaEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { Badge } from "@/components/ui/badge"
import { useAtomValue } from "jotai/react"
import capitalize from "lodash/capitalize"
import React from "react"

type RelationsRecommendationsAccordionProps = {
    entry: Anime_MediaEntry | undefined
    details: AL_MediaDetailsById_Media | undefined
}

export function RelationsRecommendationsAccordion(props: RelationsRecommendationsAccordionProps) {

    const {
        entry,
        details,
        ...rest
    } = props

    const serverStatus = useAtomValue(serverStatusAtom)

    const sourceManga = serverStatus?.settings?.library?.enableManga
        ? entry?.media?.relations?.edges?.find(edge => edge?.relationType === "SOURCE" && edge?.node?.format === "MANGA")?.node
        : undefined

    const relations = (entry?.media?.relations?.edges?.map(edge => edge) || [])
        .filter(Boolean)
        .filter(n => (n.node?.format === "TV" || n.node?.format === "OVA" || n.node?.format === "MOVIE" || n.node?.format === "SPECIAL") && (n.relationType === "PREQUEL" || n.relationType === "SEQUEL" || n.relationType === "PARENT" || n.relationType === "SIDE_STORY" || n.relationType === "ALTERNATIVE" || n.relationType === "ADAPTATION"))

    const recommendations = details?.recommendations?.edges?.map(edge => edge?.node?.mediaRecommendation)?.filter(Boolean)?.slice(0, 6) || []

    if (!entry || !details) return null

    return (
        <>
            <Accordion
                type="multiple"
                className="space-y-2 lg:space-y-4"
                itemClass="border-none"
                triggerClass="rounded-[--radius] bg-gray-900 bg-opacity-80 dark:bg-gray-900 dark:bg-opacity-80 hover:bg-gray-800 dark:hover:bg-gray-800 hover:bg-opacity-100 dark:hover:bg-opacity-100"
            >
                {(!!sourceManga || relations.length > 0) && (
                    <AccordionItem value="relations">
                        <AccordionTrigger>
                            Relations
                        </AccordionTrigger>
                        <AccordionContent className="pt-6 px-0">
                            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-3 min-[2000px]:grid-cols-4 gap-4">
                                {sourceManga && <div className="col-span-1">
                                    <MediaEntryCard
                                        media={sourceManga}
                                        overlay={<Badge
                                            className="font-semibold text-white bg-gray-950 !bg-opacity-90 rounded-md text-base rounded-bl-none rounded-tr-none"
                                            intent="gray"
                                            size="lg"
                                        >Source (Manga)</Badge>}
                                        type="manga"
                                    />
                                </div>}
                                {relations.slice(0, 4).map(edge => {
                                    return <div key={edge.node?.id} className="col-span-1">
                                        <MediaEntryCard
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
                                            type="anime"
                                        />
                                    </div>
                                })}
                            </div>
                        </AccordionContent>
                    </AccordionItem>
                )}
                <AccordionItem value="recommendations">
                    <AccordionTrigger>
                        Recommendations
                    </AccordionTrigger>
                    <AccordionContent className="pt-6 px-0">
                        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-3 min-[2000px]:grid-cols-4 gap-4">
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
                        </div>
                    </AccordionContent>
                </AccordionItem>
            </Accordion>
        </>
    )
}

