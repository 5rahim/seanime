import { describe, expect, it } from "vitest"
import { hasStartupGap } from "./hls-buffer"

function video(currentTime: number, ranges: [number, number][]) {
    return {
        currentTime,
        buffered: {
            length: ranges.length,
            start: (index: number) => ranges[index][0],
            end: (index: number) => ranges[index][1],
        },
    }
}

describe("hasStartupGap", () => {
    it("allows HLS to recover the reported startup gap", () => {
        expect(hasStartupGap(video(0, [[0.083423, 22.624]]))).toBe(true)
        expect(hasStartupGap(video(0.05, [[0.083423, 22.624]]))).toBe(true)
    })

    it("keeps empty buffers and large startup gaps on normal recovery", () => {
        expect(hasStartupGap(video(0, []))).toBe(false)
        expect(hasStartupGap(video(0, [[0.5, 22]]))).toBe(false)
        expect(hasStartupGap(video(0, [[3, 22]]))).toBe(false)
    })

    it("requires enough buffered data after the gap", () => {
        expect(hasStartupGap(video(0, [[0.08, 0.1]]))).toBe(false)
    })

    it("does not ignore stalls inside the buffer or later in playback", () => {
        expect(hasStartupGap(video(0, [[0, 22]]))).toBe(false)
        expect(hasStartupGap(video(0.09, [[0.08, 22]]))).toBe(false)
        expect(hasStartupGap(video(2, [[2.08, 22]]))).toBe(false)
    })
})
