import { getServerBaseUrl } from "@/api/client/server-url"
import type { MpvCore_ServerEvent, Player_PlaybackInfo, Player_SkipData } from "@/api/generated/types"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { WSEvents } from "@/lib/server/ws-events"
import { __isElectronDesktop__ } from "@/types/constants"
import type { MpvPrismTrack, MpvPrismTrackKind } from "@mpv-prism/core"
import { useMpvPrismPlayer } from "@mpv-prism/react"
import { useAtom } from "jotai"
import React from "react"
import { MpvCorePlayerInner } from "./mpv-core-player-inner"
import { mpvCore_stateAtom, type MpvCoreAnime4KQuality, type MpvCoreSettings } from "./mpv-core.atoms"

export type MpvCoreEnvelope = { type: MpvCore_ServerEvent, payload: unknown }

export type MpvCoreNativeChapter = {
    time?: number
    title?: string
}

export type MpvCoreChapterCue = {
    startTime: number
    endTime: number
    text: string
}

type DocumentPictureInPictureApi = {
    requestWindow(options?: { width?: number; height?: number }): Promise<Window>
}

const subtitleExtensions = ["srt", "ass", "ssa", "vtt", "ttml", "stl", "txt"]

export function normalizeMpvChapterList(value: unknown): MpvCoreNativeChapter[] {
    if (!Array.isArray(value)) return []

    return value.flatMap(item => {
        if (!item || typeof item !== "object") return []
        const chapter = item as Record<string, unknown>
        const time = Number(chapter.time)
        if (!Number.isFinite(time) || time < 0) return []
        return [{
            time,
            title: typeof chapter.title === "string" ? chapter.title : "",
        }]
    }).sort((a, b) => (a.time ?? 0) - (b.time ?? 0))
}

export function createMpvChapterCues(chapters: MpvCoreNativeChapter[], duration: number): MpvCoreChapterCue[] {
    if (!chapters.length || duration <= 0) return []

    return fillMpvChapterTimeline(chapters.flatMap((chapter, index) => {
        const startTime = Math.max(0, chapter.time ?? 0)
        const endTime = Math.min(duration, Math.max(startTime, chapters[index + 1]?.time ?? duration))
        if (startTime > duration || endTime <= startTime) return []
        return [{
            startTime,
            endTime,
            text: chapter.title ?? "",
        }]
    }), duration)
}

export function createSkipChapterCues(skipData: Player_SkipData | null, duration: number): MpvCoreChapterCue[] {
    if (!skipData || duration <= 0) return []

    const chapters = [
        skipData.op?.interval && {
            startTime: skipData.op.interval.startTime,
            endTime: skipData.op.interval.endTime,
            text: "Opening",
        },
        skipData.ed?.interval && {
            startTime: skipData.ed.interval.startTime,
            endTime: skipData.ed.interval.endTime,
            text: "Ending",
        },
    ].filter((chapter): chapter is MpvCoreChapterCue => (
        !!chapter &&
        chapter.startTime >= 0 &&
        chapter.endTime > chapter.startTime &&
        chapter.startTime <= duration
    )).sort((a, b) => a.startTime - b.startTime)

    return fillMpvChapterTimeline(chapters, duration)
}

function fillMpvChapterTimeline(chapters: MpvCoreChapterCue[], duration: number): MpvCoreChapterCue[] {
    const timeline: MpvCoreChapterCue[] = []
    let cursor = 0

    for (const chapter of chapters.toSorted((a, b) => a.startTime - b.startTime)) {
        const startTime = Math.max(cursor, Math.min(duration, chapter.startTime))
        const endTime = Math.max(startTime, Math.min(duration, chapter.endTime))
        if (startTime > cursor) {
            timeline.push({ startTime: cursor, endTime: startTime, text: "" })
        }
        if (endTime > startTime) {
            timeline.push({ ...chapter, startTime, endTime })
            cursor = endTime
        }
    }

    if (cursor < duration) {
        timeline.push({ startTime: cursor, endTime: duration, text: "" })
    }

    return timeline
}

