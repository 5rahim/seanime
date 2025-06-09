export const StreamPreviewThumbnailSize = 200
export const StreamPreviewCaptureIntervalSeconds = 12

export class StreamPreviewManager {
    private previewCache: Map<number, string> = new Map()
    private jobs: { current: Job | undefined, pending: Job | undefined } = { current: undefined, pending: undefined }
    private currentMediaSource?: string
    private videoSyncController = new AbortController()
    private videoElement: HTMLVideoElement
    private lastCapturedSegment: number = -1
    private captureThrottleTimeout: number | null = null

    private readonly _dummyVideoElement = document.createElement("video")
    private readonly _offscreenCanvas = new OffscreenCanvas(0, 0)
    private readonly _drawingContext = this._offscreenCanvas.getContext("2d")!

    constructor(videoElement: HTMLVideoElement, mediaSource?: string) {
        this.initializeDummyVideoElement()
        if (mediaSource) {
            this.loadMediaSource(mediaSource)
        }
        this.videoElement = videoElement
        this._bindToVideoPlayer()
    }

    _bindToVideoPlayer(): void {
        this.detachFromCurrentPlayer()
        this.videoSyncController = new AbortController()

        // Only capture previews occasionally during normal playback, not on every timeupdate
        this.videoElement.addEventListener("timeupdate", () => {
            const segmentIndex = this.calculateSegmentIndex(this.videoElement.currentTime)

            // Only capture if we've moved to a new segment and throttle the captures
            if (segmentIndex !== this.lastCapturedSegment && !this.previewCache.has(segmentIndex)) {
                this.throttledCaptureFrame(segmentIndex)
            }
        }, { signal: this.videoSyncController.signal })
    }

    changeMediaSource(newSource?: string): void {
        if (newSource === this.currentMediaSource || !newSource) return

        this.clearPreviewCache()
        this.resetOperationQueue()
        this.loadMediaSource(newSource)
        this.lastCapturedSegment = -1

        // Clear any pending throttled captures
        if (this.captureThrottleTimeout) {
            clearTimeout(this.captureThrottleTimeout)
            this.captureThrottleTimeout = null
        }
    }

    cleanup(): void {
        this._dummyVideoElement.remove()
        this.detachFromCurrentPlayer()
        this.clearPreviewCache()

        // Clear any pending throttled captures
        if (this.captureThrottleTimeout) {
            clearTimeout(this.captureThrottleTimeout)
            this.captureThrottleTimeout = null
        }
    }

    async retrievePreviewForSegment(segmentIndex: number): Promise<string | undefined> {
        const cachedPreview = this.previewCache.get(segmentIndex)
        if (cachedPreview) return cachedPreview

        return await this.schedulePreviewGeneration(segmentIndex)
    }

    private throttledCaptureFrame(segmentIndex: number): void {
        // Clear any pending capture
        if (this.captureThrottleTimeout) {
            clearTimeout(this.captureThrottleTimeout)
        }

        // Throttle captures to avoid spamming
        this.captureThrottleTimeout = window.setTimeout(() => {
            if (!this.previewCache.has(segmentIndex)) {
                this.captureFrameFromCurrentVideo(segmentIndex)
                this.lastCapturedSegment = segmentIndex
            }
            this.captureThrottleTimeout = null
        }, 500) // Wait 500ms before capturing
    }

    private async captureFrameFromCurrentVideo(segmentIndex: number): Promise<void> {
        if (this.videoElement.readyState < 2) return // Not enough data loaded

        const frameWidth = this.videoElement.videoWidth
        const frameHeight = this.videoElement.videoHeight

        if (!frameWidth || !frameHeight) return

        this.configureRenderingSurface(frameWidth, frameHeight)
        this._drawingContext.drawImage(this.videoElement, 0, 0, this._offscreenCanvas.width, this._offscreenCanvas.height)

        const imageBlob = await this._offscreenCanvas.convertToBlob({ type: "image/webp", quality: 0.8 })
        const previewUrl = URL.createObjectURL(imageBlob)

        this.previewCache.set(segmentIndex, previewUrl)
    }

