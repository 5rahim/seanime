"use client"
import React from "react"
import { Skeleton } from "@/components/ui/skeleton"
import { IoLibrary } from "@react-icons/all-files/io5/IoLibrary"
import { LargeEpisodeListItem } from "@/components/shared/large-episode-list-item"
import { AiOutlineDownload } from "@react-icons/all-files/ai/AiOutlineDownload"
import { useRouter } from "next/navigation"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Slider } from "@/components/shared/slider"
import { useMissingEpisodes } from "@/lib/server/hooks/library"

export function MissingEpisodes() {
    const router = useRouter()
    const { missingEpisodes, isLoading } = useMissingEpisodes()

    return (
        <>
            <AppLayoutStack spacing={"lg"}>

                <h2 className={"flex gap-3 items-center"}><IoLibrary/> Missing from your library</h2>

                <Slider>
                    {!isLoading && missingEpisodes.map(episode => {
                        return <LargeEpisodeListItem
                            key={episode.displayTitle + episode.basicMedia?.id}
                            image={episode.episodeMetadata?.image}
                            topTitle={episode.basicMedia?.title?.userPreferred}
                            title={episode.displayTitle}
                            meta={episode.episodeMetadata?.airDate ?? undefined}
                            actionIcon={<AiOutlineDownload/>}
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
                    {!isLoading && !missingEpisodes.length && (
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