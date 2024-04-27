"use client"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { OfflineAnimeLists } from "@/app/(main)/(offline)/offline/_components/offline-anime-lists"
import { OfflineMangaLists } from "@/app/(main)/(offline)/offline/_components/offline-manga-lists"
import { useOfflineSnapshot } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot-context"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { Separator } from "@/components/ui/separator"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import React from "react"

export const dynamic = "force-static"

export default function Page() {
    const status = useServerStatus()
    const ts = useThemeSettings()

    const { snapshot } = useOfflineSnapshot()

    if (!snapshot) return null

    return (
        <>
            {ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Custom && <CustomLibraryBanner />}

            <OfflineAnimeLists />
            {!!snapshot?.entries?.mangaEntries && <div className="space-y-6 p-4 relative z-[5]" id="manga">

                <Separator />

                <h1 className="text-center lg:text-left">Manga</h1>

                <OfflineMangaLists />
            </div>}

        </>
    )
}
