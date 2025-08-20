import { VideoCoreSettings } from "@/app/(main)/_features/video-core/video-core.atoms"
import { logger } from "@/lib/helpers/debug"
import {
    Anime4KPipeline,
    CNNx2M,
    CNNx2UL,
    CNNx2VL,
    DenoiseCNNx2VL,
    GANx3L,
    GANx4UUL,
    ModeA,
    ModeAA,
    ModeB,
    ModeBB,
    ModeC,
    ModeCA,
    render,
} from "anime4k-webgpu"

const log = logger("VIDEO CORE ANIME 4K MANAGER")

export type Anime4KOption =
    "off"
    | "mode-a"
    | "mode-b"
    | "mode-c"
    | "mode-aa"
    | "mode-bb"
    | "mode-ca"
    | "cnn-2x-medium"
    | "cnn-2x-very-large"
    | "denoise-cnn-2x-very-large"
    | "cnn-2x-ultra-large"
    | "gan-3x-large"
    | "gan-4x-ultra-large"

interface FrameDropState {
    enabled: boolean
    frameDropThreshold: number
    frameDropCount: number
    lastFrameTime: number
    targetFrameTime: number
    performanceGracePeriod: number
    initTime: number
}

export class VideoCoreAnime4KManager {
    canvas: HTMLCanvasElement | null = null
    private readonly videoElement: HTMLVideoElement
    private settings: VideoCoreSettings
    private _currentOption: Anime4KOption = "off"
    private _webgpuResources: { device?: GPUDevice; pipelines?: any[] } | null = null
    private _renderLoopId: number | null = null
    private _abortController: AbortController | null = null
    private _frameDropState: FrameDropState = {
        enabled: true,
        frameDropThreshold: 5,
        frameDropCount: 0,
        lastFrameTime: 0,
        targetFrameTime: 1000 / 30, // 30fps target
        performanceGracePeriod: 1000,
        initTime: 0,
    }
    private readonly _onFallback?: (message: string) => void
    private readonly _onOptionChanged?: (option: Anime4KOption) => void
    private _boxSize: { width: number; height: number } = { width: 0, height: 0 }
    private _initializationTimeout: NodeJS.Timeout | null = null
    private _initialized = false
    private _onCanvasCreatedCallbacks: Set<(canvas: HTMLCanvasElement) => void> = new Set()

    constructor({
        videoElement,
        settings,
        onFallback,
        onOptionChanged,
    }: {
        videoElement: HTMLVideoElement
        settings: VideoCoreSettings
        onFallback?: (message: string) => void
        onOptionChanged?: (option: Anime4KOption) => void
    }) {
        this.videoElement = videoElement
        this.settings = settings
        this._onFallback = onFallback
        this._onOptionChanged = onOptionChanged

        log.info("Anime4K manager initialized")
    }

    updateCanvasSize(size: { width: number; height: number }) {
        this._boxSize = size
        if (this.canvas) {
            this.canvas.width = size.width
            this.canvas.height = size.height
            this.canvas.style.top = this.videoElement.getBoundingClientRect().top + "px"
            log.info("Updating canvas size", { width: size.width, height: size.height, top: this.videoElement.getBoundingClientRect().top })
        }
    }

    // Adds a function to be called whenever the canvas is created or recreated
    registerOnCanvasCreated(callback: (canvas: HTMLCanvasElement) => void) {
        this._onCanvasCreatedCallbacks.add(callback)
    }

    // Select an Anime4K option
    async setOption(option: Anime4KOption, state?: {
        isMiniPlayer: boolean
        isPip: boolean
        seeking: boolean
    }) {

        const previousOption = this._currentOption
        this._currentOption = option

        if (option === "off") {
            log.info("Anime4K turned off")
            this.destroy()
            return
        }

        // Handle change of state
        if (state) {
            // For PIP or mini player, completely destroy the canvas
            if (state.isMiniPlayer || state.isPip) {
                log.info("Destroying canvas due to PIP/mini player mode")
                this.destroy()
                return
            }

            // For seeking, just hide the canvas
            if (state.seeking) {
                this._hideCanvas()
                return
            }
        }

        // Skip initialization if size isn't set
        if (this._boxSize.width === 0 || this._boxSize.height === 0) {
            return
        }

        // If canvas exists but is hidden, show it
        if (this.canvas && this._isCanvasHidden()) {
            log.info("Showing previously hidden canvas")
            this._showCanvas()
            return
        }

        // If option changed or no canvas exists, reinitialize
        if (previousOption !== option || !this.canvas) {
            log.info("Change detected, reinitializing canvas")
            this.destroy()
            try {
                await this._initialize()
            }
            catch (error) {
                log.error("Failed to initialize Anime4K", error)
                this._handleError(error instanceof Error ? error.message : "Unknown error")
            }
        }
    }

