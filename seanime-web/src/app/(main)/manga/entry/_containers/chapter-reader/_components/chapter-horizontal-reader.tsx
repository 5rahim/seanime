import { MangaPageContainer } from "@/app/(main)/manga/_lib/manga.types"
import {
    __manga_isLastPageAtom,
    __manga_pageFitAtom,
    __manga_pageGapAtom,
    __manga_pageStretchAtom,
    __manga_readingDirectionAtom,
    __manga_readingModeAtom,
    MangaPageFit,
    MangaReadingDirection,
    MangaReadingMode,
} from "@/app/(main)/manga/entry/_containers/chapter-reader/_lib/manga.atoms"
import { __manga_selectedChapterAtom } from "@/app/(main)/manga/entry/_containers/chapter-reader/chapter-reader-drawer"
import { cn } from "@/components/ui/core/styling"
import { useAtomValue, useSetAtom } from "jotai/react"
import React from "react"
import { useKeyPressEvent } from "react-use"

export type MangaHorizontalReaderProps = {
    pageContainer: MangaPageContainer | undefined
}

export function MangaHorizontalReader({ pageContainer }: MangaHorizontalReaderProps) {
    // Current chapter
    const selectedChapter = useAtomValue(__manga_selectedChapterAtom)

    const containerRef = React.useRef<HTMLDivElement>(null)
    const pageWrapperRef = React.useRef<HTMLDivElement>(null)

    const readingMode = useAtomValue(__manga_readingModeAtom)
    const setIsLastPage = useSetAtom(__manga_isLastPageAtom)
    const readingDirection = useAtomValue(__manga_readingDirectionAtom)
    const pageFit = useAtomValue(__manga_pageFitAtom)
    const pageStretch = useAtomValue(__manga_pageStretchAtom)
    const pageGap = useAtomValue(__manga_pageGapAtom)

    const [currentMapIndex, setCurrentMapIndex] = React.useState<number>(0)

    const paginationMap = React.useMemo(() => {
        setCurrentMapIndex(0)
        if (!pageContainer?.pages?.length) return new Map<number, number[]>()
        if (readingMode === MangaReadingMode.PAGED || !pageContainer.pageDimensions) {
            let i = 0
            const map = new Map<number, number[]>()
            while (i < pageContainer?.pages?.length) {
                map.set(i, [i])
                i++
            }
            return map
        }
        // idx -> [a, b]
        const map = new Map<number, number[]>()
        // if page x is over 2000px, we display it alone, else we display pairs
        // e.g. [[0, 1], [2], [3], [4, 5], [6], [7, 8], ...]
        let i = 0
        let mapI = 0
        while (i < pageContainer.pages.length) {
            const width = pageContainer.pageDimensions?.[i]?.width || 0
            if (width > 2000) {
                map.set(mapI, [pageContainer.pages[i].index])
                i++
            } else if (!!pageContainer.pages[i + 1] && !(!!pageContainer.pageDimensions?.[i + 1]?.width && pageContainer.pageDimensions?.[i + 1]?.width > 2000)) {
                map.set(mapI, [pageContainer.pages[i].index, pageContainer.pages[i + 1].index])
                i += 2
            } else {
                map.set(mapI, [pageContainer.pages[i].index])
                i++
            }
            mapI++
        }
        return map
    }, [pageContainer?.pages, readingMode, selectedChapter])

    // Handle pagination
    const onPaginate = React.useCallback((dir: "left" | "right") => {
        const shouldDecrement = dir === "left" && readingDirection === MangaReadingDirection.LTR || dir === "right" && readingDirection === MangaReadingDirection.RTL

        setCurrentMapIndex((draft) => {
            const newIdx = shouldDecrement ? draft - 1 : draft + 1
            if (paginationMap.has(newIdx)) {
                return newIdx
            }
            return draft
        })
    }, [paginationMap, readingDirection])

    // Arrow key navigation
    useKeyPressEvent("ArrowLeft", () => onPaginate("left"))
    useKeyPressEvent("ArrowRight", () => onPaginate("right"))

    React.useEffect(() => {
        setIsLastPage(paginationMap.size > 0 && currentMapIndex === paginationMap.size - 1)
    }, [currentMapIndex, paginationMap])

    // const getSrc = (url: string) => {
    //     if (!pageContainer?.isDownloaded) {
    //         return url
    //     }
    //     return process.env.NODE_ENV === "development"
    //         ? `http://${window?.location?.hostname}:43211/manga-backups${url}`
    //         : `http://${window?.location?.host}/manga-backups${url}`
    // }

    const onPageWrapperClick = React.useCallback((e: React.MouseEvent<HTMLDivElement, MouseEvent>) => {
        if (!pageWrapperRef.current) return

        const { clientX } = e.nativeEvent
        const divWidth = pageWrapperRef.current.offsetWidth
        const clickPosition = clientX - pageWrapperRef.current.getBoundingClientRect().left
        const clickPercentage = (clickPosition / divWidth) * 100

        if (clickPercentage <= 40) {
            onPaginate("left")
        } else if (clickPercentage >= 60) {
            onPaginate("right")
        }
    }, [onPaginate, pageWrapperRef.current])

    const currentPages = React.useMemo(() => paginationMap.get(currentMapIndex), [currentMapIndex, paginationMap])
    const showShadows = readingMode === MangaReadingMode.DOUBLE_PAGE && currentPages?.length === 2 && pageGap

    return (
        <div
            className={cn(
                "h-[calc(100dvh-60px)] overflow-y-hidden overflow-x-hidden w-full px-4 space-y-4 select-none relative",
                "focus-visible:outline-none",
                pageFit === MangaPageFit.COVER && "overflow-y-auto",
                pageFit === MangaPageFit.TRUE_SIZE && "overflow-y-auto",
                pageFit === MangaPageFit.LARGER && "overflow-y-auto",
            )}
            ref={containerRef}
            tabIndex={-1}
        >
            <div className="w-fit right-6 fixed z-[5] flex items-center bottom-2 focus-visible:outline-none">
                {!!currentPages?.length && (
                    readingMode === MangaReadingMode.DOUBLE_PAGE ? (
                        <p className="text-[--muted]">
                            {currentPages?.length > 1
                                ? `${currentPages[0] + 1}-${currentPages[1] + 1}`
                                : currentPages[0] + 1} / {pageContainer?.pages?.length}
                        </p>
                    ) : (
                        <p className="text-[--muted]">
                            {currentPages[0] + 1} / {pageContainer?.pages?.length}
                        </p>
                    )
                )}
            </div>
            {/*<div className="absolute w-full h-full right-8 flex z-[5] cursor-pointer" tabIndex={-1}>*/}
            {/*    <div className="h-full w-full flex flex-1 focus-visible:outline-none" onClick={() => onPaginate("left")} tabIndex={-1} />*/}
            {/*    <div className="h-full w-full flex flex-1 focus-visible:outline-none" onClick={() => onPaginate("right")} tabIndex={-1} />*/}
            {/*</div>*/}
            <div
                className={cn(
                    "focus-visible:outline-none",
                    showShadows && readingMode === MangaReadingMode.DOUBLE_PAGE && "flex transition-transform duration-300",
                    pageGap && "space-y-4",
                    // showShadows && readingMode === ReadingMode.DOUBLE_PAGE && readingDirection === ReadingDirection.RTL && "flex-row-reverse",
                )}
                ref={pageWrapperRef}
                onClick={onPageWrapperClick}
            >
                {pageContainer?.pages?.toSorted((a, b) => a.index - b.index)?.map((page, index) => (
                    <div
                        key={page.url}
                        className={cn(
                            "w-full h-[calc(100dvh-60px)] scroll-div min-h-[200px] relative page",
                            "focus-visible:outline-none",
                            !currentPages?.includes(index) ? "hidden" : "displayed",
                            (showShadows && readingDirection === MangaReadingDirection.RTL && readingMode === MangaReadingMode.DOUBLE_PAGE && currentPages?.[0] === index)
                            && "before:content-[''] before:absolute before:w-[3%] before:z-[5] before:h-full before:[background:_linear-gradient(-90deg,_rgba(17,_17,_17,_0)_0,_rgba(17,_17,_17,_.3)_100%)]",
                            (showShadows && readingDirection === MangaReadingDirection.RTL && readingMode === MangaReadingMode.DOUBLE_PAGE && currentPages?.[1] === index)
                            && "before:content-[''] before:absolute before:right-0 before:w-[3%] before:z-[5] before:h-full before:[background:_linear-gradient(90deg,_rgba(17,_17,_17,_0)_0,_rgba(17,_17,_17,_.3)_100%)]",
                            (showShadows && readingDirection === MangaReadingDirection.LTR && readingMode === MangaReadingMode.DOUBLE_PAGE && currentPages?.[1] === index)
                            && "before:content-[''] before:absolute before:w-[3%] before:z-[5] before:h-full before:[background:_linear-gradient(-90deg,_rgba(17,_17,_17,_0)_0,_rgba(17,_17,_17,_.3)_100%)]",
                            (showShadows && readingDirection === MangaReadingDirection.LTR && readingMode === MangaReadingMode.DOUBLE_PAGE && currentPages?.[0] === index)
                            && "before:content-[''] before:absolute before:right-0 before:w-[3%] before:z-[5] before:h-full before:[background:_linear-gradient(90deg,_rgba(17,_17,_17,_0)_0,_rgba(17,_17,_17,_.3)_100%)]",
                        )}
                        id={`page-${index}`}
                    >
                        {/*<LoadingSpinner containerClass="h-full absolute inset-0 z-[1] w-24 mx-auto" />*/}
                        <img
                            src={page.url} alt={`Page ${index}`} className={cn(
                            "focus-visible:outline-none",
                            "h-full inset-0 object-center select-none z-[4] relative",

                            pageFit === MangaPageFit.CONTAIN && "object-contain w-full h-full",
                            pageFit === MangaPageFit.LARGER && "w-[1400px] h-auto object-cover mx-auto",
                            pageFit === MangaPageFit.COVER && "w-full h-auto",
                            pageFit === MangaPageFit.TRUE_SIZE && "object-none h-auto w-auto mx-auto",

                            (showShadows && readingDirection === MangaReadingDirection.RTL && readingMode === MangaReadingMode.DOUBLE_PAGE && currentPages?.[0] === index)
                            && "[object-position:0%_50%] before:content-['']",
                            (showShadows && readingDirection === MangaReadingDirection.RTL && readingMode === MangaReadingMode.DOUBLE_PAGE && currentPages?.[1] === index)
                            && "[object-position:100%_50%]",
                            (showShadows && readingDirection === MangaReadingDirection.LTR && readingMode === MangaReadingMode.DOUBLE_PAGE && currentPages?.[0] === index)
                            && "[object-position:100%_50%]",
                            (showShadows && readingDirection === MangaReadingDirection.LTR && readingMode === MangaReadingMode.DOUBLE_PAGE && currentPages?.[1] === index)
                            && "[object-position:0%_50%]",
                        )}
                        />
                    </div>
                ))}
            </div>

        </div>
    )
}
