export function checkCodecSupport(
    codec: string,
    options: {
        isMobile: boolean
        canUseMatroskaFallback: boolean
        canPlayType: (codec: string) => "probably" | "maybe" | ""
    },
): boolean {
    if (!codec) return false
    if (options.isMobile) return false

    const isMatroska = codec.startsWith("video/x-matroska") || codec.startsWith("video/matroska")
    if (options.canPlayType(codec) === "probably") {
        return true
    }

    if (isMatroska && options.canUseMatroskaFallback) {
        const container = codec.startsWith("video/x-matroska") ? "video/x-matroska" : "video/matroska"
        const mp4 = replaceMimeContainer(codec, container, "video/mp4")
        const webm = replaceMimeContainer(codec, container, "video/webm")
        return options.canPlayType(mp4) === "probably" || options.canPlayType(webm) === "probably"
    }

    return false
}

function replaceMimeContainer(codec: string, from: string, to: string): string {
    if (codec.startsWith(from)) {
        return to + codec.substring(from.length)
    }
    return codec
}
