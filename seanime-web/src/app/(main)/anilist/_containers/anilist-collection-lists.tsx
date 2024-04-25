import { AnilistMediaEntryList } from "@/app/(main)/_features/anime/_components/anilist-media-entry-list"
import { useHandleUserAnilistLists } from "@/app/(main)/anilist/_lib/handle-user-anilist-lists"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { TextInput } from "@/components/ui/text-input"
import { useDebounce } from "@/hooks/use-debounce"
import { atom } from "jotai/index"
import { useAtom, useAtomValue } from "jotai/react"
import React, { startTransition, useState, useTransition } from "react"
import { FiSearch } from "react-icons/fi"

const selectedIndexAtom = atom("current")
const watchListSearchInputAtom = atom<string>("")

type AnilistCollectionListsProps = {}

export function AnilistCollectionLists(props: AnilistCollectionListsProps) {

    const {} = props

    const [selectedIndex, setSelectedIndex] = useAtom(selectedIndexAtom)
    const [pending, startTransition] = useTransition()
    const searchInput = useAtomValue(watchListSearchInputAtom)
    const debouncedSearchInput = useDebounce(searchInput, 500)

    const {
        currentList,
        planningList,
        pausedList,
        completedList,
        droppedList,
    } = useHandleUserAnilistLists(debouncedSearchInput)

    return (
        <>
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


                <div className="py-6">
                    <TabsContent value="current">
                        <AnilistMediaEntryList list={currentList} />
                    </TabsContent>
                    <TabsContent value="planning">
                        <AnilistMediaEntryList list={planningList} />
                    </TabsContent>
                    <TabsContent value="paused">
                        <AnilistMediaEntryList list={pausedList} />
                    </TabsContent>
                    <TabsContent value="completed">
                        <AnilistMediaEntryList list={completedList} />
                    </TabsContent>
                    <TabsContent value="dropped">
                        <AnilistMediaEntryList list={droppedList} />
                    </TabsContent>
                </div>
            </Tabs>
        </>
    )
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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