    private initializeDummyVideoElement(): void {
        this._dummyVideoElement.crossOrigin = "anonymous"
        this._dummyVideoElement.playbackRate = 0
        this._dummyVideoElement.muted = true
        this._dummyVideoElement.preload = "none"
    }

    private loadMediaSource(source: string): void {
        this._dummyVideoElement.src = this.currentMediaSource = source
        this._dummyVideoElement.load()
    }

    private detachFromCurrentPlayer(): void {
        this.videoSyncController.abort()

        // Clear any pending throttled captures when detaching
        if (this.captureThrottleTimeout) {
            clearTimeout(this.captureThrottleTimeout)
            this.captureThrottleTimeout = null
        }
    }

    private calculateSegmentIndex(currentTime: number): number {
        return Math.floor(currentTime / StreamPreviewCaptureIntervalSeconds)
    }

    private addJob(segmentIndex: number): Job {
        // @ts-ignore
        const { promise, resolve } = Promise.withResolvers<string | undefined>()

        const execute = (): void => {
            this._dummyVideoElement.requestVideoFrameCallback(async (_timestamp, metadata) => {
                const preview = await this.captureFrameAtSegment(this._dummyVideoElement, segmentIndex, metadata.width, metadata.height)
                resolve(preview)
                this.processNextJob()
            })
            this._dummyVideoElement.currentTime = segmentIndex * StreamPreviewCaptureIntervalSeconds
        }

        return { segmentIndex, execute, promise }
    }

    private processNextJob(): void {
        this.jobs.current = undefined
        if (this.jobs.pending) {
            this.jobs.current = this.jobs.pending
            this.jobs.pending = undefined
            this.jobs.current.execute()
        }
    }

    private schedulePreviewGeneration(segmentIndex: number): Promise<string | undefined> {
        if (!this.jobs.current) {
            this.jobs.current = this.addJob(segmentIndex)
            this.jobs.current.execute()
            return this.jobs.current.promise
        }

        if (this.jobs.current.segmentIndex === segmentIndex) {
            return this.jobs.current.promise
        }

        if (!this.jobs.pending) {
            this.jobs.pending = this.addJob(segmentIndex)
            return this.jobs.pending.promise
        }

        if (this.jobs.pending.segmentIndex === segmentIndex) {
            return this.jobs.pending.promise
        }

        this.jobs.pending = this.addJob(segmentIndex)
        return this.jobs.pending.promise
    }

    private async captureFrameAtSegment(
        videoElement: HTMLVideoElement,
        segmentIndex: number,
        frameWidth = videoElement.videoWidth,
        frameHeight = videoElement.videoHeight,
    ): Promise<string | undefined> {
        const existingPreview = this.previewCache.get(segmentIndex)
        if (existingPreview) return existingPreview

        if (!frameWidth || !frameHeight) return undefined

        this.configureRenderingSurface(frameWidth, frameHeight)
        this._drawingContext.drawImage(videoElement, 0, 0, this._offscreenCanvas.width, this._offscreenCanvas.height)

        const imageBlob = await this._offscreenCanvas.convertToBlob({ type: "image/webp", quality: 0.8 })
        const previewUrl = URL.createObjectURL(imageBlob)

        this.previewCache.set(segmentIndex, previewUrl)
        return previewUrl
    }

    private configureRenderingSurface(sourceWidth: number, sourceHeight: number): void {
        this._offscreenCanvas.width = StreamPreviewThumbnailSize
        this._offscreenCanvas.height = (sourceHeight / sourceWidth) * StreamPreviewThumbnailSize
    }

    private clearPreviewCache(): void {
        this.previewCache.forEach(previewUrl => URL.revokeObjectURL(previewUrl))
        this.previewCache.clear()
    }

    private resetOperationQueue(): void {
        this.jobs.current = undefined
        this.jobs.pending = undefined
    }
}

type Job = {
    segmentIndex: number
    execute: () => void
    promise: Promise<string | undefined>
}
