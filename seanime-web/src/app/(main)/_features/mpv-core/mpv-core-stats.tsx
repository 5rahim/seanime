import type { Player_PlaybackInfo } from "@/api/generated/types"
import type { MpvPrismTrack } from "@mpv-prism/core"
import React from "react"
import { mc_trackKind } from "./mpv-core"
import type { MpvCoreAnime4KQuality, MpvCoreShaderMode } from "./mpv-core.atoms"

export interface MpvCoreStatsProps {
    info: Player_PlaybackInfo | null
    tracks: MpvPrismTrack[]
    cache: unknown
    frameDrops: Record<string, number>
    diagnostics: Record<string, unknown>
    currentTime: number
    duration: number
    buffered: number
    speed: number
    buffering: boolean
    containerElement: HTMLElement | null
    shaderMode: MpvCoreShaderMode
    anime4kMode: string
    anime4kQuality: MpvCoreAnime4KQuality
    customShadersCount: number
}

export function MpvCoreStats(props: MpvCoreStatsProps) {
    const video = props.tracks.find(track => mc_trackKind(track) === "video" && track.selected)
    const audio = props.tracks.find(track => mc_trackKind(track) === "audio" && track.selected)
    const [displaySize, setDisplaySize] = React.useState({ width: 0, height: 0 })

    React.useEffect(() => {
        if (!props.containerElement) return
        const update = () => {
            const rect = props.containerElement?.getBoundingClientRect()
            if (!rect) return
            setDisplaySize({ width: Math.round(rect.width), height: Math.round(rect.height) })
        }
        update()
        const observer = new ResizeObserver(update)
        observer.observe(props.containerElement)
        return () => observer.disconnect()
    }, [props.containerElement])

    const videoParams = (props.diagnostics["video-params"] ?? {}) as Record<string, unknown>
    const videoWidth = Number(videoParams.dw ?? videoParams.w ?? video?.["demux-w"] ?? 0)
    const videoHeight = Number(videoParams.dh ?? videoParams.h ?? video?.["demux-h"] ?? 0)
    const fps = Number(props.diagnostics["estimated-vf-fps"] ?? video?.["demux-fps"] ?? 0)
    const videoBitrate = Number(props.diagnostics["video-bitrate"] ?? video?.["demux-bitrate"] ?? 0)
    const audioBitrate = Number(props.diagnostics["audio-bitrate"] ?? audio?.["demux-bitrate"] ?? 0)
    const outputDrops = props.frameDrops["frame-drop-count"] ?? 0
    const decoderDrops = props.frameDrops["decoder-frame-drop-count"] ?? 0
    const cache = props.cache as Record<string, unknown> | null
    const rawCacheDuration = Number(cache?.["cache-duration"])
    const cacheDuration = Number.isFinite(rawCacheDuration)
        ? rawCacheDuration
        : Math.max(0, props.buffered - props.currentTime)

    const StatLine = ({ label, value }: { label: string; value: React.ReactNode }) => (
        <div className="flex gap-2">
            <span className="flex-none font-semibold text-gray-400">{label}:</span>
            <span className="min-w-0 break-all">{value}</span>
        </div>
    )

    return (
        <div className="absolute left-4 top-24 z-30 max-w-lg rounded-md bg-black/80 p-4 font-mono text-xs leading-5 text-white backdrop-blur pointer-events-none select-none">
            <p className="font-bold mb-2">Stats for Nerds</p>
            <div className="space-y-1">
                <StatLine label="Source" value={props.info?.streamPath || props.info?.playbackUri || "unknown"} />
                <StatLine label="Display / Video" value={`${displaySize.width}x${displaySize.height} / ${videoWidth || "?"}x${videoHeight || "?"}`} />
                <StatLine label="Video" value={`${String(video?.codec ?? "unknown")}${videoBitrate > 0 ? ` - ${(videoBitrate / 1_000_000).toFixed(2)} Mbps` : ""}`} />
                <StatLine label="Audio" value={`${String(audio?.codec ?? "unknown")}${audioBitrate > 0 ? ` - ${(audioBitrate / 1000).toFixed(0)} kbps` : ""}`} />
                <StatLine label="Framerate" value={fps > 0 ? `${fps.toFixed(2)} fps` : "unknown"} />
                <StatLine label="Frames (Output / Decoder Drops)" value={`${outputDrops} / ${decoderDrops}`} />
                <StatLine label="Buffer Ahead" value={`${Math.max(0, cacheDuration).toFixed(2)} s${props.buffering ? " - buffering" : ""}`} />
                <StatLine label="Playback Rate" value={`${props.speed.toFixed(2)}x`} />
                <StatLine label="Time / Duration" value={`${props.currentTime.toFixed(2)}s / ${props.duration.toFixed(2)}s`} />
                <StatLine label="Hardware Decode" value={String(props.diagnostics["hwdec-current"] || "no")} />
                <StatLine label="Container" value={String(props.diagnostics["file-format"] || props.info?.mimeType || "unknown")} />
                {props.shaderMode === "anime4k" && (
                    <StatLine label="Shaders" value={`Anime4K (${props.anime4kMode}) - ${props.anime4kQuality.toUpperCase()}`} />
                )}
                {props.shaderMode === "custom" && (
                    <StatLine label="Shaders" value={`Custom (${props.customShadersCount} active)`} />
                )}
            </div>
        </div>
    )
}
