import { vc_hoveringControlBar } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_isFullscreen } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_paused } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_miniPlayer } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_busy } from "@/app/(main)/_features/video-core/video-core-atoms"
import { VideoCoreLifecycleState } from "@/app/(main)/_features/video-core/video-core.atoms"
import { useRouter } from "@/lib/navigation.ts"
import { useThemeSettings } from "@/lib/theme/theme-hooks.ts"
import { useAtomValue } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"
import { startVideoCoreMiniPlayerTransition } from "./video-core"
import { vc_fullscreenManager } from "./video-core-fullscreen"
import { MediaCoreTopSectionView, MediaCoreTopPlaybackInfoView } from "@/app/(main)/_features/media-core/media-core-playback-info"

export function VideoCoreTopSection(props: { children?: React.ReactNode, inline?: boolean }) {
    const { children, inline } = props

    const busy = useAtomValue(vc_busy)
    const paused = useAtomValue(vc_paused)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const hoveringControlBar = useAtomValue(vc_hoveringControlBar)
    const fullscreen = useAtomValue(vc_isFullscreen)

    const showTopSection = busy || paused || hoveringControlBar

    return (
        <MediaCoreTopSectionView
            inline={inline}
            fullscreen={fullscreen}
            isMiniPlayer={isMiniPlayer}
            showTopSection={showTopSection}
            paused={paused}
        >
            {children}
        </MediaCoreTopSectionView>
    )
}

export function VideoCoreTopPlaybackInfo(props: { state: VideoCoreLifecycleState, children?: React.ReactNode }) {
    const { state } = props

    const ts = useThemeSettings()
    const paused = useAtomValue(vc_paused)
    const [isMiniPlayer, setMiniPlayer] = useAtom(vc_miniPlayer)
    const hoveringControlBar = useAtomValue(vc_hoveringControlBar)
    const fullscreenManager = useAtomValue(vc_fullscreenManager)

    const router = useRouter()

    const displayTitle = state.playbackInfo?.episode?.displayTitle
    const _episodeTitle = state.playbackInfo?.episode?.episodeTitle
    const episodeTitle = (displayTitle !== _episodeTitle && (!ts.hideAnimeSpoilers || (ts.hideAnimeSpoilers && !ts.hideAnimeSpoilerTitles)))
        ? _episodeTitle
        : ""

    const onAnimeTitleClick = React.useCallback(() => {
        if (state.playbackInfo?.episode?.baseAnime?.id) {
            router.push(`/entry?id=${state.playbackInfo?.episode?.baseAnime?.id}`)
            fullscreenManager?.exitFullscreen()?.then(() => {
                startVideoCoreMiniPlayerTransition(() => {
                    setMiniPlayer(true)
                })
            })
        }
    }, [state.playbackInfo?.episode?.baseAnime?.id, router, fullscreenManager, setMiniPlayer])

    return (
        <MediaCoreTopPlaybackInfoView
            animeTitle={state.playbackInfo?.episode?.baseAnime?.title?.userPreferred}
            episodeDisplayTitle={displayTitle}
            episodeTitle={episodeTitle}
            onAnimeTitleClick={onAnimeTitleClick}
            isMiniPlayer={isMiniPlayer}
            paused={paused}
            hoveringControlBar={hoveringControlBar}
        />
    )
}
