import { useSeaMutation } from "@/lib/server/queries/utils"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useCallback, useEffect } from "react"
import { usePathname, useRouter, useSearchParams } from "next/navigation"

export function useVideoPlayer() {

    const { mutate } = useSeaMutation<null, { path: string }>({
        endpoint: SeaEndpoints.PLAY_VIDEO,
    })

    const playVideo = useCallback(({ path }: { path: string }) => mutate({ path }), [])

    return {
        playVideo,
    }

}


export function usePlayNextVideoOnMount({ onPlay }: { onPlay: () => void }) {

    const router = useRouter()
    const pathname = usePathname()
    const searchParams = useSearchParams()
    const playNext = searchParams.get("playNext")
    const id = searchParams.get("id")

    useEffect(() => {
        // Automatically play the next episode if param is present in URL
        const t = setTimeout(() => {
            if (playNext) {
                router.replace(pathname + `?id=${id}`)
                onPlay()
            }
        }, 500)

        return () => clearTimeout(t)
    }, [pathname, id, playNext])

    return null

}