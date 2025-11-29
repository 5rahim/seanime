import {
    vc_containerElement,
    vc_currentTime,
    vc_cursorBusy,
    vc_dispatchAction,
    vc_duration,
    vc_isFullscreen,
    vc_isMuted,
    vc_miniPlayer,
    vc_paused,
    vc_pip,
    vc_seeking,
    vc_volume,
    VIDEOCORE_DEBUG_ELEMENTS,
} from "@/app/(main)/_features/video-core/video-core"
import { vc_fullscreenManager } from "@/app/(main)/_features/video-core/video-core-fullscreen"
import { vc_pipManager } from "@/app/(main)/_features/video-core/video-core-pip"
import { vc_storedMutedAtom, vc_storedVolumeAtom } from "@/app/(main)/_features/video-core/video-core.atoms"
import { vc_formatTime } from "@/app/(main)/_features/video-core/video-core.utils"
import { cn } from "@/components/ui/core/styling"
import { atom, useAtomValue } from "jotai"
import { useAtom, useSetAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import { AnimatePresence, motion } from "motion/react"
import React from "react"
import { LuChevronLeft, LuChevronRight, LuVolume, LuVolume1, LuVolume2, LuVolumeOff } from "react-icons/lu"
import { RiPauseLargeLine, RiPlayLargeLine } from "react-icons/ri"
import { RxEnterFullScreen, RxExitFullScreen } from "react-icons/rx"
import { TbPictureInPicture, TbPictureInPictureOff } from "react-icons/tb"

const VIDEOCORE_CONTROL_BAR_VPADDING = 5
const VIDEOCORE_CONTROL_BAR_MAIN_SECTION_HEIGHT = 48
const VIDEOCORE_CONTROL_BAR_MAIN_SECTION_HEIGHT_MINI = 28

export const vc_hoveringControlBar = atom(false)

type VideoCoreControlBarType = "default" | "classic"
const VIDEOCORE_CONTROL_BAR_TYPE: VideoCoreControlBarType = "default"

// VideoControlBar sits on the bottom of the video container
// shows up when cursor hovers bottom of the player or video is paused
export function VideoCoreControlBar(props: {
    children?: React.ReactNode
    timeRange: React.ReactNode
}) {
    const { children, timeRange } = props

    const paused = useAtomValue(vc_paused)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const cursorBusy = useAtomValue(vc_cursorBusy)
    const [hoveringControlBar, setHoveringControlBar] = useAtom(vc_hoveringControlBar)
    const containerElement = useAtomValue(vc_containerElement)
    const seeking = useAtomValue(vc_seeking)

    const mainSectionHeight = isMiniPlayer ? VIDEOCORE_CONTROL_BAR_MAIN_SECTION_HEIGHT_MINI : VIDEOCORE_CONTROL_BAR_MAIN_SECTION_HEIGHT

    // when the user is approaching the control bar
    const [cursorPosition, setCursorPosition] = React.useState<"outside" | "approaching" | "hover">("outside")

    const showOnlyTimeRange =
        VIDEOCORE_CONTROL_BAR_TYPE === "classic" ? (
                (!paused && cursorPosition === "approaching")
            ) :
            // cursor is approaching and video is not paused
            (!paused && cursorPosition === "approaching")
            // or cursor not hovering and video is paused
            || (paused && cursorPosition === "outside") || (paused && cursorPosition === "approaching")

    const controlBarBottomPx = VIDEOCORE_CONTROL_BAR_TYPE === "classic" ? (cursorBusy || hoveringControlBar || paused) ? 0 : (
        showOnlyTimeRange ? -(mainSectionHeight) : (
            cursorPosition === "hover" ? 0 : -300
        )
    ) : (
        (cursorBusy || hoveringControlBar) ? 0 : (
            showOnlyTimeRange ? -(mainSectionHeight) : (
                cursorPosition === "hover" ? 0 : -300
            )
        )
    )

    const hideShadow = isMiniPlayer ? !paused : VIDEOCORE_CONTROL_BAR_TYPE === "classic"
        ? (!paused && cursorPosition !== "hover" && !cursorBusy)
        : (cursorPosition !== "hover" && !cursorBusy)

    // const hideControlBar = !showOnlyTimeRange && !cursorBusy && !hoveringControlBar
    const hideControlBar = !showOnlyTimeRange && !cursorBusy && !hoveringControlBar && (VIDEOCORE_CONTROL_BAR_TYPE === "classic" ? !paused : true)

    function handleVideoContainerPointerMove(e: Event) {
        if (!containerElement) {
            setCursorPosition("outside")
            return
        }

        const rect = containerElement.getBoundingClientRect()
        const y = e instanceof PointerEvent ? e.clientY - rect.top : 0
        const registerThreshold = !isMiniPlayer ? 150 : 100 // pixels from the bottom to start registering position
        const showOnlyTimeRangeOffset = !isMiniPlayer ? 50 : 50

        if ((y >= rect.height - registerThreshold && y < rect.height - registerThreshold + showOnlyTimeRangeOffset)) {
            setCursorPosition("approaching")
        } else if (y < rect.height - registerThreshold) {
            setCursorPosition("outside")
        } else {
            setCursorPosition("hover")
        }
    }

    function handleVideoContainerPointerLeave(e: Event) {
        setCursorPosition("outside")
    }

    React.useEffect(() => {
        if (!containerElement) return
        containerElement.addEventListener("pointermove", handleVideoContainerPointerMove)
        containerElement.addEventListener("pointerleave", handleVideoContainerPointerLeave)
        containerElement.addEventListener("pointercancel", handleVideoContainerPointerLeave)
        return () => {
            containerElement.removeEventListener("pointermove", handleVideoContainerPointerMove)
            containerElement.removeEventListener("pointerup", handleVideoContainerPointerLeave)
            containerElement.removeEventListener("pointercancel", handleVideoContainerPointerLeave)
        }
    }, [containerElement, paused, isMiniPlayer, seeking, hoveringControlBar])

    return (
        <>
            <div
                className={cn(
                    "vc-control-bar-bottom-gradient pointer-events-none",
                    "absolute bottom-0 left-0 right-0 w-full z-[5] h-28 transition-opacity duration-300 opacity-0",
                    "bg-gradient-to-t to-transparent",
                    !isMiniPlayer ? "from-black/40" : "from-black/80 via-black/40",
                    isMiniPlayer && "h-20",
                    !hideShadow && "opacity-100",
                )}
            />
            {!isMiniPlayer && <div
                className={cn(
                    "vc-control-bar-bottom-gradient-time-range-only pointer-events-none",
                    "absolute bottom-0 left-0 right-0 w-full z-[5] h-14 transition-opacity duration-400 opacity-0",
                    "bg-gradient-to-t to-transparent",
                    !isMiniPlayer ? "from-black/40" : "from-black/60",
                    isMiniPlayer && "h-10",
                    (showOnlyTimeRange && paused && hideShadow) && "opacity-100",
                )}
            />}
            <div
                data-vc-control-bar-section
                className={cn(
                    "vc-control-bar-section",
                    "absolute left-0 bottom-0 right-0 flex flex-col text-white",
                    "transition-all duration-300 opacity-0",
                    "z-[10] h-28",
                    !hideControlBar && "opacity-100",
                    VIDEOCORE_DEBUG_ELEMENTS && "bg-purple-500/20",
                )}
                style={{
                    bottom: `${controlBarBottomPx}px`,
                }}
                onPointerEnter={() => {
                    setHoveringControlBar(true)
                }}
                onPointerLeave={() => {
                    setHoveringControlBar(false)
                }}
                onPointerCancel={() => {
                    setHoveringControlBar(false)
                }}
            >
                <div
                    className={cn(
                        "vc-control-bar",
                        "absolute bottom-0 w-full px-4",
                        VIDEOCORE_DEBUG_ELEMENTS && "bg-purple-800/40",
                    )}
                    // style={{
                    //     paddingTop: VIDEOCORE_CONTROL_BAR_VPADDING,
                    //     paddingBottom: VIDEOCORE_CONTROL_BAR_VPADDING,
                    // }}
                >
                    {timeRange}

                    <div
                        className={cn(
                            "vc-control-bar-main-section z-[100] relative",
                            "transform-gpu duration-100 flex items-center pb-2",
                        )}
                        style={{
                            height: `${mainSectionHeight}px`,
                            // "--tw-translate-y": showOnlyTimeRange ? `-${mainSectionHeight}px` : 0,
                        } as React.CSSProperties}
                    >
                        {children}
                    </div>
                </div>
            </div>
        </>
    )
}

type VideoCoreControlButtonProps = {
    icons: [string, React.ElementType][]
    state: string
    className?: string
    iconClass?: string
    onClick: () => void
    onWheel?: (e: React.WheelEvent<HTMLButtonElement>) => void
}

export function VideoCoreControlButtonIcon(props: VideoCoreControlButtonProps) {
    const { icons, state, className, iconClass, onClick, onWheel } = props

    const isMiniPlayer = useAtomValue(vc_miniPlayer)

    const size = isMiniPlayer ? VIDEOCORE_CONTROL_BAR_MAIN_SECTION_HEIGHT_MINI : VIDEOCORE_CONTROL_BAR_MAIN_SECTION_HEIGHT

    return (
        <button
            role="button"
            style={{}}
            className={cn(
                "vc-control-button flex items-center justify-center px-2 transition-opacity hover:opacity-80 relative h-full",
                "focus-visible:outline-none focus:outline-none focus-visible:opacity-50",
                "text-3xl",
                isMiniPlayer && "text-2xl",
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
                            className="block"
                            initial={{ opacity: 0, y: 10, position: "relative" }}
                            animate={{ opacity: 1, y: 0, position: "relative" }}
                            exit={{ opacity: 0, y: 10, position: "absolute" }}
                            transition={{ duration: 0.15 }}
                        >
                            <Icon
                                className={cn(
                                    "vc-control-button-icon",
                                    iconClass,
                                )}
                            />
                        </motion.span>
                    )
                })}
            </AnimatePresence>
        </button>
    )
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export function VideoCorePlayButton() {
    const paused = useAtomValue(vc_paused)
    const action = useSetAtom(vc_dispatchAction)

    return (
        <VideoCoreControlButtonIcon
            icons={[
                ["playing", RiPauseLargeLine],
                ["paused", RiPlayLargeLine],
            ]}
            state={paused ? "paused" : "playing"}
            onClick={() => {
                action({ type: "togglePlay" })
            }}
        />
    )
}

