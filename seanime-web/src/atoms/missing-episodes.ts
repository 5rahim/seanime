import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { MediaEntryEpisode } from "@/lib/server/types"
import { atom } from "jotai"
import { useAtomValue, useSetAtom } from "jotai/react"
import { usePathname } from "next/navigation"
import { useEffect } from "react"

export const missingEpisodesAtom = atom<MediaEntryEpisode[]>([])

const missingEpisodeCount = atom(get => get(missingEpisodesAtom).length)

export function useMissingEpisodeCount() {
    return useAtomValue(missingEpisodeCount)
}

/**
 * @description
 * - When the user is not on the main page, send a request to get missing episodes
 */
export function useListenToMissingEpisodes() {
    const pathname = usePathname()
    const setter = useSetAtom(missingEpisodesAtom)

    const { data } = useSeaQuery<MediaEntryEpisode[]>({
        endpoint: SeaEndpoints.MISSING_EPISODES,
        queryKey: ["get-missing-episodes"],
        enabled: pathname !== "/schedule",
    })

    useEffect(() => {
        setter(data ?? [])
    }, [data])

    return null
}
