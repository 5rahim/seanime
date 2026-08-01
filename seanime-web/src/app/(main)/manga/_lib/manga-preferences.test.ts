import { describe, expect, it } from "vitest"
import { fromMangaPreferences, getActiveMangaFilters, toMangaPreferences } from "./manga-preferences"

describe("manga preferences", () => {
    it("preserves filters for inactive providers", () => {
        const preferences = toMangaPreferences(
            { "1": "provider-a" },
            {
                "1$provider-a": { scanlators: ["Group A"], language: "en" },
                "1$provider-b": { scanlators: ["Group B"], language: "fr" },
            },
        )

        expect(preferences.entries?.[1].provider).toBe("provider-a")
        expect(preferences.entries?.[1].filters?.["provider-b"]).toEqual({
            scanlators: ["Group B"],
            language: "fr",
        })
    })

    it("hydrates the local provider and filter maps", () => {
        const unpacked = fromMangaPreferences({
            entries: {
                5: {
                    provider: "provider-b",
                    filters: {
                        "provider-b": { scanlators: ["Group"], language: "ja" },
                    },
                },
            },
        })

        expect(unpacked.providers).toEqual({ "5": "provider-b" })
        expect(unpacked.filters).toEqual({
            "5$provider-b": { scanlators: ["Group"], language: "ja" },
        })
    })

    it("uses filters from the selected provider only", () => {
        const filters = getActiveMangaFilters(
            {
                "1$provider-a": { scanlators: ["Active Group"], language: "en" },
                "1$provider-b": { scanlators: ["Inactive Group"], language: "fr" },
            },
            { "1": "provider-a" },
            [{
                id: "provider-a",
                name: "Provider A",
                lang: "en",
                settings: { supportsMultiScanlator: true, supportsMultiLanguage: true },
            }],
        )

        expect(filters).toEqual({
            "1": { scanlators: ["Active Group"], language: "en" },
        })
    })
})
