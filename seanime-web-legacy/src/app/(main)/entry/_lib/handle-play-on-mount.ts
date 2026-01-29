import { usePlayNext } from "@/app/(main)/_atoms/playback.atoms"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { usePathname, useRouter, useSearchParams } from "next/navigation"
import React from "react"

export function usePlayNextVideoOnMount({ onPlay }: { onPlay: () => void }, enabled: boolean = true) {

    const serverStatus = useServerStatus()
    const router = useRouter()
    const pathname = usePathname()
    const searchParams = useSearchParams()

    const { playNext, resetPlayNext } = usePlayNext()

    React.useEffect(() => {
        if (!enabled) return

        // Automatically play the next episode if param is present in URL
        const t = setTimeout(() => {
            if (playNext) {
                resetPlayNext()
                onPlay()
            }
        }, 500)

        return () => clearTimeout(t)
    }, [pathname, playNext, serverStatus, onPlay, enabled])

    return null

}
