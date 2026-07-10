import { describe, expect, it } from "vitest"
import { checkCodecSupport } from "./codec-utils"

describe("checkCodecSupport", () => {
    const defaultOptions = {
        isMobile: false,
        canUseMatroskaFallback: false,
        canPlayType: () => "" as const,
    }

    it("returns false for empty/undefined input", () => {
        expect(checkCodecSupport("", defaultOptions)).toBe(false)
    })

    it("returns false in mobile environment", () => {
        expect(checkCodecSupport("video/mp4", { ...defaultOptions, isMobile: true })).toBe(false)
    })

    it("uses direct Matroska support when reported", () => {
        const canPlayType = (codec: string) => {
            if (codec === "video/x-matroska; codecs=\"vp8, opus\"" || codec === "video/matroska; codecs=\"vp8, opus\"") {
                return "probably"
            }
            return ""
        }
        expect(checkCodecSupport("video/x-matroska; codecs=\"vp8, opus\"", { ...defaultOptions, canPlayType })).toBe(true)
        expect(checkCodecSupport("video/matroska; codecs=\"vp8, opus\"", { ...defaultOptions, canPlayType })).toBe(true)
    })

    it("does not infer Matroska support outside Chromium", () => {
        const canPlayType = (codec: string) => codec.startsWith("video/mp4") ? "probably" as const : "" as const
        expect(checkCodecSupport("video/matroska; codecs=\"avc1.640028\"", { ...defaultOptions, canPlayType })).toBe(false)
    })

    it("performs MP4 fallback in Chromium", () => {
        const canPlayType = (codec: string) => {
            if (codec === "video/mp4; codecs=\"avc1.640028\"") {
                return "probably"
            }
            return ""
        }
        const options = { ...defaultOptions, canUseMatroskaFallback: true, canPlayType }
        expect(checkCodecSupport("video/x-matroska; codecs=\"avc1.640028\"", options)).toBe(true)
        expect(checkCodecSupport("video/matroska; codecs=\"avc1.640028\"", options)).toBe(true)
    })

    it("performs WebM fallback in Chromium", () => {
        const canPlayType = (codec: string) => {
            if (codec === "video/webm; codecs=\"vp09.00.40.08, opus\"") {
                return "probably"
            }
            return ""
        }
        const options = { ...defaultOptions, canUseMatroskaFallback: true, canPlayType }
        expect(checkCodecSupport("video/x-matroska; codecs=\"vp09.00.40.08, opus\"", options)).toBe(true)
        expect(checkCodecSupport("video/matroska; codecs=\"vp09.00.40.08, opus\"", options)).toBe(true)
    })

    it("returns false if direct check and fallbacks fail", () => {
        const options = { ...defaultOptions, canUseMatroskaFallback: true }
        expect(checkCodecSupport("video/x-matroska; codecs=\"hvc1.1.6.L120.B0\"", options)).toBe(false)
    })

    it("does not use Matroska fallbacks for other containers", () => {
        const canPlayType = (codec: string) => codec.startsWith("video/webm") ? "probably" as const : "" as const
        const options = { ...defaultOptions, canUseMatroskaFallback: true, canPlayType }
        expect(checkCodecSupport("video/quicktime; codecs=\"avc1.640028\"", options)).toBe(false)
    })
})
