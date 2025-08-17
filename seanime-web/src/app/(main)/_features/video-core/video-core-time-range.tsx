import {
    vc_closestBufferedTime,
    vc_currentTime,
    vc_dispatchAction,
    vc_duration,
    vc_miniPlayer,
    vc_previewManager,
    vc_previousPausedState,
    vc_seeking,
    vc_seekingTargetProgress,
    vc_videoElement,
    VIDEOCORE_DEBUG_ELEMENTS,
    VideoCoreChapterCue,
} from "@/app/(main)/_features/video-core/video-core"
import { VIDEOCORE_PREVIEW_CAPTURE_INTERVAL_SECONDS, VIDEOCORE_PREVIEW_THUMBNAIL_SIZE } from "@/app/(main)/_features/video-core/video-core-preview"
import { vc_highlightOPEDChaptersAtom, vc_showChapterMarkersAtom } from "@/app/(main)/_features/video-core/video-core.atoms"
import { vc_formatTime } from "@/app/(main)/_features/video-core/video-core.utils"
import { cn } from "@/components/ui/core/styling"
import { logger } from "@/lib/helpers/debug"
import { atom } from "jotai"
import { useAtomValue } from "jotai/index"
import { useAtom, useSetAtom } from "jotai/react"
import Image from "next/image"
import React from "react"
import { FaDiamond } from "react-icons/fa6"

