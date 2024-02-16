"use client"

import { useAnilistCollection } from "@/app/(main)/_loaders/anilist-collection"
import { AnimeListItem } from "@/components/shared/anime-list-item"
import { cn } from "@/components/ui/core"
import { LoadingOverlay } from "@/components/ui/loading-spinner"
import { TabPanels } from "@/components/ui/tabs"
import { TextInput } from "@/components/ui/text-input"
import { useDebounce } from "@/hooks/use-debounce"
import { BaseMediaFragment, MediaListStatus } from "@/lib/anilist/gql/graphql"
import { AnilistCollectionEntry, AnilistCollectionList } from "@/lib/server/types"
import { FiSearch } from "@react-icons/all-files/fi/FiSearch"
import { atom } from "jotai"
import { useAtom, useAtomValue } from "jotai/react"
import sortBy from "lodash/sortBy"
import React, { startTransition, useCallback, useMemo, useState, useTransition } from "react"

const selectedIndexAtom = atom(0)
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
        <main className={"p-8 pt-0 relative"}>

            <SearchInput/>

            <TabPanels
                navClassName={"border-none"}
                tabClassName={cn(
                    "text-lg rounded-none border-b border-b-2 border-b-transparent data-[selected=true]:text-white data-[selected=true]:border-brand-400",
                    "hover:bg-transparent dark:hover:bg-transparent hover:text-white",
                    "dark:border-transparent dark:hover:border-b-transparent dark:data-[selected=true]:border-brand-400 dark:data-[selected=true]:text-white",
                    "hover:bg-[--highlight]",
                )}
                selectedIndex={selectedIndex}
                onIndexChange={value => {
                    startTransition(() => {
                        setSelectedIndex(value)
                    })
                }}
            >
                <TabPanels.Nav>
                    <TabPanels.Tab>
                        Currently Watching
                    </TabPanels.Tab>
                    <TabPanels.Tab>
                        Planning
                    </TabPanels.Tab>
                    <TabPanels.Tab>
                        Paused
                    </TabPanels.Tab>
                    <TabPanels.Tab>
                        Completed
                    </TabPanels.Tab>
                    <TabPanels.Tab>
                        Dropped
                    </TabPanels.Tab>
                </TabPanels.Nav>
                <TabPanels.Container className="pt-8 relative">

                    {/*<SearchInput/>*/}

                    <LoadingOverlay className={cn("z-50 backdrop-blur-none", { "hidden": !pending })}/>

                    <TabPanels.Panel>
                        <WatchList list={currentList}/>
                    </TabPanels.Panel>
                    <TabPanels.Panel>
                        <WatchList list={planningList}/>
                    </TabPanels.Panel>
                    <TabPanels.Panel>
                        <WatchList list={pausedList}/>
                    </TabPanels.Panel>
                    <TabPanels.Panel>
                        <WatchList list={completedList}/>
                    </TabPanels.Panel>
                    <TabPanels.Panel>
                        <WatchList list={droppedList}/>
                    </TabPanels.Panel>
                </TabPanels.Container>
            </TabPanels>

        </main>
    )
}

const WatchList = React.memo(({ list }: { list: AnilistCollectionList | null | undefined }) => {

    return (
        <div
            className={"px-4 grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-7 min-[2000px]:grid-cols-8 gap-4"}>
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
        <div className={"mb-8"}>
            <TextInput leftIcon={<FiSearch/>} value={inputValue} onChange={e => {
                setInputValue(e.target.value)
                startTransition(() => {
                    setter(e.target.value)
                })
            }}/>
        </div>
    )
}
