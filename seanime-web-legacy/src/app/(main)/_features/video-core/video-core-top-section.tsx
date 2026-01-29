import { vc_busy, vc_isFullscreen, vc_miniPlayer, vc_paused } from "@/app/(main)/_features/video-core/video-core"
import { vc_hoveringControlBar } from "@/app/(main)/_features/video-core/video-core-control-bar"
import { VideoCoreLifecycleState } from "@/app/(main)/_features/video-core/video-core.atoms"
import { cn } from "@/components/ui/core/styling"
import { __isDesktop__ } from "@/types/constants"
import { useAtomValue } from "jotai"
import React from "react"

export function VideoCoreTopSection(props: { children?: React.ReactNode, inline?: boolean }) {
    const { children, inline, ...rest } = props

    const busy = useAtomValue(vc_busy)
    const paused = useAtomValue(vc_paused)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const hoveringControlBar = useAtomValue(vc_hoveringControlBar)
    const fullscreen = useAtomValue(vc_isFullscreen)

    return (
        <>
            <div
                data-vc-element="control-bar-top-section"
                className={cn(
                    "top-0 absolute left-0 w-full py-4 px-5 duration-200 transition-opacity opacity-0 z-[999]",
                    (__isDesktop__ && ((inline && fullscreen) || !inline)) && "top-8",
                    (busy || paused || hoveringControlBar) && "opacity-100",
                    isMiniPlayer && "top-0",
                )}
            >
                {children}
            </div>

            <div
                data-vc-element="control-bar-top-gradient"
                className={cn(
                    "pointer-events-none",
                    "absolute top-0 left-0 right-0 w-full z-[5] transition-opacity duration-300 opacity-0",
                    "bg-gradient-to-b from-black/60 to-transparent",
                    "h-20",
                    (isMiniPlayer && paused) && "opacity-100",
                )}
            />
        </>
    )
}

export function VideoCoreTopPlaybackInfo(props: { state: VideoCoreLifecycleState, children?: React.ReactNode }) {
    const { state, children, ...rest } = props

    const busy = useAtomValue(vc_busy)
    const paused = useAtomValue(vc_paused)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const hoveringControlBar = useAtomValue(vc_hoveringControlBar)

    if (isMiniPlayer) return null

    return (
        <>
            <div
                data-vc-element="top-playback-info"
                className={cn(
                    "transition-opacity duration-200 opacity-0",
                    (paused || hoveringControlBar) && "opacity-100",
                )}
            >
                {state.playbackInfo?.episode?.baseAnime?.title?.userPreferred &&
                    <p data-vc-element="top-playback-info-title" className="text-white/50 font-medium text-sm max-w-[400px] line-clamp-1">
                        {state.playbackInfo?.episode?.baseAnime?.title?.userPreferred}
                    </p>}
                <div className="flex flex-row gap-2" data-vc-element="top-playback-info-episode">
                    <p className="text-white font-bold text-base">
                        {state.playbackInfo?.episode?.displayTitle}
                    </p>
                    {state.playbackInfo?.episode?.episodeTitle && <p className="text-white/50 text-base !font-normal max-w-[400px] line-clamp-1">
                        {state.playbackInfo?.episode?.episodeTitle}
                    </p>}
                </div>
            </div>
        </>
    )
}
