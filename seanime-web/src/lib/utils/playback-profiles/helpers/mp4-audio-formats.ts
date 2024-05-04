/**
 * @deprecated - Check @/utils/playback-profiles/index
 */

import { isTizen, isTizen4, isTizen5, isTizen55, isTv, isWebOS } from "../../browser-detection"
import { getSupportedAudioCodecs } from "./audio-formats"
import { hasVp8Support } from "./mp4-video-formats"

/**
 * Checks if the client can play the AC3 codec
 *
 * @param videoTestElement - A HTML video element for testing codecs
 * @returns Determines if the browser has AC3 support
 */
export function hasAc3Support(videoTestElement: HTMLVideoElement): boolean {
    if (isTv()) {
        return true
    }

    return !!videoTestElement
        .canPlayType("audio/mp4; codecs=\"ac-3\"")
        .replace(/no/, "")
}

/**
 * Checks if the client can play AC3 in a HLS stream
 * @param videoTestElement - A HTML video element for testing codecs
 * @returns Determines if the browser has AC3 support
 */
export function hasAc3InHlsSupport(
    videoTestElement: HTMLVideoElement,
): boolean {
    if (isTizen() || isWebOS()) {
        return true
    }

    if (videoTestElement.canPlayType) {
        return !!(
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
 * Checks if the cliemt has E-AC3 codec support
 *
 * @param videoTestElement - A HTML video element for testing codecs
 * @returns Determines if browser has EAC3 support
 */
export function hasEac3Support(videoTestElement: HTMLVideoElement): boolean {
    if (isTv()) {
        return true
    }

    return !!videoTestElement
        .canPlayType("audio/mp4; codecs=\"ec-3\"")
        .replace(/no/, "")
}

/**
 * Checks if the client has AAC codec support
 *
 * @param videoTestElement - A HTML video element for testing codecs
 * @returns Determines if browser has AAC support
 */
export function hasAacSupport(videoTestElement: HTMLVideoElement): boolean {
    return !!videoTestElement
        .canPlayType("video/mp4; codecs=\"avc1.640029, mp4a.40.2\"")
        .replace(/no/, "")
}

/**
 * Checks if the client has MP2 codec support
 *
 * @returns Determines if browser has MP2 support
 */
export function hasMp2AudioSupport(): boolean {
    return isTv()
}

/**
 * Checks if the client has MP3 audio codec support
 *
 * @param videoTestElement - A HTML video element for testing codecs
 * @returns Determines if browser has Mp3 support
 */
export function hasMp3AudioSupport(
    videoTestElement: HTMLVideoElement,
): boolean {
    return !!(
        videoTestElement
            .canPlayType("video/mp4; codecs=\"avc1.640029, mp4a.69\"")
            .replace(/no/, "") ||
        videoTestElement
            .canPlayType("video/mp4; codecs=\"avc1.640029, mp4a.6B\"")
            .replace(/no/, "") ||
        videoTestElement
            .canPlayType("video/mp4; codecs=\"avc1.640029, mp3\"")
            .replace(/no/, "")
    )
}

/**
 * Determines DTS audio support
 *
 * @param videoTestElement - A HTML video element for testing codecs
 * @returns Determines if browserr has DTS audio support
 */
export function hasDtsSupport(
    videoTestElement: HTMLVideoElement,
): boolean | string {
    // DTS audio not supported in 2018 models (Tizen 4.0)
    if (isTizen4() || isTizen5() || isTizen55()) {
        return false
    }

    return (
        isTv() ||
        videoTestElement
            .canPlayType("video/mp4; codecs=\"dts-\"")
            .replace(/no/, "") ||
        videoTestElement.canPlayType("video/mp4; codecs=\"dts+\"").replace(/no/, "")
    )
}

/**
 * Gets an array of supported MP4 codecs
 *
 * @param videoTestElement - A HTML video element for testing codecs
 * @returns Array of supported MP4 audio codecs
 */
export function getSupportedMP4AudioCodecs(
    videoTestElement: HTMLVideoElement,
): string[] {
    const codecs = []

    if (hasAacSupport(videoTestElement)) {
        codecs.push("aac")
    }

    if (hasMp3AudioSupport(videoTestElement)) {
        codecs.push("mp3")
    }

    if (hasAc3Support(videoTestElement)) {
        codecs.push("ac3")

        if (hasEac3Support(videoTestElement)) {
            codecs.push("eac3")
        }
    }

    if (hasMp2AudioSupport()) {
        codecs.push("mp2")
    }

    if (hasDtsSupport(videoTestElement)) {
        codecs.push("dca", "dts")
    }

    if (isTizen() || isWebOS()) {
        codecs.push("pcm_s16le", "pcm_s24le")
    }

    if (isTizen()) {
        codecs.push("aac_latm")
    }

    if (getSupportedAudioCodecs("opus")) {
        codecs.push("opus")
    }

    if (getSupportedAudioCodecs("flac")) {
        codecs.push("flac")
    }

    if (getSupportedAudioCodecs("alac")) {
        codecs.push("alac")
    }

    if (hasVp8Support(videoTestElement) || isTizen()) {
        codecs.push("vorbis")
    }

    return codecs
}
