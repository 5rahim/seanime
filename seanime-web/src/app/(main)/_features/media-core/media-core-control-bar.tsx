import { SeaImage as Image } from "@/components/shared/sea-image"
import { cn } from "@/components/ui/core/styling"
import { AnimatePresence, motion } from "motion/react"
import React from "react"
import { FaDiamond } from "react-icons/fa6"
import { LuChevronLeft, LuChevronRight, LuVolume, LuVolume1, LuVolume2, LuVolumeOff } from "react-icons/lu"
import { RiPauseLargeLine, RiPlayLargeLine } from "react-icons/ri"
import { RxEnterFullScreen, RxExitFullScreen } from "react-icons/rx"
import { TbPictureInPicture, TbPictureInPictureOff } from "react-icons/tb"

export function formatTime(seconds: number) {
    const sign = seconds < 0 ? "-" : ""
    const absSeconds = Math.abs(seconds)
    const hours = Math.floor(absSeconds / 3600)
    const minutes = Math.floor((absSeconds % 3600) / 60)
    const secs = Math.floor(absSeconds % 60)

    if (hours > 0) {
        return `${sign}${hours}:${minutes.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`
    }
    return `${sign}${minutes}:${secs.toString().padStart(2, "0")}`
}

const CONTROL_BAR_MAIN_HEIGHT = 48
const CONTROL_BAR_MAIN_HEIGHT_MINI = 28
const CHAPTER_GAP = 3

export interface MediaCoreControlBarViewProps {
    paused: boolean
    isMiniPlayer: boolean
    cursorBusy: boolean
    hoveringControlBar: boolean
    onHoveringControlBarChange?: (hovering: boolean) => void
    containerElement: HTMLDivElement | null
    isMobile: boolean
    timeRange: React.ReactNode
    children?: React.ReactNode
}

