"use client"
import { BulkActionModal } from "@/app/(main)/(library)/_containers/bulk-actions/bulk-action-modal"
import { ContinueWatching } from "@/app/(main)/(library)/_containers/continue-watching"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { useLibraryCollection } from "@/app/(main)/(library)/_containers/library-collection/_lib/library-collection"
import { LibraryCollectionLists } from "@/app/(main)/(library)/_containers/library-collection/library-collection"
import { LibraryHeader } from "@/app/(main)/(library)/_containers/library-header"
import { LibraryToolbar } from "@/app/(main)/(library)/_containers/library-toolbar"
import { UnknownMediaManager } from "@/app/(main)/(library)/_containers/unknown-media/unknown-media-manager"
import { UnmatchedFileManager } from "@/app/(main)/(library)/_containers/unmatched-files/unmatched-file-manager"
import { CustomBackgroundImage } from "@/components/shared/custom-ui/custom-background-image"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import React, { useMemo } from "react"

export default function Library() {

    const {
        libraryCollectionList,
        isLoading,
        continueWatchingList,
        unmatchedLocalFiles,
        ignoredLocalFiles,
        unmatchedGroups,
        unknownGroups,
    } = useLibraryCollection()

    const ts = useThemeSettings()

    const hasScanned = useMemo(() => libraryCollectionList?.some(n => n.entries.length > 0), [libraryCollectionList])

    return (
        <div>
            {hasScanned && <>
                {ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Dynamic && <LibraryHeader list={continueWatchingList} />}
                {ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Custom && <CustomLibraryBanner />}
                {/*[CUSTOM UI]*/}
                <CustomBackgroundImage />
            </>}
            <LibraryToolbar
                collectionList={libraryCollectionList}
                unmatchedLocalFiles={unmatchedLocalFiles}
                ignoredLocalFiles={ignoredLocalFiles}
                unknownGroups={unknownGroups}
                isLoading={isLoading}
            />
            <ContinueWatching
                list={continueWatchingList}
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
