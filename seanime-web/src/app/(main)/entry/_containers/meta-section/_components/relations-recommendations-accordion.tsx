import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { Badge } from "@/components/ui/badge"
import { MediaDetailsByIdQuery } from "@/lib/anilist/gql/graphql"
import { MediaEntry } from "@/lib/server/types"
import capitalize from "lodash/capitalize"
import Image from "next/image"
import Link from "next/link"
import React from "react"

type RelationsRecommendationsAccordionProps = {
    entry: MediaEntry | undefined
    details: MediaDetailsByIdQuery["Media"] | undefined
}

export function RelationsRecommendationsAccordion(props: RelationsRecommendationsAccordionProps) {

    const {
        entry,
        details,
        ...rest
    } = props


    const relations = (entry?.media?.relations?.edges?.map(edge => edge) || [])
        .filter(Boolean)
        .filter(n => (n.node?.format === "TV" || n.node?.format === "OVA" || n.node?.format === "MOVIE" || n.node?.format === "SPECIAL") && (n.relationType === "PREQUEL" || n.relationType === "SEQUEL" || n.relationType === "PARENT" || n.relationType === "SIDE_STORY" || n.relationType === "ALTERNATIVE" || n.relationType === "ADAPTATION"))

    if (!entry || !details) return null

    return (
        <>
            <Accordion
                type="multiple"
                className="space-y-2 lg:space-y-4"
                itemClass="border-none"
                triggerClass="rounded-[--radius] bg-gray-900 bg-opacity-80 dark:bg-gray-900 dark:bg-opacity-80 hover:bg-gray-800 dark:hover:bg-gray-800 hover:bg-opacity-100 dark:hover:bg-opacity-100"
            >
                {relations.length > 0 && (
                    <AccordionItem value="relations">
                        <AccordionTrigger>
                            Relations
                        </AccordionTrigger>
                        <AccordionContent>
                            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                                {relations.slice(0, 4).map(edge => {
                                    return <div key={edge.node?.id} className="col-span-1">
                                        <Link href={`/entry?id=${edge.node?.id}`}>
                                            {edge.node?.coverImage?.large && <div
                                                className="h-64 w-full flex-none rounded-md object-cover object-center relative overflow-hidden group/anime-list-item"
                                            >
                                                <Image
                                                    src={edge.node?.coverImage.large}
                                                    alt={""}
                                                    fill
                                                    quality={80}
                                                    priority
                                                    sizes="10rem"
                                                    className="object-cover object-center group-hover/anime-list-item:scale-110 transition"
                                                />
                                                <div
                                                    className="z-[5] absolute bottom-0 w-full h-[60%] bg-gradient-to-t from-black to-transparent"
                                                />
                                                <Badge
                                                    className="absolute left-2 top-2 font-semibold rounded-md text-[.95rem]"
                                                    intent="white-solid"
                                                >{edge.node?.format === "MOVIE"
                                                    ? capitalize(edge.relationType || "").replace("_", " ") + " (Movie)"
                                                    : capitalize(edge.relationType || "").replace("_", " ")}</Badge>
                                                <div className="p-2 z-[5] absolute bottom-0 w-full ">
                                                    <p className="font-semibold line-clamp-2 overflow-hidden">{edge.node?.title?.userPreferred}</p>
                                                </div>
                                            </div>}
                                        </Link>
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
                    <AccordionContent>
                        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                            {details?.recommendations?.edges?.map(edge => edge?.node?.mediaRecommendation).filter(Boolean).map(media => {
                                return <div key={media.id} className="col-span-1">
                                    <Link href={`/entry?id=${media.id}`}>
                                        {media.coverImage?.large && <div
                                            className="h-64 w-full flex-none rounded-md object-cover object-center relative overflow-hidden group/anime-list-item"
                                        >
                                            <Image
                                                src={media.coverImage.large}
                                                alt={""}
                                                fill
                                                quality={80}
                                                priority
                                                sizes="10rem"
                                                className="object-cover object-center group-hover/anime-list-item:scale-110 transition"
                                            />
                                            <div
                                                className="z-[5] absolute bottom-0 w-full h-[60%] bg-gradient-to-t from-black to-transparent"
                                            />
                                            <div className="p-2 z-[5] absolute bottom-0 w-full ">
                                                <p className="font-semibold line-clamp-2 overflow-hidden">{media.title?.userPreferred}</p>
                                            </div>
                                        </div>}
                                    </Link>
                                </div>
                            })}
                        </div>
                    </AccordionContent>
                </AccordionItem>
            </Accordion>
        </>
    )
}

