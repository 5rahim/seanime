import { isChromiumBased, isMobile } from "@/lib/utils/browser-detection"
import { checkCodecSupport } from "./codec-utils"

export function useIsCodecSupported() {
    const isCodecSupported = (codec: string) => {
        const canPlayType = (c: string) => {
            if (typeof document === "undefined") return ""
            const videos = document.getElementsByTagName("video")
            const video = videos.item(0) ?? document.createElement("video")
            return video.canPlayType(c)
        }

        return checkCodecSupport(codec, {
            isMobile: isMobile(),
            canUseMatroskaFallback: isChromiumBased(),
            canPlayType,
        })
    }

    return {
        isCodecSupported,
    }
}
