import { useGetAnimeCollection } from "@/api/hooks/anilist.hooks"
import { anilistUserMediaAtom } from "@/app/(main)/_atoms/anilist.atoms"
import { useAtomValue, useSetAtom } from "jotai/react"
import { useEffect } from "react"

/**
 * @description
 * - Fetches the Anilist collection
 */
export function useAnimeCollectionLoader() {
    const setter = useSetAtom(anilistUserMediaAtom)

    const { data } = useGetAnimeCollection()

    // Store the user's media in `userMediaAtom`
    useEffect(() => {
        if (!!data) {
            const allMedia = data.MediaListCollection?.lists?.flatMap(n => n?.entries)?.filter(Boolean)?.map(n => n.media)?.filter(Boolean) ?? []
            setter(allMedia)
        }
    }, [data])

    return null
}

export function useAnilistUserMedia() {
    return useAtomValue(anilistUserMediaAtom)
}
