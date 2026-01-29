/**
 * @deprecated - Check @/utils/playback-profiles/index
 */

import { isTv } from "../../browser-detection"
import { getSupportedAudioCodecs } from "./audio-formats"
import { hasAacSupport, hasEac3Support } from "./mp4-audio-formats"
import { hasH264Support, hasH265Support } from "./mp4-video-formats"

/**
 * Check if client supports AC3 in HLS stream
 *
 * @param videoTestElement - A HTML video element for testing codecs
 * @returns Determines if the browser has AC3 in HLS support
 */
function supportsAc3InHls(
    videoTestElement: HTMLVideoElement,
): boolean | string {
    if (isTv()) {
        return true
    }

    if (videoTestElement.canPlayType) {
        return (
            videoTestElement
                .canPlayType("application/x-mpegurl; codecs=\"avc1.42E01E, ac-3\"")
                .replace(/no/, "") ||
            videoTestElement
                .canPlayType(
                    "application/vnd.apple.mpegURL; codecs=\"avc1.42E01E, ac-3\"",
                )
                .replace(/no/, "")
        )
    }

    return false
}

/**
 * Gets the supported HLS video codecs
 *
 * @param videoTestElement - A HTML video element for testing codecs
 * @returns Array of video codecs supported in HLS
 */
export function getHlsVideoCodecs(
    videoTestElement: HTMLVideoElement,
): string[] {
    const hlsVideoCodecs = []

    if (hasH264Support(videoTestElement)) {
        hlsVideoCodecs.push("h264")
    }

    if (hasH265Support(videoTestElement) || isTv()) {
        hlsVideoCodecs.push("h265", "hevc")
    }

    return hlsVideoCodecs
}

/**
 * Gets the supported HLS audio codecs
 *
 * @param videoTestElement - A HTML video element for testing codecs
 * @returns Array of audio codecs supported in HLS
 */
export function getHlsAudioCodecs(
    videoTestElement: HTMLVideoElement,
): string[] {
    const hlsVideoAudioCodecs = []

    if (supportsAc3InHls(videoTestElement)) {
        hlsVideoAudioCodecs.push("ac3")

        if (hasEac3Support(videoTestElement)) {
            hlsVideoAudioCodecs.push("eac3")
        }
    }

    if (hasAacSupport(videoTestElement)) {
        hlsVideoAudioCodecs.push("aac")
    }

    if (getSupportedAudioCodecs("opus")) {
        hlsVideoAudioCodecs.push("opus")
    }

    return hlsVideoAudioCodecs
}
