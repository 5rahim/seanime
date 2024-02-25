"use client"

import { EntryHeaderBackground } from "@/app/(main)/entry/_components/entry-header-background"
import { EpisodeSection } from "@/app/(main)/entry/_containers/episode-section/episode-section"
import { MetaSection } from "@/app/(main)/entry/_containers/meta-section/meta-section"
import { TorrentSearchDrawer } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { useMediaDetails, useMediaEntry } from "@/app/(main)/entry/_lib/media-entry"
import { Skeleton } from "@/components/ui/skeleton"
import { useRouter, useSearchParams } from "next/navigation"
import React, { useEffect } from "react"

export default function Page() {
    const router = useRouter()
    const searchParams = useSearchParams()
    const mediaId = searchParams.get("id")
    const { mediaEntry, mediaEntryLoading } = useMediaEntry(mediaId)
    const { mediaDetails, mediaDetailsLoading } = useMediaDetails(mediaId)

    useEffect(() => {
        if (!mediaId) {
            router.push("/")
        } else if ((!mediaEntryLoading && !mediaEntry)) {
            router.push("/")
        }
    }, [mediaEntry, mediaEntryLoading])


    if (mediaEntryLoading || mediaDetailsLoading) return <LoadingDisplay />
    if (!mediaEntry) return null

    return (
        <div>
            <EntryHeaderBackground entry={mediaEntry} />
            <div
                className="-mt-[8rem] relative z-10 max-w-full px-4 md:px-10 grid grid-cols-1 2xl:grid-cols-2 gap-8 pb-16"
            >
                <div className="-mt-[18rem] h-[fit-content] 2xl:sticky top-[5rem] backdrop-blur-xl">
                    {/*<div*/}
                    {/*    className="-mt-[18rem] p-8 rounded-xl backdrop-blur-2xl bg-gray-900 bg-opacity-50 backdrop-opacity-80 drop-shadow-md">*/}
                    <MetaSection entry={mediaEntry} details={mediaDetails} />
                </div>
                <div className="relative 2xl:order-first pb-10">
                    <EpisodeSection entry={mediaEntry} />
                </div>
            </div>
            <TorrentSearchDrawer entry={mediaEntry} />
        </div>
    )
}

function LoadingDisplay() {
    return (
        <div className="__header h-[30rem]">
            <div
                className="h-[30rem] w-full md:w-[calc(100%-5rem)] flex-none object-cover object-center absolute top-0 overflow-hidden"
            >
                <div
                    className="w-full absolute z-[1] top-0 h-[15rem] bg-gradient-to-b from-[--background] to-transparent via"
                />
                <Skeleton className="h-full absolute w-full" />
                <div
                    className="w-full absolute bottom-0 h-[20rem] bg-gradient-to-t from-[--background] via-transparent to-transparent"
                />
            </div>
        </div>
    )
}
