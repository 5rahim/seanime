"use client"
import { AnimeCollectionQuery, BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/queries/utils"
import { LibraryCollection } from "@/lib/server/types"
import { atom } from "jotai"
import { useSetAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import { useEffect } from "react"


export const libraryCollectionAtom = atomWithStorage<LibraryCollection | undefined>("sea-library-collection", undefined,
    undefined, {getOnInit: true})

export const userMediaAtom = atomWithStorage<BaseMediaFragment[] | undefined>("sea-user-media", undefined,
    undefined, { getOnInit: true })

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

/**
 * @description
 * - When the user is not on the main page, send a request to get the Anilist collection
 */
export function useListenToUserMedia() {
    const setter = useSetAtom(userMediaAtom)

    const { data, isLoading } = useSeaQuery<AnimeCollectionQuery>({
        endpoint: SeaEndpoints.ANILIST_COLLECTION,
        queryKey: ["get-anilist-collection"],
    })

    // Store the received data in `userMediaAtom`
    useEffect(() => {
        if (!!data) {
            const allMedia = data.MediaListCollection?.lists?.flatMap(n => n?.entries)?.filter(Boolean)?.map(n => n.media)?.filter(Boolean) ?? []
            setter(allMedia)
        }
    }, [data])

    return null
}
