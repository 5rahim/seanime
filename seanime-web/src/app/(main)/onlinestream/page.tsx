"use client"

import { useGetAnimeEntry } from "@/api/hooks/anime_entries.hooks"
import { MediaEntryPageSmallBanner } from "@/app/(main)/_features/media/_components/media-entry-page-small-banner"
import { OnlinestreamPage } from "@/app/(main)/onlinestream/_containers/onlinestream-page"
import { Skeleton } from "@/components/ui/skeleton"
import { useSearchParams } from "next/navigation"
import React from "react"


export const dynamic = "force-static"

export default function Page() {
    const searchParams = useSearchParams()
    const mediaId = searchParams.get("id")
    const { data: animeEntry, isLoading: animeEntryLoading } = useGetAnimeEntry(mediaId)

    if (!animeEntry || animeEntryLoading) return <div data-onlinestream-page-loading-container className="px-4 lg:px-8 space-y-4">
        <div className="flex gap-4 items-center relative">
            <Skeleton className="h-12" />
        </div>
        <div
            className="grid 2xl:grid-cols-[1fr,450px] gap-4 xl:gap-4"
        >
            <div className="w-full min-h-[70dvh] relative">
                <Skeleton className="h-full w-full absolute" />
            </div>

            <Skeleton className="hidden 2xl:block relative h-[78dvh] overflow-y-auto pr-4 pt-0" />

        </div>
    </div>

    return (
        <>
            <div data-onlinestream-page-container className="relative p-4 lg:p-8 z-[5] space-y-4">
                <OnlinestreamPage animeEntry={animeEntry} animeEntryLoading={animeEntryLoading} />
            </div>
            <MediaEntryPageSmallBanner bannerImage={animeEntry?.media?.bannerImage || animeEntry?.media?.coverImage?.extraLarge} />
        </>
    )

}
