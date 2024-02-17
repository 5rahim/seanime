import { missingEpisodesAtom } from "@/atoms/missing-episodes"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { MediaEntryEpisode } from "@/lib/server/types"
import { useSetAtom } from "jotai/react"
import { useEffect } from "react"

export function useMissingEpisodes() {

    const setAtom = useSetAtom(missingEpisodesAtom)

    const { data, isLoading, status } = useSeaQuery<MediaEntryEpisode[]>({
        endpoint: SeaEndpoints.MISSING_EPISODES,
        queryKey: ["get-missing-episodes"],
    })

    useEffect(() => {
        if (status === "success") {
            setAtom(data ?? [])
        }
    }, [data])

    return {
        missingEpisodes: data ?? [],
        isLoading: isLoading,
    }

}
