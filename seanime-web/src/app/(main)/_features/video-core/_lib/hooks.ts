import { isMobile } from "@/lib/utils/browser-detection"

export function useIsCodecSupported() {
    const isCodecSupported = (codec: string) => {
        if (isMobile()) return false
        if (navigator.userAgent.search("Firefox") === -1)
            codec = codec.replace("video/x-matroska", "video/mp4")
        const videos = document.getElementsByTagName("video")
        const video = videos.item(0) ?? document.createElement("video")
        return video.canPlayType(codec) === "probably"
    }

    return {
        isCodecSupported,
    }
}
