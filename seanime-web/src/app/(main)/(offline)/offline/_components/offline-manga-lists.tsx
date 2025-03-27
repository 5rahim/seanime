import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { MediaEntryPageLoadingDisplay } from "@/app/(main)/_features/media/_components/media-entry-page-loading-display"
import { LibraryHeader } from "@/app/(main)/manga/_components/library-header"
import { useHandleMangaCollection } from "@/app/(main)/manga/_lib/handle-manga-collection"
import { MangaLibraryView } from "@/app/(main)/manga/_screens/manga-library-view"
import { cn } from "@/components/ui/core/styling"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import React from "react"

export function OfflineMangaLists() {
    const {
        mangaCollection,
        filteredMangaCollection,
        genres,
        mangaCollectionLoading,
        storedProviders,
    } = useHandleMangaCollection()

    const ts = useThemeSettings()

    if (!mangaCollection || mangaCollectionLoading) return <MediaEntryPageLoadingDisplay />

    return (
        <div>
            {(
                (!!ts.libraryScreenCustomBannerImage && ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Custom)
            ) && (
                <>
                    <CustomLibraryBanner isLibraryScreen />
                    <div
                        className={cn("h-14")}
                    ></div>
                </>
            )}
            {ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Dynamic && (
                <>
                    <LibraryHeader manga={mangaCollection?.lists?.flatMap(l => l.entries)?.flatMap(e => e?.media)?.filter(Boolean) || []} />
                    <div
                        className={cn(
                            "h-28",
                            ts.hideTopNavbar && "h-40",
                        )}
                    ></div>
                </>
            )}

            <MangaLibraryView
                genres={genres}
                collection={mangaCollection}
                filteredCollection={filteredMangaCollection}
                storedProviders={storedProviders}
            />
        </div>
    )
}
