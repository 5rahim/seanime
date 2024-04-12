"use client"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { OfflineAnimeLists } from "@/app/(main)/(offline)/offline/_components/offline-anime-lists"
import { useOfflineSnapshot } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot-context"
import { serverStatusAtom } from "@/atoms/server-status"
import { OfflineMediaListAtom } from "@/components/shared/custom-ui/offline-media-list-item"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import { useAtomValue } from "jotai"
import React from "react"

export default function Page() {
    const status = useAtomValue(serverStatusAtom)
    const ts = useThemeSettings()

    const { snapshot } = useOfflineSnapshot()

    if (!snapshot) return null

    return (
        <>
            {ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Custom && <CustomLibraryBanner />}

                <OfflineAnimeLists />
                <div className="space-y-6">
                    <h2 className="text-center lg:text-left">Manga</h2>


                    <div
                        className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-7 min-[2000px]:grid-cols-8 gap-4"
                    >
                        {snapshot?.entries?.mangaEntries?.map(entry => {
                            if (!entry) return null

                            return <OfflineMediaListAtom
                                key={entry.mediaId}
                                media={entry.media!}
                                withAudienceScore={false}
                                isManga
                                assetMap={snapshot.assetMap}
                            />
                        })}
                    </div>
                </div>

        </>
    )
}
