/**
 * @deprecated - Check @/utils/playback-profiles/index
 */

import { isApple, isChrome, isEdge, isFirefox, isTizen, isWebOS } from "../../browser-detection"
import { hasH264Support, hasHevcSupport } from "./mp4-video-formats"

/**
 * Gets an array of supported fmp4 video codecs
 *
 * @param videoTestElement - A HTML video element for testing codecs
 * @returns List of supported fmp4 video codecs
 */
export function getSupportedFmp4VideoCodecs(
    videoTestElement: HTMLVideoElement,
): string[] {
    const codecs = []

    if (
        (isApple() || isEdge() || isTizen() || isWebOS()) &&
        hasHevcSupport(videoTestElement)
    ) {
        codecs.push("hevc")
    }

    if (
        hasH264Support(videoTestElement) &&
        (isChrome() ||
            isFirefox() ||
            isApple() ||
            isEdge() ||
            isTizen() ||
            isWebOS())
    ) {
        codecs.push("h264")
    }

    return codecs
}
