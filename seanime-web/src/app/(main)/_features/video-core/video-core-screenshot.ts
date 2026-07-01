import { useVideoCoreSaveScreenshot } from "@/api/hooks/videocore.hooks"
import { vc_subtitleManager } from "@/app/(main)/_features/video-core/video-core"
import { vc_anime4kManager } from "@/app/(main)/_features/video-core/video-core"
import { vc_videoElement } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_showOverlayFeedback } from "@/app/(main)/_features/video-core/video-core-overlay-display"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { atom, useAtomValue, useSetAtom } from "jotai"
import React from "react"
import { toast } from "sonner"
import { vc_anime4kOption } from "./video-core-anime-4k"

export const vc_screenshotPromptOpenAtom = atom(false)
export const vc_pendingScreenshotAtom = atom<{ blob: Blob; isAnime4K: boolean } | null>(null)

export function useVideoCoreScreenshot() {

    const videoElement = useAtomValue(vc_videoElement)
    const subtitleManager = useAtomValue(vc_subtitleManager)
    const showOverlayFeedback = useSetAtom(vc_showOverlayFeedback)
    const anime4kManager = useAtomValue(vc_anime4kManager)
    const anime4kOption = useAtomValue(vc_anime4kOption)

    const serverStatus = useServerStatus()
    const { mutateAsync: saveScreenshotMutation } = useVideoCoreSaveScreenshot()

    const setPromptOpen = useSetAtom(vc_screenshotPromptOpenAtom)
    const setPendingScreenshot = useSetAtom(vc_pendingScreenshotAtom)

    const screenshotTimeout = React.useRef<NodeJS.Timeout | null>(null)

    const blobToBase64 = (blob: Blob): Promise<string> => {
        return new Promise((resolve, reject) => {
            const reader = new FileReader()
            reader.onloadend = () => {
                const base64String = (reader.result as string).split(",")[1]
                resolve(base64String)
            }
            reader.onerror = reject
            reader.readAsDataURL(blob)
        })
    }

    async function saveScreenshot(blob: Blob, isAnime4K: boolean = false) {
        const screenshotDir = serverStatus?.settings?.mediaPlayer?.screenshotDir

        if (!screenshotDir) {
            setPendingScreenshot({ blob, isAnime4K })
            setPromptOpen(true)
            return
        }

        const filename = `seanime_screenshot_${new Date().getTime()}${isAnime4K ? "_anime4k" : ""}.png`

        try {
            const base64Data = await blobToBase64(blob)
            await saveScreenshotMutation({
                dir: screenshotDir,
                filename,
                base64Data,
            })

            try {
                await navigator.clipboard.write([new ClipboardItem({ [blob.type]: blob })])
            }
            catch (e) {
            }

            showOverlayFeedback({ message: "Screenshot saved", type: "message" })
        }
        catch (error) {
            console.error("Failed to save screenshot:", error)
            showOverlayFeedback({ message: "Screenshot failed" })
            toast.error("Failed to save screenshot to server")
        }
    }

    async function addSubtitles(canvas: HTMLCanvasElement): Promise<void> {
        const libassRenderer = subtitleManager?.libassRenderer
        if (!libassRenderer) return

        const ctx = canvas.getContext("2d")
        if (!ctx) return

        return new Promise((resolve) => {
            libassRenderer.resize(true, canvas.width, canvas.height)
            screenshotTimeout.current = setTimeout(() => {
                ctx.drawImage(libassRenderer._canvas, 0, 0, canvas.width, canvas.height)
                libassRenderer.resize(true, 0, 0)
                resolve()
            }, 300)
        })
    }

    async function createBlob(canvas: HTMLCanvasElement, type: string = "image/png"): Promise<Blob | null> {
        return new Promise((resolve) => {
            canvas.toBlob((blob) => {
                canvas.remove()
                resolve(blob)
            }, type)
        })
    }

    async function createVideoCanvas(source: HTMLVideoElement): Promise<Blob | null> {
        return new Promise(async (resolve) => {
            const canvas = document.createElement("canvas")
            const ctx = canvas.getContext("2d")
            if (!ctx) return resolve(null)

            canvas.width = source.videoWidth
            canvas.height = source.videoHeight
            ctx.drawImage(source, 0, 0)

            await addSubtitles(canvas)
            resolve(await createBlob(canvas))
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
                const blob = await createBlob(canvas)
                URL.revokeObjectURL(img.src)
                resolve(blob)
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
        showOverlayFeedback({ message: "Taking screenshot..." })

        try {
            let blob: Blob | null = null
            let isAnime4K = false

            if (anime4kOption !== "off" && anime4kManager?.canvas) {
                const anime4kBlob = await anime4kManager.captureFrame()
                if (!anime4kBlob) {
                    throw new Error("Failed to capture Anime4K frame")
                }

                if (subtitleManager?.libassRenderer) {
                    blob = await createEnhancedCanvas(anime4kBlob)
                } else {
                    blob = anime4kBlob
                }

                isAnime4K = true
            }

            if (!blob) {
                blob = await createVideoCanvas(videoElement)
            }

            if (blob) {
                await saveScreenshot(blob, isAnime4K)
            }

        }
        catch (error) {
            console.error("Screenshot failed:", error)
            showOverlayFeedback({ message: "Screenshot failed" })
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
