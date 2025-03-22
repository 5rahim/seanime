import { Anime_LibraryCollectionEntry, Anime_LibraryCollectionList } from "@/api/generated/types"
import { __mainLibrary_paramsAtom } from "@/app/(main)/(library)/_lib/handle-library-collection"
import { MediaCardLazyGrid } from "@/app/(main)/_features/media/_components/media-card-grid"
import { MediaEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { IconButton } from "@/components/ui/button"
import { DropdownMenu, DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { getLibraryCollectionTitle } from "@/lib/server/utils"
import { useAtom } from "jotai/react"
import React from "react"
import { LuListFilter } from "react-icons/lu"

export function LibraryCollectionLists({ collectionList, isLoading }: {
    collectionList: Anime_LibraryCollectionList[],
    isLoading: boolean
}) {

    return (
        <PageWrapper
            key="library-collection-lists"
            className="space-y-8"
            data-library-collection-lists
            {...{
                initial: { opacity: 0, y: 60 },
                animate: { opacity: 1, y: 0 },
                exit: { opacity: 0, scale: 0.99 },
                transition: {
                    duration: 0.25,
                },
            }}>
            {collectionList.map(collection => {
                if (!collection.entries?.length) return null
                return <LibraryCollectionListItem key={collection.type} list={collection} />
            })}
        </PageWrapper>
    )

}

export function LibraryCollectionFilteredLists({ collectionList, isLoading }: {
    collectionList: Anime_LibraryCollectionList[],
    isLoading: boolean
}) {

    // const params = useAtomValue(__mainLibrary_paramsAtom)

    return (
        <PageWrapper
            key="library-filtered-lists"
            className="space-y-8"
            data-library-filtered-lists
            {...{
                initial: { opacity: 0, y: 60 },
                animate: { opacity: 1, y: 0 },
                exit: { opacity: 0, scale: 0.99 },
                transition: {
                    duration: 0.25,
                },
            }}>
            {/*<h3 className="text-center truncate">*/}
            {/*    {params.genre?.join(", ")}*/}
            {/*</h3>*/}
            <MediaCardLazyGrid itemCount={collectionList?.flatMap(n => n.entries)?.length ?? 0}>
                {collectionList?.flatMap(n => n.entries)?.filter(Boolean)?.map(entry => {
                    return <LibraryCollectionEntryItem key={entry.mediaId} entry={entry} />
                })}
            </MediaCardLazyGrid>
        </PageWrapper>
    )

}

export const LibraryCollectionListItem = React.memo(({ list }: { list: Anime_LibraryCollectionList }) => {

    const isCurrentlyWatching = list.type === "CURRENT"

    const [params, setParams] = useAtom(__mainLibrary_paramsAtom)

    return (
        <React.Fragment key={list.type}>
            <div className="flex gap-3 items-center" data-library-collection-list-item-header data-list-type={list.type}>
                <h2 className="p-0 m-0">{getLibraryCollectionTitle(list.type)}</h2>
                <div className="flex flex-1"></div>
                {isCurrentlyWatching && <DropdownMenu
                    trigger={<IconButton
                        intent="white-basic"
                        size="xs"
                        className="mt-1"
                        icon={<LuListFilter />}
                    />}
                >
                    <DropdownMenuItem
                        onClick={() => {
                            setParams(draft => {
                                draft.continueWatchingOnly = !draft.continueWatchingOnly
                                return
                            })
                        }}
                    >
                        {params.continueWatchingOnly ? "Show all" : "Show unwatched only"}
                    </DropdownMenuItem>
                </DropdownMenu>}
            </div>
            <MediaCardLazyGrid
                itemCount={list?.entries?.length || 0}
                data-library-collection-list-item-media-card-lazy-grid
                data-list-type={list.type}
            >
                {list.entries?.map(entry => {
                    return <LibraryCollectionEntryItem key={entry.mediaId} entry={entry} />
                })}
            </MediaCardLazyGrid>
        </React.Fragment>
    )
})

export const LibraryCollectionEntryItem = React.memo(({ entry }: { entry: Anime_LibraryCollectionEntry }) => {
    return (
        <MediaEntryCard
            media={entry.media!}
            listData={entry.listData}
            libraryData={entry.libraryData}
            showListDataButton
            withAudienceScore={false}
            type="anime"
        />
    )
})
