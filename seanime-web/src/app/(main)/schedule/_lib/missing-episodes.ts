import { MediaEntryMissingEpisodes, missingEpisodesAtom, missingSilencedEpisodesAtom } from "@/atoms/missing-episodes"
import { serverStatusAtom } from "@/atoms/server-status"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { useAtomValue } from "jotai"
import { useSetAtom } from "jotai/react"
import { useEffect } from "react"

export function useMissingEpisodes() {
    const serverStatus = useAtomValue(serverStatusAtom)
    const setAtom = useSetAtom(missingEpisodesAtom)
    const setSilencedAtom = useSetAtom(missingSilencedEpisodesAtom)

    const { data, isLoading, status } = useSeaQuery<MediaEntryMissingEpisodes>({
        endpoint: SeaEndpoints.MISSING_EPISODES,
        queryKey: ["get-missing-episodes"],
    })

    useEffect(() => {
        if (status === "success") {
            if (serverStatus?.settings?.anilist?.enableAdultContent) {
                setAtom(data?.episodes ?? [])
            } else {
                setAtom(data?.episodes?.filter(episode => !episode?.basicMedia?.isAdult) ?? [])
            }
            setSilencedAtom(data?.silencedEpisodes ?? [])
        }
    }, [data, serverStatus?.settings?.anilist?.enableAdultContent])

    return {
        missingEpisodes: serverStatus?.settings?.anilist?.enableAdultContent
            ? (data?.episodes ?? [])
            : (data?.episodes?.filter(episode => !episode?.basicMedia?.isAdult) ?? []),
        silencedEpisodes: data?.silencedEpisodes ?? [],
        isLoading: isLoading,
    }

}
