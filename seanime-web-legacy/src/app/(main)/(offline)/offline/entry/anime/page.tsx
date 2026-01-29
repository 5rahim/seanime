"use client"

import { useGetAnimeEntry } from "@/api/hooks/anime_entries.hooks"
import { OfflineMetaSection } from "@/app/(main)/(offline)/offline/entry/_components/offline-meta-section"
import { MediaEntryPageLoadingDisplay } from "@/app/(main)/_features/media/_components/media-entry-page-loading-display"
import { EpisodeSection } from "@/app/(main)/entry/_containers/episode-list/episode-section"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { useRouter, useSearchParams } from "next/navigation"
import React from "react"

export default function Page() {
    const router = useRouter()
    const mediaId = useSearchParams().get("id")

    const { data: animeEntry, isLoading: animeEntryLoading } = useGetAnimeEntry(mediaId)

    React.useEffect(() => {
        if (!mediaId || (!animeEntryLoading && !animeEntry)) {
            router.push("/offline")
        }
    }, [animeEntry, animeEntryLoading])

    if (animeEntryLoading) return <MediaEntryPageLoadingDisplay />
    if (!animeEntry) return null

    return (
        <>
            <OfflineMetaSection type="anime" entry={animeEntry} />
            <PageWrapper
                className="p-4 relative"
                data-media={JSON.stringify(animeEntry.media)}
                data-anime-entry-list-data={JSON.stringify(animeEntry.listData)}
            >
                <EpisodeSection
                    entry={animeEntry}
                    details={undefined}
                    bottomSection={<></>}
                />
            </PageWrapper>
        </>
    )

}
