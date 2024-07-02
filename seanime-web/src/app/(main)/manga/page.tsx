"use client"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { LibraryHeader } from "@/app/(main)/manga/_components/library-header"
import { useHandleMangaCollection } from "@/app/(main)/manga/_lib/handle-manga-collection"
import { MangaLibraryView } from "@/app/(main)/manga/_screens/manga-library-view"
import { Skeleton } from "@/components/ui/skeleton"
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

    if (!mangaCollection || mangaCollectionLoading) return <LoadingDisplay />

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

function LoadingDisplay() {
    return (
        <div className="__header h-[30rem]">
            <div
                className="h-[30rem] w-full flex-none object-cover object-center absolute top-0 overflow-hidden"
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
