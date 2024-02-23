import { MediaEntryMissingEpisodes, missingEpisodesAtom, missingSilencedEpisodesAtom } from "@/atoms/missing-episodes"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { useSetAtom } from "jotai/react"
import { useEffect } from "react"

export function useMissingEpisodes() {

    const setAtom = useSetAtom(missingEpisodesAtom)
    const setSilencedAtom = useSetAtom(missingSilencedEpisodesAtom)

    const { data, isLoading, status } = useSeaQuery<MediaEntryMissingEpisodes>({
        endpoint: SeaEndpoints.MISSING_EPISODES,
        queryKey: ["get-missing-episodes"],
    })

    useEffect(() => {
        if (status === "success") {
            setAtom(data?.episodes ?? [])
            setSilencedAtom(data?.silencedEpisodes ?? [])
        }
    }, [data])

    return {
        missingEpisodes: data?.episodes ?? [],
        silencedEpisodes: data?.silencedEpisodes ?? [],
        isLoading: isLoading,
    }

}