    // initialize the canvas and start rendering

    // Destroy and cleanup resources
    destroy() {
        this.videoElement.style.opacity = "1"

        this._initialized = false

        if (this._initializationTimeout) {
            clearTimeout(this._initializationTimeout)
            this._initializationTimeout = null
        }

        if (this.canvas) {
            this.canvas.remove()
            this.canvas = null
        }

        if (this._renderLoopId) {
            cancelAnimationFrame(this._renderLoopId)
            this._renderLoopId = null
        }

        if (this._webgpuResources?.device) {
            this._webgpuResources.device.destroy()
            this._webgpuResources = null
        }

        if (this._abortController) {
            this._abortController.abort()
            this._abortController = null
        }

        this._frameDropState.frameDropCount = 0
        this._frameDropState.lastFrameTime = 0
    }

    // throws if initialization fails
    private async _initialize() {
        if (this._initialized || this._currentOption === "off") {
            return
        }

        log.info("Initializing Anime4K", this._currentOption)

        this._abortController = new AbortController()
        this._frameDropState = {
            ...this._frameDropState,
            frameDropCount: 0,
            initTime: performance.now(),
            lastFrameTime: 0,
        }

        // Check WebGPU support, create canvas, and start rendering
        try {
            const gpuInfo = await this.getGPUInfo()
            if (!gpuInfo) {
                throw new Error("WebGPU not supported")
            }

            if (this._abortController.signal.aborted) return

            this._createCanvas()

            if (this._abortController.signal.aborted) return

            await this._startRendering()

            this._initialized = true
            log.info("Anime4K initialized")
        }
        catch (error) {
            if (!this._abortController?.signal.aborted) {
                log.error("Initialization failed", error)
                throw error
            }
        }
    }

    // Create and position the canvas
    private _createCanvas() {
        if (this._abortController?.signal.aborted) return

        this.canvas = document.createElement("canvas")

        this.canvas.width = this._boxSize.width
        this.canvas.height = this._boxSize.height
        this.canvas.style.objectFit = "cover"
        this.canvas.style.position = "absolute"
        this.canvas.style.top = this.videoElement.getBoundingClientRect().top + "px"
        this.canvas.style.left = "0"
        this.canvas.style.right = "0"
        this.canvas.style.pointerEvents = "none"
        this.canvas.style.zIndex = "2"
        this.canvas.style.display = "block"
        this.canvas.className = "vc-anime4k-canvas"
        log.info("Creating canvas", { width: this.canvas.width, height: this.canvas.height, top: this.canvas.style.top })

        this.videoElement.parentElement?.appendChild(this.canvas)
        this.videoElement.style.opacity = "0"

        for (const callback of this._onCanvasCreatedCallbacks) {
            callback(this.canvas)
        }
    }

    // WebGPU rendering
    private async _startRendering() {
        if (!this.canvas || !this.videoElement || this._currentOption === "off") {
            console.warn("stopped started")
            return
        }

        const nativeDimensions = {
            width: this.videoElement.videoWidth,
            height: this.videoElement.videoHeight,
        }

        const targetDimensions = {
            width: this.canvas.width,
            height: this.canvas.height,
        }

        log.info("Rendering started")

        await render({
            video: this.videoElement,
            canvas: this.canvas,
            pipelineBuilder: (device, inputTexture) => {
                this._webgpuResources = { device }

                const commonProps = {
                    device,
                    inputTexture,
                    nativeDimensions,
                    targetDimensions,
                }

                return this.createPipeline(commonProps)
            },
        })

        // Start frame drop detection if enabled
        if (this._frameDropState.enabled && this._isOptionSelected(this._currentOption)) {
            this._startFrameDropDetection()
        }
    }

