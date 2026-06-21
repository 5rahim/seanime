import { useNakamaWatchParty } from "@/app/(main)/_features/nakama/nakama-manager"
import { vc_previewManager } from "@/app/(main)/_features/video-core/video-core"
import { VideoCoreChapterCue } from "@/app/(main)/_features/video-core/video-core"
import { vc_isMobile } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_isSwiping } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_swipeSeekTime } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_duration } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_currentTime } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_seeking } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_seekingTargetProgress } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_closestBufferedTime } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_miniPlayer } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_videoElement } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_previousPausedState } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_lastKnownProgress } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_skipOpeningTime } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_skipEndingTime } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_showOverlayFeedback } from "@/app/(main)/_features/video-core/video-core-overlay-display"
import { VIDEOCORE_PREVIEW_CAPTURE_INTERVAL_SECONDS, VIDEOCORE_PREVIEW_THUMBNAIL_SIZE } from "@/app/(main)/_features/video-core/video-core-preview"
import { vc_autoSkipOPEDAtom, vc_highlightOPEDChaptersAtom, vc_showChapterMarkersAtom } from "@/app/(main)/_features/video-core/video-core.atoms"
import { vc_dispatchAction } from "@/app/(main)/_features/video-core/video-core.utils"
import { vc_getOPEDChapters } from "@/app/(main)/_features/video-core/video-core.utils"
import { MediaCoreTimeRangeView } from "@/app/(main)/_features/media-core/media-core-control-bar"
import { atom, useAtomValue } from "jotai"
import { useAtom, useSetAtom } from "jotai/react"
import React from "react"

export type VideoCoreTimeRangeChapter = {
    width: number
    percentageOffset: number
    label: string | null
    start: number
    end: number
}

export interface VideoCoreTimeRangeProps {
    chapterCues: VideoCoreChapterCue[]
}

const CHAPTER_GAP = 3
const MOBILE_PREVIEW_THUMBNAIL_SIZE = 140
const PREVIEW_REQUEST_DELAY_MS = 80
const STALE_PREVIEW_GRACE_MS = 100
const PREVIEW_TOOLTIP_GUTTER = 8

export const vc_timeRangeElement = atom<HTMLDivElement | null>(null)

