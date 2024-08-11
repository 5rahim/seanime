import { Anime_LibraryCollectionList } from "@/api/generated/types"
import { useGetLibraryCollection } from "@/api/hooks/anime_collection.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useDebounce } from "@/hooks/use-debounce"
import { CollectionParams, DEFAULT_COLLECTION_PARAMS, filterCollectionEntries, filterEntriesByTitle } from "@/lib/helpers/filtering"
import { atomWithImmer } from "jotai-immer"
import { useAtomValue } from "jotai/index"
import React from "react"

export const DETAILED_LIBRARY_DEFAULT_PARAMS: CollectionParams = {
    ...DEFAULT_COLLECTION_PARAMS,
    sorting: "PROGRESS_DESC",
}

export const __library_paramsAtom = atomWithImmer<CollectionParams>(DETAILED_LIBRARY_DEFAULT_PARAMS)

export const __library_paramsInputAtom = atomWithImmer<CollectionParams>(DETAILED_LIBRARY_DEFAULT_PARAMS)

export const __library_selectedListAtom = atomWithImmer<string>("-")

export const __library_debouncedSearchInputAtom = atomWithImmer<string>("")

export function useHandleDetailedLibraryCollection() {
    const serverStatus = useServerStatus()

    /**
     * Fetch the library collection data
     */
    const { data, isLoading } = useGetLibraryCollection()

    const params = useAtomValue(__library_paramsAtom)
    const debouncedParams = useDebounce(params, 500)

    const debouncedSearchInput = useAtomValue(__library_debouncedSearchInputAtom)

    /**
     * Sort and filter the collection data
     */
    const _sortedCollection: Anime_LibraryCollectionList[] = React.useMemo(() => {
        if (!data || !data.lists) return []

        let _lists = data.lists.map(obj => {
            if (!obj) return obj
            const arr = filterCollectionEntries(obj.entries, params, serverStatus?.settings?.anilist?.enableAdultContent)
            return {
                type: obj.type,
                status: obj.status,
                entries: arr,
            }
        })
        return [
            _lists.find(n => n.type === "CURRENT"),
            _lists.find(n => n.type === "PAUSED"),
            _lists.find(n => n.type === "PLANNING"),
            _lists.find(n => n.type === "COMPLETED"),
            _lists.find(n => n.type === "DROPPED"),
        ].filter(Boolean)
    }, [data, debouncedParams, serverStatus?.settings?.anilist?.enableAdultContent])

    const sortedCollection: Anime_LibraryCollectionList[] = React.useMemo(() => {
        return _sortedCollection.map(obj => {
            if (!obj) return obj
            const arr = filterEntriesByTitle(obj.entries, debouncedSearchInput)
            return {
                type: obj.type,
                status: obj.status,
                entries: arr,
            }
        }).filter(Boolean)
    }, [_sortedCollection, debouncedSearchInput])

    const continueWatchingList = React.useMemo(() => {
        if (!data?.continueWatchingList) return []

        if (!serverStatus?.settings?.anilist?.enableAdultContent || serverStatus?.settings?.anilist?.blurAdultContent) {
            return data.continueWatchingList.filter(entry => entry.baseAnime?.isAdult === false)
        }

        return data.continueWatchingList
    }, [
        data?.continueWatchingList,
        serverStatus?.settings?.anilist?.enableAdultContent,
        serverStatus?.settings?.anilist?.blurAdultContent,
    ])

    return {
        isLoading: isLoading,
        stats: data?.stats,
        libraryCollectionList: sortedCollection,
        continueWatchingList: continueWatchingList,
        unmatchedLocalFiles: data?.unmatchedLocalFiles ?? [],
        ignoredLocalFiles: data?.ignoredLocalFiles ?? [],
        unmatchedGroups: data?.unmatchedGroups ?? [],
        unknownGroups: data?.unknownGroups ?? [],
    }

}
