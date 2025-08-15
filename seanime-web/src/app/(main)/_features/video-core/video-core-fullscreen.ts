import { logger } from "@/lib/helpers/debug"
import { atom } from "jotai"

const log = logger("VIDEO CORE FULLSCREEN")

export const vc_fullscreenManager = atom<VideoCoreFullscreenManager | null>(null)

export class VideoCoreFullscreenManager {
    private containerElement: HTMLElement | null = null
    private controller = new AbortController()
    private onFullscreenChange: (isFullscreen: boolean) => void

    constructor(onFullscreenChange: (isFullscreen: boolean) => void) {
        this.onFullscreenChange = onFullscreenChange
        this.attachDocumentListeners()
    }

    setContainer(containerElement: HTMLElement) {
        this.containerElement = containerElement
    }

    isFullscreen(): boolean {
        return !!(
            document.fullscreenElement ||
            (document as any).webkitFullscreenElement ||
            (document as any).mozFullScreenElement ||
            (document as any).msFullscreenElement
        )
    }

    async toggleFullscreen() {
        if (this.isFullscreen()) {
            await this.exitFullscreen()
        } else {
            await this.enterFullscreen()
        }
    }

    async exitFullscreen() {
        try {
            if (document.exitFullscreen) {
                await document.exitFullscreen()
            } else if ((document as any).webkitExitFullscreen) {
                await (document as any).webkitExitFullscreen()
            } else if ((document as any).mozCancelFullScreen) {
                await (document as any).mozCancelFullScreen()
            } else if ((document as any).msExitFullscreen) {
                await (document as any).msExitFullscreen()
            }
            log.info("Exited fullscreen")
        }
        catch (error) {
            log.error("Failed to exit fullscreen", error)
        }
    }

    async enterFullscreen() {
        if (!this.containerElement) {
            log.warning("Container element not set")
            return
        }

        try {
            if (this.containerElement.requestFullscreen) {
                await this.containerElement.requestFullscreen()
            } else if ((this.containerElement as any).webkitRequestFullscreen) {
                await (this.containerElement as any).webkitRequestFullscreen()
            } else if ((this.containerElement as any).mozRequestFullScreen) {
                await (this.containerElement as any).mozRequestFullScreen()
            } else if ((this.containerElement as any).msRequestFullscreen) {
                await (this.containerElement as any).msRequestFullscreen()
            }
            log.info("Entered fullscreen")
        }
        catch (error) {
            log.error("Failed to enter fullscreen", error)
        }
    }

    destroy() {
        this.controller.abort()
        this.containerElement = null
    }

    private attachDocumentListeners() {
        document.addEventListener("fullscreenchange", this.handleFullscreenChange, {
            signal: this.controller.signal,
        })
        document.addEventListener("webkitfullscreenchange", this.handleFullscreenChange, {
            signal: this.controller.signal,
        })
        document.addEventListener("mozfullscreenchange", this.handleFullscreenChange, {
            signal: this.controller.signal,
        })
        document.addEventListener("msfullscreenchange", this.handleFullscreenChange, {
            signal: this.controller.signal,
        })
    }

    private handleFullscreenChange = () => {
        const isFullscreen = this.isFullscreen()
        log.info("Fullscreen state changed:", isFullscreen)
        this.onFullscreenChange(isFullscreen)
    }
}
