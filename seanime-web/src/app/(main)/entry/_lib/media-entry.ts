import { BasicMediaFragment, MediaDetailsByIdQuery } from "@/lib/anilist/gql/graphql"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation, useSeaQuery } from "@/lib/server/query"
import { MediaEntry } from "@/lib/server/types"

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
