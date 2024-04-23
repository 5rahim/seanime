import { useGetAnilistCollection } from "@/api/hooks/anilist.hooks"
import { anilistUserMediaAtom } from "@/app/(main)/_atoms/anilist.atoms"
import { useSetAtom } from "jotai/react"
import { useEffect } from "react"

/**
 * @description
 * - Fetches the Anilist collection
 */
export function useAnilistCollectionLoader() {
    const setter = useSetAtom(anilistUserMediaAtom)

    const { data } = useGetAnilistCollection()

    // Store the user's media in `userMediaAtom`
    useEffect(() => {
        if (!!data) {
            const allMedia = data.MediaListCollection?.lists?.flatMap(n => n?.entries)?.filter(Boolean)?.map(n => n.media)?.filter(Boolean) ?? []
            setter(allMedia)
        }
    }, [data])

    return null
}
