import { MediaCoreTimeRangeView } from "@/app/(main)/_features/media-core/media-core-control-bar"
import { useMpvPrismPlayer } from "@mpv-prism/react"
import { useAtomValue } from "jotai"
import React from "react"
import { getSkipChapters, type MediaCoreChapter } from "../media-core/media-core-chapters"
import { mediaCorePreferencesAtom } from "../media-core/media-core-preferences"
import { VideoCorePreviewManager } from "../video-core/video-core-preview"
import { mc_resolveSource, type MpvCoreChapterCue } from "./mpv-core"

export interface MpvCoreTimeRangeProps {
    player: ReturnType<typeof useMpvPrismPlayer>
    currentTime: number
    duration: number
    buffered: number
    showChapterMarkers: boolean
    highlightOPEDChapters: boolean
    paused: boolean
    chapters: MpvCoreChapterCue[]
    streamUrl?: string | null
    playbackType?: string | null
    isMiniPlayer?: boolean
}

const isMobile = false
const MOBILE_PREVIEW_THUMBNAIL_SIZE = 140
const VIDEOCORE_PREVIEW_THUMBNAIL_SIZE = 200
const VIDEOCORE_PREVIEW_CAPTURE_INTERVAL_SECONDS = 4
const PREVIEW_REQUEST_DELAY_MS = 80
const STALE_PREVIEW_GRACE_MS = 100

