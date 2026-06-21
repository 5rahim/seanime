import { vc_hoveringControlBar } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_isMobile } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_isSwiping } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_duration } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_currentTime } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_isMuted } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_volume } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_isFullscreen } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_seeking } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_paused } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_miniPlayer } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_cursorBusy } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_containerElement } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_fullscreenManager } from "@/app/(main)/_features/video-core/video-core-fullscreen"
import { vc_pip } from "@/app/(main)/_features/video-core/video-core-pip"
import { vc_pipManager } from "@/app/(main)/_features/video-core/video-core-pip"
import { vc_storedMutedAtom, vc_storedVolumeAtom } from "@/app/(main)/_features/video-core/video-core.atoms"
import { vc_dispatchAction } from "@/app/(main)/_features/video-core/video-core.utils"
import {
    MediaCoreControlBarView,
    MediaCoreMobileControlBarView,
    MediaCorePlayButton,
    MediaCoreVolumeButton,
    MediaCoreNextButton,
    MediaCorePreviousButton,
    MediaCoreTimestamp,
    MediaCorePipButton,
    MediaCoreFullscreenButton,
    MediaCoreControlButtonIcon,
} from "@/app/(main)/_features/media-core/media-core-control-bar"
import { useAtomValue, atom } from "jotai"
import { useAtom, useSetAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import { mediaCorePreferencesAtom } from "@/app/(main)/_features/media-core/media-core-preferences"
import React from "react"

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
    const isMobile = useAtomValue(vc_isMobile)

    return (
        <MediaCoreControlBarView
            paused={paused}
            isMiniPlayer={isMiniPlayer}
            cursorBusy={cursorBusy}
            hoveringControlBar={hoveringControlBar}
            onHoveringControlBarChange={setHoveringControlBar}
            containerElement={containerElement}
            isMobile={isMobile}
            timeRange={timeRange}
        >
            {children}
        </MediaCoreControlBarView>
    )
}

export function VideoCoreMobileControlBar(props: {
    children?: React.ReactNode
    timeRange: React.ReactNode
    topLeftSection: React.ReactNode
    topRightSection: React.ReactNode
    bottomLeftSection: React.ReactNode
    bottomRightSection: React.ReactNode
}) {
    const { timeRange, topLeftSection, topRightSection, bottomLeftSection, bottomRightSection } = props

    const paused = useAtomValue(vc_paused)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const cursorBusy = useAtomValue(vc_cursorBusy)
    const seeking = useAtomValue(vc_seeking)
    const isSwiping = useAtomValue(vc_isSwiping)
    const [, setHoveringControlBar] = useAtom(vc_hoveringControlBar)

    React.useEffect(() => {
        setHoveringControlBar(false)
    }, [])

    return (
        <MediaCoreMobileControlBarView
            paused={paused}
            isMiniPlayer={isMiniPlayer}
            cursorBusy={cursorBusy}
            seeking={seeking}
            isSwiping={isSwiping}
            timeRange={timeRange}
            topLeftSection={topLeftSection}
            topRightSection={topRightSection}
            bottomLeftSection={bottomLeftSection}
            bottomRightSection={bottomRightSection}
        />
    )
}

export function VideoCorePlayButton() {
    const paused = useAtomValue(vc_paused)
    const action = useSetAtom(vc_dispatchAction)
    const isMobile = useAtomValue(vc_isMobile)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)

    return (
        <MediaCorePlayButton
            paused={paused}
            onTogglePlay={() => action({ type: "togglePlay" })}
            isMobile={isMobile}
            isMiniPlayer={isMiniPlayer}
        />
    )
}

export function VideoCoreVolumeButton() {
    const volume = useAtomValue(vc_volume)
    const muted = useAtomValue(vc_isMuted)
    const setVolume = useSetAtom(vc_storedVolumeAtom)
    const setMuted = useSetAtom(vc_storedMutedAtom)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const isMobile = useAtomValue(vc_isMobile)

    const onVolumeChange = React.useCallback((vol: number) => {
        setVolume(vol)
        setMuted(vol === 0)
    }, [setVolume, setMuted])

    const onMuteToggle = React.useCallback(() => {
        setMuted(p => {
            if (p && volume === 0) setVolume(0.1)
            return !p
        })
    }, [volume, setVolume, setMuted])

    return (
        <MediaCoreVolumeButton
            volume={volume}
            muted={muted}
            onVolumeChange={onVolumeChange}
            onMuteToggle={onMuteToggle}
            isMobile={isMobile}
            isMiniPlayer={isMiniPlayer}
        />
    )
}

