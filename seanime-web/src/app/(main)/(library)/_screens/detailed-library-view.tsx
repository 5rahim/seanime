import { Anime_LibraryCollectionList, Anime_MediaEntryEpisode } from "@/api/generated/types"
import { LibraryCollectionLists } from "@/app/(main)/(library)/_containers/library-collection"
import { __library_viewAtom } from "@/app/(main)/(library)/_lib/library-view.atoms"
import { IconButton } from "@/components/ui/button"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { useThemeSettings } from "@/lib/theme/hooks"
import { useSetAtom } from "jotai"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"

type LibraryViewProps = {
    collectionList: Anime_LibraryCollectionList[]
    continueWatchingList: Anime_MediaEntryEpisode[]
    isLoading: boolean
    hasScanned: boolean
}

export function DetailedLibraryView(props: LibraryViewProps) {

    const {
        collectionList,
        continueWatchingList,
        isLoading,
        hasScanned,
        ...rest
    } = props

    const ts = useThemeSettings()
    const setView = useSetAtom(__library_viewAtom)

    if (isLoading) return <LoadingSpinner />

    if (!hasScanned) return null

    return (
        <>
            <div className="flex p-4 gap-4 items-center relative w-full">
                <IconButton
                    icon={<AiOutlineArrowLeft />}
                    rounded
                    intent="white-outline"
                    size="sm"
                    onClick={() => setView("base")}
                />
                <h3 className="max-w-full lg:max-w-[50%] text-ellipsis truncate">Library</h3>
            </div>


            <LibraryCollectionLists
                collectionList={collectionList}
                isLoading={isLoading}
            />
        </>
    )
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// const SearchInput = () => {
//
//     const [input, setter] = useAtom(watchListSearchInputAtom)
//     const [inputValue, setInputValue] = useState(input)
//
//     return (
//         <div className="w-full">
//             <TextInput
//                 leftIcon={<FiSearch />}
//                 value={inputValue}
//                 onValueChange={v => {
//                     setInputValue(v)
//                     setter(v)
//                 }}
//             />
//         </div>
//     )
// }
//
// export function SearchOptions({
//     customLists,
// }: {
//     customLists?: AL_AnimeCollection_MediaListCollection_Lists[]
// }) {
//
//     const serverStatus = useServerStatus()
//     const [params, setParams] = useAtom(__myListsSearch_paramsInputAtom)
//     const setActualParams = useSetAtom(__myListsSearch_paramsAtom)
//     const debouncedParams = useDebounce(params, 500)
//     const [selectedIndex, setSelectedIndex] = useAtom(selectedIndexAtom)
//
//     React.useEffect(() => {
//         setActualParams(params)
//     }, [debouncedParams])
//
//     return (
//         <AppLayoutStack className="px-4 xl:px-0">
//             <div className="flex flex-col lg:flex-row gap-4">
//                 <Select
//                     // label="Sorting"
//                     className="w-full"
//                     fieldClass="lg:w-[200px]"
//                     options={[
//                         { value: "-", label: "All lists" },
//                         { value: "current", label: "Watching" },
//                         { value: "planning", label: "Planning" },
//                         { value: "paused", label: "Paused" },
//                         { value: "completed", label: "Completed" },
//                         { value: "dropped", label: "Dropped" },
//                         ...(customLists || []).map(list => ({ value: list.name || "N/A", label: list.name || "N/A" })),
//                     ]}
//                     value={selectedIndex || "-"}
//                     onValueChange={v => setSelectedIndex(v as any)}
//                     // disabled={!!params.title && params.title.length > 0}
//                 />
//                 <div className="flex gap-4 items-center w-full">
//                     <SearchInput />
//                     <IconButton
//                         icon={<BiTrash />} intent="gray-subtle" className="flex-none" onClick={() => {
//                         setParams(prev => ({
//                             ...prev,
//                             sorting: "SCORE_DESC",
//                             genre: null,
//                             status: null,
//                             format: null,
//                             season: null,
//                             year: null,
//                             isAdult: false,
//                         }))
//                     }}
//                     />
//                 </div>
//             </div>
//             <div className="grid grid-cols-2 lg:grid-cols-5 gap-5">
//                 <Combobox
//                     multiple
//                     leftAddon={<TbSwords />}
//                     emptyMessage="No options found"
//                     label="Genre" placeholder="All genres"
//                     className="w-full"
//                     fieldClass="w-full"
//                     options={ADVANCED_SEARCH_MEDIA_GENRES.map(genre => ({ value: genre, label: genre, textValue: genre }))}
//                     value={params.genre ? params.genre : []}
//                     onValueChange={v => setParams(draft => {
//                         draft.genre = v
//                         return
//                     })}
//                     fieldLabelClass="hidden"
//                 />
//                 <Select
//                     label="Sorting"
//                     leftAddon={<FaSortAmountDown />}
//                     className="w-full"
//                     fieldClass="flex items-center"
//                     inputContainerClass="w-full"
//                     options={MYLISTS_SORTING_OPTIONS}
//                     value={params.sorting || "SCORE_DESC"}
//                     onValueChange={v => setParams(draft => {
//                         draft.sorting = v as any
//                         return
//                     })}
//                     fieldLabelClass="hidden"
//                     // disabled={!!params.title && params.title.length > 0}
//                 />
//                 <Select
//                     leftAddon={<MdPersonalVideo />}
//                     label="Format" placeholder="All formats"
//                     className="w-full"
//                     fieldClass="w-full"
//                     options={ADVANCED_SEARCH_FORMATS}
//                     value={params.format || ""}
//                     onValueChange={v => setParams(draft => {
//                         draft.format = v as any
//                         return
//                     })}
//                     fieldLabelClass="hidden"
//                 />
//                 <Select
//                     leftAddon={<LuCalendar />}
//                     label="Year" placeholder="Timeless"
//                     className="w-full"
//                     fieldClass="w-full"
//                     options={[...Array(70)].map((v, idx) => getYear(new Date()) - idx).map(year => ({
//                         value: String(year),
//                         label: String(year),
//                     }))}
//                     value={params.year || ""}
//                     onValueChange={v => setParams(draft => {
//                         draft.year = v as any
//                         return
//                     })}
//                     fieldLabelClass="hidden"
//                 />
//                 <Select
//                     leftAddon={<RiSignalTowerLine />}
//                     label="Status" placeholder="All statuses"
//                     className="w-full"
//                     fieldClass="w-full"
//                     options={[
//                         ...ADVANCED_SEARCH_STATUS,
//                     ]}
//                     value={params.status || ""}
//                     onValueChange={v => setParams(draft => {
//                         draft.status = v as any
//                         return
//                     })}
//                     fieldLabelClass="hidden"
//                 />
//                 {/*<Select*/}
//                 {/*    leftAddon={<LuLeaf />}*/}
//                 {/*    label="Season"*/}
//                 {/*    placeholder="All seasons"*/}
//                 {/*    className="w-full"*/}
//                 {/*    fieldClass="w-full flex items-center"*/}
//                 {/*    inputContainerClass="w-full"*/}
//                 {/*    options={ADVANCED_SEARCH_SEASONS.map(season => ({ value: season.toUpperCase(), label: season }))}*/}
//                 {/*    value={params.season || ""}*/}
//                 {/*    onValueChange={v => setParams(draft => {*/}
//                 {/*        draft.season = v as any*/}
//                 {/*        return*/}
//                 {/*    })}*/}
//                 {/*    fieldLabelClass="hidden"*/}
//                 {/*/>*/}
//             </div>
//
//             {serverStatus?.settings?.anilist?.enableAdultContent && <Switch
//                 label="Adult"
//                 value={params.isAdult}
//                 onValueChange={v => setParams(draft => {
//                     draft.isAdult = v
//                     return
//                 })}
//                 fieldLabelClass="hidden"
//             />}
//
//         </AppLayoutStack>
//     )
// }
