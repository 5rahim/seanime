import { getServerBaseUrl } from "@/api/client/server-url"
import { MKVParser_SubtitleEvent, NativePlayer_PlaybackInfo } from "@/api/generated/types"
import { logger } from "@/lib/helpers/debug"
import { legacy_getAssetUrl } from "@/lib/server/assets"
import { isApple } from "@/lib/utils/browser-detection"
import JASSUB, { ASS_Event, JassubOptions } from "jassub"
import { NativePlayerSettings } from "./native-player.atoms"

const log = logger("STREAM SUBTITLE MANAGER")

const NO_TRACK = -1

const DUMMY_TRACK_HEADER = `[Script Info]
Title: English (US)
ScriptType: v4.00+
WrapStyle: 0
PlayResX: 1280
PlayResY: 720
ScaledBorderAndShadow: yes

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
Style: Default, Roboto Medium,52,&H00FFFFFF,&H00FFFFFF,&H00000000,&H00000000,0,0,0,0,100,100,0,0,1,2.6,0,2,20,20,46,1
[Events]

`

export class StreamSubtitleManager {
    // Video element
    private videoElement: HTMLVideoElement
    // LibASS renderer
    private libassRenderer: JASSUB | null = null
    // JASSUB offscreen render
    private jassubOffscreenRender: boolean
    a = 0
    // Subtitles for each track
    // private subtitleRecord: Record<number, MKVParser_SubtitleEvent[]> = {}
    // private subtitleSet: Record<string, Set<string>> = {}

    // Track the subtitle events for each track
    // Settings
    private settings: NativePlayerSettings
    // Record<trackNumber, Map<eventKey, ASS_Event>>
    private subtitleTrackMap: Record<number, Map<string, { event: MKVParser_SubtitleEvent, assEvent: ASS_Event }>> = {}
    // Playback info
    private playbackInfo: NativePlayer_PlaybackInfo
    // Current subtitle track number
    private currentTrackNumber: number = NO_TRACK

    constructor({
        videoElement,
        jassubOffscreenRender,
        playbackInfo,
        settings,
    }: {
        videoElement: HTMLVideoElement
        jassubOffscreenRender: boolean
        playbackInfo: NativePlayer_PlaybackInfo
        settings: NativePlayerSettings
    }) {
        this.videoElement = videoElement
        this.jassubOffscreenRender = jassubOffscreenRender
        this.libassRenderer = null
        this.subtitleTrackMap = {}
        this.playbackInfo = playbackInfo
        this.settings = settings

        // Select the default track
        this._selectDefaultTrack()
    }

    //
    // Selects a track to be used.
    // This should be called after the tracks are loaded.
    // When called for the first time, it will initialize the libass renderer.

    _selectDefaultTrack() {
        const foundTracks = this.playbackInfo.mkvMetadata?.subtitleTracks?.filter?.(t => t.language === this.settings.preferredSubtitleLanguage)
        if (foundTracks?.length) {
            // Find default or forced track
            const defaultIndex = foundTracks.findIndex(t => t.forced)
            this.selectTrack(foundTracks[defaultIndex >= 0 ? defaultIndex : 0].number)
            return
        }

        // No default tracks found, select the english track
        const englishTracks = this.playbackInfo.mkvMetadata?.subtitleTracks?.filter?.(t => (t.language || "eng") === "eng")
        if (englishTracks?.length) {
            const defaultIndex = englishTracks.findIndex(t => t.forced || t.default)
            this.selectTrack(englishTracks[defaultIndex >= 0 ? defaultIndex : 0].number)
            return
        }

        // No tracks found, select the first track
        this.selectTrack(this.playbackInfo.mkvMetadata?.subtitleTracks?.[0]?.number || NO_TRACK)
    }

