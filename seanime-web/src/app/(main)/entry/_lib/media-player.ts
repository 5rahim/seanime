import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { usePathname, useRouter, useSearchParams } from "next/navigation"
import { useCallback, useEffect } from "react"

export function useMediaPlayer() {

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

export function useOpenDefaultMediaPlayer() {

    const { mutate } = useSeaMutation({
        endpoint: SeaEndpoints.START_MEDIA_PLAYER,
        mutationKey: ["open-default-media-player"],
    })

    return {
        startDefaultMediaPlayer: () => mutate(),
    }

}
