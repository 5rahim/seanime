import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { OfflineAnimeLists } from "@/app/(main)/(offline)/offline/_components/offline-anime-lists"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/theme-hooks.ts"
import React from "react"


export default function Page() {
    const ts = useThemeSettings()

    return (
        <>
            {ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Custom && <CustomLibraryBanner isLibraryScreen />}
            <OfflineAnimeLists />
        </>
    )
}
