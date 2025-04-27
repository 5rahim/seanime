import { AL_AnimeCollection_MediaListCollection_Lists } from "@/api/generated/types"
import { useGetAniListStats } from "@/api/hooks/anilist.hooks"
import { AnilistAnimeEntryList } from "@/app/(main)/_features/anime/_components/anilist-media-entry-list"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { AnilistStats } from "@/app/(main)/anilist/_containers/anilist-stats"
import {
    __myLists_selectedTypeAtom,
    __myListsSearch_paramsAtom,
    __myListsSearch_paramsInputAtom,
    useHandleUserAnilistLists,
} from "@/app/(main)/anilist/_lib/handle-user-anilist-lists"
import {
    ADVANCED_SEARCH_FORMATS,
    ADVANCED_SEARCH_MEDIA_GENRES,
    ADVANCED_SEARCH_SEASONS,
    ADVANCED_SEARCH_STATUS,
} from "@/app/(main)/search/_lib/advanced-search-constants"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { IconButton } from "@/components/ui/button"
import { Combobox } from "@/components/ui/combobox"
import { cn } from "@/components/ui/core/styling"
import { Select } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { StaticTabs } from "@/components/ui/tabs"
import { TextInput } from "@/components/ui/text-input"
import { useDebounce } from "@/hooks/use-debounce"
import { COLLECTION_SORTING_OPTIONS } from "@/lib/helpers/filtering"
import { getYear } from "date-fns"
import { AnimatePresence } from "framer-motion"
import { atom } from "jotai/index"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import React from "react"
import { BiTrash } from "react-icons/bi"
import { FaSortAmountDown } from "react-icons/fa"
import { FiSearch } from "react-icons/fi"
import { LuCalendar, LuLeaf } from "react-icons/lu"
import { MdPersonalVideo } from "react-icons/md"
import { RiSignalTowerLine } from "react-icons/ri"
import { TbSwords } from "react-icons/tb"

const selectedIndexAtom = atom("-")
const watchListSearchInputAtom = atom<string>("")

