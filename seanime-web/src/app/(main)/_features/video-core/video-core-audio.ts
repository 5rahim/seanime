import { HlsAudioTrack } from "@/app/(main)/_features/video-core/video-core-hls"
import { VideoCorePlaybackInfo, VideoCoreSettings } from "@/app/(main)/_features/video-core/video-core.atoms"
import { logger } from "@/lib/helpers/debug"

const audioLog = logger("AUDIO")

export class VideoCoreAudioManager {

    onError: (error: string) => void
    private videoElement: HTMLVideoElement
    private settings: VideoCoreSettings
    // Playback info
    private playbackInfo: VideoCorePlaybackInfo
    // HLS-specific
    private readonly hlsSetAudioTrack: ((trackId: number) => void) | null = null
    private readonly hlsAudioTracks: HlsAudioTrack[] = []
    private hlsCurrentAudioTrack: number = -1

    constructor({
        videoElement,
        settings,
        playbackInfo,
        onError,
        hlsSetAudioTrack,
        hlsAudioTracks,
        hlsCurrentAudioTrack,
    }: {
        videoElement: HTMLVideoElement
        settings: VideoCoreSettings
        playbackInfo: VideoCorePlaybackInfo
        onError: (error: string) => void
        hlsSetAudioTrack?: ((trackId: number) => void) | null
        hlsAudioTracks?: HlsAudioTrack[]
        hlsCurrentAudioTrack?: number
    }) {
        this.videoElement = videoElement
        this.settings = settings
        this.playbackInfo = playbackInfo
        this.onError = onError
        this.hlsSetAudioTrack = hlsSetAudioTrack || null
        this.hlsAudioTracks = hlsAudioTracks || []
        this.hlsCurrentAudioTrack = hlsCurrentAudioTrack ?? -1

        // Check if we're dealing with HLS audio tracks
        const isHls = this.hlsSetAudioTrack !== null && this.hlsAudioTracks.length > 0

        if (!isHls && this.videoElement.audioTracks) {
            // MKV audio track handling
            // Check that audio tracks are loaded
            if (this.videoElement.audioTracks.length <= 0) {
                this.onError("The player does not support this audio codec. Please try another file or use an external player.")
                return
            }
            audioLog.info("Audio tracks", this.videoElement.audioTracks)
        }

        // Select the default track
        this._selectDefaultTrack()
    }

    _selectDefaultTrack() {
        // Check if we're dealing with HLS
        if (this.isHlsStream()) {
            this.__selectDefaultHlsTrack()
            return
        }

        // Event based track selection
        this.__selectDefaultEventTrack()
    }

    __selectDefaultEventTrack() {
        // Event based track selection
        // Split preferred languages by comma and trim whitespace
        const preferredLanguages = this.settings.preferredAudioLanguage
            .split(",")
            .map(lang => lang.trim())
            .filter(lang => lang.length > 0)

        // Try each preferred language in order
        for (const preferredLang of preferredLanguages) {
            const foundTracks = this.playbackInfo.mkvMetadata?.audioTracks?.filter?.(t => (t.language || "eng") === preferredLang)
            if (foundTracks?.length) {
                // Find default track
                const defaultIndex = foundTracks.findIndex(t => t.default)
                this.selectTrack(foundTracks[defaultIndex >= 0 ? defaultIndex : 0].number)
                return
            }
        }

        // No preferred tracks found, look for default or forced tracks
        const defaultOrForcedTracks = this.playbackInfo.mkvMetadata?.audioTracks?.filter?.(t => t.default || t.forced)
        if (defaultOrForcedTracks?.length) {
            // Prioritize default tracks over forced tracks
            const defaultIndex = defaultOrForcedTracks.findIndex(t => t.default)
            this.selectTrack(defaultOrForcedTracks[defaultIndex >= 0 ? defaultIndex : 0].number)
            return
        }
    }

    __selectDefaultHlsTrack() {
        if (!this.hlsSetAudioTrack || !this.isHlsStream()) return

        // Split preferred languages by comma and trim whitespace
        const preferredLanguages = this.settings.preferredAudioLanguage
            .split(",")
            .map(lang => lang.trim())
            .filter(lang => lang.length > 0)

        // Try each preferred language in order
        for (const preferredLang of preferredLanguages) {
            const foundTrack = this.hlsAudioTracks.find(t => (t.language || "eng") === preferredLang)
            if (foundTrack) {
                audioLog.info("Selecting preferred HLS audio track", foundTrack)
                this.hlsSetAudioTrack(foundTrack.id)
                return
            }
        }

        // No preferred track found, look for default track
        const defaultTrack = this.hlsAudioTracks.find(t => t.default)
        if (defaultTrack) {
            audioLog.info("Selecting default HLS audio track", defaultTrack)
            this.hlsSetAudioTrack(defaultTrack.id)
            return
        }

        // Otherwise, select the first track
        if (this.hlsAudioTracks.length > 0) {
            audioLog.info("Selecting first HLS audio track", this.hlsAudioTracks[0])
            this.hlsSetAudioTrack(this.hlsAudioTracks[0].id)
        }
    }

    selectTrack(trackNumber: number) {
        // If it's an HLS stream, select the track from the HLS API
        if (this.hlsSetAudioTrack) {
            audioLog.info("Selecting HLS audio track", trackNumber)
            this.hlsSetAudioTrack(trackNumber)
            return
        }

        // MKV audio track selection
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

    getSelectedTrackNumberOrNull(): number | null {
        // Check if we're dealing with HLS
        if (this.hlsSetAudioTrack) {
            return this.hlsCurrentAudioTrack
        }

        // MKV audio track
        if (!this.videoElement.audioTracks) return null

        for (let i = 0; i < this.videoElement.audioTracks.length; i++) {
            if (this.videoElement.audioTracks[i].enabled) {
                return Number(this.videoElement.audioTracks[i].id)
            }
        }

        return null
    }

    getHlsAudioTracks() {
        return this.hlsAudioTracks
    }

    // Update the current HLS audio track (called externally when track changes)
    onHlsTrackChange(trackId: number) {
        this.hlsCurrentAudioTrack = trackId
    }

    isHlsStream() {
        return this.hlsSetAudioTrack !== null && this.hlsAudioTracks.length > 0
    }

}
