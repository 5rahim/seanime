import { AL_AnimeCollection_MediaListCollection_Lists } from "@/api/generated/types"
import { useGetRawAnimeCollection } from "@/api/hooks/anilist.hooks"
import { useGetRawAnilistMangaCollection } from "@/api/hooks/manga.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { CollectionParams, CollectionType, DEFAULT_COLLECTION_PARAMS, filterEntriesByTitle, filterListEntries } from "@/lib/helpers/filtering"
import { atomWithImmer } from "jotai-immer"
import { useAtom } from "jotai/react"
import React from "react"
import { useDebounce } from "use-debounce"

export const MYLISTS_DEFAULT_PARAMS: CollectionParams<"anime"> | CollectionParams<"manga"> = {
    ...DEFAULT_COLLECTION_PARAMS,
    sorting: "SCORE_DESC",
    unreadOnly: false,
    continueWatchingOnly: false,
}

export const __myListsSearch_paramsAtom = atomWithImmer<CollectionParams<"anime"> | CollectionParams<"manga">>(MYLISTS_DEFAULT_PARAMS)

export const __myListsSearch_paramsInputAtom = atomWithImmer<CollectionParams<"anime"> | CollectionParams<"manga">>(MYLISTS_DEFAULT_PARAMS)

export const __myLists_selectedTypeAtom = atomWithImmer<"anime" | "manga" | "stats">("anime")

export function useHandleUserAnilistLists(debouncedSearchInput: string, type?: "anime" | "manga") {

    const serverStatus = useServerStatus()
    const [selectedType, setSelectedType] = useAtom(__myLists_selectedTypeAtom)
    const { data: animeData } = useGetRawAnimeCollection()
    const { data: mangaData } = useGetRawAnilistMangaCollection()

    const data = React.useMemo(() => {
        if (type) {
            return type === "anime" ? animeData : mangaData
        }
        return selectedType === "anime" ? animeData : mangaData
    }, [selectedType, animeData, mangaData, type])

    const lists = React.useMemo(() => data?.MediaListCollection?.lists, [data])

    const [params, _setParams] = useAtom(__myListsSearch_paramsAtom)
    const [debouncedParams] = useDebounce(params, 500)

    React.useLayoutEffect(() => {
        if (selectedType === "manga" && !serverStatus?.settings?.library?.enableManga) {
            setSelectedType("anime")
        }
    }, [serverStatus?.settings?.library?.enableManga])

    React.useLayoutEffect(() => {
        _setParams(MYLISTS_DEFAULT_PARAMS)
    }, [selectedType])

    const _filteredLists: AL_AnimeCollection_MediaListCollection_Lists[] = React.useMemo(() => {
        return lists?.map(obj => {
            if (!obj) return undefined
            const arr = filterListEntries(selectedType as CollectionType, obj?.entries, params, serverStatus?.settings?.anilist?.enableAdultContent)
            return {
                name: obj?.name,
                isCustomList: obj?.isCustomList,
                status: obj?.status,
                entries: arr,
            }
        }).filter(Boolean) ?? []
    }, [lists, debouncedParams, selectedType, serverStatus?.settings?.anilist?.enableAdultContent])

    const filteredLists: AL_AnimeCollection_MediaListCollection_Lists[] = React.useMemo(() => {
        return _filteredLists?.map(obj => {
            if (!obj) return undefined
            const arr = filterEntriesByTitle(obj?.entries, debouncedSearchInput)
            return {
                name: obj?.name,
                isCustomList: obj?.isCustomList,
                status: obj?.status,
                entries: arr,
            }
        })?.filter(Boolean) ?? []
    }, [_filteredLists, debouncedSearchInput])

    const customLists = React.useMemo(() => {
        return filteredLists?.filter(obj => obj?.isCustomList) ?? []
    }, [filteredLists])

    return {
        currentList: React.useMemo(() => filteredLists?.find(l => l?.status === "CURRENT"), [filteredLists]),
        repeatingList: React.useMemo(() => filteredLists?.find(l => l?.status === "REPEATING"), [filteredLists]),
        planningList: React.useMemo(() => filteredLists?.find(l => l?.status === "PLANNING"), [filteredLists]),
        pausedList: React.useMemo(() => filteredLists?.find(l => l?.status === "PAUSED"), [filteredLists]),
        completedList: React.useMemo(() => filteredLists?.find(l => l?.status === "COMPLETED"), [filteredLists]),
        droppedList: React.useMemo(() => filteredLists?.find(l => l?.status === "DROPPED"), [filteredLists]),
        customLists,
    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
