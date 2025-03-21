import { useGetLibraryCollection } from "@/api/hooks/anime_collection.hooks"
import { useGetContinuityWatchHistory } from "@/api/hooks/continuity.hooks"
import { animeLibraryCollectionAtom } from "@/app/(main)/_atoms/anime-library-collection.atoms"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import {
    CollectionParams,
    DEFAULT_ANIME_COLLECTION_PARAMS,
    filterAnimeCollectionEntries,
    sortContinueWatchingEntries,
} from "@/lib/helpers/filtering"
import { useThemeSettings } from "@/lib/theme/hooks"
import { atomWithImmer } from "jotai-immer"
import { useAtom, useSetAtom } from "jotai/react"
import React from "react"

export const MAIN_LIBRARY_DEFAULT_PARAMS: CollectionParams<"anime"> = {
    ...DEFAULT_ANIME_COLLECTION_PARAMS,
    sorting: "TITLE", // Will be set to default sorting on mount
    continueWatchingOnly: false,
}

export const __mainLibrary_paramsAtom = atomWithImmer<CollectionParams<"anime">>(MAIN_LIBRARY_DEFAULT_PARAMS)

export const __mainLibrary_paramsInputAtom = atomWithImmer<CollectionParams<"anime">>(MAIN_LIBRARY_DEFAULT_PARAMS)

export function useHandleLibraryCollection() {
    const serverStatus = useServerStatus()

    const atom_setLibraryCollection = useSetAtom(animeLibraryCollectionAtom)

    const { animeLibraryCollectionDefaultSorting, continueWatchingDefaultSorting } = useThemeSettings()

    const { data: watchHistory } = useGetContinuityWatchHistory()

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
        return [...new Set(allGenres)].filter(Boolean)?.sort((a, b) => a.localeCompare(b))
    }, [data])

    const [params, setParams] = useAtom(__mainLibrary_paramsAtom)
    // const debouncedParams = useDebounce(params, 500)

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
                            libraryData: undefined,
                        })
                    }
                }
                data.lists.find(n => n.type === "CURRENT")!.entries = entries
            }
        }

        let _lists = data.lists.map(obj => {
            if (!obj) return obj

            const newParams = { ...params, sorting: animeLibraryCollectionDefaultSorting as any } as CollectionParams<"anime">
            let arr = filterAnimeCollectionEntries(obj.entries,
                newParams,
                serverStatus?.settings?.anilist?.enableAdultContent,
                data.continueWatchingList,
                watchHistory)

            // Reset `continueWatchingOnly` to false if it's about to make the list disappear
            if (arr.length === 0 && newParams.continueWatchingOnly) {
                const newParams = { ...params, continueWatchingOnly: false, sorting: animeLibraryCollectionDefaultSorting } as CollectionParams<"anime">
                arr = filterAnimeCollectionEntries(obj.entries,
                    newParams,
                    serverStatus?.settings?.anilist?.enableAdultContent,
                    data.continueWatchingList,
                    watchHistory)
            }

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
    }, [data, params, animeLibraryCollectionDefaultSorting, serverStatus?.settings?.anilist?.enableAdultContent])

    const filteredCollection = React.useMemo(() => {
        if (!data || !data.lists) return []

        let _lists = data.lists.map(obj => {
            if (!obj) return obj
            const newParams = { ...params, sorting: animeLibraryCollectionDefaultSorting } as CollectionParams<"anime">
            const arr = filterAnimeCollectionEntries(obj.entries,
                newParams,
                serverStatus?.settings?.anilist?.enableAdultContent,
                data.continueWatchingList,
                watchHistory)
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
    }, [data, params, serverStatus?.settings?.anilist?.enableAdultContent, watchHistory])

    const continueWatchingList = React.useMemo(() => {
        if (!data?.continueWatchingList) return []

        let list = [...data.continueWatchingList]


        if (data.stream) {
            for (let entry of (data.stream.continueWatchingList ?? [])) {
                list = [...list, entry]
            }
        }

        const entries = sortedCollection.flatMap(n => n.entries)

        list = sortContinueWatchingEntries(list, continueWatchingDefaultSorting as any, entries, watchHistory)

        if (!serverStatus?.settings?.anilist?.enableAdultContent || serverStatus?.settings?.anilist?.blurAdultContent) {
            return list.filter(entry => entry.baseAnime?.isAdult === false)
        }

        return list
    }, [
        data?.stream,
        sortedCollection,
        data?.continueWatchingList,
        continueWatchingDefaultSorting,
        serverStatus?.settings?.anilist?.enableAdultContent,
        serverStatus?.settings?.anilist?.blurAdultContent,
        watchHistory,
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
