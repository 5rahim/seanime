import { atom } from "jotai"
import { LibraryCollection } from "@/lib/server/types"
import { atomWithStorage } from "jotai/utils"
import { useSetAtom } from "jotai/react"
import { useEffect } from "react"
import { useSeaQuery } from "@/lib/server/queries/utils"
import { SeaEndpoints } from "@/lib/server/endpoints"


export const libraryCollectionAtom = atomWithStorage<LibraryCollection | undefined>("sea-library-collection", undefined, undefined, { unstable_getOnInit: true })

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
 * Sends a request for the LibraryCollection and updates `libraryCollectionAtom`
 */
export function useAtomicLibraryCollectionLoader() {

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