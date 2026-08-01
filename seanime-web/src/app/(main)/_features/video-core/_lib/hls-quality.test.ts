import { describe, expect, it } from "vitest"
import { getPreferredHlsQualityLevel } from "./hls-quality"
import type { HlsQualityOption } from "./hls-quality"

const levels: HlsQualityOption[] = [
    { index: 0, height: 360, bitrate: 800_000 },
    { index: 1, height: 720, bitrate: 2_500_000 },
    { index: 2, height: 1080, bitrate: 5_000_000 },
]

describe("getPreferredHlsQualityLevel", () => {
    it("matches the stored resolution", () => {
        expect(getPreferredHlsQualityLevel(levels, "1080p - English")).toBe(2)
        expect(getPreferredHlsQualityLevel(levels, "1080p - Default")).toBe(2)
    })

    it("keeps automatic quality selection", () => {
        expect(getPreferredHlsQualityLevel(levels, "auto")).toBe(-1)
        expect(getPreferredHlsQualityLevel(levels, "default")).toBe(-1)
    })

    it("uses the lowest bitrate when the stored resolution is unavailable", () => {
        expect(getPreferredHlsQualityLevel(levels, "480p")).toBe(0)
    })

    it("does not override HLS for an unknown quality label", () => {
        expect(getPreferredHlsQualityLevel(levels, "HD")).toBeNull()
        expect(getPreferredHlsQualityLevel(levels, undefined)).toBeNull()
    })
})
