import { Manga_PageContainer } from "@/api/generated/types"
import { ChapterPage } from "@/app/(main)/manga/_containers/chapter-reader/_components/chapter-page"
import { useHandleChapterPageStatus, useHydrateMangaPaginationMap } from "@/app/(main)/manga/_lib/handle-chapter-reader"
import {
    __manga_currentPageIndexAtom,
    __manga_hiddenBarAtom,
    __manga_isLastPageAtom,
    __manga_kbsPageLeft,
    __manga_kbsPageRight,
    __manga_pageFitAtom,
    __manga_pageGapAtom,
    __manga_pageOverflowContainerWidthAtom,
    __manga_pageStretchAtom,
    __manga_paginationMapAtom,
    MangaPageFit,
    MangaPageStretch,
} from "@/app/(main)/manga/_lib/manga-chapter-reader.atoms"
import { useUpdateEffect } from "@/components/ui/core/hooks"
import { cn } from "@/components/ui/core/styling"
import { isMobile } from "@/lib/utils/browser-detection"
import { atom } from "jotai"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import mousetrap from "mousetrap"
import React from "react"
import { useEffectOnce } from "react-use"

export type MangaVerticalReaderProps = {
    pageContainer: Manga_PageContainer | undefined
}

export const ___manga_scrollSignalAtom = atom(0)

/**
 * MangaVerticalReader component
 *
 * This component is responsible for rendering the manga pages in a vertical layout.
 * It also handles the logic for scrolling and page navigation.
 */