    //
    selectTrackByLabel(trackLabel: string) {
        const track = this.playbackInfo.mkvMetadata?.subtitleTracks?.find?.(t => t.name === trackLabel)
        if (track) {
            this.selectTrack(track.number)
        } else {
            log.error("Track not found", trackLabel)
            // If track not found, disable all tracks
            this.selectTrack(NO_TRACK)
        }
    }

    // Called when the server sends a subtitle event.

    selectTrack(trackNumber: number) {
        this._initLibassRenderer()

        const track = this.playbackInfo.mkvMetadata?.subtitleTracks?.find?.(t => t.number === trackNumber)
        log.info("Selecting track", trackNumber, track)

        // Update video element's textTracks to reflect the selection in media-chrome
        if (this.videoElement.textTracks) {
            for (let i = 0; i < this.videoElement.textTracks.length; i++) {
                const textTrack = this.videoElement.textTracks[i]
                if (track && textTrack.label === track.name) {
                    textTrack.mode = "showing"
                } else {
                    textTrack.mode = "disabled"
                }
            }
        }

        if (!track) {
            this.currentTrackNumber = NO_TRACK
            this.libassRenderer?.setTrack(DUMMY_TRACK_HEADER)
            return
        }

        this.currentTrackNumber = track.number
        // const codecPrivate = track.codecPrivate?.replace("Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text", "")
        console.log("Setting track", track.codecPrivate, this.videoElement)
        // Set the track
        this.libassRenderer?.setTrack(track.codecPrivate || "")
        const subtitleTrackMap = this._getSubtitleTrackMap(track.number)
        log.info("Found", subtitleTrackMap.size, "events for track", track.number)

        // Add the events to the libass renderer
        for (const { assEvent } of subtitleTrackMap.values()) {
            this.libassRenderer?.createEvent(assEvent)
            if (this.a <= 10) {
                console.log(assEvent)
            }
            this.a++
        }

        this.libassRenderer?.resize?.()
    }

    // This will record the events and add them to the libass renderer if they are new.
    onSubtitleEvent(event: MKVParser_SubtitleEvent) {
        // Record the event
        const { isNew, assEvent } = this._recordSubtitleEvent(event)
        // if the event is new and is from the selected track, add it to the libass renderer
        // log.info("Subtitle event received", event.trackNumber, this.currentTrackNumber, isNew, assEvent.Start, assEvent.Text)

        if (!this.libassRenderer || !isNew || event.trackNumber !== this.currentTrackNumber) return

        this.libassRenderer.createEvent(assEvent)
    }

    // ----------- Private methods ----------- //

    // Returns the subtitle track map for the given track number

    terminate() {
        this.libassRenderer?.destroy()
        this.libassRenderer = null
        for (const trackNumber in this.subtitleTrackMap) {
            this.subtitleTrackMap[trackNumber].clear()
        }
        this.subtitleTrackMap = {}
        this.currentTrackNumber = NO_TRACK
    }

    // If the track map does not exist, it will be createdq
    private _getSubtitleTrackMap(trackNumber: number): Map<string, { event: MKVParser_SubtitleEvent, assEvent: ASS_Event }> {
        if (!this.subtitleTrackMap[trackNumber]) {
            this.subtitleTrackMap[trackNumber] = new Map()
        }
        return this.subtitleTrackMap[trackNumber]
    }

    private __getSubtitleTrackMapKey(event: MKVParser_SubtitleEvent): string {
        return JSON.stringify(event)
    }

