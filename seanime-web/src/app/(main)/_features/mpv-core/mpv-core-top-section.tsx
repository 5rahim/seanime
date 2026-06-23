import { Player_PlaybackInfo } from "@/api/generated/types"
import { MediaCoreTopPlaybackInfoView } from "@/app/(main)/_features/media-core/media-core-playback-info"
import { startVideoCoreMiniPlayerTransition } from "@/app/(main)/_features/video-core/video-core"
import { useRouter } from "@/lib/navigation.ts"
import { useThemeSettings } from "@/lib/theme/theme-hooks.ts"
import React from "react"

export function MpvCoreTopPlaybackInfo(props: {
    playbackInfo: Player_PlaybackInfo | null
    isMiniPlayer: boolean
    paused: boolean
    hoveringControlBar: boolean
    toggleFullscreen: (force?: boolean) => Promise<void>
    setMiniPlayer: (value: boolean) => void
}) {
    const { playbackInfo, isMiniPlayer, paused, hoveringControlBar, toggleFullscreen, setMiniPlayer } = props
    const router = useRouter()
    const ts = useThemeSettings()

    const displayTitle = playbackInfo?.episode?.displayTitle || playbackInfo?.localFile?.name || "MpvCore"
    const _episodeTitle = playbackInfo?.episode?.episodeTitle
    const episodeTitle = (_episodeTitle && displayTitle !== _episodeTitle && (!ts.hideAnimeSpoilers || (ts.hideAnimeSpoilers && !ts.hideAnimeSpoilerTitles)))
        ? _episodeTitle
        : ""

    const animeId = playbackInfo?.media?.id || playbackInfo?.episode?.baseAnime?.id
    const onAnimeTitleClick = animeId ? async () => {
        router.push(`/entry?id=${animeId}`)
        await toggleFullscreen(false)
        startVideoCoreMiniPlayerTransition(() => {
            setMiniPlayer(true)
        })
    } : undefined

    return (
        <MediaCoreTopPlaybackInfoView
            animeTitle={playbackInfo?.media?.title?.userPreferred || playbackInfo?.episode?.baseAnime?.title?.userPreferred}
            episodeDisplayTitle={displayTitle}
            episodeTitle={episodeTitle}
            onAnimeTitleClick={onAnimeTitleClick}
            isMiniPlayer={isMiniPlayer}
            paused={paused}
            hoveringControlBar={hoveringControlBar}
        />
    )
}
