import { NativePlayer_PlaybackInfo } from "@/api/generated/types"
import { VideoCoreSettings } from "@/app/(main)/_features/video-core/video-core.atoms"
import { logger } from "@/lib/helpers/debug"

const audioLog = logger("AUDIO")

export class VideoCoreAudioManager {

    onError: (error: string) => void
    private videoElement: HTMLVideoElement
    private settings: VideoCoreSettings
    // Playback info
    private playbackInfo: NativePlayer_PlaybackInfo

    constructor({
        videoElement,
        settings,
        playbackInfo,
        onError,
    }: {
        videoElement: HTMLVideoElement
        settings: VideoCoreSettings
        playbackInfo: NativePlayer_PlaybackInfo
        onError: (error: string) => void
    }) {
        this.videoElement = videoElement
        this.settings = settings
        this.playbackInfo = playbackInfo
        this.onError = onError

        if (this.videoElement.audioTracks) {
            // Check that audio tracks are loaded
            if (this.videoElement.audioTracks.length <= 0) {
                this.onError("The video element does not support the media's audio codec. Please try another media.")
                return
            }
            audioLog.info("Audio tracks", this.videoElement.audioTracks)
        }

        // Select the default track
        this._selectDefaultTrack()
    }

    _selectDefaultTrack() {
        const foundTracks = this.playbackInfo.mkvMetadata?.audioTracks?.filter?.(t => (t.language || "eng") === this.settings.preferredAudioLanguage)
        if (foundTracks?.length) {
            // Find default or forced track
            const defaultIndex = foundTracks.findIndex(t => t.forced)
            this.selectTrack(foundTracks[defaultIndex >= 0 ? defaultIndex : 0].number)
        }
    }

    selectTrackByLabel(trackLabel: string) {
        const track = this.playbackInfo.mkvMetadata?.audioTracks?.find?.(t => t.name === trackLabel)
        if (track) {
            this.selectTrack(track.number)
        } else {
            audioLog.error("Audio track not found", trackLabel)
        }
    }

    selectTrack(trackNumber: number) {
        if (!this.videoElement.audioTracks) return

        let trackChanged = false
        for (let i = 0; i < this.videoElement.audioTracks.length; i++) {
            const shouldEnable = this.videoElement.audioTracks[i].id === trackNumber.toString()
            if (this.videoElement.audioTracks[i].enabled !== shouldEnable) {
                this.videoElement.audioTracks[i].enabled = shouldEnable
                trackChanged = true
            }
        }

        if (trackChanged && this.videoElement.audioTracks.dispatchEvent) {
            this.videoElement.audioTracks.dispatchEvent(new Event("change"))
        }
    }

    getSelectedTrack(): number | null {
        if (!this.videoElement.audioTracks) return null

        for (let i = 0; i < this.videoElement.audioTracks.length; i++) {
            if (this.videoElement.audioTracks[i].enabled) {
                return Number(this.videoElement.audioTracks[i].id)
            }
        }

        return null
    }

}
