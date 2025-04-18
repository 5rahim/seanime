"use client"
import { Anime_MissingEpisodes } from "@/api/generated/types"
import { EpisodeCard } from "@/app/(main)/_features/anime/_components/episode-card"
import { useHasTorrentProvider } from "@/app/(main)/_hooks/use-server-status"
import { useHandleMissingEpisodes } from "@/app/(main)/schedule/_lib/handle-missing-episodes"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselItem } from "@/components/ui/carousel"
import { useRouter } from "next/navigation"
import React from "react"
import { AiOutlineDownload } from "react-icons/ai"
import { IoLibrary } from "react-icons/io5"
import { LuBellOff } from "react-icons/lu"

export function MissingEpisodes({ isLoading, data }: {
    data: Anime_MissingEpisodes | undefined
    isLoading: boolean
}) {
    const router = useRouter()

    const { missingEpisodes, silencedEpisodes } = useHandleMissingEpisodes(data)
    const { hasTorrentProvider } = useHasTorrentProvider()

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
                                        key={episode?.baseAnime?.id + episode.displayTitle}
                                        className="md:basis-1/2 lg:basis-1/3 2xl:basis-1/4 min-[2000px]:basis-1/5"
                                    >
                                        <EpisodeCard
                                            key={episode.displayTitle + episode.baseAnime?.id}
                                            episode={episode}
                                            image={episode.episodeMetadata?.image || episode.baseAnime?.bannerImage || episode.baseAnime?.coverImage?.extraLarge}
                                            topTitle={episode.baseAnime?.title?.userPreferred}
                                            title={episode.displayTitle}
                                            meta={episode.episodeMetadata?.airDate ?? undefined}
                                            actionIcon={hasTorrentProvider ? <AiOutlineDownload className="opacity-50" /> : null}
                                            isInvalid={episode.isInvalid}
                                            onClick={() => {
                                                if (hasTorrentProvider) {
                                                    router.push(`/entry?id=${episode.baseAnime?.id}&download=${episode.episodeNumber}`)
                                                } else {
                                                    router.push(`/entry?id=${episode.baseAnime?.id}`)
                                                }
                                            }}
                                            anime={{
                                                id: episode.baseAnime?.id,
                                                image: episode.baseAnime?.coverImage?.medium,
                                                title: episode.baseAnime?.title?.userPreferred,
                                            }}
                                        />
                                    </CarouselItem>
                                })}
                            </CarouselContent>
                        </Carousel>
                    </>
                )}

                {!!silencedEpisodes?.length && (
                    <>

                        <Accordion
                            type="multiple"
                            defaultValue={[]}
                            triggerClass="py-2 px-0 dark:hover:bg-transparent text-lg dark:text-[--muted] dark:hover:text-white"
                        >
                            <AccordionItem value="item-1">
                                <AccordionTrigger>
                                    <p className="flex gap-3 items-center text-lg text-inherit"><LuBellOff /> Silenced episodes</p>
                                </AccordionTrigger>
                                <AccordionContent className="bg-gray-950 rounded-[--radius]">
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
                                            {!isLoading && silencedEpisodes?.map(episode => {
                                                return (
                                                    <CarouselItem
                                                        key={episode.baseAnime?.id + episode.displayTitle}
                                                        className="md:basis-1/2 lg:basis-1/3 2xl:basis-1/5 min-[2000px]:basis-1/6"
                                                    >
                                                        <EpisodeCard
                                                            key={episode.displayTitle + episode.baseAnime?.id}
                                                            episode={episode}
                                                            image={episode.episodeMetadata?.image || episode.baseAnime?.bannerImage || episode.baseAnime?.coverImage?.extraLarge}
                                                            topTitle={episode.baseAnime?.title?.userPreferred}
                                                            title={episode.displayTitle}
                                                            meta={episode.episodeMetadata?.airDate ?? undefined}
                                                            actionIcon={hasTorrentProvider ? <AiOutlineDownload /> : null}
                                                            isInvalid={episode.isInvalid}
                                                            type="carousel"
                                                            onClick={() => {
                                                                if (hasTorrentProvider) {
                                                                    router.push(`/entry?id=${episode.baseAnime?.id}&download=${episode.episodeNumber}`)
                                                                } else {
                                                                    router.push(`/entry?id=${episode.baseAnime?.id}`)
                                                                }
                                                            }}
                                                            anime={{
                                                                id: episode.baseAnime?.id,
                                                                image: episode.baseAnime?.coverImage?.medium,
                                                                title: episode.baseAnime?.title?.userPreferred,
                                                            }}
                                                        />
                                                    </CarouselItem>
                                                )
                                            })}
                                        </CarouselContent>
                                    </Carousel>
                                </AccordionContent>
                            </AccordionItem>
                        </Accordion>
                    </>
                )}

            </AppLayoutStack>
        </>
    )
}
