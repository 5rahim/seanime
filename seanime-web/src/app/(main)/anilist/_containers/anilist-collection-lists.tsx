import { AL_AnimeCollection_MediaListCollection_Lists } from "@/api/generated/types"
import { AnilistMediaEntryList } from "@/app/(main)/_features/anime/_components/anilist-media-entry-list"
import { useHandleUserAnilistLists } from "@/app/(main)/anilist/_lib/handle-user-anilist-lists"
import { Separator } from "@/components/ui/separator"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { TextInput } from "@/components/ui/text-input"
import { useDebounce } from "@/hooks/use-debounce"
import { atom } from "jotai/index"
import { useAtom, useAtomValue } from "jotai/react"
import React, { useState } from "react"
import { FiSearch } from "react-icons/fi"

const selectedIndexAtom = atom("current")
const watchListSearchInputAtom = atom<string>("")

type AnilistCollectionListsProps = {}

export function AnilistCollectionLists(props: AnilistCollectionListsProps) {

    const {} = props

    const [selectedIndex, setSelectedIndex] = useAtom(selectedIndexAtom)
    const searchInput = useAtomValue(watchListSearchInputAtom)
    const debouncedSearchInput = useDebounce(searchInput, 500)

    const {
        currentList,
        planningList,
        pausedList,
        completedList,
        droppedList,
        customLists,
    } = useHandleUserAnilistLists(debouncedSearchInput)

    React.useEffect(() => {
        const lists = {
            current: currentList,
            planning: planningList,
            paused: pausedList,
            completed: completedList,
            dropped: droppedList,
        } as Record<string, AL_AnimeCollection_MediaListCollection_Lists | undefined>
        if (lists[selectedIndex]?.entries?.length === 0) {
            React.startTransition(() => {
                if (!!currentList?.entries && currentList?.entries?.length > 0) setSelectedIndex("current")
                if (!!planningList?.entries && planningList?.entries?.length > 0) setSelectedIndex("planning")
                if (!!pausedList?.entries && pausedList?.entries?.length > 0) setSelectedIndex("paused")
                if (!!completedList?.entries && completedList?.entries?.length > 0) setSelectedIndex("completed")
                if (!!droppedList?.entries && droppedList?.entries?.length > 0) setSelectedIndex("dropped")
            })
        }

    }, [selectedIndex, debouncedSearchInput])

    return (
        <>
            <SearchInput />

            <Tabs
                triggerClass="w-fit md:w-full rounded-full border-none data-[state=active]:border-none data-[state=active]:bg-[--subtle] data-[state=active]:text-[--brand]"
                listClass="w-full flex flex-wrap md:flex-nowrap h-fit md:h-12"
                value={selectedIndex}
                onValueChange={value => {
                    React.startTransition(() => {
                        setSelectedIndex(value)
                    })
                }}
            >
                <TabsList className="block lg:h-auto space-y-2">
                    <div className="inline-flex flex-wrap lg:flex-nowrap lg:h-12 items-center justify-center w-full">
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
                    </div>

                    {!!customLists?.length && (
                        <>
                            <Separator />
                            <div className="inline-flex flex-wrap lg:flex-nowrap lg:h-10 items-center justify-center w-full">
                                {customLists.map((list, i) => (
                                    <TabsTrigger key={list.name} value={list.name || ""} className="">
                                        {list?.name}
                                    </TabsTrigger>
                                ))}
                            </div>
                        </>
                    )}
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
                    {customLists?.map(list => (
                        <TabsContent key={list.name} value={list.name || ""}>
                            <AnilistMediaEntryList list={list} />
                        </TabsContent>
                    ))}
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
                    setter(v)
                }}
            />
        </div>
    )
}
