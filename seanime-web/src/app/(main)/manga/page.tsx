"use client"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { MediaEntryPageLoadingDisplay } from "@/app/(main)/_features/media/_components/media-entry-page-loading-display"
import { LibraryHeader } from "@/app/(main)/manga/_components/library-header"
import { useHandleMangaCollection } from "@/app/(main)/manga/_lib/handle-manga-collection"
import { MangaLibraryView } from "@/app/(main)/manga/_screens/manga-library-view"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import React from "react"

export const dynamic = "force-static"

export default function Page() {
    const {
        mangaCollection,
        filteredMangaCollection,
        genres,
        mangaCollectionLoading,
    } = useHandleMangaCollection()

    const ts = useThemeSettings()

    if (!mangaCollection || mangaCollectionLoading) return <MediaEntryPageLoadingDisplay />

    return (
        <div>
            {ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Custom && (
                <>
                    <CustomLibraryBanner />
                    <div className="h-32"></div>
                </>
            )}
            {ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Dynamic && (
                <>
                    <LibraryHeader manga={mangaCollection?.lists?.flatMap(l => l.entries)?.flatMap(e => e?.media)?.filter(Boolean) || []} />
                    <div className="h-10"></div>
                </>
            )}


            <MangaLibraryView
                genres={genres}
                collection={mangaCollection}
                filteredCollection={filteredMangaCollection}
            />
        </div>
    )
}
