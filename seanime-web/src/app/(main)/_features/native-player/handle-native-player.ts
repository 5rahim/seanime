import { MKVParser_SubtitleEvent, NativePlayer_PlaybackInfo } from "@/api/generated/types"
import { logger } from "@/lib/helpers/debug"
import { legacy_getAssetUrl } from "@/lib/server/assets"
import { isApple } from "@/lib/utils/browser-detection"
import JASSUB, { ASS_Event } from "jassub"

const log = logger("STREAM SUBTITLE MANAGER")

export class StreamSubtitleManager {
    // Video element
    private videoElement: HTMLVideoElement
    // LibASS renderer
    private libassRenderer: JASSUB | null = null
    // JASSUB offscreen render
    private jassubOffscreenRender: boolean
    // Subtitles for each track
    private subtitleRecord: Record<number, MKVParser_SubtitleEvent[]> = {}
    private subtitleSet: Record<string, Set<string>> = {}
    // Default fonts
    private defaultFonts: string[] = ["/jassub/default.woff2"]
    // Current subtitle track number
    private currentTrack: number | null = null
    // Playback info
    private playbackInfo: NativePlayer_PlaybackInfo

    constructor({
        videoElement,
        jassubOffscreenRender,
        playbackInfo,
    }: {
        videoElement: HTMLVideoElement
        jassubOffscreenRender: boolean
        playbackInfo: NativePlayer_PlaybackInfo
    }) {
        this.videoElement = videoElement
        this.jassubOffscreenRender = jassubOffscreenRender
        this.libassRenderer = null
        this.subtitleRecord = {}
        this.playbackInfo = playbackInfo
    }

    loadTracks() {
        if (!this.playbackInfo) {
            log.error("Cannot load tracks, no playback info")
            return
        }

        if (!this.playbackInfo.mkvMetadata?.subtitleTracks) {
            log.info("No subtitle tracks found")
            return
        }

        log.info("Adding subtitle tracks", this.playbackInfo.mkvMetadata.subtitleTracks)
        for (const track of this.playbackInfo.mkvMetadata.subtitleTracks) {
            this.videoElement.addTextTrack("subtitles", track.name, track.language)
        }

        this._initLibassRenderer()
    }

    selectTrack(trackLabel: string) {
        const track = this.playbackInfo.mkvMetadata?.subtitleTracks?.find(t => t.name === trackLabel)
        log.info("Selecting track", trackLabel, track)
        if (track) {
            this.currentTrack = track.number
            this.libassRenderer?.setTrack(track.codecPrivate || "")
            const existingEvents = this.subtitleRecord[track.number] || []
            for (const event of existingEvents) {
                this.libassRenderer?.createEvent(this._createAssEvent(event))
            }
        } else {
            log.error("Track not found", trackLabel)
        }
    }

    //
    onSubtitleEvent(event: MKVParser_SubtitleEvent) {
        const hadEvent = this._recordSubtitleEvent(event)
        if (this.libassRenderer && !hadEvent) {
            this.libassRenderer.createEvent(this._createAssEvent(event))
        }
    }

    //
    // Events

    terminate() {
        this.libassRenderer?.destroy()
        this.libassRenderer = null
        this.subtitleRecord = {}
        this.currentTrack = null
    }

    private _initLibassRenderer() {
        if (!this.libassRenderer) {
            log.info("Initializing libass renderer")

            const legacyWasmUrl = process.env.NODE_ENV === "development"
                ? "/jassub/jassub-worker.wasm.js" : legacy_getAssetUrl("/jassub/jassub-worker.wasm.js")

            this.libassRenderer = new JASSUB({
                video: this.videoElement,
                wasmUrl: "/jassub/jassub-worker.wasm",
                workerUrl: "/jassub/jassub-worker.js",
                legacyWasmUrl: legacyWasmUrl,
                // Both parameters needed for subs to work on iOS, ref: jellyfin-vue
                offscreenRender: isApple() ? false : this.jassubOffscreenRender, // should be false for iOS
                prescaleFactor: 0.8,
                onDemandRender: false,
                fonts: this.defaultFonts,
                fallbackFont: this.defaultFonts[0],
            })

        }
    }

    private _createAssEvent(event: MKVParser_SubtitleEvent): ASS_Event {
        return {
            Start: event.startTime,
            Duration: event.duration,
            Style: event.extraData?.style ?? "",
            Name: event.extraData?.name ?? "",
            MarginL: event.extraData?.marginL ? Number(event.extraData.marginL) : 0,
            MarginR: event.extraData?.marginR ? Number(event.extraData.marginR) : 0,
            MarginV: event.extraData?.marginV ? Number(event.extraData.marginV) : 0,
            Effect: event.extraData?.effect ?? "",
            Text: event.text,
            ReadOrder: event.extraData?.readOrder ? Number(event.extraData.readOrder) : 0,
            Layer: event.extraData?.layer ? Number(event.extraData.layer) : 0,
            _index: this.subtitleRecord[event.trackNumber].length,
        }
    }

    private _recordSubtitleEvent(event: MKVParser_SubtitleEvent): boolean {
        let hadEvent = false
        if (!this.subtitleRecord[event.trackNumber]) {
            log.info("Storing new subtitle track", event.trackNumber)
            this.subtitleRecord[event.trackNumber] = []
            this.subtitleSet[event.trackNumber] = new Set()
        }

        const eventKey = JSON.stringify(event)

        if (!this.subtitleSet[event.trackNumber].has(eventKey)) {
            hadEvent = true
        }

        this.subtitleRecord[event.trackNumber].push(event)
        this.subtitleSet[event.trackNumber].add(eventKey)
        return hadEvent
    }


}

export function useHandleNativePlayer() {


}
