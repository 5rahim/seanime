import {
    vc_containerElement,
    vc_cursorBusy,
    vc_miniPlayer,
    vc_paused,
    vc_seeking,
    VIDEOCORE_DEBUG_ELEMENTS,
} from "@/app/(main)/_features/video-core/video-core"
import { cn } from "@/components/ui/core/styling"
import { atom, useAtomValue } from "jotai"
import { useAtom } from "jotai/react"
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
        if (paused) {
            setBottom(0)
        }
        return () => {
            containerElement.removeEventListener("pointermove", handleVideoContainerPointerMove)
        }
    }, [containerElement, paused, isMiniPlayer, seeking, hoveringControlBar])

    return (
        <>
            <div
                className={cn(
                    "vc-control-bar-bottom-gradient",
                    "absolute bottom-0 left-0 right-0 w-full z-[1] h-32 transition-opacity duration-300 opacity-0",
                    "bg-gradient-to-t to-transparent",
                    !isMiniPlayer ? "from-black/80 via-black/50" : "from-black/80 via-black/40",
                    isMiniPlayer && "h-20",
                    ((showOnlyTimeRange || bottom != 0) && !paused) ? "opacity-0" : (cursorBusy || paused || showOnlyTimeRange) ? "opacity-100" : "",
                )}
                style={{
                    // "--tw-translate-y": (showOnlyTimeRange || bottom !== 0) && "-100%"
                } as React.CSSProperties}
            />
            <div
                data-vc-control-bar-section
                className={cn(
                    "vc-control-bar-section",
                    "absolute left-0 bottom-0 right-0 flex flex-col",
                    "transition-all duration-300 opacity-0",
                    "z-[100] h-28",
                    (cursorBusy || paused || showOnlyTimeRange) && "opacity-100",
                    VIDEOCORE_DEBUG_ELEMENTS && "bg-purple-500/20",
                )}
                style={{
                    bottom: (showOnlyTimeRange && !paused) ? `-${mainSectionHeight}px` : bottom,
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
                    className={cn(
                        "vc-control-bar",
                        "absolute bottom-0 w-full px-4",
                        VIDEOCORE_DEBUG_ELEMENTS && "bg-purple-800/40",
                    )}
                >
                    {timeRange}

                    <div
                        className={cn(
                            "vc-control-bar-main-section",
                            "transform-gpu duration-100 flex items-center",
                        )}
                        style={{
                            height: `${mainSectionHeight}px`,
                            "--tw-translate-y": (showOnlyTimeRange && !paused) ? `-${mainSectionHeight}px` : 0,
                        } as React.CSSProperties}
                    >
                        {children}
                    </div>
                </div>
            </div>
        </>
    )
}

