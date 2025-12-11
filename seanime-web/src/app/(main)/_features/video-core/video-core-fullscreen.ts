import { logger } from "@/lib/helpers/debug"
import { isApple } from "@/lib/utils/browser-detection"
import { atom } from "jotai"

const log = logger("VIDEO CORE FULLSCREEN")

export type FullscreenManagerChangedEvent = CustomEvent<{ isFullscreen: boolean }>
export type FullscreenManagerDestroyedEvent = CustomEvent
export type FullscreenManagerAttemptEvent = CustomEvent<{ method: "enter" | "exit" }>

interface VideoCoreFullscreenManagerEventMap {
    "fullscreenchanged": FullscreenManagerChangedEvent
    "destroyed": FullscreenManagerDestroyedEvent
    "enterattempt": FullscreenManagerAttemptEvent
    "exitattempt": FullscreenManagerAttemptEvent
}

export const vc_fullscreenManager = atom<VideoCoreFullscreenManager | null>(null)

export class VideoCoreFullscreenManager extends EventTarget {
    private containerElement: HTMLElement | null = null
    private videoElement: HTMLVideoElement | null = null
    private controller = new AbortController()
    private onFullscreenChange: (isFullscreen: boolean) => void
    private isElectronNativeFullscreen = false
    private attachVideoListeners?: () => void

    constructor(onFullscreenChange: (isFullscreen: boolean) => void) {
        super()
        this.onFullscreenChange = onFullscreenChange
        this.attachDocumentListeners()
        this.attachElectronListeners()
        this.initElectronFullscreenState()
    }

    addEventListener<K extends keyof VideoCoreFullscreenManagerEventMap>(
        type: K,
        listener: (this: VideoCoreFullscreenManager, ev: VideoCoreFullscreenManagerEventMap[K]) => any,
        options?: boolean | AddEventListenerOptions,
    ): void
    addEventListener(
        type: string,
        listener: EventListenerOrEventListenerObject,
        options?: boolean | AddEventListenerOptions,
    ): void
    addEventListener(
        type: string,
        listener: EventListenerOrEventListenerObject,
        options?: boolean | AddEventListenerOptions,
    ): void {
        super.addEventListener(type, listener, options)
    }

    removeEventListener<K extends keyof VideoCoreFullscreenManagerEventMap>(
        type: K,
        listener: (this: VideoCoreFullscreenManager, ev: VideoCoreFullscreenManagerEventMap[K]) => any,
        options?: boolean | EventListenerOptions,
    ): void
    removeEventListener(
        type: string,
        listener: EventListenerOrEventListenerObject,
        options?: boolean | EventListenerOptions,
    ): void

    removeEventListener(
        type: string,
        listener: EventListenerOrEventListenerObject,
        options?: boolean | EventListenerOptions,
    ): void {
        super.removeEventListener(type, listener, options)
    }

    setContainer(containerElement: HTMLElement) {
        this.containerElement = containerElement
    }

    setVideoElement(videoElement: HTMLVideoElement) {
        this.videoElement = videoElement

        // Attach iOS-specific listeners
        if (isApple() && this.attachVideoListeners) {
            this.attachVideoListeners()
        }
    }

    async toggleFullscreen() {
        if (this.isFullscreen) {
            await this.exitFullscreen()
        } else {
            await this.enterFullscreen()
        }
    }

    public get isFullscreen(): boolean {
        // Check Electron native fullscreen first
        if (this._isElectron() && this.isElectronNativeFullscreen) {
            return true
        }

        // Check iOS video fullscreen
        if (isApple() && this.videoElement) {
            return !!(this.videoElement as any).webkitDisplayingFullscreen
        }

        return !!(
            document.fullscreenElement ||
            (document as any).webkitFullscreenElement ||
            (document as any).mozFullScreenElement ||
            (document as any).msFullscreenElement
        )
    }

