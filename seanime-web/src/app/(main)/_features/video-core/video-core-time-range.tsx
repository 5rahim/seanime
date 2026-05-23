import { useNakamaWatchParty } from "@/app/(main)/_features/nakama/nakama-manager"
import { vc_previewManager } from "@/app/(main)/_features/video-core/video-core"
import { VIDEOCORE_DEBUG_ELEMENTS, VideoCoreChapterCue } from "@/app/(main)/_features/video-core/video-core"
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
import { vc_formatTime, vc_getChapterType, vc_getOPEDChapters } from "@/app/(main)/_features/video-core/video-core.utils"
import { SeaImage as Image } from "@/components/shared/sea-image"
import { cn } from "@/components/ui/core/styling"
import { logger } from "@/lib/helpers/debug"
import { atom, useAtomValue } from "jotai"
import { useAtom, useSetAtom } from "jotai/react"
import React from "react"
import { FaDiamond } from "react-icons/fa6"

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

export const vc_timeRangeElement = atom<HTMLDivElement | null>(null)

export function VideoCoreTimeRange(props: VideoCoreTimeRangeProps) {
    const {
        chapterCues,
    } = props

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
    const autoSkipIntroOutro = useAtomValue(vc_autoSkipOPEDAtom)
    const showOverlayFeedback = useSetAtom(vc_showOverlayFeedback)
    const [skipOpeningTime, setSkipOpeningTime] = useAtom(vc_skipOpeningTime)
    const [skipEndingTime, setSkipEndingTime] = useAtom(vc_skipEndingTime)
    const [restoreProgressTo, setRestoreProgressTo] = useAtom(vc_lastKnownProgress)
    const rangeRectRef = React.useRef<DOMRect | null>(null)
    const seekingTargetProgressRef = React.useRef(seekingTargetProgress)

    React.useEffect(() => {
        seekingTargetProgressRef.current = seekingTargetProgress
    }, [seekingTargetProgress])

    const bufferedPercentage = React.useMemo(() => {
        return (buffered / duration) * 100
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
                const width = (chapterDuration / duration) * 100
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

    const [progressPercentage, setProgressPercentage] = React.useState((currentTime / duration) * 100)

    React.useEffect(() => {
        const timeToUse = isSwiping && swipeSeekTime !== null ? swipeSeekTime : currentTime
        setProgressPercentage((timeToUse / duration) * 100)
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

        // e.currentTarget.currentTime >= aniSkipData.op.interval.startTime &&
        //             e.currentTarget.currentTime < aniSkipData.op.interval.endTime
        if (
            opEdChapters.opening &&
            opEdChapters.opening.end &&
            currentTime >= opEdChapters.opening.start &&
            currentTime < opEdChapters.opening.end
        ) {
            if (autoSkipIntroOutro && !restoreProgressTo) {
                console.log("auto skip", opEdChapters.opening.end)
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
                console.log("auto skip", opEdChapters.ending.end)
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
        // only check button for mouse events (touch events have button=-1)
        if (e.pointerType === "mouse" && e.button !== 0) return
        // prevent default touch behavior (scrolling, text selection)
        if (e.pointerType === "touch") {
            e.preventDefault()
        }
        e.currentTarget.setPointerCapture(e.pointerId) // capture movement outside
        rangeRectRef.current = e.currentTarget.getBoundingClientRect()
        setSeeking(true)

        if (isWatchPartyPeer) return

        // pause while seeking
        setPreviouslyPaused(videoElement.paused)
        if (!videoElement.paused) videoElement.pause()
        // move the progress
        handlePointerMove(e)
        videoElement?.dispatchEvent(new Event("seeking"))
    }

    // stop seeking
    function handlePointerUp(e: React.PointerEvent<HTMLDivElement>) {
        e.stopPropagation()
        if (!videoElement) return
        // only check button for mouse events (touch events have button=-1)
        if (e.pointerType === "mouse" && e.button !== 0) return
        // prevent default touch behavior
        if (e.pointerType === "touch") {
            e.preventDefault()
        }
        if (e.currentTarget.hasPointerCapture(e.pointerId)) {
            e.currentTarget.releasePointerCapture(e.pointerId)
        }
        rangeRectRef.current = null
        setSeeking(false)
        if (!isWatchPartyPeer) {
            // actually seek the video
            action({ type: "seekTo", payload: { time: (duration * seekingTargetProgressRef.current) / 100 } })
        }
        // resume playing
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
        // don't reset while actively seeking
        if (!seeking) {
            rangeRectRef.current = null
            if (seekingTargetProgressRef.current !== 0) {
                seekingTargetProgressRef.current = 0
                setSeekingTargetProgress(0)
            }
        }
    }

    function getPointerProgress<T extends HTMLElement>(e: React.PointerEvent<T>) { // 0-100
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
            // prevent page scroll during seeking
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

    const setTimeRangeElement = useSetAtom(vc_timeRangeElement)
    const combineRef = React.useCallback((instance: HTMLDivElement | null) => {
        // if (ref as unknown instanceof Function) (ref as any)(instance)
        // else if (ref) (ref as any).current = instance
        // if (instance) measureRef(instance)
        setTimeRangeElement(instance)
    }, [setTimeRangeElement])

    return (
        <div
            ref={combineRef}
            data-vc-element="time-range"
            data-vc-seeking-target-state={seeking}
            className={cn(
                "w-full relative group/vc-time-range z-[2] flex h-8",
                "cursor-pointer outline-none",
                "touch-none select-none [contain:layout_style]", // prevent page scroll and text selection on mobile
            )}
            role="slider"
            tabIndex={0}
            aria-valuemin={0}
            aria-valuenow={0}
            aria-valuetext="0%"
            aria-orientation="horizontal"
            aria-label="Video playback time"
            onPointerDown={handlePointerDown}
            onPointerUp={handlePointerUp}
            onPointerLeave={handlePointerLeave}
            onPointerCancel={handlePointerCancel}
            onPointerMove={handlePointerMove}
        >

            <VideoCoreTimePreview chapters={chapters} />

            {chapters.map((chapter, i) => {
                return <VideoCoreTimeRangeSegment
                    key={i}
                    idx={i}
                    progressPercentage={progressPercentage}
                    bufferedPercentage={bufferedPercentage}
                    chapter={chapter}
                    showMarker={i < chapters.length - 1 && showChapterMarkers && !isMobile}
                />
            })}

        </div>
    )
}

function VideoCoreTimeRangeSegment(props: {
    idx: number,
    progressPercentage: number,
    bufferedPercentage: number,
    chapter: VideoCoreTimeRangeChapter,
    showMarker: boolean,
}) {
    const { idx, chapter, progressPercentage, bufferedPercentage, showMarker } = props

    const duration = useAtomValue(vc_duration)
    const seekingTargetProgress = useAtomValue(vc_seekingTargetProgress)
    const action = useSetAtom(vc_dispatchAction)
    const highlightOPEDChapters = useAtomValue(vc_highlightOPEDChaptersAtom)

    const focused = !!seekingTargetProgress && chapter.percentageOffset <= seekingTargetProgress && chapter.percentageOffset + chapter.width >= seekingTargetProgress

    // returns x position of the bar in percentage
    function getChapterBarPosition(chapter: VideoCoreTimeRangeChapter, percentage: number) {
        const ret = (percentage - chapter.percentageOffset) * (100 / chapter.width)
        return (ret <= 0 ? -CHAPTER_GAP : ret >= 100 ? 100 : ret) - 100
    }

    return (
        <div
            data-vc-element="time-range-segment"
            data-vc-chapter-label={chapter.label}
            data-vc-focused-state={focused}
            data-vc-seeking-target-state={!!seekingTargetProgress}
            className={cn(
                "relative",
                "w-full h-full flex items-center",
                VIDEOCORE_DEBUG_ELEMENTS && "bg-yellow-500/10",
            )}
            style={{
                width: `${chapter.width}%`,
            }}
        >
            <div
                data-vc-element="time-range-chapter"
                className={cn(
                    "relative h-1 transition-transform transform-gpu origin-center will-change-transform flex items-center justify-center overflow-hidden rounded-lg",
                    focused && "scale-y-[2]",
                    VIDEOCORE_DEBUG_ELEMENTS && "bg-yellow-500/50",
                )}
                style={{
                    width: idx > 0 ? `calc(100% - ${CHAPTER_GAP / 2}px)` : `100%`,
                    marginLeft: idx > 0 ? `${CHAPTER_GAP}px` : `0px`,
                }}
            >
                <div
                    data-vc-element="time-range-chapter-bar"
                    data-vc-for="progress"
                    className={cn(
                        "bg-white absolute w-full h-full left-0 transform-gpu z-[10]",
                    )}
                    style={{
                        "--tw-translate-x": duration > 1 ? `${getChapterBarPosition(chapter, progressPercentage)}%` : "-100%",
                    } as React.CSSProperties}
                />
                <div
                    data-vc-element="time-range-chapter-bar"
                    data-vc-for="seeking-target"
                    className={cn(
                        "bg-white/30 absolute w-full h-full left-0 transform-gpu z-[9]",
                    )}
                    style={{
                        "--tw-translate-x": duration > 1 ? `${getChapterBarPosition(chapter, seekingTargetProgress)}%` : "-100%",
                    } as React.CSSProperties}
                />
                <div
                    data-vc-element="time-range-chapter-bar"
                    data-vc-for="buffer"
                    className={cn(
                        "bg-white/10 absolute w-full h-full left-0 transform-gpu z-[8]",
                    )}
                    style={{
                        "--tw-translate-x": duration > 1 ? `${getChapterBarPosition(chapter, bufferedPercentage)}%` : "-100%",
                    } as React.CSSProperties}
                />
                <div
                    data-vc-element="time-range-chapter-bar"
                    data-vc-for="main"
                    data-vc-highlighted-state={!!vc_getChapterType(chapter.label) && highlightOPEDChapters}
                    className={cn(
                        "bg-white/20 absolute left-0 w-full h-full z-[1]",
                        (!!vc_getChapterType(chapter.label) && highlightOPEDChapters) && "bg-blue-300/50",
                    )}
                />
            </div>
            {showMarker && (
                <button
                    data-vc-element="time-range-chapter-marker"
                    type="button"
                    onPointerDown={e => e.stopPropagation()}
                    onPointerUp={e => e.stopPropagation()}
                    onClick={e => {
                        e.stopPropagation()
                        action({ type: "seekTo", payload: { time: ((duration * (chapter.percentageOffset + chapter.width))) / 100 } })
                    }}
                    className={cn(
                        "absolute top-0 right-0 size-4 flex items-center justify-center -translate-y-1/2 translate-x-1/2 cursor-pointer z-[20] ",
                    )}
                    style={{
                        right: `-${CHAPTER_GAP / 2}px`,
                    }}
                    aria-label={`Seek to end of chapter ${idx + 1}`}
                    tabIndex={-1}
                >
                    <FaDiamond className="size-2.5 text-white/20 hover:text-white/100 transition-colors duration-100" />
                </button>
            )}
        </div>
    )
}

const timeRangeLog = logger("VIDEO CORE TIME RANGE")
const MOBILE_PREVIEW_THUMBNAIL_SIZE = 140
const PREVIEW_REQUEST_DELAY_MS = 80
const STALE_PREVIEW_GRACE_MS = 100
const PREVIEW_TOOLTIP_GUTTER = 8

function VideoCoreTimePreview(props: { chapters: VideoCoreTimeRangeChapter[] }) {
    const { chapters } = props

    const isMobile = useAtomValue(vc_isMobile)

    const duration = useAtomValue(vc_duration)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const seekingTargetProgress = useAtomValue(vc_seekingTargetProgress)
    const seeking = useAtomValue(vc_seeking)
    const isSwiping = useAtomValue(vc_isSwiping)
    const swipeSeekTime = useAtomValue(vc_swipeSeekTime)
    const previewManager = useAtomValue(vc_previewManager)
    const timeRangeElement = useAtomValue(vc_timeRangeElement)

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
        return chapter?.label
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
                timeRangeLog.error("Failed to get thumbnail", error)
            }
        }, PREVIEW_REQUEST_DELAY_MS)
    }, [cancelStalePreviewClear, previewManager, safeDuration, setPreview])

    const handleTimeRangePreview = React.useCallback((event: Event) => {
        if (!safeDuration || !timeRangeElement) {
            return
        }

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
        if (!timeRangeElement) {
            return
        }

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

    return <>

        {showThumbnail && <div
            data-vc-element="preview-thumbnail"
            className={cn(
                "absolute left-0 bottom-full aspect-video overflow-hidden rounded-md bg-black border border-white/50 pointer-events-none will-change-transform [contain:layout_paint_style]",
            )}
            style={{
                width: previewWidth,
                transform: `translate3d(${previewX - previewWidth / 2}px, ${isMobile ? "-64%" : "-54%"}, 0)`,
            }}
        >
            {!!previewThumbnail && <Image
                data-vc-element="preview-thumbnail-image"
                src={previewThumbnail || ""}
                alt="Preview"
                fill
                sizes={previewWidth + "px"}
                className="object-cover rounded-md"
                decoding="async"
                loading="lazy"
            />}
        </div>}

        {showPreview && <div
            data-vc-element="preview-tooltip"
            className={cn(
                "absolute left-0 bottom-full mb-3 px-2 py-1 bg-black/70 text-white text-center text-sm rounded-md will-change-transform",
                "whitespace-nowrap z-20 pointer-events-none",
                "",
            )}
            style={{
                transform: `translate3d(${tooltipX}px, 0, 0) translateX(-50%)`,
            }}
        >
            {chapterLabel && <p data-vc-element="preview-tooltip-chapter" className="text-xs font-medium max-w-2xl truncate">{chapterLabel}</p>}
            <p data-vc-element="preview-tooltip-time">{vc_formatTime(targetTime)}</p>

            <div
                data-vc-element="preview-tooltip-arrow"
                className="absolute top-full left-1/2 -translate-x-1/2 w-0 h-0 border-l-4 border-r-4 border-t-4 border-transparent border-t-black/70"
            />
        </div>}
    </>
}
