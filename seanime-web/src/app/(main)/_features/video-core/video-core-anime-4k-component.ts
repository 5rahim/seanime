import { vc_anime4kManager } from "@/app/(main)/_features/video-core/video-core"
import { vc_anime4kOption } from "@/app/(main)/_features/video-core/video-core-anime-4k"
import {
    vc_miniPlayer,
    vc_seeking,
    vc_videoElement,
} from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_pip } from "@/app/(main)/_features/video-core/video-core-pip"
import { useAtomValue } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"

export const VideoCoreAnime4K = () => {
    const seeking = useAtomValue(vc_seeking)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const isPip = useAtomValue(vc_pip)
    const video = useAtomValue(vc_videoElement)

    const manager = useAtomValue(vc_anime4kManager)
    const [selectedOption] = useAtom(vc_anime4kOption)

    const resizeCanvas = React.useEffectEvent(() => {
        if (!video || !manager) return

        const rect = video.getBoundingClientRect()
        if (!rect.width || !rect.height) return

        manager.resize(rect.width, rect.height)
    })

    React.useEffect(() => {
        resizeCanvas()
    }, [manager, video])

    React.useEffect(() => {
        if (video && manager) {
            manager.setOption(selectedOption, {
                isMiniPlayer,
                isPip,
                seeking,
            })
        }
    }, [video, manager, selectedOption, isMiniPlayer, isPip, seeking])

    React.useEffect(() => {
        if (!video || !manager) return

        let resizeFrame = 0

        const handleResize = () => {
            if (resizeFrame) {
                cancelAnimationFrame(resizeFrame)
            }

            resizeFrame = requestAnimationFrame(() => {
                resizeFrame = 0
                resizeCanvas()
            })
        }

        window.addEventListener("resize", handleResize)

        return () => {
            window.removeEventListener("resize", handleResize)
            if (resizeFrame) {
                cancelAnimationFrame(resizeFrame)
            }
        }
    }, [manager, video])

    return null
}