    async exitFullscreen() {
        const attemptEvent: FullscreenManagerAttemptEvent = new CustomEvent("exitattempt", { detail: { method: "exit" } })
        this.dispatchEvent(attemptEvent)

        try {
            if (this._isElectron() && this._shouldUseElectronFullscreen()) {
                await this._exitElectronFullscreen()
                this._focusVideo()
                return
            }

            if (isApple() && this.videoElement && (this.videoElement as any).webkitDisplayingFullscreen) {
                await (this.videoElement as any).webkitExitFullscreen()
                log.info("Exited iOS fullscreen")
                this._focusVideo()
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
            this._focusVideo()
        }
        catch (error) {
            log.error("Failed to exit fullscreen", error)
        }
    }

    async enterFullscreen() {
        const attemptEvent: FullscreenManagerAttemptEvent = new CustomEvent("enterattempt", { detail: { method: "enter" } })
        this.dispatchEvent(attemptEvent)

        if (!this.containerElement) {
            log.warning("Container element not set")
            return
        }

        try {
            if (this._isElectron() && this._shouldUseElectronFullscreen()) {
                await this._enterElectronFullscreen()
                this._focusVideo()
                return
            }

            if (isApple() && this.videoElement) {
                if ((this.videoElement as any).webkitEnterFullscreen) {
                    await (this.videoElement as any).webkitEnterFullscreen()
                    log.info("Entered iOS fullscreen")
                    this._focusVideo()
                    return
                }
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
            this._focusVideo()
        }
        catch (error) {
            log.error("Failed to enter fullscreen", error)
        }
    }

    destroy() {
        // this.exitFullscreen()
        this.controller.abort()
        this.containerElement = null
        this.videoElement = null

        const event: FullscreenManagerDestroyedEvent = new CustomEvent("destroyed")
        this.dispatchEvent(event)
    }

    private _isElectron(): boolean {
        return !!(window as any)?.electron
    }

    private async initElectronFullscreenState(): Promise<void> {
        if (!this._isElectron() || !window.electron?.window?.isFullscreen) {
            return
        }

        try {
            this.isElectronNativeFullscreen = await window.electron.window.isFullscreen()
            log.info("Initial Electron fullscreen state:", this.isElectronNativeFullscreen)
        }
        catch (error) {
            log.error("Failed to get initial Electron fullscreen state", error)
        }
    }

    private _focusVideo(): void {
        if (this.videoElement) {
            setTimeout(() => {
                this.videoElement?.focus()
            }, 100)
        }
    }

    private _shouldUseElectronFullscreen(): boolean {
        // return this._isElectron() && window.electron?.platform === "win32"
        return this._isElectron()
    }

    private async _enterElectronFullscreen(): Promise<void> {
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

    private async _exitElectronFullscreen(): Promise<void> {
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
        if (!this._isElectron()) return

        const removeFullscreenListener = window.electron?.on?.("window:fullscreen", (isFullscreen: boolean) => {
            this.isElectronNativeFullscreen = isFullscreen
            log.info("Electron fullscreen state changed:", isFullscreen)

            const event: FullscreenManagerChangedEvent = new CustomEvent("fullscreenchanged", { detail: { isFullscreen } })
            this.dispatchEvent(event)

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

        if (isApple()) {
            const attachVideoListeners = () => {
                if (this.videoElement) {
                    this.videoElement.addEventListener("webkitbeginfullscreen", this.handleFullscreenChange, {
                        signal: this.controller.signal,
                    })
                    this.videoElement.addEventListener("webkitendfullscreen", this.handleFullscreenChange, {
                        signal: this.controller.signal,
                    })
                }
            }

            attachVideoListeners()

            this.attachVideoListeners = attachVideoListeners
        }
    }

    private handleFullscreenChange = () => {
        const isFullscreen = this.isFullscreen
        log.info("Fullscreen state changed:", isFullscreen)

        const event: FullscreenManagerChangedEvent = new CustomEvent("fullscreenchanged", { detail: { isFullscreen } })
        this.dispatchEvent(event)

        this.onFullscreenChange(isFullscreen)
    }
}
