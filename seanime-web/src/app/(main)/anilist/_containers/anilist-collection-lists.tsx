import { AL_AnimeCollection_MediaListCollection_Lists } from "@/api/generated/types"
import { AnilistMediaEntryList } from "@/app/(main)/_features/anime/_components/anilist-media-entry-list"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import {
    __myListsSearch_paramsAtom,
    __myListsSearch_paramsInputAtom,
    useHandleUserAnilistLists,
} from "@/app/(main)/anilist/_lib/handle-user-anilist-lists"
import { ADVANCED_SEARCH_FORMATS, ADVANCED_SEARCH_MEDIA_GENRES, ADVANCED_SEARCH_STATUS } from "@/app/(main)/search/_lib/advanced-search-constants"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { IconButton } from "@/components/ui/button"
import { Combobox } from "@/components/ui/combobox"
import { Select } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { TextInput } from "@/components/ui/text-input"
import { useDebounce } from "@/hooks/use-debounce"
import { COLLECTION_SORTING_OPTIONS } from "@/lib/helpers/filtering"
import { getYear } from "date-fns"
import { atom } from "jotai/index"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import React, { useState } from "react"
import { BiTrash } from "react-icons/bi"
import { FaSortAmountDown } from "react-icons/fa"
import { FiSearch } from "react-icons/fi"
import { LuCalendar } from "react-icons/lu"
import { MdPersonalVideo } from "react-icons/md"
import { RiSignalTowerLine } from "react-icons/ri"
import { TbSwords } from "react-icons/tb"
import { useMount } from "react-use"

