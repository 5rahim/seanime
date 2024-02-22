import { _scannerModalIsOpen } from "@/app/(main)/(library)/_containers/scanner/scanner-modal"
import { DiscoverPageHeader } from "@/app/(main)/discover/_containers/discover-sections/header"
import { DiscoverTrending } from "@/app/(main)/discover/_containers/discover-sections/trending"
import { AnimeListItem } from "@/components/shared/anime-list-item"
import { Button } from "@/components/ui/button"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { LibraryCollectionEntry, LibraryCollectionList } from "@/lib/server/types"
import { getLibraryCollectionTitle } from "@/lib/server/utils"
import { FiSearch } from "@react-icons/all-files/fi/FiSearch"
import { useSetAtom } from "jotai"
import React, { memo, useMemo } from "react"

export function LibraryCollectionLists({ collectionList, isLoading }: {
    collectionList: LibraryCollectionList[],
    isLoading: boolean
}) {

    const setScannerModalOpen = useSetAtom(_scannerModalIsOpen)

    const hasScanned = useMemo(() => collectionList.some(n => n.entries.length > 0), [collectionList])

    if (isLoading) return <LoadingSpinner />

    if (!hasScanned && !isLoading) return (
        <>
            <DiscoverPageHeader />
            <div className="p-8 pt-0">

                <div className="text-center space-y-4">
                    <div className="border-2 border-[--border] border-dashed rounded-xl py-6 px-10 w-fit mx-auto space-y-4">
                        <h2>Empty library</h2>
                        <Button
                            intent={"warning-subtle"}
                            leftIcon={<FiSearch />}
                            size={"xl"}
                            rounded
                            onClick={() => setScannerModalOpen(true)}
                        >
                            Scan your library
                        </Button>
                    </div>
                </div>
                <h3>Popular this season</h3>
                <DiscoverTrending />
            </div>
        </>
    )

    return (
        <div className="p-4 space-y-8 relative">
            {collectionList.map(collection => {
                if (collection.entries.length === 0) return null
                return <LibraryCollectionListItem key={collection.type} list={collection} />
            })}
        </div>
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
        />
    )
})
