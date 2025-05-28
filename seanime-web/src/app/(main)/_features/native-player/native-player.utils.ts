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
    if (content.startsWith("<?xml")) {
        return "ttml"
    }
    if (content.startsWith("{")) {
        return "stl"
    }
    if (content.includes("-->")) {
        return "srt"
    }
    return "unknown"
}