export function VideoCoreVolumeButton() {
    const volume = useAtomValue(vc_volume)
    const muted = useAtomValue(vc_isMuted)
    const setVolume = useSetAtom(vc_storedVolumeAtom)
    const setMuted = useSetAtom(vc_storedMutedAtom)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)

    const [isSliding, setIsSliding] = React.useState(false)

    // Uses a power curve to give more granular control at lower volumes
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
        setVolume(nonLinearVolume)
        setMuted(nonLinearVolume === 0)
    }

    function handlePointerUp(e: React.PointerEvent<HTMLDivElement>) {
        if (isSliding) {
            e.stopPropagation()
            e.currentTarget.setPointerCapture(e.pointerId)
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
        setVolume(newVolume)
        setMuted(newVolume === 0)
    }

    return (
        <div
            className={cn(
                "vc-control-volume group/vc-control-volume",
                "flex items-center justify-center h-full gap-2",
            )}
        >
            <VideoCoreControlButtonIcon
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
                onClick={() => {
                    setMuted(p => {
                        if (p && volume === 0) setVolume(0.1)
                        return !p
                    })
                }}
                onWheel={handleWheel}
            />
            <div
                className={cn(
                    "vc-control-volume-slider-container relative w-0 flex group-hover/vc-control-volume:w-[6rem] h-6",
                    "transition-[width] duration-300",
                )}
            >
                <div
                    className={cn(
                        "vc-control-volume-slider",
                        "flex h-full w-full relative items-center",
                        "rounded-full",
                        "cursor-pointer",
                        "transition-all duration-300",
                    )}
                    onPointerDown={handlePointerDown}
                    onPointerMove={handlePointerMove}
                    onPointerUp={handlePointerUp}
                    onPointerCancel={handlePointerUp}
                    onWheel={handleWheel}
                >
                    <div
                        className={cn(
                            "vc-control-volume-slider-progress h-1.5",
                            "absolute bg-white",
                            "rounded-full",
                        )}
                        style={{
                            width: `${volumeToLinear(volume) * 100}%`,
                        }}
                    />
                    <div
                        className={cn(
                            "vc-control-volume-slider-progress h-1.5 w-full",
                            "absolute bg-white/20",
                            "rounded-full",
                        )}
                    />
                </div>
                <div className="w-4" />
            </div>
        </div>
    )
}

