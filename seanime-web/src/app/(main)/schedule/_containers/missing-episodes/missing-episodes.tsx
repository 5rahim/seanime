"use client"
import { LargeEpisodeListItem } from "@/components/shared/large-episode-list-item"
import { Slider } from "@/components/shared/slider"
import { Accordion } from "@/components/ui/accordion"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Skeleton } from "@/components/ui/skeleton"
import { MediaEntryEpisode } from "@/lib/server/types"
import { AiOutlineDownload } from "@react-icons/all-files/ai/AiOutlineDownload"
import { IoLibrary } from "@react-icons/all-files/io5/IoLibrary"
import { useRouter } from "next/navigation"
import React from "react"
import { LuBellOff } from "react-icons/lu"

export function MissingEpisodes({ isLoading, missingEpisodes, silencedEpisodes }: {
    missingEpisodes: MediaEntryEpisode[] | undefined,
    silencedEpisodes: MediaEntryEpisode[],
    isLoading: boolean
}) {
    const router = useRouter()

    return (
        <>
            <AppLayoutStack spacing={"lg"}>

                {!!missingEpisodes?.length && (
                    <>
                        <h2 className={"flex gap-3 items-center"}><IoLibrary /> Missing from your library</h2>

                        <Slider>
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
                                    className={"rounded-md h-auto overflow-hidden aspect-[4/2] w-96 relative flex items-end flex-none"}
                                />
                                <Skeleton
                                    className={"rounded-md h-auto overflow-hidden aspect-[4/2] w-96 relative flex items-end flex-none"}
                                />
                                <Skeleton
                                    className={"rounded-md h-auto overflow-hidden aspect-[4/2] w-96 relative flex items-end flex-none"}
                                />
                            </>}
                            {!isLoading && !missingEpisodes?.length && (
                                <div
                                    className={"rounded-md h-auto overflow-hidden aspect-[4/2] w-96 relative flex items-center justify-center flex-none bg-gray-900 text-[--muted]"}
                                >
                                    No missing episodes
                                </div>
                            )}
                        </Slider>
                    </>
                )}

                {!!silencedEpisodes.length && (
                    <>

                        <Accordion
                            containerClassName={""}
                            triggerClassName={"py-2 dark:bg-[--background] px-0 dark:hover:bg-transparent text-lg dark:text-[--muted] dark:hover:text-white"}
                        >
                            <Accordion.Item
                                title={<p className={"flex gap-3 items-center text-lg text-inherit"}><LuBellOff /> Silenced episodes</p>}
                                defaultOpen={false}
                            >
                                <Slider>
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
                                </Slider>
                            </Accordion.Item>
                        </Accordion>
                    </>
                )}

            </AppLayoutStack>
        </>
    )
}