const anime4KProfiles: Record<MpvCoreAnime4KQuality, Record<string, string[]>> = {
    fast: {
        "mode-a": [
            "Anime4K_Clamp_Highlights.glsl",
            "Anime4K_Restore_CNN_M.glsl",
            "Anime4K_Upscale_CNN_x2_M.glsl",
            "Anime4K_AutoDownscalePre_x2.glsl",
            "Anime4K_AutoDownscalePre_x4.glsl",
            "Anime4K_Upscale_CNN_x2_S.glsl",
        ],
        "mode-b": [
            "Anime4K_Clamp_Highlights.glsl",
            "Anime4K_Restore_CNN_Soft_M.glsl",
            "Anime4K_Upscale_CNN_x2_M.glsl",
            "Anime4K_AutoDownscalePre_x2.glsl",
            "Anime4K_AutoDownscalePre_x4.glsl",
            "Anime4K_Upscale_CNN_x2_S.glsl",
        ],
        "mode-c": [
            "Anime4K_Clamp_Highlights.glsl",
            "Anime4K_Upscale_Denoise_CNN_x2_M.glsl",
            "Anime4K_AutoDownscalePre_x2.glsl",
            "Anime4K_AutoDownscalePre_x4.glsl",
            "Anime4K_Upscale_CNN_x2_S.glsl",
        ],
        "mode-aa": [
            "Anime4K_Clamp_Highlights.glsl",
            "Anime4K_Restore_CNN_M.glsl",
            "Anime4K_Upscale_CNN_x2_M.glsl",
            "Anime4K_Restore_CNN_S.glsl",
            "Anime4K_AutoDownscalePre_x2.glsl",
            "Anime4K_AutoDownscalePre_x4.glsl",
            "Anime4K_Upscale_CNN_x2_S.glsl",
        ],
        "mode-bb": [
            "Anime4K_Clamp_Highlights.glsl",
            "Anime4K_Restore_CNN_Soft_M.glsl",
            "Anime4K_Upscale_CNN_x2_M.glsl",
            "Anime4K_Restore_CNN_Soft_S.glsl",
            "Anime4K_AutoDownscalePre_x2.glsl",
            "Anime4K_AutoDownscalePre_x4.glsl",
            "Anime4K_Upscale_CNN_x2_S.glsl",
        ],
        "mode-ca": [
            "Anime4K_Clamp_Highlights.glsl",
            "Anime4K_Upscale_Denoise_CNN_x2_M.glsl",
            "Anime4K_Restore_CNN_M.glsl",
            "Anime4K_AutoDownscalePre_x2.glsl",
            "Anime4K_AutoDownscalePre_x4.glsl",
            "Anime4K_Upscale_CNN_x2_S.glsl",
        ],
        "cnn-2x-medium": [
            "Anime4K_Upscale_CNN_x2_M.glsl",
        ],
        "cnn-2x-very-large": [
            "Anime4K_Upscale_CNN_x2_VL.glsl",
        ],
        "denoise-cnn-2x-very-large": [
            "Anime4K_Upscale_Denoise_CNN_x2_VL.glsl",
        ],
        "cnn-2x-ultra-large": [
            "Anime4K_Upscale_CNN_x2_UL.glsl",
        ],
    },
    hq: {
        "mode-a": [
            "Anime4K_Clamp_Highlights.glsl",
            "Anime4K_Restore_CNN_VL.glsl",
            "Anime4K_Upscale_CNN_x2_VL.glsl",
            "Anime4K_AutoDownscalePre_x2.glsl",
            "Anime4K_AutoDownscalePre_x4.glsl",
            "Anime4K_Upscale_CNN_x2_M.glsl",
        ],
        "mode-b": [
            "Anime4K_Clamp_Highlights.glsl",
            "Anime4K_Restore_CNN_Soft_VL.glsl",
            "Anime4K_Upscale_CNN_x2_VL.glsl",
            "Anime4K_AutoDownscalePre_x2.glsl",
            "Anime4K_AutoDownscalePre_x4.glsl",
            "Anime4K_Upscale_CNN_x2_M.glsl",
        ],
        "mode-c": [
            "Anime4K_Clamp_Highlights.glsl",
            "Anime4K_Upscale_Denoise_CNN_x2_VL.glsl",
            "Anime4K_AutoDownscalePre_x2.glsl",
            "Anime4K_AutoDownscalePre_x4.glsl",
            "Anime4K_Upscale_CNN_x2_M.glsl",
        ],
        "mode-aa": [
            "Anime4K_Clamp_Highlights.glsl",
            "Anime4K_Restore_CNN_VL.glsl",
            "Anime4K_Upscale_CNN_x2_VL.glsl",
            "Anime4K_Restore_CNN_M.glsl",
            "Anime4K_AutoDownscalePre_x2.glsl",
            "Anime4K_AutoDownscalePre_x4.glsl",
            "Anime4K_Upscale_CNN_x2_M.glsl",
        ],
        "mode-bb": [
            "Anime4K_Clamp_Highlights.glsl",
            "Anime4K_Restore_CNN_Soft_VL.glsl",
            "Anime4K_Upscale_CNN_x2_VL.glsl",
            "Anime4K_Restore_CNN_Soft_M.glsl",
            "Anime4K_AutoDownscalePre_x2.glsl",
            "Anime4K_AutoDownscalePre_x4.glsl",
            "Anime4K_Upscale_CNN_x2_M.glsl",
        ],
        "mode-ca": [
            "Anime4K_Clamp_Highlights.glsl",
            "Anime4K_Upscale_Denoise_CNN_x2_VL.glsl",
            "Anime4K_Restore_CNN_VL.glsl",
            "Anime4K_AutoDownscalePre_x2.glsl",
            "Anime4K_AutoDownscalePre_x4.glsl",
            "Anime4K_Upscale_CNN_x2_M.glsl",
        ],
        "cnn-2x-medium": [
            "Anime4K_Upscale_CNN_x2_M.glsl",
        ],
        "cnn-2x-very-large": [
            "Anime4K_Upscale_CNN_x2_VL.glsl",
        ],
        "denoise-cnn-2x-very-large": [
            "Anime4K_Upscale_Denoise_CNN_x2_VL.glsl",
        ],
        "cnn-2x-ultra-large": [
            "Anime4K_Upscale_CNN_x2_UL.glsl",
        ],
    },
}

