import { MKVParser_ChapterInfo, NativePlayer_PlaybackInfo } from "@/api/generated/types"
import {
    vc_buffering,
    vc_currentTime,
    vc_duration,
    vc_ended,
    vc_isMuted,
    vc_paused,
    vc_playbackRate,
    vc_readyState,
    vc_timeRanges,
    vc_videoElement,
    vc_videoSize,
    vc_volume,
    VideoCoreChapterCue,
} from "@/app/(main)/_features/video-core/video-core"
import { useAtomValue } from "jotai"
import { useSetAtom } from "jotai/react"
import { useEffect } from "react"

export type VideoCoreChapter = {
    start: number
    end: number
    title: string
}

export function useVideoCoreBindings(playbackInfo: NativePlayer_PlaybackInfo | null | undefined) {

    const v = useAtomValue(vc_videoElement)
    const setVideoSize = useSetAtom(vc_videoSize)
    const setDuration = useSetAtom(vc_duration)
    const setCurrentTime = useSetAtom(vc_currentTime)
    const setPlaybackRate = useSetAtom(vc_playbackRate)
    const setReadyState = useSetAtom(vc_readyState)
    const setBuffering = useSetAtom(vc_buffering)
    const setIsMuted = useSetAtom(vc_isMuted)
    const setVolume = useSetAtom(vc_volume)
    const setBuffered = useSetAtom(vc_timeRanges)
    const setEnded = useSetAtom(vc_ended)
    const setPaused = useSetAtom(vc_paused)

    useEffect(() => {
        if (!v) return
        const handler = () => {
            setVideoSize({
                width: v.videoWidth,
                height: v.videoHeight,
            })
            setDuration(v.duration)
            setCurrentTime(v.currentTime)
            setPlaybackRate(v.playbackRate)
            setReadyState(v.readyState)
            // Set buffering to true if readyState is less than HAVE_ENOUGH_DATA (3) and video is not paused
            setBuffering(v.readyState < 3 && !v.paused)
            setIsMuted(v.muted)
            setVolume(v.volume)
            setBuffered(v.buffered.length > 0 ? v.buffered : null)
            setEnded(v.ended)
            setPaused(v.paused)
        }
        const events = ["timeupdate", "loadedmetadata", "progress", "play", "pause", "ratechange", "volumechange", "ended", "loadeddata", "resize",
            "waiting", "canplay", "stalled"]
        events.forEach(e => v.addEventListener(e, handler))
        handler() // initialize state once

        return () => {
            console.log("Removing video event listeners")
            events.forEach(e => v.removeEventListener(e, handler))
        }
    }, [v, playbackInfo])

}

export const vc_createChapterCues = (chapters: Array<MKVParser_ChapterInfo> | undefined, duration: number): VideoCoreChapterCue[] => {
    if (!chapters || chapters.length === 0 || duration === 0) {
        return []
    }

    return chapters.map((chapter, index) => ({
        startTime: chapter.start / 1e6,
        endTime: chapter.end ? chapter.end / 1e6 : (chapters[index + 1]?.start ? chapters[index + 1].start / 1e6 : duration),
        text: chapter.text || ``,
    }))
}

export const vc_createChapterVTT = (chapters: Array<MKVParser_ChapterInfo> | undefined, duration: number) => {
    if (!chapters || chapters.length === 0 || duration === 0) {
        return ""
    }

    let vttContent = "WEBVTT\n\n"

    chapters.forEach((chapter, index) => {
        const startTime = chapter.start / 1e6
        const endTime = chapter.end ? chapter.end / 1e6 : (chapters[index + 1]?.start ? chapters[index + 1].start / 1e6 : duration)

        const formatTime = (seconds: number) => {
            const hours = Math.floor(seconds / 3600)
            const minutes = Math.floor((seconds % 3600) / 60)
            const secs = (seconds % 60).toFixed(3)
            return `${hours.toString().padStart(2, "0")}:${minutes.toString().padStart(2, "0")}:${secs.padStart(6, "0")}`
        }

        vttContent += `${index + 1}\n`
        vttContent += `${formatTime(startTime)} --> ${formatTime(endTime)}\n`
        vttContent += `${chapter.text || ``}\n\n`
    })

    return vttContent
}

export function isSubtitleFile(filename: string) {
    const subRx = /\.srt$|\.ass$|\.ssa$|\.vtt$|\.txt$|\.ttml$|\.stl$/i
    return subRx.test(filename)
}

export function detectSubtitleType(content: string): "ass" | "vtt" | "ttml" | "stl" | "srt" | "unknown" {
    const trimmed = content.trim()

    // ASS/SSA: [Script Info] or [V4+ Styles] or [V4 Styles]
    if (
        /^\[Script Info\]/im.test(trimmed) ||
        /^\[V4\+ Styles\]/im.test(trimmed) ||
        /^\[V4 Styles\]/im.test(trimmed)
    ) {
        return "ass"
    }

    // VTT: WEBVTT at start, optionally with BOM or comments
    if (/^(?:\uFEFF)?WEBVTT\b/im.test(trimmed)) {
        return "vtt"
    }

    // TTML: XML root with <tt> or <tt:tt>
    if (
        /^<\?xml[\s\S]*?<tt[:\s>]/im.test(trimmed) ||
        /^<tt[:\s>]/im.test(trimmed)
    ) {
        return "ttml"
    }

    // STL: { ... } lines (MicroDVD/other curly-brace formats)
    if (/^\{\d+\}/m.test(trimmed)) {
        return "stl"
    }

    // SRT: 1\n00:00:00,000 --> 00:00:05,000
    if (
        /^\d+\s*\n\s*\d{2}:\d{2}:\d{2},\d{3}\s*-->\s*\d{2}:\d{2}:\d{2},\d{3}/m.test(trimmed) ||
        /\d{2}:\d{2}:\d{2},\d{3}\s*-->\s*\d{2}:\d{2}:\d{2},\d{3}/.test(trimmed)
    ) {
        return "srt"
    }

    // Fallback: check for VTT/SRT timecodes
    if (/\d{2}:\d{2}:\d{2}\.\d{3}\s*-->\s*\d{2}:\d{2}:\d{2}\.\d{3}/.test(trimmed)) {
        return "vtt"
    }
    if (/\d{2}:\d{2}:\d{2},\d{3}\s*-->\s*\d{2}:\d{2}:\d{2},\d{3}/.test(trimmed)) {
        return "srt"
    }

    return "unknown"
}

