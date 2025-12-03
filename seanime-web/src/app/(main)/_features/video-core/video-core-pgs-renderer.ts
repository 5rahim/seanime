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
    private _canvas: HTMLCanvasElement | null = null
    private _ctx: CanvasRenderingContext2D | null = null
    private _events: Map<string, PgsEvent> = new Map()
    private _imageCache: Map<string, HTMLImageElement> = new Map()
    private _currentEvent: PgsEvent | null = null
    private _currentEventRendered: boolean = false
    private _animationFrameId: number | null = null
    private _timeOffset: number = 0 // offset internal video events
    private _debug: boolean = false
    private _isDestroyed: boolean = false

    constructor(options: VideoCorePgsRendererOptions) {
        this._videoElement = options.videoElement
        this._debug = options.debug ?? false
        this._setupCanvas()
        this._startRenderLoop()
    }

    addEvent(event: PgsEvent) {
        const key = this._getEventKey(event)

        if (this._events.has(key)) {
            return // Already have this event
        }

        this._events.set(key, event)

        // Preload the image
        this._preloadImage(event.imageData).then(img => {
            this._imageCache.set(event.imageData, img)
            if (this._debug) {
                log.info("Preloaded image", {
                    startTime: event.startTime,
                    width: img.width,
                    height: img.height,
                })
            }
        }).catch(err => {
            log.error("Failed to preload image", err)
        })

        if (this._debug) {
            log.info("Added PGS event", {
                startTime: event.startTime,
                endTime: event.startTime + event.duration,
                duration: event.duration,
                width: event.width,
                height: event.height,
                x: event.x,
                y: event.y,
                canvasWidth: event.canvasWidth,
                canvasHeight: event.canvasHeight,
            })
        }
    }

    resize() {
        if (!this._canvas) {
            return
        }

        const videoContentSize = this._getRenderedVideoContentSize()
        if (!videoContentSize) {
            return
        }

        const { displayedWidth, displayedHeight, offsetX, offsetY } = videoContentSize

        // Set canvas size to match the actual video content
        this._canvas.width = displayedWidth
        this._canvas.height = displayedHeight

        // Position the canvas to match the video content position
        this._canvas.style.width = `${displayedWidth}px`
        this._canvas.style.height = `${displayedHeight}px`
        this._canvas.style.left = `${offsetX}px`
        this._canvas.style.top = `${offsetY}px`

        // Force re-render on resize
        this._currentEventRendered = false

        if (this._debug) {
            log.info("Resized canvas", {
                width: this._canvas.width,
                height: this._canvas.height,
                left: offsetX,
                top: offsetY,
            })
        }
    }

    setTimeOffset(offset: number) {
        this._timeOffset = offset
    }

    stop() {
        // this._currentEvent = null
        // this._currentEventRendered = false
        // if (this._ctx && this._canvas) {
        //     this._ctx.clearRect(0, 0, this._canvas.width, this._canvas.height)
        // }
    }

    clear() {
        this._events.clear()
        this._imageCache.clear()
        this._currentEvent = null
        this._currentEventRendered = false

        if (this._ctx && this._canvas) {
            this._ctx.clearRect(0, 0, this._canvas.width, this._canvas.height)
        }
    }

    destroy() {
        this._isDestroyed = true

        if (this._animationFrameId !== null) {
            cancelAnimationFrame(this._animationFrameId)
            this._animationFrameId = null
        }

        this.clear()

        // Clean up resize observer
        if (this._canvas && (this._canvas as any)._resizeObserver) {
            (this._canvas as any)._resizeObserver.disconnect()
            (this._canvas as any)._resizeObserver = null
        }

        if (this._canvas && this._canvas.parentElement) {
            this._canvas.parentElement.removeChild(this._canvas)
        }

        this._canvas = null
        this._ctx = null
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

        // Get 2D context
        this._ctx = this._canvas.getContext("2d", { alpha: true })

        // Insert canvas after video element
        const parent = this._videoElement.parentElement
        if (parent) {
            parent.style.position = "relative"
            parent.appendChild(this._canvas)
        }

        this.resize()

        // Add resize observer to handle video element size changes
        this._setupResizeObserver()
    }

    private _setupResizeObserver() {
        if (!this._videoElement || !this._canvas) return

        const resizeObserver = new ResizeObserver(() => {
            this.resize()
        })

        resizeObserver.observe(this._videoElement)

        // Store for cleanup
        ;(this._canvas as any)._resizeObserver = resizeObserver
    }

    private _getEventKey(event: PgsEvent): string {
        return `${event.startTime}-${event.duration}-${event.imageData.substring(0, 50)}`
    }

    private async _preloadImage(base64Data: string): Promise<HTMLImageElement> {
        return new Promise((resolve, reject) => {
            const img = new Image()
            img.onload = () => resolve(img)
            img.onerror = reject
            img.src = base64Data
        })
    }

    private _startRenderLoop() {
        const render = () => {
            if (this._isDestroyed) {
                return
            }

            // Only render if video is playing or seeking
            if (!this._videoElement.paused || this._videoElement.seeking) {
                this._render()
            }

            this._animationFrameId = requestAnimationFrame(render)
        }

        render()
    }

    private _render() {
        if (!this._ctx || !this._canvas) {
            return
        }

        const currentTime = this._videoElement.currentTime + this._timeOffset

        // Find the event that should be displayed at current time
        let eventToDisplay: PgsEvent | null = null

        for (const event of this._events.values()) {
            const startTime = event.startTime
            const endTime = event.startTime + event.duration

            if (currentTime >= startTime && currentTime <= endTime) {
                eventToDisplay = event
                break
            }
        }

        // If event changed, clear canvas and log
        if (eventToDisplay !== this._currentEvent) {
            this._ctx.clearRect(0, 0, this._canvas.width, this._canvas.height)
            this._currentEventRendered = false

            if (this._debug && eventToDisplay) {
                log.info("Displaying new PGS event", {
                    currentTime,
                    startTime: eventToDisplay.startTime,
                    endTime: eventToDisplay.startTime + eventToDisplay.duration,
                    canvasWidth: this._canvas.width,
                    canvasHeight: this._canvas.height,
                })
            } else if (this._debug && this._currentEvent && !eventToDisplay) {
                log.info("Cleared PGS event", { currentTime })
            }

            this._currentEvent = eventToDisplay
        }

        // Render current event only if it hasn't been rendered yet
        if (eventToDisplay && !this._currentEventRendered) {
            this._renderEvent(eventToDisplay)
            this._currentEventRendered = true
        } else if (!eventToDisplay) {
            // No event to display, ensure canvas is clear
            this._ctx.clearRect(0, 0, this._canvas.width, this._canvas.height)
            this._currentEventRendered = false
        }
    }

    private _renderEvent(event: PgsEvent) {
        if (!this._ctx || !this._canvas) {
            return
        }

        const img = this._imageCache.get(event.imageData)
        if (!img || !img.complete) {
            return
        }

        // Canvas dimensions
        const canvasWidth = this._canvas.width
        const canvasHeight = this._canvas.height

        // Video canvas dimensions from event or use canvas size
        const videoCanvasWidth = event.canvasWidth || canvasWidth
        const videoCanvasHeight = event.canvasHeight || canvasHeight

        // Scale factors
        const scaleX = canvasWidth / videoCanvasWidth
        const scaleY = canvasHeight / videoCanvasHeight

        // Position (default to bottom center if not specified)
        let x = event.x !== undefined ? event.x : (videoCanvasWidth - event.width) / 2
        let y = event.y !== undefined ? event.y : videoCanvasHeight - event.height - 20

        // Apply scaling
        x *= scaleX
        y *= scaleY

        const width = event.width * scaleX
        const height = event.height * scaleY

        if (this._debug) {
            log.info("Rendering PGS image", {
                x,
                y,
                width,
                height,
                scaleX,
                scaleY,
                imgWidth: img.width,
                imgHeight: img.height,
                dataURL: img.src,
            })
        }

        // Handle cropping if specified
        if (event.cropX !== undefined && event.cropY !== undefined &&
            event.cropWidth !== undefined && event.cropHeight !== undefined) {

            const sx = event.cropX
            const sy = event.cropY
            const sWidth = event.cropWidth
            const sHeight = event.cropHeight

            this._ctx.drawImage(
                img,
                sx, sy, sWidth, sHeight,
                x, y, width, height,
            )
        } else {
            // Draw the full image
            this._ctx.drawImage(img, x, y, width, height)
        }

        if (this._debug) {
            // Draw translucent overlay over entire canvas when subtitle is present
            this._ctx.fillStyle = "rgba(255, 0, 255, 0.1)"
            this._ctx.fillRect(0, 0, canvasWidth, canvasHeight)

            // Draw debug border around subtitle
            this._ctx.strokeStyle = "purple"
            this._ctx.lineWidth = 2
            this._ctx.strokeRect(x, y, width, height)
        }
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
