import { vc_realVideoSize, vc_seeking, vc_videoElement } from "@/app/(main)/_features/video-core/video-core"
import { ModeA, render } from "anime4k-webgpu"
import { atom, useAtomValue } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"
import { useUpdateEffect } from "react-use"

// maintain a single canvas
// destroy when window resizes
const canvasAtom = atom<HTMLCanvasElement | null>(null)

export const useVideoCoreAnime4K = () => {
    const video = useAtomValue(vc_videoElement)
    const realVideoSize = useAtomValue(vc_realVideoSize)
    const seeking = useAtomValue(vc_seeking)
    const [canvas, setCanvas] = useAtom(canvasAtom)

    const initialized = React.useRef(false)

    function destroyCanvas() {
        if (canvas) {
            canvas.remove()
            setCanvas(null)
        }
    }

    // throttle canvas creation
    const timeoutRef = React.useRef<NodeJS.Timeout | null>(null)

    React.useEffect(() => {
        // if already initialized, destroy the canvas
        if (initialized || seeking) {
            destroyCanvas()
        }
        if (timeoutRef.current) {
            clearTimeout(timeoutRef.current)
        }

        if (!video || seeking) return

        timeoutRef.current = setTimeout(() => {
            if (!video) return
            const canvas = document.createElement("canvas")
            canvas.width = realVideoSize.width
            canvas.height = realVideoSize.height
            canvas.style.objectFit = "cover"
            canvas.style.position = "absolute"
            canvas.style.top = video.getBoundingClientRect().top + "px"
            canvas.style.left = "0"
            canvas.style.right = "0"
            canvas.style.bottom = "0"
            canvas.style.pointerEvents = "none"
            canvas.style.zIndex = "2"
            setCanvas(canvas)
            video.parentElement?.appendChild(canvas)
            initialized.current = true
        }, 1000)

    }, [video, realVideoSize, seeking])

    useUpdateEffect(() => {
        if (!canvas || !video) return

        async function init() {
            await render({
                video: video!,
                canvas: canvas!,
                pipelineBuilder: (device, inputTexture) => {
                    // const restore = new GANx3L({
                    //     device,
                    //     inputTexture,
                    // });
                    // return [restore] as any
                    // const upscale = new CNNx2M({
                    //     device,
                    //     inputTexture,
                    // });
                    const preset = new ModeA({
                        device,
                        inputTexture,
                        nativeDimensions: {
                            width: video!.videoWidth,
                            height: video!.videoHeight,
                        },
                        targetDimensions: {
                            width: canvas!.width,
                            height: canvas!.height,
                        },
                    })
                    return [preset] as any
                },
            })
        }

        init()
    }, [canvas])

}


