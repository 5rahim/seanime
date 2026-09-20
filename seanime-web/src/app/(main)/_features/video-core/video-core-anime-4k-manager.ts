import { VideoCoreSettings } from "@/app/(main)/_features/video-core/video-core.atoms"
import { logger } from "@/lib/helpers/debug"
import type { Anime4KPipeline } from "anime4k-webgpu"

const log = logger("VIDEO CORE ANIME 4K MANAGER")

const anime4kCaptureShader = /* wgsl */`
struct VertexOutput {
    @builtin(position) position: vec4<f32>,
    @location(0) uv: vec2<f32>,
}

@vertex
fn vert_main(@builtin(vertex_index) vertexIndex: u32) -> VertexOutput {
    var positions = array<vec2<f32>, 6>(
        vec2<f32>(-1.0, -1.0),
        vec2<f32>(1.0, -1.0),
        vec2<f32>(-1.0, 1.0),
        vec2<f32>(-1.0, 1.0),
        vec2<f32>(1.0, -1.0),
        vec2<f32>(1.0, 1.0),
    );

    var uvs = array<vec2<f32>, 6>(
        vec2<f32>(0.0, 1.0),
        vec2<f32>(1.0, 1.0),
        vec2<f32>(0.0, 0.0),
        vec2<f32>(0.0, 0.0),
        vec2<f32>(1.0, 1.0),
        vec2<f32>(1.0, 0.0),
    );

    var output: VertexOutput;
    output.position = vec4<f32>(positions[vertexIndex], 0.0, 1.0);
    output.uv = uvs[vertexIndex];
    return output;
}

@group(0) @binding(0) var inputTexture: texture_2d<f32>;

@fragment
fn frag_main(input: VertexOutput) -> @location(0) vec4<f32> {
    let textureSize = vec2<f32>(textureDimensions(inputTexture));
    let coords = vec2<i32>(clamp(input.uv * textureSize, vec2<f32>(0.0), textureSize - vec2<f32>(1.0)));
    return textureLoad(inputTexture, coords, 0);
}
`

// Presents the last pipeline's output onto the canvas, sampled the same way anime4k-webgpu's own renderer does
const anime4kBlitShader = /* wgsl */`
struct VertexOutput {
    @builtin(position) Position: vec4<f32>,
    @location(0) fragUV: vec2<f32>,
}

@vertex
fn vert_main(@builtin(vertex_index) VertexIndex: u32) -> VertexOutput {
    const pos = array(vec2(-1.0, -1.0), vec2(3.0, -1.0), vec2(-1.0, 3.0));

    var output: VertexOutput;
    output.Position = vec4(pos[VertexIndex], 0.0, 1.0);
    output.fragUV = vec2(pos[VertexIndex].x * 0.5 + 0.5, 0.5 - pos[VertexIndex].y * 0.5);
    return output;
}

@group(0) @binding(1) var mySampler: sampler;
@group(0) @binding(2) var myTexture: texture_2d<f32>;

@fragment
fn main(@location(0) fragUV: vec2f) -> @location(0) vec4f {
    return textureSampleBaseClampToEdge(myTexture, mySampler, fragUV);
}
`

export type Anime4KManagerCanvasCreatedEvent = CustomEvent<{ canvas: HTMLCanvasElement }>
export type Anime4KManagerOptionChangedEvent = CustomEvent<{ newOption: Anime4KOption }>
export type Anime4KManagerErrorEvent = CustomEvent<{ message: string }>
export type Anime4KManagerCanvasResizedEvent = CustomEvent<{ width: number; height: number }>
export type Anime4KManagerDestroyedEvent = CustomEvent

interface VideoCoreAnime4KManagerEventMap {
    "canvascreated": Anime4KManagerCanvasCreatedEvent
    "optionchanged": Anime4KManagerOptionChangedEvent
    "error": Anime4KManagerErrorEvent
    "canvasresized": Anime4KManagerCanvasResizedEvent
    "destroyed": Anime4KManagerDestroyedEvent
}


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
    totalFrameDrops: number
    lastFrameTime: number
    targetFrameTime: number
    performanceGracePeriod: number
    initTime: number
}

