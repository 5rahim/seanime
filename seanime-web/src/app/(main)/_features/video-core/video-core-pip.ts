import { MediaCaptionsManager } from "@/app/(main)/_features/video-core/video-core-media-captions"
import { VideoCoreSubtitleManager } from "@/app/(main)/_features/video-core/video-core-subtitles"
import { VideoCorePlaybackInfo } from "@/app/(main)/_features/video-core/video-core.atoms"
import { logger } from "@/lib/helpers/debug"
import { atom } from "jotai"

const log = logger("VIDEO CORE PIP")

export const vc_pipElement = atom<HTMLVideoElement | null>(null)
export const vc_pipManager = atom<VideoCorePipManager | null>(null)

export class VideoCorePipManager {
    private video: HTMLVideoElement | null = null
    private subtitleManager: VideoCoreSubtitleManager | null = null
    private mediaCaptionsManager: MediaCaptionsManager | null = null
    private controller = new AbortController()
    private canvasController: AbortController | null = null
    private readonly onPipElementChange: (element: HTMLVideoElement | null) => void
    private pipProxy: HTMLVideoElement | null = null
    private isSyncingFromMain = false
    private isSyncingFromPip = false
    private playbackInfo: VideoCorePlaybackInfo | null = null

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

    setVideo(video: HTMLVideoElement, playbackInfo: VideoCorePlaybackInfo) {
        this.video = video

        if (this.video) {
            this.video.addEventListener("play", this.handleMainVideoPlay, {
                signal: this.controller.signal,
            })
            this.video.addEventListener("pause", this.handleMainVideoPause, {
                signal: this.controller.signal,
            })
        }
        this.playbackInfo = playbackInfo
    }

    setSubtitleManager(subtitleManager: VideoCoreSubtitleManager) {
        this.subtitleManager = subtitleManager
    }

    setMediaCaptionsManager(mediaCaptionsManager: MediaCaptionsManager) {
        this.mediaCaptionsManager = mediaCaptionsManager
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
            const hasLibassSubtitles = this.subtitleManager?.getSelectedTrackNumberOrNull?.() != null
            const hasMediaCaptions = this.mediaCaptionsManager?.getSelectedTrackIndexOrNull?.() != null
            const hasActiveSubtitles = hasLibassSubtitles || hasMediaCaptions

            if (!hasActiveSubtitles) {
                log.info("Entering PiP without subtitles")
                await this.video.requestPictureInPicture()
                return
            }

            log.info("Entering PiP with subtitle burning", { hasLibassSubtitles, hasMediaCaptions })
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
        this.mediaCaptionsManager = null
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
        this.pipProxy = null
    }

    private newPipVideo() {
        const element = document.createElement("video")
        element.muted = true
        element.addEventListener("enterpictureinpicture", this.handleEnterPip, {
            signal: this.controller.signal,
        })
        element.addEventListener("leavepictureinpicture", this.handleLeavePip, {
            signal: this.controller.signal,
        })
        return element
    }



    private renderToCanvas = (
        pipVideo: HTMLVideoElement,
        context: CanvasRenderingContext2D,
        animationFrameRef: { current: number },
    ) => (now?: number, metadata?: VideoFrameCallbackMetadata) => {
        if (!this.video || !context) return

        // sync play/pause state
        if (now !== undefined) {
            if (this.video.paused && !pipVideo.paused) {
                if (!this.isSyncingFromPip) {
                    pipVideo.pause()
                }
            } else if (!this.video.paused && pipVideo.paused) {
                if (!this.isSyncingFromPip) {
                    pipVideo.play().catch(() => {})
                }
            }
        }

        context.drawImage(this.video, 0, 0)

        // Draw ASS/SSA subtitles
        const subtitleCanvas = this.subtitleManager?.libassRenderer?._canvas
        if (subtitleCanvas && context.canvas.width && context.canvas.height) {
            context.drawImage(subtitleCanvas, 0, 0, context.canvas.width, context.canvas.height)
        }

        // Draw PGS subtitles
        const pgsCanvas = this.subtitleManager?.pgsRenderer?._canvas
        if (pgsCanvas && context.canvas.width && context.canvas.height) {
            context.drawImage(pgsCanvas, 0, 0, context.canvas.width, context.canvas.height)
        }

        // Draw media captions
        if (this.mediaCaptionsManager) {
            this.mediaCaptionsManager.renderToCanvas(context, context.canvas.width, context.canvas.height, this.video.currentTime)
        }

        animationFrameRef.current = this.video.requestVideoFrameCallback(this.renderToCanvas(pipVideo, context, animationFrameRef))
    }

