"use client"
import { LargeEpisodeListItem } from "@/components/shared/large-episode-list-item"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { HorizontalDraggableScroll } from "@/components/ui/horizontal-draggable-scroll"
import { Skeleton } from "@/components/ui/skeleton"
import { MediaEntryEpisode } from "@/lib/server/types"
import { useRouter } from "next/navigation"
import React from "react"
import { AiOutlineDownload } from "react-icons/ai"
import { IoLibrary } from "react-icons/io5"
import { LuBellOff } from "react-icons/lu"

export function MissingEpisodes({ isLoading, missingEpisodes, silencedEpisodes }: {
    missingEpisodes: MediaEntryEpisode[] | undefined,
    silencedEpisodes: MediaEntryEpisode[],
    isLoading: boolean
}) {
    const router = useRouter()

    if (!missingEpisodes?.length && !silencedEpisodes?.length) return null

    return (
        <>
            <AppLayoutStack spacing="lg">

                {!!missingEpisodes?.length && (
                    <>
                        <h2 className="flex gap-3 items-center"><IoLibrary /> Missing from your library</h2>

                        <HorizontalDraggableScroll>
                            {!isLoading && missingEpisodes?.map(episode => {
                                return <LargeEpisodeListItem
                                    key={episode.displayTitle + episode.basicMedia?.id}
                                    image={episode.episodeMetadata?.image}
                                    topTitle={episode.basicMedia?.title?.userPreferred}
                                    title={episode.displayTitle}
                                    meta={episode.episodeMetadata?.airDate ?? undefined}
                                    actionIcon={<AiOutlineDownload />}
                                    isInvalid={episode.isInvalid}
                                    onClick={() => {
                                        router.push(`/entry?id=${episode.basicMedia?.id}&download=${episode.episodeNumber}`)
                                    }}
                                />
                            })}
                            {isLoading && <>
                                <Skeleton
                                    className="rounded-md h-auto overflow-hidden aspect-[4/2] w-96 relative flex items-end flex-none"
                                />
                                <Skeleton
                                    className="rounded-md h-auto overflow-hidden aspect-[4/2] w-96 relative flex items-end flex-none"
                                />
                                <Skeleton
                                    className="rounded-md h-auto overflow-hidden aspect-[4/2] w-96 relative flex items-end flex-none"
                                />
                            </>}
                            {!isLoading && !missingEpisodes?.length && (
                                <div
                                    className="rounded-md h-auto overflow-hidden aspect-[4/2] w-96 relative flex items-center justify-center flex-none bg-gray-900 text-[--muted]"
                                >
                                    No missing episodes
                                </div>
                            )}
                        </HorizontalDraggableScroll>
                    </>
                )}

                {!!silencedEpisodes.length && (
                    <>

                        <Accordion
                            type="multiple"
                            defaultValue={[]}
                            triggerClass="py-2 dark:bg-[--background] px-0 dark:hover:bg-transparent text-lg dark:text-[--muted] dark:hover:text-white"
                        >
                            <AccordionItem value="item-1">
                                <AccordionTrigger>
                                    <p className="flex gap-3 items-center text-lg text-inherit"><LuBellOff /> Silenced episodes</p>
                                </AccordionTrigger>
                                <AccordionContent className="bg-gray-950 rounded-[--radius]">
                                    <HorizontalDraggableScroll>
                                        {!isLoading && silencedEpisodes?.map(episode => {
                                            return <LargeEpisodeListItem
                                                key={episode.displayTitle + episode.basicMedia?.id}
                                                image={episode.episodeMetadata?.image}
                                                topTitle={episode.basicMedia?.title?.userPreferred}
                                                title={episode.displayTitle}
                                                meta={episode.episodeMetadata?.airDate ?? undefined}
                                                actionIcon={<AiOutlineDownload />}
                                                isInvalid={episode.isInvalid}
                                                onClick={() => {
                                                    router.push(`/entry?id=${episode.basicMedia?.id}&download=${episode.episodeNumber}`)
                                                }}
                                            />
                                        })}
                                    </HorizontalDraggableScroll>
                                </AccordionContent>
                            </AccordionItem>
                        </Accordion>
                    </>
                )}

            </AppLayoutStack>
        </>
    )
}
