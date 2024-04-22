import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"

export function useOpenMediaEntryInExplorer() {

    const { mutate } = useSeaMutation<boolean, { mediaId: number }>({
        endpoint: SeaEndpoints.OPEN_ANIME_ENTRY_IN_EXPLORER,
        mutationKey: ["open-media-entry-in-explorer"],
    })

    return {
        openEntryInExplorer: (mediaId: number) => mutate({
            mediaId: mediaId,
        }),
    }

}

export function useOpenInExplorer() {

    const { mutate } = useSeaMutation<boolean, { path: string }>({
        endpoint: SeaEndpoints.OPEN_IN_EXPLORER,
        mutationKey: ["open-in-explorer"],
    })

    return {
        openInExplorer: (path: string) => mutate({
            path: path,
        }),
    }

}
