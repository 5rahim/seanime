import { Manga_PageContainer } from "@/api/generated/types"
import { ChapterPage } from "@/app/(main)/manga/_containers/chapter-reader/_components/chapter-page"
import { useHandleChapterPageStatus, useHydrateMangaPaginationMap } from "@/app/(main)/manga/_lib/handle-chapter-reader"
import {
    __manga_currentPageIndexAtom,
    __manga_currentPaginationMapIndexAtom,
    __manga_hiddenBarAtom,
    __manga_isLastPageAtom,
    __manga_kbsPageLeft,
    __manga_kbsPageRight,
    __manga_pageFitAtom,
    __manga_pageGapAtom,
    __manga_pageGapShadowAtom,
    __manga_pageOverflowContainerWidthAtom,
    __manga_paginationMapAtom,
    __manga_readingDirectionAtom,
    __manga_readingModeAtom,
    MangaPageFit,
    MangaReadingDirection,
    MangaReadingMode,
} from "@/app/(main)/manga/_lib/manga-chapter-reader.atoms"
import { cn } from "@/components/ui/core/styling"
import { isMobile } from "@/lib/utils/browser-detection"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import mousetrap from "mousetrap"
import React from "react"
import { useUpdateEffect } from "react-use"

export type MangaHorizontalReaderProps = {
    pageContainer: Manga_PageContainer | undefined
}

