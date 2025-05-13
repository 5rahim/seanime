import { MKVParser_SubtitleEvent } from "@/api/generated/types"
import { legacy_getAssetUrl } from "@/lib/server/assets"
import { LibASSTextRenderer } from "@vidstack/react"

class SubtitleManager {
    // LibASS renderer
    private libassRenderer: LibASSTextRenderer | null = null
    // Subtitles for each track
    private trackSubtitles: Record<number, MKVParser_SubtitleEvent[]> = {}
    // Default fonts
    private defaultFonts: string[] = ["/jassub/default.woff2"]
    // Current subtitle track number
    private currentTrack: number | null = null

    constructor({
        jassubOffscreenRender,
    }: {
        jassubOffscreenRender: boolean
    }) {
        // User can add legacy WASM file to assets folder if needed
        const legacyWasmUrl = process.env.NODE_ENV === "development"
            ? "/jassub/jassub-worker.wasm.js" : legacy_getAssetUrl("/jassub/jassub-worker.wasm.js")

        // @ts-expect-error
        const renderer = new LibASSTextRenderer(() => import("jassub"), {
            wasmUrl: "/jassub/jassub-worker.wasm",
            workerUrl: "/jassub/jassub-worker.js",
            legacyWasmUrl: legacyWasmUrl,
            // Both parameters needed for subs to work on iOS, ref: jellyfin-vue
            offscreenRender: jassubOffscreenRender, // should be false for iOS
            prescaleFactor: 0.8,
            onDemandRender: false,
            fonts: this.defaultFonts,
            fallbackFont: this.defaultFonts[0],
        })
        this.trackSubtitles = {}
    }

    private _addSubtitleEvent(event: MKVParser_SubtitleEvent) {
        if (!this.trackSubtitles[event.trackNumber]) {
            this.trackSubtitles[event.trackNumber] = []
        }
        this.trackSubtitles[event.trackNumber].push(event)
    }


}

export function useHandleNativePlayer() {


}
