import { useGetMangaEntry } from "@/api/hooks/manga.hooks"
import { MediaEntryPageLoadingDisplay } from "@/app/(main)/_features/media/_components/media-entry-page-loading-display"
import { OfflineMetaSection } from "@/app/(main)/_features/offline/entry/_components/offline-meta-section"
import { OfflineChapterList } from "@/app/(main)/_features/offline/entry/manga/_components/offline-chapter-list"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { usePathname } from "@/lib/navigation"
import { useRouter, useSearchParams } from "@/lib/navigation"
import React from "react"

export default function Page() {
    const router = useRouter()
    const mediaId = useSearchParams().get("id")
    const pathname = usePathname()

    const { data: mangaEntry, isLoading: mangaEntryLoading } = useGetMangaEntry(mediaId)

    React.useEffect(() => {
        if (!pathname.startsWith("/offline/entry/manga")) return
        if (!mediaId || (!mangaEntryLoading && !mangaEntry)) {
            router.push("/offline")
        }
    }, [mangaEntry, mangaEntryLoading, pathname])

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
