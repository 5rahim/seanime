import { getDefaultSubtitleTrackNumber } from "@/app/(main)/_features/video-core/video-core-subtitles"
import { VideoCoreSettings } from "@/app/(main)/_features/video-core/video-core.atoms"
import { logger } from "@/lib/helpers/debug"
import { CaptionsRenderer, parseResponse, VTTCue } from "media-captions"
import "media-captions/styles/captions.css"
import "media-captions/styles/regions.css"

const log = logger("VIDEO CORE MEDIA CAPTIONS")

export type MediaCaptionsTrackInfo = {
    src: string
    label: string
    language: string
    type?: "vtt" | "srt" | "ssa" | "ass"
    default?: boolean
}

export type MediaCaptionsTrack = {
    number: number
    label: string
    language: string
    selected: boolean
}

export type MediaCaptionsManagerOptions = {
    videoElement: HTMLVideoElement
    tracks: MediaCaptionsTrackInfo[]
    settings: VideoCoreSettings
}

type LoadedTrack = {
    metadata: MediaCaptionsTrackInfo
    cues: VTTCue[]
    regions: any[]
}

const NO_TRACK_IDX = -1

/**
 * Manages non-ASS subtitles.
 * ```tsx
 * <VideoCore
 *   state={{
 *     active: true,
 *     playbackInfo: {
 *       id: "video-1",
 *       playbackType: "onlinestream",
 *       streamUrl: "https://example.com/video.mp4",
 *       streamType: "stream",
 *       subtitleTracks: [
 *         {
 *           src: "https://example.com/subtitles/en.vtt",
 *           label: "English",
 *           language: "en",
 *           type: "vtt",
 *           default: true
 *         },
 *         {
 *           src: "https://example.com/subtitles/es.srt",
 *           label: "Spanish",
 *           language: "es",
 *           type: "srt"
 *         }
 *       ]
 *     },
 *     playbackError: null,
 *     loadingState: null
 *   }}
 * />
 * ```
 */
export class MediaCaptionsManager {
    private videoElement: HTMLVideoElement
    private tracks: MediaCaptionsTrackInfo[] = []
    private loadedTracks: LoadedTrack[] = []
    private renderer: CaptionsRenderer | null = null
    private overlayElement: HTMLDivElement | null = null
    private currentTrackIndex: number = NO_TRACK_IDX
    private timeUpdateListener: (() => void) | null = null
    private readonly settings: VideoCoreSettings

    private _onSelectedTrackChanged?: (track: number | null) => void

    constructor(options: MediaCaptionsManagerOptions) {
        this.videoElement = options.videoElement
        this.tracks = options.tracks
        this.settings = options.settings

        this.init()
    }

    addTrackChangedEventListener(callback: (track: number | null) => void) {
        this._onSelectedTrackChanged = callback
    }

    public selectTrack(index: number) {
        if (index < 0 || index >= this.tracks.length) {
            this.setNoTrack()
            return
        }

        const track = this.loadedTracks[index]
        if (!track) {
            log.error("Track not loaded", index)
            return
        }

        this.currentTrackIndex = index
        log.info(`Selected track: ${this.tracks[index].label}`)

        if (this.renderer) {
            this.renderer.changeTrack({
                cues: track.cues,
                regions: track.regions,
            })
            this.renderer.currentTime = this.videoElement.currentTime
        }

        this._onSelectedTrackChanged?.(index)
    }

    public setNoTrack() {
        this.currentTrackIndex = NO_TRACK_IDX
        if (this.renderer) {
            this.renderer.reset()
        }
        this._onSelectedTrackChanged?.(NO_TRACK_IDX)
        log.info("Disabled subtitles")
    }

    public getTracks(): MediaCaptionsTrack[] {
        return this.tracks.map((track, index) => ({
            number: index,
            label: track.label,
            language: track.language,
            selected: this.currentTrackIndex === index,
        }))
    }

    public getSelectedTrack(): MediaCaptionsTrackInfo | null {
        if (this.currentTrackIndex === NO_TRACK_IDX) return null
        return this.tracks[this.currentTrackIndex]
    }

    public getTrack(index: number | null): MediaCaptionsTrackInfo | undefined {
        return this.tracks[index ?? NO_TRACK_IDX]
    }

    public getSelectedTrackIndexOrNull() {
        if (this.currentTrackIndex === NO_TRACK_IDX) return null
        return this.currentTrackIndex
    }

    public destroy() {
        log.info("Destroying media-captions manager")

        if (this.timeUpdateListener) {
            this.videoElement.removeEventListener("timeupdate", this.timeUpdateListener)
            this.timeUpdateListener = null
        }

        if (this.renderer) {
            this.renderer.destroy()
            this.renderer = null
        }

        if (this.overlayElement) {
            this.overlayElement.remove()
            this.overlayElement = null
        }

        this.loadedTracks = []
        this.tracks = []
        this.currentTrackIndex = NO_TRACK_IDX
    }

    private async init() {
        log.info("Initializing media-captions manager", this.tracks)

        // Create overlay element for captions
        this.overlayElement = document.createElement("div")
        this.overlayElement.id = "video-core-captions-overlay"
        this.overlayElement.style.position = "absolute"
        this.overlayElement.style.inset = "0"
        this.overlayElement.style.pointerEvents = "none"
        this.overlayElement.style.zIndex = "10"
        this.overlayElement.style.overflow = "hidden"

        // Insert overlay after video element
        this.videoElement.parentElement?.appendChild(this.overlayElement)

        // Create renderer
        this.renderer = new CaptionsRenderer(this.overlayElement)

        // Load all tracks
        await this.loadTracks()

        // Select default track
        // const defaultTrackIndex = this.tracks.findIndex(t => t.default)
        // if (defaultTrackIndex !== -1) {
        //     this.selectTrack(defaultTrackIndex)
        // }
        const defaultTrackNumber = getDefaultSubtitleTrackNumber(this.settings, this.tracks.map((t, idx) => ({ ...t, number: idx })))
        this.selectTrack(defaultTrackNumber)

        // Setup time update listener
        this.timeUpdateListener = () => {
            if (this.renderer && this.currentTrackIndex !== NO_TRACK_IDX) {
                this.renderer.currentTime = this.videoElement.currentTime
            }
        }
        this.videoElement.addEventListener("timeupdate", this.timeUpdateListener)
    }

    private async loadTracks() {
        for (const track of this.tracks) {
            try {
                const result = await parseResponse(fetch(track.src), {
                    // type: track.type,
                })

                this.loadedTracks.push({
                    metadata: track,
                    cues: result.cues,
                    regions: result.regions,
                })

                log.info(`Loaded track: ${track.label}`, result)
            }
            catch (error) {
                log.error(`Failed to load track: ${track.label}`, error)
            }
        }
    }
}


