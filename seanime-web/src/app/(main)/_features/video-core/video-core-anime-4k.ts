import { vc_anime4kManager, vc_miniPlayer, vc_pip, vc_realVideoSize, vc_seeking, vc_videoElement } from "@/app/(main)/_features/video-core/video-core"
import { Anime4KOption } from "@/app/(main)/_features/video-core/video-core-anime-4k-manager"
import { logger } from "@/lib/helpers/debug"
import { useAtomValue } from "jotai"
import { useAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React from "react"

const log = logger("VIDEO CORE ANIME 4K")

export const vc_anime4kOption = atomWithStorage<Anime4KOption>("sea-video-core-anime4k", "off", undefined, { getOnInit: true })

export const VideoCoreAnime4K = () => {
    const realVideoSize = useAtomValue(vc_realVideoSize)
    const seeking = useAtomValue(vc_seeking)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const isPip = useAtomValue(vc_pip)
    const video = useAtomValue(vc_videoElement)

    const manager = useAtomValue(vc_anime4kManager)
    const [selectedOption] = useAtom(vc_anime4kOption)

    // Update manager with real video size
    React.useEffect(() => {
        if (manager) {
            manager.updateCanvasSize(realVideoSize)
        }
    }, [manager, realVideoSize])

    // Handle option changes
    React.useEffect(() => {
        if (video && manager) {
            log.info("Setting Anime4K option", selectedOption)
            manager.setOption(selectedOption, {
                isMiniPlayer,
                isPip,
                seeking,
            })
        }
    }, [video, manager, selectedOption, isMiniPlayer, isPip, seeking])

    return null
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

