import { NativePlayerState } from "@/app/(main)/_features/native-player/native-player.atoms"
import { vc_busy, vc_miniPlayer, vc_paused } from "@/app/(main)/_features/video-core/video-core"
import { vc_hoveringControlBar } from "@/app/(main)/_features/video-core/video-core-control-bar"
import { cn } from "@/components/ui/core/styling"
import { useAtomValue } from "jotai"
import { motion } from "motion/react"
import React from "react"

export function VideoCoreTopSection(props: { children?: React.ReactNode }) {
    const { children, ...rest } = props

    const busy = useAtomValue(vc_busy)
    const paused = useAtomValue(vc_paused)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const hoveringControlBar = useAtomValue(vc_hoveringControlBar)

    return (
        <>
            <div
                data-vc-control-bar-top-section
                className={cn(
                    "vc-control-bar-top-section",
                    "top-8 absolute left-0 w-full py-4 px-5 duration-200 transition-opacity opacity-0 z-[999]",
                    (busy || paused || hoveringControlBar) && "opacity-100",
                    isMiniPlayer && "top-0",
                )}
            >
                {children}
            </div>

            <div
                className={cn(
                    "vc-control-bar-top-gradient pointer-events-none",
                    "absolute top-0 left-0 right-0 w-full z-[5] transition-opacity duration-300 opacity-0",
                    "bg-gradient-to-b from-black/60 to-transparent",
                    "h-20",
                    (isMiniPlayer && paused) && "opacity-100",
                )}
            />
        </>
    )
}

export function VideoCoreTopPlaybackInfo(props: { state: NativePlayerState, children?: React.ReactNode }) {
    const { state, children, ...rest } = props

    const busy = useAtomValue(vc_busy)
    const paused = useAtomValue(vc_paused)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const hoveringControlBar = useAtomValue(vc_hoveringControlBar)

    if (isMiniPlayer) return null

    return (
        <>
            <div
                className={cn(
                    "transition-opacity duration-200 opacity-0",
                    (paused || hoveringControlBar) && "opacity-100",
                )}
            >
                <p className="text-white font-bold text-lg">
                    {state.playbackInfo?.episode?.displayTitle}
                </p>
                <p className="text-white/50 text-base !font-normal">
                    {state.playbackInfo?.episode?.episodeTitle}
                </p>
            </div>
        </>
    )
}
