import { MangaChapterDetails, MangaEntry, MangaPageContainer } from "@/app/(main)/manga/_lib/manga.types"
import {
    ChapterReaderSettings,
    MANGA_PAGE_FIT_OPTIONS,
    MANGA_PAGE_STRETCH_OPTIONS,
    MANGA_READING_DIRECTION_OPTIONS,
    MANGA_READING_MODE_OPTIONS,
} from "@/app/(main)/manga/entry/_containers/chapter-reader/_components/chapter-reader-settings"
import {
    __manga_currentPageIndexAtom,
    __manga_pageFitAtom,
    __manga_pageStretchAtom,
    __manga_paginationMapAtom,
    __manga_readingDirectionAtom,
    __manga_readingModeAtom,
    MangaPageStretch,
    MangaReadingDirection,
    MangaReadingMode,
} from "@/app/(main)/manga/entry/_containers/chapter-reader/_lib/manga-chapter-reader.atoms"
import { __manga_selectedChapterAtom } from "@/app/(main)/manga/entry/_containers/chapter-reader/chapter-reader-drawer"
import { Badge } from "@/components/ui/badge"
import { IconButton } from "@/components/ui/button"
import { useAtom, useAtomValue } from "jotai/react"
import React from "react"
import { AiOutlineArrowLeft, AiOutlineArrowRight, AiOutlineCloseCircle } from "react-icons/ai"

type MangaReaderBarProps = {
    children?: React.ReactNode
    previousChapter: MangaChapterDetails | undefined
    nextChapter: MangaChapterDetails | undefined
    pageContainer: MangaPageContainer | undefined
    entry: MangaEntry | undefined
}