interface RenderStats {
    currentFps: number
    frameTimeSamples: number[]
    lastRenderTime: number
    renderCallbackId: number | null
}

interface Anime4KWebGPUResources {
    device?: GPUDevice
    pipelines?: Anime4KPipeline[]
    outputTexture?: GPUTexture
    context?: GPUCanvasContext
    inputTexture?: GPUTexture
}

export class VideoCoreAnime4KManager extends EventTarget {
    canvas: HTMLCanvasElement | null = null
    private readonly videoElement: HTMLVideoElement
    private settings: VideoCoreSettings
    private _currentOption: Anime4KOption = "off"
    private _webgpuResources: Anime4KWebGPUResources | null = null
    private _renderLoopId: number | null = null
    private _abortController: AbortController | null = null
    private _frameDropState: FrameDropState = {
        enabled: true,
        frameDropThreshold: 8,
        frameDropCount: 0,
        totalFrameDrops: 0,
        lastFrameTime: 0,
        targetFrameTime: 1000 / 16, // 30fps target
        performanceGracePeriod: 1000,
        initTime: 0,
    }
    private _renderStats: RenderStats = {
        currentFps: 0,
        frameTimeSamples: [],
        lastRenderTime: 0,
        renderCallbackId: null,
    }
    private readonly _onFallback?: (message: string) => void
    private readonly _onOptionChanged?: (option: Anime4KOption) => void
    private _boxSize: { width: number; height: number } = { width: 0, height: 0 }
    private _initializationTimeout: NodeJS.Timeout | null = null
    private _initialized = false
    // Incremented by _stopRendering() to invalidate the running render session.
    private _renderSessionId = 0
    private _frameCallbackId: number | null = null
    private _framesInFlight = 0
    // Frames allowed on the GPU queue at once, beyond which frames are skipped instead of queued.
    private _maxFramesInFlight = 2
    private _onCanvasCreatedCallbacks: Set<(canvas: HTMLCanvasElement) => void> = new Set()
    private _onCanvasCreatedCallbacksOnce: Set<(canvas: HTMLCanvasElement) => void> = new Set()

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
        super()
        this.videoElement = videoElement
        this.settings = settings
        this._onFallback = onFallback
        this._onOptionChanged = onOptionChanged

