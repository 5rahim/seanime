import { usePlaybackPlayVideo } from "@/api/hooks/playback_manager.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useMediastreamCurrentFile, useMediastreamMediaToTranscode } from "@/app/(main)/mediastream/_lib/mediastream.atoms"
import { isMobile, isPs4, isTv, isXbox } from "@/lib/utils/browser-detection"
import { useRouter } from "next/navigation"
import React from "react"
import { toast } from "sonner"

export function useHandlePlayMedia() {
    const router = useRouter()
    const serverStatus = useServerStatus()

    const { mediaToTranscode } = useMediastreamMediaToTranscode()

    const { mutate: playVideo } = usePlaybackPlayVideo()

    const { setFilePath } = useMediastreamCurrentFile()

    function playMediaFile({ path, mediaId }: { path: string, mediaId: number }) {

        if (serverStatus?.mediastreamSettings?.transcodeEnabled && mediaToTranscode.includes(String(mediaId))) {
            setFilePath(path)
            React.startTransition(() => {
                router.push(`/mediastream?id=${mediaId}`)
            })
            return
        }

        if (isMobile() || isTv() || isPs4() || isXbox()) { // TODO: Find a way to override this
            if (serverStatus?.mediastreamSettings?.transcodeEnabled) {
                setFilePath(path)
                React.startTransition(() => {
                    router.push(`/mediastream?id=${mediaId}`)
                })
            } else {
                toast.error("Playback is not supported on this device.")
            }
            return
        } else {
            return playVideo({ path })
        }
    }

    return {
        playMediaFile,
    }
}
