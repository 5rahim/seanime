import { describe, expect, it } from "vitest"
import { sortMangaChaptersAscending } from "./sort-manga-chapters"

describe("sortMangaChaptersAscending", () => {
    it("orders chapters numerically, keeps duplicate chapters stable, and puts unnumbered chapters last", () => {
        const chapters = [
            { chapter: "10", id: "ten" },
            { chapter: "1.5", id: "one-and-a-half" },
            { chapter: "", id: "unnumbered" },
            { chapter: "1", id: "first" },
            { chapter: "1", id: "second-group" },
            { chapter: "2", id: "two" },
        ]

        expect(sortMangaChaptersAscending(chapters).map(chapter => chapter.id)).toEqual([
            "first", "second-group", "one-and-a-half", "two", "ten", "unnumbered",
        ])
        expect(chapters[0].id).toBe("ten")
    })
})
