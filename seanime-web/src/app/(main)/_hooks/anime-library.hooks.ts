import { LibraryCollection } from "@/app/(main)/(library)/_lib/anime-library.types"
import { libraryCollectionAtom } from "@/app/(main)/_atoms/anime-library-collection.atoms"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { useSetAtom } from "jotai/react"
import { useEffect } from "react"

/**
 * @description
 * - Fetches the library collection and sets it in the atom
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
