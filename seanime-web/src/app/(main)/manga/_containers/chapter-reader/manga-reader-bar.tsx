import { AL_BaseManga, Manga_PageContainer } from "@/api/generated/types"
import { ___manga_scrollSignalAtom } from "@/app/(main)/manga/_containers/chapter-reader/_components/chapter-vertical-reader"
import {
    ChapterReaderSettings,
    MANGA_PAGE_FIT_OPTIONS,
    MANGA_PAGE_STRETCH_OPTIONS,
    MANGA_READING_DIRECTION_OPTIONS,
    MANGA_READING_MODE_OPTIONS,
} from "@/app/(main)/manga/_containers/chapter-reader/chapter-reader-settings"
import { __manga_selectedChapterAtom, MangaReader_SelectedChapter, useHandleChapterPageStatus } from "@/app/(main)/manga/_lib/handle-chapter-reader"
import {
    __manga_currentPageIndexAtom,
    __manga_currentPaginationMapIndexAtom,
    __manga_hiddenBarAtom,
    __manga_pageFitAtom,
    __manga_pageStretchAtom,
    __manga_paginationMapAtom,
    __manga_readerProgressBarAtom,
    __manga_readingDirectionAtom,
    __manga_readingModeAtom,
    MangaPageStretch,
    MangaReadingDirection,
    MangaReadingMode,
} from "@/app/(main)/manga/_lib/manga-chapter-reader.atoms"
import { Badge } from "@/components/ui/badge"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Popover } from "@/components/ui/popover"
import { Select } from "@/components/ui/select"
import { useSetAtom } from "jotai"
import { useAtom, useAtomValue } from "jotai/react"
import React from "react"
import { BiX } from "react-icons/bi"
import { LuChevronLeft, LuChevronRight, LuInfo } from "react-icons/lu"

type MangaReaderBarProps = {
    children?: React.ReactNode
    previousChapter: MangaReader_SelectedChapter | undefined
    nextChapter: MangaReader_SelectedChapter | undefined
    goToChapter: (dir: "previous" | "next") => void
    pageContainer: Manga_PageContainer | undefined
    entry: { mediaId: number, media?: AL_BaseManga } | undefined
}

