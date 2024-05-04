/**
 * @deprecated - Check @/utils/playback-profiles/index
 */

import { hasAacSupport, hasAc3InHlsSupport, hasAc3Support, hasEac3Support, hasMp3AudioSupport } from "./mp4-audio-formats"

/**
 * List of supported Ts audio codecs
 */
export function getSupportedTsAudioCodecs(
    videoTestElement: HTMLVideoElement,
): string[] {
    const codecs = []

    if (hasAacSupport(videoTestElement)) {
        codecs.push("aac")
    }

    if (hasMp3AudioSupport(videoTestElement)) {
        codecs.push("mp3")
    }

    if (hasAc3Support(videoTestElement) && hasAc3InHlsSupport(videoTestElement)) {
        codecs.push("ac3")

        if (hasEac3Support(videoTestElement)) {
            codecs.push("eac3")
        }
    }

    return codecs
}
