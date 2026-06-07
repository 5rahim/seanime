import { atom } from "jotai"
import { atomWithStorage } from "jotai/utils"

export const __manga_currentPageIndexAtom = atom(0)
export const __manga_currentPaginationMapIndexAtom = atom(0) // HORIZONTAL MODE
export const __manga_paginationMapAtom = atom<Record<number, number[]>>({})

export type MangaReaderResumeLocation = {
    chapterId: string
    provider: string
    pageIndex: number
    updatedAt: number
}

export const MANGA_READER_RESUME_MAX_ENTRIES = 250
export const MANGA_READER_RESUME_MAX_AGE_MS = 1000 * 60 * 60 * 24 * 90

export function cleanupMangaResumeLocations(locations: Record<string, MangaReaderResumeLocation>) {
    const now = Date.now()

    const trimmedLocations = Object.entries(locations)
        .filter(([key, value]) => {
            return (
                !!key
                && typeof value?.chapterId === "string"
                && value.chapterId.length > 0
                && typeof value?.provider === "string"
                && value.provider.length > 0
                && Number.isInteger(value?.pageIndex)
                && value.pageIndex >= 0
                && Number.isFinite(value?.updatedAt)
                && now - value.updatedAt <= MANGA_READER_RESUME_MAX_AGE_MS
            )
        })
        .sort((a, b) => b[1].updatedAt - a[1].updatedAt)
        .slice(0, MANGA_READER_RESUME_MAX_ENTRIES)

    const nextLocations = Object.fromEntries(trimmedLocations)

    const sameSize = Object.keys(nextLocations).length === Object.keys(locations).length
    if (!sameSize) return nextLocations

    for (const [key, value] of Object.entries(nextLocations)) {
        if (
            locations[key]?.pageIndex !== value.pageIndex
            || locations[key]?.updatedAt !== value.updatedAt
        ) {
            return nextLocations
        }
    }

    return locations
}

export const __manga_resumeLocationsAtom = atomWithStorage<Record<string, MangaReaderResumeLocation>>(
    "sea-manga-resume-locations",
    {},
    undefined,
    { getOnInit: true },
)

export const __manga_hiddenBarAtom = atom(false)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// e.g. { "mangaId": { "sea-manga-reading-mode": "long-stop" } }
// DEVNOTE: change key by adding "-vx" when settings system changes
export const __manga_entryReaderSettings = atomWithStorage<Record<string, Record<string, any>>>("sea-manga-entry-reader-settings", {})

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const MANGA_KBS_ATOM_KEYS = {
    kbsChapterLeft: "sea-manga-chapter-left",
    kbsChapterRight: "sea-manga-chapter-right",
    kbsPageLeft: "sea-manga-page-left",
    kbsPageRight: "sea-manga-page-right",
}

export const MANGA_DEFAULT_KBS = {
    [MANGA_KBS_ATOM_KEYS.kbsChapterLeft]: "[",
    [MANGA_KBS_ATOM_KEYS.kbsChapterRight]: "]",
    [MANGA_KBS_ATOM_KEYS.kbsPageLeft]: "left",
    [MANGA_KBS_ATOM_KEYS.kbsPageRight]: "right",
}


export const __manga_kbsChapterLeft = atomWithStorage(MANGA_KBS_ATOM_KEYS.kbsChapterLeft, MANGA_DEFAULT_KBS[MANGA_KBS_ATOM_KEYS.kbsChapterLeft])
export const __manga_kbsChapterRight = atomWithStorage(MANGA_KBS_ATOM_KEYS.kbsChapterRight, MANGA_DEFAULT_KBS[MANGA_KBS_ATOM_KEYS.kbsChapterRight])
export const __manga_kbsPageLeft = atomWithStorage(MANGA_KBS_ATOM_KEYS.kbsPageLeft, MANGA_DEFAULT_KBS[MANGA_KBS_ATOM_KEYS.kbsPageLeft])
export const __manga_kbsPageRight = atomWithStorage(MANGA_KBS_ATOM_KEYS.kbsPageRight, MANGA_DEFAULT_KBS[MANGA_KBS_ATOM_KEYS.kbsPageRight])

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const MANGA_SETTINGS_ATOM_KEYS = {
    readingMode: "sea-manga-reading-mode",
    readingDirection: "sea-manga-reading-direction",
    pageFit: "sea-manga-page-fit",
    pageStretch: "sea-manga-page-stretch",
    pageGap: "sea-manga-page-gap",
    pageGapShadow: "sea-manga-page-gap-shadow",
    doublePageOffset: "sea-manga-double-page-offset",
    overflowPageContainerWidth: "sea-manga-overflow-page-container-width",
}

export const MANGA_PAGE_ZOOM_MIN = 0.5
export const MANGA_PAGE_ZOOM_MAX = 3
export const MANGA_PAGE_ZOOM_STEP = 0.1
export const MANGA_PAGE_ZOOM_DEFAULT = 1
export const __manga_pageZoomAtom = atom(MANGA_PAGE_ZOOM_DEFAULT)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const MangaReadingDirection = {
    LTR: "ltr",
    RTL: "rtl",
}

export const __manga_readingDirectionAtom = atomWithStorage<string>(MANGA_SETTINGS_ATOM_KEYS.readingDirection, MangaReadingDirection.LTR)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const MangaReadingMode = {
    LONG_STRIP: "long-strip",
    PAGED: "paged",
    DOUBLE_PAGE: "double-page",
}

export const __manga_readingModeAtom = atomWithStorage<string>(MANGA_SETTINGS_ATOM_KEYS.readingMode, MangaReadingMode.LONG_STRIP)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const MangaPageFit = {
    CONTAIN: "contain",
    LARGER: "larger",
    COVER: "cover",
    TRUE_SIZE: "true-size",
}

export const __manga_pageFitAtom = atomWithStorage<string>(MANGA_SETTINGS_ATOM_KEYS.pageFit, MangaPageFit.CONTAIN)

export const __manga_pageOverflowContainerWidthAtom = atomWithStorage<number>(MANGA_SETTINGS_ATOM_KEYS.overflowPageContainerWidth, 50)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const MangaPageStretch = {
    NONE: "none",
    STRETCH: "stretch",
}


export const __manga_pageStretchAtom = atomWithStorage<string>(MANGA_SETTINGS_ATOM_KEYS.pageStretch, MangaPageStretch.NONE)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const __manga_readerProgressBarAtom = atomWithStorage<boolean>("sea-manga-reader-progress-bar", false)

export const __manga_pageGapAtom = atomWithStorage<boolean>(MANGA_SETTINGS_ATOM_KEYS.pageGap, true)

export const __manga_pageGapShadowAtom = atomWithStorage(MANGA_SETTINGS_ATOM_KEYS.pageGapShadow, true)

export const __manga_doublePageOffsetAtom = atomWithStorage(MANGA_SETTINGS_ATOM_KEYS.doublePageOffset, 0)

export const __manga_isLastPageAtom = atom(false)
