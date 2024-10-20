import { useGetLibraryCollection } from "@/api/hooks/anime_collection.hooks"
import { animeLibraryCollectionAtom } from "@/app/(main)/_atoms/anime-library-collection.atoms"
import { useAtomValue, useSetAtom } from "jotai/react"
import React from "react"

/**
 * @description
 * - Fetches the library collection and sets it in the atom
 */
export function useAnimeLibraryCollectionLoader() {

    const setter = useSetAtom(animeLibraryCollectionAtom)

    const { data, status } = useGetLibraryCollection()

    React.useEffect(() => {
        if (status === "success") {
            setter(data)
        }
    }, [data, status])

    return null
}

export function useLibraryCollection() {
    return useAtomValue(animeLibraryCollectionAtom)
}