const diagnosticsProperties = new Set([
    "video-params",
    "audio-params",
    "estimated-vf-fps",
    "display-fps",
    "video-bitrate",
    "audio-bitrate",
    "file-format",
    "hwdec-current",
])

export function isEditableKeyboardTarget(target: EventTarget | null) {
    if (!(target instanceof Element)) return false
    if (target instanceof HTMLElement && target.isContentEditable) return true
    return !!target.closest("input, textarea, select, [contenteditable='true'], [role='textbox']")
}

function mpvColor(color: string, opacity = 1) {
    const normalized = /^#[0-9a-f]{6}$/i.test(color) ? color.slice(1).toUpperCase() : "FFFFFF"
    const alpha = Math.round(Math.max(0, Math.min(1, opacity)) * 255).toString(16).padStart(2, "0").toUpperCase()
    return `#${alpha}${normalized}`
}

export async function applyMpvSubtitleSettings(
    player: NonNullable<ReturnType<typeof useMpvPrismPlayer>>,
    settings: MpvCoreSettings,
) {
    const style = settings.subtitleCustomization
    const fontName = style.fontName.trim().replace(/\.(woff2?|ttf|otf)$/i, "") || "sans-serif"
    const properties: Array<[string, string | number]> = style.enabled
        ? [
            ["sub-ass-override", "force"],
            ["sub-font", fontName],
            ["sub-font-size", style.fontSize],
            ["sub-color", mpvColor(style.primaryColor)],
            ["sub-outline-color", mpvColor(style.outlineColor)],
            ["sub-back-color", mpvColor(style.backColor, style.backColorOpacity)],
            ["sub-outline-size", style.outline],
            ["sub-shadow-offset", style.shadow],
            ["sub-border-style", "outline-and-shadow"],
        ]
        : [
            ["sub-ass-override", "no"],
            ["sub-font", "sans-serif"],
            ["sub-font-size", 38],
            ["sub-color", "#FFFFFFFF"],
            ["sub-outline-color", "#FF000000"],
            ["sub-back-color", "#00000000"],
            ["sub-outline-size", 1.65],
            ["sub-shadow-offset", 0],
        ]

    await Promise.all([
        player.setProperty("sub-delay", settings.subtitleDelay),
        ...properties.map(([name, value]) => player.setProperty(name, value)),
    ])
}

export function mc_selectPreferredTrack(
    tracks: MpvPrismTrack[],
    kind: "audio" | "subtitle",
    preferredValue: string,
    blacklistValue = "",
) {
    const preferred = preferredValue.split(",").map(value => value.trim().toLowerCase()).filter(Boolean)
    const blacklist = blacklistValue.split(",").map(value => value.trim().toLowerCase()).filter(Boolean)
    const candidates = tracks.filter(track => mc_trackKind(track) === kind)

    return candidates.find(track => {
        const language = String(track.lang ?? "").toLowerCase()
        const title = String(track.title ?? "").toLowerCase()
        if (blacklist.some(value => title.includes(value))) return false
        return preferred.some(value => language === value || language.includes(value) || title.includes(value))
    })
}

