"use client"
import { LargeEpisodeListItem } from "@/components/shared/large-episode-list-item"
import { Slider } from "@/components/shared/slider"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Skeleton } from "@/components/ui/skeleton"
import { MediaEntryEpisode } from "@/lib/server/types"
import { AiOutlineDownload } from "@react-icons/all-files/ai/AiOutlineDownload"
import { IoLibrary } from "@react-icons/all-files/io5/IoLibrary"
import { useRouter } from "next/navigation"
import React from "react"

export function MissingEpisodes({ isLoading, missingEpisodes }: { missingEpisodes: MediaEntryEpisode[] | undefined, isLoading: boolean }) {
    const router = useRouter()

    if (!missingEpisodes?.length) return null

    return (
        <>
            <AppLayoutStack spacing={"lg"}>

                <h2 className={"flex gap-3 items-center"}><IoLibrary/> Missing from your library</h2>

                <Slider>
                    {!isLoading && missingEpisodes?.map(episode => {
                        return <LargeEpisodeListItem
                            key={episode.displayTitle + episode.basicMedia?.id}
                            image={episode.episodeMetadata?.image}
                            topTitle={episode.basicMedia?.title?.userPreferred}
                            title={episode.displayTitle}
                            meta={episode.episodeMetadata?.airDate ?? undefined}
                            actionIcon={<AiOutlineDownload/>}
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
            </AppLayoutStack>
        </>
    )
}
