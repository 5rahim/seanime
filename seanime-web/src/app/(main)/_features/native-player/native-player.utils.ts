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
    content = content.trim()
    if (content.startsWith("[Script Info]")) {
        return "ass"
    }
    if (content.startsWith("WEBVTT")) {
        return "vtt"
    }
    if (content.startsWith("<?xml") || content.startsWith("<tt")) {
        return "ttml"
    }
    if (content.startsWith("{")) {
        return "stl"
    }
    if (content.includes(" -->")) {
        return "srt"
    }
    return "unknown"
}