export function MediaCoreControlBarView(props: MediaCoreControlBarViewProps) {
    const {
        paused,
        isMiniPlayer,
        cursorBusy,
        hoveringControlBar,
        onHoveringControlBarChange,
        containerElement,
        isMobile,
        timeRange,
        children,
    } = props

    const mainSectionHeight = isMiniPlayer ? CONTROL_BAR_MAIN_HEIGHT_MINI : CONTROL_BAR_MAIN_HEIGHT
    const [cursorPosition, setCursorPosition] = React.useState<"outside" | "approaching" | "hover">("outside")
    const cursorPositionRef = React.useRef(cursorPosition)
    const containerRectRef = React.useRef<DOMRect | null>(null)

    React.useEffect(() => {
        cursorPositionRef.current = cursorPosition
    }, [cursorPosition])

    const setCursorPositionC = React.useEffectEvent((nextPosition: "outside" | "approaching" | "hover") => {
        if (cursorPositionRef.current === nextPosition) return
        cursorPositionRef.current = nextPosition
        setCursorPosition(nextPosition)
    })

    const showOnlyTimeRange = isMobile ? false : (
        (!paused && cursorPosition === "approaching") ||
        (paused && cursorPosition === "outside") ||
        (paused && cursorPosition === "approaching")
    )

    const controlBarTranslateY = isMobile ? (
        (paused || cursorBusy || hoveringControlBar) ? 0 : 300
    ) : (
        (cursorBusy || hoveringControlBar) ? 0 : (
            showOnlyTimeRange ? mainSectionHeight : (
                cursorPosition === "hover" ? 0 : 300
            )
        )
    )

    const hideShadow = isMobile ? !paused : (
        isMiniPlayer ? !paused : (cursorPosition !== "hover" && !cursorBusy)
    )

    const hideControlBar = isMobile ? (!paused && !cursorBusy && !hoveringControlBar) : (
        !showOnlyTimeRange && !cursorBusy && !hoveringControlBar
    )

    React.useEffect(() => {
        if (isMobile || !containerElement) return

        const handlePointerMove = (e: PointerEvent) => {
            const rect = containerElement.getBoundingClientRect()
            containerRectRef.current = rect
            const y = e.clientY - rect.top
            const registerThreshold = !isMiniPlayer ? 150 : 100
            const showOnlyTimeRangeOffset = 50

            if (y >= rect.height - registerThreshold && y < rect.height - registerThreshold + showOnlyTimeRangeOffset) {
                setCursorPositionC("approaching")
            } else if (y < rect.height - registerThreshold) {
                setCursorPositionC("outside")
            } else {
                setCursorPositionC("hover")
            }
        }

        const handlePointerLeave = () => {
            containerRectRef.current = null
            setCursorPositionC("outside")
        }

        const updateContainerRect = () => {
            containerRectRef.current = containerElement.getBoundingClientRect()
        }

        updateContainerRect()
        const resizeObserver = new ResizeObserver(updateContainerRect)
        resizeObserver.observe(containerElement)

        containerElement.addEventListener("pointerenter", updateContainerRect)
        containerElement.addEventListener("pointermove", handlePointerMove)
        containerElement.addEventListener("pointerleave", handlePointerLeave)
        containerElement.addEventListener("pointercancel", handlePointerLeave)
        window.addEventListener("resize", updateContainerRect)

        return () => {
            resizeObserver.disconnect()
            containerElement.removeEventListener("pointerenter", updateContainerRect)
            containerElement.removeEventListener("pointermove", handlePointerMove)
            containerElement.removeEventListener("pointerleave", handlePointerLeave)
            containerElement.removeEventListener("pointercancel", handlePointerLeave)
            window.removeEventListener("resize", updateContainerRect)
            containerRectRef.current = null
        }
    }, [containerElement, isMobile, isMiniPlayer])

    React.useLayoutEffect(() => {
        if (!containerElement || isMobile) return
        const captionsOverlay = containerElement.querySelector("#video-core-captions-wrapper") as HTMLElement
        if (!captionsOverlay) return
        if (controlBarTranslateY === 0 || showOnlyTimeRange) {
            captionsOverlay.style.setProperty("--tw-translate-y", `-${showOnlyTimeRange ? 20 : 50}px`, "important")
        } else {
            captionsOverlay.style.setProperty("--tw-translate-y", "0%")
        }
        return () => {
            captionsOverlay.style.removeProperty("--tw-translate-y")
        }
    }, [controlBarTranslateY, containerElement, isMobile, showOnlyTimeRange])

    return (
        <>
            <div
                data-vc-element="control-bar-gradient-bottom"
                className={cn(
                    "pointer-events-none absolute bottom-0 left-0 right-0 w-full z-[5] h-28 transition-opacity duration-300 opacity-0",
                    "bg-gradient-to-t to-transparent",
                    !isMiniPlayer ? "from-black/40" : "from-black/80 via-black/40",
                    isMiniPlayer && "h-20",
                    !hideShadow && "opacity-100",
                )}
            />
            {!isMiniPlayer && (
                <div
                    data-vc-element="control-bar-time-range-bottom-gradient"
                    className={cn(
                        "pointer-events-none absolute bottom-0 left-0 right-0 w-full z-[5] h-14 transition-opacity duration-400 opacity-0",
                        "bg-gradient-to-t to-transparent",
                        !isMiniPlayer ? "from-black/40" : "from-black/60",
                        isMiniPlayer && "h-10",
                        (showOnlyTimeRange && paused && hideShadow) && "opacity-100",
                    )}
                />
            )}
            <div
                data-vc-element="control-bar"
                data-vc-state-visible={!hideControlBar}
                className={cn(
                    "absolute left-0 bottom-0 right-0 flex flex-col text-white",
                    "transition-[opacity,transform] duration-300 opacity-0",
                    "z-[10] h-28 transform-gpu will-change-[opacity,transform]",
                    !hideControlBar && "opacity-100",
                )}
                style={{
                    transform: `translateY(${controlBarTranslateY}px)`,
                }}
                onPointerEnter={() => {
                    if (!isMobile) onHoveringControlBarChange?.(true)
                }}
                onPointerLeave={() => {
                    if (!isMobile) onHoveringControlBarChange?.(false)
                }}
                onPointerCancel={() => {
                    if (!isMobile) onHoveringControlBarChange?.(false)
                }}
            >
                <div
                    data-vc-element="control-bar-wrapper"
                    className={cn(
                        "absolute bottom-0 w-full",
                        isMobile ? "px-2" : "px-4",
                    )}
                >
                    {timeRange}

                    <div
                        data-vc-element="control-bar-main-section"
                        className={cn(
                            "z-[100] relative transform-gpu duration-100 flex items-center",
                            isMobile ? "pb-1" : "pb-2",
                        )}
                        style={{
                            height: `${mainSectionHeight}px`,
                        } as React.CSSProperties}
                    >
                        {children}
                    </div>
                </div>
            </div>
        </>
    )
}

