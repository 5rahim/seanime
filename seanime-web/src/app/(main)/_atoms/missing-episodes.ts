import { Anime_MediaEntryEpisode } from "@/api/generated/types"
import { useGetMissingEpisodes } from "@/api/hooks/anime_entries.hooks"
import { atom } from "jotai"
import { useAtomValue, useSetAtom } from "jotai/react"
import { usePathname } from "next/navigation"
import { useEffect } from "react"

export const missingEpisodesAtom = atom<Anime_MediaEntryEpisode[]>([])
export const missingSilencedEpisodesAtom = atom<Anime_MediaEntryEpisode[]>([])

const missingEpisodeCount = atom(get => get(missingEpisodesAtom).length)

export function useMissingEpisodeCount() {
    return useAtomValue(missingEpisodeCount)
}

/**
 * @description
 * - When the user is not on the main page, send a request to get missing episodes
 */
export function useMissingEpisodeListener() {
    const pathname = usePathname()
    const setter = useSetAtom(missingEpisodesAtom)
    const silencedSetter = useSetAtom(missingSilencedEpisodesAtom)

    const { data } = useGetMissingEpisodes(pathname !== "/schedule")

    useEffect(() => {
        setter(data?.episodes ?? [])
        silencedSetter(data?.silencedEpisodes ?? [])
    }, [data])

    return null
}
