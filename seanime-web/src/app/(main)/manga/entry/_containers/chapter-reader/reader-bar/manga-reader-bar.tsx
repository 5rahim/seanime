import { MangaChapterDetails, MangaEntry } from "@/app/(main)/manga/_lib/manga.types"
import {
    ChapterReaderSettings,
    MANGA_READING_DIRECTION_OPTIONS,
    MANGA_READING_MODE_OPTIONS,
} from "@/app/(main)/manga/entry/_containers/chapter-reader/_components/chapter-reader-settings"
import {
    __manga_isLastPageAtom,
    __manga_readingDirectionAtom,
    __manga_readingModeAtom,
    MangaReadingDirection,
    MangaReadingMode,
} from "@/app/(main)/manga/entry/_containers/chapter-reader/_lib/manga.atoms"
import { __manga_selectedChapterAtom } from "@/app/(main)/manga/entry/_containers/chapter-reader/chapter-reader-drawer"
import { IconButton } from "@/components/ui/button"
import { useAtom, useAtomValue } from "jotai/react"
import React from "react"
import { AiOutlineArrowLeft, AiOutlineArrowRight, AiOutlineCloseCircle } from "react-icons/ai"

type MangaReaderBarProps = {
    children?: React.ReactNode
    previousChapter?: MangaChapterDetails
    nextChapter?: MangaChapterDetails
    entry?: MangaEntry
}

export function MangaReaderBar(props: MangaReaderBarProps) {

    const {
        children,
        previousChapter,
        nextChapter,
        entry,
        ...rest
    } = props

    const [selectedChapter, setSelectedChapter] = useAtom(__manga_selectedChapterAtom)

    const [readingMode, setReadingMode] = useAtom(__manga_readingModeAtom)
    const readingDirection = useAtomValue(__manga_readingDirectionAtom)
    const isLastPage = useAtomValue(__manga_isLastPageAtom)

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
                        onClick={() => setSelectedChapter(nextChapter)}
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
                        onClick={() => setSelectedChapter(previousChapter)}
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
                        onClick={() => setSelectedChapter(previousChapter)}
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
                        onClick={() => setSelectedChapter(nextChapter)}
                        disabled={!nextChapter}
                    />
                )
            }
        }
    }, [selectedChapter, nextChapter, previousChapter, readingDirection, readingMode])

    if (!entry) return null

    return (
        <>
            <div className="fixed bottom-0 w-full h-12 gap-4 flex items-center px-4 z-[10] bg-[#0c0c0c]">

                <IconButton
                    icon={<AiOutlineCloseCircle />}
                    rounded
                    intent="white-outline"
                    size="xs"
                    onClick={() => setSelectedChapter(undefined)}
                />

                <h4 className="flex gap-1 items-center">
                    <span className="max-w-[180px] text-ellipsis truncate block">{entry?.media?.title?.userPreferred}</span>
                </h4>

                <div className="flex gap-3 items-center">
                    <ChapterNavButton dir="left" />
                    {selectedChapter?.title || ""}
                    <ChapterNavButton dir="right" />
                </div>

                <div className="flex flex-1"></div>


                <div className="flex flex-1"></div>

                <p className="flex gap-4 items-center opacity-50">
                    <span className="flex items-center gap-1">
                        {MANGA_READING_MODE_OPTIONS.find((option) => option.value === readingMode)?.label}
                    </span>
                    {readingMode !== MangaReadingMode.LONG_STRIP && (
                        <span className="flex items-center gap-1">
                            <span>{MANGA_READING_DIRECTION_OPTIONS.find((option) => option.value === readingDirection)?.label}</span>
                        </span>
                    )}
                </p>

                <ChapterReaderSettings />
            </div>
        </>
    )
}
