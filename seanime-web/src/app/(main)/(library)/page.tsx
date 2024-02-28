"use client"
import { BulkActionModal } from "@/app/(main)/(library)/_containers/bulk-actions/bulk-action-modal"
import { ContinueWatching } from "@/app/(main)/(library)/_containers/continue-watching"
import { useLibraryCollection } from "@/app/(main)/(library)/_containers/library-collection/_lib/library-collection"
import { LibraryCollectionLists } from "@/app/(main)/(library)/_containers/library-collection/library-collection"
import { LibraryHeader } from "@/app/(main)/(library)/_containers/library-header"
import { LibraryToolbar } from "@/app/(main)/(library)/_containers/library-toolbar"
import { UnknownMediaManager } from "@/app/(main)/(library)/_containers/unknown-media/unknown-media-manager"
import { UnmatchedFileManager } from "@/app/(main)/(library)/_containers/unmatched-files/unmatched-file-manager"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { cn } from "@/components/ui/core/styling"
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
            <LibraryHeader list={continueWatchingList} />
            <PageWrapper
                className={cn(
                    "mt-10",
                )}
            >
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
            </PageWrapper>
        </div>
    )
}
