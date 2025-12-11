import { vc_containerElement, vc_isFullscreen, vc_miniPlayer } from "@/app/(main)/_features/video-core/video-core"
import { VideoCoreControlButtonIcon } from "@/app/(main)/_features/video-core/video-core-control-bar"
import { vc_hlsCurrentQuality, vc_hlsQualityLevels, vc_hlsSetQuality } from "@/app/(main)/_features/video-core/video-core-hls"
import { VideoCoreMenu, VideoCoreMenuBody, VideoCoreMenuTitle, VideoCoreSettingSelect } from "@/app/(main)/_features/video-core/video-core-menu"
import { VideoCoreLifecycleState, VideoCore_VideoSource } from "@/app/(main)/_features/video-core/video-core.atoms"
import { atom, useAtomValue } from "jotai"
import React from "react"
import { LuFilm } from "react-icons/lu"

export const vc_videoSources = atom<VideoCore_VideoSource[]>([])

export function VideoCoreResolutionMenu({ state, onVideoSourceChange }: {
    state: VideoCoreLifecycleState,
    onVideoSourceChange: ((source: VideoCore_VideoSource) => void) | undefined
}) {
    const isFullscreen = useAtomValue(vc_isFullscreen)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const containerElement = useAtomValue(vc_containerElement)

    // Streams
    const videoSources = state.playbackInfo?.videoSources ?? []

    // HLS
    const hlsQualityLevels = useAtomValue(vc_hlsQualityLevels)
    const hlsCurrentQuality = useAtomValue(vc_hlsCurrentQuality)
    const hlsSetQuality = useAtomValue(vc_hlsSetQuality)

    const isHls = !videoSources?.length

    const levels = React.useMemo<VideoCore_VideoSource[]>(() => {
        // Use HLS levels if no video sources are provided
        if (!videoSources?.length) {
            return hlsQualityLevels?.map(level => ({
                resolution: level.name,
                index: level.index,
                label: `${Math.round(level.bitrate / 1000)}kbps`,
            }))
        }

        return videoSources
    }, [hlsQualityLevels, videoSources])

    if (isMiniPlayer || !levels?.length) return null

    return (
        <VideoCoreMenu
            name="Video"
            trigger={<VideoCoreControlButtonIcon
                icons={[
                    ["default", LuFilm],
                ]}
                state="default"
                onClick={() => {}}
                className="text-xl lg:text-2xl"
            />}
        >
            <VideoCoreMenuTitle>Quality</VideoCoreMenuTitle>
            <VideoCoreMenuBody>
                <VideoCoreSettingSelect
                    isFullscreen={isFullscreen}
                    containerElement={containerElement}
                    options={[
                        ...(isHls ? [{
                            label: "Auto",
                            value: -1,
                        }] : []),
                        ...levels.map(level => ({
                            label: level.resolution,
                            value: level.index,
                            moreInfo: level.label,
                        })).toReversed(),
                    ]}
                    onValueChange={(value: number) => {
                        if (isHls) {
                            hlsSetQuality?.(value)
                        } else {
                            onVideoSourceChange?.(videoSources.find(source => source.index === value)!)
                        }
                    }}
                    value={isHls ? hlsCurrentQuality : state.playbackInfo?.selectedVideoSource}
                />
            </VideoCoreMenuBody>
        </VideoCoreMenu>
    )
}
