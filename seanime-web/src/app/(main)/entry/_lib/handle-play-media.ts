import { usePlaybackPlayVideo } from "@/api/hooks/playback_manager.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useMediastreamCurrentFile } from "@/app/(main)/mediastream/_lib/mediastream.atoms"
import { isMobile, isPs4, isTv, isXbox } from "@/lib/utils/browser-detection"
import { useRouter } from "next/navigation"
import React from "react"
import { toast } from "sonner"

export function useHandlePlayMedia() {
    const router = useRouter()
    const serverStatus = useServerStatus()

    const { mutate: playVideo } = usePlaybackPlayVideo()

    const { setFilePath } = useMediastreamCurrentFile()

    function playMediaFile({ path }: { path: string }) {
        if (isMobile() || isTv() || isPs4() || isXbox()) { // TODO: Find a way to override this
            if (serverStatus?.featureFlags?.experimental?.mediastream) { // TODO: Remove when stable
                setFilePath(path)
                React.startTransition(() => {
                    router.push("/mediastream")
                })
            } else {
                toast.error("Playback is not supported on this device.")
            }
            return
        } else {
            return playVideo({ path })
        }
    }

    function streamMediaFile({ path }: { path: string }) {
        if (serverStatus?.featureFlags?.experimental?.mediastream) {
            setFilePath(path)
            React.startTransition(() => {
                router.push("/mediastream")
            })
        }
    }

    return {
        playMediaFile,
        streamMediaFile,
    }
}
