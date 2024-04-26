"use client"
import { LibraryHeader } from "@/app/(main)/(library)/_components/library-header"
import { BulkActionModal } from "@/app/(main)/(library)/_containers/bulk-action-modal"
import { ContinueWatching } from "@/app/(main)/(library)/_containers/continue-watching"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { LibraryCollectionLists } from "@/app/(main)/(library)/_containers/library-collection"
import { LibraryToolbar } from "@/app/(main)/(library)/_containers/library-toolbar"
import { UnknownMediaManager } from "@/app/(main)/(library)/_containers/unknown-media-manager"
import { UnmatchedFileManager } from "@/app/(main)/(library)/_containers/unmatched-file-manager"
import { useHandleLibraryCollection } from "@/app/(main)/(library)/_hooks/library-collection"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import React, { useMemo } from "react"

export const dynamic = "force-static"

export default function Library() {

    const {
        libraryCollectionList,
        isLoading,
        continueWatchingList,
        unmatchedLocalFiles,
        ignoredLocalFiles,
        unmatchedGroups,
        unknownGroups,
    } = useHandleLibraryCollection()

    const ts = useThemeSettings()

    const hasScanned = useMemo(() => libraryCollectionList?.some(n => !!n.entries?.length), [libraryCollectionList])

    return (
        <div>
            {hasScanned && <>
                {ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Dynamic && <LibraryHeader list={continueWatchingList} />}
                {ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Custom && <CustomLibraryBanner />}
            </>}
            <LibraryToolbar
                collectionList={libraryCollectionList}
                unmatchedLocalFiles={unmatchedLocalFiles}
                ignoredLocalFiles={ignoredLocalFiles}
                unknownGroups={unknownGroups}
                isLoading={isLoading}
            />
            <ContinueWatching
                episodes={continueWatchingList}
                isLoading={isLoading}
            />
            <LibraryCollectionLists
                collectionList={libraryCollectionList}
                isLoading={isLoading}
            />
            <UnmatchedFileManager
                unmatchedGroups={unmatchedGroups}
            />
            <UnknownMediaManager
                unknownGroups={unknownGroups}
            />
            <BulkActionModal />
        </div>
    )
}