export function mc_resolveAnime4KProfile(
    directory: MpvCoreAnime4KDirectory | null,
    anime4kMode: string,
    quality: MpvCoreAnime4KQuality,
) {
    const byBasename = new Map(
        (directory?.shaders ?? []).map(shader => [
            shader.name.split("/").pop()?.toLowerCase() ?? shader.name.toLowerCase(),
            shader.path,
        ]),
    )
    const expected = anime4KProfiles[quality][anime4kMode] || []
    const missing = expected.filter(name => !byBasename.has(name.toLowerCase()))
    const paths = expected.map(name => byBasename.get(name.toLowerCase())).filter((value): value is string => !!value)
    return { paths, missing }
}

export function mc_resolveSource(value: string | undefined) {
    return (value ?? "").replace("{{SERVER_URL}}", getServerBaseUrl())
}

export function mc_trackKind(track: MpvPrismTrack): MpvPrismTrackKind | null {
    if (track.type === "audio") return "audio"
    if (track.type === "sub" || track.type === "subtitle") return "subtitle"
    if (track.type === "video") return "video"
    return null
}

export function mc_trackLabel(track: MpvPrismTrack) {
    const detail = [track.title, track.lang, track.codec].filter(Boolean).join(" · ")
    return detail || `${mc_trackKind(track) ?? "track"} ${track.id ?? ""}`
}

export function getMpvSubtitleCodecType(codec: string) {
    const clean = codec.toLowerCase()
    if (clean.includes("ass")) return "ASS"
    if (clean.includes("ssa")) return "SSA"
    if (clean.includes("pgs") || clean.includes("hdmv")) return "PGS"
    if (clean.includes("srt") || clean.includes("subrip")) return "SRT"
    if (clean.includes("vtt") || clean.includes("webvtt")) return "VTT"
    return codec.toUpperCase()
}

export function getMpvAudioCodecType(codec: string) {
    const clean = codec.toLowerCase()
    if (clean.includes("aac")) return "AAC"
    if (clean.includes("ac3") || clean.includes("ac-3")) return "AC3"
    if (clean.includes("dts")) return "DTS"
    if (clean.includes("flac")) return "FLAC"
    if (clean.includes("opus")) return "OPUS"
    if (clean.includes("truehd")) return "TRUEHD"
    return codec.toUpperCase()
}

export function mc_formatSubtitleTrack(track: MpvPrismTrack) {
    const codecStr = track.codec ? String(track.codec) : undefined
    const formattedCodec = codecStr ? getMpvSubtitleCodecType(codecStr) : undefined
    const label = track.title || track.lang?.toUpperCase() || `Track ${track.id}`
    const isLangSameAsLabel = track.lang?.toLowerCase() === track.title?.toLowerCase() || track.lang?.toUpperCase() === label
    const moreInfo = track.lang && !isLangSameAsLabel
        ? `${track.lang.toUpperCase()}${formattedCodec ? "/" + formattedCodec : ""}`
        : formattedCodec

    return {
        label,
        value: track.id,
        moreInfo,
    }
}

export function mc_formatAudioTrack(track: MpvPrismTrack) {
    const codecStr = track.codec ? String(track.codec) : undefined
    const formattedCodec = codecStr ? getMpvAudioCodecType(codecStr) : undefined
    const label = track.title || track.lang?.toUpperCase() || `Track ${track.id}`
    const isLangSameAsLabel = track.lang?.toLowerCase() === track.title?.toLowerCase() || track.lang?.toUpperCase() === label
    const moreInfo = track.lang && !isLangSameAsLabel
        ? `${track.lang.toUpperCase()}${formattedCodec ? "/" + formattedCodec : ""}`
        : formattedCodec

    return {
        label,
        value: track.id,
        moreInfo,
    }
}


export function mc_cacheBufferedSeconds(value: unknown, duration: number, currentTime: number) {
    if (!value || typeof value !== "object") return currentTime
    const state = value as Record<string, unknown>
    const ranges = state["seekable-ranges"]
    if (Array.isArray(ranges) && ranges.length > 0) {
        const last = ranges[ranges.length - 1]
        if (last && typeof last === "object") {
            const end = Number((last as Record<string, unknown>).end)
            if (Number.isFinite(end)) return Math.min(duration, end)
        }
    }
    const cacheDuration = Number(state["cache-duration"])
    return Number.isFinite(cacheDuration) ? Math.min(duration, currentTime + cacheDuration) : currentTime
}