export function VideoCoreTimeRange(props: VideoCoreTimeRangeProps) {
    const { chapterCues } = props

    const { isPeer: isWatchPartyPeer } = useNakamaWatchParty()

    const videoElement = useAtomValue(vc_videoElement)
    const isMobile = useAtomValue(vc_isMobile)
    const isSwiping = useAtomValue(vc_isSwiping)
    const swipeSeekTime = useAtomValue(vc_swipeSeekTime)

    const currentTime = useAtomValue(vc_currentTime)
    const duration = useAtomValue(vc_duration)
    const buffered = useAtomValue(vc_closestBufferedTime)
    const [seekingTargetProgress, setSeekingTargetProgress] = useAtom(vc_seekingTargetProgress)
    const [seeking, setSeeking] = useAtom(vc_seeking)
    const [previouslyPaused, setPreviouslyPaused] = useAtom(vc_previousPausedState)
    const action = useSetAtom(vc_dispatchAction)
    const showChapterMarkers = useAtomValue(vc_showChapterMarkersAtom)
    const highlightOPEDChapters = useAtomValue(vc_highlightOPEDChaptersAtom)
    const autoSkipIntroOutro = useAtomValue(vc_autoSkipOPEDAtom)
    const showOverlayFeedback = useSetAtom(vc_showOverlayFeedback)
    const [skipOpeningTime, setSkipOpeningTime] = useAtom(vc_skipOpeningTime)
    const [skipEndingTime, setSkipEndingTime] = useAtom(vc_skipEndingTime)
    const [restoreProgressTo, setRestoreProgressTo] = useAtom(vc_lastKnownProgress)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)

    const [timeRangeElement, setTimeRangeElement] = useAtom(vc_timeRangeElement)

    const rangeRectRef = React.useRef<DOMRect | null>(null)
    const seekingTargetProgressRef = React.useRef(seekingTargetProgress)

    React.useEffect(() => {
        seekingTargetProgressRef.current = seekingTargetProgress
    }, [seekingTargetProgress])

    const bufferedPercentage = React.useMemo(() => {
        return duration > 0 ? (buffered / duration) * 100 : 0
    }, [buffered, duration])

    const chapters = React.useMemo<VideoCoreTimeRangeChapter[]>(() => {
        if (!chapterCues?.length) return [{
            width: 100,
            percentageOffset: 0,
            label: null,
            start: 0,
            end: 0,
        }]

        let percentageOffset = 0
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
                const result = {
                    width,
                    percentageOffset,
                    label: chapter.text || null,
                    start,
                    end,
                }
                percentageOffset += width
                return result
            })
    }, [chapterCues, duration])

    const [progressPercentage, setProgressPercentage] = React.useState(duration > 0 ? (currentTime / duration) * 100 : 0)

    React.useEffect(() => {
        const timeToUse = isSwiping && swipeSeekTime !== null ? swipeSeekTime : currentTime
        setProgressPercentage(duration > 0 ? (timeToUse / duration) * 100 : 0)
    }, [currentTime, duration, isSwiping, swipeSeekTime])

    const opEdChapters = React.useMemo(() => vc_getOPEDChapters(chapters), [chapters])

    // handle auto skip
    React.useEffect(() => {
        if (!opEdChapters.opening?.end && !opEdChapters.ending?.end) return
        if (isNaN(duration) || duration <= 1) return
        if (isWatchPartyPeer) {
            setSkipOpeningTime(0)
            setSkipEndingTime(0)
            return
        }

        if (
            opEdChapters.opening &&
            opEdChapters.opening.end &&
            currentTime >= opEdChapters.opening.start &&
            currentTime < opEdChapters.opening.end
        ) {
            if (autoSkipIntroOutro && !restoreProgressTo) {
                action({ type: "seekTo", payload: { time: opEdChapters.opening.end } })
                showOverlayFeedback({ message: "Skipped OP", duration: 1000 })
            } else {
                setSkipOpeningTime(opEdChapters.opening.end)
            }
        } else {
            setSkipOpeningTime(0)
        }

        if (
            opEdChapters.ending &&
            opEdChapters.ending.end &&
            currentTime >= opEdChapters.ending.start &&
            currentTime < opEdChapters.ending.end &&
            currentTime < duration
        ) {
            if (autoSkipIntroOutro && !restoreProgressTo) {
                action({ type: "seekTo", payload: { time: opEdChapters.ending.end } })
                showOverlayFeedback({ message: "Skipped ED", duration: 1000 })
            } else {
                setSkipEndingTime(opEdChapters.ending.end)
            }
        } else {
            setSkipEndingTime(0)
        }

    }, [currentTime, autoSkipIntroOutro, opEdChapters, duration, restoreProgressTo, isWatchPartyPeer])

    // start seeking
    function handlePointerDown(e: React.PointerEvent<HTMLDivElement>) {
        e.stopPropagation()
        if (!videoElement) return
        if (e.pointerType === "mouse" && e.button !== 0) return
        if (e.pointerType === "touch") {
            e.preventDefault()
        }
        e.currentTarget.setPointerCapture(e.pointerId)
        rangeRectRef.current = e.currentTarget.getBoundingClientRect()
        setSeeking(true)

        if (isWatchPartyPeer) return

        setPreviouslyPaused(videoElement.paused)
        if (!videoElement.paused) videoElement.pause()
        handlePointerMove(e)
        videoElement?.dispatchEvent(new Event("seeking"))
    }

    // stop seeking
    function handlePointerUp(e: React.PointerEvent<HTMLDivElement>) {
        e.stopPropagation()
        if (!videoElement) return
        if (e.pointerType === "mouse" && e.button !== 0) return
        if (e.pointerType === "touch") {
            e.preventDefault()
        }
        if (e.currentTarget.hasPointerCapture(e.pointerId)) {
            e.currentTarget.releasePointerCapture(e.pointerId)
        }
        rangeRectRef.current = null
        setSeeking(false)
        if (!isWatchPartyPeer) {
            action({ type: "seekTo", payload: { time: (duration * seekingTargetProgressRef.current) / 100 } })
        }
        if (!previouslyPaused) videoElement?.play()?.catch()
    }

    // handle interrupted touch/pointer
    function handlePointerCancel(e: React.PointerEvent<HTMLDivElement>) {
        if (!videoElement) return
        if (e.currentTarget.hasPointerCapture(e.pointerId)) {
            e.currentTarget.releasePointerCapture(e.pointerId)
        }
        rangeRectRef.current = null
        if (seeking) {
            setSeeking(false)
            action({ type: "seekTo", payload: { time: (duration * seekingTargetProgressRef.current) / 100 } })
            if (!previouslyPaused) videoElement?.play()?.catch()
        }
    }

    // stop seeking
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

    // move progress
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

    const combineRef = React.useCallback((instance: HTMLDivElement | null) => {
        setTimeRangeElement(instance)
    }, [setTimeRangeElement])

    // Thumbnail & preview states
    const previewManager = useAtomValue(vc_previewManager)
    const [previewThumbnail, setPreviewThumbnail] = React.useState<string | null>(null)
    const [timeRangeWidth, setTimeRangeWidth] = React.useState(0)
    const requestedSegmentRef = React.useRef<number | null>(null)
    const requestVersionRef = React.useRef(0)
    const pendingClientXRef = React.useRef<number | null>(null)
    const previewRafRef = React.useRef<number | null>(null)
    const previewRequestTimeoutRef = React.useRef<number | null>(null)
    const stalePreviewClearTimeoutRef = React.useRef<number | null>(null)

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

    React.useEffect(() => {
        clearPreview()
    }, [previewManager, clearPreview])

    React.useEffect(() => {
        return () => clearPreview()
    }, [clearPreview])

    const targetProgress = React.useMemo(() => {
        if (!safeDuration) return 0
        const progress = isSwiping && swipeSeekTime !== null
            ? (swipeSeekTime / safeDuration) * 100
            : seekingTargetProgress
        if (!Number.isFinite(progress)) return 0
        return Math.max(0, Math.min(100, progress))
    }, [isSwiping, swipeSeekTime, safeDuration, seekingTargetProgress])

    const targetTime = React.useMemo(() => {
        if (!safeDuration) return 0
        if (isSwiping && swipeSeekTime !== null) {
            return Math.max(0, Math.min(safeDuration, swipeSeekTime))
        }
        return (safeDuration * targetProgress) / 100
    }, [isSwiping, swipeSeekTime, safeDuration, targetProgress])

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
        if (timeRangeWidth <= PREVIEW_TOOLTIP_GUTTER * 2) return timeRangeWidth / 2
        return Math.max(PREVIEW_TOOLTIP_GUTTER, Math.min(timeRangeWidth - PREVIEW_TOOLTIP_GUTTER, x))
    }, [targetProgress, timeRangeWidth])

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
        if (!isSwiping || swipeSeekTime === null || !safeDuration) {
            return
        }
        requestPreview(Math.max(0, Math.min(safeDuration, swipeSeekTime)))
    }, [isSwiping, swipeSeekTime, safeDuration, requestPreview])

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
    }, [clearPreview, handleTimeRangePreview, timeRangeElement, isMobile])

    React.useEffect(() => {
        if (!isSwiping) {
            clearPreview()
        }
    }, [clearPreview, isSwiping])

    const showPreview = !isMiniPlayer && !!previewManager && (seeking || isSwiping || targetTime > 0)
    const showThumbnail = showPreview

    return (
        <MediaCoreTimeRangeView
            seeking={seeking}
            progressPercentage={progressPercentage}
            bufferedPercentage={bufferedPercentage}
            seekingTargetProgress={seekingTargetProgress}
            chapters={chapters}
            showChapterMarkers={showChapterMarkers}
            highlightOPEDChapters={highlightOPEDChapters}
            showPreview={showPreview}
            showThumbnail={showThumbnail}
            previewWidth={previewWidth}
            previewX={previewX}
            tooltipX={tooltipX}
            chapterLabel={chapterLabel}
            targetTime={targetTime}
            previewThumbnailUrl={previewThumbnail}
            timeRangeRef={combineRef}
            onPointerDown={handlePointerDown}
            onPointerUp={handlePointerUp}
            onPointerMove={handlePointerMove}
            onPointerLeave={handlePointerLeave}
            onPointerCancel={handlePointerCancel}
            isMobile={isMobile}
            duration={duration}
            onMarkerClick={(time) => {
                action({ type: "seekTo", payload: { time } })
            }}
        />
    )
}
