import { Anime_Episode, Anime_LibraryCollectionEntry, Anime_LibraryCollectionList } from "@/api/generated/types"
import {
    __library_debouncedSearchInputAtom,
    __library_paramsAtom,
    __library_selectedListAtom,
    DETAILED_LIBRARY_DEFAULT_PARAMS,
    useHandleDetailedLibraryCollection,
} from "@/app/(main)/(library)/_lib/handle-detailed-library-collection"
import { __library_viewAtom } from "@/app/(main)/(library)/_lib/library-view.atoms"
import { MediaCardLazyGrid } from "@/app/(main)/_features/media/_components/media-card-grid"
import { MediaEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"
import { MediaGenreSelector } from "@/app/(main)/_features/media/_components/media-genre-selector"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import {
    ADVANCED_SEARCH_FORMATS,
    ADVANCED_SEARCH_MEDIA_GENRES,
    ADVANCED_SEARCH_SEASONS,
    ADVANCED_SEARCH_STATUS,
} from "@/app/(main)/search/_lib/advanced-search-constants"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Select } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { StaticTabs } from "@/components/ui/tabs"
import { TextInput } from "@/components/ui/text-input"
import { useDebounce } from "@/hooks/use-debounce"
import { ANIME_COLLECTION_SORTING_OPTIONS } from "@/lib/helpers/filtering"
import { getLibraryCollectionTitle } from "@/lib/server/utils"
import { useThemeSettings } from "@/lib/theme/hooks"
import { getYear } from "date-fns"
import { useAtomValue, useSetAtom } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"
import { BiTrash } from "react-icons/bi"
import { FaSortAmountDown } from "react-icons/fa"
import { FiSearch } from "react-icons/fi"
import { LuCalendar, LuLeaf } from "react-icons/lu"
import { MdPersonalVideo } from "react-icons/md"
import { RiSignalTowerLine } from "react-icons/ri"

type LibraryViewProps = {
    collectionList: Anime_LibraryCollectionList[]
    continueWatchingList: Anime_Episode[]
    isLoading: boolean
    hasScanned: boolean
}

export function DetailedLibraryView(props: LibraryViewProps) {

    const {
        // collectionList: _collectionList,
        continueWatchingList,
        isLoading,
        hasScanned,
        ...rest
    } = props

    const ts = useThemeSettings()
    const setView = useSetAtom(__library_viewAtom)

    const {
        stats,
        libraryCollectionList,
    } = useHandleDetailedLibraryCollection()

    if (isLoading) return <LoadingSpinner />

    if (!hasScanned) return null

    return (
        <PageWrapper className="p-4 space-y-8 relative z-[4]" data-detailed-library-view-container>

            <div className="flex flex-col md:flex-row gap-4 justify-between" data-detailed-library-view-header-container>
                <div className="flex gap-4 items-center relative w-fit">
                    <IconButton
                        icon={<AiOutlineArrowLeft />}
                        rounded
                        intent="white-outline"
                        size="sm"
                        onClick={() => setView("base")}
                    />
                    <h3 className="text-ellipsis truncate">Library</h3>
                </div>

                <SearchInput />
            </div>

            <div
                className="grid grid-cols-3 lg:grid-cols-6 gap-4 [&>div]:text-center [&>div>p]:text-[--muted]"
                data-detailed-library-view-stats-container
            >
                <div>
                    <h3>{stats?.totalSize}</h3>
                    <p>Library</p>
                </div>
                <div>
                    <h3>{stats?.totalFiles}</h3>
                    <p>Files</p>
                </div>
                <div>
                    <h3>{stats?.totalEntries}</h3>
                    <p>Entries</p>
                </div>
                <div>
                    <h3>{stats?.totalShows}</h3>
                    <p>TV Shows</p>
                </div>
                <div>
                    <h3>{stats?.totalMovies}</h3>
                    <p>Movies</p>
                </div>
                <div>
                    <h3>{stats?.totalSpecials}</h3>
                    <p>Specials</p>
                </div>
            </div>

            <SearchOptions />

            <GenreSelector />

            {libraryCollectionList.map(collection => {
                if (!collection.entries?.length) return null
                return <LibraryCollectionListItem key={collection.type} list={collection} />
            })}
        </PageWrapper>
    )
}

const LibraryCollectionListItem = React.memo(({ list }: { list: Anime_LibraryCollectionList }) => {

    const selectedList = useAtomValue(__library_selectedListAtom)

    if (selectedList !== "-" && selectedList !== list.type) return null

    return (
        <React.Fragment key={list.type}>
            <h2>{getLibraryCollectionTitle(list.type)} <span className="text-[--muted] font-medium ml-3">{list?.entries?.length ?? 0}</span></h2>
            <MediaCardLazyGrid itemCount={list?.entries?.length || 0}>
                {list.entries?.map(entry => {
                    return <LibraryCollectionEntryItem key={entry.mediaId} entry={entry} />
                })}
            </MediaCardLazyGrid>
        </React.Fragment>
    )
})

