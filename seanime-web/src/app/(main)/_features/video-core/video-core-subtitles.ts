import { getServerBaseUrl } from "@/api/client/server-url"
import { MKVParser_SubtitleEvent, MKVParser_TrackInfo } from "@/api/generated/types"

import { getSubtitleTrackType } from "@/app/(main)/_features/video-core/video-core-subtitle-menu"
import { VideoCorePlaybackInfo, VideoCoreSettings, VideoCoreSubtitleTrack } from "@/app/(main)/_features/video-core/video-core.atoms"
import { logger } from "@/lib/helpers/debug"
import { getAssetUrl, legacy_getAssetUrl } from "@/lib/server/assets"
import JASSUB, { ASS_Event, ASS_Style, JassubOptions } from "jassub"
import { toast } from "sonner"

const subtitleLog = logger("SUBTITLE")

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

export type NormalizedTrackInfo = {
    language?: string
    languageIETF?: string
    codecID?: string
    label?: string
    number: number
    forced: boolean
    default: boolean
}

// Manages ASS subtitle streams.
export class VideoCoreSubtitleManager {
    private readonly videoElement: HTMLVideoElement
    private readonly jassubOffscreenRender: boolean
    libassRenderer: JASSUB | null = null
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
    private mkvTracks: Record<string, {
        info: MKVParser_TrackInfo
        events: Map<string, { event: MKVParser_SubtitleEvent, assEvent: ASS_Event }>
        styles: Record<string, number>
    }> = {}

    // URL-based tracks (will use internal API to convert to ASS)
    private nonMkvTracks: Record<string, {
        info: VideoCoreSubtitleTrack
        content: string | null // converted content
    }> = {}

    private readonly fetchAndConvertToASS?: (url: string) => Promise<string | undefined>

    private playbackInfo: VideoCorePlaybackInfo
    private currentTrackNumber: number = NO_TRACK_NUMBER
    private fonts: string[] = []

    private _onSelectedTrackChanged?: (track: number | null) => void

    constructor({
        videoElement,
        jassubOffscreenRender,
        playbackInfo,
        settings,
        fetchAndConvertToASS,
    }: {
        videoElement: HTMLVideoElement
        jassubOffscreenRender: boolean
        playbackInfo: VideoCorePlaybackInfo
        settings: VideoCoreSettings
        fetchAndConvertToASS?: (url: string) => Promise<string | undefined>
    }) {
        this.videoElement = videoElement
        this.jassubOffscreenRender = jassubOffscreenRender
        this.playbackInfo = playbackInfo
        this.settings = settings
        this.fetchAndConvertToASS = fetchAndConvertToASS

        /*
         * MKV Tracks
         */
        if (this.playbackInfo?.mkvMetadata?.subtitleTracks) {
            for (const track of this.playbackInfo.mkvMetadata.subtitleTracks) {
                this._mkvAddTrack(track)
            }
            this._storeMkvTrackStyles() // Store their styles
        }

        /*
         * Non-MKV Tracks
         */
        if (this.playbackInfo?.subtitleTracks) {
            let trackNumber = 1000 // Start from 1000
            for (const track of this.playbackInfo.subtitleTracks) {
                if (track.useLibassRenderer) {
                    this.nonMkvTracks[trackNumber] = {
                        info: track,
                        content: null,
                    }
                    trackNumber++
                }
            }
        }

        // Select default track if we have any tracks
        if (this.playbackInfo?.mkvMetadata?.subtitleTracks || Object.keys(this.nonMkvTracks).length > 0) {
            this._selectDefaultTrack()
        }

        subtitleLog.info("Text tracks", this.videoElement.textTracks)
        subtitleLog.info("Tracks", this.mkvTracks)
        subtitleLog.info("Non-MKV tracks", this.nonMkvTracks)
    }

    addTrackChangedEventListener(callback: (track: number | null) => void) {
        this._onSelectedTrackChanged = callback
    }

