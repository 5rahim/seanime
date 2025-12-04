export const VIDEOCORE_PREVIEW_THUMBNAIL_SIZE = 200
export const VIDEOCORE_PREVIEW_CAPTURE_INTERVAL_SECONDS = 4
const MAX_CONCURRENT_JOBS = 5
const PREFETCH_AHEAD_COUNT = 10

export class VideoCorePreviewManager {
    private previewCache: Map<number, string> = new Map()
    private inFlightPromises: Map<number, Promise<string | undefined>> = new Map()
    private jobQueue: Job[] = []
    private activeJobs: Set<number> = new Set()
    private currentMediaSource?: string
    private videoSyncController = new AbortController()
    private videoElement: HTMLVideoElement
    private lastCapturedSegment: number = -1
    private captureThrottleTimeout: number | null = null
    private highestCachedIndex: number = -1
    private canvasSizeConfigured: boolean = false

    private readonly _dummyVideoElement = document.createElement("video")
    private readonly _offscreenCanvas = new OffscreenCanvas(0, 0)
    private readonly _drawingContext = this._offscreenCanvas.getContext("2d", {
        alpha: false,
        desynchronized: true,
    })!

    constructor(videoElement: HTMLVideoElement, mediaSource?: string) {
        this.initializeDummyVideoElement()
        // Non-HLS streams will use _dummyVideoElement
        if (mediaSource) {
            this.loadMediaSource(mediaSource + "&thumbnail=true")
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

    cleanup(): void {
        this.detachFromCurrentPlayer()
        this.resetOperationQueue()
        this.clearPreviewCache()

        if (this.captureThrottleTimeout) {
            clearTimeout(this.captureThrottleTimeout)
            this.captureThrottleTimeout = null
        }

        this._dummyVideoElement.pause()
        this._dummyVideoElement.removeAttribute("src")
        this._dummyVideoElement.load()
        this._dummyVideoElement.remove()

        this.currentMediaSource = undefined
        this.lastCapturedSegment = -1
        this.highestCachedIndex = -1
        this.canvasSizeConfigured = false
    }

    async retrievePreviewForSegment(segmentIndex: number): Promise<string | undefined> {
        const cachedPreview = this.previewCache.get(segmentIndex)
        if (cachedPreview) {
            // Prefetch upcoming segments in the background
            this.prefetchUpcomingSegments(segmentIndex)
            return cachedPreview
        }

        const inFlight = this.inFlightPromises.get(segmentIndex)
        if (inFlight) return inFlight

        const promise = this.schedulePreviewGeneration(segmentIndex)

        // Also prefetch nearby segments
        this.prefetchUpcomingSegments(segmentIndex)

        return promise
    }

    private prefetchUpcomingSegments(currentIndex: number): void {
        // Prefetch next segments
        for (let i = 1; i <= PREFETCH_AHEAD_COUNT; i++) {
            const nextIndex = currentIndex + i
            if (!this.previewCache.has(nextIndex) && !this.inFlightPromises.has(nextIndex)) {
                this.schedulePreviewGeneration(nextIndex).catch(() => {
                    // Ignore prefetch errors
                })
            }
        }
    }

    getLastestCachedIndex(): number {
        return this.highestCachedIndex
    }

    calculateTimeFromIndex(segmentIndex: number): number {
        return segmentIndex * VIDEOCORE_PREVIEW_CAPTURE_INTERVAL_SECONDS
    }

    private throttledCaptureFrame(segmentIndex: number): void {
        // Clear any pending capture
        if (this.captureThrottleTimeout) {
            clearTimeout(this.captureThrottleTimeout)
        }

        // Throttle captures to avoid spamming
        this.captureThrottleTimeout = window.setTimeout(() => {
            console.warn("Capturing preview for segment", segmentIndex)
            if (!this.previewCache.has(segmentIndex) && !this.inFlightPromises.has(segmentIndex)) {
                const promise = this.captureFrameFromCurrentVideo(segmentIndex)
                this.inFlightPromises.set(segmentIndex, promise)
                promise.finally(() => this.inFlightPromises.delete(segmentIndex))
                this.lastCapturedSegment = segmentIndex
            }
            this.captureThrottleTimeout = null
        }, 500) // Wait 500ms before capturing
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
        return Math.floor(currentTime / VIDEOCORE_PREVIEW_CAPTURE_INTERVAL_SECONDS)
    }

    private async captureFrameFromCurrentVideo(segmentIndex: number): Promise<string | undefined> {
        if (this.videoElement.readyState < 2) return

        const frameWidth = this.videoElement.videoWidth
        const frameHeight = this.videoElement.videoHeight

        if (!frameWidth || !frameHeight) return

        try {
            this.configureRenderingSurface(frameWidth, frameHeight)
            this._drawingContext.drawImage(this.videoElement, 0, 0, this._offscreenCanvas.width, this._offscreenCanvas.height)

            const imageBlob = await this._offscreenCanvas.convertToBlob({ type: "image/webp", quality: 0.6 })
            const previewUrl = URL.createObjectURL(imageBlob)

            this.previewCache.set(segmentIndex, previewUrl)
            if (segmentIndex > this.highestCachedIndex) {
                this.highestCachedIndex = segmentIndex
            }
            return previewUrl
        }
        catch (error) {
            if (error instanceof DOMException && error.name === "SecurityError") {
                console.warn("Cannot capture preview: CORS restrictions prevent canvas export")
                return undefined
            }
            throw error
        }
    }

    private addJob(segmentIndex: number): Job {
        // @ts-ignore
        const { promise, resolve } = Promise.withResolvers<string | undefined>()

        const execute = async (): Promise<void> => {
            try {
                this.activeJobs.add(segmentIndex)

                // seek and wait
                this._dummyVideoElement.currentTime = segmentIndex * VIDEOCORE_PREVIEW_CAPTURE_INTERVAL_SECONDS

                // Wait for seek to complete
                await new Promise<void>((seekResolve) => {
                    const onSeeked = () => {
                        this._dummyVideoElement.removeEventListener("seeked", onSeeked)
                        seekResolve()
                    }
                    this._dummyVideoElement.addEventListener("seeked", onSeeked, { once: true })
                })

                const preview = await this.captureFrameAtSegment(
                    this._dummyVideoElement,
                    segmentIndex,
                    this._dummyVideoElement.videoWidth,
                    this._dummyVideoElement.videoHeight,
                )
                resolve(preview)
            }
            catch (error) {
                resolve(undefined)
            }
            finally {
                this.activeJobs.delete(segmentIndex)
                this.processNextJob()
            }
        }

        return { segmentIndex, execute, promise }
    }

    private processNextJob(): void {
        // Process multiple jobs concurrently
        while (this.activeJobs.size < MAX_CONCURRENT_JOBS && this.jobQueue.length > 0) {
            const job = this.jobQueue.shift()
            if (job) {
                job.execute()
            }
        }
    }

    private schedulePreviewGeneration(segmentIndex: number): Promise<string | undefined> {
        // Check if already in flight
        const existingPromise = this.inFlightPromises.get(segmentIndex)
        if (existingPromise) {
            return existingPromise
        }

        // Check if already in queue
        const existingJob = this.jobQueue.find(j => j.segmentIndex === segmentIndex)
        if (existingJob) {
            return existingJob.promise
        }

        // Create new job
        const job = this.addJob(segmentIndex)
        this.inFlightPromises.set(segmentIndex, job.promise)

        job.promise.finally(() => {
            this.inFlightPromises.delete(segmentIndex)
        })

        // Add to queue or execute immediately
        if (this.activeJobs.size < MAX_CONCURRENT_JOBS) {
            job.execute()
        } else {
            this.jobQueue.push(job)
        }

        return job.promise
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

        try {
            this.configureRenderingSurface(frameWidth, frameHeight)
            this._drawingContext.drawImage(videoElement, 0, 0, this._offscreenCanvas.width, this._offscreenCanvas.height)

            const imageBlob = await this._offscreenCanvas.convertToBlob({ type: "image/webp", quality: 0.6 })
            const previewUrl = URL.createObjectURL(imageBlob)

            this.previewCache.set(segmentIndex, previewUrl)
            if (segmentIndex > this.highestCachedIndex) {
                this.highestCachedIndex = segmentIndex
            }
            return previewUrl
        }
        catch (error) {
            if (error instanceof DOMException && error.name === "SecurityError") {
                console.warn("Cannot capture preview: CORS restrictions prevent canvas export")
                return undefined
            }
            throw error
        }
    }

    private configureRenderingSurface(sourceWidth: number, sourceHeight: number): void {
        // Only resize canvas if needed
        if (!this.canvasSizeConfigured || this._offscreenCanvas.width !== VIDEOCORE_PREVIEW_THUMBNAIL_SIZE) {
            this._offscreenCanvas.width = VIDEOCORE_PREVIEW_THUMBNAIL_SIZE
            this._offscreenCanvas.height = (sourceHeight / sourceWidth) * VIDEOCORE_PREVIEW_THUMBNAIL_SIZE
            this.canvasSizeConfigured = true
        }
    }

    private clearPreviewCache(): void {
        this.previewCache.forEach(previewUrl => URL.revokeObjectURL(previewUrl))
        this.previewCache.clear()
        this.inFlightPromises.clear()
    }

    private resetOperationQueue(): void {
        this.jobQueue = []
        this.activeJobs.clear()
    }
}

type Job = {
    segmentIndex: number
    execute: () => Promise<void>
    promise: Promise<string | undefined>
}
