import { libraryCollectionAtom } from "@/app/(main)/_loaders/library-collection"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { LibraryCollection } from "@/lib/server/types"
import { useAtom } from "jotai/react"
import { useEffect, useMemo } from "react"

export function useLibraryCollection() {

    const [prev, setLibraryCollectionAtom] = useAtom(libraryCollectionAtom)

    const { data, isLoading } = useSeaQuery<LibraryCollection>({
        endpoint: SeaEndpoints.LIBRARY_COLLECTION,
        queryKey: ["get-library-collection"],
        placeholderData: prev,
    })

    // Store the received data in `libraryCollectionAtom`
    useEffect(() => {
        if (!!data) {
            setLibraryCollectionAtom(data)
        }
    }, [data])

    const sortedCollection = useMemo(() => {
        if (!data || !data.lists) return []
        return [
            data.lists.find(n => n.type === "current"),
            data.lists.find(n => n.type === "paused"),
            data.lists.find(n => n.type === "planned"),
            data.lists.find(n => n.type === "completed"),
            data.lists.find(n => n.type === "dropped"),
        ].filter(Boolean)
    }, [data])

    return {
        isLoading: isLoading,
        libraryCollectionList: sortedCollection,
        continueWatchingList: data?.continueWatchingList ?? [],
        unmatchedLocalFiles: data?.unmatchedLocalFiles ?? [],
        ignoredLocalFiles: data?.ignoredLocalFiles ?? [],
        unmatchedGroups: data?.unmatchedGroups ?? [],
        unknownGroups: data?.unknownGroups ?? [],
    }

}
