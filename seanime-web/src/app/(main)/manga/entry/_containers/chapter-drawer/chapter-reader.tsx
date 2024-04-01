import { MangaPageContainer } from "@/app/(main)/manga/_lib/types"
import { __manga_selectedChapterAtom } from "@/app/(main)/manga/entry/_containers/chapter-drawer/chapter-drawer"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { atom } from "jotai/index"
import { useAtomValue, useSetAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React from "react"
import { useKeyPressEvent } from "react-use"

export const enum MangaReadingDirection {
    LTR = "ltr",
    RTL = "rtl",
}

export const readingDirectionOptions = [
    { value: MangaReadingDirection.LTR, label: "Left to right" },
    { value: MangaReadingDirection.RTL, label: "Right to left" },
]

export const enum MangaReadingMode {
    LONG_STRIP = "long-strip",
    PAGED = "paged",
    DOUBLE_PAGE = "double-page",
}

export const mangaReadingModeOptions = [
    { value: MangaReadingMode.LONG_STRIP, label: "Long strip" },
    { value: MangaReadingMode.PAGED, label: "Paged" },
    // { value: ReadingMode.DOUBLE_PAGE, label: "Double page" },
]
export const __manga_readingDirectionAtom = atomWithStorage<MangaReadingDirection>("sea-manga-reading-direction", MangaReadingDirection.LTR)
export const __manga_readingModeAtom = atomWithStorage<MangaReadingMode>("sea-manga-reading-mode", MangaReadingMode.LONG_STRIP)
export const __manga_isLastPageAtom = atom(false)

export type MangaHorizontalReaderProps = {
    pageContainer: MangaPageContainer | undefined
}

export function MangaHorizontalReader({ pageContainer }: MangaHorizontalReaderProps) {
    // Current chapter
    const selectedChapter = useAtomValue(__manga_selectedChapterAtom)

    const containerRef = React.useRef<HTMLDivElement>(null)

    const readingMode = useAtomValue(__manga_readingModeAtom)
    const setIsLastPage = useSetAtom(__manga_isLastPageAtom)
    const readingDirection = useAtomValue(__manga_readingDirectionAtom)

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
    //
    //     return process.env.NODE_ENV === "development"
    //         ? `http://${window?.location?.hostname}:43211/manga-backups${url}`
    //         : `http://${window?.location?.host}/manga-backups${url}`
    // }

    const currentPages = React.useMemo(() => paginationMap.get(currentMapIndex), [currentMapIndex, paginationMap])
    const twoDisplayed = readingMode === MangaReadingMode.DOUBLE_PAGE && currentPages?.length === 2

    return (
        <div
            className="h-[calc(100dvh-60px)] overflow-y-hidden overflow-x-hidden w-full px-4 space-y-4 select-none relative"
            ref={containerRef}
            tabIndex={-1}
        >
            <div className="w-fit right-6 absolute z-[5] flex items-center bottom-2">
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
            <div className="absolute w-full h-[calc(100dvh-60px)] flex z-[5] cursor-pointer" tabIndex={-1}>
                <div className="h-full w-full flex flex-1 focus-visible:outline-none" onClick={() => onPaginate("left")} tabIndex={-1} />
                <div className="h-full w-full flex flex-1 focus-visible:outline-none" onClick={() => onPaginate("right")} tabIndex={-1} />
            </div>
            <div
                className={cn(
                    twoDisplayed && readingMode === MangaReadingMode.DOUBLE_PAGE && "flex space-x-2 transition-transform duration-300",
                    // twoDisplayed && readingMode === ReadingMode.DOUBLE_PAGE && readingDirection === ReadingDirection.RTL && "flex-row-reverse",
                )}
            >
                {pageContainer?.pages?.toSorted((a, b) => a.index - b.index)?.map((page, index) => (
                    <div
                        key={page.url}
                        className={cn(
                            "w-full h-[calc(100dvh-60px)] scroll-div min-h-[200px] relative page",
                            !currentPages?.includes(index) ? "hidden" : "displayed",
                            (twoDisplayed && readingDirection === MangaReadingDirection.RTL && readingMode === MangaReadingMode.DOUBLE_PAGE && currentPages?.[0] === index)
                            && "before:content-[''] before:absolute before:w-[3%] before:z-[5] before:h-full before:[background:_linear-gradient(-90deg,_rgba(17,_17,_17,_0)_0,_rgba(17,_17,_17,_.3)_100%)]",
                            (twoDisplayed && readingDirection === MangaReadingDirection.RTL && readingMode === MangaReadingMode.DOUBLE_PAGE && currentPages?.[1] === index)
                            && "before:content-[''] before:absolute before:right-0 before:w-[3%] before:z-[5] before:h-full before:[background:_linear-gradient(90deg,_rgba(17,_17,_17,_0)_0,_rgba(17,_17,_17,_.3)_100%)]",
                            (twoDisplayed && readingDirection === MangaReadingDirection.LTR && readingMode === MangaReadingMode.DOUBLE_PAGE && currentPages?.[1] === index)
                            && "before:content-[''] before:absolute before:w-[3%] before:z-[5] before:h-full before:[background:_linear-gradient(-90deg,_rgba(17,_17,_17,_0)_0,_rgba(17,_17,_17,_.3)_100%)]",
                            (twoDisplayed && readingDirection === MangaReadingDirection.LTR && readingMode === MangaReadingMode.DOUBLE_PAGE && currentPages?.[0] === index)
                            && "before:content-[''] before:absolute before:right-0 before:w-[3%] before:z-[5] before:h-full before:[background:_linear-gradient(90deg,_rgba(17,_17,_17,_0)_0,_rgba(17,_17,_17,_.3)_100%)]",
                        )}
                        id={`page-${index}`}
                    >
                        {/*<LoadingSpinner containerClass="h-full absolute inset-0 z-[1] w-24 mx-auto" />*/}
                        <img
                            src={page.url} alt={`Page ${index}`} className={cn(
                            "w-full h-full inset-0 object-contain object-center select-none z-[4] relative",
                            (twoDisplayed && readingDirection === MangaReadingDirection.RTL && readingMode === MangaReadingMode.DOUBLE_PAGE && currentPages?.[0] === index)
                            && "[object-position:0%_50%] before:content-['']",
                            (twoDisplayed && readingDirection === MangaReadingDirection.RTL && readingMode === MangaReadingMode.DOUBLE_PAGE && currentPages?.[1] === index)
                            && "[object-position:100%_50%]",
                            (twoDisplayed && readingDirection === MangaReadingDirection.LTR && readingMode === MangaReadingMode.DOUBLE_PAGE && currentPages?.[0] === index)
                            && "[object-position:100%_50%]",
                            (twoDisplayed && readingDirection === MangaReadingDirection.LTR && readingMode === MangaReadingMode.DOUBLE_PAGE && currentPages?.[1] === index)
                            && "[object-position:0%_50%]",
                        )}
                        />
                    </div>
                ))}
            </div>

        </div>
    )
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export type MangaVerticalReaderProps = {
    pageContainer: MangaPageContainer | undefined
}

export function MangaVerticalReader({ pageContainer }: MangaVerticalReaderProps) {

    const containerRef = React.useRef<HTMLDivElement>(null)
    const setIsLastPage = useSetAtom(__manga_isLastPageAtom)

    const [currentPageIndex, setCurrentPageIndex] = React.useState<number>(0)

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
            <div className="w-fit right-6 absolute z-[5] flex items-center bottom-2 focus-visible:outline-none" tabIndex={-1}>
                {!!(currentPageIndex + 1) && (
                    <p className="text-[--muted]">
                        {currentPageIndex + 1} / {pageContainer?.pages?.length}
                    </p>
                )}
            </div>
            <div
                className="w-full h-[calc(100dvh-60px)] overflow-y-auto overflow-x-hidden px-4 space-y-4 select-none relative focus-visible:outline-none"
                ref={containerRef}
                tabIndex={-1}
            >
                <div className="absolute w-full h-full z-[5] focus-visible:outline-none" tabIndex={-1}>

                </div>
                {pageContainer?.pages?.map((page, index) => (
                    <div
                        key={page.url}
                        className="max-w-[1400px] mx-auto scroll-div min-h-[200px] relative focus-visible:outline-none"
                        id={`page-${index}`}
                        tabIndex={-1}
                    >
                        <LoadingSpinner containerClass="h-full absolute inset-0 z-[1] focus-visible:outline-none" tabIndex={-1} />
                        <img
                            src={page.url}
                            alt={`Page ${index}`}
                            className="max-w-full h-auto mx-auto select-none z-[4] relative focus-visible:outline-none"
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