export interface MediaCoreMobileControlBarViewProps {
    paused: boolean
    isMiniPlayer: boolean
    cursorBusy: boolean
    seeking: boolean
    isSwiping: boolean
    timeRange: React.ReactNode
    topLeftSection?: React.ReactNode
    topRightSection?: React.ReactNode
    bottomLeftSection?: React.ReactNode
    bottomRightSection?: React.ReactNode
}

export function MediaCoreMobileControlBarView(props: MediaCoreMobileControlBarViewProps) {
    const {
        paused,
        isMiniPlayer,
        cursorBusy,
        seeking,
        isSwiping,
        timeRange,
        topLeftSection,
        topRightSection,
        bottomLeftSection,
        bottomRightSection,
    } = props

    const [isSwipingDebounced, setIsSwipingDebounced] = React.useState(false)
    const sieT = React.useRef<ReturnType<typeof setTimeout> | null>(null)
    React.useEffect(() => {
        if (isSwiping) {
            setIsSwipingDebounced(true)
        } else {
            sieT.current = setTimeout(() => {
                setIsSwipingDebounced(false)
            }, 200)
        }
        return () => {
            if (sieT.current) clearTimeout(sieT.current)
        }
    }, [isSwiping])

    const showShadow = paused || cursorBusy
    const bottomSectionTranslateY = (paused || cursorBusy) ? 0 : 300

    return (
        <>
            <div
                data-vc-element="mobile-control-bar-gradient-bottom"
                className={cn(
                    "vc-mobile-control-bar-bottom-gradient pointer-events-none absolute bottom-0 left-0 right-0 w-full z-[10] h-28 transition-opacity duration-300 opacity-0",
                    "bg-gradient-to-t to-transparent",
                    !isMiniPlayer ? "from-black/40" : "from-black/80 via-black/40",
                    "h-20",
                    (showShadow || isSwiping) && "opacity-100",
                )}
            />
            <div
                data-vc-element="mobile-control-bar-gradient-top"
                className={cn(
                    "vc-mobile-control-bar-top-gradient pointer-events-none absolute top-0 left-0 right-0 w-full z-[10] h-28 transition-opacity duration-300 opacity-0",
                    "bg-gradient-to-b to-transparent",
                    !isMiniPlayer ? "from-black/40" : "from-black/80 via-black/40",
                    "h-20",
                    (showShadow) && "opacity-100",
                )}
            />

            {/*Top Section*/}
            <div
                data-vc-element="mobile-control-bar-top-section"
                className={cn(
                    "vc-mobile-control-bar-top-section absolute transition-transform left-0 right-0 top-0 w-full z-[11] transform-gpu px-2 pt-3",
                )}
                style={{
                    transform: `translateY(${-bottomSectionTranslateY}px)`,
                }}
            >
                <div data-vc-element="mobile-control-bar-top-content" className="transform-gpu duration-100 flex items-center">
                    {topLeftSection}
                    <div className="flex flex-1" />
                    {topRightSection}
                </div>
            </div>

            {/*Bottom Section*/}
            <div
                data-vc-element="mobile-control-bar-bottom-section"
                className={cn(
                    "vc-mobile-control-bar-bottom-section absolute transition-transform left-0 right-0 bottom-0 w-full z-[11] transform-gpu px-2",
                    isSwiping && "transition-none",
                )}
                style={{
                    transform: isSwiping ? "translateY(0px)" : `translateY(${bottomSectionTranslateY}px)`,
                }}
            >
                <div
                    data-vc-element="mobile-control-bar-bottom-content"
                    className={cn(
                        "transform-gpu duration-100 flex items-center",
                        (isSwiping || isSwipingDebounced) && "hidden",
                    )}
                >
                    {bottomLeftSection}
                    <div className="flex flex-1" />
                    {bottomRightSection}
                </div>
                {timeRange}
            </div>
        </>
    )
}

export interface MediaCoreControlButtonIconProps {
    icons: [string, React.ElementType][]
    state: string
    className?: string
    iconClass?: string
    onClick: () => void
    onWheel?: (e: React.WheelEvent<HTMLButtonElement>) => void
    children?: React.ReactNode
    isMobile: boolean
    isMiniPlayer: boolean
}

