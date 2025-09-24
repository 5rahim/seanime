"use client"
import { HomeScreen } from "@/app/(main)/(library)/_home/home-screen"
import { useThemeSettings } from "@/lib/theme/hooks"
import React from "react"

export const dynamic = "force-static"

export default function Page() {

    // const {
    //     libraryGenres,
    //     libraryCollectionList,
    //     filteredLibraryCollectionList,
    //     isLoading,
    //     continueWatchingList,
    //     unmatchedLocalFiles,
    //     ignoredLocalFiles,
    //     unmatchedGroups,
    //     unknownGroups,
    //     streamingMediaIds,
    //     hasEntries,
    //     isStreamingOnly,
    //     isNakamaLibrary,
    // } = useHandleLibraryCollection()
    //
    // const [view, setView] = useAtom(__library_viewAtom)

    const ts = useThemeSettings()

    return <HomeScreen />

    // return (
    //     <div data-library-page-container>
    //
    //         {hasEntries && ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Custom && <CustomLibraryBanner isLibraryScreen />}
    //         {hasEntries && ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Dynamic && <LibraryHeader list={continueWatchingList} />}
    //         <LibraryToolbar
    //             collectionList={libraryCollectionList}
    //             unmatchedLocalFiles={unmatchedLocalFiles}
    //             ignoredLocalFiles={ignoredLocalFiles}
    //             unknownGroups={unknownGroups}
    //             isLoading={isLoading}
    //             hasEntries={hasEntries}
    //             isStreamingOnly={isStreamingOnly}
    //             isNakamaLibrary={isNakamaLibrary}
    //         />
    //
    //         <EmptyLibraryView isLoading={isLoading} hasEntries={hasEntries} />
    //
    //         <AnimatePresence mode="wait">
    //             {view === "base" && <PageWrapper
    //                 key="base"
    //                 className="relative 2xl:order-first pb-10 pt-4"
    //                 {...{
    //                     initial: { opacity: 0, y: 5 },
    //                     animate: { opacity: 1, y: 0 },
    //                     exit: { opacity: 0, scale: 0.99 },
    //                     transition: {
    //                         duration: 0.25,
    //                     },
    //                 }}
    //             >
    //                 <LibraryView
    //                     genres={libraryGenres}
    //                     collectionList={libraryCollectionList}
    //                     filteredCollectionList={filteredLibraryCollectionList}
    //                     continueWatchingList={continueWatchingList}
    //                     isLoading={isLoading}
    //                     hasEntries={hasEntries}
    //                     streamingMediaIds={streamingMediaIds}
    //                 />
    //             </PageWrapper>}
    //             {view === "detailed" && <PageWrapper
    //                 key="detailed"
    //                 className="relative 2xl:order-first pb-10 pt-4"
    //                 {...{
    //                     initial: { opacity: 0, y: 5 },
    //                     animate: { opacity: 1, y: 0 },
    //                     exit: { opacity: 0, scale: 0.99 },
    //                     transition: {
    //                         duration: 0.25,
    //                     },
    //                 }}
    //             >
    //                 <DetailedLibraryView
    //                     collectionList={libraryCollectionList}
    //                     continueWatchingList={continueWatchingList}
    //                     isLoading={isLoading}
    //                     hasEntries={hasEntries}
    //                     streamingMediaIds={streamingMediaIds}
    //                     isNakamaLibrary={isNakamaLibrary}
    //                 />
    //             </PageWrapper>}
    //         </AnimatePresence>
    //
    //         <HomeSettingsModal />
    //
    //         <UnmatchedFileManager
    //             unmatchedGroups={unmatchedGroups}
    //         />
    //         <UnknownMediaManager
    //             unknownGroups={unknownGroups}
    //         />
    //         <IgnoredFileManager
    //             files={ignoredLocalFiles}
    //         />
    //         <BulkActionModal />
    //     </div>
    // )
}
