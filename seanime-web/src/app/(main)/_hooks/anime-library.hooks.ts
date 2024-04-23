import { useGetLibraryCollection } from "@/api/hooks/anime_collection.hooks"
import { libraryCollectionAtom } from "@/app/(main)/_atoms/anime-library-collection.atoms"
import { useAtomValue, useSetAtom } from "jotai/react"
import { useEffect } from "react"

/**
 * @description
 * - Fetches the library collection and sets it in the atom
 */
export function useLibraryCollectionLoader() {

    const setter = useSetAtom(libraryCollectionAtom)

    const { data, status } = useGetLibraryCollection()

    useEffect(() => {
        if (status === "success") {
            setter(data)
        }
    }, [data, status])

    return null
}

export function useLibraryCollection() {
    return useAtomValue(libraryCollectionAtom)
}
