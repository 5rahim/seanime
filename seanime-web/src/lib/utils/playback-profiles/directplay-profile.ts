/**
 * @deprecated - Check @/utils/playback-profiles/index
 */

import { DirectPlayProfile, DlnaProfileType } from "@/lib/utils/playback-profiles/jellyfin-types"
import { getSupportedAudioCodecs } from "./helpers/audio-formats"
import { getSupportedMP4AudioCodecs } from "./helpers/mp4-audio-formats"
import { getSupportedMP4VideoCodecs } from "./helpers/mp4-video-formats"
import { hasMkvSupport } from "./helpers/transcoding-formats"
import { getSupportedWebMAudioCodecs } from "./helpers/webm-audio-formats"
import { getSupportedWebMVideoCodecs } from "./helpers/webm-video-formats"

/**
 * Returns a valid DirectPlayProfile for the current platform.
 *
 * @param videoTestElement - A HTML video element for testing codecs
 * @returns An array of direct play profiles for the current platform.
 */
export function getDirectPlayProfiles(
    videoTestElement: HTMLVideoElement,
): DirectPlayProfile[] {
    const DirectPlayProfiles: DirectPlayProfile[] = []

    const webmVideoCodecs = getSupportedWebMVideoCodecs(videoTestElement)
    const webmAudioCodecs = getSupportedWebMAudioCodecs(videoTestElement)

    const mp4VideoCodecs = getSupportedMP4VideoCodecs(videoTestElement)
    const mp4AudioCodecs = getSupportedMP4AudioCodecs(videoTestElement)

    if (webmVideoCodecs.length > 0) {
        DirectPlayProfiles.push({
            Container: "webm",
            Type: DlnaProfileType.Video,
            VideoCodec: webmVideoCodecs.join(","),
            AudioCodec: webmAudioCodecs.join(","),
        })
    }

    if (mp4VideoCodecs.length > 0) {
        DirectPlayProfiles.push({
            Container: "mp4,m4v",
            Type: DlnaProfileType.Video,
            VideoCodec: mp4VideoCodecs.join(","),
            AudioCodec: mp4AudioCodecs.join(","),
        })
    }

    if (hasMkvSupport(videoTestElement) && mp4VideoCodecs.length > 0) {
        DirectPlayProfiles.push({
            Container: "mkv",
            Type: DlnaProfileType.Video,
            VideoCodec: mp4VideoCodecs.join(","),
            AudioCodec: mp4AudioCodecs.join(","),
        })
    }

    const supportedAudio = [
        "opus",
        "mp3",
        "mp2",
        "aac",
        "flac",
        "alac",
        "webma",
        "wma",
        "wav",
        "ogg",
        "oga",
        "eac3",
    ]

    for (const audioFormat of supportedAudio.filter(format =>
        getSupportedAudioCodecs(format),
    )) {
        DirectPlayProfiles.push({
            Container: audioFormat,
            Type: DlnaProfileType.Audio,
        })

        if (audioFormat === "opus" || audioFormat === "webma") {
            DirectPlayProfiles.push({
                Container: "webm",
                Type: DlnaProfileType.Audio,
                AudioCodec: audioFormat,
            })
        }

        // Aac also appears in the m4a and m4b container
        if (audioFormat === "aac" || audioFormat === "alac") {
            DirectPlayProfiles.push(
                {
                    Container: "m4a",
                    AudioCodec: audioFormat,
                    Type: DlnaProfileType.Audio,
                },
                {
                    Container: "m4b",
                    AudioCodec: audioFormat,
                    Type: DlnaProfileType.Audio,
                },
            )
        }
    }

    return DirectPlayProfiles
}
