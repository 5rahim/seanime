import type { Anime4KOption } from "@/app/(main)/_features/video-core/video-core-anime-4k-manager"
import { atomWithStorage } from "jotai/utils"

export const vc_anime4kOption = atomWithStorage<Anime4KOption>("sea-video-core-anime4k", "off", undefined, { getOnInit: true })

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
    const lower = (gpu || "").toLowerCase()

    const isHighEnd = !!gpu && (
        // Marketing names (rarely exposed by WebGPU; kept for non-browser callers)
        gpu.includes("RTX 40") ||
        gpu.includes("RTX 3080") ||
        gpu.includes("RTX 3090") ||
        gpu.includes("RX 9070") ||
        gpu.includes("RX 7900") ||
        gpu.includes("RX 7800") ||
        gpu.includes("RX 6800") ||
        gpu.includes("RX 6900") ||
        gpu.includes("Radeon Pro W7") ||
        gpu.includes("M1 Pro") ||
        gpu.includes("M1 Max") ||
        gpu.includes("M2") ||
        gpu.includes("M3") ||
        // WebGPU architecture strings (Chrome/Edge expose these)
        lower.includes("rdna-3") ||
        lower.includes("rdna-4") ||
        lower.includes("ada-lovelace") ||
        lower.includes("blackwell") ||
        lower.includes("hopper") ||
        lower.includes("apple-m2") ||
        lower.includes("apple-m3") ||
        lower.includes("apple-m4")
    )

    const isMidRange = !!gpu && (
        gpu.includes("RTX 30") ||
        gpu.includes("RTX 20") ||
        gpu.includes("GTX 16") ||
        gpu.includes("RX 7700") ||
        gpu.includes("RX 7600") ||
        gpu.includes("RX 6700") ||
        gpu.includes("RX 6600") ||
        gpu.includes("RX 6500") ||
        gpu.includes("RX 6400") ||
        gpu.includes("RX 5") ||
        gpu.includes("M1") ||
        // Architecture strings
        lower.includes("rdna-2") ||
        lower.includes("ampere") ||
        lower.includes("turing") ||
        lower.includes("xe-2") ||
        lower.includes("xe-hpg") ||
        lower.includes("apple-m1")
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

export type GPUVendor = "AMD" | "NVIDIA" | "Intel" | "Apple" | "Other"

export const getGPUVendor = (gpu?: string): GPUVendor => {
    if (!gpu) return "Other"
    const lower = gpu.toLowerCase()
    // Vendor strings (WebGPU normalized) and marketing names
    if (lower.includes("amd") || lower.includes("radeon") || lower.includes("rx ") || lower.includes("vega")) return "AMD"
    if (lower.includes("nvidia") || lower.includes("geforce") || lower.includes("rtx") || lower.includes("gtx")) return "NVIDIA"
    if (lower.includes("intel") || lower.includes("arc ") || lower.includes("iris")) return "Intel"
    if (lower.includes("apple") || lower.startsWith("m1") || lower.startsWith("m2") || lower.startsWith("m3") || lower.startsWith("m4")) return "Apple"
    // WebGPU architecture strings (no vendor word but still identifiable)
    if (lower.startsWith("rdna") || lower === "gcn" || lower.startsWith("gcn-")) return "AMD"
    if (
        lower.includes("ada-lovelace") || lower.includes("ampere") || lower.includes("turing") ||
        lower.includes("pascal") || lower.includes("maxwell") || lower.includes("blackwell") || lower.includes("hopper")
    ) return "NVIDIA"
    if (lower.startsWith("xe") || lower.includes("gen-")) return "Intel"
    return "Other"
}

// Returns the label that should appear in the menu's `moreInfo` slot for an
// option, given the detected GPU. When the option is in the recommended bucket
// for the GPU it shows e.g. "AMD ✓"; otherwise it falls back to the legacy
// "Heavy" badge for heavy-tier options, or undefined for normal ones.
// `gpuArch` is matched against marketing names AND WebGPU architecture strings;
// `gpuVendor` is the normalized vendor token used solely to label the badge.
export const getAnime4KOptionRecommendation = (
    optionValue: Anime4KOption,
    gpuArch?: string,
    gpuVendor?: string,
): string | undefined => {
    const option = anime4kOptions.find(o => o.value === optionValue)
    if (!option || option.value === "off") return undefined

    if (gpuArch) {
        const recommendation = getPerformanceRecommendation(gpuArch)
        const isRecommended = recommendation.recommendedOptions.some(o => o.value === option.value)
        if (isRecommended) {
            const vendor = getGPUVendor(gpuVendor || gpuArch)
            return vendor === "Other" ? "✓" : `${vendor} ✓`
        }
    }

    return option.performance === "heavy" ? "Heavy" : undefined
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

export const getGPUInfo = async (): Promise<{ gpu: string; vendor: string; architecture: string } | null> => {
    if (!navigator.gpu) return null

    try {
        const adapter = await navigator.gpu.requestAdapter()
        if (!adapter) return null

        const info = (adapter as any).info || {}
        const architecture: string = info.architecture || ""
        const vendor: string = info.vendor || ""
        // Prefer the more descriptive architecture string ("rdna-3", "rx-7800-xt", ...)
        // when available, falling back to the vendor.
        const gpu = architecture || vendor || "Unknown GPU"

        return {
            gpu,
            vendor: vendor || "Unknown",
            architecture,
        }
    }
    catch {
        return null
    }
}

