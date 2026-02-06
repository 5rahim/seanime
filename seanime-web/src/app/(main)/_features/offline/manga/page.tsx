import { CustomLibraryBanner } from "@/app/(main)/_features/anime-library/_containers/custom-library-banner"
import { OfflineMangaLists } from "@/app/(main)/_features/offline/_components/offline-manga-lists"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/theme-hooks"
import React from "react"


export default function Page() {
    const ts = useThemeSettings()

    return (
        <>
            {ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Custom && <CustomLibraryBanner isLibraryScreen />}
            <OfflineMangaLists />
        </>
    )
}
