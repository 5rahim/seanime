import { getServerBaseUrl } from "@/api/client/server-url"
import { MKVParser_SubtitleEvent, MKVParser_TrackInfo, NativePlayer_PlaybackInfo } from "@/api/generated/types"
import { logger } from "@/lib/helpers/debug"
import { legacy_getAssetUrl } from "@/lib/server/assets"
import { isApple } from "@/lib/utils/browser-detection"
import JASSUB, { ASS_Event, JassubOptions } from "jassub"
import { toast } from "sonner"
import { NativePlayerSettings } from "./native-player.atoms"

const subtitleLog = logger("SUBTITLE")
const audioLog = logger("AUDIO")

const NO_TRACK_NUMBER = -1

const DEFAULT_SUBTITLE_HEADER = `[Script Info]
Title: English (US)
ScriptType: v4.00+
WrapStyle: 0
PlayResX: 640
PlayResY: 360
ScaledBorderAndShadow: yes

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
Style: Default, Roboto Medium,24,&H00FFFFFF,&H000000FF,&H00000000,&H00000000,0,0,0,0,100,100,0,0,1,1.3,0,2,20,20,23,0
[Events]

`

export class StreamSubtitleManager {
    private readonly videoElement: HTMLVideoElement
    private readonly jassubOffscreenRender: boolean
    private libassRenderer: JASSUB | null = null
    private settings: NativePlayerSettings

    private tracks: Record<string, {
        info: MKVParser_TrackInfo
        events: Map<string, { event: MKVParser_SubtitleEvent, assEvent: ASS_Event }>
        styles: Record<string, number>
    }> = {}

    private playbackInfo: NativePlayer_PlaybackInfo
    private currentTrackNumber: number = NO_TRACK_NUMBER
    private fonts: string[] = []

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
        this.playbackInfo = playbackInfo
        this.settings = settings

        if (!this.playbackInfo?.mkvMetadata?.subtitleTracks) return

        // Store the tracks
        for (const track of this.playbackInfo.mkvMetadata.subtitleTracks) {
            this._addTrack(track)
        }
        this._storeTrackStyles() // Store their styles
        this._selectDefaultTrack()