export function MediaCoreControlButtonIcon(props: MediaCoreControlButtonIconProps) {
    const { icons, state, className, iconClass, onClick, onWheel, children, isMobile, isMiniPlayer } = props

    return (
        <button
            role="button"
            data-vc-element="control-button"
            data-vc-state={state}
            className={cn(
                "vc-control-button flex items-center justify-center transition-opacity relative h-full focus-visible:outline-none focus:outline-none focus-visible:opacity-50",
                isMobile ? "px-1 text-2xl" : "px-2 text-3xl hover:opacity-80",
                isMiniPlayer && !isMobile && "text-2xl",
                className,
            )}
            onClick={onClick}
            onWheel={onWheel}
        >
            <AnimatePresence>
                {icons.map(n => {
                    const [iconState, Icon] = n
                    if (state !== iconState) return null
                    return (
                        <motion.span
                            key={iconState}
                            data-vc-element="control-button-icon"
                            data-vc-state={iconState}
                            className="block"
                            initial={{ opacity: 0, y: 10, position: "relative" }}
                            animate={{ opacity: 1, y: 0, position: "relative" }}
                            exit={{ opacity: 0, y: 10, position: "absolute" }}
                            transition={{ duration: 0.15 }}
                        >
                            <Icon className={cn("vc-control-button-icon", iconClass)} />
                        </motion.span>
                    )
                })}
            </AnimatePresence>
            {children}
        </button>
    )
}

export function MediaCorePlayButton(props: {
    paused: boolean
    onTogglePlay: () => void
    isMobile: boolean
    isMiniPlayer: boolean
}) {
    return (
        <MediaCoreControlButtonIcon
            icons={[
                ["playing", RiPauseLargeLine],
                ["paused", RiPlayLargeLine],
            ]}
            state={props.paused ? "paused" : "playing"}
            onClick={props.onTogglePlay}
            isMobile={props.isMobile}
            isMiniPlayer={props.isMiniPlayer}
        />
    )
}

export interface MediaCoreVolumeButtonProps {
    volume: number
    muted: boolean
    onVolumeChange: (vol: number) => void
    onMuteToggle: () => void
    isMobile: boolean
    isMiniPlayer: boolean
}

export function MediaCoreVolumeButton(props: MediaCoreVolumeButtonProps) {
    const { volume, muted, onVolumeChange, onMuteToggle, isMobile, isMiniPlayer } = props
    const [isSliding, setIsSliding] = React.useState(false)

    function linearToVolume(linear: number): number {
        return Math.pow(linear, 2)
    }

    function volumeToLinear(vol: number): number {
        return Math.pow(vol, 1 / 2)
    }

    function handlePointerDown(e: React.PointerEvent<HTMLDivElement>) {
        e.stopPropagation()
        e.currentTarget.setPointerCapture(e.pointerId)
        setIsSliding(true)
    }

    function handleSetVolume(e: React.PointerEvent<HTMLDivElement>) {
        const rect = e.currentTarget.getBoundingClientRect()
        const x = e.clientX - rect.left
        const width = e.currentTarget.clientWidth
        const linearPosition = Math.max(0, Math.min(1, x / width))
        const nonLinearVolume = linearToVolume(linearPosition)
        onVolumeChange(nonLinearVolume)
    }

    function handlePointerUp(e: React.PointerEvent<HTMLDivElement>) {
        if (isSliding) {
            e.stopPropagation()
            e.currentTarget.releasePointerCapture(e.pointerId)
            setIsSliding(false)
            handleSetVolume(e)
        }
    }

    function handlePointerMove(e: React.PointerEvent<HTMLDivElement>) {
        if (isSliding) {
            e.stopPropagation()
            handleSetVolume(e)
        }
    }

    function handleWheel(e: React.WheelEvent<HTMLButtonElement | HTMLDivElement>) {
        e.stopPropagation()
        const delta = -e.deltaY / 1000
        const newVolume = Math.max(0, Math.min(1, volume + delta))
        onVolumeChange(newVolume)
    }

    return (
        <div data-vc-element="control-volume" className="vc-control-volume group/vc-control-volume flex items-center justify-center h-full gap-2 text-white">
            <MediaCoreControlButtonIcon
                icons={[
                    ["low", LuVolume],
                    ["mid", LuVolume1],
                    ["high", LuVolume2],
                    ["muted", LuVolumeOff],
                ]}
                state={
                    muted ? "muted" :
                        volume >= 0.5 ? "high" :
                            volume > 0.1 ? "mid" :
                                "low"
                }
                className={isMiniPlayer ? "text-[1.3rem]" : "text-2xl"}
                onClick={onMuteToggle}
                onWheel={handleWheel}
                isMobile={isMobile}
                isMiniPlayer={isMiniPlayer}
            />
            {!isMobile && (
                <div
                    data-vc-element="control-volume-slider-container"
                    className="vc-control-volume-slider-container relative w-0 flex group-hover/vc-control-volume:w-[6rem] h-6 transition-[width] duration-300"
                >
                    <div
                        data-vc-element="control-volume-slider"
                        className="vc-control-volume-slider flex h-full w-full relative items-center rounded-full cursor-pointer transition-[width,background-color] duration-300"
                        onPointerDown={handlePointerDown}
                        onPointerMove={handlePointerMove}
                        onPointerUp={handlePointerUp}
                        onPointerCancel={handlePointerUp}
                        onWheel={handleWheel}
                    >
                        <div
                            data-vc-element="control-volume-slider-progress"
                            className="vc-control-volume-slider-progress h-1.5 absolute bg-white rounded-full"
                            style={{
                                width: `${volumeToLinear(volume) * 100}%`,
                            }}
                        />
                        <div
                            data-vc-element="control-volume-slider-background"
                            className="vc-control-volume-slider-progress h-1.5 w-full absolute bg-white/20 rounded-full"
                        />
                    </div>
                    <div className="w-4" />
                </div>
            )}
        </div>
    )
}