    private createPipeline(commonProps: any): [Anime4KPipeline] {
        switch (this._currentOption) {
            case "mode-a":
                return [new ModeA(commonProps)]
            case "mode-b":
                return [new ModeB(commonProps)]
            case "mode-c":
                return [new ModeC(commonProps)]
            case "mode-aa":
                return [new ModeAA(commonProps)]
            case "mode-bb":
                return [new ModeBB(commonProps)]
            case "mode-ca":
                return [new ModeCA(commonProps)]
            case "cnn-2x-medium":
                return [new CNNx2M(commonProps)]
            case "cnn-2x-very-large":
                return [new CNNx2VL(commonProps)]
            case "denoise-cnn-2x-very-large":
                return [new DenoiseCNNx2VL(commonProps)]
            case "cnn-2x-ultra-large":
                return [new CNNx2UL(commonProps)]
            case "gan-3x-large":
                return [new GANx3L(commonProps)]
            case "gan-4x-ultra-large":
                return [new GANx4UUL(commonProps)]
            default:
                return [new ModeA(commonProps)]
        }
    }

    // Start frame drop detection loop
    private _startFrameDropDetection() {
        const frameDetectionLoop = () => {
            if (this._isOptionSelected(this._currentOption) && this._renderLoopId !== null) {
                this._detectFrameDrops()
                this._renderLoopId = requestAnimationFrame(frameDetectionLoop)
            }
        }
        this._renderLoopId = requestAnimationFrame(frameDetectionLoop)
    }

    // Detect frame drops and stop when it gets bad
    private _detectFrameDrops() {
        if (!this._isOptionSelected(this._currentOption)) {
            return
        }

        const now = performance.now()
        const timeSinceInit = now - this._frameDropState.initTime

        // Skip detection during grace period
        if (timeSinceInit < this._frameDropState.performanceGracePeriod) {
            this._frameDropState.lastFrameTime = now
            return
        }

        if (this._frameDropState.lastFrameTime > 0) {
            const frameTime = now - this._frameDropState.lastFrameTime
            const isFrameDrop = frameTime > this._frameDropState.targetFrameTime * 1.5 // 50% tolerance

            if (isFrameDrop) {
                this._frameDropState.frameDropCount++

                if (this._frameDropState.frameDropCount >= this._frameDropState.frameDropThreshold) {
                    log.warning(`Detected ${this._frameDropState.frameDropCount} consecutive frame drops. Falling back to 'off' mode.`)
                    this._handlePerformanceFallback()
                    return
                }
            } else {
                // Reset on successful frame
                this._frameDropState.frameDropCount = 0
            }
        }

        this._frameDropState.lastFrameTime = now
    }

    private _handlePerformanceFallback() {
        this._onFallback?.("Performance degraded. Turning off Anime4K.")
        this.setOption("off")
        this._onOptionChanged?.("off")
    }

    private _handleError(message: string) {
        this._onFallback?.(`Anime4K: ${message}`)
        this.setOption("off")
        this._onOptionChanged?.("off")
    }

    // Get GPU information
    private async getGPUInfo() {
        if (!navigator.gpu) return null

        try {
            const adapter = await navigator.gpu.requestAdapter()
            if (!adapter) return null

            const device = await adapter.requestDevice()
            if (!device) return null

            const info = (adapter as any).info || {}

            return {
                gpu: info.vendor || info.architecture || "Unknown GPU",
                vendor: info.vendor || "Unknown",
                device,
            }
        }
        catch {
            return null
        }
    }

    private _isOptionSelected(option: Anime4KOption): boolean {
        return option !== "off"
    }

    private _hideCanvas() {
        if (this.canvas) {
            this.canvas.style.display = "none"
            this.videoElement.style.opacity = "1"
        }
    }

    private _showCanvas() {
        if (this.canvas) {
            this.canvas.style.display = "block"
            this.videoElement.style.opacity = "0"
        }
    }

    private _isCanvasHidden(): boolean {
        return this.canvas ? this.canvas.style.display === "none" : false
    }
}