export function MangaHorizontalReader({ pageContainer }: MangaHorizontalReaderProps) {
    const containerRef = React.useRef<HTMLDivElement>(null)
    const pageWrapperRef = React.useRef<HTMLDivElement>(null)

    const readingMode = useAtomValue(__manga_readingModeAtom)
    const setIsLastPage = useSetAtom(__manga_isLastPageAtom)
    const readingDirection = useAtomValue(__manga_readingDirectionAtom)
    const pageFit = useAtomValue(__manga_pageFitAtom)
    const pageGap = useAtomValue(__manga_pageGapAtom)
    const pageGapShadow = useAtomValue(__manga_pageGapShadowAtom)
    const pageOverflowContainerWidth = useAtomValue(__manga_pageOverflowContainerWidthAtom)

    const [hiddenBar, setHideBar] = useAtom(__manga_hiddenBarAtom)

    const kbsPageLeft = useAtomValue(__manga_kbsPageLeft)
    const kbsPageRight = useAtomValue(__manga_kbsPageRight)

    // Global page index
    const setCurrentPageIndex = useSetAtom(__manga_currentPageIndexAtom)

    const paginationMap = useAtomValue(__manga_paginationMapAtom)

    /**
     * For this horizontal reader [currentMapIndex] is the actual variable that controls what pages are displayed
     * [currentPageIndex] is updated AFTER [currentMapIndex] changes
     */
    const [currentMapIndex, setCurrentMapIndex] = useAtom(__manga_currentPaginationMapIndexAtom)

    useHydrateMangaPaginationMap(pageContainer)

    const { handlePageLoad } = useHandleChapterPageStatus(pageContainer)

    /**
     * When the current map index changes, scroll to the top of the container
     */
    useUpdateEffect(() => {
        containerRef.current?.scrollTo({ top: 0 })
    }, [currentMapIndex])

    /**
     * Set [isLastPage] state when the current map index changes
     */
    React.useEffect(() => {
        setIsLastPage(Object.keys(paginationMap).length > 0 && currentMapIndex === Object.keys(paginationMap).length - 1)
    }, [currentMapIndex, paginationMap])

    /**
     * Function to handle page navigation when the user clicks on the left or right side of the page
     */
    const onPaginate = React.useCallback((dir: "left" | "right") => {
        const shouldDecrement = dir === "left" && readingDirection === MangaReadingDirection.LTR || dir === "right" && readingDirection === MangaReadingDirection.RTL

        setCurrentMapIndex((draft) => {
            const newIdx = shouldDecrement ? draft - 1 : draft + 1
            if (paginationMap.hasOwnProperty(newIdx)) {
                return newIdx
            }
            return draft
        })
    }, [paginationMap, readingDirection])

    /**
     * Key bindings for page navigation
     */
    React.useEffect(() => {
        mousetrap.bind(kbsPageLeft, () => onPaginate("left"))
        mousetrap.bind(kbsPageRight, () => onPaginate("right"))
        mousetrap.bind("up", () => {
            containerRef.current?.scrollBy(0, -100)
        })
        mousetrap.bind("down", () => {
            containerRef.current?.scrollBy(0, 100)
        })

        return () => {
            mousetrap.unbind(kbsPageLeft)
            mousetrap.unbind(kbsPageRight)
            mousetrap.unbind("up")
            mousetrap.unbind("down")
        }
    }, [kbsPageLeft, kbsPageRight, paginationMap, readingDirection])

    /**
     * Function to handle page navigation when the user clicks on the left or right side of the page
     */
    const onPageWrapperClick = React.useCallback((e: React.MouseEvent<HTMLDivElement, MouseEvent>) => {
        if (!pageWrapperRef.current) return

        if ((e.target as HTMLElement).id === "retry-button") return
        if ((e.target as HTMLElement).id === "retry-icon") return

        const { clientX } = e.nativeEvent
        const divWidth = pageWrapperRef.current.offsetWidth
        const clickPosition = clientX - pageWrapperRef.current.getBoundingClientRect().left
        const clickPercentage = (clickPosition / divWidth) * 100

        if (clickPercentage <= 40) {
            onPaginate("left")
        } else if (clickPercentage >= 60) {
            onPaginate("right")
        } else {
            if (!isMobile()) {
                setHideBar(prev => !prev)
            }
        }
    }, [onPaginate, pageWrapperRef.current])

    /**
     * Update the current page index when the current map index changes
     */
    React.useEffect(() => {
        if (!pageContainer?.pages?.length) return

        const currentPages = paginationMap[currentMapIndex]
        if (!currentPages) return

        setCurrentPageIndex(currentPages[0])
    }, [currentMapIndex])

    // Current page indexes displayed
    const currentPages = React.useMemo(() => paginationMap[currentMapIndex], [currentMapIndex, paginationMap])
    // Two pages are currently displayed
    const twoPages = readingMode === MangaReadingMode.DOUBLE_PAGE && currentPages?.length === 2
    // Show shadows
    const showShadows = twoPages && pageGap && !(pageFit === MangaPageFit.COVER || pageFit === MangaPageFit.TRUE_SIZE) && pageGapShadow

    return (
        <div
            data-chapter-horizontal-reader-container
            className={cn(
                "h-[calc(100dvh-3rem)] overflow-y-hidden overflow-x-hidden w-full px-4 select-none relative",
                hiddenBar && "h-dvh max-h-full",
                "focus-visible:outline-none",
                pageFit === MangaPageFit.COVER && "overflow-y-auto",
                pageFit === MangaPageFit.TRUE_SIZE && "overflow-y-auto",
                pageFit === MangaPageFit.LARGER && "overflow-y-auto",

                // Double page + PageFit = LARGER
                pageFit === MangaPageFit.LARGER && readingMode === MangaReadingMode.DOUBLE_PAGE && "w-full px-40 mx-auto",
            )}
            ref={containerRef}
            tabIndex={-1}
        >
            {/*<div className="absolute w-full h-full right-8 flex z-[5] cursor-pointer" tabIndex={-1}>*/}
            {/*    <div className="h-full w-full flex flex-1 focus-visible:outline-none" onClick={() => onPaginate("left")} tabIndex={-1} />*/}
            {/*    <div className="h-full w-full flex flex-1 focus-visible:outline-none" onClick={() => onPaginate("right")} tabIndex={-1} />*/}
            {/*</div>*/}
            <div
                data-chapter-horizontal-reader-page-wrapper
                className={cn(
                    "focus-visible:outline-none",
                    twoPages && readingMode === MangaReadingMode.DOUBLE_PAGE && "flex transition-transform duration-300",
                    twoPages && readingMode === MangaReadingMode.DOUBLE_PAGE && pageGap && "gap-2",
                    twoPages && readingMode === MangaReadingMode.DOUBLE_PAGE && "flex-row-reverse",
                )}
                ref={pageWrapperRef}
                onClick={onPageWrapperClick}
            >
                {pageContainer?.pages?.toSorted((a, b) => a.index - b.index)?.map((page, index) => (
                    <ChapterPage
                        key={page.url}
                        page={page}
                        index={index}
                        readingMode={readingMode}
                        pageContainer={pageContainer}
                        onFinishedLoading={() => {
                            handlePageLoad(index)
                        }}
                        containerClass={cn(
                            "w-full h-[calc(100dvh-3rem)] scroll-div min-h-[200px] relative page",
                            hiddenBar && "h-dvh max-h-full",
                            "focus-visible:outline-none",
                            !currentPages?.includes(index) ? "hidden" : "displayed",
                            // Double Page, gap
                            (showShadows && readingMode === MangaReadingMode.DOUBLE_PAGE && currentPages?.[0] === index)
                            && "before:content-[''] before:absolute before:w-[3%] before:z-[5] before:h-full before:[background:_linear-gradient(-90deg,_rgba(17,_17,_17,_0)_0,_rgba(17,_17,_17,_.3)_100%)]",
                            (showShadows && readingMode === MangaReadingMode.DOUBLE_PAGE && currentPages?.[1] === index)
                            && "before:content-[''] before:absolute before:right-0 before:w-[3%] before:z-[5] before:h-full before:[background:_linear-gradient(90deg,_rgba(17,_17,_17,_0)_0,_rgba(17,_17,_17,_.3)_100%)]",
                            // Page fit
                            pageFit === MangaPageFit.LARGER && "h-full",
                        )}
                        imageClass={cn(
                            "focus-visible:outline-none",
                            "h-full inset-0 object-center select-none z-[4] relative",

                            //
                            // Page fit
                            //

                            // Single page
                            (readingMode === MangaReadingMode.PAGED
                                && pageFit === MangaPageFit.CONTAIN) && "object-contain w-full h-full",
                            (readingMode === MangaReadingMode.PAGED
                                && pageFit === MangaPageFit.LARGER) && "h-auto object-cover mx-auto",
                            (readingMode === MangaReadingMode.PAGED
                                && pageFit === MangaPageFit.COVER) && "w-full h-auto",
                            (readingMode === MangaReadingMode.PAGED
                                && pageFit === MangaPageFit.TRUE_SIZE) && "object-none h-auto w-auto mx-auto",
                            // Double page
                            (readingMode === MangaReadingMode.DOUBLE_PAGE
                                && pageFit === MangaPageFit.CONTAIN) && "object-contain w-full h-full",
                            (readingMode === MangaReadingMode.DOUBLE_PAGE
                                && pageFit === MangaPageFit.LARGER) && "w-[1400px] h-auto object-cover mx-auto",
                            (readingMode === MangaReadingMode.DOUBLE_PAGE
                                && pageFit === MangaPageFit.COVER) && "w-full h-auto",
                            (readingMode === MangaReadingMode.DOUBLE_PAGE
                                && pageFit === MangaPageFit.TRUE_SIZE) && cn(
                                "object-none h-auto w-auto",
                                (twoPages && currentPages?.[0] === index)
                                    ? "mr-auto" :
                                    (twoPages && currentPages?.[1] === index)
                                        ? "ml-auto" : "mx-auto",
                            ),

                            //
                            // Double page - Page position
                            //
                            (twoPages && currentPages?.[0] === index)
                            && "[object-position:0%_50%] before:content-['']",
                            (twoPages && currentPages?.[1] === index)
                            && "[object-position:100%_50%]",
                        )}
                        imageWidth={pageFit === MangaPageFit.LARGER && readingMode === MangaReadingMode.PAGED
                            ? pageOverflowContainerWidth + "%"
                            : undefined}
                    />
                ))}
            </div>

        </div>
    )
}