type VideoCoreTimeRangeChapter = {
    width: number
    percentageOffset: number
    label: string | null
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

    const videoElement = useAtomValue(vc_videoElement)

    const currentTime = useAtomValue(vc_currentTime)
    const duration = useAtomValue(vc_duration)
    const buffered = useAtomValue(vc_closestBufferedTime)
    const [seekingTargetProgress, setSeekingTargetProgress] = useAtom(vc_seekingTargetProgress)
    const [seeking, setSeeking] = useAtom(vc_seeking)
    const [previouslyPaused, setPreviouslyPaused] = useAtom(vc_previousPausedState)
    const action = useSetAtom(vc_dispatchAction)
    const [showChapterMarkers] = useAtom(vc_showChapterMarkersAtom)

    const bufferedPercentage = React.useMemo(() => {
        return (buffered / duration) * 100
    }, [buffered])

    const chapters = React.useMemo(() => {
        if (!chapterCues?.length) return [{
            width: 100,
            percentageOffset: 0,
            label: null,
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
                }
                percentageOffset += width
                return result
            })
    }, [chapterCues, duration])

    const [progressPercentage, setProgressPercentage] = React.useState((currentTime / duration) * 100)

    React.useEffect(() => {
        setProgressPercentage((currentTime / duration) * 100)
    }, [currentTime, duration])


    // start seeking
    function handlePointerDown(e: React.PointerEvent<HTMLDivElement>) {
        e.stopPropagation()
        if (!videoElement) return
        if (e.button !== 0) return
        e.currentTarget.setPointerCapture(e.pointerId) // capture movement outside
        setSeeking(true)
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
        if (e.button !== 0) return
        e.currentTarget.releasePointerCapture(e.pointerId)
        setSeeking(false)
        // actually seek the video
        action({ type: "seekTo", payload: { time: (duration * seekingTargetProgress) / 100 } })
        // resume playing
        if (!previouslyPaused) videoElement?.play()
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
            className={cn(
                "vc-time-range",
                "w-full relative group/vc-time-range z-[2] flex h-8",
                "cursor-pointer outline-none",
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
            onPointerCancel={handlePointerLeave}
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
                    showMarker={i < chapters.length - 1 && showChapterMarkers}
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
            className={cn(
                "vc-time-range-chapter-segment",
                "relative",
                "w-full h-full flex items-center",
                VIDEOCORE_DEBUG_ELEMENTS && "bg-yellow-500/10",
            )}
            style={{
                width: `${chapter.width}%`,
            }}
        >
            <div
                className={cn(
                    "vc-time-range-chapter",
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
                    className={cn(
                        "vc-time-range-chapter-progress-bar",
                        "bg-white absolute w-full h-full left-0 transform-gpu hover:duration-[30ms] z-[10]",
                        focused && "duration-[30ms]",
                    )}
                    style={{
                        "--tw-translate-x": duration > 1 ? `${getChapterBarPosition(chapter, progressPercentage)}%` : "-100%",
                    } as React.CSSProperties}
                />
                <div
                    className={cn(
                        "vc-time-range-chapter-seeking-target-bar",
                        "bg-white/30 absolute w-full h-full left-0 transform-gpu z-[9]",
                    )}
                    style={{
                        "--tw-translate-x": duration > 1 ? `${getChapterBarPosition(chapter, seekingTargetProgress)}%` : "-100%",
                    } as React.CSSProperties}
                />
                <div
                    className={cn(
                        "vc-time-range-chapter-buffer-bar",
                        "bg-white/10 absolute w-full h-full left-0 transform-gpu z-[8]",
                    )}
                    style={{
                        "--tw-translate-x": duration > 1 ? `${getChapterBarPosition(chapter, bufferedPercentage)}%` : "-100%",
                    } as React.CSSProperties}
                />
                <div
                    className={cn(
                        "vc-time-range-chapter-bar",
                        "bg-white/20 absolute left-0 w-full h-full z-[1]",
                        (["opening", "op", "ending",
                            "ed"].includes(chapter.label?.toLowerCase?.() || "") && highlightOPEDChapters) && "bg-blue-300/50",
                    )}
                />
            </div>
            {showMarker && (
                <button
                    type="button"
                    onPointerDown={e => e.stopPropagation()}
                    onPointerUp={e => e.stopPropagation()}
                    onClick={e => {
                        e.stopPropagation()
                        action({ type: "seekTo", payload: { time: ((duration * (chapter.percentageOffset + chapter.width))) / 100 } })
                    }}
                    className={cn(
                        "vc-time-range-chapter-marker",
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

const timePreviewLog = logger("VIDEO CORE / TIME PREVIEW")

function VideoCoreTimePreview(props: { chapters: VideoCoreTimeRangeChapter[] }) {
    const { chapters } = props

    const videoElement = useAtomValue(vc_videoElement)

    const duration = useAtomValue(vc_duration)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const seekingTargetProgress = useAtomValue(vc_seekingTargetProgress)
    const seeking = useAtomValue(vc_seeking)
    const action = useSetAtom(vc_dispatchAction)
    const previewManager = useAtomValue(vc_previewManager)
    const timeRangeElement = useAtomValue(vc_timeRangeElement)

    const [previewThumbnail, setPreviewThumbnail] = React.useState<string | null>(null)

    const targetTime = (duration * seekingTargetProgress) / 100 // in seconds

    const chapterLabel = React.useMemo(() => {
        // returns chapter name at the current target
        const chapter = chapters.find(chapter => chapter.percentageOffset <= seekingTargetProgress && chapter.percentageOffset + chapter.width >= seekingTargetProgress)
        return chapter?.label
    }, [seekingTargetProgress, chapters])

    const handleTimeRangePreview = React.useCallback(async (event: MouseEvent) => {
        if (!previewManager || !duration || !timeRangeElement) {
            return
        }

        setPreviewThumbnail(null)
        timeRangeElement.removeAttribute("data-preview-image")

        // Calculate preview time based on mouse position
        const rect = timeRangeElement.getBoundingClientRect()
        const x = event.clientX - rect.left
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
                timePreviewLog.error("Failed to get thumbnail", error)
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

        timeRangeElement.addEventListener("mouseleave", handleMouseLeave)
        timeRangeElement.addEventListener("mousemove", handleTimeRangePreview)

        return () => {
            timeRangeElement.removeEventListener("mouseleave", handleMouseLeave)
            timeRangeElement.removeEventListener("mousemove", handleTimeRangePreview)
        }
    }, [handleTimeRangePreview, timeRangeElement])

    const showThumbnail = (!isMiniPlayer && previewManager && (seeking || !!targetTime)) &&
        targetTime <= previewManager.getLastestCachedIndex() * VIDEOCORE_PREVIEW_CAPTURE_INTERVAL_SECONDS

    return <>

        {showThumbnail && <div
            className={cn(
                "absolute bottom-full aspect-video overflow-hidden rounded-md bg-black border border-white/50",
            )}
            style={{
                left: `clamp(${VIDEOCORE_PREVIEW_THUMBNAIL_SIZE / 2}px, ${(targetTime / duration) * 100}%, calc(100% - ${VIDEOCORE_PREVIEW_THUMBNAIL_SIZE / 2}px))`,
                width: VIDEOCORE_PREVIEW_THUMBNAIL_SIZE + "px",
                transform: "translateX(-50%) translateY(-54%)",
            }}
        >
            {!!previewThumbnail && <Image
                src={previewThumbnail || ""}
                alt="Preview"
                fill
                sizes={VIDEOCORE_PREVIEW_THUMBNAIL_SIZE + "px"}
                className="object-cover rounded-md"
                decoding="async"
                loading="lazy"
            />}
        </div>}

        {(seeking || !!targetTime) && <div
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
            {chapterLabel && <p className="text-xs font-medium max-w-2xl truncate">{chapterLabel}</p>}
            <p>{vc_formatTime(targetTime)}</p>

            <div
                className="absolute top-full left-1/2 -translate-x-1/2 w-0 h-0 border-l-4 border-r-4 border-t-4 border-transparent border-t-black/70"
            />
        </div>}
    </>
}
