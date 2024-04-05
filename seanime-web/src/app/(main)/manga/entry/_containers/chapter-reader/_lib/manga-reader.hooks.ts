import { MangaPageContainer } from "@/app/(main)/manga/_lib/manga.types"
import {
    __manga_currentPageIndexAtom,
    __manga_currentPaginationMapIndexAtom,
    __manga_pageFitAtom,
    __manga_pageStretchAtom,
    __manga_paginationMapAtom,
    __manga_readingDirectionAtom,
    __manga_readingModeAtom,
    MangaPageFit,
    MangaPageStretch,
    MangaReadingDirection,
    MangaReadingMode,
} from "@/app/(main)/manga/entry/_containers/chapter-reader/_lib/manga-chapter-reader.atoms"
import { __manga_selectedChapterAtom } from "@/app/(main)/manga/entry/_containers/chapter-reader/chapter-reader-drawer"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import mousetrap from "mousetrap"
import React from "react"

export function useHydrateMangaPaginationMap(pageContainer?: MangaPageContainer) {
    // Current chapter
    const selectedChapter = useAtomValue(__manga_selectedChapterAtom)
    const readingMode = useAtomValue(__manga_readingModeAtom)

    // Global page index
    const currentPageIndex = useAtomValue(__manga_currentPageIndexAtom)

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
                fullSpreadThreshold = recWidth
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
                const width = pageContainer.pageDimensions?.[i]?.width || 0
                if (width > fullSpreadThreshold) {
                    map.set(mapI, [pageContainer.pages[i].index])
                    i++
                } else if (!!pageContainer.pages[i + 1] && !(!!pageContainer.pageDimensions?.[i + 1]?.width && pageContainer.pageDimensions?.[i + 1]?.width > fullSpreadThreshold)) {
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

    }, [pageContainer?.pages, readingMode, selectedChapter])

}

export function useSwitchSettingsWithKeys() {
    const [readingMode, setReadingMode] = useAtom(__manga_readingModeAtom)
    const [readingDirection, setReadingDirection] = useAtom(__manga_readingDirectionAtom)
    const [pageFit, setPageFit] = useAtom(__manga_pageFitAtom)
    const [pageStretch, setPageStretch] = useAtom(__manga_pageStretchAtom)

    const switchValue = (currentValue: string, possibleValues: string[], setValue: (v: any) => void) => {
        const currentIndex = possibleValues.indexOf(currentValue)
        const nextIndex = (currentIndex + 1) % possibleValues.length
        setValue(possibleValues[nextIndex])
    }

    React.useEffect(() => {
        mousetrap.bind("m", () => switchValue(readingMode, Object.values(MangaReadingMode), setReadingMode))
        mousetrap.bind("d", () => switchValue(readingDirection, Object.values(MangaReadingDirection), setReadingDirection))
        mousetrap.bind("f", () => switchValue(pageFit, Object.values(MangaPageFit), setPageFit))
        mousetrap.bind("s", () => switchValue(pageStretch, Object.values(MangaPageStretch), setPageStretch))

        return () => {
            mousetrap.unbind("m")
            mousetrap.unbind("d")
            mousetrap.unbind("f")
            mousetrap.unbind("s")
        }
    }, [readingMode, readingDirection, pageFit, pageStretch])
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