    private _initLibassRenderer() {
        if (!this.libassRenderer) {
            log.info("Initializing libass renderer")

            const wasmUrl = new URL("/jassub/jassub-worker.wasm", window.location.origin).toString()
            const workerUrl = new URL("/jassub/jassub-worker.js", window.location.origin).toString()
            // const legacyWasmUrl = new URL("/jassub/jassub-worker.wasm.js", window.location.origin).toString()
            const modernWasmUrl = new URL("/jassub/jassub-worker-modern.wasm", window.location.origin).toString()

            const fonts = this.playbackInfo.mkvMetadata?.attachments?.filter(a => a.type === "font")
                ?.map(a => `${getServerBaseUrl()}/api/v1/directstream/att/${a.filename}`) || []

            log.info("Fonts", fonts)

            // Extracted fonts
            let availableFonts: Record<string, string> = {}
            let firstFont = ""
            if (!!fonts?.length) {
                for (const font of fonts) {
                    const name = font.split("/").pop()?.split(".")[0]
                    if (name) {
                        if (!firstFont) {
                            firstFont = name.toLowerCase()
                        }
                        availableFonts[name.toLowerCase()] = font
                    }
                }
            }

            // Fallback font if no fonts are available
            if (!firstFont) {
                firstFont = "Roboto Medium"
            }
            if (Object.keys(availableFonts).length === 0) {
                availableFonts = {
                    "Roboto Medium": process.env.NODE_ENV !== "development"
                        ? getServerBaseUrl() + `/jassub/Roboto-Medium.ttf`
                        : "/jassub/Roboto-Medium.ttf",
                }
            }

            const legacyWasmUrl = process.env.NODE_ENV === "development"
                ? "/jassub/jassub-worker.wasm.js" : legacy_getAssetUrl("/jassub/jassub-worker.wasm.js")

            this.libassRenderer = new JASSUB({
                video: this.videoElement,
                subContent: DUMMY_TRACK_HEADER, // needed
                wasmUrl: wasmUrl,
                workerUrl: workerUrl,
                legacyWasmUrl: legacyWasmUrl,
                modernWasmUrl: modernWasmUrl,
                // Both parameters needed for subs to work on iOS, ref: jellyfin-vue
                offscreenRender: isApple() ? false : this.jassubOffscreenRender, // should be false for iOS
                prescaleFactor: 0.8,
                onDemandRender: false,
                fonts: fonts,
                availableFonts: availableFonts,
                fallbackFont: firstFont,
                libassMemoryLimit: 1024,
                libassGlyphLimit: 80000,
            })
        }
    }

    private _createAssEvent(event: MKVParser_SubtitleEvent, index: number): ASS_Event {
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
            ReadOrder: event.extraData?.readOrder ? Number(event.extraData.readOrder) : 1,
            Layer: event.extraData?.layer ? Number(event.extraData.layer) : 0,
            // index is based on the order of the events in the record
            _index: index,
        }
    }


    // Adds the event to the record and returns true if it's new.
    // Returns false if the event is already in the record.
    private _recordSubtitleEvent(event: MKVParser_SubtitleEvent): { isNew: boolean, assEvent: ASS_Event } {
        const subtitleTrackMap = this._getSubtitleTrackMap(event.trackNumber)

        const eventKey = this.__getSubtitleTrackMapKey(event)

        // Check if the event is already in the record
        // If it is, return false
        if (subtitleTrackMap.has(eventKey)) {
            return { isNew: false, assEvent: subtitleTrackMap.get(eventKey)?.assEvent! }
        }

        // record the event
        const assEvent = this._createAssEvent(event, subtitleTrackMap.size)
        subtitleTrackMap.set(eventKey, { event, assEvent })
        return { isNew: true, assEvent }
    }


}

export function useHandleNativePlayer() {


}

export class StreamAudioManager {

    onError: (error: string) => void
    private videoElement: HTMLVideoElement
    private settings: NativePlayerSettings
    // Playback info
    private playbackInfo: NativePlayer_PlaybackInfo

    constructor({
        videoElement,
        settings,
        playbackInfo,
        onError,
    }: {
        videoElement: HTMLVideoElement
        settings: NativePlayerSettings
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
            log.error("Audio track not found", trackLabel)
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

        // Dispatch change event to notify media-chrome
        if (trackChanged && this.videoElement.audioTracks.dispatchEvent) {
            this.videoElement.audioTracks.dispatchEvent(new Event("change"))
        }
    }

}
