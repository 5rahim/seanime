import { isMobile } from "@/lib/utils/browser-detection"

export function useIsCodecSupported() {
    const isCodecSupported = (codec: string) => {
        if (!codec) return false
        if (isMobile()) return false

        const videos = document.getElementsByTagName("video")
        const video = videos.item(0) ?? document.createElement("video")

        if (codec.startsWith("video/x-matroska")) {
            // Firefox cannot demux Matroska at all
            if (navigator.userAgent.includes("Firefox")) return false
            // Chromium-based browsers demux mkv through their mp4/webm pipelines,
            // but report no support for "video/x-matroska" itself. Test the codec
            // set against both containers: h264/hevc/av1 + aac/opus/flac resolve
            // through mp4, while vp8/vp9 + vorbis resolve through webm.
            const mp4 = codec.replace("video/x-matroska", "video/mp4")
            const webm = codec.replace("video/x-matroska", "video/webm")
            return video.canPlayType(mp4) === "probably" || video.canPlayType(webm) === "probably"
        }

        return video.canPlayType(codec) === "probably"
    }

    return {
        isCodecSupported,
    }
}