export function MangaReaderBar(props: MangaReaderBarProps) {

    const {
        children,
        previousChapter,
        nextChapter,
        pageContainer,
        entry,
        goToChapter,
        ...rest
    } = props

    const [selectedChapter, setSelectedChapter] = useAtom(__manga_selectedChapterAtom)

    const [currentPageIndex, setCurrentPageIndex] = useAtom(__manga_currentPageIndexAtom)
    const paginationMap = useAtomValue(__manga_paginationMapAtom)
    const pageFit = useAtomValue(__manga_pageFitAtom)
    const pageStretch = useAtomValue(__manga_pageStretchAtom)
    const readingMode = useAtomValue(__manga_readingModeAtom)
    const readingDirection = useAtomValue(__manga_readingDirectionAtom)
    const readerProgressBar = useAtomValue(__manga_readerProgressBarAtom)

    const hiddenBar = useAtomValue(__manga_hiddenBarAtom)

    const ChapterNavButton = React.useCallback(({ dir }: { dir: "left" | "right" }) => {
        const reversed = (readingDirection === MangaReadingDirection.RTL && (readingMode === MangaReadingMode.PAGED || readingMode === MangaReadingMode.DOUBLE_PAGE))
        if (reversed) {
            if (dir === "left") {
                return (
                    <IconButton
                        icon={<LuChevronLeft />}
                        rounded
                        intent="white-outline"
                        size="sm"
                        onClick={() => {
                            if (nextChapter && entry) {
                                goToChapter("next")
                            }
                        }}
                        disabled={!nextChapter}
                        className="border-transparent"
                    />
                )
            } else {
                return (
                    <IconButton
                        icon={<LuChevronRight />}
                        rounded
                        intent="gray-outline"
                        size="sm"
                        onClick={() => {
                            if (previousChapter && entry) {
                                goToChapter("previous")
                            }
                        }}
                        disabled={!previousChapter}
                        className="border-transparent"
                    />
                )
            }
        } else {
            if (dir === "left") {
                return (
                    <IconButton
                        icon={<LuChevronLeft />}
                        rounded
                        intent="gray-outline"
                        size="sm"
                        onClick={() => {
                            if (previousChapter && entry) {
                                goToChapter("previous")
                            }
                        }}
                        disabled={!previousChapter}
                        className="border-transparent"
                    />
                )
            } else {
                return (
                    <IconButton
                        icon={<LuChevronRight />}
                        rounded
                        intent="white-outline"
                        size="sm"
                        onClick={() => {
                            if (nextChapter && entry) {
                                goToChapter("next")
                            }
                        }}
                        disabled={!nextChapter}
                        className="border-transparent"
                    />
                )
            }
        }
    }, [selectedChapter, nextChapter, previousChapter, readingDirection, readingMode])

    /**
     * Format the second part of pagination text
     * e.g. 1-2 / 10
     */
    const secondPageText = React.useMemo(() => {
        let secondPageIndex = 0
        for (const [key, values] of Object.entries(paginationMap)) {
            if (paginationMap[Number(key)].includes(currentPageIndex)) {
                secondPageIndex = values[1]
            }
        }
        if (isNaN(secondPageIndex) || secondPageIndex === 0 || secondPageIndex === currentPageIndex) return ""
        return "-" + (secondPageIndex + 1)
    }, [currentPageIndex, paginationMap])

    /**
     * Pagination
     */
    const [currentMapIndex, setCurrentMapIndex] = useAtom(__manga_currentPaginationMapIndexAtom)
    const setScrollSignal = useSetAtom(___manga_scrollSignalAtom)
    const handlePageChange = React.useCallback((pageIdx: number) => {
        if (readingMode === MangaReadingMode.PAGED) {
            setCurrentPageIndex(pageIdx)
            setCurrentMapIndex(pageIdx)
        } else if (readingMode === MangaReadingMode.DOUBLE_PAGE) {
            setCurrentMapIndex(prevMapIdx => {
                // Find the new map index based on the page index
                // e.g., { 0: [0, 1], 1: [2, 3], 2: [4, 5] }
                //   if pageIdx is 3, then the new map index is 1
                const newMapIdx = Object.keys(paginationMap).find(key => paginationMap[Number(key)].includes(pageIdx))
                if (newMapIdx === undefined) return prevMapIdx
                return Number(newMapIdx)
            })
            setCurrentPageIndex(pageIdx)
        } else {
            setCurrentPageIndex(pageIdx)
            setScrollSignal(p => p + 1)
        }
    }, [readingMode, paginationMap])

    const { allPagesLoaded } = useHandleChapterPageStatus(pageContainer)

    if (!entry) return null

    return (
        <>
            {(pageContainer && readerProgressBar && allPagesLoaded) && <div
                data-manga-reader-bar-container
                className={cn(
                    "bottom-12 w-full fixed z-10 hidden lg:block group/bp",
                    hiddenBar && "bottom-0",
                )}
            >
                <div data-manga-reader-bar-inner-container className="flex max-w-full items-center">
                    {pageContainer.pages?.map((_, index) => (
                        <div
                            key={index}
                            data-manga-reader-bar-pagination
                            className={cn(
                                "w-full h-6 cursor-pointer",
                                "transition-all duration-200 bg-gradient-to-t via-transparent from-transparent from-10% to-transparent hover:from-gray-800",
                                index === currentPageIndex && "from-gray-800",
                                index < currentPageIndex && "from-[--subtle] from-5%",
                            )}
                            onClick={() => handlePageChange(index)}
                        >
                            <p
                                className={cn(
                                    "w-full h-full flex items-center rounded-t-md justify-center text-transparent group-hover/bp:text-[--muted] transition",
                                    "hover:text-white hover:bg-gray-800",
                                    index === currentPageIndex && "text-white hover:text-white group-hover/bp:text-white",
                                )}
                            >{index + 1}</p>
                        </div>
                    ))}
                </div>
            </div>}


            <div
                data-manga-reader-bar
                className={cn(
                    "fixed bottom-0 w-full h-12 gap-4 flex items-center px-4 z-[10] bg-[var(--background)] transition-transform",
                    hiddenBar && "translate-y-60",
                )} id="manga-reader-bar"
            >

                <IconButton
                    icon={<BiX />}
                    rounded
                    intent="gray-outline"
                    size="xs"
                    onClick={() => setSelectedChapter(undefined)}
                />

                <h4 data-manga-reader-bar-title className="lg:flex gap-1 items-center hidden">
                    <span className="max-w-[180px] text-ellipsis truncate block">{entry?.media?.title?.userPreferred}</span>
                </h4>

                {!!selectedChapter &&
                    <div data-manga-reader-bar-chapter-nav-container className="flex gap-3 items-center flex-none whitespace-nowrap ">
                    <ChapterNavButton dir="left" />
                    <span className="hidden md:inline-block">Chapter </span>
                    {`${selectedChapter?.chapterNumber}`}
                    <ChapterNavButton dir="right" />
                </div>}

                <div data-manga-reader-bar-spacer className="flex flex-1"></div>

                <div data-manga-reader-bar-page-container className="flex items-center gap-2">

                    {pageContainer && <Popover
                        trigger={
                            <Badge
                                size="lg"
                                className="w-fit cursor-pointer rounded-[--radius-md] z-[5] flex bg-gray-950 items-center bottom-2 focus-visible:outline-none"
                                tabIndex={-1}
                                data-manga-reader-bar-page-container-badge
                            >
                                {!!(currentPageIndex + 1) && (
                                    <p className="">
                                        {currentPageIndex + 1}{secondPageText}
                                        <span className="text-[--muted]"> / {pageContainer?.pages?.length}</span>
                                    </p>
                                )}
                            </Badge>
                        }
                    >
                        <Select
                            data-manga-reader-bar-page-container-select
                            options={pageContainer.pages?.map((_, index) => ({ label: String(index + 1), value: String(index) })) ?? []}
                            value={String(currentPageIndex)}
                            onValueChange={e => {
                                handlePageChange(Number(e))
                            }}
                        />
                    </Popover>}

                    <div data-manga-reader-bar-info-container className="hidden lg:flex">
                        <Popover
                            modal={true}
                            trigger={
                                <IconButton
                                    icon={<LuInfo />}
                                    intent="gray-basic"
                                    className="opacity-50 outline-0"
                                    tabIndex={-1}
                                />
                            }
                            className="text-[--muted] space-y-1"
                        >
                            <div data-manga-reader-bar-info-container-provider className="hidden lg:block">
                                <p className="text-[--muted] text-sm">{selectedChapter?.provider}</p>
                            </div>
                            <div data-manga-reader-bar-info-container-mode className="flex items-center gap-1">
                                <span className="text-white w-6">m:</span>
                                {MANGA_READING_MODE_OPTIONS.find((option) => option.value === readingMode)?.label}
                            </div>
                            <div data-manga-reader-bar-info-container-fit className="flex items-center gap-1">
                                <span className="text-white w-6">f:</span>
                                {MANGA_PAGE_FIT_OPTIONS.find((option) => option.value === pageFit)?.label}
                            </div>
                            {pageStretch !== MangaPageStretch.NONE &&
                                <div data-manga-reader-bar-info-container-stretch className="flex items-center gap-1">
                                <span className="text-white w-6">s:</span>
                                {MANGA_PAGE_STRETCH_OPTIONS.find((option) => option.value === pageStretch)?.label}
                            </div>}
                            {readingMode !== MangaReadingMode.LONG_STRIP && (
                                <div data-manga-reader-bar-info-container-direction className="flex items-center gap-1">
                                    <span className="text-white w-6">d:</span>
                                    <span>{MANGA_READING_DIRECTION_OPTIONS.find((option) => option.value === readingDirection)?.label}</span>
                                </div>
                            )}
                        </Popover>
                    </div>


                    <ChapterReaderSettings mediaId={entry.mediaId} />

                </div>
            </div>
        </>
    )
}
