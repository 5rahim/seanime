import { MangaPageContainer } from "@/app/(main)/manga/_lib/manga.types"
import {
    __manga_currentPageIndexAtom,
    __manga_currentPaginationMapIndexAtom,
    __manga_isLastPageAtom,
    __manga_pageFitAtom,
    __manga_pageGapAtom,
    __manga_pageStretchAtom,
    __manga_paginationMapAtom,
    MangaPageFit,
    MangaPageStretch,
} from "@/app/(main)/manga/entry/_containers/chapter-reader/_lib/manga.atoms"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import React from "react"
import { useEffectOnce, useKeyPressEvent } from "react-use"

export type MangaVerticalReaderProps = {
    pageContainer: MangaPageContainer | undefined
}

export function MangaVerticalReader({ pageContainer }: MangaVerticalReaderProps) {

    const containerRef = React.useRef<HTMLDivElement>(null)
    const setIsLastPage = useSetAtom(__manga_isLastPageAtom)
    const pageFit = useAtomValue(__manga_pageFitAtom)
    const pageStretch = useAtomValue(__manga_pageStretchAtom)
    const pageGap = useAtomValue(__manga_pageGapAtom)

    const [currentPageIndex, setCurrentPageIndex] = useAtom(__manga_currentPageIndexAtom)
    const [currentMapIndex, setCurrentMapIndex] = useAtom(__manga_currentPaginationMapIndexAtom)
    const [paginationMap, setPaginationMap] = useAtom(__manga_paginationMapAtom)

    React.useEffect(() => {
        setCurrentMapIndex(0)

        if (!pageContainer?.pages?.length) {
            setPaginationMap({})
            return
        }

        let i = 0
        const map = {} as Record<number, number[]>
        while (i < pageContainer?.pages?.length) {
            map[i] = [i]
            i++
        }
        setPaginationMap(map)
        return
    }, [pageContainer?.pages])

    useEffectOnce(() => {
        if (currentPageIndex !== 0) {
            const pageDiv = containerRef.current?.querySelector(`#page-${currentPageIndex}`)
            pageDiv?.scrollIntoView({ behavior: "smooth" })
        }
    })

    // Function to handle scroll event
    const handleScroll = () => {
        if (!!containerRef.current) {
            const scrollTop = containerRef.current.scrollTop
            const scrollHeight = containerRef.current.scrollHeight
            const clientHeight = containerRef.current.clientHeight

            if (scrollTop > 1000 && !!pageContainer?.pages?.length && (scrollTop + clientHeight >= scrollHeight - 1500)) {
                setIsLastPage(true)
            } else {
                setIsLastPage(false)
            }

            containerRef.current?.querySelectorAll(".scroll-div")?.forEach((div) => {
                if (isElementXPercentInViewport(div) && pageContainer?.pages?.length) {
                    const idx = Number(div.id.split("-")[1])
                    setCurrentPageIndex(idx)
                }
            })
        }
    }

    React.useEffect(() => {
        setIsLastPage(false)
    }, [pageContainer?.pages])

    // Add scroll event listener when component mounts
    React.useEffect(() => {
        containerRef.current?.addEventListener("scroll", handleScroll)
        return () => containerRef.current?.removeEventListener("scroll", handleScroll)
    }, [containerRef.current])

    useKeyPressEvent("ArrowUp", () => {
        containerRef.current?.scrollBy(0, -50)
    })

    useKeyPressEvent("ArrowDown", () => {
        containerRef.current?.scrollBy(0, 50)
    })


    return (
        <div className="max-h-[calc(100dvh-3rem)] relative focus-visible:outline-none" tabIndex={-1}>
            {/*<div className="w-fit right-6 absolute z-[5] flex items-center bottom-2 focus-visible:outline-none" tabIndex={-1}>*/}
            {/*    {!!(currentPageIndex + 1) && (*/}
            {/*        <p className="text-[--muted]">*/}
            {/*            {currentPageIndex + 1} / {pageContainer?.pages?.length}*/}
            {/*        </p>*/}
            {/*    )}*/}
            {/*</div>*/}
            <div
                className={cn(
                    "w-full h-[calc(100dvh-60px)] overflow-y-auto overflow-x-hidden px-4 select-none relative focus-visible:outline-none",
                    pageGap && "space-y-4",
                )}
                ref={containerRef}
                tabIndex={-1}
            >
                <div className="absolute w-full h-full z-[5] focus-visible:outline-none" tabIndex={-1}>

                </div>
                {pageContainer?.pages?.map((page, index) => (
                    <div
                        key={page.url}
                        className={cn(
                            "max-w-[1400px] mx-auto scroll-div min-h-[200px] relative focus-visible:outline-none",
                            pageFit === MangaPageFit.CONTAIN && "max-w-full h-[calc(100dvh-60px)]",
                            pageFit === MangaPageFit.TRUE_SIZE && "max-w-full",
                            pageFit === MangaPageFit.COVER && "max-w-full",
                        )}
                        id={`page-${index}`}
                        tabIndex={-1}
                    >
                        <LoadingSpinner containerClass="h-full absolute inset-0 z-[1] focus-visible:outline-none" tabIndex={-1} />
                        <img
                            src={page.url}
                            alt={`Page ${index}`}
                            className={cn(
                                "max-w-full h-auto mx-auto select-none z-[4] relative focus-visible:outline-none",

                                // "h-full inset-0 object-center select-none z-[4] relative",

                                pageFit === MangaPageFit.CONTAIN ?
                                    pageStretch === MangaPageStretch.NONE ? "w-auto h-full object-center" : "object-fill w-[1400px] h-full" :
                                    undefined,
                                pageFit === MangaPageFit.LARGER ?
                                    pageStretch === MangaPageStretch.NONE ? "w-auto h-full object-center" : "w-[1400px] h-auto object-cover mx-auto" :
                                    undefined,
                                pageFit === MangaPageFit.COVER && "w-full h-auto",
                                pageFit === MangaPageFit.TRUE_SIZE && "object-none h-auto w-auto mx-auto",
                            )}
                        />
                    </div>
                ))}

            </div>
        </div>
    )
}

// source: https://stackoverflow.com/a/51121566
const isElementXPercentInViewport = function (el: any, percentVisible = 30) {
    let
        rect = el.getBoundingClientRect(),
        windowHeight = (window.innerHeight || document.documentElement.clientHeight)

    return !(
        Math.floor(100 - (((rect.top >= 0 ? 0 : rect.top) / +-rect.height) * 100)) < percentVisible ||
        Math.floor(100 - ((rect.bottom - windowHeight) / rect.height) * 100) < percentVisible
    )
}
