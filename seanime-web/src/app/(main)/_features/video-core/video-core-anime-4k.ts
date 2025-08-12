import { vc_miniPlayer, vc_pip, vc_realVideoSize, vc_seeking, vc_videoElement } from "@/app/(main)/_features/video-core/video-core"
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
import { atom, useAtomValue } from "jotai"
import { useAtom, useSetAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React from "react"
import { useUpdateEffect } from "react-use"
import { vc_doFlashAction } from "./video-core-action-display"

type Anime4KOption =
    | "off"
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

export const vc_anime4kOption = atomWithStorage<Anime4KOption>("sea-video-core-anime4k", "off", undefined, { getOnInit: true })

const frameDropDetectionAtom = atom({
    enabled: true,
    frameDropThreshold: 5, // number of consecutive frame drops before fallback
    frameDropCount: 0,
    lastFrameTime: 0,
    targetFrameTime: 1000 / 30, // 40fps target
    performanceGracePeriod: 1000, // grace period after initialization
    initTime: 0,
})

// maintain a single canvas
// destroy when window resizes
export const vc_anime4kCanvas = atom<HTMLCanvasElement | null>(null)

export const useVideoCoreAnime4K = () => {
    const video = useAtomValue(vc_videoElement)
    const realVideoSize = useAtomValue(vc_realVideoSize)
    const seeking = useAtomValue(vc_seeking)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const isPip = useAtomValue(vc_pip)

    const flashAction = useSetAtom(vc_doFlashAction)

    const [canvas, setCanvas] = useAtom(vc_anime4kCanvas)
    const [frameDropState, setFrameDropState] = useAtom(frameDropDetectionAtom)

    const [selectedOption, setAnime4kOption] = useAtom(vc_anime4kOption)

    const canvasCreated = React.useRef(false)
    const renderLoopRef = React.useRef<number | null>(null)
    const originalOptionRef = React.useRef<Anime4KOption>("off")
    const currentOptionRef = React.useRef<Anime4KOption>(selectedOption)
    const webgpuResourcesRef = React.useRef<{ device?: GPUDevice; pipelines?: any[] } | null>(null)
    const initializingRef = React.useRef(false)
    const abortControllerRef = React.useRef<AbortController | null>(null)

    React.useEffect(() => {
        currentOptionRef.current = selectedOption
    }, [selectedOption])

    const detectFrameDrops = React.useCallback(() => {
        if (currentOptionRef.current === "off") {
            return
        }

        setFrameDropState(prev => {
            const now = performance.now()
            const timeSinceInit = now - prev.initTime

            // skip detection during grace period
            if (timeSinceInit < prev.performanceGracePeriod) {
                return { ...prev, lastFrameTime: now }
            }

            if (prev.lastFrameTime > 0) {
                const frameTime = now - prev.lastFrameTime
                const isFrameDrop = frameTime > prev.targetFrameTime * 1.5 // 50% tolerance

                if (isFrameDrop) {
                    const newDropCount = prev.frameDropCount + 1

                    if (newDropCount >= prev.frameDropThreshold) {
                        setAnime4kOption("off")
                        flashAction({ message: "Performance degraded. Turning off Anime4K.", duration: 2000 })
                        console.warn(`Anime4K: Detected ${newDropCount} consecutive frame drops. Falling back to 'off' mode.`)
                        destroyCanvas()
                        return {
                            ...prev,
                            frameDropCount: 0,
                            lastFrameTime: now,
                        }
                    }

                    return {
                        ...prev,
                        frameDropCount: newDropCount,
                        lastFrameTime: now,
                    }
                } else {
                    // reset on successful frame
                    return {
                        ...prev,
                        frameDropCount: 0,
                        lastFrameTime: now,
                    }
                }
            } else {
                return { ...prev, lastFrameTime: now }
            }
        })
    }, [])

    function destroyCanvas() {
        if (canvas) {
            canvas.remove()
            setCanvas(null)
            canvasCreated.current = false
        }
        if (renderLoopRef.current) {
            cancelAnimationFrame(renderLoopRef.current)
            renderLoopRef.current = null
        }
        if (webgpuResourcesRef.current?.device) {
            webgpuResourcesRef.current.device.destroy()
            webgpuResourcesRef.current = null
        }
        if (abortControllerRef.current) {
            abortControllerRef.current.abort()
            abortControllerRef.current = null
        }
        initializingRef.current = false
    }

    const timeoutRef = React.useRef<NodeJS.Timeout | null>(null)

    React.useEffect(() => {
        const init = async () => {
            if (initializingRef.current) {
                return
            }

            // if already created, destroy the canvas
            if (canvasCreated || seeking || selectedOption === "off" || isMiniPlayer || isPip) {
                destroyCanvas()
            }
            if (timeoutRef.current) {
                clearTimeout(timeoutRef.current)
            }

            if (!video || seeking || selectedOption === "off" || isMiniPlayer || isPip) return

            initializingRef.current = true
            abortControllerRef.current = new AbortController()

            try {
                // frame drop detection state
                setFrameDropState(prev => ({
                    ...prev,
                    frameDropCount: 0,
                    initTime: performance.now(),
                    lastFrameTime: 0,
                }))

                // original option for potential restoration
                if ((selectedOption as Anime4KOption) !== "off") {
                    originalOptionRef.current = selectedOption
                }

                if (abortControllerRef.current?.signal.aborted) return

                // get gpu info, if not available, fallback to off
                const gpuInfo = await getGPUInfo()
                if (abortControllerRef.current?.signal.aborted) return

                if (!gpuInfo) {
                    setAnime4kOption("off")
                    flashAction({ message: "Anime4K: WebGPU not supported.", duration: 2000 })
                    destroyCanvas()
                    return
                }

                // Check if aborted before creating canvas
                if (abortControllerRef.current?.signal.aborted) return

                // create canvas
                timeoutRef.current = setTimeout(() => {
                    if (!video || abortControllerRef.current?.signal.aborted) return
                    const canvas = document.createElement("canvas")
                    canvas.width = realVideoSize.width
                    canvas.height = realVideoSize.height
                    canvas.style.objectFit = "cover"
                    canvas.style.position = "absolute"
                    canvas.style.top = video.getBoundingClientRect().top + "px"
                    canvas.style.left = "0"
                    canvas.style.right = "0"
                    canvas.style.pointerEvents = "none"
                    canvas.style.zIndex = "2"
                    canvas.className = "vc-anime4k-canvas"
                    canvasCreated.current = true
                    video.parentElement?.appendChild(canvas)
                    setCanvas(canvas)
                    initializingRef.current = false
                }, 1000)
            }
            catch (error) {
                if (!abortControllerRef.current?.signal.aborted) {
                    console.error("Anime4K initialization error:", error)
                }
                initializingRef.current = false
            }
        }

        init()

        return () => {
            if (timeoutRef.current) clearTimeout(timeoutRef.current)
            destroyCanvas()
        }
    }, [selectedOption, video, realVideoSize, seeking, isMiniPlayer, isPip])

    useUpdateEffect(() => {
        // clean up any existing frame detection loop
        if (renderLoopRef.current) {
            cancelAnimationFrame(renderLoopRef.current)
            renderLoopRef.current = null
        }

        if (!canvas || !video || selectedOption === "off" || !canvasCreated.current) return

        const nativeDimensions = {
            width: video.videoWidth,
            height: video.videoHeight,
        }

        const targetDimensions = {
            width: canvas.width,
            height: canvas.height,
        }

        async function init() {
            try {
                const renderResult = await render({
                    video: video!,
                    canvas: canvas!,
                    pipelineBuilder: (device, inputTexture) => {
                        webgpuResourcesRef.current = { device }

                        const commonProps = {
                            device,
                            inputTexture,
                            nativeDimensions,
                            targetDimensions,
                        }

                        switch (selectedOption) {
                            case "mode-a":
                                return [new ModeA(commonProps)] as [Anime4KPipeline]

                            case "mode-b":
                                return [new ModeB(commonProps)] as [Anime4KPipeline]

                            case "mode-c":
                                return [new ModeC(commonProps)] as [Anime4KPipeline]

                            case "mode-aa":
                                return [new ModeAA(commonProps)] as [Anime4KPipeline]

                            case "mode-bb":
                                return [new ModeBB(commonProps)] as [Anime4KPipeline]

                            case "mode-ca":
                                return [new ModeCA(commonProps)] as [Anime4KPipeline]

                            case "cnn-2x-medium":
                                return [new CNNx2M(commonProps)] as [Anime4KPipeline]

                            case "cnn-2x-very-large":
                                return [new CNNx2VL(commonProps)] as [Anime4KPipeline]

                            case "denoise-cnn-2x-very-large":
                                return [new DenoiseCNNx2VL(commonProps)] as [Anime4KPipeline]

                            case "cnn-2x-ultra-large":
                                return [new CNNx2UL(commonProps)] as [Anime4KPipeline]

                            case "gan-3x-large":
                                return [new GANx3L(commonProps)] as [Anime4KPipeline]

                            case "gan-4x-ultra-large":
                                return [new GANx4UUL(commonProps)] as [Anime4KPipeline]

                            default:
                                // fallback to Mode A
                                return [new ModeA(commonProps)] as [Anime4KPipeline]
                        }
                    },
                })

                if (frameDropState.enabled && selectedOption !== "off") {
                    const frameDetectionLoop = () => {
                        if (currentOptionRef.current !== "off") {
                            detectFrameDrops()
                            renderLoopRef.current = requestAnimationFrame(frameDetectionLoop)
                        }
                    }
                    renderLoopRef.current = requestAnimationFrame(frameDetectionLoop)
                }
            }
            catch (error: unknown) {
                console.warn("Anime4K initialization failed:", error)
                if (error instanceof Error) {
                    flashAction({ message: `Anime4K: Initialization failed: ${error.message}`, duration: 2000 })
                }
                // fallback to off on failure
                setAnime4kOption("off")
            }
        }

        init()

        return () => {
            if (renderLoopRef.current) {
                cancelAnimationFrame(renderLoopRef.current)
                renderLoopRef.current = null
            }
        }
    }, [canvas, selectedOption, frameDropState.enabled])

}

export const anime4kOptions: { value: Anime4KOption; label: string; description: string; performance: "light" | "medium" | "heavy" }[] = [
    { value: "off", label: "Off", description: "Disabled", performance: "light" },
    { value: "mode-a", label: "Mode A", description: "Removes compression artifacts then upscales", performance: "light" },
    { value: "mode-b", label: "Mode B", description: "Gentle artifact removal then upscales", performance: "light" },
    { value: "mode-c", label: "Mode C", description: "Upscales with denoising then upscales again", performance: "light" },
    { value: "mode-aa", label: "Mode A+A", description: "Enhanced restoration for better quality", performance: "medium" },
    { value: "mode-bb", label: "Mode B+B", description: "Double soft restoration", performance: "medium" },
    { value: "mode-ca", label: "Mode C+A", description: "Denoising + restoration hybrid", performance: "medium" },
    { value: "cnn-2x-medium", label: "CNN 2x M", description: "Balanced speed and quality", performance: "medium" },
    { value: "cnn-2x-very-large", label: "CNN 2x VL", description: "High quality neural network", performance: "heavy" },
    { value: "denoise-cnn-2x-very-large", label: "Denoise CNN 2x VL", description: "Removes noise while upscaling", performance: "heavy" },
    { value: "cnn-2x-ultra-large", label: "CNN 2x UL", description: "Maximum CNN quality", performance: "heavy" },
    { value: "gan-3x-large", label: "GAN 3x L", description: "Generative adversarial network for perceptual quality", performance: "heavy" },
    { value: "gan-4x-ultra-large", label: "GAN 4x UL", description: "Maximum upscaling with GAN technology", performance: "heavy" },
]

export const getAnime4KOptionByValue = (value: Anime4KOption) => {
    return anime4kOptions.find(option => option.value === value)
}

export const getRecommendedAnime4KOptions = (videoResolution: { width: number; height: number }) => {
    const is720pOrLower = videoResolution.height <= 720
    const is1080pOrLower = videoResolution.height <= 1080

    if (is720pOrLower) {
        return anime4kOptions.filter(option =>
            ["mode-a", "mode-b", "mode-aa", "mode-bb", "cnn-2x-medium", "cnn-2x-very-large"].includes(option.value),
        )
    } else if (is1080pOrLower) {
        return anime4kOptions.filter(option =>
            ["mode-a", "mode-b", "mode-c", "cnn-2x-medium"].includes(option.value),
        )
    } else {
        return anime4kOptions.filter(option =>
            ["mode-a", "mode-b", "mode-c"].includes(option.value),
        )
    }
}

export const getPerformanceRecommendation = (gpu?: string) => {
    const isHighEnd = gpu && (
        gpu.includes("RTX 40") ||
        gpu.includes("RTX 3080") ||
        gpu.includes("RTX 3090") ||
        gpu.includes("RX 6800") ||
        gpu.includes("RX 6900") ||
        gpu.includes("M1 Pro") ||
        gpu.includes("M1 Max") ||
        gpu.includes("M2") ||
        gpu.includes("M3")
    )

    const isMidRange = gpu && (
        gpu.includes("RTX 30") ||
        gpu.includes("RTX 20") ||
        gpu.includes("GTX 16") ||
        gpu.includes("RX 6600") ||
        gpu.includes("RX 5") ||
        gpu.includes("M1")
    )

    if (isHighEnd) {
        return {
            maxPerformance: "heavy" as const,
            recommendedOptions: anime4kOptions.filter(opt => opt.performance !== "heavy").slice(0, 8),
        }
    } else if (isMidRange) {
        return {
            maxPerformance: "medium" as const,
            recommendedOptions: anime4kOptions.filter(opt => opt.performance === "light" || opt.performance === "medium"),
        }
    } else {
        return {
            maxPerformance: "light" as const,
            recommendedOptions: anime4kOptions.filter(opt => opt.performance === "light"),
        }
    }
}

export const isWebGPUAvailable = async (): Promise<boolean> => {
    if (!navigator.gpu) {
        return false
    }

    try {
        const adapter = await navigator.gpu.requestAdapter()
        if (!adapter) return false

        const device = await adapter.requestDevice()
        return !!device
    }
    catch {
        return false
    }
}

export const getOptimalAnime4KSettings = async (videoResolution: { width: number; height: number }) => {
    const webGPUAvailable = await isWebGPUAvailable()

    if (!webGPUAvailable) {
        return {
            supported: false,
            recommendation: "off" as Anime4KOption,
            reason: "WebGPU not supported on this device",
        }
    }

    const gpuInfo = await getGPUInfo()
    const recommendation = getPerformanceRecommendation(gpuInfo?.gpu)
    const videoRecommendations = getRecommendedAnime4KOptions(videoResolution)

    const optimalOption = anime4kOptions.find(option =>
        videoRecommendations.some(vr => vr.value === option.value) &&
        recommendation.recommendedOptions.some(pr => pr.value === option.value),
    )

    return {
        supported: true,
        recommendation: optimalOption?.value || "mode-a" as Anime4KOption,
        reason: `Recommended for ${videoResolution.height}p video on ${gpuInfo?.gpu || "current GPU"}`,
        alternatives: recommendation.recommendedOptions.slice(0, 3),
    }
}

const getGPUInfo = async () => {
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

