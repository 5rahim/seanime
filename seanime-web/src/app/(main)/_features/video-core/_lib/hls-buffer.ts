export function hasStartupGap(video: Pick<HTMLVideoElement, "currentTime" | "buffered">) {
    if (video.currentTime < 0 || video.currentTime > 0.1 || !video.buffered.length) return false

    const start = video.buffered.start(0)
    return start > video.currentTime && start < 0.5 && video.buffered.end(0) > start + 0.05
}
