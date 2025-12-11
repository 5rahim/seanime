import { VideoCore_VideoPlaybackInfo } from "@/app/(main)/_features/video-core/video-core.atoms"
import { logger } from "@/lib/helpers/debug"
import Hls from "hls.js"

const log = logger("VIDEO CORE PREVIEW")

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
    private hlsInstance: Hls | null = null
    private isHlsSource: boolean = false
    private lastKnownTime: number = 0

    private readonly _dummyVideoElement = document.createElement("video")
    private readonly _offscreenCanvas = new OffscreenCanvas(0, 0)
    private readonly _drawingContext = this._offscreenCanvas.getContext("2d", {
        alpha: false,
        desynchronized: true,
    })!

    constructor(
        videoElement: HTMLVideoElement,
        mediaSource: string,
        streamType: VideoCore_VideoPlaybackInfo["streamType"],
        useCustomThumbnailRequest?: boolean,
    ) {
        this.initializeDummyVideoElement()
        this.videoElement = videoElement

        this.isHlsSource = streamType === "hls"
        this.loadMediaSource(
            mediaSource + (useCustomThumbnailRequest ? "&thumbnail=true" : ""),
        )

        this._bindToVideoPlayer()
    }

    _bindToVideoPlayer(): void {
        this.detachFromCurrentPlayer()
        this.videoSyncController = new AbortController()

        // reset lastCapturedSegment to allow capture at new position
        this.videoElement.addEventListener("seeked", () => {
            const currentSegment = this.calculateSegmentIndex(this.videoElement.currentTime)
            const previousSegment = this.calculateSegmentIndex(this.lastKnownTime)

            // If we seeked more than 1 segment away, reset to allow fresh captures
            if (Math.abs(currentSegment - previousSegment) > 1) {
                log.info("Seek detected, resetting lastCapturedSegment")
                this.lastCapturedSegment = -1

                // Cancel any pending throttled capture since we've moved to a new position
                if (this.captureThrottleTimeout) {
                    clearTimeout(this.captureThrottleTimeout)
                    this.captureThrottleTimeout = null
                }
            }

            this.lastKnownTime = this.videoElement.currentTime
        }, { signal: this.videoSyncController.signal })

        this.videoElement.addEventListener("timeupdate", () => {
            const segmentIndex = this.calculateSegmentIndex(this.videoElement.currentTime)

            // Update last known time for seek detection
            this.lastKnownTime = this.videoElement.currentTime

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

        if (this.hlsInstance) {
            this.hlsInstance.destroy()
            this.hlsInstance = null
        }

        this._dummyVideoElement.pause()
        this._dummyVideoElement.removeAttribute("src")
        this._dummyVideoElement.load()
        this._dummyVideoElement.remove()

        this.currentMediaSource = undefined
        this.lastCapturedSegment = -1
        this.highestCachedIndex = -1
        this.canvasSizeConfigured = false
        this.isHlsSource = false
        this.lastKnownTime = 0
    }

    async retrievePreviewForSegment(
        segmentIndex: number,
    ): Promise<string | undefined> {
        const cachedPreview = this.previewCache.get(segmentIndex)
        if (cachedPreview) {
            this.prefetchUpcomingSegments(segmentIndex)
            return cachedPreview
        }

        const inFlight = this.inFlightPromises.get(segmentIndex)
        if (inFlight) return inFlight

        const promise = this.schedulePreviewGeneration(segmentIndex)
        this.prefetchUpcomingSegments(segmentIndex)

        return promise
    }

    private loadMediaSource(source: string): void {
        this.currentMediaSource = source

        if (this.hlsInstance) {
            this.hlsInstance.destroy()
            this.hlsInstance = null
        }

        if (this.isHlsSource) {
            if (Hls.isSupported()) {
                this.hlsInstance = new Hls({
                    enableWorker: false,
                    lowLatencyMode: false,
                    backBufferLength: 0,
                    maxBufferLength: 5,
                    maxMaxBufferLength: 10,
                })
                this.hlsInstance.loadSource(source)
                this.hlsInstance.attachMedia(this._dummyVideoElement)
            } else if (this._dummyVideoElement.canPlayType("application/vnd.apple.mpegurl")) {
                this._dummyVideoElement.src = source
                this._dummyVideoElement.load()
            } else {
                log.warning("HLS is not supported in this browser for thumbnails")
            }
        } else {
            this._dummyVideoElement.src = source
            this._dummyVideoElement.load()
        }
    }

    private prefetchUpcomingSegments(currentIndex: number): void {
        for (let i = 1; i <= PREFETCH_AHEAD_COUNT; i++) {
            const nextIndex = currentIndex + i
            if (
                !this.previewCache.has(nextIndex) &&
                !this.inFlightPromises.has(nextIndex)
            ) {
                this.schedulePreviewGeneration(nextIndex).catch(() => {})
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
        if (this.captureThrottleTimeout) {
            clearTimeout(this.captureThrottleTimeout)
        }

        this.captureThrottleTimeout = window.setTimeout(async () => {
            // log.info("Capturing preview for segment", segmentIndex)
            if (!this.previewCache.has(segmentIndex) && !this.inFlightPromises.has(segmentIndex)) {
                const promise = this.captureFrameFromCurrentVideo(segmentIndex)
                this.inFlightPromises.set(segmentIndex, promise)

                try {
                    const result = await promise
                    // Only update lastCapturedSegment if capture succeeded
                    if (result) {
                        this.lastCapturedSegment = segmentIndex
                    }
                    // If capture failed, don't update lastCapturedSegment to allow retry
                }
                finally {
                    this.inFlightPromises.delete(segmentIndex)
                }
            }
            this.captureThrottleTimeout = null
        }, 300)
    }

    private initializeDummyVideoElement(): void {
        this._dummyVideoElement.crossOrigin = "anonymous"
        this._dummyVideoElement.playbackRate = 0
        this._dummyVideoElement.muted = true
        this._dummyVideoElement.preload = "metadata"
    }

    private detachFromCurrentPlayer(): void {
        this.videoSyncController.abort()

        if (this.captureThrottleTimeout) {
            clearTimeout(this.captureThrottleTimeout)
            this.captureThrottleTimeout = null
        }
    }

    private calculateSegmentIndex(currentTime: number): number {
        return Math.floor(currentTime / VIDEOCORE_PREVIEW_CAPTURE_INTERVAL_SECONDS)
    }

    private async captureFrameFromCurrentVideo(
        segmentIndex: number,
    ): Promise<string | undefined> {
        // Wait for video to stabilize after seek
        if (this.videoElement.readyState < 2) {
            try {
                await this.waitForReadyState(this.videoElement, 2, 5000)
            }
            catch {
                // Don't update lastCapturedSegment on failure to allow retry
                // log.info("Video not ready for capture, skipping segment", segmentIndex)
                return undefined
            }
        }

        const frameWidth = this.videoElement.videoWidth
        const frameHeight = this.videoElement.videoHeight

        if (!frameWidth || !frameHeight) return undefined

        try {
            this.configureRenderingSurface(frameWidth, frameHeight)
            this._drawingContext.drawImage(
                this.videoElement,
                0,
                0,
                this._offscreenCanvas.width,
                this._offscreenCanvas.height,
            )

            const imageBlob = await this._offscreenCanvas.convertToBlob({
                type: "image/webp",
                quality: 0.6,
            })
            const previewUrl = URL.createObjectURL(imageBlob)

            this.previewCache.set(segmentIndex, previewUrl)
            if (segmentIndex > this.highestCachedIndex) {
                this.highestCachedIndex = segmentIndex
            }
            return previewUrl
        }
        catch (error) {
            if (error instanceof DOMException && error.name === "SecurityError") {
                log.warning(
                    "Cannot capture preview: CORS restrictions prevent canvas export",
                )
                return undefined
            }
            throw error
        }
    }

    private waitForReadyState(
        video: HTMLVideoElement,
        minState: number,
        timeout: number,
    ): Promise<void> {
        return new Promise((resolve, reject) => {
            if (video.readyState >= minState) {
                resolve()
                return
            }

            const timeoutId = setTimeout(() => {
                cleanup()
                reject(new Error("Timeout waiting for ready state"))
            }, timeout)

            const onCanPlay = () => {
                if (video.readyState >= minState) {
                    cleanup()
                    resolve()
                }
            }

            const cleanup = () => {
                clearTimeout(timeoutId)
                video.removeEventListener("canplay", onCanPlay)
                video.removeEventListener("canplaythrough", onCanPlay)
            }

            video.addEventListener("canplay", onCanPlay)
            video.addEventListener("canplaythrough", onCanPlay)
        })
    }

    private addJob(segmentIndex: number): Job {
        // @ts-expect-error Promise.withResolvers
        const { promise, resolve } = Promise.withResolvers<string | undefined>()

        const execute = async (): Promise<void> => {
            try {
                this.activeJobs.add(segmentIndex)

                const targetTime = segmentIndex * VIDEOCORE_PREVIEW_CAPTURE_INTERVAL_SECONDS

                await this.waitForVideoReady()

                this._dummyVideoElement.currentTime = targetTime

                await this.waitForSeek()

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

    private waitForVideoReady(): Promise<void> {
        return new Promise((resolve, reject) => {
            if (this._dummyVideoElement.readyState >= 1) {
                resolve()
                return
            }

            const timeout = setTimeout(() => {
                cleanup()
                reject(new Error("Timeout waiting for video ready"))
            }, 10000)

            const onLoadedMetadata = () => {
                cleanup()
                resolve()
            }

            const onError = () => {
                cleanup()
                reject(new Error("Video failed to load"))
            }

            const cleanup = () => {
                clearTimeout(timeout)
                this._dummyVideoElement.removeEventListener(
                    "loadedmetadata",
                    onLoadedMetadata,
                )
                this._dummyVideoElement.removeEventListener("error", onError)
            }

            this._dummyVideoElement.addEventListener("loadedmetadata", onLoadedMetadata, { once: true })
            this._dummyVideoElement.addEventListener("error", onError, { once: true })
        })
    }

    private waitForSeek(): Promise<void> {
        return new Promise((resolve, reject) => {
            const timeout = setTimeout(() => {
                cleanup()
                reject(new Error("Seek timeout"))
            }, 5000)

            const onSeeked = () => {
                cleanup()
                resolve()
            }

            const onError = () => {
                cleanup()
                reject(new Error("Seek failed"))
            }

            const cleanup = () => {
                clearTimeout(timeout)
                this._dummyVideoElement.removeEventListener("seeked", onSeeked)
                this._dummyVideoElement.removeEventListener("error", onError)
            }

            this._dummyVideoElement.addEventListener("seeked", onSeeked, {
                once: true,
            })
            this._dummyVideoElement.addEventListener("error", onError, { once: true })
        })
    }

    private processNextJob(): void {
        while (
            this.activeJobs.size < MAX_CONCURRENT_JOBS &&
            this.jobQueue.length > 0
            ) {
            const job = this.jobQueue.shift()
            if (job) {
                job.execute()
            }
        }
    }

    private schedulePreviewGeneration(
        segmentIndex: number,
    ): Promise<string | undefined> {
        const existingPromise = this.inFlightPromises.get(segmentIndex)
        if (existingPromise) {
            return existingPromise
        }

        const existingJob = this.jobQueue.find((j) => j.segmentIndex === segmentIndex)
        if (existingJob) {
            return existingJob.promise
        }

        const job = this.addJob(segmentIndex)
        this.inFlightPromises.set(segmentIndex, job.promise)

        job.promise.finally(() => {
            this.inFlightPromises.delete(segmentIndex)
        })

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

        if (videoElement.readyState < 2) {
            try {
                await this.waitForReadyState(videoElement, 2, 3000)
            }
            catch {
                return undefined
            }
        }

        try {
            this.configureRenderingSurface(frameWidth, frameHeight)
            this._drawingContext.drawImage(
                videoElement,
                0,
                0,
                this._offscreenCanvas.width,
                this._offscreenCanvas.height,
            )

            const imageBlob = await this._offscreenCanvas.convertToBlob({
                type: "image/webp",
                quality: 0.6,
            })
            const previewUrl = URL.createObjectURL(imageBlob)

            this.previewCache.set(segmentIndex, previewUrl)
            if (segmentIndex > this.highestCachedIndex) {
                this.highestCachedIndex = segmentIndex
            }
            return previewUrl
        }
        catch (error) {
            if (error instanceof DOMException && error.name === "SecurityError") {
                log.warning("Cannot capture preview: CORS restrictions prevent canvas export")
                return undefined
            }
            throw error
        }
    }

    private configureRenderingSurface(
        sourceWidth: number,
        sourceHeight: number,
    ): void {
        const targetHeight = (sourceHeight / sourceWidth) * VIDEOCORE_PREVIEW_THUMBNAIL_SIZE

        if (
            !this.canvasSizeConfigured ||
            this._offscreenCanvas.width !== VIDEOCORE_PREVIEW_THUMBNAIL_SIZE ||
            this._offscreenCanvas.height !== targetHeight
        ) {
            this._offscreenCanvas.width = VIDEOCORE_PREVIEW_THUMBNAIL_SIZE
            this._offscreenCanvas.height = targetHeight
            this.canvasSizeConfigured = true
        }
    }

    private clearPreviewCache(): void {
        this.previewCache.forEach((previewUrl) => URL.revokeObjectURL(previewUrl))
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
