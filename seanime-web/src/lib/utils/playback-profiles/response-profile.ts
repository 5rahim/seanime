/**
 * @deprecated - Check @/utils/playback-profiles/index
 */

import { DlnaProfileType } from "@/lib/utils/playback-profiles/jellyfin-types"

/**
 * Returns a valid ResponseProfile for the current platform.
 *
 * @returns An array of subtitle profiles for the current platform.
 */
export function getResponseProfiles() {
    const ResponseProfiles = []

    ResponseProfiles.push({
        Type: DlnaProfileType.Video,
        Container: "m4v",
        MimeType: "video/mp4",
    })

    return ResponseProfiles
}
