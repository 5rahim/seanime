import { Manga_ChapterContainer, Manga_PageContainer, Nullish } from "@/api/generated/types"
import { useMangaEntryDownloadedChapters } from "@/app/(main)/manga/_lib/handle-manga-downloads"
import { getDecimalFromChapter, isChapterAfter, isChapterBefore } from "@/app/(main)/manga/_lib/handle-manga-utils"
import {
    __manga_currentPageIndexAtom,
    __manga_currentPaginationMapIndexAtom,
    __manga_doublePageOffsetAtom,
    __manga_pageFitAtom,
    __manga_pageStretchAtom,
    __manga_paginationMapAtom,
    __manga_readingDirectionAtom,
    __manga_readingModeAtom,
    MangaPageFit,
    MangaPageStretch,
    MangaReadingDirection,
    MangaReadingMode,
} from "@/app/(main)/manga/_lib/manga-chapter-reader.atoms"
import { logger } from "@/lib/helpers/debug"
import { atom } from "jotai"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import mousetrap from "mousetrap"
import React from "react"

const __manga_readerLoadedPagesAtom = atom<number[]>([])

export function useHandleChapterPageStatus(pageContainer: Manga_PageContainer | undefined) {
    const currentChapter = useCurrentChapter()

    /**
     * Keep track of loaded page indexes
     * - Well compare the length to the number of pages to determine if we can show the progress bar
     */
    const [loadedPages, setLoadedPages] = useAtom(__manga_readerLoadedPagesAtom)

    React.useEffect(() => {
        setLoadedPages([])
    }, [currentChapter])

    const handlePageLoad = React.useCallback((pageIndex: number) => {
        setLoadedPages(prev => [...prev, pageIndex])
    }, [])

    return {
        allPagesLoaded: loadedPages.length > 0 && loadedPages.length === pageContainer?.pages?.length,
        loadedPages,
        handlePageLoad,
    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

/**
 * Current chapter being read
 */
export type MangaReader_SelectedChapter = {
    chapterNumber: string
    provider: string
    chapterId: string
    mediaId: number
}

/**
 * Stores the current chapter being read
 */
export const __manga_selectedChapterAtom = atomWithStorage<MangaReader_SelectedChapter | undefined>("sea-manga-chapter",
    undefined,
    undefined,
    { getOnInit: true })

export function useSetCurrentChapter() {
    return useSetAtom(__manga_selectedChapterAtom)
}

export function useCurrentChapter() {
    return useAtomValue(__manga_selectedChapterAtom)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export function useHandleChapterPagination(mId: Nullish<string | number>, chapterContainer: Manga_ChapterContainer | undefined) {
    const currentChapter = useCurrentChapter()
    const setCurrentChapter = useSetCurrentChapter()
    /**
     * Get the entry's downloaded chapters from the atom
     */
    const entryDownloadedChapters = useMangaEntryDownloadedChapters()

    // Get previous and next chapters
    const previousChapter = React.useMemo<MangaReader_SelectedChapter | undefined>(() => {
        if (!mId) return undefined
        if (!currentChapter) return undefined

        // First, look in downloaded chapters
        // e.g., current is 14.2, look for the highest chapter number that is less than 14.2
        const _1 = entryDownloadedChapters
            .filter(ch => ch.chapterId !== currentChapter.chapterId)
            .sort((a, b) => getDecimalFromChapter(b.chapterNumber) - getDecimalFromChapter(a.chapterNumber)) // Sort in descending order
            .find(ch => isChapterBefore(ch.chapterNumber, currentChapter.chapterNumber)) // Find the first chapter that is before the current chapter
        // Save the chapter if it exists
        const downloadedCh = _1 ? {
            chapterId: _1.chapterId,
            chapterNumber: _1.chapterNumber,
            provider: _1.provider as string,
            mediaId: Number(mId),
        } : undefined

        // Return it if there's no container
        if (!chapterContainer?.chapters) return downloadedCh

        // Look for the previous chapter in the chapter container
        const idx = chapterContainer.chapters.findIndex((chapter) => chapter.id === currentChapter?.chapterId)

        let previousContainerCh: MangaReader_SelectedChapter | undefined = undefined
        if (idx !== -1 && !!chapterContainer.chapters[idx - 1]) {
            previousContainerCh = {
                chapterId: chapterContainer.chapters[idx - 1].id,
                chapterNumber: chapterContainer.chapters[idx - 1].chapter,
                provider: chapterContainer.chapters[idx - 1].provider as string,
                mediaId: chapterContainer.mediaId,
            }
        }

        // Look in the chapter container, but this time, by sorting the chapters in descending order to find the adjacent chapter
        let _2 = chapterContainer.chapters
            .filter(ch => ch.id !== currentChapter.chapterId)
            .sort((a, b) => getDecimalFromChapter(b.chapter) - getDecimalFromChapter(a.chapter))
            .find(ch => isChapterBefore(ch.chapter, currentChapter.chapterNumber))
        const adjacentContainerCh = _2 ? {
            chapterId: _2.id,
            chapterNumber: _2.chapter,
            provider: _2.provider as string,
            mediaId: chapterContainer.mediaId,
        } : undefined

        // Now we compare the three options and return the one that is closer to the current chapter
        const chapters = [downloadedCh, previousContainerCh, adjacentContainerCh].filter(Boolean)
        if (chapters.length === 0) return undefined
        if (chapters.length === 1) return chapters[0]

        const returnedCh = chapters.reduce((prev, curr) => {
            if (!prev) return curr
            if (!curr) return prev
            const prevDiff = Math.abs(getDecimalFromChapter(prev.chapterNumber) - getDecimalFromChapter(currentChapter.chapterNumber))
            const currDiff = Math.abs(getDecimalFromChapter(curr.chapterNumber) - getDecimalFromChapter(currentChapter.chapterNumber))
            return prevDiff < currDiff ? prev : curr
        }, chapters[0])

        // Make sure to always return the downloaded chapter if it exists
        if (!!downloadedCh && getDecimalFromChapter(downloadedCh.chapterNumber) === getDecimalFromChapter(returnedCh.chapterNumber)) {
            return downloadedCh
        }

        return returnedCh
    }, [mId, currentChapter, entryDownloadedChapters, chapterContainer?.chapters])

    const nextChapter = React.useMemo<MangaReader_SelectedChapter | undefined>(() => {
        if (!mId) return undefined
        if (!currentChapter) return undefined

        // First, look in downloaded chapters
        // e.g., current is 14.2, look for the lowest chapter number that is greater than 14.2
        const _1 = entryDownloadedChapters
            .filter(ch => ch.chapterId !== currentChapter.chapterId)
            .sort((a, b) => getDecimalFromChapter(a.chapterNumber) - getDecimalFromChapter(b.chapterNumber)) // Sort in ascending order
            .find(ch => isChapterAfter(ch.chapterNumber, currentChapter.chapterNumber)) // Find the first chapter that is after the current chapter
        // Save the chapter if it exists
        const downloadedCh = _1 ? {
            chapterId: _1.chapterId,
            chapterNumber: _1.chapterNumber,
            provider: _1.provider as string,
            mediaId: Number(mId),
        } : undefined

        // Return it if there's no container
        if (!chapterContainer?.chapters) return downloadedCh

        // Look for the next chapter in the chapter container
        const idx = chapterContainer.chapters.findIndex((chapter) => chapter.id === currentChapter?.chapterId)

        let nextContainerCh: MangaReader_SelectedChapter | undefined = undefined
        if (idx !== -1 && !!chapterContainer.chapters[idx + 1]) {
            nextContainerCh = {
                chapterId: chapterContainer.chapters[idx + 1].id,
                chapterNumber: chapterContainer.chapters[idx + 1].chapter,
                provider: chapterContainer.chapters[idx + 1].provider as string,
                mediaId: chapterContainer.mediaId,
            }
        }

        // Look in the chapter container, but this time, by sorting the chapters in ascending order to find the adjacent chapter
        let _2 = chapterContainer.chapters
            .filter(ch => ch.id !== currentChapter.chapterId)
            .sort((a, b) => getDecimalFromChapter(a.chapter) - getDecimalFromChapter(b.chapter))
            .find(ch => isChapterAfter(ch.chapter, currentChapter.chapterNumber))
        const adjacentContainerCh = _2 ? {
            chapterId: _2.id,
            chapterNumber: _2.chapter,
            provider: _2.provider as string,
            mediaId: chapterContainer.mediaId,
        } : undefined

        // Now we compare the three options and return the one that is closer to the current chapter
        const chapters = [downloadedCh, nextContainerCh, adjacentContainerCh].filter(Boolean)
        if (chapters.length === 0) return undefined
        if (chapters.length === 1) return chapters[0]

        const returnedCh = chapters.reduce((prev, curr) => {
            if (!prev) return curr
            if (!curr) return prev
            const prevDiff = Math.abs(getDecimalFromChapter(prev.chapterNumber) - getDecimalFromChapter(currentChapter.chapterNumber))
            const currDiff = Math.abs(getDecimalFromChapter(curr.chapterNumber) - getDecimalFromChapter(currentChapter.chapterNumber))
            return prevDiff < currDiff ? prev : curr
        }, chapters[0])

        // Make sure to always return the downloaded chapter if it exists
        if (!!downloadedCh && getDecimalFromChapter(downloadedCh.chapterNumber) === getDecimalFromChapter(returnedCh.chapterNumber)) {
            return downloadedCh
        }

        return returnedCh
    }, [mId, currentChapter, entryDownloadedChapters, chapterContainer?.chapters])

    const goToChapter = React.useCallback((dir: "previous" | "next") => {
        if (dir === "previous" && previousChapter) {
            logger("handle-chapter-reader").info("Going to previous chapter", previousChapter)
            setCurrentChapter(previousChapter)
        } else if (dir === "next" && nextChapter) {
            logger("handle-chapter-reader").info("Going to next chapter", nextChapter)
            setCurrentChapter(nextChapter)
        }
    }, [setCurrentChapter, previousChapter, nextChapter])

    return {
        goToChapter,
        previousChapter,
        nextChapter,
    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////


export function useHydrateMangaPaginationMap(pageContainer?: Manga_PageContainer) {
    // Current chapter
    const selectedChapter = useAtomValue(__manga_selectedChapterAtom)
    const readingMode = useAtomValue(__manga_readingModeAtom)

    // Global page index
    const currentPageIndex = useAtomValue(__manga_currentPageIndexAtom)

    const [doublePageOffset, setDoublePageOffset] = useAtom(__manga_doublePageOffsetAtom)

    /**
     * Pagination map is used to determine how many pages to display at once.
     * The key is the index of the map, and the value is an array of page indexes to display.
     * e.g. { 0: [0, 1], 1: [2], 2: [3, 4], ... }
     */
    const setPaginationMap = useSetAtom(__manga_paginationMapAtom)
    const setCurrentMapIndex = useSetAtom(__manga_currentPaginationMapIndexAtom)

    React.useEffect(() => {
        if (!pageContainer?.pages?.length) return

        let _map = {} as Record<number, number[]>
        /**
         * Setting [paginationMap]
         * If the reading mode is PAGED or the page dimensions are not set, we display each page individually.
         * i.e. [0], [1], [2], [3], ...
         */
        if (
            readingMode === MangaReadingMode.PAGED
            || readingMode === MangaReadingMode.LONG_STRIP
            || (!pageContainer.pageDimensions || Object.keys(pageContainer.pageDimensions).length === 0)
        ) {
            let i = 0
            while (i < pageContainer?.pages?.length) {
                _map[i] = [i]
                i++
            }
            setPaginationMap(_map)
        } else {
            // \/ Double Page logic

            /**
             * Get the recurring width of the pages to determine the threshold for displaying full spreads.
             */
            let fullSpreadThreshold = 2000
            const recWidth = getRecurringNumber(Object.values(pageContainer.pageDimensions).map(n => n.width))
            if (!!recWidth && recWidth > 0) {
                fullSpreadThreshold = recWidth + 50 // Add padding to the width to account for any discrepancies
            }

            const map = new Map<number, number[]>()

            /**
             * Hydrate the map with the page indexes to display.
             * This basically groups pages with the same width together.
             * Pages with a width greater than [fullSpreadThreshold] are displayed individually.
             * e.g. If Page index 2 has a larger width -> [0, 1], [2], [3, 4], ...
             */
            let i = 0
            let mapI = 0
            while (i < pageContainer.pages.length) {

                if (doublePageOffset > 0 && i + 1 <= doublePageOffset) {
                    map.set(mapI, [pageContainer.pages[i].index])
                    i++
                    mapI++
                    continue
                }

                const width = pageContainer.pageDimensions?.[i]?.width || 0
                if (width > fullSpreadThreshold) {
                    map.set(mapI, [pageContainer.pages[i].index])
                    i++
                } else if (
                    !!pageContainer.pages[i + 1]
                    && !(!!pageContainer.pageDimensions?.[i + 1]?.width && pageContainer.pageDimensions?.[i + 1]?.width > fullSpreadThreshold)
                ) {
                    map.set(mapI, [pageContainer.pages[i].index, pageContainer.pages[i + 1].index])
                    i += 2
                } else {
                    map.set(mapI, [pageContainer.pages[i].index])
                    i++
                }
                mapI++
            }
            map.forEach((value, key) => {
                _map[key] = value
            })
            // Set the pagination map to the newly created map
            setPaginationMap(_map)
            map.clear()
        }

        /**
         * This handles navigating to the correct map index when switching reading modes
         *
         * After setting the pagination map, we need to determine which map index to scroll to.
         * This is done by finding the map index that contains the current page index.
         */
        let mapIndexToScroll = 0
        for (const [index, pages] of Object.entries(_map)) {
            if (pages.includes(currentPageIndex)) {
                mapIndexToScroll = Number(index)
                break
            }
        }
        // Set the current map index to the map index to scroll to
        setCurrentMapIndex(mapIndexToScroll)

    }, [pageContainer?.pages, readingMode, selectedChapter, doublePageOffset])

}

export function useSwitchSettingsWithKeys() {
    const [readingMode, setReadingMode] = useAtom(__manga_readingModeAtom)
    const [readingDirection, setReadingDirection] = useAtom(__manga_readingDirectionAtom)
    const [pageFit, setPageFit] = useAtom(__manga_pageFitAtom)
    const [pageStretch, setPageStretch] = useAtom(__manga_pageStretchAtom)
    const [doublePageOffset, setDoublePageOffset] = useAtom(__manga_doublePageOffsetAtom)

    const switchValue = (currentValue: string, possibleValues: string[], setValue: (v: any) => void) => {
        const currentIndex = possibleValues.indexOf(currentValue)
        const nextIndex = (currentIndex + 1) % possibleValues.length
        setValue(possibleValues[nextIndex])
    }

    const incrementOffset = () => {
        setDoublePageOffset(prev => Math.max(0, prev + 1))
    }

    const decrementOffset = () => {
        setDoublePageOffset(prev => Math.max(0, prev - 1))
    }

    React.useEffect(() => {
        mousetrap.bind("m", () => switchValue(readingMode, Object.values(MangaReadingMode), setReadingMode))
        mousetrap.bind("d", () => switchValue(readingDirection, Object.values(MangaReadingDirection), setReadingDirection))
        mousetrap.bind("f", () => switchValue(pageFit, Object.values(MangaPageFit), setPageFit))
        mousetrap.bind("s", () => switchValue(pageStretch, Object.values(MangaPageStretch), setPageStretch))
        mousetrap.bind("shift+right", () => incrementOffset())
        mousetrap.bind("shift+left", () => decrementOffset())

        return () => {
            mousetrap.unbind("m")
            mousetrap.unbind("d")
            mousetrap.unbind("f")
            mousetrap.unbind("s")
            mousetrap.unbind("shift+right")
            mousetrap.unbind("shift+left")
        }
    }, [readingMode, readingDirection, pageFit, pageStretch, doublePageOffset])
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

function getRecurringNumber(arr: number[]): number | undefined {
    const counts = new Map<number, number>()

    arr.forEach(num => {
        counts.set(num, (counts.get(num) || 0) + 1)
    })

    let highestCount = 0
    let highestNumber: number | undefined

    counts.forEach((count, num) => {
        if (count > highestCount) {
            highestCount = count
            highestNumber = num
        }
    })

    return highestNumber
}

