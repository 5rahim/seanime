"use client"
import { LibraryHeader } from "@/app/(main)/(library)/_containers/library-header"
import React from "react"
import { LibraryCollectionLists } from "@/app/(main)/(library)/_containers/library-collection"
import { useLibraryCollection } from "@/lib/server/hooks/library"
import { ContinueWatching } from "@/app/(main)/(library)/_containers/continue-watching"
import { LibraryToolbar } from "@/app/(main)/(library)/_containers/library-toolbar"
import { ScannerModal } from "@/app/(main)/(library)/_components/scanner-modal"
import { ScanProgressBar } from "@/app/(main)/(library)/_components/scan-progress-bar"
import { UnmatchedFileManager } from "@/app/(main)/(library)/_components/unmatched-file-manager"
import { BulkActionModal } from "@/app/(main)/(library)/_components/bulk-action-modal"

export default function Library() {

    const {
        libraryCollectionList,
        isLoading,
        continueWatchingList,
        unmatchedLocalFiles,
        ignoredLocalFiles,
        unmatchedGroups,
    } = useLibraryCollection()

    return (
        <div>
            <ScanProgressBar/>
            <LibraryHeader/>
            <LibraryToolbar
                collectionList={libraryCollectionList}
                unmatchedLocalFiles={unmatchedLocalFiles}
                ignoredLocalFiles={ignoredLocalFiles}
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
            <ScannerModal/>
            <UnmatchedFileManager
                unmatchedGroups={unmatchedGroups}
            />
            <BulkActionModal/>
        </div>
    )
}