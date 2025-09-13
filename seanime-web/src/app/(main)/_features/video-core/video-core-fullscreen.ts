import { logger } from "@/lib/helpers/debug"
import { atom } from "jotai"

const log = logger("VIDEO CORE FULLSCREEN")

export const vc_fullscreenManager = atom<VideoCoreFullscreenManager | null>(null)

export class VideoCoreFullscreenManager {
    private containerElement: HTMLElement | null = null
    private controller = new AbortController()
    private onFullscreenChange: (isFullscreen: boolean) => void
    private isElectronNativeFullscreen = false

    constructor(onFullscreenChange: (isFullscreen: boolean) => void) {
        this.onFullscreenChange = onFullscreenChange
        this.attachDocumentListeners()
        this.attachElectronListeners()
    }

    setContainer(containerElement: HTMLElement) {
        this.containerElement = containerElement
    }

    isFullscreen(): boolean {
        // Check Electron native fullscreen first
        if (this.isElectron() && this.isElectronNativeFullscreen) {
            return true
        }

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
            if (this.isElectron() && this.shouldUseElectronFullscreen()) {
                await this.exitElectronFullscreen()
                return
            }

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
            if (this.isElectron() && this.shouldUseElectronFullscreen()) {
                await this.enterElectronFullscreen()
                return
            }

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

    private isElectron(): boolean {
        return !!(window as any)?.electron
    }

    private shouldUseElectronFullscreen(): boolean {
        // return this.isElectron() && window.electron?.platform === "win32"
        return this.isElectron()
    }

    private async enterElectronFullscreen(): Promise<void> {
        if (!(window as any)?.electron?.window?.setFullscreen) {
            log.warning("Electron fullscreen API not available")
            return
        }

        try {
            await (window as any).electron.window.setFullscreen(true)
            this.isElectronNativeFullscreen = true
            log.info("Entered Electron native fullscreen")
        }
        catch (error) {
            log.error("Failed to enter Electron fullscreen", error)
        }
    }

    private async exitElectronFullscreen(): Promise<void> {
        if (!window.electron?.window?.setFullscreen) {
            log.warning("Electron fullscreen API not available")
            return
        }

        try {
            await window.electron?.window?.setFullscreen(false)
            this.isElectronNativeFullscreen = false
            log.info("Exited Electron native fullscreen")
        }
        catch (error) {
            log.error("Failed to exit Electron fullscreen", error)
        }
    }

    private attachElectronListeners() {
        if (!this.isElectron()) return

        const removeFullscreenListener = window.electron?.on?.("window:fullscreen", (isFullscreen: boolean) => {
            this.isElectronNativeFullscreen = isFullscreen
            log.info("Electron fullscreen state changed:", isFullscreen)
            this.onFullscreenChange(isFullscreen)
        })

        if (removeFullscreenListener) {
            const originalAbort = this.controller.abort.bind(this.controller)
            this.controller.abort = () => {
                removeFullscreenListener()
                originalAbort()
            }
        }
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
