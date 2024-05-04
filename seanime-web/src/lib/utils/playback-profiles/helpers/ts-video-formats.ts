/**
 * @deprecated - Check @/utils/playback-profiles/index
 */

import { hasH264Support } from "./mp4-video-formats"

/**
 * List of supported ts video codecs
 */
export function getSupportedTsVideoCodecs(
    videoTestElement: HTMLVideoElement,
): string[] {
    const codecs = []

    if (hasH264Support(videoTestElement)) {
        codecs.push("h264")
    }

    return codecs
}
