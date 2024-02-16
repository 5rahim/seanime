import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { LibraryCollection } from "@/lib/server/types"
import { atom } from "jotai/index"
import { useSetAtom } from "jotai/react"
import { useEffect } from "react"

export const libraryCollectionAtom = atom<LibraryCollection | undefined>(undefined)

export const getAtomicLibraryEntryAtom = atom(get => get(libraryCollectionAtom),
    (get, set, payload: number) => {
        const lists = get(libraryCollectionAtom)?.lists
        if (!lists) {
            return undefined
        }
        return lists.flatMap(n => n.entries).find(n => n.mediaId === payload)
    },
)

/**
 * @description
 * - Top level hook for fetching the LibraryCollection
 * - Sends a request for the LibraryCollection and updates `libraryCollectionAtom`
 */
export function useLibraryCollectionLoader() {

    const setter = useSetAtom(libraryCollectionAtom)

    const { data, status } = useSeaQuery<LibraryCollection>({
        endpoint: SeaEndpoints.LIBRARY_COLLECTION,
        queryKey: ["get-library-collection"],
    })

    useEffect(() => {
        if (status === "success") {
            setter(data)
        }
    }, [data, status])

    return null
}
