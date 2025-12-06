import { logger } from "@/lib/helpers/debug"

interface VideoCorePgsRendererOptions {
    videoElement: HTMLVideoElement
    debug?: boolean
}

export interface PgsEvent {
    startTime: number // in seconds
    duration: number // in seconds
    imageData: string // base64 encoded PNG
    width: number // image width
    height: number // image height
    x?: number // position on canvas
    y?: number // position on canvas
    canvasWidth?: number // total canvas dimensions
    canvasHeight?: number
    cropX?: number
    cropY?: number
    cropWidth?: number
    cropHeight?: number
}

const log = logger("PGS RENDERER")

export class VideoCorePgsRenderer {
    private _videoElement: HTMLVideoElement
    _canvas: HTMLCanvasElement | null = null
    private _worker: Worker | null = null
    private _animationFrameId: number | null = null
    private _debug: boolean = false
    private _isDestroyed: boolean = false
    private _canvasWidth: number = 0
    private _canvasHeight: number = 0
    private _resizeObserver: ResizeObserver | null = null

    constructor(options: VideoCorePgsRendererOptions) {
        this._videoElement = options.videoElement
        this._debug = options.debug ?? false
        this._setupCanvas()
        this._setupWorker()
        this._startRenderLoop()
    }

    addEvent(event: PgsEvent) {
        if (!this._worker) {
            return
        }

        this._worker.postMessage({
            type: "addEvent",
            payload: event,
        })
    }

    resize() {
        if (!this._canvas || !this._worker) {
            return
        }

        const videoContentSize = this._getRenderedVideoContentSize()
        if (!videoContentSize) {
            return
        }

        const { displayedWidth, displayedHeight, offsetX, offsetY } = videoContentSize

        // Store dimensions for render loop
        this._canvasWidth = displayedWidth
        this._canvasHeight = displayedHeight

        // Update canvas style (position and display size)
        // devnote: can't modify width/height after transferControlToOffscreen
        this._canvas.style.width = `${displayedWidth}px`
        this._canvas.style.height = `${displayedHeight}px`
        this._canvas.style.left = `${offsetX}px`
        this._canvas.style.top = `${offsetY}px`

        // Notify worker of resize to update internal canvas dimensions
        this._worker.postMessage({
            type: "resize",
            payload: {
                width: displayedWidth,
                height: displayedHeight,
            },
        })

        if (this._debug) {
            log.info("Resized canvas", {
                width: displayedWidth,
                height: displayedHeight,
                left: offsetX,
                top: offsetY,
            })
        }
    }

    setTimeOffset(offset: number) {
        if (!this._worker) {
            return
        }

        this._worker.postMessage({
            type: "setTimeOffset",
            payload: { offset },
        })
    }

    stop() {
        // no-op
    }

    clear() {
        if (!this._worker) {
            return
        }

        this._worker.postMessage({
            type: "clear",
        })
    }

    destroy() {
        this._isDestroyed = true

        if (this._animationFrameId !== null) {
            cancelAnimationFrame(this._animationFrameId)
            this._animationFrameId = null
        }

        // Terminate worker
        if (this._worker) {
            this._worker.terminate()
            this._worker = null
        }

        // Clean up resize observer
        if (this._resizeObserver) {
            this._resizeObserver.disconnect()
            this._resizeObserver = null
        }

        if (this._canvas && this._canvas.parentElement) {
            this._canvas.parentElement.removeChild(this._canvas)
        }

        this._canvas = null
    }

