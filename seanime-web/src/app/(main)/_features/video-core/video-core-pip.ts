import { VideoCoreSubtitleManager } from "@/app/(main)/_features/video-core/video-core-subtitles"
import { logger } from "@/lib/helpers/debug"
import { atom } from "jotai"

const log = logger("VIDEO CORE PIP")

export const vc_pipElement = atom<HTMLVideoElement | null>(null)
export const vc_pipManager = atom<VideoCorePipManager | null>(null)

export class VideoCorePipManager {
    private video: HTMLVideoElement | null = null
    private subtitleManager: VideoCoreSubtitleManager | null = null
    private controller = new AbortController()
    private canvasController: AbortController | null = null
    private readonly onPipElementChange: (element: HTMLVideoElement | null) => void

    constructor(onPipElementChange: (element: HTMLVideoElement | null) => void) {
        this.onPipElementChange = onPipElementChange
        document.addEventListener("enterpictureinpicture", this.handleEnterPip, {
            signal: this.controller.signal,
        })
        document.addEventListener("leavepictureinpicture", this.handleLeavePip, {
            signal: this.controller.signal,
        })
        window.addEventListener("visibilitychange", () => {
            const shouldAutoPip = document.visibilityState !== "visible" &&
                this.video &&
                !this.video.paused

            if (shouldAutoPip) {
                this.togglePip(true)
            }
        }, { signal: this.controller.signal })
    }

    setVideo(video: HTMLVideoElement) {
        this.video = video
    }

    setSubtitleManager(subtitleManager: VideoCoreSubtitleManager) {
        this.subtitleManager = subtitleManager
    }

    togglePip(enable?: boolean) {
        const isCurrentlyInPip = document.pictureInPictureElement !== null
        const shouldEnable = enable !== undefined ? enable : !isCurrentlyInPip

        if (shouldEnable) {
            this.enterPip()
        } else {
            this.exitPip()
        }
    }

    exitPip() {
        if (document.pictureInPictureElement) {
            document.exitPictureInPicture().catch(err => {
                log.error("Failed to exit PiP", err)
            })
        }
    }

    async enterPip() {
        if (document.pictureInPictureElement || !this.video) {
            log.warning("PiP already in use or video not set")
            return
        }

        try {
            const hasActiveSubtitles = this.subtitleManager?.getSelectedTrack?.() !== null
            if (!hasActiveSubtitles) {
                log.info("Entering PiP without subtitles")
                await this.video.requestPictureInPicture()
                return
            }

            log.info("Entering PiP with subtitle burning")
            await this.enterPipWithSubtitles()
        }
        catch (error) {
            log.error("Failed to enter PiP", error)
        }
    }

    destroy() {
        this.exitPip()
        this.canvasController?.abort()
        this.controller.abort()
        this.video = null
        this.subtitleManager = null
    }

    private handleEnterPip = () => {
        const pipElement = document.pictureInPictureElement as HTMLVideoElement | null
        log.info("Entered PiP", pipElement)
        this.onPipElementChange(pipElement)
    }

    private handleLeavePip = () => {
        log.info("Left PiP")
        this.onPipElementChange(null)

        if (this.video) {
            this.video.focus()
        }
    }

    private newPipVideo() {
        const element = document.createElement("video")
        element.addEventListener("enterpictureinpicture", this.handleEnterPip, {
            signal: this.controller.signal,
        })
        element.addEventListener("leavepictureinpicture", this.handleLeavePip, {
            signal: this.controller.signal,
        })
        return element
    }

    private async enterPipWithSubtitles() {
        if (!this.video || !this.subtitleManager) return

        const canvas = document.createElement("canvas")
        const context = canvas.getContext("2d")
        if (!context) {
            log.error("Failed to get canvas context")
            return
        }

        const pipVideo = this.newPipVideo()
        pipVideo.srcObject = canvas.captureStream()
        pipVideo.muted = true

        canvas.width = this.video.videoWidth
        canvas.height = this.video.videoHeight

        if (this.subtitleManager?.libassRenderer) {
            this.subtitleManager.libassRenderer.resize(this.video.videoWidth, this.video.videoHeight)
        }

        this.canvasController = new AbortController()
        let animationFrame: number

        // draw initial frame
        context.drawImage(this.video, 0, 0)
        const subtitleCanvas = this.subtitleManager?.libassRenderer?._canvas
        if (subtitleCanvas && canvas.width && canvas.height) {
            context.drawImage(subtitleCanvas, 0, 0, canvas.width, canvas.height)
        }

        // wait for metadata
        await new Promise<void>((resolve, reject) => {
            const timeout = setTimeout(() => {
                reject(new Error("Timeout waiting for PiP video metadata"))
            }, 5000)

            pipVideo.addEventListener("loadedmetadata", () => {
                clearTimeout(timeout)
                resolve()
            }, { once: true })

            pipVideo.addEventListener("error", () => {
                clearTimeout(timeout)
                reject(new Error("Error loading PiP video metadata"))
            }, { once: true })
        })

        const renderFrame = (now?: number, metadata?: VideoFrameCallbackMetadata) => {
            if (!this.video || !context) return

            // sync play/pause state
            if (now !== undefined) {
                if (this.video.paused && !pipVideo.paused) {
                    pipVideo.pause()
                } else if (!this.video.paused && pipVideo.paused) {
                    pipVideo.play().catch()
                }
            }

            context.drawImage(this.video, 0, 0)
            const subtitleCanvas = this.subtitleManager?.libassRenderer?._canvas
            if (subtitleCanvas && canvas.width && canvas.height) {
                context.drawImage(subtitleCanvas, 0, 0, canvas.width, canvas.height)
            }
            animationFrame = this.video.requestVideoFrameCallback(renderFrame)
        }

        const cleanup = () => {
            if (this.subtitleManager?.libassRenderer) {
                this.subtitleManager.libassRenderer.resize()
            }
            if (animationFrame && this.video) {
                this.video.cancelVideoFrameCallback(animationFrame)
            }
            canvas.remove()
            pipVideo.remove()
        }

        this.canvasController.signal.addEventListener("abort", cleanup)
        this.controller.signal.addEventListener("abort", () => {
            this.canvasController?.abort()
        })
        pipVideo.addEventListener("leavepictureinpicture", () => {
            this.canvasController?.abort()
        }, { signal: this.canvasController.signal })

        try {
            // start the continuous rendering loop
            renderFrame(performance.now())

            // always start the canvas stream
            try {
                await pipVideo.play()
                if (this.video.paused) {
                    pipVideo.pause()
                }
            }
            catch (playError) {
                if (playError instanceof DOMException && playError.name === "AbortError") {
                } else {
                    throw playError
                }
            }

            const pipWindow = await pipVideo.requestPictureInPicture()

            pipWindow.addEventListener("resize", () => {
                const { width, height } = pipWindow
                if (isNaN(width) || isNaN(height) || !isFinite(width) || !isFinite(height)) {
                    return
                }
                this.subtitleManager?.libassRenderer?.resize(width, height)
            }, { signal: this.canvasController.signal })

            log.info("Successfully entered PiP")
        }
        catch (error) {
            log.error("Failed to enter PiP", error)
            this.canvasController?.abort()
            throw error
        }
    }
}