export function VideoCoreNextButton({ onClick }: { onClick: () => void }) {
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    if (isMiniPlayer) return null

    return (
        <VideoCoreControlButtonIcon
            icons={[
                ["default", LuChevronRight],
            ]}
            state="default"
            onClick={onClick}
        />
    )
}


export function VideoCorePreviousButton({ onClick }: { onClick: () => void }) {
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    if (isMiniPlayer) return null

    return (
        <VideoCoreControlButtonIcon
            icons={[
                ["default", LuChevronLeft],
            ]}
            state="default"
            onClick={onClick}
        />
    )
}

const vc_timestampType = atomWithStorage("sea-video-core-timestamp-type", "elapsed", undefined, { getOnInit: true })

export function VideoCoreTimestamp() {
    const duration = useAtomValue(vc_duration)
    const currentTime = useAtomValue(vc_currentTime)
    const [type, setType] = useAtom(vc_timestampType)

    function handleSwitchType() {
        setType(p => p === "elapsed" ? "remaining" : "elapsed")
    }

    if (duration <= 1 || isNaN(duration)) return null

    return (
        <p className="font-medium text-sm opacity-100 hover:opacity-80 cursor-pointer" onClick={handleSwitchType}>
            {type === "remaining" ? "-" : ""}{vc_formatTime(Math.max(0,
            Math.min(duration, type === "elapsed" ? currentTime : duration - currentTime)))} / {vc_formatTime(duration)}
        </p>
    )
}

export function VideoCorePipButton() {
    const pipManager = useAtomValue(vc_pipManager)
    const isPip = useAtomValue(vc_pip)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)

    if (isMiniPlayer) return null

    return (
        <VideoCoreControlButtonIcon
            icons={[
                ["default", TbPictureInPicture],
                ["pip", TbPictureInPictureOff],
            ]}
            state={isPip ? "pip" : "default"}
            onClick={() => {
                pipManager?.togglePip()
            }}
        />
    )
}

export function VideoCoreFullscreenButton() {
    const fullscreenManager = useAtomValue(vc_fullscreenManager)
    const isFullscreen = useAtomValue(vc_isFullscreen)
    const [isMiniPlayer, setMiniPlayer] = useAtom(vc_miniPlayer)

    return (
        <VideoCoreControlButtonIcon
            icons={[
                ["default", RxEnterFullScreen],
                ["fullscreen", RxExitFullScreen],
            ]}
            state={isFullscreen ? "fullscreen" : "default"}
            onClick={() => {
                setMiniPlayer(false)
                fullscreenManager?.toggleFullscreen()
            }}
        />
    )
}
