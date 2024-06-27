import { useGetLibraryCollection } from "@/api/hooks/anime_collection.hooks"
import { libraryCollectionAtom } from "@/app/(main)/_atoms/anime-library-collection.atoms"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useSetAtom } from "jotai/react"
import React, { useEffect, useMemo } from "react"

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
    useEffect(() => {
        if (!!data) {
            atom_setLibraryCollection(data)
        }
    }, [data])

    /**
     * Sort and filter the collection data
     */
    const sortedCollection = useMemo(() => {
        if (!data || !data.lists) return []

        let _lists = data.lists
        if (!serverStatus?.settings?.anilist?.enableAdultContent) {
            _lists = _lists.map(list => {
                list.entries = list.entries?.filter(entry => !entry.media?.isAdult) ?? []
                return list
            })
        }

        return [
            _lists.find(n => n.type === "current"),
            _lists.find(n => n.type === "paused"),
            _lists.find(n => n.type === "planned"),
            _lists.find(n => n.type === "completed"),
            _lists.find(n => n.type === "dropped"),
        ].filter(Boolean)
    }, [data, serverStatus?.settings?.anilist?.enableAdultContent])

    const continueWatchingList = React.useMemo(() => {
        if (!data?.continueWatchingList) return []

        if (!serverStatus?.settings?.anilist?.enableAdultContent || serverStatus?.settings?.anilist?.blurAdultContent) {
            return data.continueWatchingList.filter(entry => entry.basicMedia?.isAdult === false)
        }

        return data.continueWatchingList
    }, [
        data?.continueWatchingList,
        serverStatus?.settings?.anilist?.enableAdultContent,
        serverStatus?.settings?.anilist?.blurAdultContent,
    ])

    return {
        isLoading: isLoading,
        libraryCollectionList: sortedCollection,
        continueWatchingList: continueWatchingList,
        unmatchedLocalFiles: data?.unmatchedLocalFiles ?? [],
        ignoredLocalFiles: data?.ignoredLocalFiles ?? [],
        unmatchedGroups: data?.unmatchedGroups ?? [],
        unknownGroups: data?.unknownGroups ?? [],
    }

}