export function MediaCoreNextButton(props: { onClick: () => void; isMobile: boolean; isMiniPlayer: boolean }) {
    if (props.isMiniPlayer) return null
    return (
        <MediaCoreControlButtonIcon
            icons={[["default", LuChevronRight]]}
            state="default"
            onClick={props.onClick}
            isMobile={props.isMobile}
            isMiniPlayer={props.isMiniPlayer}
        />
    )
}

export function MediaCorePreviousButton(props: { onClick: () => void; isMobile: boolean; isMiniPlayer: boolean }) {
    if (props.isMiniPlayer) return null
    return (
        <MediaCoreControlButtonIcon
            icons={[["default", LuChevronLeft]]}
            state="default"
            onClick={props.onClick}
            isMobile={props.isMobile}
            isMiniPlayer={props.isMiniPlayer}
        />
    )
}

export interface MediaCoreTimestampProps {
    currentTime: number
    duration: number
    timestampMode: "elapsed" | "remaining"
    onTimestampModeToggle: () => void
    isMobile: boolean
}

export function MediaCoreTimestamp(props: MediaCoreTimestampProps) {
    const { currentTime, duration, timestampMode, onTimestampModeToggle, isMobile } = props

    if (duration <= 1 || isNaN(duration)) return null

    const timeToShow = timestampMode === "remaining"
        ? Math.max(0, duration - currentTime)
        : currentTime

    return (
        <p
            data-vc-element="timestamp"
            data-vc-timestamp-type={timestampMode}
            className={cn(
                "tabular-nums font-medium opacity-100 cursor-pointer text-white",
                isMobile ? "text-xs" : "text-sm hover:opacity-80",
            )}
            onClick={onTimestampModeToggle}
        >
            {timestampMode === "remaining" ? "-" : ""}
            {formatTime(Math.max(0, Math.min(duration, timeToShow)))} / {formatTime(duration)}
        </p>
    )
}

export function MediaCorePipButton(props: {
    isPip: boolean
    onTogglePip: () => void
    isMobile: boolean
    isMiniPlayer: boolean
}) {
    if (props.isMiniPlayer) return null
    return (
        <MediaCoreControlButtonIcon
            icons={[
                ["default", TbPictureInPicture],
                ["pip", TbPictureInPictureOff],
            ]}
            state={props.isPip ? "pip" : "default"}
            onClick={props.onTogglePip}
            isMobile={props.isMobile}
            isMiniPlayer={props.isMiniPlayer}
        />
    )
}

export function MediaCoreFullscreenButton(props: {
    isFullscreen: boolean
    onToggleFullscreen: () => void
    isMobile: boolean
    isMiniPlayer: boolean
}) {
    return (
        <MediaCoreControlButtonIcon
            icons={[
                ["default", RxEnterFullScreen],
                ["fullscreen", RxExitFullScreen],
            ]}
            state={props.isFullscreen ? "fullscreen" : "default"}
            onClick={props.onToggleFullscreen}
            isMobile={props.isMobile}
            isMiniPlayer={props.isMiniPlayer}
        />
    )
}

