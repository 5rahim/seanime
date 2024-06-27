import { AL_AnimeCollection_MediaListCollection_Lists } from "@/api/generated/types"
import { useGetRawAnimeCollection } from "@/api/hooks/anilist.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { CollectionParams, DEFAULT_COLLECTION_PARAMS, filterEntriesByTitle, filterListEntries } from "@/lib/helpers/filtering"
import { useAtomValue } from "jotai"
import { atomWithImmer } from "jotai-immer"
import React from "react"
import { useDebounce } from "use-debounce"

export const __myListsSearch_paramsAtom = atomWithImmer<CollectionParams>({
    ...DEFAULT_COLLECTION_PARAMS,
    sorting: "SCORE_DESC",
})

export const __myListsSearch_paramsInputAtom = atomWithImmer<CollectionParams>({
    ...DEFAULT_COLLECTION_PARAMS,
    sorting: "SCORE_DESC",
})

export function useHandleUserAnilistLists(debouncedSearchInput: string) {

    const serverStatus = useServerStatus()
    const { data } = useGetRawAnimeCollection()

    const lists = React.useMemo(() => data?.MediaListCollection?.lists, [data])

    const params = useAtomValue(__myListsSearch_paramsAtom)
    const [debouncedParams] = useDebounce(params, 500)

    const _filteredLists: AL_AnimeCollection_MediaListCollection_Lists[] = React.useMemo(() => {
        return lists?.map(obj => {
            if (!obj) return undefined
            const arr = filterListEntries(obj?.entries, params, serverStatus?.settings?.anilist?.enableAdultContent)
            return {
                name: obj?.name,
                isCustomList: obj?.isCustomList,
                status: obj?.status,
                entries: arr,
            }
        }).filter(Boolean) ?? []
    }, [lists, debouncedParams, serverStatus?.settings?.anilist?.enableAdultContent])

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
        planningList: React.useMemo(() => filteredLists?.find(l => l?.status === "PLANNING"), [filteredLists]),
        pausedList: React.useMemo(() => filteredLists?.find(l => l?.status === "PAUSED"), [filteredLists]),
        completedList: React.useMemo(() => filteredLists?.find(l => l?.status === "COMPLETED"), [filteredLists]),
        droppedList: React.useMemo(() => filteredLists?.find(l => l?.status === "DROPPED"), [filteredLists]),
        customLists,
    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
