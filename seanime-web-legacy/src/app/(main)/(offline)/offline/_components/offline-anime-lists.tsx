import { LibraryHeader } from "@/app/(main)/(library)/_components/library-header"
import { ContinueWatching } from "@/app/(main)/(library)/_containers/continue-watching"
import { useHandleLibraryCollection } from "@/app/(main)/(library)/_lib/handle-library-collection"
import { LibraryView } from "@/app/(main)/(library)/_screens/library-view"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { cn } from "@/components/ui/core/styling"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import React from "react"

export function OfflineAnimeLists() {
    const ts = useThemeSettings()

    const {
        libraryGenres,
        libraryCollectionList,
        filteredLibraryCollectionList,
        isLoading,
        continueWatchingList,
        streamingMediaIds,
    } = useHandleLibraryCollection()

    return (
        <>
            {ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Dynamic && <>
                <LibraryHeader list={continueWatchingList} />
                <div
                    className={cn(
                        "h-40",
                        ts.hideTopNavbar && "h-40",
                    )}
                ></div>
            </>}
            <PageWrapper
                className="pt-4 relative space-y-8"
            >
                <ContinueWatching
                    episodes={continueWatchingList}
                    isLoading={isLoading}
                    withTitle
                />
                <LibraryView
                    genres={libraryGenres}
                    collectionList={libraryCollectionList}
                    filteredCollectionList={filteredLibraryCollectionList}
                    continueWatchingList={continueWatchingList}
                    isLoading={isLoading}
                    hasEntries={true}
                    streamingMediaIds={streamingMediaIds}
                />
            </PageWrapper>
        </>
    )
}
