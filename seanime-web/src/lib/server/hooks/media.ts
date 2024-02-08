import { userMediaAtom } from "@/atoms/collection"
import { useWebsocketMessageListener } from "@/atoms/websocket"
import { AnimeCollectionQuery, BasicMediaFragment, MediaDetailsByIdQuery } from "@/lib/anilist/gql/graphql"
import { SeaEndpoints, WSEvents } from "@/lib/server/endpoints"
import { useSeaMutation, useSeaQuery } from "@/lib/server/queries/utils"
import { LocalFile, MediaEntry } from "@/lib/server/types"
import { useQueryClient } from "@tanstack/react-query"
import { useSetAtom } from "jotai/react"
import toast from "react-hot-toast"


/**
 * @description
 * Listens to REFRESHED_ANILIST_COLLECTION events and re-fetches queries associated with AniList collection.
 */
export function useAnilistCollectionListener() {

    const qc = useQueryClient()

    useWebsocketMessageListener<AnimeCollectionQuery>({
        type: WSEvents.REFRESHED_ANILIST_COLLECTION,
        onMessage: data => {
            (async () => {
                await qc.refetchQueries({ queryKey: ["get-library-collection"] })
                await qc.refetchQueries({ queryKey: ["get-anilist-collection"] })
                await qc.refetchQueries({ queryKey: ["get-missing-episodes"] })
            })()
        },
    })

}

/**
 * @description
 * Fetches the MediaEntry associated with the ID
 * @param _mId
 */
export function useMediaEntry(_mId: string | number | null) {

    const mId = typeof _mId === "string" ? Number(_mId) : _mId

    const { data, isLoading } = useSeaQuery<MediaEntry, { mediaId: number }>({
        endpoint: SeaEndpoints.MEDIA_ENTRY.replace("{id}", String(mId)),
        queryKey: ["get-media-entry", mId],
        enabled: !!mId,
    })

    return {
        mediaEntry: data,
        mediaEntryLoading: isLoading,
    }

}

/**
 * @description
 * Fetches the SimpleMediaEntry associated with the ID
 * @param _mId
 */
export function useSimpleMediaEntry(_mId: string | number | null) {

    const mId = typeof _mId === "string" ? Number(_mId) : _mId

    const { data, isLoading } = useSeaQuery<MediaEntry, { mediaId: number }>({
        endpoint: SeaEndpoints.SIMPLE_MEDIA_ENTRY.replace("{id}", String(mId)),
        queryKey: ["get-simple-media-entry", mId],
        enabled: !!mId,
    })

    return {
        mediaEntry: data,
        mediaEntryLoading: isLoading,
    }

}

export function useLatestAnilistCollection() {

    const { data, isLoading } = useSeaQuery<AnimeCollectionQuery>({
        endpoint: SeaEndpoints.ANILIST_COLLECTION,
        queryKey: ["get-anilist-collection"],
        method: "post",
    })

    return {
        anilistLists: data?.MediaListCollection?.lists ?? [],
        isLoading,
    }

}

/**
 * @description
 * Fetches the (cached) AniList collection
 */
export function useAnilistCollection() {

    const setUserMedia = useSetAtom(userMediaAtom)

    const { data, isLoading } = useSeaQuery<AnimeCollectionQuery>({
        endpoint: SeaEndpoints.ANILIST_COLLECTION,
        queryKey: ["get-anilist-collection"],
    })


    return {
        anilistLists: data?.MediaListCollection?.lists ?? [],
        isLoading,
    }

}

/**
 * @description
 * - Asks the server to fetch an up-to-date version of the user's AniList collection.
 * - When the request succeeds, we refetch queries related to the AniList collection.
 */
export function useRefreshAnilistCollection() {

    const qc = useQueryClient()

    // Return data is ignored
    const { mutate, isPending } = useSeaMutation<AnimeCollectionQuery>({
        endpoint: SeaEndpoints.ANILIST_COLLECTION,
        mutationKey: ["refresh-anilist-collection"],
        onSuccess: async () => {
            // Refetch library collection
            toast.success("AniList is up-to-date")
            await qc.refetchQueries({ queryKey: ["get-library-collection"] })
            await qc.refetchQueries({ queryKey: ["get-anilist-collection"] })
            await qc.refetchQueries({ queryKey: ["get-missing-episodes"] })
        },
    })

    return {
        refreshAnilistCollection: mutate,
        isPending,
    }

}

/**
 * @description
 * Get MediaDetails (genre, recommendations...) associated with an AniList Media.
 * @param _mId
 */
export function useMediaDetails(_mId: string | number | null) {

    const mId = typeof _mId === "string" ? Number(_mId) : _mId

    const { data, isLoading } = useSeaQuery<MediaDetailsByIdQuery["Media"], { mediaId: number }>({
        endpoint: SeaEndpoints.ANILIST_MEDIA_DETAILS.replace("{id}", String(mId)),
        queryKey: ["get-anilist-media-details", mId],
        enabled: !!mId,
    })

    return {
        mediaDetails: data,
        mediaDetailsLoading: isLoading,
    }

}

/**
 * @description
 * - Used by the "Unmatched file manager"
 * - Fetches AniList Media suggestions based on the files located in the specified directory.
 */
export function useFetchMediaEntrySuggestions() {

    const { mutate, data, isPending, reset } = useSeaMutation<BasicMediaFragment[], { dir: string }>({
        endpoint: SeaEndpoints.MEDIA_ENTRY_SUGGESTIONS,
        mutationKey: ["media-entry-suggestions"],
    })

    return {
        fetchSuggestions: (dir: string) => mutate({
            dir: dir,
        }),
        suggestions: data ?? [],
        isPending,
        resetSuggestions: reset,
    }

}

export function useManuallyMatchLocalFiles() {

    const qc = useQueryClient()

    type Props = { dir: string, mediaId: number }

    // Return data is ignored
    const { mutate, isPending } = useSeaMutation<LocalFile[], Props>({
        endpoint: SeaEndpoints.MEDIA_ENTRY_MANUAL_MATCH,
        mutationKey: ["media-entry-manual-match"],
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["get-library-collection"] })
        },
    })

    return {
        manuallyMatchEntry: (props: Props, callback: () => void) => {
            mutate(props, {
                onSuccess: async () => {
                    if (props.mediaId) {
                        await qc.refetchQueries({ queryKey: ["get-media-entry", props.mediaId] })
                    }
                    callback()
                },
            })
        },
        isPending,
    }

}
