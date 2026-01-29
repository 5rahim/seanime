/**
 * @deprecated - Check @/utils/playback-profiles/index
 */

import { isWebOS } from "../../browser-detection"

/**
 * Get an array of supported codecs
 */
export function getSupportedWebMAudioCodecs(
    videoTestElement: HTMLVideoElement,
): string[] {
    const codecs = []

    codecs.push("vorbis")

    if (
        !isWebOS() &&
        videoTestElement.canPlayType("audio/ogg; codecs=\"opus\"").replace(/no/, "")
    ) {
        codecs.push("opus")
    }

    return codecs
}