    // Sets the track to no track.
    setNoTrack() {
        this.currentTrackNumber = NO_TRACK_NUMBER
        this.libassRenderer?.setTrack(this.defaultSubtitleHeader)
        this.libassRenderer?.resize?.()
        this._onSelectedTrackChanged?.(NO_TRACK_NUMBER)
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
         * Non-MKV track
         */

        // Check if this is a non-MKV track (non-MKV)
        // If it is, fetch/convert the content and add it to the libass renderer
        const nonMkvTrack = this.nonMkvTracks[trackNumber]
        if (nonMkvTrack) {
            this._handleNonMkvTrack(trackNumber, nonMkvTrack)
            return
        }

        /*
         * MKV track
         */

        const mkvTrack = this.mkvTracks[trackNumber]
        if (!mkvTrack) {
            subtitleLog.warning("MKV Track not found", trackNumber)
            return
        }

        // Handle MKV track
        const codecPrivate = mkvTrack.info.codecPrivate?.slice?.(0, -1) || this.defaultSubtitleHeader

        // Set the track
        this.libassRenderer?.setTrack(codecPrivate)

        // Apply customization to Default styles
        this._applySubtitleCustomization()

        const trackEventMap = this.mkvTracks[track.number]?.events
        if (!trackEventMap) {
            return
        }
        subtitleLog.info("Found", trackEventMap.size, "events for track", track.number)

        // Add the events to the libass renderer
        for (const { assEvent } of trackEventMap.values()) {
            this.libassRenderer?.createEvent(assEvent)
        }

        this.libassRenderer?.resize?.()
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

    destroy() {
        subtitleLog.info("Destroying subtitle manager")
        this.libassRenderer?.destroy()
        this.libassRenderer = null
        for (const trackNumber in this.mkvTracks) {
            this.mkvTracks[trackNumber].events.clear()
        }
        this.mkvTracks = {}
        this.nonMkvTracks = {}
        this.currentTrackNumber = NO_TRACK_NUMBER
    }

    getSelectedTrackNumberOrNull(): number | null {
        if (this.currentTrackNumber === NO_TRACK_NUMBER) return null
        return this.currentTrackNumber
    }

    // Update settings and reapply subtitle customization to current track
    updateSettings(newSettings: VideoCoreSettings) {
        this.settings = newSettings
        // Reapply customization if a track is currently selected
        if (this.currentTrackNumber !== NO_TRACK_NUMBER) {
            this._applySubtitleCustomization()
        }
    }

    onMkvTrackAdded(track: MKVParser_TrackInfo) {
        subtitleLog.info("Subtitle track added", track)
        toast.success(`Subtitle track added: ${track.name}`)
        this._mkvAddTrack(track)
        this._storeMkvTrackStyles()
        // Add the track to the video element
        const trackEl = document.createElement("track")
        trackEl.id = track.number.toString()
        trackEl.kind = "subtitles"
        trackEl.label = track.name || ""
        trackEl.srclang = track.language || "eng"
        this.videoElement.appendChild(trackEl)
        // this._selectDefaultTrack()
        this.selectTrack(track.number)
        this._init()
        this.libassRenderer?.resize?.()
    }

    private _getTracks(): NormalizedTrackInfo[] {
        const mkvTracks = Object.values(this.mkvTracks).map(t => <NormalizedTrackInfo>({
            language: t.info.language,
            number: t.info.number,
            label: t.info.name,
            forced: t.info.forced,
            default: t.info.default,
            languageIETF: t.info.languageIETF,
            codecID: t.info.codecID,
        }))

        const nonMkvTracks = Object.entries(this.nonMkvTracks).map(([trackNumber, t]) => <NormalizedTrackInfo>({
            language: t.info.language,
            number: Number(trackNumber),
            label: t.info.label,
            forced: false,
            default: t.info.default,
        }))

        return [...mkvTracks, ...nonMkvTracks].sort((a, b) => a.number - b.number)
    }

    // Selects a track to be used.
    // This should be called after the tracks are loaded.

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
            subContent: this.defaultSubtitleHeader, // needed
            // subUrl: new URL("/jassub/test.ass", window.location.origin).toString(),
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
        const mkvTrack = this.mkvTracks[this.currentTrackNumber]
        if (mkvTrack) {
            found = true
            if (mkvTrack.styles && Object.keys(mkvTrack.styles).length > 1) {
                subtitleLog.info("Track has multiple styles, not applying customization")
                return
            }
        }
        const nonMkvTrack = this.nonMkvTracks[this.currentTrackNumber]
        if (nonMkvTrack) {
            found = true
            // if it's a non-MKV track, it was converted from another format so has only one style
        }

        if (!found) return

        const opts = this.settings.subtitleCustomization

        const primaryColor = hexToASSColor(opts.primaryColor || "#FFFFFF", 0)
        const outlineColor = hexToASSColor(opts.outlineColor || "#000000", 0)
        const backColor = hexToASSColor(opts.backColor || "#000000", 0)

        // devnote: jassub scales down to 30% of the og scale
        // /jassub/blob/main/src/JASSUB.cpp#L709
        const customStyle = {
            Name: "CustomDefault",
            FontName: DEFAULT_FONT_NAME, // opts.fontName || DEFAULT_FONT_NAME,
            FontSize: opts.fontSize || 62,
            PrimaryColour: primaryColor,
            SecondaryColour: primaryColor,
            OutlineColour: outlineColor,
            BackColour: backColor,
            ScaleX: ((opts.scaleX || 100) / 100),
            ScaleY: ((opts.scaleY || 100) / 100),
            Outline: opts.outline ?? 3,
            Shadow: opts.shadow ?? 0,
            MarginV: opts.marginV ?? 120,
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

    // +-----------------------+
    // |       MKV Tracks      |
    // +-----------------------+

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

    // When called for the first time, it will initialize the libass renderer.
    private _selectDefaultTrack() {
        if (this.currentTrackNumber !== NO_TRACK_NUMBER) return
        const tracks = this._getTracks()

        if (!tracks?.length) {
            this.setNoTrack()
            return
        }

        if (tracks.length === 1) {
            this.selectTrack(tracks[0].number)
            return
        }

        // Split preferred languages by comma and trim whitespace
        const defaultTrackNumber = getDefaultSubtitleTrackNumber(this.settings, tracks)
        this.selectTrack(defaultTrackNumber)
    }

    isTrackSupported(trackNumber: number): boolean {
        if (trackNumber === NO_TRACK_NUMBER) return true

        const track = this.playbackInfo?.mkvMetadata?.subtitleTracks?.find(t => t.number === trackNumber)
        if (!track) return true

        return getSubtitleTrackType(track.codecID) !== "PGS"
    }

    private _mkvAddTrack(track: MKVParser_TrackInfo) {
        this.mkvTracks[track.number] = {
            info: track,
            events: new Map(),
            styles: {},
        }
        return this.mkvTracks[track.number]
    }

    // Stores the styles for each track.
    private _storeMkvTrackStyles() {
        if (!this.playbackInfo?.mkvMetadata?.subtitleTracks) return
        for (const track of this.playbackInfo.mkvMetadata.subtitleTracks) {
            const codecPrivate = track.codecPrivate?.slice?.(0, -1) || this.defaultSubtitleHeader
            const lines = codecPrivate.replaceAll("\r\n", "\n").split("\n").filter(line => line.startsWith("Style:"))
            let index = 1
            const s: Record<string, number> = {}
            this.mkvTracks[track.number].styles = s // reset styles
            for (const line of lines) {
                let styleName = line.split("Style:")[1]
                styleName = (styleName.split(",")[0] || "").trim()
                if (styleName && !s[styleName]) {
                    s[styleName] = index++
                }
            }
            this.mkvTracks[track.number].styles = s
        }
    }

    private __eventMapKey(event: MKVParser_SubtitleEvent): string {
        return JSON.stringify(event)
    }

    private _createAssEvent(event: MKVParser_SubtitleEvent, index: number): ASS_Event {
        return {
            Start: event.startTime,
            Duration: event.duration,
            Style: String(event.extraData?.style ? this.mkvTracks[event.trackNumber]?.styles?.[event.extraData?.style ?? "Default"] : 1),
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
        const trackEventMap = this.mkvTracks[event.trackNumber]?.events // get the map
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

    // +-----------------------+
    // |     Non-MKV Tracks    |
    // +-----------------------+

    // Fetches the track's content and converts it to ASS.
    // If the content is already fetched, it will load it.
    private async _handleNonMkvTrack(trackNumber: number, nonMkvTrack: { info: VideoCoreSubtitleTrack, content: string | null }) {
        subtitleLog.info("Handling non-MKV track", trackNumber, nonMkvTrack.info)

        if (!this.fetchAndConvertToASS) {
            subtitleLog.error("fetchAndConvertToASS callback not provided")
            return
        }

        // If content is already loaded, use it
        if (!!nonMkvTrack.content) {
            subtitleLog.info("Using cached converted content for track", trackNumber)
            this.libassRenderer?.setTrack(nonMkvTrack.content)
            this._applySubtitleCustomization()
            this.libassRenderer?.resize?.()
            return
        }

        // Convert the subtitle to ASS format
        if (nonMkvTrack.info.type === "ass") {
            try {
                subtitleLog.info("Fetching subtitle content", nonMkvTrack.info.src)
                const content = await fetch(nonMkvTrack.info.src).then(res => res.text())
                this.libassRenderer?.setTrack(content)
                this._applySubtitleCustomization()
                this.libassRenderer?.resize?.()
            }
            catch (error) {
                subtitleLog.error("Error fetching subtitle content", error)
                toast.error("Failed to load subtitle track")
            }
        } else {
            try {
                subtitleLog.info("Converting subtitle to ASS format", nonMkvTrack.info.src)
                const assContent = await this.fetchAndConvertToASS(nonMkvTrack.info.src)

                if (!assContent) {
                    subtitleLog.error("Failed to convert subtitle to ASS format")
                    toast.error("Failed to convert subtitle track")
                    return
                }

                // Cache the converted content
                this.nonMkvTracks[trackNumber].content = assContent

                // Load the converted content
                subtitleLog.info("Loading converted ASS content")
                this.libassRenderer?.setTrack(assContent)
                this._applySubtitleCustomization()
                this.libassRenderer?.resize?.()
            }
            catch (error) {
                subtitleLog.error("Error converting subtitle to ASS", error)
                toast.error("Failed to convert subtitle track")
            }
        }
    }
}

export function getDefaultSubtitleTrackNumber(
    settings: VideoCoreSettings,
    tracks: { label?: string, language?: string, number: number, forced?: boolean, default?: boolean }[] | null = null,
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

    // Try each preferred language in order
    for (const preferredLang of preferredLanguages) {
        const foundTracks = tracks?.filter?.(t => t.language === preferredLang && !blacklistLabels.includes(t.label?.toLowerCase() || ""))
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
