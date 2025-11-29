import { vc_containerElement, vc_isFullscreen, vc_miniPlayer } from "@/app/(main)/_features/video-core/video-core"
import { VideoCoreControlButtonIcon } from "@/app/(main)/_features/video-core/video-core-control-bar"
import { vc_hlsCurrentQuality, vc_hlsQualityLevels, vc_hlsSetQuality } from "@/app/(main)/_features/video-core/video-core-hls"
import { VideoCoreMenu, VideoCoreMenuBody, VideoCoreMenuTitle, VideoCoreSettingSelect } from "@/app/(main)/_features/video-core/video-core-menu"
import { useAtomValue } from "jotai"
import React from "react"
import { LuFilm } from "react-icons/lu"

export function VideoCoreQualityMenu() {
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
