export type MediaCoreChapter = {
    label: string | null
    start: number
    end: number
}

export function getChapterType(name: string | null | undefined) {
    if (!name) return false
    if (/opening$|^opening\s|^op$/mi.test(name)) return "Opening"
    if (/ending$|^ending\s|^ed$|^credits/mi.test(name)) return "Ending"
    if (/^intro$|recap/mi.test(name)) return "Intro"
    if (/^outro$/mi.test(name)) return "Outro"
    return false
}

export function introIsOpening(chapters: Array<{ label: string | null }>) {
    const types = chapters.map(chapter => getChapterType(chapter.label)).filter(Boolean)
    return types.includes("Intro") && !types.includes("Opening")
}

type SkipOptions = {
    guardIntro?: boolean
}

export function getDefaultSkipChapters<T extends { label: string | null }>(chapters: T[], options: SkipOptions = {}) {
    let opening: T | null = null
    let ending: T | null = null
    const usesIntro = options.guardIntro !== false && introIsOpening(chapters)

    for (const chapter of chapters) {
        const type = getChapterType(chapter.label)
        if (!opening && !usesIntro && type === "Opening") opening = chapter
        if (!ending && !usesIntro && type === "Ending") ending = chapter
        if (opening && ending) break
    }

    return { opening, ending }
}

function getPatterns(value: string) {
    return value
        .split(",")
        .map(pattern => pattern.trim())
        .filter(Boolean)
}

function getRegexes(value: string) {
    return getPatterns(value).flatMap(pattern => {
        try {
            return [new RegExp(pattern, "i")]
        }
        catch {
            return []
        }
    })
}

export function getSkipPatternError(value: string) {
    for (const pattern of getPatterns(value)) {
        try {
            new RegExp(pattern, "i")
        }
        catch {
            return `Invalid regex: ${pattern}`
        }
    }
    return ""
}

export function getSkipChapters<T extends { label: string | null }>(chapters: T[], patterns: string, options: SkipOptions = {}) {
    const defaults = getDefaultSkipChapters(chapters, options)
    const regexes = getRegexes(patterns)

    return chapters.filter(chapter => {
        if (chapter === defaults.opening || chapter === defaults.ending) return true
        const label = chapter.label
        if (!label) return false
        return regexes.some(regex => regex.test(label))
    })
}

export function getSkipLabel(name: string | null) {
    const type = getChapterType(name)
    if (type === "Opening" || type === "Ending") return type
    return name?.trim() || "Chapter"
}