/* -------------------------------------------------------------------------------------------------
 * TimeRange / Timeline View
 * -----------------------------------------------------------------------------------------------*/

type MediaCoreTimeRangeChapter = {
    width: number
    percentageOffset: number
    label: string | null
    start: number
    end: number
}

export interface MediaCoreTimeRangeViewProps {
    seeking: boolean
    progressPercentage: number
    bufferedPercentage: number
    seekingTargetProgress: number
    chapters: MediaCoreTimeRangeChapter[]
    showChapterMarkers: boolean
    highlightChapters?: boolean
    skipChapters?: MediaCoreTimeRangeChapter[]
    showPreview: boolean
    showThumbnail: boolean
    previewWidth: number
    previewX: number
    tooltipX: number
    chapterLabel: string | null
    targetTime: number
    previewThumbnailUrl: string | null
    timeRangeRef: React.Ref<HTMLDivElement>
    onPointerDown: (e: React.PointerEvent<HTMLDivElement>) => void
    onPointerUp: (e: React.PointerEvent<HTMLDivElement>) => void
    onPointerMove: (e: React.PointerEvent<HTMLDivElement>) => void
    onPointerLeave: (e: React.PointerEvent<HTMLDivElement>) => void
    onPointerCancel: (e: React.PointerEvent<HTMLDivElement>) => void
    onMarkerClick?: (time: number) => void
    isMobile: boolean
    duration: number
}


export function MediaCoreTimeRangeView(props: MediaCoreTimeRangeViewProps) {
    const {
        seeking,
        progressPercentage,
        bufferedPercentage,
        seekingTargetProgress,
        chapters,
        showChapterMarkers,
        highlightChapters = true,
        skipChapters = [],
        showPreview,
        showThumbnail,
        previewWidth,
        previewX,
        tooltipX,
        chapterLabel,
        targetTime,
        previewThumbnailUrl,
        timeRangeRef,
        onPointerDown,
        onPointerUp,
        onPointerMove,
        onPointerLeave,
        onPointerCancel,
        onMarkerClick,
        isMobile,
        duration,
    } = props

    return (
        <div
            ref={timeRangeRef}
            data-vc-element="time-range"
            data-vc-seeking-target-state={seeking}
            className="w-full relative group/vc-time-range z-[2] flex h-8 cursor-pointer outline-none touch-none select-none [contain:layout_style] text-white"
            role="slider"
            tabIndex={0}
            aria-valuemin={0}
            aria-valuenow={progressPercentage}
            aria-valuetext={`${Math.round(progressPercentage)}%`}
            aria-orientation="horizontal"
            aria-label="Video playback time"
            onPointerDown={onPointerDown}
            onPointerUp={onPointerUp}
            onPointerLeave={onPointerLeave}
            onPointerCancel={onPointerCancel}
            onPointerMove={onPointerMove}
        >
            {showThumbnail && (
                <div
                    data-vc-element="preview-thumbnail"
                    className="absolute left-0 bottom-full aspect-video overflow-hidden rounded-md bg-black border border-white/50 pointer-events-none will-change-transform [contain:layout_paint_style]"
                    style={{
                        width: previewWidth,
                        transform: `translate3d(${previewX - previewWidth / 2}px, ${isMobile ? "-64%" : "-54%"}, 0)`,
                    }}
                >
                    {!!previewThumbnailUrl && (
                        <Image
                            data-vc-element="preview-thumbnail-image"
                            src={previewThumbnailUrl}
                            alt="Preview"
                            fill
                            sizes={`${previewWidth}px`}
                            className="object-cover rounded-md"
                            decoding="async"
                            loading="lazy"
                        />
                    )}
                </div>
            )}

            {showPreview && (
                <div
                    data-vc-element="preview-tooltip"
                    className="absolute left-0 bottom-full mb-3 px-2 py-1 bg-black/70 text-white text-center text-sm rounded-md will-change-transform whitespace-nowrap z-20 pointer-events-none"
                    style={{
                        transform: `translate3d(${tooltipX}px, 0, 0) translateX(-50%)`,
                    }}
                >
                    {chapterLabel && <p data-vc-element="preview-tooltip-chapter" className="text-xs font-medium max-w-2xl truncate">{chapterLabel}</p>}
                    <p data-vc-element="preview-tooltip-time">{formatTime(targetTime)}</p>
                    <div
                        data-vc-element="preview-tooltip-arrow"
                        className="absolute top-full left-1/2 -translate-x-1/2 w-0 h-0 border-l-4 border-r-4 border-t-4 border-transparent border-t-black/70"
                    />
                </div>
            )}

            {chapters.map((chapter, i) => (
                <MediaCoreTimeRangeSegment
                    key={i}
                    idx={i}
                    progressPercentage={progressPercentage}
                    bufferedPercentage={bufferedPercentage}
                    chapter={chapter}
                    showMarker={i < chapters.length - 1 && showChapterMarkers && !isMobile}
                    duration={duration}
                    seekingTargetProgress={seekingTargetProgress}
                    highlighted={highlightChapters && skipChapters.includes(chapter)}
                    onMarkerClick={onMarkerClick}
                />
            ))}
        </div>
    )
}

