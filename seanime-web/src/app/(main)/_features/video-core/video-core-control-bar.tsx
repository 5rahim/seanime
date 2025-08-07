import {
    vc_containerElement,
    vc_currentTime,
    vc_cursorBusy,
    vc_doAction,
    vc_duration,
    vc_miniPlayer,
    vc_paused,
    vc_previousPausedState,
    vc_seeking,
    vc_seekingTargetProgress,
    vc_videoElement,
    VIDEOCORE_DEBUG_ELEMENTS,
    VideoCoreChapterCue,
} from "@/app/(main)/_features/video-core/video-core"
import { cn } from "@/components/ui/core/styling"
import { atom, useAtomValue } from "jotai"
import { useAtom, useSetAtom } from "jotai/react"
import React from "react"

const VIDEOCORE_CONTROL_BAR_MAIN_SECTION_HEIGHT = 48
const VIDEOCORE_CONTROL_BAR_MAIN_SECTION_HEIGHT_MINI = 28

export const vc_hoveringControlBar = atom(false)

// VideoControlBar sits on the bottom of the video container
// shows up when cursor hovers bottom of the player or video is paused
export function VideoCoreControlBar(props: {
    children?: React.ReactNode
    timeRange: React.ReactNode
}) {
    const { children, timeRange } = props

    const paused = useAtomValue(vc_paused)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const [cursorBusy, setCursorBusy] = useAtom(vc_cursorBusy)
    const [hoveringControlBar, setHoveringControlBar] = useAtom(vc_hoveringControlBar)
    const [bottom, setBottom] = React.useState(-300)
    const seeking = useAtomValue(vc_seeking)

    const [showOnlyTimeRange, setShowOnlyTimeRange] = React.useState(false)

    // gradually show the control bar as cursor moves down
    // display it completely after a certain threshold or when the video is paused
    const containerElement = useAtomValue(vc_containerElement)

    const mainSectionHeight = isMiniPlayer ? VIDEOCORE_CONTROL_BAR_MAIN_SECTION_HEIGHT_MINI : VIDEOCORE_CONTROL_BAR_MAIN_SECTION_HEIGHT

    function handleVideoContainerPointerMove(e: Event) {
        if (!containerElement) return

        if (seeking || paused || hoveringControlBar) {
            setBottom(0)
            setShowOnlyTimeRange(false)
            return
        }

        const rect = containerElement.getBoundingClientRect()
        const y = e instanceof PointerEvent ? e.clientY - rect.top : 0
        const registerThreshold = !isMiniPlayer ? 150 : 100 // pixels from the bottom to start registering position
        const showOnlyTimeRangeOffset = !isMiniPlayer ? 50 : 50

        console.log(y >= rect.height - registerThreshold, y < rect.height - registerThreshold + showOnlyTimeRangeOffset)

        if ((y >= rect.height - registerThreshold && y < rect.height - registerThreshold + showOnlyTimeRangeOffset)) {
            setShowOnlyTimeRange(true)
            setBottom(0)
        } else if (y < rect.height - registerThreshold && !paused) {
            setBottom(-100)
            setShowOnlyTimeRange(false)
        } else {
            setBottom(0)
            setShowOnlyTimeRange(false)
        }
    }

    React.useEffect(() => {
        if (!containerElement) return
        containerElement.addEventListener("pointermove", handleVideoContainerPointerMove)
        return () => {
            containerElement.removeEventListener("pointermove", handleVideoContainerPointerMove)
        }
    }, [containerElement, paused, isMiniPlayer, seeking, hoveringControlBar])


    return (
        <div
            data-vc-control-bar-section
            className={cn(
                "absolute left-0 bottom-0 right-0 flex flex-col",
                "transition-all duration-300 opacity-0",
                "z-[100] h-28",
                (cursorBusy || paused || showOnlyTimeRange) && "opacity-100",
                VIDEOCORE_DEBUG_ELEMENTS && "bg-purple-500/20",
            )}
            style={{
                bottom: showOnlyTimeRange ? `-${mainSectionHeight}px` : bottom,
            }}
            onPointerEnter={() => {
                setCursorBusy(true)
                setHoveringControlBar(true)
            }}
            onPointerLeave={() => {
                setCursorBusy(false)
                setHoveringControlBar(false)
            }}
            onPointerCancel={() => {
                setCursorBusy(false)
                setHoveringControlBar(false)
            }}
        >
            <div
                data-vc-control-bar-bottom-gradient
                className={cn(
                    "absolute bottom-0 left-0 right-0 w-full z-[1] h-28 transition-opacity duration-100",
                    "bg-gradient-to-t from-black/100 to-transparent",
                    !isMiniPlayer ? "via-black/40" : "via-black/40",
                    isMiniPlayer && "h-20",
                    (showOnlyTimeRange || bottom != 0) && "opacity-0",
                )}
            />
            <div
                data-vc-control-bar
                className={cn(
                    "absolute bottom-0 w-full px-4",
                    VIDEOCORE_DEBUG_ELEMENTS && "bg-purple-800/40",
                )}
            >
                {timeRange}

                <div
                    className={cn(
                        "transition-gpu duration-100",
                    )}
                    style={{
                        height: `${mainSectionHeight}px`,
                        "--tw-translate-y": showOnlyTimeRange ? `-${mainSectionHeight}px` : 0,
                    } as React.CSSProperties}
                >
                    {children}
                </div>
            </div>
        </div>
    )
}

