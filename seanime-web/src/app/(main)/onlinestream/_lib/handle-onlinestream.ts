import { VideoCore_OnlinestreamParams } from "@/api/generated/types"
import { useNakamaStatus } from "@/app/(main)/_features/nakama/nakama-manager"
import { __anime_entryPageViewAtom } from "@/app/(main)/entry/_containers/anime-entry-page"
import { logger, useLatestFunction } from "@/lib/helpers/debug"
import { usePathname, useRouter, useSearchParams } from "@/lib/navigation"
import { atom } from "jotai"
import { useAtom, useSetAtom } from "jotai/react"
import React from "react"
import { toast } from "sonner"

export type OnlineStreamParams = {
    mediaId: number
    provider: string
    episodeNumber: number
    server: string
    quality: string
    dubbed: boolean
}

const __onlinestream_streamToLoadAtom = atom<OnlineStreamParams | null>(null)

export function useNakamaOnlineStreamWatchParty() {
    const [streamToLoad, setStreamToLoad] = useAtom(__onlinestream_streamToLoadAtom)
    const pathname = usePathname()
    const searchParams = useSearchParams()
    const router = useRouter()
    const nakamaStatus = useNakamaStatus()

    const setCurrentView = useSetAtom(__anime_entryPageViewAtom)

    const redirectToStream = useLatestFunction((params: OnlineStreamParams) => {
        router.push("/entry?id=" + params.mediaId + "&tab=onlinestream")
        setCurrentView("onlinestream")
    })

    const startOnlineStreamWatchParty = useLatestFunction((params: VideoCore_OnlinestreamParams) => {
        if (nakamaStatus?.isHost) {
            logger("ONLINESTREAM").info("Start online stream event sent to peers", params)
            return
        }
        logger("ONLINESTREAM").info("Starting online stream watch party", params)
        toast.info("Starting online streaming watch party", { duration: 2000 })
        redirectToStream(params)
        React.startTransition(() => {
            setStreamToLoad(params)
        })
    })

    const onLoadedStream = useLatestFunction(() => {
        setStreamToLoad(null)
    })

    const removeParamsFromUrl = useLatestFunction(() => {
        const params = new URLSearchParams(searchParams.toString())
        params.delete("tab")
        params.delete("provider")
        params.delete("episodeNumber")
        params.delete("server")
        params.delete("quality")
        params.delete("dubbed")
        const newUrl = pathname + (params.toString() ? `?${params.toString()}` : "")
        router.replace(newUrl, { scroll: false })
    })

    return {
        // Peers
        streamToLoad,
        redirectToStream,
        onLoadedStream,
        startOnlineStreamWatchParty,
        removeParamsFromUrl,
    }
}