        subtitleLog.info("Text tracks", this.videoElement.textTracks)
        subtitleLog.info("Tracks", this.tracks)
    }

    // Selects a track by its label.
    selectTrackByLabel(trackLabel: string) {
        const track = this.playbackInfo.mkvMetadata?.subtitleTracks?.find?.(t => t.name === trackLabel)
        if (track) {
            this.selectTrack(track.number)
        } else {
            subtitleLog.error("Track not found", trackLabel)
            this.setNoTrack()
        }
    }

    // Sets the track to no track.
    setNoTrack() {
        this.currentTrackNumber = NO_TRACK_NUMBER
        this.libassRenderer?.setTrack(DEFAULT_SUBTITLE_HEADER)
        this.libassRenderer?.resize?.()
    }

    // Selects a track by its number.
    selectTrack(trackNumber: number) {
        subtitleLog.info("Track selection requested", trackNumber)
        this._init()

        if (this.currentTrackNumber === trackNumber) {
            subtitleLog.info("Track already selected", trackNumber)
            return
        }

        if (trackNumber === NO_TRACK_NUMBER) {
            subtitleLog.info("No track selected", trackNumber)
            this.setNoTrack()
            return
        }

        const track = this._getTracks()?.find?.(t => t.info.number === trackNumber)
        subtitleLog.info("Selecting track", trackNumber, track)

        // Update video element's textTracks to reflect the selection in media-chrome
        if (this.videoElement.textTracks) {
            subtitleLog.info("Updating video element's textTracks", this.videoElement.textTracks)
            for (const textTrack of this.videoElement.textTracks) {
                if (track && textTrack.id === track.info.number.toString()) {
                    textTrack.mode = "showing"
                } else {
                    textTrack.mode = "disabled"
                }
            }
        }

        if (!track) {
            subtitleLog.info("Track not found", trackNumber)
            this.setNoTrack()
            return
        }

        const codecPrivate = track.info.codecPrivate?.slice?.(0, -1) || DEFAULT_SUBTITLE_HEADER

        this.currentTrackNumber = track.info.number // update the current track number

        // Set the track
        this.libassRenderer?.setTrack(codecPrivate)
        const trackEventMap = this.tracks[track.info.number]?.events
        if (!trackEventMap) {
            return
        }
        subtitleLog.info("Found", trackEventMap.size, "events for track", track.info.number)

        // Add the events to the libass renderer
        for (const { assEvent } of trackEventMap.values()) {
            this.libassRenderer?.createEvent(assEvent)
        }

        this.libassRenderer?.resize?.()
    }

    // This will record the events and add them to the libass renderer if they are new.
    onSubtitleEvent(event: MKVParser_SubtitleEvent) {
        // Record the event
        const { isNew, assEvent } = this._recordSubtitleEvent(event)
        // subtitleLog.info("Subtitle event received", event.trackNumber, this.currentTrackNumber)

        if (!assEvent) return

        // if the event is new and is from the selected track, add it to the libass renderer
        if (this.libassRenderer && isNew && event.trackNumber === this.currentTrackNumber) {
            // console.log("Creating event", event.text)
            // console.table(assEvent)
            this.libassRenderer.createEvent(assEvent)
        }
    }

    onTrackAdded(track: MKVParser_TrackInfo) {
        subtitleLog.info("Subtitle track added", track)
        toast.info(`Subtitle track added: ${track.name}`)
        this._addTrack(track)
        this._storeTrackStyles()
        // Add the track to the video element
        const trackEl = document.createElement("track")
        trackEl.id = track.number.toString()
        trackEl.kind = "subtitles"
        trackEl.label = track.name || ""
        trackEl.srclang = track.language || "eng"
        this.videoElement.appendChild(trackEl)
        this._selectDefaultTrack()
        this._init()
        this.libassRenderer?.resize?.()

    }

    terminate() {
        this.libassRenderer?.destroy()
        this.libassRenderer = null
        for (const trackNumber in this.tracks) {
            this.tracks[trackNumber].events.clear()
        }
        this.tracks = {}
        this.currentTrackNumber = NO_TRACK_NUMBER
    }

    // ----------- Private methods ----------- //

    //
    // Selects a track to be used.
    // This should be called after the tracks are loaded.
    // When called for the first time, it will initialize the libass renderer.
    //
    private _selectDefaultTrack() {
        if (this.currentTrackNumber !== NO_TRACK_NUMBER) return
        const tracks = this._getTracks()

        if (!tracks?.length) {
            this.setNoTrack()
            return
        }

        if (tracks.length === 1) {
            this.selectTrack(tracks[0].info.number)
            return
        }

        const foundTracks = tracks?.filter?.(t => t.info.language === this.settings.preferredSubtitleLanguage)
        if (foundTracks?.length) {
            // Find default or forced track
            const defaultIndex = foundTracks.findIndex(t => t.info.forced)
            this.selectTrack(foundTracks[defaultIndex >= 0 ? defaultIndex : 0].info.number)
            return
        }

        // No default tracks found, select the english track
        const englishTracks = tracks?.filter?.(t => (t.info.language || "eng") === "eng" || t.info.language === "en")
        if (englishTracks?.length) {
            const defaultIndex = englishTracks.findIndex(t => t.info.forced || t.info.default)
            this.selectTrack(englishTracks[defaultIndex >= 0 ? defaultIndex : 0].info.number)
            return
        }

        // No tracks found, select the first track
        this.selectTrack(tracks?.[0]?.info?.number ?? NO_TRACK_NUMBER)
    }

    private _addTrack(track: MKVParser_TrackInfo) {
        this.tracks[track.number] = {
            info: track,
            events: new Map(),
            styles: {},
        }
        return this.tracks[track.number]
    }

    private _getTracks() {
        return Object.values(this.tracks).sort((a, b) => a.info.number - b.info.number)
    }

    //
    // Stores the styles for each track.
    //
    private _storeTrackStyles() {
        if (!this.playbackInfo?.mkvMetadata?.subtitleTracks) return
        for (const track of this.playbackInfo.mkvMetadata.subtitleTracks) {
            const codecPrivate = track.codecPrivate?.slice?.(0, -1) || DEFAULT_SUBTITLE_HEADER
            const lines = codecPrivate.replaceAll("\r\n", "\n").split("\n").filter(line => line.startsWith("Style:"))
            let index = 1
            const s: Record<string, number> = {}
            this.tracks[track.number].styles = s // reset styles
            for (const line of lines) {
                let styleName = line.split("Style:")[1]
                styleName = (styleName.split(",")[0] || "").trim()
                if (styleName && !s[styleName]) {
                    s[styleName] = index++
                }
            }
            this.tracks[track.number].styles = s
        }
    }

    private __eventMapKey(event: MKVParser_SubtitleEvent): string {
        return `${event.trackNumber}-${event.startTime}-${event.duration}-${event.extraData?.style}-${event.extraData?.name}-${event.extraData?.marginL}-${event.extraData?.marginR}-${event.extraData?.marginV}-${event.extraData?.effect}-${event.extraData?.readOrder}-${event.extraData?.layer}`
    }

    private _init() {
        if (!!this.libassRenderer) return

        subtitleLog.info("Initializing libass renderer")

        const wasmUrl = new URL("/jassub/jassub-worker.wasm", window.location.origin).toString()
        const workerUrl = new URL("/jassub/jassub-worker.js", window.location.origin).toString()
        // const legacyWasmUrl = new URL("/jassub/jassub-worker.wasm.js", window.location.origin).toString()
        const modernWasmUrl = new URL("/jassub/jassub-worker-modern.wasm", window.location.origin).toString()

        const legacyWasmUrl = process.env.NODE_ENV === "development"
            ? "/jassub/jassub-worker.wasm.js" : legacy_getAssetUrl("/jassub/jassub-worker.wasm.js")

        const defaultFontUrl = "/jassub/Roboto-Medium.ttf"

        this.libassRenderer = new JASSUB({
            video: this.videoElement,
            subContent: DEFAULT_SUBTITLE_HEADER, // needed
            // subUrl: new URL("/jassub/test.ass", window.location.origin).toString(),
            wasmUrl: wasmUrl,
            workerUrl: workerUrl,
            legacyWasmUrl: legacyWasmUrl,
            modernWasmUrl: modernWasmUrl,
            // Both parameters needed for subs to work on iOS, ref: jellyfin-vue
            offscreenRender: isApple() ? false : this.jassubOffscreenRender, // should be false for iOS
            prescaleFactor: 0.8,
            onDemandRender: false,
            fonts: this.fonts,
            fallbackFont: "roboto medium",
            availableFonts: {
                "roboto medium": defaultFontUrl,
            },
            libassGlyphLimit: 60500,
        })

        this.fonts = this.playbackInfo.mkvMetadata?.attachments?.filter(a => a.type === "font")
            ?.map(a => `${getServerBaseUrl()}/api/v1/directstream/att/${a.filename}`) || []

        this.fonts = [defaultFontUrl, ...this.fonts]

        for (const font of this.fonts) {
            this.libassRenderer.addFont(font)
        }
    }

    private _createAssEvent(event: MKVParser_SubtitleEvent, index: number): ASS_Event {
        return {
            Start: event.startTime,
            Duration: event.duration,
            Style: String(event.extraData?.style ? this.tracks[event.trackNumber]?.styles?.[event.extraData?.style ?? "Default"] : 1),
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


    private _recordSubtitleEvent(event: MKVParser_SubtitleEvent): { isNew: boolean, assEvent: ASS_Event | null } {
        const trackEventMap = this.tracks[event.trackNumber]?.events // get the map
        if (!trackEventMap) {
            return { isNew: false, assEvent: null }
        }

        const eventKey = this.__eventMapKey(event)

        if (event.text.includes("never imagined something")) {
            subtitleLog.info("KEY", eventKey, "isNew", !trackEventMap.has(eventKey))
        }

        // Check if the event is already in the record
        // If it is, return false
        if (trackEventMap.has(eventKey)) {
            return { isNew: false, assEvent: trackEventMap.get(eventKey)?.assEvent! }
        }

        // record the event
        const assEvent = this._createAssEvent(event, trackEventMap.size)
        trackEventMap.set(eventKey, { event, assEvent })
        return { isNew: true, assEvent }
    }
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

        // Dispatch change event to notify media-chrome
        if (trackChanged && this.videoElement.audioTracks.dispatchEvent) {
            this.videoElement.audioTracks.dispatchEvent(new Event("change"))
        }
    }

}
