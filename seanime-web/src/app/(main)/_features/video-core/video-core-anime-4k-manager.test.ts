import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { VideoCoreAnime4KManager } from "./video-core-anime-4k-manager"

vi.mock("@/lib/helpers/debug", () => ({ logger: () => ({ info: vi.fn(), error: vi.fn(), warning: vi.fn() }) }))
vi.mock("anime4k-webgpu", () => ({
    ModeA: class {
        pass = vi.fn()
        getOutputTexture() { return { createView: vi.fn() } }
    },
}))

function setup() {
    const video = Object.assign(new EventTarget(), {
        readyState: 4, HAVE_CURRENT_DATA: 2, videoWidth: 640, videoHeight: 360, paused: false,
        requestVideoFrameCallback: vi.fn(() => 1), cancelVideoFrameCallback: vi.fn(),
    })
    const context = { configure: vi.fn(), unconfigure: vi.fn(), getCurrentTexture: () => ({ createView: vi.fn() }) }
    const texture = { destroy: vi.fn() }
    const pass = { setPipeline: vi.fn(), setBindGroup: vi.fn(), draw: vi.fn(), end: vi.fn() }
    const releases: (() => void)[] = []
    const device = {
        destroy: vi.fn(), createTexture: vi.fn(() => texture),
        createBindGroupLayout: vi.fn(), createPipelineLayout: vi.fn(), createRenderPipeline: vi.fn(),
        createShaderModule: vi.fn(), createBindGroup: vi.fn(), createSampler: vi.fn(),
        createCommandEncoder: () => ({ beginRenderPass: () => pass, finish: vi.fn() }),
        queue: {
            copyExternalImageToTexture: vi.fn(), submit: vi.fn(),
            onSubmittedWorkDone: vi.fn(() => new Promise<void>(resolve => releases.push(resolve))),
        },
    }
    const requestDevice = vi.fn(async () => device)
    vi.stubGlobal("navigator", { gpu: { requestAdapter: async () => ({ requestDevice }), getPreferredCanvasFormat: () => "bgra8unorm" } })
    const canvas = { width: 640, height: 360, getContext: () => context, remove: vi.fn(), style: {} }
    const manager: any = new VideoCoreAnime4KManager({ videoElement: video as any, settings: {} as any })
    manager._boxSize = { width: 640, height: 360 }
    const changed = vi.fn()
    manager._onOptionChanged = changed
    vi.spyOn(manager, "_createCanvas").mockImplementation(() => { manager.canvas = canvas })
    return { manager, video, device, requestDevice, context, texture, releases, changed }
}

describe("Anime4K rendering", () => {
    beforeEach(() => {
        vi.useFakeTimers()
        vi.stubGlobal("GPUTextureUsage", { TEXTURE_BINDING: 1, COPY_DST: 2, RENDER_ATTACHMENT: 4 })
        vi.stubGlobal("GPUShaderStage", { FRAGMENT: 1 })
        vi.stubGlobal("requestAnimationFrame", vi.fn(() => 1))
        vi.stubGlobal("cancelAnimationFrame", vi.fn())
    })
    afterEach(() => { vi.useRealTimers(); vi.unstubAllGlobals(); vi.restoreAllMocks() })

    it("does not turn Anime4K back on after disabling it during setup", async () => {
        const { manager, changed } = setup()
        let finish!: () => void
        const render = vi.spyOn(manager, "_startRendering").mockImplementationOnce(() => new Promise<void>(resolve => { finish = resolve }))
            .mockResolvedValue(undefined)
        const pending = manager.setOption("mode-a")
        await vi.waitFor(() => expect(render).toHaveBeenCalledOnce())
        await manager.setOption("off")
        changed.mockClear()
        finish()
        await pending
        expect(changed).not.toHaveBeenCalled()
        expect(manager._initialized).toBe(false)
        manager.destroy()
    })

    it("waits for video data before allocating textures and cancels that wait on close", async () => {
        const { manager, video, requestDevice } = setup()
        video.readyState = 0
        video.videoWidth = 0
        const pending = manager.setOption("mode-a")
        await vi.waitFor(() => expect(manager._createCanvas).toHaveBeenCalledOnce())
        await Promise.resolve()
        await Promise.resolve()
        expect(requestDevice).not.toHaveBeenCalled()
        manager.destroy()
        await pending
        expect(requestDevice).not.toHaveBeenCalled()
    })

    it("releases the device when pipeline setup fails", async () => {
        const { manager, device } = setup()
        vi.spyOn(manager, "createPipeline").mockImplementation(() => { throw new Error("pipeline failed") })
        await manager.setOption("mode-a")
        expect(device.destroy).toHaveBeenCalledOnce()
    })

    it("starts rendering after video data becomes available", async () => {
        const { manager, video, requestDevice } = setup()
        video.readyState = 0
        const pending = manager.setOption("mode-a")
        await vi.waitFor(() => expect(manager._createCanvas).toHaveBeenCalledOnce())
        video.readyState = 2
        video.dispatchEvent(new Event("loadeddata"))
        await pending
        expect(requestDevice).toHaveBeenCalledOnce()
        expect(video.requestVideoFrameCallback).toHaveBeenCalledOnce()
        manager.destroy()
    })

    it("does not wait for an event that already fired", async () => {
        const { manager, video, requestDevice } = setup()
        video.readyState = 2
        await manager.setOption("mode-a")
        expect(requestDevice).toHaveBeenCalledOnce()
        manager.destroy()
    })

    it("bounds queued frames and stops the loop on close", async () => {
        const { manager, video, device, releases, context, texture } = setup()
        await manager.setOption("mode-a")
        const frame = video.requestVideoFrameCallback.mock.calls[0][0] as unknown as () => void
        frame(); frame(); frame()
        expect(device.queue.submit).toHaveBeenCalledTimes(2)
        releases[0]()
        await Promise.resolve()
        frame()
        expect(device.queue.submit).toHaveBeenCalledTimes(3)
        manager.destroy()
        const callbacks = video.requestVideoFrameCallback.mock.calls.length
        frame()
        expect(video.requestVideoFrameCallback).toHaveBeenCalledTimes(callbacks)
        expect(device.destroy).toHaveBeenCalledOnce()
        expect(texture.destroy).toHaveBeenCalledOnce()
        expect(context.unconfigure).toHaveBeenCalledOnce()
    })
})