export function VideoCoreNextButton({ onClick }: { onClick: () => void }) {
    const isMobile = useAtomValue(vc_isMobile)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)

    return (
        <MediaCoreNextButton
            onClick={onClick}
            isMobile={isMobile}
            isMiniPlayer={isMiniPlayer}
        />
    )
}

export function VideoCorePreviousButton({ onClick }: { onClick: () => void }) {
    const isMobile = useAtomValue(vc_isMobile)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)

    return (
        <MediaCorePreviousButton
            onClick={onClick}
            isMobile={isMobile}
            isMiniPlayer={isMiniPlayer}
        />
    )
}

export const vc_timestampType = atom(
    (get) => get(mediaCorePreferencesAtom).timestampMode,
    (get, set, newValue: "elapsed" | "remaining" | ((prev: "elapsed" | "remaining") => "elapsed" | "remaining")) => {
        const current = get(mediaCorePreferencesAtom)
        const next = typeof newValue === "function" ? newValue(current.timestampMode) : newValue
        set(mediaCorePreferencesAtom, { ...current, timestampMode: next })
    }
)

export function VideoCoreTimestamp() {
    const duration = useAtomValue(vc_duration)
    const currentTime = useAtomValue(vc_currentTime)
    const [type, setType] = useAtom(vc_timestampType)
    const isMobile = useAtomValue(vc_isMobile)

    const onTimestampModeToggle = React.useCallback(() => {
        setType((p: "elapsed" | "remaining") => p === "elapsed" ? "remaining" : "elapsed")
    }, [setType])

    return (
        <MediaCoreTimestamp
            currentTime={currentTime}
            duration={duration}
            timestampMode={type}
            onTimestampModeToggle={onTimestampModeToggle}
            isMobile={isMobile}
        />
    )
}

type VideoCoreControlButtonProps = {
    icons: [string, React.ElementType][]
    state: string
    className?: string
    iconClass?: string
    onClick: () => void
    onWheel?: (e: React.WheelEvent<HTMLButtonElement>) => void
    children?: React.ReactNode
}

export function VideoCoreControlButtonIcon(props: VideoCoreControlButtonProps) {
    const isMobile = useAtomValue(vc_isMobile)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)

    return (
        <MediaCoreControlButtonIcon
            {...props}
            isMobile={isMobile}
            isMiniPlayer={isMiniPlayer}
        />
    )
}

export function VideoCorePipButton() {
    const pipManager = useAtomValue(vc_pipManager)
    const isPip = useAtomValue(vc_pip)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const isMobile = useAtomValue(vc_isMobile)

    const onTogglePip = React.useCallback(() => {
        pipManager?.togglePip()
    }, [pipManager])

    return (
        <MediaCorePipButton
            isPip={isPip}
            onTogglePip={onTogglePip}
            isMobile={isMobile}
            isMiniPlayer={isMiniPlayer}
        />
    )
}

export function VideoCoreFullscreenButton() {
    const fullscreenManager = useAtomValue(vc_fullscreenManager)
    const isFullscreen = useAtomValue(vc_isFullscreen)
    const [, setMiniPlayer] = useAtom(vc_miniPlayer)
    const isMobile = useAtomValue(vc_isMobile)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)

    const onToggleFullscreen = React.useCallback(() => {
        setMiniPlayer(false)
        fullscreenManager?.toggleFullscreen()
    }, [fullscreenManager, setMiniPlayer])

    return (
        <MediaCoreFullscreenButton
            isFullscreen={isFullscreen}
            onToggleFullscreen={onToggleFullscreen}
            isMobile={isMobile}
            isMiniPlayer={isMiniPlayer}
        />
    )
}
