import { describe, expect, it } from "vitest"
import { getSkipChapters, getSkipPatternError } from "./media-core-chapters"

function chapters(...labels: string[]) {
    return labels.map((label, index) => ({
        label,
        start: index * 60,
        end: (index + 1) * 60,
    }))
}

describe("chapter skipping", () => {
    it("keeps the first default opening and ending", () => {
        const list = chapters("Opening", "Episode", "Opening 2", "Credits")

        expect(getSkipChapters(list, "")).toEqual([list[0], list[3]])
    })

    it("keeps the existing intro chapter rule", () => {
        const list = chapters("Intro", "Episode", "Ending")

        expect(getSkipChapters(list, "")).toEqual([])
        expect(getSkipChapters(list, "", { guardIntro: false })).toEqual([list[2]])
    })

    it("adds all chapters matching custom regexes", () => {
        const list = chapters("Intro", "Episode", "Next Episode Preview", "Outro")

        expect(getSkipChapters(list, "^intro$,preview,^outro$")).toEqual([list[0], list[2], list[3]])
    })

    it("matches custom regexes case insensitively", () => {
        const list = chapters("PREVIEW")

        expect(getSkipChapters(list, "^preview$")).toEqual(list)
    })

    it("reports invalid regexes", () => {
        expect(getSkipPatternError("^Preview$,(")).toBe("Invalid regex: (")
        expect(getSkipPatternError("^Preview$")).toBe("")
    })
})
