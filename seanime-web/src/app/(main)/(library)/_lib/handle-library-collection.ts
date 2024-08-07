import { useGetLibraryCollection } from "@/api/hooks/anime_collection.hooks"
import { libraryCollectionAtom } from "@/app/(main)/_atoms/anime-library-collection.atoms"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { CollectionParams, DEFAULT_COLLECTION_PARAMS, filterCollectionEntries } from "@/lib/helpers/filtering"
import { atomWithImmer } from "jotai-immer"
import { useAtom, useSetAtom } from "jotai/react"
import React from "react"

export const MAIN_LIBRARY_DEFAULT_PARAMS: CollectionParams = {
    ...DEFAULT_COLLECTION_PARAMS,
    sorting: "PROGRESS_DESC",
}

export const __mainLibrary_paramsAtom = atomWithImmer<CollectionParams>(MAIN_LIBRARY_DEFAULT_PARAMS)

export const __mainLibrary_paramsInputAtom = atomWithImmer<CollectionParams>(MAIN_LIBRARY_DEFAULT_PARAMS)

export function useHandleLibraryCollection() {
    const serverStatus = useServerStatus()

    const atom_setLibraryCollection = useSetAtom(libraryCollectionAtom)

    /**
     * Fetch the library collection data
     */
    const { data, isLoading } = useGetLibraryCollection()

    /**
     * Store the received data in `libraryCollectionAtom`
     */
    React.useEffect(() => {
        if (!!data) {
            atom_setLibraryCollection(data)
        }
    }, [data])

    const libraryGenres = React.useMemo(() => {
        const allGenres = data?.lists?.flatMap(l => {
            return l.entries?.flatMap(e => e.media?.genres) ?? []
        })
        return [...new Set(allGenres)].filter(Boolean)
    }, [data])

    const [params, setParams] = useAtom(__mainLibrary_paramsAtom)
    // const debouncedParams = useDebounce(params, 500)

    // Reset params when data changes
    React.useEffect(() => {
        if (!!data) {
            setParams(MAIN_LIBRARY_DEFAULT_PARAMS)
        }
    }, [data])

    /**
     * Sort and filter the collection data
     */
    const sortedCollection = React.useMemo(() => {
        if (!data || !data.lists) return []

        let _lists = data.lists.map(obj => {
            if (!obj) return obj
            const arr = filterCollectionEntries(obj.entries, MAIN_LIBRARY_DEFAULT_PARAMS, serverStatus?.settings?.anilist?.enableAdultContent)
            return {
                type: obj.type,
                status: obj.status,
                entries: arr,
            }
        })
        return [
            _lists.find(n => n.type === "current"),
            _lists.find(n => n.type === "paused"),
            _lists.find(n => n.type === "planned"),
            _lists.find(n => n.type === "completed"),
            _lists.find(n => n.type === "dropped"),
        ].filter(Boolean)
    }, [data, params, serverStatus?.settings?.anilist?.enableAdultContent])

    const filteredCollection = React.useMemo(() => {
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
            _lists.find(n => n.type === "current"),
            _lists.find(n => n.type === "paused"),
            _lists.find(n => n.type === "planned"),
            _lists.find(n => n.type === "completed"),
            _lists.find(n => n.type === "dropped"),
        ].filter(Boolean)
    }, [data, params, serverStatus?.settings?.anilist?.enableAdultContent])

    const continueWatchingList = React.useMemo(() => {
        if (!data?.continueWatchingList) return []

        if (!serverStatus?.settings?.anilist?.enableAdultContent || serverStatus?.settings?.anilist?.blurAdultContent) {
            return data.continueWatchingList.filter(entry => entry.baseMedia?.isAdult === false)
        }

        return data.continueWatchingList
    }, [
        data?.continueWatchingList,
        serverStatus?.settings?.anilist?.enableAdultContent,
        serverStatus?.settings?.anilist?.blurAdultContent,
    ])

    return {
        libraryGenres,
        isLoading: isLoading,
        libraryCollectionList: sortedCollection,
        filteredLibraryCollectionList: filteredCollection,
        continueWatchingList: continueWatchingList,
        unmatchedLocalFiles: data?.unmatchedLocalFiles ?? [],
        ignoredLocalFiles: data?.ignoredLocalFiles ?? [],
        unmatchedGroups: data?.unmatchedGroups ?? [],
        unknownGroups: data?.unknownGroups ?? [],
    }

}
