/**
 * @deprecated - Check @/utils/playback-profiles/index
 */

import { hasAv1Support, hasVp8Support, hasVp9Support } from "./mp4-video-formats"

/**
 * Get an array of supported codecs WebM video codecs
 */
export function getSupportedWebMVideoCodecs(
    videoTestElement: HTMLVideoElement,
): string[] {
    const codecs = []

    if (hasVp8Support(videoTestElement)) {
        codecs.push("vp8")
    }

    if (hasVp9Support(videoTestElement)) {
        codecs.push("vp9")
    }

    if (hasAv1Support(videoTestElement)) {
        codecs.push("av1")
    }

    return codecs
}
