"use client"
import { BulkActionModal } from "@/app/(main)/(library)/_containers/bulk-action-modal"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { LibraryToolbar } from "@/app/(main)/(library)/_containers/library-toolbar"
import { UnknownMediaManager } from "@/app/(main)/(library)/_containers/unknown-media-manager"
import { UnmatchedFileManager } from "@/app/(main)/(library)/_containers/unmatched-file-manager"
import { useHandleLibraryCollection } from "@/app/(main)/(library)/_lib/handle-library-collection"
import { __library_viewAtom } from "@/app/(main)/(library)/_lib/library-view.atoms"
import { DetailedLibraryView } from "@/app/(main)/(library)/_screens/detailed-library-view"
import { LibraryView } from "@/app/(main)/(library)/_screens/library-view"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import { AnimatePresence } from "framer-motion"
import { useAtom } from "jotai/react"
import React from "react"

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

    const [view, setView] = useAtom(__library_viewAtom)

    const ts = useThemeSettings()

    const hasScanned = React.useMemo(() => libraryCollectionList?.some(n => !!n.entries?.length), [libraryCollectionList])

    return (
        <div>
            {hasScanned && ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Custom && <CustomLibraryBanner />}
            <LibraryToolbar
                collectionList={libraryCollectionList}
                unmatchedLocalFiles={unmatchedLocalFiles}
                ignoredLocalFiles={ignoredLocalFiles}
                unknownGroups={unknownGroups}
                isLoading={isLoading}
                hasScanned={hasScanned}
            />

            <AnimatePresence mode="wait" initial={false}>
                {view === "base" && <PageWrapper
                    key="base"
                    className="relative 2xl:order-first pb-10 pt-4"
                    {...{
                        initial: { opacity: 0, y: 0 },
                        animate: { opacity: 1, y: 0 },
                        exit: { opacity: 0 },
                        transition: {
                            duration: 0.25,
                        },
                    }}
                >
                    <LibraryView
                        collectionList={libraryCollectionList}
                        continueWatchingList={continueWatchingList}
                        isLoading={isLoading}
                        hasScanned={hasScanned}
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
                        hasScanned={hasScanned}
                    />
                </PageWrapper>}
            </AnimatePresence>

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
