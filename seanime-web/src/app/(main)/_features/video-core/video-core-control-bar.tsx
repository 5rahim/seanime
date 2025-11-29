import { nativePlayer_stateAtom } from "@/app/(main)/_features/native-player/native-player.atoms"
import {
    MediaCaptionsTrack,
    vc_audioManager,
    vc_containerElement,
    vc_currentTime,
    vc_cursorBusy,
    vc_dispatchAction,
    vc_duration,
    vc_isFullscreen,
    vc_isMuted,
    vc_mediaCaptionsManager,
    vc_miniPlayer,
    vc_paused,
    vc_pip,
    vc_playbackRate,
    vc_seeking,
    vc_subtitleManager,
    vc_videoElement,
    vc_volume,
    VIDEOCORE_DEBUG_ELEMENTS,
} from "@/app/(main)/_features/video-core/video-core"
import { anime4kOptions, getAnime4KOptionByValue, vc_anime4kOption } from "@/app/(main)/_features/video-core/video-core-anime-4k"
import { Anime4KOption } from "@/app/(main)/_features/video-core/video-core-anime-4k-manager"
import { vc_fullscreenManager } from "@/app/(main)/_features/video-core/video-core-fullscreen"
import {
    vc_hlsAudioTracks,
    vc_hlsCurrentAudioTrack,
    vc_hlsCurrentQuality,
    vc_hlsQualityLevels,
    vc_hlsSetQuality,
} from "@/app/(main)/_features/video-core/video-core-hls"
import {
    vc_menuOpen,
    vc_menuSectionOpen,
    vc_menuSubSectionOpen,
    VideoCoreMenu,
    VideoCoreMenuBody,
    VideoCoreMenuOption,
    VideoCoreMenuSectionBody,
    VideoCoreMenuSubmenuBody,
    VideoCoreMenuSubOption,
    VideoCoreMenuSubSubmenuBody,
    VideoCoreMenuTitle,
    VideoCoreSettingSelect,
    VideoCoreSettingTextInput,
} from "@/app/(main)/_features/video-core/video-core-menu"
import { vc_pipManager } from "@/app/(main)/_features/video-core/video-core-pip"
import { videoCorePreferencesModalAtom } from "@/app/(main)/_features/video-core/video-core-preferences"
import { NormalizedTrackInfo } from "@/app/(main)/_features/video-core/video-core-subtitles"
import {
    vc_autoNextAtom,
    vc_autoPlayVideoAtom,
    vc_autoSkipOPEDAtom,
    vc_beautifyImageAtom,
    vc_highlightOPEDChaptersAtom,
    vc_initialSettings,
    vc_settings,
    vc_showChapterMarkersAtom,
    vc_storedMutedAtom,
    vc_storedPlaybackRateAtom,
    vc_storedVolumeAtom,
    VideoCoreSettings,
} from "@/app/(main)/_features/video-core/video-core.atoms"
import { vc_formatTime } from "@/app/(main)/_features/video-core/video-core.utils"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Switch } from "@/components/ui/switch"
import { Tooltip } from "@/components/ui/tooltip"
import { atom, useAtomValue } from "jotai"
import { useAtom, useSetAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import { AnimatePresence, motion } from "motion/react"
import React, { useState } from "react"
import { AiFillInfoCircle } from "react-icons/ai"
import { HiFastForward } from "react-icons/hi"
import { ImFileText } from "react-icons/im"
import { IoCaretForwardCircleOutline } from "react-icons/io5"
import {
    LuCaptions,
    LuChevronLeft,
    LuChevronRight,
    LuChevronUp,
    LuFilm,
    LuHeading,
    LuHeadphones,
    LuPaintbrush,
    LuPalette,
    LuSettings2,
    LuSparkles,
    LuTvMinimalPlay,
    LuVolume,
    LuVolume1,
    LuVolume2,
    LuVolumeOff,
} from "react-icons/lu"
import { MdOutlineSubtitles, MdSpeed } from "react-icons/md"
import { RiPauseLargeLine, RiPlayLargeLine, RiShadowLine } from "react-icons/ri"
import { RxEnterFullScreen, RxExitFullScreen } from "react-icons/rx"
import { TbArrowForwardUp, TbPictureInPicture, TbPictureInPictureOff } from "react-icons/tb"
import { VscTextSize } from "react-icons/vsc"
import { toast } from "sonner"

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

function VideoCoreControlButtonIcon(props: VideoCoreControlButtonProps) {
    const { icons, state, className, iconClass, onClick, onWheel } = props

    const isMiniPlayer = useAtomValue(vc_miniPlayer)

    const size = isMiniPlayer ? VIDEOCORE_CONTROL_BAR_MAIN_SECTION_HEIGHT_MINI : VIDEOCORE_CONTROL_BAR_MAIN_SECTION_HEIGHT

    return (
        <button
            role="button"
            style={{}}
            className={cn(
                "vc-control-button flex items-center justify-center px-2 transition-opacity hover:opacity-80 relative h-full",
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

export function VideoCoreAudioButton() {
    const action = useSetAtom(vc_dispatchAction)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const state = useAtomValue(nativePlayer_stateAtom)
    const audioManager = useAtomValue(vc_audioManager)
    const videoElement = useAtomValue(vc_videoElement)
    const isFullscreen = useAtomValue(vc_isFullscreen)
    const containerElement = useAtomValue(vc_containerElement)
    const [selectedTrack, setSelectedTrack] = React.useState<number | null>(null)

    // Get MKV audio tracks
    const mkvAudioTracks = state.playbackInfo?.mkvMetadata?.audioTracks

    // Get HLS audio tracks
    const hlsAudioTracks = useAtomValue(vc_hlsAudioTracks)
    const hlsCurrentAudioTrack = useAtomValue(vc_hlsCurrentAudioTrack)

    // Determine which audio tracks to use
    const audioTracks = mkvAudioTracks || (hlsAudioTracks.length > 0 ? hlsAudioTracks : null)
    const isHls = !mkvAudioTracks && hlsAudioTracks.length > 0

    function onAudioChange() {
        setSelectedTrack(audioManager?.getSelectedTrack?.() ?? null)
    }

    React.useEffect(() => {
        if (!videoElement || !audioManager) return

        videoElement?.audioTracks?.addEventListener?.("change", onAudioChange)
        return () => {
            videoElement?.audioTracks?.removeEventListener?.("change", onAudioChange)
        }
    }, [videoElement, audioManager])

    React.useEffect(() => {
        onAudioChange()
    }, [audioManager])

    // Update selected track when HLS audio track changes
    React.useEffect(() => {
        if (isHls && hlsCurrentAudioTrack !== -1) {
            setSelectedTrack(hlsCurrentAudioTrack)
        }
    }, [hlsCurrentAudioTrack, isHls])

    if (isMiniPlayer || !audioTracks?.length || audioTracks.length === 1) return null

    return (
        <VideoCoreMenu
            name="audio"
            trigger={<VideoCoreControlButtonIcon
                icons={[
                    ["default", LuHeadphones],
                ]}
                state="default"
                className="text-2xl"
                onClick={() => {

                }}
            />}
        >
            <VideoCoreMenuTitle>Audio</VideoCoreMenuTitle>
            <VideoCoreMenuBody>
                <VideoCoreSettingSelect
                    isFullscreen={isFullscreen}
                    containerElement={containerElement}
                    options={audioTracks.map(track => {
                        if (isHls) {
                            // HLS track format
                            const hlsTrack = track as any
                            return {
                                label: hlsTrack.name || hlsTrack.language?.toUpperCase() || `Track ${hlsTrack.id + 1}`,
                                value: hlsTrack.id,
                                moreInfo: hlsTrack.language?.toUpperCase(),
                            }
                        } else {
                            // MKV track format
                            const mkvTrack = track as any
                            return {
                                label: `${mkvTrack.name || mkvTrack.language?.toUpperCase() || mkvTrack.languageIETF?.toUpperCase()}`,
                                value: mkvTrack.number,
                                moreInfo: mkvTrack.language?.toUpperCase(),
                            }
                        }
                    })}
                    onValueChange={(value: number) => {
                        audioManager?.selectTrack(value)
                        if (!isHls) {
                            action({ type: "seek", payload: { time: -1 } })
                        }
                    }}
                    value={selectedTrack || 0}
                />
            </VideoCoreMenuBody>
        </VideoCoreMenu>
    )
}

export function VideoCoreQualityButton() {
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const qualityLevels = useAtomValue(vc_hlsQualityLevels)
    const currentQuality = useAtomValue(vc_hlsCurrentQuality)
    const setQuality = useAtomValue(vc_hlsSetQuality)
    const isFullscreen = useAtomValue(vc_isFullscreen)
    const containerElement = useAtomValue(vc_containerElement)

    if (isMiniPlayer || !qualityLevels?.length) return null

    return (
        <VideoCoreMenu
            name="quality"
            trigger={<VideoCoreControlButtonIcon
                icons={[
                    ["default", LuFilm],
                ]}
                state="default"
                onClick={() => {}}
                className="text-2xl"
            />}
        >
            <VideoCoreMenuTitle>Quality</VideoCoreMenuTitle>
            <VideoCoreMenuBody>
                <VideoCoreSettingSelect
                    isFullscreen={isFullscreen}
                    containerElement={containerElement}
                    options={[
                        {
                            label: "Auto",
                            value: -1,
                        },
                        ...qualityLevels.map((level: any) => ({
                            label: level.name,
                            value: level.index,
                            moreInfo: `${Math.round(level.bitrate / 1000)}kbps`,
                        })).toReversed(),
                    ]}
                    onValueChange={(value: number) => {
                        setQuality?.(value)
                    }}
                    value={currentQuality}
                />
            </VideoCoreMenuBody>
        </VideoCoreMenu>
    )
}

export function VideoCoreSubtitleButton() {
    const action = useSetAtom(vc_dispatchAction)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const state = useAtomValue(nativePlayer_stateAtom)
    const subtitleManager = useAtomValue(vc_subtitleManager)
    const mediaCaptionsManager = useAtomValue(vc_mediaCaptionsManager)
    const videoElement = useAtomValue(vc_videoElement)
    const isFullscreen = useAtomValue(vc_isFullscreen)
    const containerElement = useAtomValue(vc_containerElement)
    const [selectedTrack, setSelectedTrack] = React.useState<number | null>(null)

    const [subtitleTracks, setSubtitleTracks] = React.useState<NormalizedTrackInfo[]>([])
    const [mediaCaptionsTracks, setMediaCaptionsTracks] = React.useState<MediaCaptionsTrack[]>([])

    function onTextTrackChange() {
        setSubtitleTracks(p => subtitleManager?.getTracks?.() ?? p)
    }

    function onTrackChange(trackNumber: number | null) {
        setSelectedTrack(trackNumber)
        if (trackNumber !== null && subtitleManager && !subtitleManager?.isTrackSupported(trackNumber)) {
            toast.error("This subtitle format is not supported by the player. Select another subtitle track or use an external player.")
        }
    }

    const firstRender = React.useRef(true)

    React.useEffect(() => {
        if (!videoElement) return

        /**
         * MKV subtitle tracks
         */
        if (subtitleManager) {
            if (firstRender.current) {
                // firstRender.current = false
                onTrackChange(subtitleManager?.getSelectedTrackNumberOrNull?.() ?? null)
            }

            // Listen for subtitle track changes
            subtitleManager.addTrackChangedEventListener(onTrackChange)

            // Listen for when the subtitle tracks are mounted
            videoElement?.textTracks?.addEventListener?.("change", onTextTrackChange)
            return () => {
                videoElement?.textTracks?.removeEventListener?.("change", onTextTrackChange)
            }
        } else if (mediaCaptionsManager) {
            /**
             * Media captions tracks
             */
            if (firstRender.current) {
                // firstRender.current = false
                setSelectedTrack(mediaCaptionsManager.getSelectedTrackIndexOrNull?.() ?? null)
            }

            // Listen for subtitle track changes
            mediaCaptionsManager.addTrackChangedEventListener(onTrackChange)

            const tracks = mediaCaptionsManager.getTracks?.() ?? []
            setMediaCaptionsTracks(tracks)
        }
    }, [videoElement, subtitleManager, mediaCaptionsManager])

    React.useEffect(() => {
        onTextTrackChange()
    }, [subtitleManager])

    // Get active manager
    const activeManager = subtitleManager || mediaCaptionsManager
    const activeTracks = subtitleManager ? subtitleTracks : mediaCaptionsTracks

    if (isMiniPlayer || !activeTracks?.length) return null

    return (
        <VideoCoreMenu
            name="subtitle"
            trigger={<VideoCoreControlButtonIcon
                icons={[
                    ["default", LuCaptions],
                ]}
                state="default"
                onClick={() => {

                }}
            />}
        >
            <VideoCoreMenuTitle>Subtitles {!!subtitleManager && <Tooltip
                trigger={<AiFillInfoCircle className="text-sm" />}
                className="z-[150]"
                portalContainer={containerElement ?? undefined}
            >
                You can add subtitles by dragging and dropping files onto the player.
            </Tooltip>}</VideoCoreMenuTitle>
            <VideoCoreMenuBody>
                <VideoCoreSettingSelect
                    isFullscreen={isFullscreen}
                    containerElement={containerElement}
                    options={[
                        {
                            label: "Off",
                            value: -1,
                        },
                        ...subtitleTracks.map(track => {
                            // MKV subtitle tracks
                            return {
                                label: `${track.label || track.language?.toUpperCase() || track.languageIETF?.toUpperCase()}`,
                                value: track.number,
                                moreInfo: track.language
                                    ? `${track.language.toUpperCase()}${track.codecID ? "/" + getSubtitleTrackType(track.codecID) : ``}`
                                    : undefined,
                            }
                        }),
                        ...mediaCaptionsTracks.map(track => {
                            return {
                                label: track.label,
                                value: track.number,
                                moreInfo: track.language?.toUpperCase(),
                            }
                        }),
                    ]}
                    onValueChange={(value: number) => {
                        if (value === -1) {
                            activeManager?.setNoTrack()
                            setSelectedTrack(null)
                            return
                        }
                        if (subtitleManager) {
                            subtitleManager.selectTrack(value)
                        } else if (mediaCaptionsManager) {
                            mediaCaptionsManager.selectTrack(value)
                            setSelectedTrack(value)
                        }
                    }}
                    value={selectedTrack ?? -1}
                />
            </VideoCoreMenuBody>
        </VideoCoreMenu>
    )
}

export function getSubtitleTrackType(codecID: string) {
    switch (codecID) {
        case "S_TEXT/ASS":
            return "ASS"
        case "S_TEXT/SSA":
            return "SSA"
        case "S_TEXT/UTF8":
            return "TEXT"
        case "S_HDMV/PGS":
            return "PGS"
    }
    return "unknown"
}

export function VideoCoreSettingsButton() {
    const action = useSetAtom(vc_dispatchAction)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const playbackRate = useAtomValue(vc_playbackRate)
    const setPlaybackRate = useSetAtom(vc_storedPlaybackRateAtom)
    const isFullscreen = useAtomValue(vc_isFullscreen)
    const containerElement = useAtomValue(vc_containerElement)
    const subtitleManager = useAtomValue(vc_subtitleManager)

    const [anime4kOption, setAnime4kOption] = useAtom(vc_anime4kOption)
    const currentAnime4kOption = getAnime4KOptionByValue(anime4kOption)

    const [, setKeybindingsModelOpen] = useAtom(videoCorePreferencesModalAtom)

    const [showChapterMarkers, setShowChapterMarkers] = useAtom(vc_showChapterMarkersAtom)
    const [highlightOPEDChapters, setHighlightOPEDChapters] = useAtom(vc_highlightOPEDChaptersAtom)
    const [beautifyImage, setBeautifyImage] = useAtom(vc_beautifyImageAtom)
    const [autoNext, setAutoNext] = useAtom(vc_autoNextAtom)
    const [autoPlay, setAutoPlay] = useAtom(vc_autoPlayVideoAtom)
    const [autoSkipOPED, setAutoSkipOPED] = useAtom(vc_autoSkipOPEDAtom)

    const [menuOpen, setMenuOpen] = useAtom(vc_menuOpen)
    const [openMenuSection, setOpenMenuSection] = useAtom(vc_menuSectionOpen)
    const [openMenuSubSection, setOpenMenuSubSection] = useAtom(vc_menuSubSectionOpen)

    const [settings, setSettings] = useAtom(vc_settings)

    const [editedSubCustomization, setEditedSubCustomization] = useState<VideoCoreSettings["subtitleCustomization"]>(
        settings.subtitleCustomization || vc_initialSettings.subtitleCustomization,
    )

    const [subFontName, setSubFontName] = useState<string>(editedSubCustomization?.fontName || "")

    React.useEffect(() => {
        if (openMenuSection === "Subtitle Styles") {
            setEditedSubCustomization(settings.subtitleCustomization || vc_initialSettings.subtitleCustomization)
        }
    }, [openMenuSection, settings])

    const handleSaveSettings = (customization?: VideoCoreSettings["subtitleCustomization"]) => {
        const newSettings = {
            ...settings,
            subtitleCustomization: customization || editedSubCustomization,
        }
        setSettings(newSettings)
        subtitleManager?.updateSettings(newSettings)

        // // Go back to submenu after saving from sub-submenu
        // setOpenMenuSubSection(null)
    }

    const handleSubtitleCustomizationChange = <K extends keyof VideoCoreSettings["subtitleCustomization"]>(
        key: K,
        value: VideoCoreSettings["subtitleCustomization"][K],
    ): void => {
        const newCustomization = {
            ...editedSubCustomization,
            [key]: value,
        }
        setEditedSubCustomization(newCustomization)
        React.startTransition(() => {
            handleSaveSettings(newCustomization)
        })
    }

    if (isMiniPlayer) return null

    return (
        <>
            {playbackRate !== 1 && (
                <p
                    className="text-sm text-[--muted] cursor-pointer" onClick={() => {
                    setMenuOpen("settings")
                    React.startTransition(() => {
                        setOpenMenuSection("Playback Speed")
                    })
                }}
                >
                    {`${(playbackRate).toFixed(2)}x`}
                </p>
            )}
            <VideoCoreMenu
                name="settings"
                trigger={<VideoCoreControlButtonIcon
                    icons={[
                        ["default", LuChevronUp],
                    ]}
                    state="default"
                    onClick={() => {
                    }}
                />}
            >
                <VideoCoreMenuSectionBody>
                    <VideoCoreMenuTitle>Settings</VideoCoreMenuTitle>
                    <VideoCoreMenuOption title="Playback Speed" icon={MdSpeed} value={`${(playbackRate).toFixed(2)}x`} />
                    <VideoCoreMenuOption title="Auto Play" icon={IoCaretForwardCircleOutline} value={autoPlay ? "On" : "Off"} />
                    <VideoCoreMenuOption title="Auto Next" icon={HiFastForward} value={autoNext ? "On" : "Off"} />
                    <VideoCoreMenuOption title="Skip OP/ED" icon={TbArrowForwardUp} value={autoSkipOPED ? "On" : "Off"} />
                    <VideoCoreMenuOption title="Anime4K" icon={LuSparkles} value={currentAnime4kOption?.label || "Off"} />
                    <VideoCoreMenuOption
                        title="Subtitle Styles"
                        icon={MdOutlineSubtitles}
                        value={editedSubCustomization?.enabled ? `On${!!editedSubCustomization?.fontName ? ", Font" : ""}` : "Off"}
                    />
                    <VideoCoreMenuOption title="Player Appearance" icon={LuTvMinimalPlay} />
                    <VideoCoreMenuOption title="Preferences" icon={LuSettings2} onClick={() => setKeybindingsModelOpen(true)} />
                </VideoCoreMenuSectionBody>
                <VideoCoreMenuSubmenuBody>
                    <VideoCoreMenuOption title="Subtitle Styles" icon={MdOutlineSubtitles}>
                        <p className="text-sm text-[--muted] mb-2">Subtitle customization will not override ASS/SAA tracks that contain multiple
                                                                   styles.</p>
                        <VideoCoreSettingSelect
                            options={[
                                { label: "On", value: 1 },
                                { label: "Off", value: 0 },
                            ]}
                            onValueChange={(v: number) => handleSubtitleCustomizationChange("enabled", v === 1)}
                            value={editedSubCustomization.enabled ? 1 : 0}
                        />
                        {editedSubCustomization.enabled && <>
                            <p className="text-[--muted] text-sm my-2">Options</p>
                            <VideoCoreMenuSubOption
                                title="Font"
                                icon={LuHeading}
                                parentId="Subtitle Styles"
                                value={editedSubCustomization.fontName?.slice(0,
                                    11) + (!!editedSubCustomization.fontName?.length && editedSubCustomization.fontName?.length > 10
                                    ? "..."
                                    : "")}
                            />
                            <VideoCoreMenuSubOption title="Font Size" icon={VscTextSize} parentId="Subtitle Styles" />
                            <VideoCoreMenuSubOption title="Text Color" icon={LuPalette} parentId="Subtitle Styles" />
                            <VideoCoreMenuSubOption title="Outline" icon={ImFileText} parentId="Subtitle Styles" />
                            <VideoCoreMenuSubOption title="Shadow" icon={RiShadowLine} parentId="Subtitle Styles" />
                        </>}
                    </VideoCoreMenuOption>
                    <VideoCoreMenuOption title="Playback Speed" icon={MdSpeed}>
                        <VideoCoreSettingSelect
                            options={[
                                { label: "0.5x", value: 0.5 },
                                { label: "0.9x", value: 0.9 },
                                { label: "1x", value: 1 },
                                { label: "1.1x", value: 1.1 },
                                { label: "1.5x", value: 1.5 },
                                { label: "2x", value: 2 },
                            ]}
                            onValueChange={(v: number) => {
                                setPlaybackRate(v)
                            }}
                            value={playbackRate}
                        />
                    </VideoCoreMenuOption>
                    <VideoCoreMenuOption title="Auto Play" icon={IoCaretForwardCircleOutline}>
                        <VideoCoreSettingSelect
                            options={[
                                { label: "On", value: 1 },
                                { label: "Off", value: 0 },
                            ]}
                            onValueChange={(v: number) => {
                                setAutoPlay(!!v)
                            }}
                            value={autoPlay ? 1 : 0}
                        />
                    </VideoCoreMenuOption>
                    <VideoCoreMenuOption title="Auto Next" icon={HiFastForward}>
                        <VideoCoreSettingSelect
                            options={[
                                { label: "On", value: 1 },
                                { label: "Off", value: 0 },
                            ]}
                            onValueChange={(v: number) => {
                                setAutoNext(!!v)
                            }}
                            value={autoNext ? 1 : 0}
                        />
                    </VideoCoreMenuOption>
                    <VideoCoreMenuOption title="Skip OP/ED" icon={TbArrowForwardUp}>
                        <VideoCoreSettingSelect
                            options={[
                                { label: "On", value: 1 },
                                { label: "Off", value: 0 },
                            ]}
                            onValueChange={(v: number) => {
                                setAutoSkipOPED(!!v)
                            }}
                            value={autoSkipOPED ? 1 : 0}
                        />
                    </VideoCoreMenuOption>
                    <VideoCoreMenuOption title="Anime4K" icon={LuSparkles}>
                        <p className="text-[--muted] text-sm mb-2">
                            Real-time sharpening. GPU-intensive.
                        </p>
                        <VideoCoreSettingSelect
                            isFullscreen={isFullscreen}
                            containerElement={containerElement}
                            options={anime4kOptions.map(option => ({
                                label: `${option.label}`,
                                value: option.value,
                                moreInfo: option.performance === "heavy" ? "Heavy" : undefined,
                                description: option.description,
                            }))}
                            onValueChange={(value: Anime4KOption) => {
                                setAnime4kOption(value)
                            }}
                            value={anime4kOption}
                        />
                    </VideoCoreMenuOption>
                    <VideoCoreMenuOption title="Player Appearance" icon={LuPaintbrush}>
                        <Switch
                            label="Show Chapter Markers"
                            side="right"
                            fieldClass="hover:bg-transparent hover:border-transparent px-0 ml-0 w-full"
                            size="sm"
                            value={showChapterMarkers}
                            onValueChange={setShowChapterMarkers}
                        />
                        <Switch
                            label="Highlight OP/ED Chapters"
                            side="right"
                            fieldClass="hover:bg-transparent hover:border-transparent px-0 ml-0 w-full"
                            size="sm"
                            value={highlightOPEDChapters}
                            onValueChange={setHighlightOPEDChapters}
                        />
                        <Switch
                            label="Increase Saturation"
                            side="right"
                            fieldClass="hover:bg-transparent hover:border-transparent px-0 ml-0 w-full"
                            size="sm"
                            value={beautifyImage}
                            onValueChange={setBeautifyImage}
                        />
                    </VideoCoreMenuOption>
                </VideoCoreMenuSubmenuBody>
                <VideoCoreMenuSubSubmenuBody>
                    <VideoCoreMenuSubOption title="Font" icon={VscTextSize} parentId="Subtitle Styles">
                        <div className="">
                            <p className="text-sm mb-2">Custom Font</p>
                            <p className="text-sm text-[--muted] mb-2">
                                Place the font files in a folder named <strong>assets</strong> in the Seanime data directory. The file name must match
                                the font name exactly.
                            </p>
                            <div className="space-y-2">
                                <VideoCoreSettingTextInput
                                    label="File Name"
                                    value={subFontName ?? ""}
                                    onValueChange={(v: string) => setSubFontName(v)}
                                />
                                <div className="flex w-full">
                                    <Button
                                        size="sm" intent="gray-glass" onClick={() => {
                                        handleSubtitleCustomizationChange("fontName", subFontName)
                                    }}
                                    >
                                        Save
                                    </Button>
                                </div>
                            </div>
                        </div>
                    </VideoCoreMenuSubOption>
                    <VideoCoreMenuSubOption title="Font Size" icon={LuHeading} parentId="Subtitle Styles">
                        <p className="text-[--muted] text-sm mb-2">Font Size</p>
                        <VideoCoreSettingSelect
                            options={[
                                { label: "Small", value: 54 },
                                { label: "Medium", value: 62 },
                                { label: "Large", value: 72 },
                                { label: "Extra Large", value: 82 },
                            ]}
                            onValueChange={(v: number) => handleSubtitleCustomizationChange("fontSize", v)}
                            value={editedSubCustomization.fontSize ?? 62}
                        />
                    </VideoCoreMenuSubOption>
                    <VideoCoreMenuSubOption title="Text Color" icon={LuPalette} parentId="Subtitle Styles">
                        <VideoCoreSettingSelect
                            options={[
                                { label: "White", value: "#FFFFFF" },
                                { label: "Black", value: "#000000" },
                                { label: "Yellow", value: "#FFD700" },
                                { label: "Cyan", value: "#00FFFF" },
                                { label: "Pink", value: "#FF69B4" },
                                { label: "Purple", value: "#9370DB" },
                                { label: "Lime", value: "#00FF00" },
                            ]}
                            onValueChange={(v: string) => handleSubtitleCustomizationChange("primaryColor", v)}
                            value={editedSubCustomization.primaryColor ?? "#FFFFFF"}
                        />
                    </VideoCoreMenuSubOption>
                    <VideoCoreMenuSubOption title="Outline" icon={LuPalette} parentId="Subtitle Styles">
                        <p className="text-[--muted] text-sm mb-2">Outline Width</p>
                        <VideoCoreSettingSelect
                            options={[
                                { label: "None", value: 0 },
                                { label: "Small", value: 2 },
                                { label: "Medium", value: 3 },
                                { label: "Large", value: 4 },
                            ]}
                            onValueChange={(v: number) => handleSubtitleCustomizationChange("outline", v)}
                            value={editedSubCustomization.outline ?? 3}
                        />
                        <p className="text-[--muted] text-sm my-2">Outline Color</p>
                        <VideoCoreSettingSelect
                            options={[
                                { label: "Black", value: "#000000" },
                                { label: "White", value: "#FFFFFF" },
                                { label: "Yellow", value: "#FFD700" },
                                { label: "Cyan", value: "#00FFFF" },
                                { label: "Pink", value: "#FF69B4" },
                                { label: "Purple", value: "#9370DB" },
                                { label: "Lime", value: "#00FF00" },
                            ]}
                            onValueChange={(v: string) => handleSubtitleCustomizationChange("outlineColor", v)}
                            value={editedSubCustomization.outlineColor ?? "#000000"}
                        />
                    </VideoCoreMenuSubOption>
                    <VideoCoreMenuSubOption title="Shadow" icon={LuPalette} parentId="Subtitle Styles">
                        <p className="text-[--muted] text-sm mb-2">Shadow Depth</p>
                        <VideoCoreSettingSelect
                            options={[
                                { label: "None", value: 0 },
                                { label: "Small", value: 1 },
                                { label: "Medium", value: 2 },
                                { label: "Large", value: 3 },
                            ]}
                            onValueChange={(v: number) => handleSubtitleCustomizationChange("shadow", v)}
                            value={editedSubCustomization.shadow ?? 0}
                        />
                        <p className="text-[--muted] text-sm my-2">Shadow Color</p>
                        <VideoCoreSettingSelect
                            options={[
                                { label: "Black", value: "#000000" },
                                { label: "White", value: "#FFFFFF" },
                                { label: "Yellow", value: "#FFD700" },
                                { label: "Cyan", value: "#00FFFF" },
                                { label: "Pink", value: "#FF69B4" },
                                { label: "Purple", value: "#9370DB" },
                                { label: "Lime", value: "#00FF00" },
                            ]}
                            onValueChange={(v: string) => handleSubtitleCustomizationChange("backColor", v)}
                            value={editedSubCustomization.backColor ?? "#000000"}
                        />
                    </VideoCoreMenuSubOption>
                </VideoCoreMenuSubSubmenuBody>
            </VideoCoreMenu>
        </>
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
