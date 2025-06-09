import { MKVParser_ChapterInfo } from "@/api/generated/types"

export const nativeplayer_createChapterCues = (chapters: Array<MKVParser_ChapterInfo> | undefined, duration: number) => {
    if (!chapters || chapters.length === 0 || duration === 0) {
        return []
    }

    return chapters.map((chapter, index) => ({
        startTime: chapter.start / 1e6,
        endTime: chapter.end ? chapter.end / 1e6 : (chapters[index + 1]?.start ? chapters[index + 1].start / 1e6 : duration),
        text: chapter.text || `Chapter ${index + 1}`,
    }))
}

export const nativeplayer_createChapterVTT = (chapters: Array<MKVParser_ChapterInfo> | undefined, duration: number) => {
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
        vttContent += `${chapter.text || `Chapter ${index + 1}`}\n\n`
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

export function nativeplayer_createChaptersFromAniSkip(
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
            text: "Intro",
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
            text: ((duration) - aniSkipData.ed.interval.endTime) > 0.5 * 60 ? "Outro" : "Preview",
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