export function AnilistCollectionLists() {
    const serverStatus = useServerStatus()
    const [pageType, setPageType] = useAtom(__myLists_selectedTypeAtom)
    const [selectedIndex, setSelectedIndex] = useAtom(selectedIndexAtom)
    const searchInput = useAtomValue(watchListSearchInputAtom)
    const debouncedSearchInput = useDebounce(searchInput, 500)

    const {
        currentList,
        repeatingList,
        planningList,
        pausedList,
        completedList,
        droppedList,
        customLists,
    } = useHandleUserAnilistLists(debouncedSearchInput)

    const { data: stats, isLoading: statsLoading } = useGetAniListStats()

    const setParams = useSetAtom(__myListsSearch_paramsAtom)

    // useMount(() => {
    //     setParams({
    //         sorting: "SCORE_DESC",
    //         genre: null,
    //         status: null,
    //         format: null,
    //         season: null,
    //         year: null,
    //         isAdult: false,
    //         unreadOnly: false,
    //         continueWatchingOnly: false,
    //     })
    // })

    return (
        <AppLayoutStack className="space-y-6" data-anilist-collection-lists>
            <div className="w-full flex justify-center" data-anilist-collection-lists-tabs-container>
                <StaticTabs
                    data-anilist-collection-lists-tabs
                    className="h-10 w-fit border rounded-full"
                    triggerClass="px-4 py-1"
                    items={[
                        { name: "Anime", isCurrent: pageType === "anime", onClick: () => setPageType("anime") },
                        ...[serverStatus?.settings?.library?.enableManga && {
                            name: "Manga",
                            isCurrent: pageType === "manga",
                            onClick: () => setPageType("manga"),
                        }],
                        { name: "Stats", isCurrent: pageType === "stats", onClick: () => setPageType("stats") },
                    ].filter(Boolean)}
                />
            </div>


            <AnimatePresence mode="wait" initial={false} data-anilist-collection-lists-content>
                {pageType !== "stats" && <PageWrapper
                    key="lists"
                    className="space-y-6"
                    {...{
                        initial: { opacity: 0 },
                        animate: { opacity: 1 },
                        exit: { opacity: 0 },
                        transition: {
                            duration: 0.35,
                        },
                    }}
                >
                    <SearchOptions customLists={customLists} />

                    <div className="py-6 space-y-6" data-anilist-collection-lists-stack>
                        {(!!currentList?.entries?.length && ["-", "CURRENT"].includes(selectedIndex)) && <>
                            <h2>Current <span className="text-[--muted] font-medium ml-3">{currentList?.entries?.length}</span></h2>
                            <AnilistAnimeEntryList type={pageType} list={currentList} />
                        </>}
                        {(!!repeatingList?.entries?.length && ["-", "REPEATING"].includes(selectedIndex)) && <>
                            <h2>Repeating <span className="text-[--muted] font-medium ml-3">{repeatingList?.entries?.length}</span></h2>
                            <AnilistAnimeEntryList type={pageType} list={repeatingList} />
                        </>}
                        {(!!planningList?.entries?.length && ["-", "PLANNING"].includes(selectedIndex)) && <>
                            <h2>Planning <span className="text-[--muted] font-medium ml-3">{planningList?.entries?.length}</span></h2>
                            <AnilistAnimeEntryList type={pageType} list={planningList} />
                        </>}
                        {(!!pausedList?.entries?.length && ["-", "PAUSED"].includes(selectedIndex)) && <>
                            <h2>Paused <span className="text-[--muted] font-medium ml-3">{pausedList?.entries?.length}</span></h2>
                            <AnilistAnimeEntryList type={pageType} list={pausedList} />
                        </>}
                        {(!!completedList?.entries?.length && ["-", "COMPLETED"].includes(selectedIndex)) && <>
                            <h2>Completed <span className="text-[--muted] font-medium ml-3">{completedList?.entries?.length}</span></h2>
                            <AnilistAnimeEntryList type={pageType} list={completedList} />
                        </>}
                        {(!!droppedList?.entries?.length && ["-", "DROPPED"].includes(selectedIndex)) && <>
                            <h2>Dropped <span className="text-[--muted] font-medium ml-3">{droppedList?.entries?.length}</span></h2>
                            <AnilistAnimeEntryList type={pageType} list={droppedList} />
                        </>}
                        {customLists?.map(list => {
                            return (!!list.entries?.length && ["-", list.name || "N/A"].includes(selectedIndex)) ? <div
                                key={list.name}
                                className="space-y-6"
                            >
                                <h2>{list.name}</h2>
                                <AnilistAnimeEntryList type={pageType} list={list} />
                            </div> : null
                        })}
                    </div>
                </PageWrapper>}

                {pageType === "stats" && <PageWrapper
                    key="stats"
                    className="space-y-6"
                    {...{
                        initial: { opacity: 0, y: 60 },
                        animate: { opacity: 1, y: 0 },
                        exit: { opacity: 0, scale: 0.99 },
                        transition: {
                            duration: 0.35,
                        },
                    }}
                    data-anilist-collection-lists-stats-wrapper
                >
                    <AnilistStats
                        stats={stats}
                        isLoading={statsLoading}
                    />
                </PageWrapper>}
            </AnimatePresence>

        </AppLayoutStack>
    )
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const SearchInput = () => {

    const [input, setter] = useAtom(watchListSearchInputAtom)

    return (
        <div className="w-full">
            <TextInput
                leftIcon={<FiSearch />}
                value={input}
                onValueChange={v => {
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
    const [pageType, setPageType] = useAtom(__myLists_selectedTypeAtom)

    React.useEffect(() => {
        setActualParams(params)
    }, [debouncedParams])

    const [input, setInput] = useAtom(watchListSearchInputAtom)

    const highlightTrash = React.useMemo(() => {
        return !(!input.length && params.sorting === "SCORE_DESC" && (params.genre === null || !params.genre.length) && params.status === null && params.format === null && params.season === null && params.year === null && params.isAdult === false)
    }, [params, input])

    return (
        <AppLayoutStack className="px-4 xl:px-0" data-anilist-collection-lists-search-options-stack>
            <div className="flex flex-col lg:flex-row gap-4" data-anilist-collection-lists-search-options-container>
                <Select
                    // label="Sorting"
                    className="w-full"
                    fieldClass="lg:w-[200px]"
                    options={[
                        { value: "-", label: "All lists" },
                        { value: "CURRENT", label: "Watching" },
                        { value: "REPEATING", label: "Repeating" },
                        { value: "PLANNING", label: "Planning" },
                        { value: "PAUSED", label: "Paused" },
                        { value: "COMPLETED", label: "Completed" },
                        { value: "DROPPED", label: "Dropped" },
                        ...(customLists || []).map(list => ({ value: list.name || "N/A", label: list.name || "N/A" })),
                    ]}
                    value={selectedIndex || "-"}
                    onValueChange={v => setSelectedIndex(v as any)}
                    // disabled={!!params.title && params.title.length > 0}
                />
                <div className="flex gap-4 items-center w-full" data-anilist-collection-lists-search-options-actions>
                    <SearchInput />
                    <IconButton
                        icon={<BiTrash />} intent={highlightTrash ? "alert" : "gray-subtle"} className="flex-none" onClick={() => {
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
                        setInput("")
                    }}
                        disabled={!highlightTrash}
                    />
                </div>
            </div>
            <div
                className={cn(
                    "grid grid-cols-2 gap-5",
                    pageType === "anime" ? "xl:grid-cols-6" : "lg:grid-cols-4",
                )}
                data-anilist-collection-lists-search-options-grid
            >
                <Combobox
                    multiple
                    leftAddon={<TbSwords className={cn((params.genre !== null && !!params.genre?.length) && "text-indigo-300 font-bold text-xl")} />}
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
                    leftAddon={<FaSortAmountDown className={cn((params.sorting !== "SCORE_DESC") && "text-indigo-300 font-bold text-xl")} />}
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
                {pageType === "anime" && <Select
                    leftAddon={
                        <MdPersonalVideo className={cn((params.format !== null && !!params.format?.length) && "text-indigo-300 font-bold text-xl")} />}
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
                />}
                <Select
                    leftAddon={
                        <RiSignalTowerLine className={cn((params.status !== null && !!params.status?.length) && "text-indigo-300 font-bold text-xl")} />}
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
                {pageType === "anime" && <Select
                    leftAddon={<LuLeaf className={cn((params.season !== null && !!params.season?.length) && "text-indigo-300 font-bold text-xl")} />}
                    label="Season"
                    placeholder="All seasons"
                    className="w-full"
                    fieldClass="w-full flex items-center"
                    inputContainerClass="w-full"
                    options={ADVANCED_SEARCH_SEASONS.map(season => ({ value: season.toUpperCase(), label: season }))}
                    value={params.season || ""}
                    onValueChange={v => setParams(draft => {
                        draft.season = v as any
                        return
                    })}
                    fieldLabelClass="hidden"
                />}
                <Select
                    leftAddon={<LuCalendar className={cn((params.year !== null && !!params.year?.length) && "text-indigo-300 font-bold text-xl")} />}
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