    private _setupCanvas() {
        // Create canvas element
        this._canvas = document.createElement("canvas")
        this._canvas.style.position = "absolute"
        this._canvas.style.pointerEvents = "none"
        this._canvas.style.zIndex = "10"
        this._canvas.style.objectFit = "contain"
        this._canvas.style.objectPosition = "center"
        this._canvas.className = "vc-pgs-canvas"

        // Insert canvas after video element
        const parent = this._videoElement.parentElement
        if (parent) {
            parent.style.position = "relative"
            parent.appendChild(this._canvas)
        }

        // Set initial size before transferring
        const videoContentSize = this._getRenderedVideoContentSize()
        if (videoContentSize) {
            const { displayedWidth, displayedHeight, offsetX, offsetY } = videoContentSize
            this._canvas.width = displayedWidth
            this._canvas.height = displayedHeight
            this._canvasWidth = displayedWidth
            this._canvasHeight = displayedHeight
            this._canvas.style.width = `${displayedWidth}px`
            this._canvas.style.height = `${displayedHeight}px`
            this._canvas.style.left = `${offsetX}px`
            this._canvas.style.top = `${offsetY}px`
        }

        // Add resize observer to handle video element size changes
        this._setupResizeObserver()
    }

    private _setupWorker() {
        if (!this._canvas) {
            return
        }

        this._worker = new window.Worker("/pgs-renderer.worker.js")

        // Handle messages
        this._worker.onmessage = (e) => {
            const { type, payload } = e.data

            if (type === "debug" && this._debug) {
                log.info(payload.message, payload.data)
            } else if (type === "error") {
                log.error(payload.message, payload.error)
            }
        }

        // Transfer canvas to worker
        const offscreenCanvas = this._canvas.transferControlToOffscreen()
        this._worker.postMessage({
            type: "init",
            payload: {
                canvas: offscreenCanvas,
                debug: this._debug,
            },
        }, [offscreenCanvas])
    }

    private _setupResizeObserver() {
        if (!this._videoElement || !this._canvas) return

        this._resizeObserver = new ResizeObserver(() => {
            this.resize()
        })

        this._resizeObserver.observe(this._videoElement)
    }

    private _startRenderLoop() {
        const render = () => {
            if (this._isDestroyed) {
                return
            }

            // Send render request to worker with current video state
            if (this._worker && (!this._videoElement.paused || this._videoElement.seeking)) {
                this._worker.postMessage({
                    type: "render",
                    payload: {
                        currentTime: this._videoElement.currentTime,
                        canvasWidth: this._canvasWidth,
                        canvasHeight: this._canvasHeight,
                        isPlaying: !this._videoElement.paused,
                    },
                })
            }

            this._animationFrameId = requestAnimationFrame(render)
        }

        render()
    }


    private _getRenderedVideoContentSize() {
        const containerWidth = this._videoElement.clientWidth
        const containerHeight = this._videoElement.clientHeight

        const videoWidth = this._videoElement.videoWidth
        const videoHeight = this._videoElement.videoHeight

        if (!videoWidth || !videoHeight) return null

        const containerRatio = containerWidth / containerHeight
        const videoRatio = videoWidth / videoHeight

        let displayedWidth: number
        let displayedHeight: number
        let offsetX = 0
        let offsetY = 0

        const objectFit = getComputedStyle(this._videoElement).objectFit || "contain"

        if (objectFit === "cover") {
            if (videoRatio > containerRatio) {
                displayedHeight = containerHeight
                displayedWidth = containerHeight * videoRatio
                offsetX = (containerWidth - displayedWidth) / 2
            } else {
                displayedWidth = containerWidth
                displayedHeight = containerWidth / videoRatio
                offsetY = (containerHeight - displayedHeight) / 2
            }
        } else if (objectFit === "contain") {
            if (videoRatio > containerRatio) {
                displayedWidth = containerWidth
                displayedHeight = containerWidth / videoRatio
                offsetY = (containerHeight - displayedHeight) / 2
            } else {
                displayedHeight = containerHeight
                displayedWidth = containerHeight * videoRatio
                offsetX = (containerWidth - displayedWidth) / 2
            }
        } else {
            // object-fit: fill or none
            displayedWidth = containerWidth
            displayedHeight = containerHeight
        }

        return { displayedWidth, displayedHeight, offsetX, offsetY }
    }
}