const LibraryCollectionEntryItem = React.memo(({ entry }: { entry: Anime_LibraryCollectionEntry }) => {
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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const SearchInput = () => {

    const [inputValue, setInputValue] = React.useState("")
    const setDebouncedInput = useSetAtom(__library_debouncedSearchInputAtom)
    const debouncedInput = useDebounce(inputValue, 500)

    React.useEffect(() => {
        setDebouncedInput(inputValue)
    }, [debouncedInput])


    return (
        <div className="w-full md:w-[300px]">
            <TextInput
                leftIcon={<FiSearch />}
                value={inputValue}
                onValueChange={v => {
                    setInputValue(v)
                }}
                className="rounded-full bg-gray-900/50"
            />
        </div>
    )
}

export function SearchOptions() {

    const serverStatus = useServerStatus()
    const [params, setParams] = useAtom(__library_paramsAtom)
    const [selectedIndex, setSelectedIndex] = useAtom(__library_selectedListAtom)

    return (
        <AppLayoutStack className="px-4 xl:px-0" data-detailed-library-view-search-options-container>
            <div className="flex w-full justify-center">
                <StaticTabs
                    className="h-10 w-fit pb-6"
                    triggerClass="px-4 py-1"
                    items={[
                        { name: "All", isCurrent: selectedIndex === "-", onClick: () => setSelectedIndex("-") },
                        { name: "Watching", isCurrent: selectedIndex === "CURRENT", onClick: () => setSelectedIndex("CURRENT") },
                        { name: "Planning", isCurrent: selectedIndex === "PLANNING", onClick: () => setSelectedIndex("PLANNING") },
                        { name: "Paused", isCurrent: selectedIndex === "PAUSED", onClick: () => setSelectedIndex("PAUSED") },
                        { name: "Completed", isCurrent: selectedIndex === "COMPLETED", onClick: () => setSelectedIndex("COMPLETED") },
                        { name: "Dropped", isCurrent: selectedIndex === "DROPPED", onClick: () => setSelectedIndex("DROPPED") },
                    ]}
                />
            </div>
            <div
                className="grid grid-cols-2 md:grid-cols-3 2xl:grid-cols-[1fr_1fr_1fr_1fr_1fr_auto_auto] gap-4"
                data-detailed-library-view-search-options-grid
            >
                <Select
                    label="Sorting"
                    leftAddon={<FaSortAmountDown className={cn(params.sorting !== "TITLE" && "text-indigo-300 font-bold text-xl")} />}
                    className="w-full"
                    fieldClass="flex items-center"
                    inputContainerClass="w-full"
                    options={ANIME_COLLECTION_SORTING_OPTIONS}
                    value={params.sorting || "TITLE"}
                    onValueChange={v => setParams(draft => {
                        draft.sorting = v as any
                        return
                    })}
                    fieldLabelClass="hidden"
                    // disabled={!!params.title && params.title.length > 0}
                />
                <Select
                    leftAddon={
                        <MdPersonalVideo className={cn((params.format as any) !== null && (params.format as any) !== "" && "text-indigo-300 font-bold text-xl")} />}
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
                    leftAddon={
                        <RiSignalTowerLine className={cn((params.status as any) !== null && (params.status as any) !== "" && "text-indigo-300 font-bold text-xl")} />}
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
                <Select
                    leftAddon={
                        <LuLeaf className={cn((params.season as any) !== null && (params.season as any) !== "" && "text-indigo-300 font-bold text-xl")} />}
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
                />
                <Select
                    leftAddon={<LuCalendar className={cn((params.year !== null && params.year !== "") && "text-indigo-300 font-bold text-xl")} />}
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
                <div className="flex gap-4 items-center w-full">
                    <IconButton
                        icon={<BiTrash />} intent="alert-subtle" className="flex-none" onClick={() => {
                        setParams(DETAILED_LIBRARY_DEFAULT_PARAMS)
                    }}
                    />
                </div>
                {serverStatus?.settings?.anilist?.enableAdultContent && <div className="flex h-full items-center">
                    <Switch
                        label="Adult"
                        value={params.isAdult}
                        onValueChange={v => setParams(draft => {
                            draft.isAdult = v
                            return
                        })}
                        fieldLabelClass="hidden"
                    />
                </div>}
            </div>

        </AppLayoutStack>
    )
}

function GenreSelector() {
    const [params, setParams] = useAtom(__library_paramsAtom)
    return (
        <MediaGenreSelector
            items={[
                {
                    name: "All",
                    isCurrent: !params!.genre?.length,
                    onClick: () => setParams(draft => {
                        draft.genre = []
                        return
                    }),
                },
                ...ADVANCED_SEARCH_MEDIA_GENRES.map(genre => ({
                    name: genre,
                    isCurrent: params!.genre?.includes(genre) ?? false,
                    onClick: () => setParams(draft => {
                        if (draft.genre?.includes(genre)) {
                            draft.genre = draft.genre?.filter(g => g !== genre)
                        } else {
                            draft.genre = [...(draft.genre || []), genre]
                        }
                        return
                    }),
                })),
            ]}
        />
    )
}
