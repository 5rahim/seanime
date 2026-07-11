// @ts-ignore
import { NativePlayer_SubtitleEventsPayload } from "@/api/generated/types"
import { describe, expect, it } from "vitest"
import { getSubtitleEvents, isSubtitleBatchCurrent } from "./native-player-subtitles"

function newBatch(playbackId: string, generationId: number, text: string): NativePlayer_SubtitleEventsPayload {
    return {
        playbackId,
        generationId,
        seekTime: 0,
        events: [{
            trackNumber: 1,
            text,
            startTime: 0,
            duration: 1,
            codecID: "S_TEXT/ASS",
        }],
    }
}

describe("native player subtitle batches", () => {
    it("rejects previous playbacks and generations", () => {
        expect(isSubtitleBatchCurrent(newBatch("old", 4, "old playback"), "current", 3)).toBe(false)
        expect(isSubtitleBatchCurrent(newBatch("current", 2, "old seek"), "current", 3)).toBe(false)
        expect(isSubtitleBatchCurrent(newBatch("current", 3, "current"), "current", 3)).toBe(true)
    })

    it("flushes only the latest playback generation", () => {
        const events = getSubtitleEvents([
            newBatch("old", 2, "old playback"),
            newBatch("current", 1, "old seek"),
            newBatch("current", 2, "first"),
            newBatch("current", 2, "second"),
        ], "current", 2)

        expect(events.map(event => event.text)).toEqual(["first", "second"])
    })
})
