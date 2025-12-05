import { vc_getCaptionStyle } from "@/app/(main)/_features/video-core/video-core-settings-menu"
import { getDefaultSubtitleTrackNumber } from "@/app/(main)/_features/video-core/video-core-subtitles"
import { VideoCoreSettings } from "@/app/(main)/_features/video-core/video-core.atoms"
import { logger } from "@/lib/helpers/debug"
import { CaptionsRenderer, ParsedCaptionsResult, parseResponse, VTTCue, VTTRegion } from "media-captions"
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
    index: number
    metadata: MediaCaptionsTrackInfo
    cues: VTTCue[]
    regions: VTTRegion[]
    loaded: boolean
    loadFn: () => Promise<ParsedCaptionsResult> | null
}

const NO_TRACK_IDX = -1

/**
 * Manages subtitles rendered using media-captions.
 * ```tsx
 * <VideoCore
 *   state={{
 *     active: true,
 *     playbackInfo: {
 *       id: "video-1",
 *       playbackType: "onlinestream",
 *       streamUrl: "https://example.com/video.mp4",
 *       streamType: "native",
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
    private wrapperElement: HTMLDivElement | null = null
    private overlayElement: HTMLDivElement | null = null
    private currentTrackIndex: number = NO_TRACK_IDX
    private timeUpdateListener: (() => void) | null = null
    private readonly settings: VideoCoreSettings
    private captionCustomization: VideoCoreSettings["captionCustomization"]
    private subtitleDelay = 0

    private _onSelectedTrackChanged?: (track: number | null) => void
    private _onTracksLoaded?: (tracks: MediaCaptionsTrack[]) => void

    constructor(options: MediaCaptionsManagerOptions) {
        this.videoElement = options.videoElement
        this.tracks = options.tracks
        this.settings = options.settings
        this.captionCustomization = options.settings.captionCustomization
        this.subtitleDelay = options.settings.subtitleDelay ?? 0

        this.init()
    }

    public updateSettings(settings: VideoCoreSettings) {
        this.captionCustomization = settings.captionCustomization
        this.setSubtitleDelay(settings.subtitleDelay ?? 0)
        this.applyCaptionStyles()

        if (this.renderer && this.currentTrackIndex !== NO_TRACK_IDX) {
            this.renderer.currentTime = this.videoElement.currentTime + (-this.subtitleDelay)
        }
    }

    addTracksLoadedEventListener(callback: (tracks: MediaCaptionsTrack[]) => void) {
        this._onTracksLoaded = callback
    }

    addTrackChangedEventListener(callback: (track: number | null) => void) {
        this._onSelectedTrackChanged = callback
    }

    setSubtitleDelay(delay: number) {
        this.subtitleDelay = delay
    }

    public getTracks(): MediaCaptionsTrack[] {
        return this.loadedTracks.map((loadedTrack, index) => {
            return {
                number: loadedTrack.index,
                label: loadedTrack.metadata.label,
                language: loadedTrack.metadata.language,
                selected: this.currentTrackIndex === index,
            }
        })
    }

    public async selectTrack(index: number) {
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

        this._onSelectedTrackChanged?.(index)

        if (this.renderer) {
            if (!track.loaded) {
                log.info("Loading track", index)
                const res = await track.loadFn()
                if (res) {
                    track.cues = res.cues
                    track.regions = res.regions
                }
                track.loaded = true
            }
            this.renderer.changeTrack({
                cues: track.cues,
                regions: track.regions,
            })
            this.renderer.currentTime = this.videoElement.currentTime + (-this.subtitleDelay)
        }
    }

    public setNoTrack() {
        this.currentTrackIndex = NO_TRACK_IDX
        if (this.renderer) {
            this.renderer.reset()
        }
        this._onSelectedTrackChanged?.(NO_TRACK_IDX)
        log.info("Disabled subtitles")
    }

    /*
     * Render captions to a canvas context for PIP mode
     */
    public async renderToCanvas(context: CanvasRenderingContext2D, width: number, height: number, currentTime: number) {
        if (this.currentTrackIndex === NO_TRACK_IDX || !this.renderer) return

        const track = this.loadedTracks[this.currentTrackIndex]
        if (!track) return

        if (!track.loaded) {
            const res = await track.loadFn()
            if (res) {
                track.cues = res.cues
                track.regions = res.regions
            }
            track.loaded = true
        }

        // Find active cues for current time
        const activeCues = track.cues.filter(cue =>
            currentTime >= cue.startTime && currentTime <= cue.endTime,
        )

        if (activeCues.length === 0) return

        // Render each active cue
        activeCues.forEach((cue) => {
            const text = cue.text
            if (!text) return

            context.save()

            // Calculate position (bottom center by default)
            const fontSize = Math.max(width * 0.04, 20) // 4% of video width
            const padding = fontSize * 0.5
            const bottomMargin = height * 0.1 // 10% from bottom
            const maxWidth = width * 0.9 // Use 90% of canvas width
            const lineHeight = fontSize * 1.3

            // Setup text rendering
            context.font = `bold ${fontSize}px Inter, Arial, sans-serif`
            context.textAlign = "center"
            context.textBaseline = "bottom"

            // Word wrap the text
            const words = text.split(" ")
            const lines: string[] = []
            let currentLine = ""

            for (const word of words) {
                const testLine = currentLine ? `${currentLine} ${word}` : word
                const metrics = context.measureText(testLine)

                if (metrics.width > maxWidth && currentLine) {
                    lines.push(currentLine)
                    currentLine = word
                } else {
                    currentLine = testLine
                }
            }
            if (currentLine) {
                lines.push(currentLine)
            }

            // Calculate dimensions for all lines
            let maxLineWidth = 0
            lines.forEach(line => {
                const metrics = context.measureText(line)
                maxLineWidth = Math.max(maxLineWidth, metrics.width)
            })

            const totalHeight = lines.length * lineHeight
            const x = width / 2
            const y = height - bottomMargin

            // Draw background box
            context.fillStyle = "rgba(0, 0, 0, 0.8)"
            context.fillRect(
                x - maxLineWidth / 2 - padding,
                y - totalHeight - padding,
                maxLineWidth + padding * 2,
                totalHeight + padding * 2,
            )

            // Draw each line with outline
            context.strokeStyle = "black"
            context.lineWidth = fontSize * 0.1
            context.lineJoin = "round"

            lines.forEach((line, index) => {
                const lineY = y - (lines.length - 1 - index) * lineHeight
                context.strokeText(line, x, lineY)
            })

            context.fillStyle = "white"
            lines.forEach((line, index) => {
                const lineY = y - (lines.length - 1 - index) * lineHeight
                context.fillText(line, x, lineY)
            })

            context.restore()
        })
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

    private applyCaptionStyles() {
        if (!this.overlayElement) return

        const custom = this.captionCustomization
        const useCustom = true // custom.enabled

        if (!useCustom) {
            // Remove custom styles
            this.overlayElement.style.removeProperty("--media-font-family")
            this.overlayElement.style.removeProperty("--cue-height")
            this.overlayElement.style.removeProperty("--cue-font-size")
            this.overlayElement.style.removeProperty("--cue-line-height")
            this.overlayElement.style.removeProperty("--media-text-color")
            this.overlayElement.style.removeProperty("--cue-color")
            this.overlayElement.style.removeProperty("--cue-bg-color")
            this.overlayElement.style.removeProperty("--cue-font-weight")
            this.overlayElement.style.removeProperty("--cue-padding-x")
            this.overlayElement.style.removeProperty("--cue-padding-y")
            this.overlayElement.style.removeProperty("--cue-text-shadow")
            this.overlayElement.style.removeProperty("--overlay-padding")
            return
        }

        this.overlayElement.style.setProperty("--overlay-padding", "3%")

        // if (custom.fontFamily) {
        //     this.overlayElement.style.setProperty("--media-font-family", custom.fontFamily)
        // }
        const fontSize = vc_getCaptionStyle(custom, "fontSize")
        if (fontSize) {
            // Override the calculated font size and recalculate all dependent values
            const newFontSize = `calc(var(--overlay-height) / 100 * ${fontSize})`
            this.overlayElement.style.setProperty("--cue-font-size", newFontSize)
            // this.overlayElement.style.setProperty("--cue-line-height", `calc(${newFontSize} * 1.2)`)
            // this.overlayElement.style.setProperty("--cue-padding-x", `calc(${newFontSize} * 0.6)`)
            // this.overlayElement.style.setProperty("--cue-padding-y", `calc(${newFontSize} * 0.4)`)
        }
        const textColor = vc_getCaptionStyle(custom, "textColor")
        if (textColor) {
            this.overlayElement.style.setProperty("--media-text-color", textColor)
            this.overlayElement.style.setProperty("--cue-color", textColor)
        }
        const backgroundColor = vc_getCaptionStyle(custom, "backgroundColor")
        const backgroundOpacity = vc_getCaptionStyle(custom, "backgroundOpacity")
        if (backgroundColor) {
            const opacity = backgroundOpacity !== undefined ? backgroundOpacity : 0.8
            const hex = backgroundColor.replace("#", "")
            const r = parseInt(hex.substring(0, 2), 16)
            const g = parseInt(hex.substring(2, 4), 16)
            const b = parseInt(hex.substring(4, 6), 16)
            this.overlayElement.style.setProperty("--cue-bg-color", `rgba(${r}, ${g}, ${b}, ${opacity})`)
        }
        // if (custom.bold !== undefined) {
        //     this.overlayElement.style.setProperty("--cue-font-weight", custom.bold ? "bold" : "normal")
        // }
        const textShadow = vc_getCaptionStyle(custom, "textShadow")
        const textShadowColor = vc_getCaptionStyle(custom, "textShadowColor")
        if (textShadow !== undefined) {
            if (textShadow === 0) {
                this.overlayElement.style.setProperty("--cue-text-shadow", "none")
            } else {
                const shadowColor = textShadowColor || "#000000"
                this.overlayElement.style.setProperty("--cue-text-shadow", `${shadowColor} 1px 1px ${textShadow}px`)
            }
        }
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

        if (this.wrapperElement) {
            this.wrapperElement.remove()
            this.wrapperElement = null
        }

        this.loadedTracks = []
        this.tracks = []
        this.currentTrackIndex = NO_TRACK_IDX
    }

    private async init() {
        log.info("Initializing media-captions manager", this.tracks)

        this.wrapperElement = document.createElement("div")
        this.wrapperElement.id = "video-core-captions-wrapper"
        this.wrapperElement.style.position = "absolute"
        this.wrapperElement.style.inset = "0"
        this.wrapperElement.style.pointerEvents = "none"
        this.wrapperElement.style.zIndex = "10"
        this.wrapperElement.style.overflow = "hidden"
        this.wrapperElement.classList.add("transform-gpu", "transition-all", "duration-300", "ease-in-out")
        this.videoElement.parentElement?.appendChild(this.wrapperElement)

        // Create overlay element for captions
        this.overlayElement = document.createElement("div")
        this.overlayElement.id = "video-core-captions-overlay"
        this.overlayElement.style.position = "absolute"
        this.overlayElement.style.inset = "0"
        this.overlayElement.style.pointerEvents = "none"
        this.overlayElement.style.zIndex = "10"
        this.overlayElement.style.overflow = "hidden"

        // Insert overlay after video element
        this.wrapperElement.appendChild(this.overlayElement)

        // Create renderer
        this.renderer = new CaptionsRenderer(this.overlayElement)

        // Apply custom styles
        this.applyCaptionStyles()

        // Load all tracks
        await this.loadTracks()

    }

    private async loadTracks() {
        for (let i = 0; i < this.tracks.length; i++) {
            const track = this.tracks[i]
            try {
                this.loadedTracks.push({
                    index: i,
                    metadata: track,
                    cues: [],
                    regions: [],
                    loaded: false,
                    loadFn: async () => {
                        return await parseResponse(fetch(track.src), {
                            // type: track.type,
                        })
                    },
                })

                log.info(`Loaded track: ${track.label}`)
            }
            catch (error) {
                log.error(`Failed to load track: ${track.label}`, error)
            }
        }
        // When the first track is loaded, start rendering captions
        // Select default track
        const defaultTrackNumber = getDefaultSubtitleTrackNumber(this.settings, this.tracks.map((t, idx) => ({ ...t, number: idx })))
        await this.selectTrack(defaultTrackNumber)
        // Setup time update listener
        this.timeUpdateListener = () => {
            if (this.renderer && this.currentTrackIndex !== NO_TRACK_IDX) {
                this.renderer.currentTime = this.videoElement.currentTime + (-this.subtitleDelay)
            }
        }
        this.videoElement.addEventListener("timeupdate", this.timeUpdateListener)
        this._onTracksLoaded?.(this.getTracks())
    }
}