export const BLOCKED_MPV_OPTIONS = new Set([
    "vo", "wid", "osc", "osd",
    "input-ipc-server", "input-unix-socket",
    "idle", "keep-open", "log-file"
])

export function mc_parseCustomMpvConfig(config: string): { parsed: Record<string, string>, ignored: string[] } {
    const parsed: Record<string, string> = {}
    const ignored: string[] = []

    if (!config) return { parsed, ignored }

    const lines = config.split(/\r?\n/)
    let inSection = false
    for (const line of lines) {
        let cleanLine = line.trim()
        if (!cleanLine || cleanLine.startsWith("#") || cleanLine.startsWith("//")) {
            continue
        }

        // Handle inline '#' comment (preceded by whitespace, not part of a hex color)
        let hashIndex = cleanLine.indexOf("#")
        while (hashIndex !== -1) {
            if (hashIndex > 0 && /\s/.test(cleanLine[hashIndex - 1])) {
                const remaining = cleanLine.slice(hashIndex)
                const isHexColor = /^#[0-9a-fA-F]{3,8}(?:\b|['"]|$)/.test(remaining)
                if (!isHexColor) {
                    cleanLine = cleanLine.slice(0, hashIndex).trim()
                    break
                }
            }
            hashIndex = cleanLine.indexOf("#", hashIndex + 1)
        }

        // Handle inline '//' comments (not part of a URL)
        const doubleSlashIndex = cleanLine.indexOf("//")
        if (doubleSlashIndex !== -1) {
            const prefix = cleanLine.slice(0, doubleSlashIndex)
            if (!prefix.match(/https?:$/i)) {
                cleanLine = prefix.trim()
            }
        }

        if (!cleanLine) {
            continue
        }

        // Only parse global options before the first profile block.
        if (cleanLine.startsWith("[") && cleanLine.endsWith("]")) {
            inSection = true
            continue
        }
        if (inSection) continue

        const eqIndex = cleanLine.indexOf("=")
        let rawKey = ""
        let rawValue = ""

        if (eqIndex === -1) {
            rawKey = cleanLine
            rawValue = "yes"
        } else {
            rawKey = cleanLine.slice(0, eqIndex)
            rawValue = cleanLine.slice(eqIndex + 1)
        }

        const key = rawKey.trim().replace(/^--/, "").toLowerCase().replace(/_/g, "-")
        let value = rawValue.trim()

        if ((value.startsWith("'") && value.endsWith("'")) || (value.startsWith("\"") && value.endsWith("\""))) {
            value = value.slice(1, -1).trim()
        }

        if (!key) continue

        if (!/^[a-zA-Z0-9_\-\.\/\:@\+]+$/.test(key)) {
            ignored.push(key)
            continue
        }

        if (BLOCKED_MPV_OPTIONS.has(key)) {
            ignored.push(key)
            continue
        }

        parsed[key] = value
    }

    return { parsed, ignored }
}

export function MpvCore() {
    const [state, setState] = useAtom(mpvCore_stateAtom)
    const serverStatus = useServerStatus()

    React.useEffect(() => {
        if (!__isElectronDesktop__ || !window.electron?.mpvCore) return
        const enabled = serverStatus?.settings?.mediaPlayer?.mpvPrismLogging ?? false
        window.electron.mpvCore.setLoggingEnabled(enabled)
    }, [serverStatus?.settings?.mediaPlayer?.mpvPrismLogging])

    useWebsocketMessageListener<MpvCoreEnvelope>({
        type: WSEvents.MPVCORE,
        onMessage: ({ type, payload }) => {
            switch (type) {
                case "open-and-await":
                    setState(draft => {
                        draft.active = true
                        draft.miniPlayer = false
                        draft.loadingState = String(payload || "Preparing stream...")
                        draft.playbackInfo = null
                        draft.playbackError = null
                    })
                    break
                case "watch":
                    setState(draft => {
                        draft.active = true
                        draft.miniPlayer = false
                        draft.loadingState = "Loading..."
                        draft.playbackInfo = payload as Player_PlaybackInfo
                        draft.playbackError = null
                    })
                    break
            }
        },
        deps: [],
    })

    if (!state.active) return null

    return <MpvCorePlayerInner />
}
