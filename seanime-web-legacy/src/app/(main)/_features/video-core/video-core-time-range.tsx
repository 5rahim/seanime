import { useNakamaWatchParty } from "@/app/(main)/_features/nakama/nakama-manager"
import {
    vc_closestBufferedTime,
    vc_currentTime,
    vc_dispatchAction,
    vc_duration,
    vc_isMobile,
    vc_isSwiping,
    vc_lastKnownProgress,
    vc_miniPlayer,
    vc_previewManager,
    vc_previousPausedState,
    vc_seeking,
    vc_seekingTargetProgress,
    vc_skipEndingTime,
    vc_skipOpeningTime,
    vc_swipeSeekTime,
    vc_videoElement,
    VIDEOCORE_DEBUG_ELEMENTS,
    VideoCoreChapterCue,
} from "@/app/(main)/_features/video-core/video-core"
import { vc_showOverlayFeedback } from "@/app/(main)/_features/video-core/video-core-overlay-display"
import { VIDEOCORE_PREVIEW_CAPTURE_INTERVAL_SECONDS, VIDEOCORE_PREVIEW_THUMBNAIL_SIZE } from "@/app/(main)/_features/video-core/video-core-preview"
import { vc_autoSkipOPEDAtom, vc_highlightOPEDChaptersAtom, vc_showChapterMarkersAtom } from "@/app/(main)/_features/video-core/video-core.atoms"
import { vc_formatTime, vc_getChapterType, vc_getOPEDChapters } from "@/app/(main)/_features/video-core/video-core.utils"
import { cn } from "@/components/ui/core/styling"
import { logger } from "@/lib/helpers/debug"
import { atom, useAtomValue } from "jotai"
import { useAtom, useSetAtom } from "jotai/react"
import Image from "next/image"
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

    const bufferedPercentage = React.useMemo(() => {
        return (buffered / duration) * 100
    }, [buffered])

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

    const opEdChapters = vc_getOPEDChapters(chapters)

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
        setSeeking(false)
        if (!isWatchPartyPeer) {
            // actually seek the video
            action({ type: "seekTo", payload: { time: (duration * seekingTargetProgress) / 100 } })
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
        if (seeking) {
            setSeeking(false)
            action({ type: "seekTo", payload: { time: (duration * seekingTargetProgress) / 100 } })
            if (!previouslyPaused) videoElement?.play()?.catch()
        }
    }

    // stop seeking
    function handlePointerLeave(e: React.PointerEvent<HTMLDivElement>) {
        // don't reset while actively seeking
        if (!seeking) {
            setSeekingTargetProgress(0)
        }
    }

    function getPointerProgress<T extends HTMLElement>(e: React.PointerEvent<T>) { // 0-100
        const rect = e.currentTarget.getBoundingClientRect()
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
            setProgressPercentage(target)
        }
        setSeekingTargetProgress(target)
    }

    const setTimeRangeElement = useSetAtom(vc_timeRangeElement)
    const combineRef = (instance: HTMLDivElement | null) => {
        // if (ref as unknown instanceof Function) (ref as any)(instance)
        // else if (ref) (ref as any).current = instance
        // if (instance) measureRef(instance)
        setTimeRangeElement(instance)
    }

    return (
        <div
            ref={combineRef}
            data-vc-element="time-range"
            data-vc-seeking-target-state={seeking}
            className={cn(
                "w-full relative group/vc-time-range z-[2] flex h-8",
                "cursor-pointer outline-none",
                "touch-none select-none", // prevent page scroll and text selection on mobile
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
                    "relative h-1 transition-[height] flex items-center justify-center overflow-hidden rounded-lg",
                    focused && "h-2",
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

function VideoCoreTimePreview(props: { chapters: VideoCoreTimeRangeChapter[] }) {
    const { chapters } = props

    const isMobile = useAtomValue(vc_isMobile)

    const duration = useAtomValue(vc_duration)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const seekingTargetProgress = useAtomValue(vc_seekingTargetProgress)
    const seeking = useAtomValue(vc_seeking)
    const isSwiping = useAtomValue(vc_isSwiping)
    const swipeSeekTime = useAtomValue(vc_swipeSeekTime)
    const action = useSetAtom(vc_dispatchAction)
    const previewManager = useAtomValue(vc_previewManager)
    const timeRangeElement = useAtomValue(vc_timeRangeElement)

    const [previewThumbnail, setPreviewThumbnail] = React.useState<string | null>(null)

    const targetTime = React.useMemo(() => {
        if (isSwiping && swipeSeekTime !== null) {
            return swipeSeekTime
        }
        return (duration * seekingTargetProgress) / 100
    }, [isSwiping, swipeSeekTime, duration, seekingTargetProgress])

    const chapterLabel = React.useMemo(() => {
        // returns chapter name at the current target
        const targetPercentage = isSwiping && swipeSeekTime !== null
            ? (swipeSeekTime / duration) * 100
            : seekingTargetProgress
        const chapter = chapters.find(chapter =>
            chapter.percentageOffset <= targetPercentage &&
            chapter.percentageOffset + chapter.width >= targetPercentage,
        )
        return chapter?.label
    }, [isSwiping, swipeSeekTime, duration, seekingTargetProgress, chapters])

    const handleTimeRangePreview = React.useCallback(async (event: Event) => {
        if (!previewManager || !duration || !timeRangeElement) {
            return
        }

        setPreviewThumbnail(null)
        timeRangeElement.removeAttribute("data-preview-image")

        // Calculate preview time based on mouse or touch position
        const rect = timeRangeElement.getBoundingClientRect()
        let clientX: number

        if (event instanceof TouchEvent && event.touches.length > 0) {
            clientX = event.touches[0].clientX
        } else if (event instanceof MouseEvent) {
            clientX = event.clientX
        } else {
            return
        }

        const x = clientX - rect.left
        const percentage = Math.max(0, Math.min(1, x / rect.width))
        const previewTime = percentage * duration

        if (previewTime >= 0 && previewTime <= duration) {
            const thumbnailIndex = Math.floor(previewTime / VIDEOCORE_PREVIEW_CAPTURE_INTERVAL_SECONDS)

            try {
                const thumbnail = await previewManager.retrievePreviewForSegment(thumbnailIndex)
                if (thumbnail) {
                    timeRangeElement.setAttribute("data-preview-image", thumbnail)
                    setPreviewThumbnail(thumbnail)
                }
            }
            catch (error) {
                timeRangeLog.error("Failed to get thumbnail", error)
            }
        }
    }, [previewManager, timeRangeElement, duration])

    React.useEffect(() => {
        if (!timeRangeElement) {
            return
        }

        const handleMouseLeave = () => {
            timeRangeElement.removeAttribute("data-preview-image")
            setPreviewThumbnail(null)
        }

        const handleTouchEnd = () => {
            timeRangeElement.removeAttribute("data-preview-image")
            setPreviewThumbnail(null)
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
    }, [handleTimeRangePreview, timeRangeElement, isMobile])

    // Fetch thumbnail preview during swipe
    React.useEffect(() => {
        if (!isSwiping || !swipeSeekTime || !previewManager || !duration) {
            return
        }

        const fetchSwipeThumbnail = async () => {
            const thumbnailIndex = Math.floor(swipeSeekTime / VIDEOCORE_PREVIEW_CAPTURE_INTERVAL_SECONDS)

            try {
                const thumbnail = await previewManager.retrievePreviewForSegment(thumbnailIndex)
                if (thumbnail) {
                    setPreviewThumbnail(thumbnail)
                }
            }
            catch (error) {
                timeRangeLog.error("Failed to get swipe thumbnail", error)
            }
        }

        fetchSwipeThumbnail()
    }, [isSwiping, swipeSeekTime, previewManager, duration])

    // Clear thumbnail when swipe ends
    React.useEffect(() => {
        if (!isSwiping) {
            setPreviewThumbnail(null)
        }
    }, [isSwiping])

    const showThumbnail = (!isMiniPlayer && previewManager && (seeking || isSwiping || !!targetTime)) &&
        targetTime <= previewManager.getLastestCachedIndex() * VIDEOCORE_PREVIEW_CAPTURE_INTERVAL_SECONDS

    return <>

        {showThumbnail && <div
            data-vc-element="preview-thumbnail"
            className={cn(
                "absolute bottom-full aspect-video overflow-hidden rounded-md bg-black border border-white/50 pointer-events-none",
            )}
            style={!isMobile ? {
                left: `clamp(${VIDEOCORE_PREVIEW_THUMBNAIL_SIZE / 2}px, ${(targetTime / duration) * 100}%, calc(100% - ${VIDEOCORE_PREVIEW_THUMBNAIL_SIZE / 2}px))`,
                width: VIDEOCORE_PREVIEW_THUMBNAIL_SIZE + "px",
                transform: "translateX(-50%) translateY(-54%)",
            } : {
                left: `clamp(${140 / 2}px, ${(targetTime / duration) * 100}%, calc(100% - ${140 / 2}px))`,
                width: 140 + "px",
                transform: "translateX(-50%) translateY(-64%)",
            }}
        >
            {!!previewThumbnail && <Image
                data-vc-element="preview-thumbnail-image"
                src={previewThumbnail || ""}
                alt="Preview"
                fill
                sizes={VIDEOCORE_PREVIEW_THUMBNAIL_SIZE + "px"}
                className="object-cover rounded-md"
                decoding="async"
                loading="lazy"
            />}
        </div>}

        {(seeking || isSwiping || !!targetTime) && <div
            data-vc-element="preview-tooltip"
            className={cn(
                "absolute bottom-full mb-3 px-2 py-1 bg-black/70 text-white text-center text-sm rounded-md",
                "whitespace-nowrap z-20 pointer-events-none",
                "",
            )}
            style={{
                left: `${(targetTime / duration) * 100}%`,
                transform: "translateX(-50%)",
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
