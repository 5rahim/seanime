/**
 * @deprecated
 * Since we're targeting modern environments/devices only, it makes sense to switch
 * to the native MediaCapabilities API, widely supported on modern devices, but not in older.
 *
 * Given a media file, we should test with MC the compatibility of video, audio and subtitle streams
 * independently:
 * If success: Don't request transcoding and direct play that specific stream.
 * If failure: Request transcoding of the failing streams to a previously hardcoded
 * bitrate/codec combination
 *
 * For the hardcoded bitrate/codecs combination we can use what we know that are universally
 * compatible, even without testing for explicit compatibility (we can do simple checks,
 * but the more we do, the complex/less portable our solution can get).
 * Examples: H264, AAC and VTT/SASS (thanks to JASSUB).
 *
 * Other codec combinations can be hardcoded, even if they're not direct-playable in
 * most browsers (like H265 or AV1), so the few browsers that support them benefits from less bandwidth
 * usage (although this will rarely happen: The most expected situations when transcoding
 * is when the media's codecs are more "powerful" than what the client is capable of, and H265 is
 * pretty modern, so it would've been catched-up by MediaCapabilities. However,
 * we must take into account the playback of really old codecs like MPEG or H263,
 * whose support are probably likely going to be removed from browsers,
 * so MediaCapabilities reports as unsupported, so we would be going from an "inferior" codec to a
 * "superior" codec in this situation)
 */

import { getDirectPlayProfiles } from "./directplay-profile"
import { getCodecProfiles } from "./helpers/codec-profiles"
import { getResponseProfiles } from "./response-profile"
import { getSubtitleProfiles } from "./subtitle-profile"
import { getTranscodingProfiles } from "./transcoding-profile"

/**
 * Creates a device profile containing supported codecs for the active Cast device.
 *
 * @param videoTestElement - Dummy video element for compatibility tests
 */
function getDeviceProfile(videoTestElement: HTMLVideoElement) {
    // MaxStaticBitrate seems to be for offline sync only
    return {
        MaxStreamingBitrate: 120_000_000,
        MaxStaticBitrate: 0,
        MusicStreamingTranscodingBitrate: Math.min(120_000_000, 192_000),
        DirectPlayProfiles: getDirectPlayProfiles(videoTestElement),
        TranscodingProfiles: getTranscodingProfiles(videoTestElement),
        ContainerProfiles: [],
        CodecProfiles: getCodecProfiles(videoTestElement),
        SubtitleProfiles: getSubtitleProfiles(),
        ResponseProfiles: getResponseProfiles(),
    }
}

const videoTestElement = document.createElement("video")
const playbackProfile = getDeviceProfile(videoTestElement)

export default playbackProfile
