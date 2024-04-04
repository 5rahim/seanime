"use client"
import { atom } from "jotai/index"
import { atomWithStorage } from "jotai/utils"

export const __manga_currentPageIndexAtom = atom(0)
export const __manga_currentPaginationMapIndexAtom = atom(0) // HORIZONTAL MODE
export const __manga_paginationMapAtom = atom<Record<number, number[]>>({})

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

export const __manga_isLastPageAtom = atom(false)
