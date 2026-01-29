"use client"
import { atom } from "jotai/index"
import { atomWithStorage } from "jotai/utils"

export const __manga_currentPageIndexAtom = atom(0)
export const __manga_currentPaginationMapIndexAtom = atom(0) // HORIZONTAL MODE
export const __manga_paginationMapAtom = atom<Record<number, number[]>>({})

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