        log.info("Anime4K manager initialized")
    }

    getStats() {
        return {
            currentOption: this._currentOption,
            frameDropCount: this._frameDropState.frameDropCount,
            totalFrameDrops: this._frameDropState.totalFrameDrops,
            currentFps: this._renderStats.currentFps,
            targetFrameTime: this._frameDropState.targetFrameTime,
            lastFrameTime: this._frameDropState.lastFrameTime,
        }
    }

    getCurrentOption(): Anime4KOption {
        return this._currentOption
    }

    async captureFrame(): Promise<Blob | null> {
        const device = this._webgpuResources?.device
        const outputTexture = this._webgpuResources?.outputTexture

        if (!device || !outputTexture) {
            return null
        }

        const width = outputTexture.width
        const height = outputTexture.height

        if (!width || !height) {
            return null
        }

        const captureTexture = device.createTexture({
            label: "anime4k-screenshot-texture",
            size: [width, height, 1],
            format: "rgba8unorm",
            usage: GPUTextureUsage.RENDER_ATTACHMENT | GPUTextureUsage.COPY_SRC,
        })

        const bindGroupLayout = device.createBindGroupLayout({
            entries: [{
                binding: 0,
                visibility: GPUShaderStage.FRAGMENT,
                texture: { sampleType: "float" },
            }],
        })

        const pipeline = device.createRenderPipeline({
            layout: device.createPipelineLayout({ bindGroupLayouts: [bindGroupLayout] }),
            vertex: {
                module: device.createShaderModule({ code: anime4kCaptureShader }),
                entryPoint: "vert_main",
            },
            fragment: {
                module: device.createShaderModule({ code: anime4kCaptureShader }),
                entryPoint: "frag_main",
                targets: [{ format: "rgba8unorm" }],
            },
            primitive: { topology: "triangle-list" },
        })

        const bindGroup = device.createBindGroup({
            layout: bindGroupLayout,
            entries: [{ binding: 0, resource: outputTexture.createView() }],
        })

        const bytesPerPixel = 4
        const unpaddedBytesPerRow = width * bytesPerPixel
        const bytesPerRow = Math.ceil(unpaddedBytesPerRow / 256) * 256
        const buffer = device.createBuffer({
            label: "anime4k-screenshot-buffer",
            size: bytesPerRow * height,
            usage: GPUBufferUsage.COPY_DST | GPUBufferUsage.MAP_READ,
        })

        try {
            const encoder = device.createCommandEncoder()
            const pass = encoder.beginRenderPass({
                colorAttachments: [{
                    view: captureTexture.createView(),
                    clearValue: { r: 0, g: 0, b: 0, a: 1 },
                    loadOp: "clear",
                    storeOp: "store",
                }],
            })

            pass.setPipeline(pipeline)
            pass.setBindGroup(0, bindGroup)
            pass.draw(6)
            pass.end()

            encoder.copyTextureToBuffer(
                { texture: captureTexture },
                { buffer, bytesPerRow, rowsPerImage: height },
                { width, height, depthOrArrayLayers: 1 },
            )

            device.queue.submit([encoder.finish()])

            await buffer.mapAsync(GPUMapMode.READ)

            const mappedRange = new Uint8Array(buffer.getMappedRange())
            const pixels = new Uint8ClampedArray(unpaddedBytesPerRow * height)

            for (let row = 0; row < height; row++) {
                const sourceOffset = row * bytesPerRow
                const destinationOffset = row * unpaddedBytesPerRow
                pixels.set(mappedRange.subarray(sourceOffset, sourceOffset + unpaddedBytesPerRow), destinationOffset)
            }

            const canvas = document.createElement("canvas")
            const ctx = canvas.getContext("2d")
            if (!ctx) {
                canvas.remove()
                return null
            }

            canvas.width = width
            canvas.height = height
            ctx.putImageData(new ImageData(pixels, width, height), 0, 0)

            return await new Promise((resolve) => {
                canvas.toBlob((blob) => {
                    canvas.remove()
                    resolve(blob)
                }, "image/png")
            })
        }
        catch (error) {
            log.error("Failed to capture Anime4K frame", error)
            return null
        }
        finally {
            if (buffer.mapState === "mapped") {
                buffer.unmap()
            }
            buffer.destroy()
            captureTexture.destroy()
        }
    }

    addEventListener<K extends keyof VideoCoreAnime4KManagerEventMap>(
        type: K,
        listener: (this: VideoCoreAnime4KManager, ev: VideoCoreAnime4KManagerEventMap[K]) => any,
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

    removeEventListener<K extends keyof VideoCoreAnime4KManagerEventMap>(
        type: K,
        listener: (this: VideoCoreAnime4KManager, ev: VideoCoreAnime4KManagerEventMap[K]) => any,
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

    updateCanvasSize(_size: { width: number; height: number }) {
        const rect = this.videoElement.getBoundingClientRect()
        const containerWidth = rect.width
        const containerHeight = rect.height
        const videoContentSize = this.getRenderedVideoContentSize(this.videoElement, containerWidth, containerHeight)
        this._boxSize = { width: videoContentSize?.displayedWidth || containerWidth, height: videoContentSize?.displayedHeight || containerHeight }
        if (this.canvas) {
            this.canvas.width = this._boxSize.width
            this.canvas.height = this._boxSize.height
            this.canvas.style.width = this._boxSize.width + "px"
            this.canvas.style.height = this._boxSize.height + "px"
            log.info("Updating canvas size", { ...this._boxSize })
        }

        const event: Anime4KManagerCanvasResizedEvent = new CustomEvent("canvasresized",
            { detail: { width: this._boxSize.width, height: this._boxSize.height } })
        this.dispatchEvent(event)
    }

    resize(containerWidth: number, containerHeight: number) {
        const videoContentSize = this.getRenderedVideoContentSize(this.videoElement, containerWidth, containerHeight)
        this._boxSize = { width: videoContentSize?.displayedWidth || 0, height: videoContentSize?.displayedHeight || 0 }
        if (this.canvas) {
            this.canvas.width = this._boxSize.width
            this.canvas.height = this._boxSize.height
            this.canvas.style.width = this._boxSize.width + "px"
            this.canvas.style.height = this._boxSize.height + "px"
            // log.info("Updating canvas size", { ...this._boxSize })
        }

        const event: Anime4KManagerCanvasResizedEvent = new CustomEvent("canvasresized",
            { detail: { width: this._boxSize.width, height: this._boxSize.height } })
        this.dispatchEvent(event)
    }

    // Adds a function to be called whenever the canvas is created or recreated
    registerOnCanvasCreated(callback: (canvas: HTMLCanvasElement) => void) {
        this._onCanvasCreatedCallbacks.add(callback)
    }

    // Adds a function to be called whenever the canvas is created or recreated
    registerOnCanvasCreatedOnce(callback: (canvas: HTMLCanvasElement) => void) {
        this._onCanvasCreatedCallbacksOnce.add(callback)
    }

    // Select an Anime4K option
    async setOption(option: Anime4KOption, state?: {
        isMiniPlayer: boolean
        isPip: boolean
        seeking: boolean
    }) {

        const previousOption = this._currentOption
        this._currentOption = option

        if (previousOption !== option && option === "off") {
            // log.info("Anime4K turned off")
            this.destroy()
            return
        }

        // Handle change of state
        if (state) {
            // For PIP or mini player, completely destroy the canvas
            if (state.isMiniPlayer || state.isPip) {
                log.info("Destroying canvas due to PIP/mini player mode")
                if (previousOption !== "off") this.destroy()
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
            if (previousOption !== "off") this.destroy()
            try {
                await this._initialize()
            }
            catch (error) {
                log.error("Failed to initialize Anime4K", error)
                this._handleError(error instanceof Error ? error.message : "Unknown error")
            }
            this._onOptionChanged?.(option)
        }

    }

    // initialize the canvas and start rendering

    // Destroy and cleanup resources
    destroy() {
        // this.videoElement.style.opacity = "1"

        this._initialized = false

        if (this._renderStats.renderCallbackId !== null) {
            this.videoElement.cancelVideoFrameCallback(this._renderStats.renderCallbackId)
            this._renderStats.renderCallbackId = null
        }

        this._renderStats = {
            currentFps: 0,
            frameTimeSamples: [],
            lastRenderTime: 0,
            renderCallbackId: null,
        }

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

        if (this._renderStats.renderCallbackId !== null) {
            this.videoElement.cancelVideoFrameCallback(this._renderStats.renderCallbackId)
            this._renderStats.renderCallbackId = null
        }

        this._stopRendering()

        if (this._abortController) {
            this._abortController.abort()
            this._abortController = null
        }

        this._frameDropState.frameDropCount = 0
        this._frameDropState.lastFrameTime = 0

        const event: Anime4KManagerDestroyedEvent = new CustomEvent("destroyed")
        this.dispatchEvent(event)
    }

    // Stop the render loop and release the GPU device it owns.
    // Safe to call at any point, including while initialization is in progress.
    private _stopRendering() {
        this._renderSessionId++
        this._framesInFlight = 0

        if (this._frameCallbackId !== null) {
            this.videoElement.cancelVideoFrameCallback(this._frameCallbackId)
            this._frameCallbackId = null
        }

        this._webgpuResources?.inputTexture?.destroy()
        this._webgpuResources?.context?.unconfigure()
        // Destroying the device also frees the pipeline textures, which expose no disposal of their own
        this._webgpuResources?.device?.destroy()
        this._webgpuResources = null
    }

    // throws if initialization fails
    private async _initialize() {
        if (this._initialized || this._currentOption === "off") {
            return
        }

        log.info("Initializing Anime4K", this._currentOption)

        const event: Anime4KManagerOptionChangedEvent = new CustomEvent("optionchanged", { detail: { newOption: this._currentOption } })
        this.dispatchEvent(event)

        const abortController = new AbortController()
        this._abortController = abortController
        this._frameDropState = {
            ...this._frameDropState,
            frameDropCount: 0,
            totalFrameDrops: 0,
            initTime: performance.now(),
            lastFrameTime: 0,
        }

        this._renderStats = {
            currentFps: 0,
            frameTimeSamples: [],
            lastRenderTime: 0,
            renderCallbackId: null,
        }

        // Check WebGPU support, create canvas, and start rendering
        try {
            const gpuInfo = await this.getGPUInfo()
            if (!gpuInfo) {
                throw new Error("WebGPU not supported")
            }

            if (this._isStale(abortController)) return

            this._createCanvas()

            if (this._isStale(abortController)) return

            await this._startRendering(abortController)

            this._initialized = true
            log.info("Anime4K initialized")
        }
        catch (error) {
            if (!abortController.signal.aborted) {
                log.error("Initialization failed", error)
                throw error
            }
        }
    }

    // destroy() nulls the controller and the next _initialize() installs a fresh one, so reading
    // this._abortController directly would let an already aborted run carry on allocating
    private _isStale(abortController: AbortController): boolean {
        return abortController.signal.aborted || this._abortController !== abortController
    }

    private getRenderedVideoContentSize(video: HTMLVideoElement, containerWidth: number, containerHeight: number) {
        const videoWidth = video.videoWidth
        const videoHeight = video.videoHeight

        if (!videoWidth || !videoHeight) return null // not ready yet

        const containerRatio = containerWidth / containerHeight
        const videoRatio = videoWidth / videoHeight

        let displayedWidth, displayedHeight

        if (videoRatio > containerRatio) {
            displayedWidth = containerWidth
            displayedHeight = containerWidth / videoRatio
        } else {
            displayedHeight = containerHeight
            displayedWidth = containerHeight * videoRatio
        }

        return { displayedWidth, displayedHeight }
    }


    // Create and position the canvas
    private _createCanvas() {
        if (this._abortController?.signal.aborted) return

        const rect = this.videoElement.getBoundingClientRect()
        const videoContentSize = this.getRenderedVideoContentSize(this.videoElement, rect.width, rect.height)
        if (videoContentSize) {
            this._boxSize = { width: videoContentSize.displayedWidth, height: videoContentSize.displayedHeight }
        }

        // Drop any previous canvas, otherwise it stays in the DOM holding its own GPU context
        this.canvas?.remove()

        this.canvas = document.createElement("canvas")

        this.canvas.width = this._boxSize.width
        this.canvas.height = this._boxSize.height
        this.canvas.style.objectFit = "cover"
        this.canvas.style.position = "absolute"
        this.canvas.style.pointerEvents = "none"
        this.canvas.style.zIndex = "2"
        this.canvas.style.objectFit = "contain"
        this.canvas.style.objectPosition = "center"
        this.canvas.style.width = this._boxSize.width + "px"
        this.canvas.style.height = this._boxSize.height + "px"
        this.canvas.style.top = ""
        this.canvas.style.display = "block"
        this.canvas.className = "vc-anime4k-canvas"
        log.info("Creating canvas", { width: this.canvas.width, height: this.canvas.height, top: this.canvas.style.top })

        this.videoElement.parentElement?.appendChild(this.canvas)
        // this.videoElement.style.opacity = "0"
    }

    // WebGPU rendering
    private async _startRendering(abortController: AbortController) {
        if (!this.canvas || !this.videoElement || this._currentOption === "off") {
            console.warn("stopped started")
            return
        }

        const anime4k = await import("anime4k-webgpu")

        const nativeDimensions = {
            width: this.videoElement.videoWidth,
            height: this.videoElement.videoHeight,
        }

        const targetDimensions = {
            width: this.canvas.width,
            height: this.canvas.height,
        }

        log.info("Rendering started")

        // Below is anime4k.render() inlined, because that helper leaks by construction:
        //  - its frame callback reschedules itself unconditionally and it returns no handle, so the
        //    loop runs for the lifetime of the page and keeps the video element, canvas, pipelines
        //    and device reachable even after this manager is destroyed
        //  - it requests its own adapter and device, so overlapping calls leave earlier devices
        //    unreachable and therefore never destroyed
        //  - it submits once per presented frame with no regard for whether the GPU kept up
        // Owning the loop here is what lets _stopRendering() actually stop it.
        const adapter = await navigator.gpu.requestAdapter()
        if (!adapter) throw new Error("WebGPU not supported")

        const device = await adapter.requestDevice()

        if (this._isStale(abortController) || !this.canvas) {
            device.destroy()
            return
        }

        const context = this.canvas.getContext("webgpu")!
        const presentationFormat = navigator.gpu.getPreferredCanvasFormat()
        context.configure({ device, format: presentationFormat, alphaMode: "premultiplied" })

        const inputTexture = device.createTexture({
            size: [nativeDimensions.width, nativeDimensions.height, 1],
            format: "rgba16float",
            usage: GPUTextureUsage.TEXTURE_BINDING | GPUTextureUsage.COPY_DST | GPUTextureUsage.RENDER_ATTACHMENT,
        })

        const pipelines = this.createPipeline({ device, inputTexture, nativeDimensions, targetDimensions }, anime4k)

        const bindGroupLayout = device.createBindGroupLayout({
            label: "Render Bind Group Layout",
            entries: [
                { binding: 1, visibility: GPUShaderStage.FRAGMENT, sampler: {} },
                { binding: 2, visibility: GPUShaderStage.FRAGMENT, texture: {} },
            ],
        })

        const renderPipeline = device.createRenderPipeline({
            layout: device.createPipelineLayout({ bindGroupLayouts: [bindGroupLayout] }),
            vertex: { module: device.createShaderModule({ code: anime4kBlitShader }), entryPoint: "vert_main" },
            fragment: {
                module: device.createShaderModule({ code: anime4kBlitShader }),
                entryPoint: "main",
                targets: [{ format: presentationFormat }],
            },
            primitive: { topology: "triangle-list" },
        })

        const bindGroup = device.createBindGroup({
            layout: bindGroupLayout,
            entries: [
                { binding: 1, resource: device.createSampler({ magFilter: "linear", minFilter: "linear" }) },
                { binding: 2, resource: pipelines.at(-1)!.getOutputTexture().createView() },
            ],
        })

        this._webgpuResources = {
            device,
            pipelines,
            outputTexture: pipelines.at(-1)?.getOutputTexture(),
            context,
            inputTexture,
        }

        // Claim the session only once every resource above exists, so nothing can be half-published
        const sessionId = ++this._renderSessionId

        const renderFrame = () => {
            this._frameCallbackId = null
            if (sessionId !== this._renderSessionId) return

            // Drop the frame rather than queue it when the GPU is behind, see _maxFramesInFlight
            if (this._framesInFlight < this._maxFramesInFlight) {
                this._framesInFlight++

                if (!this.videoElement.paused) {
                    device.queue.copyExternalImageToTexture({ source: this.videoElement }, { texture: inputTexture },
                        [nativeDimensions.width, nativeDimensions.height])
                }

                const encoder = device.createCommandEncoder()
                pipelines.forEach(pipeline => pipeline.pass(encoder))

                const pass = encoder.beginRenderPass({
                    colorAttachments: [{
                        view: context.getCurrentTexture().createView(),
                        clearValue: { r: 0, g: 0, b: 0, a: 1 },
                        loadOp: "clear",
                        storeOp: "store",
                    }],
                })
                pass.setPipeline(renderPipeline)
                pass.setBindGroup(0, bindGroup)
                pass.draw(3)
                pass.end()
                device.queue.submit([encoder.finish()])

                // Rejects if the device goes away before the work completes, so release on both paths.
                // _stopRendering() resets the counter, hence the session check.
                const released = () => {
                    if (sessionId === this._renderSessionId) this._framesInFlight--
                }
                device.queue.onSubmittedWorkDone().then(released, released)
            }

            this._frameCallbackId = this.videoElement.requestVideoFrameCallback(renderFrame)
        }

        this._frameCallbackId = this.videoElement.requestVideoFrameCallback(renderFrame)

        // Deferred, so the session check matters: a teardown inside this window would otherwise
        // resurrect the fps tracker below on a session that no longer exists
        setTimeout(() => {
            requestAnimationFrame(() => {
                if (this.canvas && sessionId === this._renderSessionId) {
                    const rect = this.videoElement.getBoundingClientRect()
                    const videoContentSize = this.getRenderedVideoContentSize(this.videoElement, rect.width, rect.height)
                    if (videoContentSize && (videoContentSize.displayedWidth !== this._boxSize.width || videoContentSize.displayedHeight !== this._boxSize.height)) {
                        this._boxSize = { width: videoContentSize.displayedWidth, height: videoContentSize.displayedHeight }
                        this.canvas.width = this._boxSize.width
                        this.canvas.height = this._boxSize.height
                        this.canvas.style.width = this._boxSize.width + "px"
                        this.canvas.style.height = this._boxSize.height + "px"
                        log.info("Post-init canvas resize correction", { ...this._boxSize })
                    }

                    for (const callback of this._onCanvasCreatedCallbacks) {
                        callback(this.canvas)
                    }
                    for (const callback of this._onCanvasCreatedCallbacksOnce) {
                        callback(this.canvas)
                    }
                    this._onCanvasCreatedCallbacksOnce.clear()

                    const event: Anime4KManagerCanvasCreatedEvent = new CustomEvent("canvascreated", { detail: { canvas: this.canvas } })
                    this.dispatchEvent(event)

                    this._startRenderFpsTracking(sessionId)
                }
            })
        }, 100)

        setTimeout(() => {
            requestAnimationFrame(() => {
                if (this.canvas && sessionId === this._renderSessionId && !this._abortController?.signal.aborted) {
                    const rect = this.videoElement.getBoundingClientRect()
                    const videoContentSize = this.getRenderedVideoContentSize(this.videoElement, rect.width, rect.height)
                    if (videoContentSize && (videoContentSize.displayedWidth !== this._boxSize.width || videoContentSize.displayedHeight !== this._boxSize.height)) {
                        this._boxSize = { width: videoContentSize.displayedWidth, height: videoContentSize.displayedHeight }
                        this.canvas.width = this._boxSize.width
                        this.canvas.height = this._boxSize.height
                        this.canvas.style.width = this._boxSize.width + "px"
                        this.canvas.style.height = this._boxSize.height + "px"
                        log.info("Deferred canvas resize correction", { ...this._boxSize })
                    }
                }
            })
        }, 500)

        // Start frame drop detection if enabled
        if (this._frameDropState.enabled && this._isOptionSelected(this._currentOption)) {
            this._startFrameDropDetection()
        }
    }

    private createPipeline(commonProps: any, anime4k: typeof import("anime4k-webgpu")): [Anime4KPipeline] {
        switch (this._currentOption) {
            case "mode-a":
                return [new anime4k.ModeA(commonProps)]
            case "mode-b":
                return [new anime4k.ModeB(commonProps)]
            case "mode-c":
                return [new anime4k.ModeC(commonProps)]
            case "mode-aa":
                return [new anime4k.ModeAA(commonProps)]
            case "mode-bb":
                return [new anime4k.ModeBB(commonProps)]
            case "mode-ca":
                return [new anime4k.ModeCA(commonProps)]
            case "cnn-2x-medium":
                return [new anime4k.CNNx2M(commonProps)]
            case "cnn-2x-very-large":
                return [new anime4k.CNNx2VL(commonProps)]
            case "denoise-cnn-2x-very-large":
                return [new anime4k.DenoiseCNNx2VL(commonProps)]
            case "cnn-2x-ultra-large":
                return [new anime4k.CNNx2UL(commonProps)]
            case "gan-3x-large":
                return [new anime4k.GANx3L(commonProps)]
            case "gan-4x-ultra-large":
                return [new anime4k.GANx4UUL(commonProps)]
            default:
                return [new anime4k.ModeA(commonProps)]
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
                this._frameDropState.totalFrameDrops++

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
        // Dispatch Fallback Event
        const errorEvent: Anime4KManagerErrorEvent = new CustomEvent("error", { detail: { message: "Performance degraded. Turning off Anime4K." } })
        this.dispatchEvent(errorEvent)

        this.setOption("off")
        this._onOptionChanged?.("off")
    }

    private _handleError(message: string) {
        this._onFallback?.(`Anime4K: ${message}`)
        const errorEvent: Anime4KManagerErrorEvent = new CustomEvent("error", { detail: { message: message } })
        this.dispatchEvent(errorEvent)

        this.setOption("off")
        this._onOptionChanged?.("off")
    }

    // Get GPU information
    // Adapter only: requesting a device here would leak one on every initialization, since this is
    // a support probe and the caller never gets to release it
    private async getGPUInfo() {
        if (!navigator.gpu) return null

        try {
            const adapter = await navigator.gpu.requestAdapter()
            if (!adapter) return null

            const info = (adapter as any).info || {}

            return {
                gpu: info.vendor || info.architecture || "Unknown GPU",
                vendor: info.vendor || "Unknown",
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
            // this.videoElement.style.opacity = "1"
        }
    }

    private _showCanvas() {
        if (this.canvas) {
            this.canvas.style.display = "block"
            // this.videoElement.style.opacity = "0"
        }
    }

    private _isCanvasHidden(): boolean {
        return this.canvas ? this.canvas.style.display === "none" : false
    }

    private _startRenderFpsTracking(sessionId: number) {
        if (!this.videoElement.requestVideoFrameCallback) return

        const trackFrame = (now: number, metadata: VideoFrameCallbackMetadata) => {
            if (this._renderStats.lastRenderTime > 0) {
                const frameTime = now - this._renderStats.lastRenderTime

                this._renderStats.frameTimeSamples.push(frameTime)
                if (this._renderStats.frameTimeSamples.length > 10) {
                    this._renderStats.frameTimeSamples.shift()
                }

                const avgFrameTime = this._renderStats.frameTimeSamples.reduce((a, b) => a + b, 0) / this._renderStats.frameTimeSamples.length
                this._renderStats.currentFps = avgFrameTime > 0 ? 1000 / avgFrameTime : 0
            }

            this._renderStats.lastRenderTime = now

            // destroy() leaves _currentOption untouched, so without the session check this keeps
            // rescheduling forever and holds on to the manager and its video element
            if (sessionId === this._renderSessionId && this._isOptionSelected(this._currentOption)) {
                this._renderStats.renderCallbackId = this.videoElement.requestVideoFrameCallback(trackFrame)
            }
        }

        this._renderStats.renderCallbackId = this.videoElement.requestVideoFrameCallback(trackFrame)
    }
}
