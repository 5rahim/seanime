"use client"

import { useAnilistCollection } from "@/app/(main)/_loaders/anilist-collection"
import { AnimeListItem } from "@/components/shared/anime-list-item"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { TextInput } from "@/components/ui/text-input"
import { useDebounce } from "@/hooks/use-debounce"
import { BaseMediaFragment, MediaListStatus } from "@/lib/anilist/gql/graphql"
import { AnilistCollectionEntry, AnilistCollectionList } from "@/lib/server/types"
import { atom } from "jotai"
import { useAtom, useAtomValue } from "jotai/react"
import sortBy from "lodash/sortBy"
import React, { startTransition, useCallback, useMemo, useState, useTransition } from "react"
import { FiSearch } from "react-icons/fi"

const selectedIndexAtom = atom("current")
const watchListSearchInputAtom = atom<string>("")

function anilist_filterEntriesByTitle(arr: AnilistCollectionEntry[], input: string) {
    if (arr.length > 0 && input.length > 0) {
        const _input = input.toLowerCase().trim().replace(/\s+/g, " ")
        return (arr as { media: BaseMediaFragment | null | undefined }[]).filter(entry => (
            entry.media?.title?.english?.toLowerCase().includes(_input)
            || entry.media?.title?.userPreferred?.toLowerCase().includes(_input)
            || entry.media?.title?.romaji?.toLowerCase().includes(_input)
            || entry.media?.synonyms?.some(syn => syn?.toLowerCase().includes(_input))
        )) as AnilistCollectionEntry[]
    }
    return arr
}

export default function Home() {

    const [selectedIndex, setSelectedIndex] = useAtom(selectedIndexAtom)
    const [pending, startTransition] = useTransition()

    const { anilistLists } = useAnilistCollection()

    const searchInput = useAtomValue(watchListSearchInputAtom)
    const search = useDebounce(searchInput, 500)

    const sortedLists = useMemo(() => {
        return anilistLists.map(obj => {
            if (!obj) return undefined
            let arr = obj?.entries as AnilistCollectionEntry[]
            // Sort by name
            arr = sortBy(arr, n => n?.media?.title?.userPreferred).reverse()
            // Sort by score
            arr = (sortBy(arr, n => n?.score).reverse())
            obj.entries = arr
            return obj
        })
    }, [anilistLists])

    const getList = useCallback((status: MediaListStatus) => {
        let obj = structuredClone(sortedLists?.find(n => n?.status === status))
        if (!obj) return undefined
        obj.entries = anilist_filterEntriesByTitle(obj.entries as AnilistCollectionEntry[], search)
        return obj
    }, [sortedLists, search])

    const currentList = useMemo(() => getList("CURRENT"), [search, getList, anilistLists])
    const planningList = useMemo(() => getList("PLANNING"), [search, getList, anilistLists])
    const pausedList = useMemo(() => getList("PAUSED"), [search, getList, anilistLists])
    const completedList = useMemo(() => getList("COMPLETED"), [search, getList, anilistLists])
    const droppedList = useMemo(() => getList("DROPPED"), [search, getList, anilistLists])

    return (
        <PageWrapper
            className="p-4 sm:p-8 pt-4 relative"
        >
            <SearchInput />

            <Tabs
                triggerClass="w-fit md:w-full rounded-full border-none data-[state=active]:border-none data-[state=active]:bg-[--subtle] data-[state=active]:text-[--brand]"
                listClass="w-full flex flex-wrap md:flex-nowrap h-fit md:h-12"
                value={selectedIndex}
                onValueChange={value => {
                    startTransition(() => {
                        setSelectedIndex(value)
                    })
                }}
            >
                <TabsList>
                    <TabsTrigger value="current">
                        Currently Watching
                    </TabsTrigger>
                    <TabsTrigger value="planning">
                        Planning
                    </TabsTrigger>
                    <TabsTrigger value="paused">
                        Paused
                    </TabsTrigger>
                    <TabsTrigger value="completed">
                        Completed
                    </TabsTrigger>
                    <TabsTrigger value="dropped">
                        Dropped
                    </TabsTrigger>
                </TabsList>

                {/*<SearchInput/>*/}

                <div className="py-6">
                    {/*<LoadingOverlay className={cn("z-50 backdrop-blur-none", { "hidden": !pending })} />*/}

                    <TabsContent value="current">
                        <WatchList list={currentList} />
                    </TabsContent>
                    <TabsContent value="planning">
                        <WatchList list={planningList} />
                    </TabsContent>
                    <TabsContent value="paused">
                        <WatchList list={pausedList} />
                    </TabsContent>
                    <TabsContent value="completed">
                        <WatchList list={completedList} />
                    </TabsContent>
                    <TabsContent value="dropped">
                        <WatchList list={droppedList} />
                    </TabsContent>
                </div>
            </Tabs>

        </PageWrapper>
    )
}

const WatchList = React.memo(({ list }: { list: AnilistCollectionList | null | undefined }) => {

    return (
        <div
            className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-7 min-[2000px]:grid-cols-8 gap-4"
        >
            {list?.entries?.filter(Boolean)?.map((entry) => (
                <AnimeListItem
                    key={`${entry.media?.id}`}
                    listData={{
                        progress: entry.progress!,
                        score: entry.score!,
                        status: entry.status!,
                    }}
                    showLibraryBadge
                    media={entry.media!}
                />
            ))}
        </div>
    )

})

const SearchInput = () => {

    const [input, setter] = useAtom(watchListSearchInputAtom)
    const [inputValue, setInputValue] = useState(input)

    return (
        <div className="mb-4">
            <TextInput
                leftIcon={<FiSearch />}
                value={inputValue}
                onValueChange={v => {
                    setInputValue(v)
                    startTransition(() => {
                        setter(v)
                    })
                }}
            />
        </div>
    )
}
