import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { useVideoCoreVolumeBoost } from "./video-core-volume-boost"

const effects = vi.hoisted(() => [] as (() => void)[])
vi.mock("react", () => ({ default: { useEffect: (fn: () => void | (() => void)) => {
    const cleanup = fn()
    if (cleanup) effects.push(cleanup)
} } }))
vi.mock("@/lib/helpers/debug", () => ({ logger: () => ({ info: vi.fn(), error: vi.fn() }) }))
vi.mock("@/app/(main)/_features/media-core/media-core-volume-boost", () => import("../media-core/media-core-volume-boost"))

function cleanup() { effects.splice(0).forEach(fn => fn()) }

describe("VideoCore volume boost", () => {
    let source: any, gain: any, limiter: any, context: any
    beforeEach(() => {
        source = { connect: vi.fn(), disconnect: vi.fn() }
        gain = { connect: vi.fn(), disconnect: vi.fn(), gain: { setTargetAtTime: vi.fn() } }
        limiter = { connect: vi.fn(), disconnect: vi.fn(), threshold: {}, knee: {}, ratio: {}, attack: {}, release: {} }
        context = {
            state: "running", currentTime: 0, destination: {},
            createMediaElementSource: vi.fn(() => source), createGain: () => gain, createDynamicsCompressor: () => limiter,
        }
        gain.context = context
        vi.stubGlobal("window", Object.assign(new EventTarget(), { AudioContext: class { constructor() { return context } } }))
    })
    afterEach(() => { cleanup(); context.state = "closed"; vi.unstubAllGlobals() })

    it("leaves unboosted playback alone", () => {
        useVideoCoreVolumeBoost(new EventTarget() as HTMLVideoElement, 1)
        expect(context.createMediaElementSource).not.toHaveBeenCalled()
    })

    it("disconnects audio nodes when the video closes", () => {
        useVideoCoreVolumeBoost(new EventTarget() as HTMLVideoElement, 3)
        expect(gain.gain.setTargetAtTime).toHaveBeenCalledWith(3, 0, 0.02)
        cleanup()
        expect(source.disconnect).toHaveBeenCalledOnce()
        expect(gain.disconnect).toHaveBeenCalledOnce()
        expect(limiter.disconnect).toHaveBeenCalledOnce()
    })

    it("reuses the media source and bypasses the limiter when boost is turned off", () => {
        const video = new EventTarget() as HTMLVideoElement
        useVideoCoreVolumeBoost(video, 2)
        cleanup()
        useVideoCoreVolumeBoost(video, 1)
        expect(context.createMediaElementSource).toHaveBeenCalledOnce()
        expect(gain.connect).toHaveBeenLastCalledWith(context.destination)
        expect(gain.gain.setTargetAtTime).toHaveBeenLastCalledWith(1, 0, 0.02)
    })
})
