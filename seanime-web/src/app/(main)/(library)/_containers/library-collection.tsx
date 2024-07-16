import { Anime_LibraryCollectionEntry, Anime_LibraryCollectionList } from "@/api/generated/types"
import { MediaCardLazyGrid } from "@/app/(main)/_features/media/_components/media-card-grid"
import { AnimeEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { getLibraryCollectionTitle } from "@/lib/server/utils"
import React from "react"

export function LibraryCollectionLists({ collectionList, isLoading }: {
    collectionList: Anime_LibraryCollectionList[],
    isLoading: boolean
}) {

    return (
        <PageWrapper
            key="library-collection-lists"
            className="space-y-8"
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
    return (
        <React.Fragment key={list.type}>
            <h2>{getLibraryCollectionTitle(list.type)}</h2>
            <MediaCardLazyGrid itemCount={list?.entries?.length || 0}>
                {list.entries?.map(entry => {
                    return <LibraryCollectionEntryItem key={entry.mediaId} entry={entry} />
                })}
            </MediaCardLazyGrid>
        </React.Fragment>
    )
})

export const LibraryCollectionEntryItem = React.memo(({ entry }: { entry: Anime_LibraryCollectionEntry }) => {
    return (
        <AnimeEntryCard
            media={entry.media!}
            listData={entry.listData}
            libraryData={entry.libraryData}
            showListDataButton
            withAudienceScore={false}
            type="anime"
        />
    )
})