export function MangaReaderBar(props: MangaReaderBarProps) {

    const {
        children,
        previousChapter,
        nextChapter,
        pageContainer,
        entry,
        ...rest
    } = props

    const [selectedChapter, setSelectedChapter] = useAtom(__manga_selectedChapterAtom)

    const [currentPageIndex, setCurrentPageIndex] = useAtom(__manga_currentPageIndexAtom)
    const paginationMap = useAtomValue(__manga_paginationMapAtom)
    const pageFit = useAtomValue(__manga_pageFitAtom)
    const pageStretch = useAtomValue(__manga_pageStretchAtom)
    const readingMode = useAtomValue(__manga_readingModeAtom)
    const readingDirection = useAtomValue(__manga_readingDirectionAtom)

    const ChapterNavButton = React.useCallback(({ dir }: { dir: "left" | "right" }) => {
        const reversed = (readingDirection === MangaReadingDirection.RTL && (readingMode === MangaReadingMode.PAGED || readingMode === MangaReadingMode.DOUBLE_PAGE))
        if (reversed) {
            if (dir === "left") {
                return (
                    <IconButton
                        icon={<AiOutlineArrowLeft />}
                        rounded
                        intent="white-outline"
                        size="sm"
                        onClick={() => {
                            if (nextChapter && entry) {
                                setSelectedChapter({
                                    chapterId: nextChapter.id,
                                    chapterNumber: nextChapter.chapter,
                                    provider: nextChapter.provider,
                                    mediaId: entry.mediaId,
                                })
                            }
                        }}
                        disabled={!nextChapter}
                    />
                )
            } else {
                return (
                    <IconButton
                        icon={<AiOutlineArrowRight />}
                        rounded
                        intent="gray-outline"
                        size="sm"
                        onClick={() => {
                            if (previousChapter && entry) {
                                setSelectedChapter({
                                    chapterId: previousChapter.id,
                                    chapterNumber: previousChapter.chapter,
                                    provider: previousChapter.provider,
                                    mediaId: entry.mediaId,
                                })
                            }
                        }}
                        disabled={!previousChapter}
                    />
                )
            }
        } else {
            if (dir === "left") {
                return (
                    <IconButton
                        icon={<AiOutlineArrowLeft />}
                        rounded
                        intent="gray-outline"
                        size="sm"
                        onClick={() => {
                            if (previousChapter && entry) {
                                setSelectedChapter({
                                    chapterId: previousChapter.id,
                                    chapterNumber: previousChapter.chapter,
                                    provider: previousChapter.provider,
                                    mediaId: entry.mediaId,
                                })
                            }
                        }}
                        disabled={!previousChapter}
                    />
                )
            } else {
                return (
                    <IconButton
                        icon={<AiOutlineArrowRight />}
                        rounded
                        intent="white-outline"
                        size="sm"
                        onClick={() => {
                            if (nextChapter && entry) {
                                setSelectedChapter({
                                    chapterId: nextChapter.id,
                                    chapterNumber: nextChapter.chapter,
                                    provider: nextChapter.provider,
                                    mediaId: entry.mediaId,
                                })
                            }
                        }}
                        disabled={!nextChapter}
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

    if (!entry) return null

    return (
        <>
            <div className="fixed bottom-0 w-full h-12 gap-4 flex items-center px-4 z-[10] bg-[#0c0c0c]" id="manga-reader-bar">

                <IconButton
                    icon={<AiOutlineCloseCircle />}
                    rounded
                    intent="white-outline"
                    size="xs"
                    onClick={() => setSelectedChapter(undefined)}
                />

                <h4 className="lg:flex gap-1 items-center hidden">
                    <span className="max-w-[180px] text-ellipsis truncate block">{entry?.media?.title?.userPreferred}</span>
                </h4>

                {!!selectedChapter && <div className="flex gap-3 items-center flex-none whitespace-nowrap ">
                    <ChapterNavButton dir="left" />
                    {`Chapter ${selectedChapter?.chapterNumber}`}
                    <ChapterNavButton dir="right" />
                </div>}

                <div className="flex flex-1"></div>

                {pageContainer && <Badge
                    size="lg"
                    className="w-fit rounded-md z-[5] flex bg-gray-950 items-center bottom-2 focus-visible:outline-none"
                    tabIndex={-1}
                >
                    {!!(currentPageIndex + 1) && (
                        <p className="">
                            {currentPageIndex + 1}{secondPageText}<span className="text-[--muted]"> / {pageContainer?.pages?.length}</span>
                        </p>
                    )}
                </Badge>}

                {/*{pageContainer && <Popover trigger={*/}
                {/*    <Badge size="lg" className="w-fit cursor-pointer rounded-md z-[5] flex bg-gray-950 items-center bottom-2 focus-visible:outline-none" tabIndex={-1}>*/}
                {/*        {!!(currentPageIndex + 1) && (*/}
                {/*            <p className="">*/}
                {/*                {currentPageIndex + 1}{secondPageText} <span className="text-[--muted]">/ {pageContainer?.pages?.length}</span>*/}
                {/*            </p>*/}
                {/*        )}*/}
                {/*    </Badge>*/}
                {/*}>*/}
                {/*    <Select*/}
                {/*        options={pageContainer.pages?.map((_, index) => ({ label: String(index + 1), value: String(index) })) ?? []}*/}
                {/*        value={String(currentPageIndex)}*/}
                {/*        onValueChange={value => {*/}
                {/*            setCurrentPageIndex(Number(value))*/}
                {/*        }}*/}
                {/*    />*/}
                {/*</Popover>}*/}

                <p className="hidden lg:flex gap-4 items-center text-[--muted]">
                    <span className="flex items-center gap-1">
                        <span className="text-white">m:</span>
                        {MANGA_READING_MODE_OPTIONS.find((option) => option.value === readingMode)?.label}
                    </span>
                    <span className="flex items-center gap-1">
                        <span className="text-white">f:</span>
                        {MANGA_PAGE_FIT_OPTIONS.find((option) => option.value === pageFit)?.label}
                    </span>
                    {pageStretch !== MangaPageStretch.NONE && <span className="flex items-center gap-1">
                        <span className="text-white">s:</span>
                        {MANGA_PAGE_STRETCH_OPTIONS.find((option) => option.value === pageStretch)?.label}
                    </span>}
                    {readingMode !== MangaReadingMode.LONG_STRIP && (
                        <span className="flex items-center gap-1">
                            <span className="text-white">d:</span>
                            <span>{MANGA_READING_DIRECTION_OPTIONS.find((option) => option.value === readingDirection)?.label}</span>
                        </span>
                    )}
                </p>

                <ChapterReaderSettings />
            </div>
        </>
    )
}
