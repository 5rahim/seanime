"use client"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { OfflineMangaLists } from "@/app/(main)/(offline)/offline/_components/offline-manga-lists"
import { useOfflineSnapshot } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot-context"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import React from "react"

export const dynamic = "force-static"

export default function Page() {
    const ts = useThemeSettings()

    const { snapshot } = useOfflineSnapshot()

    if (!snapshot) return null

    return (
        <>
            {ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Custom && <CustomLibraryBanner isLibraryScreen />}
            <OfflineMangaLists />
        </>
    )
}