export function MpvCoreTimeRange({
    player,
    currentTime,
    duration,
    buffered,
    showChapterMarkers,
    highlightOPEDChapters,
    paused,
    chapters: chapterCues,
    streamUrl,
    playbackType,
    isMiniPlayer = false,
}: MpvCoreTimeRangeProps) {
    const [seeking, setSeeking] = React.useState(false)
    const [seekingTargetProgress, setSeekingTargetProgress] = React.useState(0)
    const dragWasPausedRef = React.useRef(paused)
    const skipPatterns = useAtomValue(mediaCorePreferencesAtom).skipPatterns

    const rangeRectRef = React.useRef<DOMRect | null>(null)
    const seekingTargetProgressRef = React.useRef(seekingTargetProgress)

    React.useEffect(() => {
        seekingTargetProgressRef.current = seekingTargetProgress
    }, [seekingTargetProgress])

    // Thumbnails
    const [previewThumbnail, setPreviewThumbnail] = React.useState<string | null>(null)
    const [previewManager, setPreviewManager] = React.useState<VideoCorePreviewManager | null>(null)
    const requestedSegmentRef = React.useRef<number | null>(null)
    const requestVersionRef = React.useRef(0)
    const pendingClientXRef = React.useRef<number | null>(null)
    const previewRafRef = React.useRef<number | null>(null)
    const previewRequestTimeoutRef = React.useRef<number | null>(null)
    const stalePreviewClearTimeoutRef = React.useRef<number | null>(null)

    React.useEffect(() => {
        if (!streamUrl) {
            setPreviewManager(null)
            return
        }

        const resolvedUrl = mc_resolveSource(streamUrl)
        const isHls = resolvedUrl.includes(".m3u8") || resolvedUrl.includes("/hls/")
        const streamType = isHls ? "hls" : "direct"

        const dummyVideo = document.createElement("video")

        const manager = new VideoCorePreviewManager(
            dummyVideo,
            resolvedUrl,
            streamType,
            playbackType !== "onlinestream"
        )

        setPreviewManager(manager)

        return () => {
            manager.cleanup()
        }
    }, [streamUrl, playbackType])

    const bufferedPercentage = React.useMemo(() => {
        return duration > 0 ? (buffered / duration) * 100 : 0
    }, [buffered, duration])

    const chapters = React.useMemo<Array<MediaCoreChapter & { width: number, percentageOffset: number }>>(() => {
        if (!chapterCues.length) return [{
            width: 100,
            percentageOffset: 0,
            label: null,
            start: 0,
            end: duration,
        }]

        return chapterCues
            .toSorted((a, b) => a.startTime - b.startTime)
            .filter(chapter =>
                (chapter.startTime || chapter.endTime) &&
                !(chapter.endTime > 0 && chapter.endTime < chapter.startTime),
            )
            .map(chapter => {
                const start = chapter.startTime ?? 0
                const end = chapter.endTime ?? duration
                const chapterDuration = end - start
                const width = duration > 0 ? (chapterDuration / duration) * 100 : 0
                return {
                    width,
                    percentageOffset: duration > 0 ? (start / duration) * 100 : 0,
                    label: chapter.text,
                    start,
                    end,
                }
            })
    }, [chapterCues, duration])
    const skipChapters = React.useMemo(() => getSkipChapters(chapters, skipPatterns, { guardIntro: false }), [chapters, skipPatterns])

    const [progressPercentage, setProgressPercentage] = React.useState(duration > 0 ? (currentTime / duration) * 100 : 0)

    React.useEffect(() => {
        if (!seeking) {
            setProgressPercentage(duration > 0 ? (currentTime / duration) * 100 : 0)
        }
    }, [currentTime, duration, seeking])

    function handlePointerDown(e: React.PointerEvent<HTMLDivElement>) {
        e.stopPropagation()
        if (!player) return
        if (e.pointerType === "mouse" && e.button !== 0) return
        if (e.pointerType === "touch") {
            e.preventDefault()
        }
        e.currentTarget.setPointerCapture(e.pointerId)
        rangeRectRef.current = e.currentTarget.getBoundingClientRect()
        setSeeking(true)
        dragWasPausedRef.current = paused
        if (!paused) player.pause()
        handlePointerMove(e)
    }

    function handlePointerUp(e: React.PointerEvent<HTMLDivElement>) {
        e.stopPropagation()
        if (!player) return
        if (e.pointerType === "mouse" && e.button !== 0) return
        if (e.pointerType === "touch") {
            e.preventDefault()
        }
        if (e.currentTarget.hasPointerCapture(e.pointerId)) {
            e.currentTarget.releasePointerCapture(e.pointerId)
        }
        rangeRectRef.current = null
        setSeeking(false)
        const targetTime = (duration * seekingTargetProgressRef.current) / 100
        player.seek(targetTime, "absolute+exact").then(() => {
            if (!dragWasPausedRef.current) player.play()
        })
    }

    function handlePointerCancel(e: React.PointerEvent<HTMLDivElement>) {
        if (!player) return
        if (e.currentTarget.hasPointerCapture(e.pointerId)) {
            e.currentTarget.releasePointerCapture(e.pointerId)
        }
        rangeRectRef.current = null
        if (seeking) {
            setSeeking(false)
            const targetTime = (duration * seekingTargetProgressRef.current) / 100
            player.seek(targetTime, "absolute+exact").then(() => {
                if (!dragWasPausedRef.current) player.play()
            })
        }
    }

    function handlePointerLeave(e: React.PointerEvent<HTMLDivElement>) {
        if (!seeking) {
            rangeRectRef.current = null
            if (seekingTargetProgressRef.current !== 0) {
                seekingTargetProgressRef.current = 0
                setSeekingTargetProgress(0)
            }
        }
    }

    function getPointerProgress<T extends HTMLElement>(e: React.PointerEvent<T>) {
        const rect = rangeRectRef.current ?? e.currentTarget.getBoundingClientRect()
        rangeRectRef.current = rect
        if (rect.width <= 0) return 0
        const x = e.clientX - rect.left
        return Math.max(0, Math.min(100, (x / rect.width * 100)))
    }

    function handlePointerMove(e: React.PointerEvent<HTMLDivElement>) {
        const target = getPointerProgress(e)
        if (seeking) {
            if (e.pointerType === "touch") {
                e.preventDefault()
            }
            e.stopPropagation()
            setProgressPercentage(prev => Math.abs(prev - target) < 0.1 ? prev : target)
        }
        if (Math.abs(seekingTargetProgressRef.current - target) >= 0.1) {
            seekingTargetProgressRef.current = target
            setSeekingTargetProgress(target)
        }
    }

    const [timeRangeWidth, setTimeRangeWidth] = React.useState(0)
    const [timeRangeElement, setTimeRangeElement] = React.useState<HTMLDivElement | null>(null)

    React.useEffect(() => {
        if (!timeRangeElement) return
        const updateWidth = (width: number) => {
            setTimeRangeWidth(prev => Math.abs(prev - width) < 0.5 ? prev : width)
        }
        updateWidth(timeRangeElement.getBoundingClientRect().width)
        const resizeObserver = new ResizeObserver(entries => {
            updateWidth(entries[0]?.contentRect.width ?? 0)
        })
        resizeObserver.observe(timeRangeElement)
        return () => resizeObserver.disconnect()
    }, [timeRangeElement])

    const previewWidth = isMobile ? MOBILE_PREVIEW_THUMBNAIL_SIZE : VIDEOCORE_PREVIEW_THUMBNAIL_SIZE
    const safeDuration = Number.isFinite(duration) && duration > 0 ? duration : 0

    const setPreview = React.useCallback((thumbnail: string | null) => {
        setPreviewThumbnail(prev => prev === thumbnail ? prev : thumbnail)
    }, [])

    const cancelStalePreviewClear = React.useCallback(() => {
        if (stalePreviewClearTimeoutRef.current !== null) {
            window.clearTimeout(stalePreviewClearTimeoutRef.current)
            stalePreviewClearTimeoutRef.current = null
        }
    }, [])

    const clearPreview = React.useCallback(() => {
        requestVersionRef.current += 1
        requestedSegmentRef.current = null
        pendingClientXRef.current = null
        cancelStalePreviewClear()

        if (previewRafRef.current !== null) {
            window.cancelAnimationFrame(previewRafRef.current)
            previewRafRef.current = null
        }

        if (previewRequestTimeoutRef.current !== null) {
            window.clearTimeout(previewRequestTimeoutRef.current)
            previewRequestTimeoutRef.current = null
        }

        setPreview(null)
    }, [cancelStalePreviewClear, setPreview])

    React.useEffect(() => {
        clearPreview()
    }, [previewManager, clearPreview])

    React.useEffect(() => {
        return () => clearPreview()
    }, [clearPreview])

    const requestPreview = React.useCallback((previewTime: number) => {
        if (!previewManager || !safeDuration || previewTime < 0 || previewTime > safeDuration) return

        const segmentIndex = Math.floor(previewTime / VIDEOCORE_PREVIEW_CAPTURE_INTERVAL_SECONDS)
        if (requestedSegmentRef.current === segmentIndex) return

        requestedSegmentRef.current = segmentIndex
        cancelStalePreviewClear()
        stalePreviewClearTimeoutRef.current = window.setTimeout(() => {
            stalePreviewClearTimeoutRef.current = null
            setPreview(null)
        }, STALE_PREVIEW_GRACE_MS)

        if (previewRequestTimeoutRef.current !== null) {
            window.clearTimeout(previewRequestTimeoutRef.current)
        }

        previewRequestTimeoutRef.current = window.setTimeout(async () => {
            previewRequestTimeoutRef.current = null
            const requestVersion = ++requestVersionRef.current

            try {
                const thumbnail = await previewManager.retrievePreviewForSegment(segmentIndex, false)
                if (requestVersion !== requestVersionRef.current || requestedSegmentRef.current !== segmentIndex) return
                if (!thumbnail) {
                    requestedSegmentRef.current = null
                    return
                }
                cancelStalePreviewClear()
                setPreview(thumbnail)
            }
            catch (error) {
                if (requestedSegmentRef.current === segmentIndex) {
                    requestedSegmentRef.current = null
                }
                console.error("Failed to get thumbnail", error)
            }
        }, PREVIEW_REQUEST_DELAY_MS)
    }, [cancelStalePreviewClear, previewManager, safeDuration, setPreview])

    const handleTimeRangePreview = React.useCallback((event: Event) => {
        if (!safeDuration || !timeRangeElement) return

        let clientX: number
        if (event instanceof TouchEvent && event.touches.length > 0) {
            clientX = event.touches[0].clientX
        } else if (event instanceof MouseEvent) {
            clientX = event.clientX
        } else {
            return
        }

        pendingClientXRef.current = clientX
        if (previewRafRef.current !== null) return

        previewRafRef.current = window.requestAnimationFrame(() => {
            previewRafRef.current = null
            const pendingClientX = pendingClientXRef.current
            if (pendingClientX === null) return

            const rect = timeRangeElement.getBoundingClientRect()
            if (rect.width <= 0) return

            setTimeRangeWidth(prev => Math.abs(prev - rect.width) < 0.5 ? prev : rect.width)

            const x = pendingClientX - rect.left
            const percentage = Math.max(0, Math.min(1, x / rect.width))
            requestPreview(percentage * safeDuration)
        })
    }, [requestPreview, safeDuration, timeRangeElement])

    React.useEffect(() => {
        if (!timeRangeElement) return

        const handleMouseLeave = () => {
            clearPreview()
        }

        const handleTouchEnd = () => {
            clearPreview()
        }

        if (isMobile) {
            timeRangeElement.addEventListener("touchmove", handleTimeRangePreview, { passive: true })
            timeRangeElement.addEventListener("touchend", handleTouchEnd)
        } else {
            timeRangeElement.addEventListener("mouseleave", handleMouseLeave)
            timeRangeElement.addEventListener("mousemove", handleTimeRangePreview)
        }

        return () => {
            if (isMobile) {
                timeRangeElement.removeEventListener("touchmove", handleTimeRangePreview)
                timeRangeElement.removeEventListener("touchend", handleTouchEnd)
            } else {
                timeRangeElement.removeEventListener("mouseleave", handleMouseLeave)
                timeRangeElement.removeEventListener("mousemove", handleTimeRangePreview)
            }
        }
    }, [clearPreview, handleTimeRangePreview, timeRangeElement])

    const targetProgress = React.useMemo(() => {
        if (!safeDuration) return 0
        const progress = seekingTargetProgress
        if (!Number.isFinite(progress)) return 0
        return Math.max(0, Math.min(100, progress))
    }, [safeDuration, seekingTargetProgress])

    const targetTime = React.useMemo(() => {
        if (!safeDuration) return 0
        return (safeDuration * targetProgress) / 100
    }, [safeDuration, targetProgress])

    const chapterLabel = React.useMemo(() => {
        const chapter = chapters.find(chapter =>
            chapter.percentageOffset <= targetProgress &&
            chapter.percentageOffset + chapter.width >= targetProgress,
        )
        return chapter?.label ?? null
    }, [targetProgress, chapters])

    const previewX = React.useMemo(() => {
        if (timeRangeWidth <= 0) return previewWidth / 2
        const x = (targetProgress / 100) * timeRangeWidth
        if (!Number.isFinite(x)) return previewWidth / 2
        if (timeRangeWidth <= previewWidth) return timeRangeWidth / 2
        return Math.max(previewWidth / 2, Math.min(timeRangeWidth - previewWidth / 2, x))
    }, [targetProgress, timeRangeWidth, previewWidth])

    const tooltipX = React.useMemo(() => {
        if (timeRangeWidth <= 0) return 0
        const x = (targetProgress / 100) * timeRangeWidth
        if (!Number.isFinite(x)) return 0
        if (timeRangeWidth <= 8 * 2) return timeRangeWidth / 2
        return Math.max(8, Math.min(timeRangeWidth - 8, x))
    }, [targetProgress, timeRangeWidth])

    const showPreview = !isMiniPlayer && !!previewManager && (seeking || targetProgress > 0)
    const showThumbnail = showPreview

    return (
        <MediaCoreTimeRangeView
            seeking={seeking}
            progressPercentage={progressPercentage}
            bufferedPercentage={bufferedPercentage}
            seekingTargetProgress={seekingTargetProgress}
            chapters={chapters}
            showChapterMarkers={showChapterMarkers}
            highlightChapters={highlightOPEDChapters}
            skipChapters={skipChapters}
            showPreview={showPreview}
            showThumbnail={showThumbnail}
            previewWidth={previewWidth}
            previewX={previewX}
            tooltipX={tooltipX}
            chapterLabel={chapterLabel}
            targetTime={targetTime}
            previewThumbnailUrl={previewThumbnail}
            timeRangeRef={setTimeRangeElement}
            onPointerDown={handlePointerDown}
            onPointerUp={handlePointerUp}
            onPointerMove={handlePointerMove}
            onPointerLeave={handlePointerLeave}
            onPointerCancel={handlePointerCancel}
            isMobile={false}
            duration={duration}
            onMarkerClick={(time) => {
                player?.seek(time, "absolute+exact")
            }}
        />
    )
}