    private handleMainVideoPlay = () => {
        if (this.isSyncingFromPip) return
        if (this.pipProxy && this.pipProxy.paused) {
            this.isSyncingFromMain = true
            this.pipProxy.play().catch(() => {})
            this.isSyncingFromMain = false
        }
    }

    private handleMainVideoPause = () => {
        if (this.isSyncingFromPip) return
        if (this.pipProxy && !this.pipProxy.paused) {
            this.isSyncingFromMain = true
            this.pipProxy.pause()
            this.isSyncingFromMain = false
        }
    }

    private async enterPipWithSubtitles() {
        if (!this.video || (!this.subtitleManager && !this.mediaCaptionsManager)) return

        const canvas = document.createElement("canvas")
        const context = canvas.getContext("2d")
        if (!context) {
            log.error("Failed to get canvas context")
            return
        }

        const pipVideo = this.newPipVideo()
        pipVideo.srcObject = canvas.captureStream()
        pipVideo.muted = true
        this.pipProxy = pipVideo

        canvas.width = this.video.videoWidth
        canvas.height = this.video.videoHeight

        if (this.subtitleManager?.libassRenderer) {
            this.subtitleManager.libassRenderer.resize(this.video.videoWidth, this.video.videoHeight)
        }

        this.canvasController = new AbortController()

        // Forward PiP overlay play/pause controls to the main video element
        // In the canvas path the PiP element is not the main <video> and PiP UI
        // controls act on this proxy element instead.
        const forwardPlay = () => {
            if (this.video && this.video.paused) {
                this.isSyncingFromPip = true
                this.video.play().catch(err => {
                    log.error("Failed to play main video from PiP overlay", err)
                }).finally(() => {
                    this.isSyncingFromPip = false
                })
            }
        }
        const forwardPause = () => {
            if (this.video && !this.video.paused) {
                this.isSyncingFromPip = true
                this.video.pause()
                this.isSyncingFromPip = false
            }
        }
        pipVideo.addEventListener("play", forwardPlay, { signal: this.canvasController.signal })
        pipVideo.addEventListener("pause", forwardPause, { signal: this.canvasController.signal })
        const animationFrameRef = { current: 0 }

        // draw initial frame
        context.drawImage(this.video, 0, 0)
        const subtitleCanvas = this.subtitleManager?.libassRenderer?._canvas
        if (subtitleCanvas && canvas.width && canvas.height) {
            context.drawImage(subtitleCanvas, 0, 0, canvas.width, canvas.height)
        }
        const pgsCanvas = this.subtitleManager?.pgsRenderer?._canvas
        if (pgsCanvas && canvas.width && canvas.height) {
            context.drawImage(pgsCanvas, 0, 0, canvas.width, canvas.height)
        }
        if (this.mediaCaptionsManager) {
            this.mediaCaptionsManager.renderToCanvas(context, canvas.width, canvas.height, this.video.currentTime)
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

        const cleanup = () => {
            if (this.subtitleManager?.libassRenderer) {
                this.subtitleManager.libassRenderer.resize()
            }
            if (animationFrameRef.current && this.video) {
                this.video.cancelVideoFrameCallback(animationFrameRef.current)
            }
            canvas.remove()
            pipVideo.remove()
            this.pipProxy = null
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
            this.renderToCanvas(pipVideo, context, animationFrameRef)(performance.now())

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
                this.subtitleManager?.pgsRenderer?.resize()
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
