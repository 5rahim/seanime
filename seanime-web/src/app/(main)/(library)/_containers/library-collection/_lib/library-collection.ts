import { LibraryCollection } from "@/app/(main)/(library)/_lib/anime-library.types"
import { libraryCollectionAtom } from "@/app/(main)/_loaders/library-collection"
import { serverStatusAtom } from "@/atoms/server-status"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { useAtomValue, useSetAtom } from "jotai/react"
import React, { useEffect, useMemo } from "react"

export function useLibraryCollection() {
    const serverStatus = useAtomValue(serverStatusAtom)

    const setLibraryCollectionAtom = useSetAtom(libraryCollectionAtom)

    const { data, isLoading } = useSeaQuery<LibraryCollection>({
        endpoint: SeaEndpoints.LIBRARY_COLLECTION,
        queryKey: ["get-library-collection"],
    })

    // Store the received data in `libraryCollectionAtom`
    useEffect(() => {
        if (!!data) {
            setLibraryCollectionAtom(data)
        }
    }, [data])

    const sortedCollection = useMemo(() => {
        if (!data || !data.lists) return []

        let _lists = data.lists
        if (!serverStatus?.settings?.anilist?.enableAdultContent) {
            _lists = _lists.map(list => {
                list.entries = list.entries.filter(entry => !entry.media?.isAdult)
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
