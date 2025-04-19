"use client"
import { LibraryHeader } from "@/app/(main)/(library)/_components/library-header"
import { BulkActionModal } from "@/app/(main)/(library)/_containers/bulk-action-modal"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { IgnoredFileManager } from "@/app/(main)/(library)/_containers/ignored-file-manager"
import { LibraryToolbar } from "@/app/(main)/(library)/_containers/library-toolbar"
import { UnknownMediaManager } from "@/app/(main)/(library)/_containers/unknown-media-manager"
import { UnmatchedFileManager } from "@/app/(main)/(library)/_containers/unmatched-file-manager"
import { useHandleLibraryCollection } from "@/app/(main)/(library)/_lib/handle-library-collection"
import { __library_viewAtom } from "@/app/(main)/(library)/_lib/library-view.atoms"
import { DetailedLibraryView } from "@/app/(main)/(library)/_screens/detailed-library-view"
import { EmptyLibraryView } from "@/app/(main)/(library)/_screens/empty-library-view"
import { LibraryView } from "@/app/(main)/(library)/_screens/library-view"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import { AnimatePresence } from "framer-motion"
import { useAtom } from "jotai/react"
import React from "react"

export const dynamic = "force-static"

export default function Library() {

    const {
        libraryGenres,
        libraryCollectionList,
        filteredLibraryCollectionList,
        isLoading,
        continueWatchingList,
        unmatchedLocalFiles,
        ignoredLocalFiles,
        unmatchedGroups,
        unknownGroups,
    } = useHandleLibraryCollection()

    const [view, setView] = useAtom(__library_viewAtom)

    const ts = useThemeSettings()

    const hasEntries = React.useMemo(() => libraryCollectionList?.some(n => !!n.entries?.length), [libraryCollectionList])

    return (
        <div data-library-page-container>

            {hasEntries && ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Custom && <CustomLibraryBanner isLibraryScreen />}
            {hasEntries && ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Dynamic && <LibraryHeader list={continueWatchingList} />}
            <LibraryToolbar
                collectionList={libraryCollectionList}
                unmatchedLocalFiles={unmatchedLocalFiles}
                ignoredLocalFiles={ignoredLocalFiles}
                unknownGroups={unknownGroups}
                isLoading={isLoading}
                hasEntries={hasEntries}
            />

            <EmptyLibraryView isLoading={isLoading} hasEntries={hasEntries} />

            <AnimatePresence mode="wait">
                {view === "base" && <PageWrapper
                    key="base"
                    className="relative 2xl:order-first pb-10 pt-4"
                    {...{
                        initial: { opacity: 0, y: 60 },
                        animate: { opacity: 1, y: 0 },
                        exit: { opacity: 0, scale: 0.99 },
                        transition: {
                            duration: 0.25,
                        },
                    }}
                >
                    <LibraryView
                        genres={libraryGenres}
                        collectionList={libraryCollectionList}
                        filteredCollectionList={filteredLibraryCollectionList}
                        continueWatchingList={continueWatchingList}
                        isLoading={isLoading}
                        hasEntries={hasEntries}
                    />
                </PageWrapper>}
                {view === "detailed" && <PageWrapper
                    key="detailed"
                    className="relative 2xl:order-first pb-10 pt-4"
                    {...{
                        initial: { opacity: 0, y: 60 },
                        animate: { opacity: 1, y: 0 },
                        exit: { opacity: 0, scale: 0.99 },
                        transition: {
                            duration: 0.25,
                        },
                    }}
                >
                    <DetailedLibraryView
                        collectionList={libraryCollectionList}
                        continueWatchingList={continueWatchingList}
                        isLoading={isLoading}
                        hasEntries={hasEntries}
                    />
                </PageWrapper>}
            </AnimatePresence>

            <UnmatchedFileManager
                unmatchedGroups={unmatchedGroups}
            />
            <UnknownMediaManager
                unknownGroups={unknownGroups}
            />
            <IgnoredFileManager
                files={ignoredLocalFiles}
            />
            <BulkActionModal />
        </div>
    )
}
