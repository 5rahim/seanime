import { logger } from "@/lib/helpers/debug"
import React from "react"

const log = logger("VIDEO CORE VOLUME BOOST")

export const VOLUME_BOOST_MIN = 1
export const VOLUME_BOOST_MAX = 3
export const VOLUME_BOOST_STEP = 0.25

export const VOLUME_BOOST_OPTIONS = [
    { label: "Off", value: 1 },
    { label: "125%", value: 1.25 },
    { label: "150%", value: 1.5 },
    { label: "200%", value: 2 },
    { label: "250%", value: 2.5 },
    { label: "300%", value: 3 },
]

export function clampVolumeBoost(value: number) {
    if (!Number.isFinite(value)) return VOLUME_BOOST_MIN
    return Math.min(VOLUME_BOOST_MAX, Math.max(VOLUME_BOOST_MIN, Math.round(value * 100) / 100))
}

export function formatVolumeBoost(value: number) {
    return value <= VOLUME_BOOST_MIN ? "Off" : `${Math.round(value * 100)}%`
}

type BoostGraph = {
    gain: GainNode
}

// A media element can only be connected to a single MediaElementAudioSourceNode for its whole lifetime,
// so the graph is cached per element and the AudioContext is shared.
let sharedContext: AudioContext | null = null
const graphs = new WeakMap<HTMLMediaElement, BoostGraph>()

function getContext() {
    if (!sharedContext || sharedContext.state === "closed") {
        const Ctx = window.AudioContext || (window as any).webkitAudioContext
        if (!Ctx) return null
        sharedContext = new Ctx()
    }
    return sharedContext
}

function getOrCreateGraph(element: HTMLMediaElement): BoostGraph | null {
    const existing = graphs.get(element)
    if (existing) return existing

    const ctx = getContext()
    if (!ctx) return null

    try {
        const source = ctx.createMediaElementSource(element)
        const gain = ctx.createGain()
        // Acts as a limiter so boosted peaks get squashed instead of clipping
        const limiter = ctx.createDynamicsCompressor()
        limiter.threshold.value = -1
        limiter.knee.value = 0
        limiter.ratio.value = 20
        limiter.attack.value = 0.003
        limiter.release.value = 0.25

        source.connect(gain)
        gain.connect(limiter)
        limiter.connect(ctx.destination)

        const graph = { gain }
        graphs.set(element, graph)
        log.info("Audio graph created")
        return graph
    }
    catch (e) {
        log.error("Failed to create audio graph", e)
        return null
    }
}

function resumeContext() {
    if (sharedContext && sharedContext.state === "suspended") {
        sharedContext.resume().catch(() => undefined)
    }
}

/**
 * Amplifies the video element's audio past 100% using the Web Audio API.
 * The audio graph is only created once a boost above 100% is requested, so playback is untouched otherwise.
 */
export function useVideoCoreVolumeBoost(element: HTMLVideoElement | null, boost: number) {
    const value = clampVolumeBoost(boost)

    React.useEffect(() => {
        if (!element) return

        const graph = value > VOLUME_BOOST_MIN ? getOrCreateGraph(element) : graphs.get(element)
        if (!graph) return

        graph.gain.gain.setTargetAtTime(value, graph.gain.context.currentTime, 0.02)
        resumeContext()
    }, [element, value])

    // Browsers start an AudioContext suspended until there's a user gesture, which would silence the routed audio
    React.useEffect(() => {
        if (!element) return

        element.addEventListener("play", resumeContext)
        window.addEventListener("pointerdown", resumeContext)
        window.addEventListener("keydown", resumeContext)
        return () => {
            element.removeEventListener("play", resumeContext)
            window.removeEventListener("pointerdown", resumeContext)
            window.removeEventListener("keydown", resumeContext)
        }
    }, [element])
}