const selectedIndexAtom = atom("-")
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

    const setParams = useSetAtom(__myListsSearch_paramsAtom)

    useMount(() => {
        setParams({
            sorting: "SCORE_DESC",
            genre: null,
            status: null,
            format: null,
            season: null,
            year: null,
            isAdult: false,
        })
    })

    return (
        <>
            <SearchOptions customLists={customLists} />

            <div className="py-6 space-y-6">
                {(!!currentList?.entries?.length && ["-", "current"].includes(selectedIndex)) && <>
                    <h2>Watching</h2>
                    <AnilistMediaEntryList list={currentList} />
                </>}
                {(!!planningList?.entries?.length && ["-", "planning"].includes(selectedIndex)) && <>
                    <h2>Planning</h2>
                    <AnilistMediaEntryList list={planningList} />
                </>}
                {(!!pausedList?.entries?.length && ["-", "paused"].includes(selectedIndex)) && <>
                    <h2>Paused</h2>
                    <AnilistMediaEntryList list={pausedList} />
                </>}
                {(!!completedList?.entries?.length && ["-", "completed"].includes(selectedIndex)) && <>
                    <h2>Completed</h2>
                    <AnilistMediaEntryList list={completedList} />
                </>}
                {(!!droppedList?.entries?.length && ["-", "dropped"].includes(selectedIndex)) && <>
                    <h2>Dropped</h2>
                    <AnilistMediaEntryList list={droppedList} />
                </>}
                {customLists?.map(list => {
                    return (!!list.entries?.length && ["-", list.name || "N/A"].includes(selectedIndex)) ? <div key={list.name} className="space-y-6">
                        <h2>{list.name}</h2>
                        <AnilistMediaEntryList list={list} />
                    </div> : null
                })}
            </div>
        </>
    )
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const SearchInput = () => {

    const [input, setter] = useAtom(watchListSearchInputAtom)
    const [inputValue, setInputValue] = useState(input)

    return (
        <div className="w-full">
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

export function SearchOptions({
    customLists,
}: {
    customLists?: AL_AnimeCollection_MediaListCollection_Lists[]
}) {

    const serverStatus = useServerStatus()
    const [params, setParams] = useAtom(__myListsSearch_paramsInputAtom)
    const setActualParams = useSetAtom(__myListsSearch_paramsAtom)
    const debouncedParams = useDebounce(params, 500)
    const [selectedIndex, setSelectedIndex] = useAtom(selectedIndexAtom)

    React.useEffect(() => {
        setActualParams(params)
    }, [debouncedParams])

    return (
        <AppLayoutStack className="px-4 xl:px-0">
            <div className="flex flex-col lg:flex-row gap-4">
                <Select
                    // label="Sorting"
                    className="w-full"
                    fieldClass="lg:w-[200px]"
                    options={[
                        { value: "-", label: "All lists" },
                        { value: "current", label: "Watching" },
                        { value: "planning", label: "Planning" },
                        { value: "paused", label: "Paused" },
                        { value: "completed", label: "Completed" },
                        { value: "dropped", label: "Dropped" },
                        ...(customLists || []).map(list => ({ value: list.name || "N/A", label: list.name || "N/A" })),
                    ]}
                    value={selectedIndex || "-"}
                    onValueChange={v => setSelectedIndex(v as any)}
                    // disabled={!!params.title && params.title.length > 0}
                />
                <div className="flex gap-4 items-center w-full">
                    <SearchInput />
                    <IconButton
                        icon={<BiTrash />} intent="gray-subtle" className="flex-none" onClick={() => {
                        setParams(prev => ({
                            ...prev,
                            sorting: "SCORE_DESC",
                            genre: null,
                            status: null,
                            format: null,
                            season: null,
                            year: null,
                            isAdult: false,
                        }))
                    }}
                    />
                </div>
            </div>
            <div className="grid grid-cols-2 lg:grid-cols-5 gap-5">
                <Combobox
                    multiple
                    leftAddon={<TbSwords />}
                    emptyMessage="No options found"
                    label="Genre" placeholder="All genres"
                    className="w-full"
                    fieldClass="w-full"
                    options={ADVANCED_SEARCH_MEDIA_GENRES.map(genre => ({ value: genre, label: genre, textValue: genre }))}
                    value={params.genre ? params.genre : []}
                    onValueChange={v => setParams(draft => {
                        draft.genre = v
                        return
                    })}
                    fieldLabelClass="hidden"
                />
                <Select
                    label="Sorting"
                    leftAddon={<FaSortAmountDown />}
                    className="w-full"
                    fieldClass="flex items-center"
                    inputContainerClass="w-full"
                    options={COLLECTION_SORTING_OPTIONS}
                    value={params.sorting || "SCORE_DESC"}
                    onValueChange={v => setParams(draft => {
                        draft.sorting = v as any
                        return
                    })}
                    fieldLabelClass="hidden"
                    // disabled={!!params.title && params.title.length > 0}
                />
                <Select
                    leftAddon={<MdPersonalVideo />}
                    label="Format" placeholder="All formats"
                    className="w-full"
                    fieldClass="w-full"
                    options={ADVANCED_SEARCH_FORMATS}
                    value={params.format || ""}
                    onValueChange={v => setParams(draft => {
                        draft.format = v as any
                        return
                    })}
                    fieldLabelClass="hidden"
                />
                <Select
                    leftAddon={<LuCalendar />}
                    label="Year" placeholder="Timeless"
                    className="w-full"
                    fieldClass="w-full"
                    options={[...Array(70)].map((v, idx) => getYear(new Date()) - idx).map(year => ({
                        value: String(year),
                        label: String(year),
                    }))}
                    value={params.year || ""}
                    onValueChange={v => setParams(draft => {
                        draft.year = v as any
                        return
                    })}
                    fieldLabelClass="hidden"
                />
                <Select
                    leftAddon={<RiSignalTowerLine />}
                    label="Status" placeholder="All statuses"
                    className="w-full"
                    fieldClass="w-full"
                    options={[
                        ...ADVANCED_SEARCH_STATUS,
                    ]}
                    value={params.status || ""}
                    onValueChange={v => setParams(draft => {
                        draft.status = v as any
                        return
                    })}
                    fieldLabelClass="hidden"
                />
                {/*<Select*/}
                {/*    leftAddon={<LuLeaf />}*/}
                {/*    label="Season"*/}
                {/*    placeholder="All seasons"*/}
                {/*    className="w-full"*/}
                {/*    fieldClass="w-full flex items-center"*/}
                {/*    inputContainerClass="w-full"*/}
                {/*    options={ADVANCED_SEARCH_SEASONS.map(season => ({ value: season.toUpperCase(), label: season }))}*/}
                {/*    value={params.season || ""}*/}
                {/*    onValueChange={v => setParams(draft => {*/}
                {/*        draft.season = v as any*/}
                {/*        return*/}
                {/*    })}*/}
                {/*    fieldLabelClass="hidden"*/}
                {/*/>*/}
            </div>

            {serverStatus?.settings?.anilist?.enableAdultContent && <Switch
                label="Adult"
                value={params.isAdult}
                onValueChange={v => setParams(draft => {
                    draft.isAdult = v
                    return
                })}
                fieldLabelClass="hidden"
            />}

        </AppLayoutStack>
    )
}
