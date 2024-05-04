/**
 * @deprecated - Check @/utils/playback-profiles/index
 */

import { isEdge } from "../../browser-detection"
import { getSupportedAudioCodecs } from "./audio-formats"
import { hasAacSupport, hasAc3InHlsSupport, hasAc3Support, hasEac3Support, hasMp3AudioSupport } from "./mp4-audio-formats"

/**
 * Gets an array with the supported fmp4 codecs
 *
 * @param videoTestElement - A HTML video element for testing codecs
 * @returns List of supported FMP4 audio codecs
 */
export function getSupportedFmp4AudioCodecs(
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

    if (getSupportedAudioCodecs("flac") && !isEdge()) {
        codecs.push("flac")
    }

    if (getSupportedAudioCodecs("alac")) {
        codecs.push("alac")
    }

    return codecs
}
