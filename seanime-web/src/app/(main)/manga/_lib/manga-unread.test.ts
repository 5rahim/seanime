import { describe, expect, it } from "vitest"
import { getMangaEntryLatestChapterNumber, getMangaEntryUnreadState } from "./manga-unread"

const latestChapterNumbers = {
    1: [
        { provider: "provider-a", language: "en", scanlator: "A", number: 12 },
        { provider: "provider-a", language: "fr", scanlator: "A", number: 10 },
        { provider: "provider-a", language: "en", scanlator: "B", number: 14 },
        { provider: "provider-b", language: "en", scanlator: "A", number: 100 },
    ],
}

describe("manga unread state", () => {
    it("uses only the selected provider and active filters", () => {
        expect(getMangaEntryLatestChapterNumber(1, latestChapterNumbers, { 1: "provider-a" }, {
            1: { language: "en", scanlators: ["A"] },
        })).toBe(12)

        expect(getMangaEntryLatestChapterNumber(1, latestChapterNumbers, { 1: "provider-a" }, {
            1: { language: "en", scanlators: [] },
        })).toBe(14)
    })

    it("returns unknown when the selected source or filter has no data", () => {
        expect(getMangaEntryLatestChapterNumber(1, latestChapterNumbers, {}, {})).toBeNull()
        expect(getMangaEntryLatestChapterNumber(1, latestChapterNumbers, { 1: "provider-a" }, {
            1: { language: "jp", scanlators: [] },
        })).toBeNull()
        expect(getMangaEntryLatestChapterNumber(2, latestChapterNumbers, { 2: "provider-a" }, {})).toBeNull()
    })

    it("never reports a negative unread count", () => {
        expect(getMangaEntryUnreadState(1, 9, latestChapterNumbers, { 1: "provider-a" }, {}).unread).toBe(5)
        expect(getMangaEntryUnreadState(1, 20, latestChapterNumbers, { 1: "provider-a" }, {})).toEqual({
            known: true,
            latest: 14,
            unread: 0,
        })
        expect(getMangaEntryUnreadState(2, 0, latestChapterNumbers, { 2: "provider-a" }, {})).toEqual({
            known: false,
            latest: null,
            unread: 0,
        })
    })
})
