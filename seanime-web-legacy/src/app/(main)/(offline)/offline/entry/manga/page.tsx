"use client"

import { useGetMangaEntry } from "@/api/hooks/manga.hooks"
import { OfflineMetaSection } from "@/app/(main)/(offline)/offline/entry/_components/offline-meta-section"
import { OfflineChapterList } from "@/app/(main)/(offline)/offline/entry/manga/_components/offline-chapter-list"
import { MediaEntryPageLoadingDisplay } from "@/app/(main)/_features/media/_components/media-entry-page-loading-display"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { useRouter, useSearchParams } from "next/navigation"
import React from "react"

export default function Page() {
    const router = useRouter()
    const mediaId = useSearchParams().get("id")

    const { data: mangaEntry, isLoading: mangaEntryLoading } = useGetMangaEntry(mediaId)

    React.useEffect(() => {
        if (!mediaId || (!mangaEntryLoading && !mangaEntry)) {
            router.push("/offline")
        }
    }, [mangaEntry, mangaEntryLoading])

    if (mangaEntryLoading) return <MediaEntryPageLoadingDisplay />
    if (!mangaEntry) return null

    return (
        <>
            <OfflineMetaSection type="manga" entry={mangaEntry} />
            <PageWrapper className="p-4 space-y-6">

                <h2>Chapters</h2>

                <OfflineChapterList entry={mangaEntry} />
            </PageWrapper>
        </>
    )

}
