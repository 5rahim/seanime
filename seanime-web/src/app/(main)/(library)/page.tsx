"use client"
import { BulkActionModal } from "@/app/(main)/(library)/_components/bulk-action-modal"
import { ScanProgressBar } from "@/app/(main)/(library)/_components/scan-progress-bar"
import { ScannerModal } from "@/app/(main)/(library)/_components/scanner-modal"
import { UnknownMediaManager } from "@/app/(main)/(library)/_components/unknown-media-manager"
import { UnmatchedFileManager } from "@/app/(main)/(library)/_components/unmatched-file-manager"
import { ContinueWatching } from "@/app/(main)/(library)/_containers/continue-watching"
import { LibraryCollectionLists } from "@/app/(main)/(library)/_containers/library-collection"
import { LibraryHeader } from "@/app/(main)/(library)/_containers/library-header"
import { LibraryToolbar } from "@/app/(main)/(library)/_containers/library-toolbar"
import { LibraryWatcher } from "@/components/application/library-watcher"
import { useLibraryCollection } from "@/lib/server/hooks/library"
import React from "react"

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

    return (
        <div>
            <ScanProgressBar />
            <LibraryWatcher />
            <LibraryHeader />
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
            <ScannerModal />
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
