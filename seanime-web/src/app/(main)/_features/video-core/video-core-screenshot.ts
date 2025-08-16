import { useAtomValue, useSetAtom } from "jotai"
import React from "react"
import { vc_subtitleManager, vc_videoElement } from "./video-core"
import { vc_doFlashAction } from "./video-core-action-display"
import { vc_anime4kCanvas, vc_anime4kOption } from "./video-core-anime-4k"

export function useVideoCoreScreenshot() {

    const videoElement = useAtomValue(vc_videoElement)
    const subtitleManager = useAtomValue(vc_subtitleManager)
    const flashAction = useSetAtom(vc_doFlashAction)
    const anime4kCanvas = useAtomValue(vc_anime4kCanvas)
    const anime4kOption = useAtomValue(vc_anime4kOption)

    const screenshotTimeout = React.useRef<NodeJS.Timeout | null>(null)

    async function saveToClipboard(blob: Blob, isAnime4K: boolean = false) {
        await navigator.clipboard.write([new ClipboardItem({ [blob.type]: blob })])
        flashAction({ message: "Screenshot saved to clipboard", type: "message" })
    }

    async function addSubtitles(canvas: HTMLCanvasElement): Promise<void> {
        const libassRenderer = subtitleManager?.libassRenderer
        if (!libassRenderer) return

        const ctx = canvas.getContext("2d")
        if (!ctx) return

        return new Promise((resolve) => {
            libassRenderer.resize(canvas.width, canvas.height)
            screenshotTimeout.current = setTimeout(() => {
                ctx.drawImage(libassRenderer._canvas, 0, 0, canvas.width, canvas.height)
                libassRenderer.resize(0, 0, 0, 0)
                resolve()
            }, 300)
        })
    }

    async function createVideoCanvas(source: HTMLVideoElement | HTMLCanvasElement): Promise<Blob | null> {
        return new Promise(async (resolve) => {
            if (source instanceof HTMLCanvasElement) {
                source.toBlob(resolve, "image/png")
            } else {
                const canvas = document.createElement("canvas")
                const ctx = canvas.getContext("2d")
                if (!ctx) return resolve(null)

                canvas.width = source.videoWidth
                canvas.height = source.videoHeight
                ctx.drawImage(source, 0, 0)

                await addSubtitles(canvas)
                canvas.toBlob((blob) => {
                    canvas.remove()
                    resolve(blob)
                })
            }
        })
    }

    async function createEnhancedCanvas(anime4kBlob: Blob): Promise<Blob | null> {
        return new Promise((resolve) => {
            const img = new Image()
            img.onload = async () => {
                const canvas = document.createElement("canvas")
                const ctx = canvas.getContext("2d")
                if (!ctx) return resolve(null)

                canvas.width = img.width
                canvas.height = img.height
                ctx.drawImage(img, 0, 0)

                await addSubtitles(canvas)
                canvas.toBlob((blob) => {
                    canvas.remove()
                    URL.revokeObjectURL(img.src)
                    resolve(blob)
                })
            }
            img.src = URL.createObjectURL(anime4kBlob)
        })
    }

    async function takeScreenshot() {
        if (screenshotTimeout.current) {
            clearTimeout(screenshotTimeout.current)
        }

        if (!videoElement) return

        const isPaused = videoElement.paused

        videoElement.pause()
        flashAction({ message: "Taking screenshot..." })

        try {
            let blob: Blob | null = null
            let isAnime4K = false

            if (anime4kOption !== "off" && anime4kCanvas) {
                const anime4kBlob = await createVideoCanvas(anime4kCanvas)
                if (anime4kBlob) {
                    if (subtitleManager?.libassRenderer) {
                        blob = await createEnhancedCanvas(anime4kBlob)
                    } else {
                        blob = anime4kBlob
                    }
                    isAnime4K = true
                }
            }

            if (!blob) {
                blob = await createVideoCanvas(videoElement)
            }

            if (blob) {
                await saveToClipboard(blob, isAnime4K)
            }

        }
        catch (error) {
            console.error("Screenshot failed:", error)
            flashAction({ message: "Screenshot failed" })
        }
        finally {
            if (!isPaused) {
                videoElement.play()
            }
        }
    }

    return {
        takeScreenshot,
    }
}
