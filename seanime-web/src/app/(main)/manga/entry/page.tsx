"use client"
import { useGetMangaEntry, useGetMangaEntryDetails } from "@/api/hooks/manga.hooks"
import { MediaEntryCharactersSection } from "@/app/(main)/_features/media/_components/media-entry-characters-section"
import { MediaEntryPageLoadingDisplay } from "@/app/(main)/_features/media/_components/media-entry-page-loading-display"
import { MangaRecommendations } from "@/app/(main)/manga/_components/manga-recommendations"
import { MetaSection } from "@/app/(main)/manga/_components/meta-section"
import { ChapterList } from "@/app/(main)/manga/_containers/chapter-list/chapter-list"
import { useHandleMangaDownloadData } from "@/app/(main)/manga/_lib/handle-manga-downloads"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { useRouter, useSearchParams } from "next/navigation"
import React from "react"

export const dynamic = "force-static"

export default function Page() {
    const router = useRouter()
    const searchParams = useSearchParams()
    const mediaId = searchParams.get("id")
    const { data: mangaEntry, isLoading: mangaEntryLoading } = useGetMangaEntry(mediaId)
    const { data: mangaDetails, isLoading: mangaDetailsLoading } = useGetMangaEntryDetails(mediaId)

    /**
     * Fetch manga download data
     */
    const { downloadData, downloadDataLoading } = useHandleMangaDownloadData(mediaId)

    React.useEffect(() => {
        if (!mediaId) {
            router.push("/")
        } else if ((!mangaEntryLoading && !mangaEntry)) {
            router.push("/")
        }
    }, [mangaEntry, mangaEntryLoading])

    React.useEffect(() => {
        try {
            if (mangaEntry?.media?.title?.userPreferred) {
                document.title = `${mangaEntry?.media?.title?.userPreferred} | Seanime`
            }
        }
        catch {
        }
    }, [mangaEntry])

    if (!mangaEntry || mangaEntryLoading || mangaDetailsLoading) return <MediaEntryPageLoadingDisplay />

    return (
        <div
            data-manga-entry-page
            data-media={JSON.stringify(mangaEntry.media)}
            data-manga-entry-list-data={JSON.stringify(mangaEntry.listData)}
        >
            <MetaSection entry={mangaEntry} details={mangaDetails} />

            <div data-manga-entry-page-content-container className="px-4 md:px-8 relative z-[8]">

                <PageWrapper
                    data-manga-entry-page-content
                    key="chapter-list"
                    className="relative 2xl:order-first pb-10 pt-4 space-y-10"
                    {...{
                        initial: { opacity: 0, y: 60 },
                        animate: { opacity: 1, y: 0 },
                        exit: { opacity: 0, y: 60 },
                        transition: {
                            type: "spring",
                            damping: 10,
                            stiffness: 80,
                            delay: 0.6,
                        },
                    }}
                >

                    <div
                        data-manga-entry-page-grid
                        className="grid gap-8 xl:grid-cols-[1fr,480px] 2xl:grid-cols-[1fr,650px]"
                    >
                        <div className="space-y-2">
                            <ChapterList
                                entry={mangaEntry}
                                mediaId={mediaId}
                                details={mangaDetails}
                                downloadData={downloadData}
                                downloadDataLoading={downloadDataLoading}
                            />
                        </div>

                        <div data-manga-entry-page-characters-section-container className="pt-12">
                            <MediaEntryCharactersSection details={mangaDetails} isMangaPage />
                        </div>
                    </div>

                    <MangaRecommendations entry={mangaEntry} details={mangaDetails} />

                </PageWrapper>
            </div>
        </div>
    )
}