export function vc_createChaptersFromAniSkip(
    aniSkipData: {
        op: { interval: { startTime: number; endTime: number } } | null;
        ed: { interval: { startTime: number; endTime: number } } | null
    } | undefined,
    duration: number,
    mediaFormat?: string,
): Array<MKVParser_ChapterInfo> {
    if (!aniSkipData?.op?.interval || duration <= 0) {
        return []
    }

    let chapters: MKVParser_ChapterInfo[] = []

    if (aniSkipData?.op?.interval) {
        chapters.push({
            uid: 91,
            start: aniSkipData.op.interval.startTime > 5 ? aniSkipData.op.interval.startTime : 0,
            end: aniSkipData.op.interval.endTime,
            text: "Opening",
        })
    }

    if (aniSkipData?.ed?.interval) {
        chapters.push({
            uid: 92,
            start: aniSkipData.ed.interval.startTime,
            end: aniSkipData.ed.interval.endTime,
            text: "Ending",
        })
    }

    if (chapters.length === 0) return []

    // Add beginning chapter
    if (aniSkipData.op?.interval?.startTime > 5) {
        chapters.push({
            uid: 90,
            start: 0,
            end: aniSkipData.op.interval.startTime,
            // text: aniSkipData.op.interval.startTime > 1.5 * 60 ? "Intro" : "Recap",
            text: "Prologue",
        })
    }

    // Add middle chapter
    chapters.push({
        uid: 93,
        start: aniSkipData.op?.interval?.endTime || 0,
        end: aniSkipData.ed?.interval?.startTime || duration,
        text: mediaFormat !== "MOVIE" ? "Episode" : "Movie",
    })

    // Add ending chapter
    if (aniSkipData.ed?.interval?.endTime && aniSkipData.ed.interval.endTime < duration - 5) {
        chapters.push({
            uid: 94,
            start: aniSkipData.ed.interval.endTime,
            end: duration,
            text: ((duration) - aniSkipData.ed.interval.endTime) > 0.5 * 60 ? "Ending" : "Preview",
        })
    }

    chapters.sort((a, b) => a.start - b.start)
    // Make sure last chapter is clamped to the end of the video
    if (chapters.length > 0) {
        chapters[chapters.length - 1].end = duration
    }

    chapters = chapters.map((chapter, index) => ({
        ...chapter,
        start: chapter.start * 1e6,
        end: chapter.end ? index === chapters.length - 1 ? duration * 1e6 : chapter.end * 1e6 : undefined,
    }))

    return chapters
}

export const vc_formatTime = (seconds: number) => {
    const sign = seconds < 0 ? "-" : ""
    const absSeconds = Math.abs(seconds)
    const hours = Math.floor(absSeconds / 3600)
    const minutes = Math.floor((absSeconds % 3600) / 60)
    const secs = Math.floor(absSeconds % 60)

    if (hours > 0) {
        return `${sign}${hours}:${minutes.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`
    }
    return `${sign}${minutes}:${secs.toString().padStart(2, "0")}`
}

export const vc_logGeneralInfo = (video: HTMLVideoElement | null) => {
    if (!video) return
    // MP4 container codec tests
    console.log("HEVC main ->", video.canPlayType("video/mp4;codecs=\"hev1.1.6.L120.90\"") || "❌")
    console.log("HEVC main 10 ->", video.canPlayType("video/mp4;codecs=\"hev1.2.4.L120.90\"") || "❌")
    console.log("HEVC main still-picture ->", video.canPlayType("video/mp4;codecs=\"hev1.3.E.L120.90\"") || "❌")
    console.log("HEVC range extensions ->", video.canPlayType("video/mp4;codecs=\"hev1.4.10.L120.90\"") || "❌")

    // Audio codec tests
    console.log("Dolby AC3 ->", video.canPlayType("audio/mp4; codecs=\"ac-3\"") || "❌")
    console.log("Dolby EC3 ->", video.canPlayType("audio/mp4; codecs=\"ec-3\"") || "❌")

    // GPU and hardware acceleration status
    const canvas = document.createElement("canvas")
    const gl = canvas.getContext("webgl2") || canvas.getContext("webgl")
    if (gl) {
        const debugInfo = gl.getExtension("WEBGL_debug_renderer_info")
        if (debugInfo) {
            console.log("GPU Vendor ->", gl.getParameter(debugInfo.UNMASKED_VENDOR_WEBGL))
            console.log("GPU Renderer ->", gl.getParameter(debugInfo.UNMASKED_RENDERER_WEBGL))
        }
    }
    console.log("Hardware concurrency ->", navigator.hardwareConcurrency)
    console.log("User agent ->", navigator.userAgent)

    // Web GPU
    if (navigator.gpu) {
        navigator.gpu.requestAdapter().then(adapter => {
            if (adapter) {
                console.log("WebGPU adapter ->", adapter)
                console.log("WebGPU adapter features ->", adapter.features)
            } else {
                console.log("⚠️ No WebGPU adapter found.")
            }
        })
    } else {
        console.log("❌ WebGPU not supported.")
    }
}
