/**
 * @deprecated - Check @/utils/playback-profiles/index
 */

import { isEdge, isTizen, isTv, supportsMediaSource } from "../../browser-detection"

/**
 * Checks if the client can play native HLS
 *
 * @param videoTestElement - A HTML video element for testing codecs
 * @returns Determines if the browser can play native Hls
 */
export function canPlayNativeHls(videoTestElement: HTMLVideoElement): boolean {
    if (isTizen()) {
        return true
    }

    return !!(
        videoTestElement.canPlayType("application/x-mpegURL").replace(/no/, "") ||
        videoTestElement
            .canPlayType("application/vnd.apple.mpegURL")
            .replace(/no/, "")
    )
}

/**
 * Determines if the browser can play Hls with Media Source Extensions
 */
export function canPlayHlsWithMSE(): boolean {
    return supportsMediaSource()
}

/**
 * Determines if the browser can play Mkvs
 */
export function hasMkvSupport(videoTestElement: HTMLVideoElement): boolean {
    if (isTv()) {
        return true
    }

    if (
        videoTestElement.canPlayType("video/x-matroska").replace(/no/, "") ||
        videoTestElement.canPlayType("video/mkv").replace(/no/, "")
    ) {
        return true
    }

    return !!isEdge()
}
