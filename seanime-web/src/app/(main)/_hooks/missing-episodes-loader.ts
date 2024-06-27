import { useGetMissingEpisodes } from "@/api/hooks/anime_entries.hooks"
import { missingEpisodeCountAtom, missingEpisodesAtom, missingSilencedEpisodesAtom } from "@/app/(main)/_atoms/missing-episodes.atoms"
import { useAtomValue, useSetAtom } from "jotai/react"
import { usePathname } from "next/navigation"
import { useEffect } from "react"

export function useMissingEpisodeCount() {
    return useAtomValue(missingEpisodeCountAtom)
}

export function useMissingEpisodes() {
    return useAtomValue(missingEpisodesAtom)
}

/**
 * @description
 * - When the user is not on the main page, send a request to get missing episodes
 */
export function useMissingEpisodesLoader() {
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