export function MangaVerticalReader({ pageContainer }: MangaVerticalReaderProps) {

    const containerRef = React.useRef<HTMLDivElement>(null)
    const setIsLastPage = useSetAtom(__manga_isLastPageAtom)
    const pageFit = useAtomValue(__manga_pageFitAtom)
    const pageStretch = useAtomValue(__manga_pageStretchAtom)
    const pageGap = useAtomValue(__manga_pageGapAtom)
    const pageOverflowContainerWidth = useAtomValue(__manga_pageOverflowContainerWidthAtom)
    const [currentPageIndex, setCurrentPageIndex] = useAtom(__manga_currentPageIndexAtom)
    const paginationMap = useAtom(__manga_paginationMapAtom)

    const [hiddenBar, setHideBar] = useAtom(__manga_hiddenBarAtom)

    const kbsPageLeft = useAtomValue(__manga_kbsPageLeft)
    const kbsPageRight = useAtomValue(__manga_kbsPageRight)

    useHydrateMangaPaginationMap(pageContainer)

    const { handlePageLoad } = useHandleChapterPageStatus(pageContainer)

    /**
     * When the reader mounts (reading mode changes), scroll to the current page
     */
    useEffectOnce(() => {
        if (currentPageIndex !== 0) {
            const pageDiv = containerRef.current?.querySelector(`#page-${currentPageIndex}`)
            pageDiv?.scrollIntoView()
        }
    })


    /**
     * When there is a signal, scroll to the current page
     */
    const scrollSignal = useAtomValue(___manga_scrollSignalAtom)
    useUpdateEffect(() => {
        const pageDiv = containerRef.current?.querySelector(`#page-${currentPageIndex}`)
        pageDiv?.scrollIntoView()
    }, [scrollSignal])

    /**
     * Function to handle scroll event
     *
     * This function is responsible for handling the scroll event on the container div.
     * It checks if the user has scrolled past a certain point and sets the [isLastPage] state accordingly.
     * It also checks which page is currently in the viewport and sets the [currentPageIndex] state.
     */
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

    // Reset isLastPage state when pages change
    React.useEffect(() => {
        setIsLastPage(false)
    }, [pageContainer?.pages])

    // Add scroll event listener when component mounts
    React.useEffect(() => {
        // Add a scroll event listener to container
        containerRef.current?.addEventListener("scroll", handleScroll)
        return () => containerRef.current?.removeEventListener("scroll", handleScroll)
    }, [containerRef.current])

    // Page navigation
    React.useEffect(() => {
        mousetrap.bind("up", () => {
            containerRef.current?.scrollBy(0, -100)
        })
        mousetrap.bind("down", () => {
            containerRef.current?.scrollBy(0, 100)
        })


        return () => {
            mousetrap.unbind("up")
            mousetrap.unbind("down")
        }
    }, [paginationMap])

    /**
     * Key bindings for page navigation
     */
    React.useEffect(() => {
        mousetrap.bind(kbsPageLeft, () => {
            if (currentPageIndex > 0) {
                const pageDiv = containerRef.current?.querySelector(`#page-${currentPageIndex - 1}`)
                pageDiv?.scrollIntoView()
            }

        })
        mousetrap.bind(kbsPageRight, () => {
            if (pageContainer?.pages?.length && currentPageIndex < pageContainer?.pages?.length - 1) {
                const pageDiv = containerRef.current?.querySelector(`#page-${currentPageIndex + 1}`)
                pageDiv?.scrollIntoView()
            }
        })

        return () => {
            mousetrap.unbind(kbsPageLeft)
            mousetrap.unbind(kbsPageRight)
        }
    }, [kbsPageLeft, kbsPageRight, paginationMap])

    return (
        <div
            data-chapter-vertical-reader-container
            className={cn(
                "max-h-[calc(100dvh-3rem)] overflow-hidden relative focus-visible:outline-none",
                hiddenBar && "h-full max-h-full",
            )} tabIndex={-1}
            onClick={() => {
                if (!isMobile()) {
                    setHideBar(prev => !prev)
                }
            }}
        >
            <div
                data-chapter-vertical-reader-inner-container
                className={cn(
                    "w-full h-[calc(100dvh-3rem)] overflow-y-auto overflow-x-hidden px-4 select-none relative focus-visible:outline-none",
                    hiddenBar && "h-dvh",
                    pageGap && "space-y-4",
                )}
                ref={containerRef}
                tabIndex={-1}
            >
                <div
                    data-chapter-vertical-reader-inner-container-spacer
                    className="absolute w-full h-full z-[5] focus-visible:outline-none"
                    tabIndex={-1}
                >

                </div>
                {pageContainer?.pages?.map((page, index) => (
                    <ChapterPage
                        key={page.url}
                        page={page}
                        index={index}
                        readingMode={"paged"}
                        pageContainer={pageContainer}
                        onFinishedLoading={() => {
                            // If the first page is loaded, set the current page index to 0
                            // This is to avoid the current page index to remain incorrect when multiple pages are loading
                            if (index === 0) {
                                setCurrentPageIndex(0)
                            }
                            handlePageLoad(index)
                        }}
                        containerClass={cn(
                            "mx-auto scroll-div min-h-[200px] relative focus-visible:outline-none",
                            pageFit === MangaPageFit.CONTAIN && "max-w-full h-[calc(100dvh-60px)]",
                            pageFit === MangaPageFit.TRUE_SIZE && "max-w-full",
                            pageFit === MangaPageFit.COVER && "max-w-full",
                        )}
                        containerMaxWidth={pageFit === MangaPageFit.LARGER ? pageOverflowContainerWidth + "%" : undefined}
                        imageClass={cn(
                            "max-w-full h-auto mx-auto select-none z-[4] relative focus-visible:outline-none",

                            // "h-full inset-0 object-center select-none z-[4] relative",

                            pageFit === MangaPageFit.CONTAIN ?
                                pageStretch === MangaPageStretch.NONE ? "w-auto h-full object-center" : "object-fill w-full max-w-[1400px] h-full" :
                                undefined,
                            pageFit === MangaPageFit.LARGER ?
                                pageStretch === MangaPageStretch.NONE ? "w-auto h-full object-center" : "w-full h-auto object-cover mx-auto" :
                                undefined,
                            pageFit === MangaPageFit.COVER && "w-full h-auto",
                            pageFit === MangaPageFit.TRUE_SIZE && "object-none h-auto w-auto mx-auto",
                        )}

                        // imageMaxWidth={pageFit === MangaPageFit.LARGER ? pageOverflowContainerWidth+"%" : undefined}
                    />
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
