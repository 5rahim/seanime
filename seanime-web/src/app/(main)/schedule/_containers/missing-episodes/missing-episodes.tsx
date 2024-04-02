"use client"
import { MediaEntryEpisode } from "@/app/(main)/(library)/_lib/anime-library.types"
import { LargeEpisodeListItem } from "@/components/shared/large-episode-list-item"
import { GenericSliderEpisodeItem } from "@/components/shared/slider-episode-item"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselItem } from "@/components/ui/carousel"
import { HorizontalDraggableScroll } from "@/components/ui/horizontal-draggable-scroll"
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

                        <Carousel
                            className="w-full max-w-full"
                            gap="md"
                            opts={{
                                align: "start",
                            }}
                            autoScroll
                        >
                            <CarouselDotButtons />
                            <CarouselContent>
                                {!isLoading && missingEpisodes?.map(episode => {
                                    return <CarouselItem
                                        key={episode?.basicMedia?.id + episode.displayTitle}
                                        className="md:basis-1/2 lg:basis-1/3 2xl:basis-1/4 min-[2000px]:basis-1/5"
                                    >
                                        <GenericSliderEpisodeItem
                                            key={episode.displayTitle + episode.basicMedia?.id}
                                            image={episode.episodeMetadata?.image}
                                            topTitle={episode.basicMedia?.title?.userPreferred}
                                            title={episode.displayTitle}
                                            meta={episode.episodeMetadata?.airDate ?? undefined}
                                            actionIcon={<AiOutlineDownload className="opacity-50" />}
                                            isInvalid={episode.isInvalid}
                                            onClick={() => {
                                                router.push(`/entry?id=${episode.basicMedia?.id}&download=${episode.episodeNumber}`)
                                            }}
                                        />
                                    </CarouselItem>
                                })}
                            </CarouselContent>
                        </Carousel>
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