type VideoCoreTimeRangeChapter = {
    width: number
    percentageOffset: number
    label: string | null
}

export interface VideoCoreTimeRangeProps {
    chapterCues: VideoCoreChapterCue[]
}

const CHAPTER_GAP = 3

export function VideoCoreTimeRange(props: VideoCoreTimeRangeProps) {
    const {
        chapterCues,
    } = props

    const videoElement = useAtomValue(vc_videoElement)

    const currentTime = useAtomValue(vc_currentTime)
    const duration = useAtomValue(vc_duration)
    const [seekingTargetProgress, setSeekingTargetProgress] = useAtom(vc_seekingTargetProgress)
    const [seeking, setSeeking] = useAtom(vc_seeking)
    const [previouslyPaused, setPreviouslyPaused] = useAtom(vc_previousPausedState)
    const action = useSetAtom(vc_doAction)

    const chapters = React.useMemo(() => {
        if (!chapterCues?.length) return [{
            width: 100,
            percentageOffset: 0,
            label: null,
        }]

        let ret = [] as VideoCoreTimeRangeChapter[]

        let percentageOffset = 0
        for (const chapter of chapterCues.toSorted((a, b) => a.startTime - b.startTime)) {
            if (!chapter.startTime && !chapter.endTime) continue
            if (chapter.endTime > 0 && chapter.endTime < chapter.startTime) continue

            const start = chapter.startTime ?? 0
            const end = chapter.endTime ?? duration
            const chapterDuration = end - start
            const width = (chapterDuration / duration) * 100
            ret.push({
                width,
                percentageOffset: percentageOffset,
                label: chapter.text || null,
            })
            percentageOffset += width
        }

        return ret
    }, [chapterCues, duration])

    const [progressPercentage, setProgressPercentage] = React.useState((currentTime / duration) * 100)

    React.useEffect(() => {
        setProgressPercentage((currentTime / duration) * 100)
    }, [currentTime, duration])


    // start seeking
    function handlePointerDown(e: React.PointerEvent<HTMLDivElement>) {
        if (!videoElement) return
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
        e.currentTarget.releasePointerCapture(e.pointerId)
        setSeeking(false)
        // actually seek the video
        action({ type: "seekTo", payload: { time: (duration * seekingTargetProgress) / 100 } })
        // const newTime = (duration * seekingTargetProgress) / 100
        // if (videoElement) {
        //     videoElement.currentTime = newTime
        // }
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
        return clampPercentage(x / rect.width * 100)
    }

    function clampPercentage(n: number) {
        return Math.max(0, Math.min(100, n))
    }

    // move progress
    function handlePointerMove(e: React.PointerEvent<HTMLDivElement>) {
        const target = getPointerProgress(e)
        if (seeking) {
            setProgressPercentage(target)
        }
        setSeekingTargetProgress(target)
    }

    function getChapterBarPosition(chapter: VideoCoreTimeRangeChapter, percentage: number) {
        const ret = (100 / chapter.width) * (percentage - chapter.percentageOffset)
        return (ret <= 0 ? -CHAPTER_GAP : ret >= 100 ? 100 : ret) - 100
    }

    return (
        <div
            className={cn(
                "w-full relative group/vc-time-range z-[2] flex h-8",
                "cursor-pointer",
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

            {chapters.map((chapter, i) => {
                const focused = !!seekingTargetProgress && chapter.percentageOffset <= seekingTargetProgress && chapter.percentageOffset + chapter.width >= seekingTargetProgress
                return (
                    <div
                        key={i}
                        className={cn(
                            "vc-time-range-chapter",
                            "w-full h-full flex items-center",
                            VIDEOCORE_DEBUG_ELEMENTS && "bg-yellow-500/10",
                        )}
                        style={{
                            width: `${chapter.width}%`,
                        }}
                    >
                        <div
                            className={cn(
                                "vc-time-range-chapter-bg",
                                "relative h-1 transition-[height] flex items-center justify-center overflow-clip rounded-lg",
                                focused && "h-2",
                                VIDEOCORE_DEBUG_ELEMENTS && "bg-yellow-500/50",
                            )}
                            style={{
                                width: i > 0 ? `calc(100% - ${CHAPTER_GAP / 2}px)` : `100%`,
                                marginLeft: i > 0 ? `${CHAPTER_GAP}px` : `0px`,
                            }}
                        >
                            <div
                                className={cn(
                                    "bg-white absolute w-full h-full left-0 transform-gpu hover:duration-[1000ms] z-[10]",
                                    focused && "duration-75",
                                )}
                                style={{
                                    "--tw-translate-x": duration > 1 ? `${getChapterBarPosition(chapter, progressPercentage)}%` : "-100%",
                                    // "--tw-translate-x": `${getChapterBarPosition(chapter, progressPercentage)}%`,
                                } as React.CSSProperties}
                            />
                            <div
                                className={cn(
                                    "bg-white/20 absolute left-0 w-full h-full z-[1]",
                                )}
                            />
                        </div>
                    </div>
                )
            })}

        </div>
    )
}
