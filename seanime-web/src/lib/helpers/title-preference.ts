export type TitleLanguage = "romaji" | "english" | "native"

export const DEFAULT_TITLE_PREFERENCE: TitleLanguage[] = ["romaji", "english", "native"]

export const TITLE_LANGUAGE_OPTIONS: { id: TitleLanguage; label: string }[] = [
    { id: "romaji", label: "Romaji" },
    { id: "english", label: "English" },
    { id: "native", label: "Native" },
]

export function parsePreferredTitleLanguage(pref: string | undefined): TitleLanguage[] {
    if (!pref) return DEFAULT_TITLE_PREFERENCE
    const parsed = pref.split(",").filter(Boolean) as TitleLanguage[]
    if (parsed.length !== 3) return DEFAULT_TITLE_PREFERENCE
    const validOptions = ["romaji", "english", "native"]
    if (!parsed.every(p => validOptions.includes(p))) return DEFAULT_TITLE_PREFERENCE
    return parsed
}

export function serializePreferredTitleLanguage(prefs: TitleLanguage[]): string {
    return prefs.join(",")
}

export function getPreferredTitle(
    title: { english?: string | null; romaji?: string | null; native?: string | null } | undefined,
    preferences: TitleLanguage[],
): string {
    if (!title) return "Untitled"

    for (const pref of preferences) {
        switch (pref) {
            case "english":
                if (title.english) return title.english
                break
            case "romaji":
                if (title.romaji) return title.romaji
                break
            case "native":
                if (title.native) return title.native
                break
        }
    }

    return title.romaji || title.english || title.native || "Untitled"
}
