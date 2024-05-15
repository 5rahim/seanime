import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { usePathname, useRouter, useSearchParams } from "next/navigation"
import React from "react"

export function usePlayNextVideoOnMount({ onPlay }: { onPlay: () => void }) {

    const serverStatus = useServerStatus()
    const router = useRouter()
    const pathname = usePathname()
    const searchParams = useSearchParams()
    const playNext = searchParams.get("playNext")
    const id = searchParams.get("id")

    React.useEffect(() => {
        // Automatically play the next episode if param is present in URL
        const t = setTimeout(() => {
            if (playNext) {
                router.replace(pathname + `?id=${id}`)
                onPlay()
            }
        }, 500)

        return () => clearTimeout(t)
    }, [pathname, id, playNext, serverStatus, onPlay])

    return null

}
