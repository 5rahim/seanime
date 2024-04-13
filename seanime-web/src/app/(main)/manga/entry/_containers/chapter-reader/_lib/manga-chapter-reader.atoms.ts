"use client"
import { atom } from "jotai/index"
import { atomWithStorage } from "jotai/utils"

export const __manga_currentPageIndexAtom = atom(0)
export const __manga_currentPaginationMapIndexAtom = atom(0) // HORIZONTAL MODE
export const __manga_paginationMapAtom = atom<Record<number, number[]>>({})


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const __manga_entryReaderSettings = atomWithStorage("sea-manga-entry-reader-settings", {})

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const MANGA_KBS = {
    kbsChapterLeft: "sea-manga-chapter-left",
    kbsChapterRight: "sea-manga-chapter-right",
    kbsPageLeft: "sea-manga-page-left",
    kbsPageRight: "sea-manga-page-right",

}

export const MANGA_DEFAULT_KBS = {
    [MANGA_KBS.kbsChapterLeft]: "[",
    [MANGA_KBS.kbsChapterRight]: "]",
    [MANGA_KBS.kbsPageLeft]: "left",
    [MANGA_KBS.kbsPageRight]: "right",
}


export const __manga_kbsChapterLeft = atomWithStorage(MANGA_KBS.kbsChapterLeft, MANGA_DEFAULT_KBS[MANGA_KBS.kbsChapterLeft])
export const __manga_kbsChapterRight = atomWithStorage(MANGA_KBS.kbsChapterRight, MANGA_DEFAULT_KBS[MANGA_KBS.kbsChapterRight])
export const __manga_kbsPageLeft = atomWithStorage(MANGA_KBS.kbsPageLeft, MANGA_DEFAULT_KBS[MANGA_KBS.kbsPageLeft])
export const __manga_kbsPageRight = atomWithStorage(MANGA_KBS.kbsPageRight, MANGA_DEFAULT_KBS[MANGA_KBS.kbsPageRight])

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const MangaReadingDirection = {
    LTR: "ltr",
    RTL: "rtl",
}

export const __manga_readingDirectionAtom = atomWithStorage<string>("sea-manga-reading-direction", MangaReadingDirection.LTR)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const MangaReadingMode = {
    LONG_STRIP: "long-strip",
    PAGED: "paged",
    DOUBLE_PAGE: "double-page",
}

export const __manga_readingModeAtom = atomWithStorage<string>("sea-manga-reading-mode", MangaReadingMode.LONG_STRIP)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const MangaPageFit = {
    CONTAIN: "contain",
    LARGER: "larger",
    COVER: "cover",
    TRUE_SIZE: "true-size",
}


export const __manga_pageFitAtom = atomWithStorage<string>("sea-manga-page-fit", MangaPageFit.CONTAIN)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const MangaPageStretch = {
    NONE: "none",
    STRETCH: "stretch",
}


export const __manga_pageStretchAtom = atomWithStorage<string>("sea-manga-page-stretch", MangaPageStretch.NONE)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const __manga_pageGapAtom = atomWithStorage<boolean>("sea-manga-page-gap", true)

export const __manga_pageGapShadowAtom = atomWithStorage("sea-manga-page-gap-shadow", true)

export const __manga_doublePageOffsetAtom = atomWithStorage("sea-manga-double-page-offset", 0)

export const __manga_isLastPageAtom = atom(false)
