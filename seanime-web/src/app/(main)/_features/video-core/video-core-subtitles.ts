import { getServerBaseUrl } from "@/api/client/server-url"
import { MKVParser_SubtitleEvent, MKVParser_TrackInfo } from "@/api/generated/types"
import { VideoCorePgsRenderer } from "@/app/(main)/_features/video-core/video-core-pgs-renderer"
import { vc_getSubtitleStyle } from "@/app/(main)/_features/video-core/video-core-settings-menu"
import { VideoCore_VideoPlaybackInfo, VideoCore_VideoSubtitleTrack, VideoCoreSettings } from "@/app/(main)/_features/video-core/video-core.atoms"
import { logger } from "@/lib/helpers/debug"
import { getAssetUrl, legacy_getAssetUrl } from "@/lib/server/assets"
import JASSUB, { ASS_Event } from "jassub"
import { toast } from "sonner"

const subtitleLog = logger("VIDEO CORE SUBTITLE")

const NO_TRACK_NUMBER = -1
const DEFAULT_FONT_NAME = "roboto medium"

function hexToASSColor(hex: string, alpha: number = 0): number {
    hex = hex.replace(/^#/, "")
    if (hex.length === 3) {
        hex = hex.split("").map(c => c + c).join("")
    }
    const val = parseInt(hex, 16)
    const r = (val >> 16) & 0xFF
    const g = (val >> 8) & 0xFF
    const b = val & 0xFF
    return ((r << 24) | (g << 16) | (b << 8) | alpha) >>> 0
}

function isPGS(str: string) {
    return str === "S_HDMV/PGS"
}

// Event or file track info.
export type NormalizedTrackInfo = {
    type: "event" | "file"
    language?: string
    languageIETF?: string
    codecID?: string
    label?: string
    number: number
    forced: boolean
    default: boolean
}

export type SubtitleManagerTrackSelectedEvent = CustomEvent<{ trackNumber: number, kind: "file" | "event" }>
export type SubtitleManagerTrackDeselectedEvent = CustomEvent
export type SubtitleManagerTrackAddedEvent = CustomEvent<{ track: NormalizedTrackInfo }>
export type SubtitleManagerTracksLoadedEvent = CustomEvent<{ tracks: NormalizedTrackInfo[] }>
export type SubtitleManagerDestroyedEvent = CustomEvent
export type SubtitleManagerSettingsUpdatedEvent = CustomEvent<{ settings: VideoCoreSettings }>

interface VideoCoreSubtitleManagerEventMap {
    "trackselected": SubtitleManagerTrackSelectedEvent
    "trackdeselected": SubtitleManagerTrackDeselectedEvent
    "trackadded": SubtitleManagerTrackAddedEvent
    "tracksloaded": SubtitleManagerTracksLoadedEvent
    "destroyed": SubtitleManagerDestroyedEvent
    "settingsupdated": SubtitleManagerSettingsUpdatedEvent
}


// Manages ASS and PGS subtitle streams.
export class VideoCoreSubtitleManager extends EventTarget {
    private readonly videoElement: HTMLVideoElement
    private readonly jassubOffscreenRender: boolean
    libassRenderer: JASSUB | null = null
    pgsRenderer: VideoCorePgsRenderer | null = null
    private settings: VideoCoreSettings
    private defaultSubtitleHeader = `[Script Info]
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

    // Event-based tracks
    private eventTracks: Record<string, {
        info: MKVParser_TrackInfo
        events: Map<string, { event: MKVParser_SubtitleEvent, assEvent: ASS_Event }>
        styles: Record<string, number>
    }> = {}

    // PGS event-based tracks
    private pgsEventTracks: Record<string, {
        info: MKVParser_TrackInfo
        events: Map<string, MKVParser_SubtitleEvent>
    }> = {}

    // URL-based tracks (will use internal API to convert to ASS)
    private fileTracks: Record<string, {
        info: VideoCore_VideoSubtitleTrack
        content: string | null // converted content
    }> = {}

    private readonly fetchAndConvertToASS?: (url?: string, content?: string) => Promise<string | undefined>

    private playbackInfo: VideoCore_VideoPlaybackInfo
    private currentTrackNumber: number = NO_TRACK_NUMBER
    private fonts: string[] = []

    private _onSelectedTrackChanged?: (track: number | null) => void
    private _onTracksLoaded?: (tracks: NormalizedTrackInfo[]) => void

    constructor({
        videoElement,
        jassubOffscreenRender,
        playbackInfo,
        settings,
        fetchAndConvertToASS,
    }: {
        videoElement: HTMLVideoElement
        jassubOffscreenRender: boolean
        playbackInfo: VideoCore_VideoPlaybackInfo
        settings: VideoCoreSettings
        fetchAndConvertToASS?: (url?: string, content?: string) => Promise<string | undefined>
    }) {
        super()
        this.videoElement = videoElement
        this.jassubOffscreenRender = jassubOffscreenRender
        this.playbackInfo = playbackInfo
        this.settings = settings
        this.fetchAndConvertToASS = fetchAndConvertToASS

        /*
         * Event Tracks
         */
        if (this.playbackInfo?.mkvMetadata?.subtitleTracks) {
            for (const track of this.playbackInfo.mkvMetadata.subtitleTracks) {
                this._addEventTrack(track)
            }
            this._storeEventTrackStyles()
        }

        /*
         * File Tracks
         */
        if (this.playbackInfo?.subtitleTracks) {
            let trackNumber = 1000
            for (const track of this.playbackInfo.subtitleTracks) {
                if (track.useLibassRenderer) {
                    this.fileTracks[trackNumber] = {
                        info: track,
                        content: null,
                    }
                    trackNumber++
                }
            }
        }

        this._onTracksLoaded?.(this._getTracks())

        // Select default track if we have any tracks
        if (this.playbackInfo?.mkvMetadata?.subtitleTracks || Object.keys(this.fileTracks).length > 0) {
            this._selectDefaultTrack()
        }

        // Apply subtitle delay from settings
        this.setSubtitleDelay(settings.subtitleDelay)

        subtitleLog.info("Text tracks", this.videoElement.textTracks)
        subtitleLog.info("Event Tracks", this.eventTracks)
        subtitleLog.info("PGS Event Tracks", this.pgsEventTracks)
        subtitleLog.info("File tracks", this.fileTracks)
    }

    addEventListener<K extends keyof VideoCoreSubtitleManagerEventMap>(
        type: K,
        listener: (this: VideoCoreSubtitleManager, ev: VideoCoreSubtitleManagerEventMap[K]) => any,
        options?: boolean | AddEventListenerOptions,
    ): void
    addEventListener(
        type: string,
        listener: EventListenerOrEventListenerObject,
        options?: boolean | AddEventListenerOptions,
    ): void

    addEventListener(
        type: string,
        listener: EventListenerOrEventListenerObject,
        options?: boolean | AddEventListenerOptions,
    ): void {
        super.addEventListener(type, listener, options)
    }

    removeEventListener<K extends keyof VideoCoreSubtitleManagerEventMap>(
        type: K,
        listener: (this: VideoCoreSubtitleManager, ev: VideoCoreSubtitleManagerEventMap[K]) => any,
        options?: boolean | EventListenerOptions,
    ): void
    removeEventListener(
        type: string,
        listener: EventListenerOrEventListenerObject,
        options?: boolean | EventListenerOptions,
    ): void

    removeEventListener(
        type: string,
        listener: EventListenerOrEventListenerObject,
        options?: boolean | EventListenerOptions,
    ): void {
        super.removeEventListener(type, listener, options)
    }

    getSelectedTrackNumberOrNull(): number | null {
        if (this.currentTrackNumber === NO_TRACK_NUMBER) return null
        return this.currentTrackNumber
    }

    //////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

    // Sets the track to no track.
    setNoTrack() {
        this.currentTrackNumber = NO_TRACK_NUMBER
        this.libassRenderer?.setTrack(this.defaultSubtitleHeader)
        this.libassRenderer?.resize?.()
        this.pgsRenderer?.clear()
        this._onSelectedTrackChanged?.(NO_TRACK_NUMBER)

        const event: SubtitleManagerTrackDeselectedEvent = new CustomEvent("trackdeselected")
        this.dispatchEvent(event)
    }

    setTrackChangedEventListener(callback: (track: number | null) => void) {
        this._onSelectedTrackChanged = callback
    }

    setTracksLoadedEventListener(callback: ((tracks: NormalizedTrackInfo[]) => void)) {
        this._onTracksLoaded = callback
    }

    // Selects a track by its number.
    selectTrack(trackNumber: number) {
        subtitleLog.info("Track selection requested", trackNumber)
        this._init()

        // if (this.currentTrackNumber === trackNumber) {
        //     subtitleLog.info("Track already selected", trackNumber)
        //     return
        // }

        if (trackNumber === NO_TRACK_NUMBER) {
            subtitleLog.info("No track selected", trackNumber)
            this.setNoTrack()
            return
        }

        const track = this._getTracks()?.find?.(t => t.number === trackNumber)
        subtitleLog.info("Selecting track", trackNumber, track)

        // Update the text track which is showing in the video element
        if (this.videoElement.textTracks) {
            subtitleLog.info("Updating video element's textTracks", this.videoElement.textTracks)
            for (const textTrack of this.videoElement.textTracks) {
                if (track && textTrack.id === track.number.toString()) {
                    textTrack.mode = "showing"
                } else {
                    textTrack.mode = "disabled"
                }
            }
            // Dispatch a change event to update the player
            this.videoElement.textTracks.dispatchEvent(new Event("change"))
        }

        if (!track) {
            subtitleLog.error("Track not found", trackNumber)
            this.setNoTrack()
            return
        }

        // Dispatch the selected track change event
        this._onSelectedTrackChanged?.(trackNumber)

        this.currentTrackNumber = track.number // update the current track number

        /*
         * File track
         */

        // Check if this is a file track
        // If it is, fetch/convert the content and add it to the libass renderer
        const fileTrack = this.fileTracks[trackNumber]
        if (fileTrack) {
            this._handleFileTrack(trackNumber, fileTrack)
            return
        }

        /*
         * Event track
         */

        const eventTrack = this.eventTracks[trackNumber]
        if (!eventTrack) {
            subtitleLog.warning("Event track not found", trackNumber)
            return
        }

        // Handle event track
        const codecPrivate = eventTrack.info.codecPrivate?.slice?.(0, -1) || this.defaultSubtitleHeader

        // Check if this is a PGS track
        if (isPGS(eventTrack.info.codecID)) {
            // Clear PGS renderer and libass
            this.pgsRenderer?.clear()
            this.libassRenderer?.setTrack(this.defaultSubtitleHeader)

            // Add all cached PGS events from the event map
            const pgsTrack = this.pgsEventTracks[track.number]
            if (pgsTrack?.events) {
                subtitleLog.info("Found", pgsTrack.events.size, "PGS events for track", track.number)
                for (const event of pgsTrack.events.values()) {
                    this._addPgsEvent(event)
                }
                this.pgsRenderer?.resize?.()
            } else {
                subtitleLog.warning("No PGS events found for track", track.number)
            }
        } else {
            // Handle regular ASS/text subtitles
            this.pgsRenderer?.clear()

            // Set the track
            this.libassRenderer?.setTrack(codecPrivate)

            // Apply customization to Default styles
            this._applySubtitleCustomization()

            const trackEventMap = this.eventTracks[track.number]?.events
            if (!trackEventMap) {
                return
            }
            subtitleLog.info("Found", trackEventMap.size, "events for track", track.number)

            // Add the cached events to the libass renderer
            for (const { assEvent } of trackEventMap.values()) {
                this.libassRenderer?.createEvent(assEvent)
            }

            this.libassRenderer?.resize?.()
        }

        const selectedEvent: SubtitleManagerTrackSelectedEvent = new CustomEvent("trackselected", { detail: { trackNumber, kind: "event" } })
        this.dispatchEvent(selectedEvent)
    }

    destroy() {
        subtitleLog.info("Destroying subtitle manager")
        this.libassRenderer?.destroy()
        this.libassRenderer = null
        this.pgsRenderer?.destroy()
        this.pgsRenderer = null
        for (const trackNumber in this.eventTracks) {
            this.eventTracks[trackNumber].events.clear()
        }
        this.eventTracks = {}
        for (const trackNumber in this.pgsEventTracks) {
            this.pgsEventTracks[trackNumber].events.clear()
        }
        this.pgsEventTracks = {}
        this.fileTracks = {}
        this.currentTrackNumber = NO_TRACK_NUMBER

        const event: SubtitleManagerDestroyedEvent = new CustomEvent("destroyed")
        this.dispatchEvent(event)
    }

    getTracks() {
        return this._getTracks()
    }

    getTrack(trackNumber: number | null) {
        return this._getTracks()?.find(t => t.number === (trackNumber ?? NO_TRACK_NUMBER))
    }

    getNextTrackNumber(trackNumber: number | null) {
        const tracks = this._getTracks()
        const nextTrackNumber = tracks.find(t => t.number > (trackNumber ?? NO_TRACK_NUMBER))?.number
        return nextTrackNumber ?? NO_TRACK_NUMBER
    }

    // Update settings and reapply subtitle customization to current track
    updateSettings(newSettings: VideoCoreSettings) {
        this.settings = newSettings
        // Apply subtitle delay
        this.setSubtitleDelay(newSettings.subtitleDelay)
        // Reapply customization if a track is currently selected
        if (this.currentTrackNumber !== NO_TRACK_NUMBER) {
            this._applySubtitleCustomization()
        }

        // Dispatch Settings Updated Event
        const event: SubtitleManagerSettingsUpdatedEvent = new CustomEvent("settingsupdated", { detail: { settings: newSettings } })
        this.dispatchEvent(event)
    }

    setSubtitleDelay(subtitleDelay: number) {
        if (this.libassRenderer) (this.libassRenderer as any).timeOffset = (-subtitleDelay)
        if (this.pgsRenderer) this.pgsRenderer.setTimeOffset(-subtitleDelay)
    }

    // This will record the events and add them to the libass renderer if they are new.
    onSubtitleEvent(event: MKVParser_SubtitleEvent) {
        // Check if this is a PGS event
        if (isPGS(event.codecID)) {
            this._handlePgsEvent(event)
            return
        }

        // Handle ASS events
        const { isNew, assEvent } = this._recordSubtitleEvent(event)

        if (!assEvent) return

        // if the event is new and is from the selected track, add it to the libass renderer
        if (this.libassRenderer && isNew && event.trackNumber === this.currentTrackNumber) {
            this.libassRenderer.createEvent(assEvent)
        }
    }

    getFileTrack(trackNumber: number) {
        return this.fileTracks[trackNumber] || null
    }

    private _init() {
        if (!this.libassRenderer) {
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
                subContent: this.defaultSubtitleHeader,
                wasmUrl: wasmUrl,
                workerUrl: workerUrl,
                legacyWasmUrl: legacyWasmUrl,
                modernWasmUrl: modernWasmUrl,
                // Both parameters needed for subs to work on iOS, ref: jellyfin-vue
                // offscreenRender: isApple() ? false : this.jassubOffscreenRender, // should be false for iOS
                offscreenRender: true,
                // onDemandRender: false,
                // prescaleFactor: 0.8,
                fonts: this.fonts,
                fallbackFont: DEFAULT_FONT_NAME,
                availableFonts: {
                    [DEFAULT_FONT_NAME]: defaultFontUrl,
                },
                libassGlyphLimit: 60500,
                libassMemoryLimit: 1024,
                dropAllBlur: true,
                debug: false,
            })

            this.fonts = this.playbackInfo.mkvMetadata?.attachments?.filter(a => a.type === "font")
                ?.map(a => `${getServerBaseUrl()}/api/v1/directstream/att/${a.filename}`) || []

            this.fonts = [defaultFontUrl, ...this.fonts]

            for (const font of this.fonts) {
                this.libassRenderer.addFont(font)
            }
        }

        if (!this.pgsRenderer && this.playbackInfo.mkvMetadata?.tracks?.some(t => isPGS(t.codecID))) {
            this.pgsRenderer = new VideoCorePgsRenderer({
                videoElement: this.videoElement,
                // debug: process.env.NODE_ENV === "development",
            })
        }
    }

    private _getTracks(): NormalizedTrackInfo[] {
        const eventTracks = Object.values(this.eventTracks).map(t => <NormalizedTrackInfo>({
            type: "event",
            language: t.info.language,
            number: t.info.number,
            label: t.info.name,
            forced: t.info.forced,
            default: t.info.default,
            languageIETF: t.info.languageIETF,
            codecID: t.info.codecID,
        }))

        const fileTracks = Object.entries(this.fileTracks).map(([trackNumber, t]) => <NormalizedTrackInfo>({
            type: "file",
            language: t.info.language,
            number: Number(trackNumber),
            label: t.info.label,
            forced: false,
            default: t.info.default,
        }))

        return [...eventTracks, ...fileTracks].sort((a, b) => a.number - b.number)
    }

    //////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////\

    // +-----------------------+
    // |      Event Tracks     |
    // +-----------------------+

    addEventTrack(track: MKVParser_TrackInfo) {
        subtitleLog.info("Subtitle track added", track)
        this._addEventTrack(track)
        this._storeEventTrackStyles()
        // Select the track
        this.selectTrack(track.number)
        this._init()
        this.libassRenderer?.resize?.()
        this.pgsRenderer?.resize()

        const tracks = this._getTracks()
        const normalizedTrack = tracks.find(t => t.number === track.number)
        if (normalizedTrack) {
            const event: SubtitleManagerTrackAddedEvent = new CustomEvent("trackadded", { detail: { track: normalizedTrack } })
            this.dispatchEvent(event)
        }
        const event: SubtitleManagerTracksLoadedEvent = new CustomEvent("tracksloaded", { detail: { tracks: tracks } })
        this.dispatchEvent(event)
        this._onTracksLoaded?.(tracks)
    }

    // When called for the first time, it will initialize the libass renderer.
    private _selectDefaultTrack() {
        if (this.currentTrackNumber !== NO_TRACK_NUMBER) {
            subtitleLog.warning("A track is already selected, cannot select default track")
            return
        }
        const tracks = this._getTracks()
        subtitleLog.info("Selecting default track", tracks)

        if (!tracks?.length) {
            this.setNoTrack()
            return
        }

        if (tracks.length === 1) {
            subtitleLog.info("Only one track found, selecting it")
            this.selectTrack(tracks[0].number)
            return
        }

        // Split preferred languages by comma and trim whitespace
        const defaultTrackNumber = getDefaultSubtitleTrackNumber(this.settings, tracks)
        subtitleLog.info("Default subtitle track number",
            defaultTrackNumber,
            this.settings.preferredSubtitleLanguage,
            this.settings.preferredSubtitleBlacklist)
        this.selectTrack(defaultTrackNumber)
    }

    private _handlePgsEvent(event: MKVParser_SubtitleEvent) {
        // Ensure the PGS track exists
        if (!this.pgsEventTracks[event.trackNumber]) {
            subtitleLog.warning("PGS track not initialized for track number", event.trackNumber)
            return
        }

        const trackEventMap = this.pgsEventTracks[event.trackNumber].events
        const eventKey = this._getPgsEventKey(event)

        // Check if the event is already recorded
        if (trackEventMap.has(eventKey)) {
            return
        }

        // Store the event
        trackEventMap.set(eventKey, event)

        // If this is the currently selected track, add the event to the renderer
        if (event.trackNumber === this.currentTrackNumber && this.pgsRenderer) {
            this._addPgsEvent(event)
        }
    }

    private _getPgsEventKey(event: MKVParser_SubtitleEvent): string {
        return `${event.startTime}-${event.duration}-${event.text.substring(0, 50)}`
    }

    private _addPgsEvent(event: MKVParser_SubtitleEvent) {
        if (!this.pgsRenderer) {
            return
        }

        const pgsEvent = {
            startTime: event.startTime / 1e3,
            duration: event.duration / 1e3,
            imageData: event.text, // base64 PNG
            width: parseInt(event.extraData?.width || "0", 10),
            height: parseInt(event.extraData?.height || "0", 10),
            x: event.extraData?.x ? parseInt(event.extraData.x, 10) : undefined,
            y: event.extraData?.y ? parseInt(event.extraData.y, 10) : undefined,
            canvasWidth: event.extraData?.canvas_width ? parseInt(event.extraData.canvas_width, 10) : undefined,
            canvasHeight: event.extraData?.canvas_height ? parseInt(event.extraData.canvas_height, 10) : undefined,
            cropX: event.extraData?.crop_x ? parseInt(event.extraData.crop_x, 10) : undefined,
            cropY: event.extraData?.crop_y ? parseInt(event.extraData.crop_y, 10) : undefined,
            cropWidth: event.extraData?.crop_width ? parseInt(event.extraData.crop_width, 10) : undefined,
            cropHeight: event.extraData?.crop_height ? parseInt(event.extraData.crop_height, 10) : undefined,
        }

        this.pgsRenderer.addEvent(pgsEvent)
    }

    private _applySubtitleCustomization() {
        if (!this.libassRenderer) {
            return
        }

        // Handle undefined or disabled customization
        if (!this.settings.subtitleCustomization?.enabled) {
            // Disable style override if customization is disabled
            this.libassRenderer.disableStyleOverride()
            this.libassRenderer.setDefaultFont(DEFAULT_FONT_NAME)
            return
        }

        // check if the track has only one style, if so, apply the customization to that style
        let found = false
        const eventTrack = this.eventTracks[this.currentTrackNumber]
        if (eventTrack) {
            found = true
            if (eventTrack.styles && Object.keys(eventTrack.styles).length > 1) {
                subtitleLog.info("Track has multiple styles, not applying customization")
                return
            }
        }
        const fileTrack = this.fileTracks[this.currentTrackNumber]
        if (fileTrack) {
            found = true
            // if it's a file track, it was converted from another format so has only one style
        }

        if (!found) return

        const opts = this.settings.subtitleCustomization

        const primaryColor = hexToASSColor(vc_getSubtitleStyle(opts, "primaryColor"), 0)
        const outlineColor = hexToASSColor(vc_getSubtitleStyle(opts, "outlineColor"), 0)
        const backColor = hexToASSColor(vc_getSubtitleStyle(opts, "backColor"), vc_getSubtitleStyle(opts, "backColorOpacity"))

        // devnote: jassub scales down to 30% of the og scale
        // /jassub/blob/main/src/JASSUB.cpp#L709
        const customStyle = {
            Name: "CustomDefault",
            FontName: DEFAULT_FONT_NAME, // opts.fontName || DEFAULT_FONT_NAME,
            FontSize: vc_getSubtitleStyle(opts, "fontSize"),
            PrimaryColour: primaryColor,
            SecondaryColour: primaryColor,
            OutlineColour: outlineColor,
            BackColour: backColor,
            ScaleX: ((100) / 100),
            ScaleY: ((100) / 100),
            Outline: vc_getSubtitleStyle(opts, "outline"),
            Shadow: vc_getSubtitleStyle(opts, "shadow"),
            MarginV: 120,
            BorderStyle: 1,
            Alignment: 2, // Bottom center
            MarginL: 20,
            MarginR: 20,
            Bold: 0, // customization.bold ? 1 : 0,
            Encoding: 1,
            Justify: 0,
            Blur: 0,
            Italic: 0,
            Underline: 0,
            StrikeOut: 0,
            Spacing: 0,
            Angle: 0,
            treat_fontname_as_pattern: 0,
        }

        // Apply the style override
        this.libassRenderer.styleOverride(customStyle)
        subtitleLog.info("Applied subtitle customization override", customStyle)

        // Apply font change
        // fontName can be something like "Noto Sans SC" or "Noto Sans SC.ttf"
        if (opts.fontName) {
            const _fontName = opts.fontName.trim()
            let url = getAssetUrl(`${_fontName}.woff2`)
            if (_fontName.includes(".")) {
                url = getAssetUrl(_fontName) // use the fontname as filename if there's an extension
            }

            // if the font is not already loaded, load it
            const fontName = _fontName.split(".")[0]
            if (this.fonts.includes(url)) {
                subtitleLog.info("Setting default font to", fontName)
                this.libassRenderer.setDefaultFont(fontName)
                return
            }

            subtitleLog.info("Applying font change", url, ", setting default font to", fontName)
            this.fonts.push(url)
            this.libassRenderer.addFont(url)
            this.libassRenderer!.setDefaultFont(fontName)
        } else {
            this.libassRenderer.setDefaultFont(DEFAULT_FONT_NAME)
        }
    }

    private __eventMapKey(event: MKVParser_SubtitleEvent): string {
        return JSON.stringify(event)
    }

    // Stores the styles for each track.
    private _storeEventTrackStyles() {
        if (!this.playbackInfo?.mkvMetadata?.subtitleTracks) return
        for (const track of this.playbackInfo.mkvMetadata.subtitleTracks) {
            const codecPrivate = track.codecPrivate?.slice?.(0, -1) || this.defaultSubtitleHeader
            const lines = codecPrivate.replaceAll("\r\n", "\n").split("\n").filter(line => line.startsWith("Style:"))
            let index = 1
            const s: Record<string, number> = {}
            this.eventTracks[track.number].styles = s // reset styles
            for (const line of lines) {
                let styleName = line.split("Style:")[1]
                styleName = (styleName.split(",")[0] || "").trim()
                if (styleName && !s[styleName]) {
                    s[styleName] = index++
                }
            }
            this.eventTracks[track.number].styles = s
        }
    }

    private _createAssEvent(event: MKVParser_SubtitleEvent, index: number): ASS_Event {
        return {
            Start: event.startTime,
            Duration: event.duration,
            Style: String(event.extraData?.style ? this.eventTracks[event.trackNumber]?.styles?.[event.extraData?.style ?? "Default"] : 1),
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

    private _addEventTrack(track: MKVParser_TrackInfo) {
        this.eventTracks[track.number] = {
            info: track,
            events: new Map(),
            styles: {},
        }

        // If this is a PGS track, initialize it in the PGS events map
        // PGS tracks will also have an entry in eventTracks
        if (isPGS(track.codecID)) {
            this.pgsEventTracks[track.number] = {
                info: track,
                events: new Map(),
            }
        }
    }

    private _recordSubtitleEvent(event: MKVParser_SubtitleEvent): { isNew: boolean, assEvent: ASS_Event | null } {
        const trackEventMap = this.eventTracks[event.trackNumber]?.events // get the map
        if (!trackEventMap) {
            return { isNew: false, assEvent: null }
        }

        const eventKey = this.__eventMapKey(event)

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

    //////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

    // +-----------------------+
    // |      File Tracks      |
    // +-----------------------+

    // // Adds a new track AFTER initialization and selects it
    addFileTrack(track: VideoCore_VideoSubtitleTrack) {
        subtitleLog.info("Subtitle file track added", track)
        toast.success(`Subtitle track added: ${track.label}`)
        const lastFileTrackNumber = Object.keys(this.fileTracks).length
            ? Number(Object.keys(this.fileTracks)[Object.keys(this.fileTracks).length - 1])
            : 999
        const number = lastFileTrackNumber + 1
        this.fileTracks[number] = {
            info: track,
            content: null,
        }
        // Select the track
        this.selectTrack(number)
        this._init()
        this.libassRenderer?.resize?.()
        this.pgsRenderer?.resize()

        const tracks = this._getTracks()
        const normalizedTrack = tracks.find(t => t.number === number)
        if (normalizedTrack) {
            const event: SubtitleManagerTrackAddedEvent = new CustomEvent("trackadded", { detail: { track: normalizedTrack } })
            this.dispatchEvent(event)
        }
        const event: SubtitleManagerTracksLoadedEvent = new CustomEvent("tracksloaded", { detail: { tracks: tracks } })
        this.dispatchEvent(event)
        this._onTracksLoaded?.(tracks)
    }

    // Fetches the track's content and converts it to ASS.
    // If the content is already fetched, it will load it.
    private async _handleFileTrack(trackNumber: number, fileTrack: { info: VideoCore_VideoSubtitleTrack, content: string | null }) {
        subtitleLog.info("Handling file track", trackNumber, fileTrack.info)

        if (!this.fetchAndConvertToASS) {
            subtitleLog.error("fetchAndConvertToASS callback not provided")
            return
        }

        // If content is already loaded, use it
        if (!!fileTrack.content) {
            subtitleLog.info("Using cached converted content for track", trackNumber)
            this.libassRenderer?.setTrack(fileTrack.content)
            this._applySubtitleCustomization()
            this.libassRenderer?.resize?.()
            this.pgsRenderer?.resize()
            return
        }

        // Convert the subtitle to ASS format
        if (fileTrack.info.type === "ass") {
            try {
                if (fileTrack.info.src) subtitleLog.info("Fetching subtitle content", fileTrack.info.src)
                const content = fileTrack.info.src ? await fetch(fileTrack.info.src).then(res => res.text()) : (fileTrack.info.content || "")
                this.libassRenderer?.setTrack(content)
                this._applySubtitleCustomization()
                this.libassRenderer?.resize?.()
                this.pgsRenderer?.resize()
            }
            catch (error) {
                subtitleLog.error("Error fetching subtitle content", error)
                toast.error("Failed to load subtitle track")
            }
        } else {
            try {
                subtitleLog.info("Converting subtitle to ASS format")
                const assContent = await this.fetchAndConvertToASS(fileTrack.info.src, fileTrack.info.content)

                if (!assContent) {
                    subtitleLog.error("Failed to convert subtitle to ASS format")
                    toast.error("Failed to convert subtitle track")
                    return
                }

                // Cache the converted content
                this.fileTracks[trackNumber].content = assContent

                // Load the converted content
                subtitleLog.info("Loading converted ASS content")
                this.libassRenderer?.setTrack(assContent)
                this._applySubtitleCustomization()
                this.libassRenderer?.resize?.()
                this.pgsRenderer?.resize()
            }
            catch (error) {
                subtitleLog.error("Error converting subtitle to ASS", error)
                toast.error("Failed to convert subtitle track")
            }
        }

        const selectedEvent: SubtitleManagerTrackSelectedEvent = new CustomEvent("trackselected", { detail: { trackNumber, kind: "file" } })
        this.dispatchEvent(selectedEvent)
    }
}

export function getDefaultSubtitleTrackNumber(
    settings: VideoCoreSettings,
    _tracks: { label?: string, language?: string, number: number, forced?: boolean, default?: boolean }[] | null = null,
): number {
    // Split preferred languages by comma and trim whitespace
    const preferredLanguages = settings.preferredSubtitleLanguage
        .split(",")
        .map(lang => lang.trim())
        .filter(lang => lang.length > 0)

    const blacklistLabels = (settings.preferredSubtitleBlacklist ?? "")
        .split(",")
        .map(label => label.trim().toLowerCase())
        .filter(label => label.length > 0)

    let tracks = _tracks ?? []
    // remove blacklisted tracks if there are more than one
    if (blacklistLabels.length && tracks.length > 1) {
        tracks = tracks?.filter?.(t => !t.label || !blacklistLabels.includes(t.label?.toLowerCase())) ?? []
    }

    // Try each preferred language in order
    for (const preferredLang of preferredLanguages) {
        let foundTracks = tracks?.filter?.(t => t.language?.toLowerCase() === preferredLang?.toLowerCase())
        if (foundTracks?.length) {
            // Find default or forced track
            const defaultIndex = foundTracks.findIndex(t => t.forced)
            return foundTracks[defaultIndex >= 0 ? defaultIndex : 0].number
        }
        if (preferredLang === "none") {
            return NO_TRACK_NUMBER
        }
    }

    // No preferred tracks found, look for default or forced tracks
    const defaultOrForcedTracks = tracks?.filter?.(t => t.default || t.forced)
    if (defaultOrForcedTracks?.length) {
        // Prioritize default tracks over forced tracks
        const defaultIndex = defaultOrForcedTracks.findIndex(t => t.default)
        return defaultOrForcedTracks[defaultIndex >= 0 ? defaultIndex : 0].number
    }

    // No forced/default tracks found, select the first track
    return tracks?.[0]?.number ?? NO_TRACK_NUMBER
}
