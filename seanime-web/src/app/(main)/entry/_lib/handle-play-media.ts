import { usePlaybackPlayVideo } from "@/api/hooks/playback_manager.hooks"
import { isMobile, isPs4, isTv, isXbox } from "@/lib/utils/browser-detection"
import { toast } from "sonner"

export function useHandlePlayMedia() {

    const { mutate: playVideo } = usePlaybackPlayVideo()

    function playMediaFile({ path }: { path: string }) {
        if (isMobile() || isTv() || isPs4() || isXbox()) {
            toast.error("Playback is not supported on this device.")
            return
        } else {
            return playVideo({ path })
        }
    }

    return {
        playMediaFile,
    }
}
