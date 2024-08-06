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

        // Stream
        if (data.stream) {
            // Add to current list
            let currentList = data.lists.find(n => n.type === "CURRENT")
            if (currentList) {
                let entries = [...(currentList.entries ?? [])]
                for (let anime of (data.stream.anime ?? [])) {
                    if (!entries.some(e => e.mediaId === anime.id)) {
                        entries.push({
                            media: anime,
                            mediaId: anime.id,
                            listData: data.stream.listData?.[anime.id],
                        })
                    }
                }
                data.lists.find(n => n.type === "CURRENT")!.entries = entries
            }
        }

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
            _lists.find(n => n.type === "CURRENT"),
            _lists.find(n => n.type === "PAUSED"),
            _lists.find(n => n.type === "PLANNING"),
            _lists.find(n => n.type === "COMPLETED"),
            _lists.find(n => n.type === "DROPPED"),
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
            _lists.find(n => n.type === "CURRENT"),
            _lists.find(n => n.type === "PAUSED"),
            _lists.find(n => n.type === "PLANNING"),
            _lists.find(n => n.type === "COMPLETED"),
            _lists.find(n => n.type === "DROPPED"),
        ].filter(Boolean)
    }, [data, params, serverStatus?.settings?.anilist?.enableAdultContent])

    const continueWatchingList = React.useMemo(() => {
        if (!data?.continueWatchingList) return []

        let list = [...data.continueWatchingList]
        if (data.stream) {
            for (let entry of (data.stream.continueWatchingList ?? [])) {
                list = [...list, entry]
            }
        }

        if (!serverStatus?.settings?.anilist?.enableAdultContent || serverStatus?.settings?.anilist?.blurAdultContent) {
            return list.filter(entry => entry.baseAnime?.isAdult === false)
        }

        return list?.sort((a, b) => b.episodeNumber - a.episodeNumber)
    }, [
        data?.stream,
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