function MediaCoreTimeRangeSegment(props: {
    idx: number
    progressPercentage: number
    bufferedPercentage: number
    chapter: {
        width: number
        percentageOffset: number
        label: string | null
        start: number
        end: number
    }
    showMarker: boolean
    duration: number
    seekingTargetProgress: number
    highlighted: boolean
    onMarkerClick?: (time: number) => void
}) {
    const { idx, chapter, progressPercentage, bufferedPercentage, showMarker, duration, seekingTargetProgress, highlighted, onMarkerClick } = props
    const focused = !!seekingTargetProgress && chapter.percentageOffset <= seekingTargetProgress && chapter.percentageOffset + chapter.width >= seekingTargetProgress

    function getChapterBarPosition(chapter: any, percentage: number) {
        const ret = (percentage - chapter.percentageOffset) * (100 / chapter.width)
        return (ret <= 0 ? -CHAPTER_GAP : ret >= 100 ? 100 : ret) - 100
    }

    return (
        <div
            data-vc-element="time-range-segment"
            data-vc-chapter-label={chapter.label}
            data-vc-focused-state={focused}
            data-vc-seeking-target-state={!!seekingTargetProgress}
            className="relative w-full h-full flex items-center"
            style={{
                width: `${chapter.width}%`,
            }}
        >
            <div
                data-vc-element="time-range-chapter"
                className={cn(
                    "relative h-1 transition-transform transform-gpu origin-center will-change-transform flex items-center justify-center overflow-hidden rounded-lg",
                    focused && "scale-y-[2]",
                )}
                style={{
                    width: idx > 0 ? `calc(100% - ${CHAPTER_GAP / 2}px)` : `100%`,
                    marginLeft: idx > 0 ? `${CHAPTER_GAP}px` : `0px`,
                }}
            >
                <div
                    data-vc-element="time-range-chapter-bar"
                    data-vc-for="progress"
                    className="bg-white absolute w-full h-full left-0 transform-gpu z-[10]"
                    style={{
                        "--tw-translate-x": duration > 1 ? `${getChapterBarPosition(chapter, progressPercentage)}%` : "-100%",
                    } as React.CSSProperties}
                />
                <div
                    data-vc-element="time-range-chapter-bar"
                    data-vc-for="seeking-target"
                    className="bg-white/30 absolute w-full h-full left-0 transform-gpu z-[9]"
                    style={{
                        "--tw-translate-x": duration > 1 ? `${getChapterBarPosition(chapter, seekingTargetProgress)}%` : "-100%",
                    } as React.CSSProperties}
                />
                <div
                    data-vc-element="time-range-chapter-bar"
                    data-vc-for="buffer"
                    className="bg-white/10 absolute w-full h-full left-0 transform-gpu z-[8]"
                    style={{
                        "--tw-translate-x": duration > 1 ? `${getChapterBarPosition(chapter, bufferedPercentage)}%` : "-100%",
                    } as React.CSSProperties}
                />
                <div
                    data-vc-element="time-range-chapter-bar"
                    data-vc-for="main"
                    data-vc-highlighted-state={highlighted}
                    className={cn(
                        "bg-white/20 absolute left-0 w-full h-full z-[1]",
                        highlighted && "bg-blue-300/50",
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
                        onMarkerClick?.(chapter.start + (chapter.end - chapter.start))
                    }}
                    className="absolute top-0 right-0 size-4 flex items-center justify-center -translate-y-1/2 translate-x-1/2 cursor-pointer z-[20]"
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
