"use client"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { MediaEntryPageLoadingDisplay } from "@/app/(main)/_features/media/_components/media-entry-page-loading-display"
import { LibraryHeader } from "@/app/(main)/manga/_components/library-header"
import { useHandleMangaCollection } from "@/app/(main)/manga/_lib/handle-manga-collection"
import { MangaLibraryView } from "@/app/(main)/manga/_screens/manga-library-view"
import { cn } from "@/components/ui/core/styling"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import React from "react"

export const dynamic = "force-static"

export default function Page() {
    const {
        mangaCollection,
        filteredMangaCollection,
        mangaCollectionLoading,
        storedFilters,
        storedProviders,
        mangaCollectionGenres,
    } = useHandleMangaCollection()

    const ts = useThemeSettings()

    if (!mangaCollection || mangaCollectionLoading) return <MediaEntryPageLoadingDisplay />

    return (
        <div
            data-manga-page-container
            data-stored-filters={JSON.stringify(storedFilters)}
            data-stored-providers={JSON.stringify(storedProviders)}
        >
            {(
                (!!ts.libraryScreenCustomBannerImage && ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Custom)
            ) && (
                <>
                    <CustomLibraryBanner isLibraryScreen />
                    <div
                        data-manga-page-custom-banner-spacer
                        className={cn("h-14")}
                    ></div>
                </>
            )}
            {ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Dynamic && (
                <>
                    <LibraryHeader manga={mangaCollection?.lists?.flatMap(l => l.entries)?.flatMap(e => e?.media)?.filter(Boolean) || []} />
                    <div
                        data-manga-page-dynamic-banner-spacer
                        className={cn(
                            process.env.NEXT_PUBLIC_PLATFORM !== "desktop" && "h-28",
                            (process.env.NEXT_PUBLIC_PLATFORM !== "desktop" && ts.hideTopNavbar) && "h-40",
                            process.env.NEXT_PUBLIC_PLATFORM === "desktop" && "h-40",
                        )}
                    ></div>
                </>
            )}

            <MangaLibraryView
                genres={mangaCollectionGenres}
                collection={mangaCollection}
                filteredCollection={filteredMangaCollection}
                storedProviders={storedProviders}
            />
        </div>
    )
}
