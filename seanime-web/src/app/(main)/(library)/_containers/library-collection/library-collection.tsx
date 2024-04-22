import { _scannerModalIsOpen } from "@/app/(main)/(library)/_containers/scanner/scanner-modal"
import { LibraryCollectionEntry, LibraryCollectionList } from "@/app/(main)/(library)/_lib/anime-library.types"
import { DiscoverPageHeader } from "@/app/(main)/discover/_containers/discover-sections/header"
import { DiscoverTrending } from "@/app/(main)/discover/_containers/discover-sections/trending"
import { AnimeListItem } from "@/components/shared/anime-list-item"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Skeleton } from "@/components/ui/skeleton"
import { getLibraryCollectionTitle } from "@/lib/server/utils"
import { useSetAtom } from "jotai"
import React, { memo, useMemo } from "react"
import { FiSearch } from "react-icons/fi"

export function LibraryCollectionLists({ collectionList, isLoading }: {
    collectionList: LibraryCollectionList[],
    isLoading: boolean
}) {

    const setScannerModalOpen = useSetAtom(_scannerModalIsOpen)

    const hasScanned = useMemo(() => collectionList.some(n => n.entries.length > 0), [collectionList])

    if (isLoading) return <React.Fragment>
        <div className="p-4 space-y-4 relative z-[4]">
            <Skeleton className="h-12 w-full max-w-lg relative" />
            <div
                className={cn(
                    "grid h-[22rem] min-[2000px]:h-[24rem] grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-7 min-[2000px]:grid-cols-8 gap-4",
                    // "md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-7 min-[2000px]:grid-cols-8"
                )}
            >
                {[1, 2, 3, 4, 5, 6, 7, 8]?.map((_, idx) => {
                    return <Skeleton
                        key={idx} className={cn(
                        "h-[22rem] min-[2000px]:h-[24rem] col-span-1 aspect-[6/7] flex-none rounded-md relative overflow-hidden",
                        "[&:nth-child(8)]:hidden min-[2000px]:[&:nth-child(8)]:block",
                        "[&:nth-child(7)]:hidden 2xl:[&:nth-child(7)]:block",
                        "[&:nth-child(6)]:hidden xl:[&:nth-child(6)]:block",
                        "[&:nth-child(5)]:hidden xl:[&:nth-child(5)]:block",
                        "[&:nth-child(4)]:hidden lg:[&:nth-child(4)]:block",
                        "[&:nth-child(3)]:hidden md:[&:nth-child(3)]:block",
                    )}
                    />
                })}
            </div>
        </div>
    </React.Fragment>

    if (!hasScanned && !isLoading) return (
        <>
            <DiscoverPageHeader />
            <PageWrapper className="p-4 sm:p-8 pt-0 space-y-8 relative z-[4]">
                <div className="text-center space-y-4">
                    <div className="w-fit mx-auto space-y-4">
                        <h2>Empty library</h2>
                        <Button
                            intent="warning-subtle"
                            leftIcon={<FiSearch />}
                            size="xl"
                            rounded
                            onClick={() => setScannerModalOpen(true)}
                        >
                            Scan your library
                        </Button>
                    </div>
                </div>
                <div>
                    <h3>Trending this season</h3>
                    <DiscoverTrending />
                </div>
            </PageWrapper>
        </>
    )

    return (
        <PageWrapper className="p-4 space-y-8 relative z-[4]">
            {process.env.NEXT_PUBLIC_PLATFORM}
            {collectionList.map(collection => {
                if (collection.entries.length === 0) return null
                return <LibraryCollectionListItem key={collection.type} list={collection} />
            })}
        </PageWrapper>
    )

}

export const LibraryCollectionListItem = memo(({ list }: { list: LibraryCollectionList }) => {
    return (
        <React.Fragment key={list.type}>
            <h2>{getLibraryCollectionTitle(list.type)}</h2>
            <div
                className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-7 min-[2000px]:grid-cols-8 gap-4"
            >
                {list.entries?.map(entry => {
                    return <LibraryCollectionEntryItem key={entry.mediaId} entry={entry} />
                })}
            </div>
        </React.Fragment>
    )
})
export const LibraryCollectionEntryItem = memo(({ entry }: { entry: LibraryCollectionEntry }) => {
    return (
        <AnimeListItem
            media={entry.media!}
            listData={entry.listData}
            libraryData={entry.libraryData}
            showListDataButton
            withAudienceScore={false}
        />
    )
})
