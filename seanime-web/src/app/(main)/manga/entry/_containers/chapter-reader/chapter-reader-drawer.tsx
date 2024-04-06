import { useDiscordMangaPresence } from "@/app/(main)/manga/_lib/discord-manga-presence"
import { useMangaPageContainer, useUpdateMangaProgress } from "@/app/(main)/manga/_lib/manga.hooks"
import { MangaChapterContainer, MangaChapterDetails, MangaEntry } from "@/app/(main)/manga/_lib/manga.types"
import { MangaHorizontalReader } from "@/app/(main)/manga/entry/_containers/chapter-reader/_components/chapter-horizontal-reader"
import { MangaVerticalReader } from "@/app/(main)/manga/entry/_containers/chapter-reader/_components/chapter-vertical-reader"
import { MangaReaderBar } from "@/app/(main)/manga/entry/_containers/chapter-reader/_components/manga-reader-bar"
import {
    __manga_currentPageIndexAtom,
    __manga_currentPaginationMapIndexAtom,
    __manga_isLastPageAtom,
    __manga_kbsChapterLeft,
    __manga_kbsChapterRight,
    __manga_paginationMapAtom,
    __manga_readingDirectionAtom,
    __manga_readingModeAtom,
    MangaReadingDirection,
    MangaReadingMode,
} from "@/app/(main)/manga/entry/_containers/chapter-reader/_lib/manga-chapter-reader.atoms"
import { useSwitchSettingsWithKeys } from "@/app/(main)/manga/entry/_containers/chapter-reader/_lib/manga-reader.hooks"
import { LuffyError } from "@/components/shared/luffy-error"
import { Button } from "@/components/ui/button"
import { Card, CardFooter, CardHeader } from "@/components/ui/card"
import { cn } from "@/components/ui/core/styling"
import { Drawer } from "@/components/ui/drawer"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import mousetrap from "mousetrap"
import React from "react"
import { toast } from "sonner"

type ChapterDrawerProps = {
    entry: MangaEntry
    chapterContainer: MangaChapterContainer
    chapterIdToNumbersMap: Map<string, number>
}

export const __manga_selectedChapterAtom = atomWithStorage<MangaChapterDetails | undefined>("sea-manga-chapter",
    undefined,
    undefined,
    { getOnInit: true })


export function ChapterReaderDrawer(props: ChapterDrawerProps) {

    const {
        entry,
        chapterContainer,
        chapterIdToNumbersMap,
        ...rest
    } = props

    // Discord rich presence
    useDiscordMangaPresence(entry)

    const [selectedChapter, setSelectedChapter] = useAtom(__manga_selectedChapterAtom)
    const setCurrentPageIndex = useSetAtom(__manga_currentPageIndexAtom)
    const setCurrentPaginationMapIndex = useSetAtom(__manga_currentPaginationMapIndexAtom)
    const [readingMode, setReadingMode] = useAtom(__manga_readingModeAtom)
    const isLastPage = useAtomValue(__manga_isLastPageAtom)
    const kbsChapterLeft = useAtomValue(__manga_kbsChapterLeft)
    const kbsChapterRight = useAtomValue(__manga_kbsChapterRight)
    const paginationMap = useAtomValue(__manga_paginationMapAtom)
    const readingDirection = useAtomValue(__manga_readingDirectionAtom)

    useSwitchSettingsWithKeys()

    const { pageContainer, pageContainerLoading, pageContainerError } = useMangaPageContainer(String(entry?.media?.id || "0"), selectedChapter?.id)

    /**
     * Switch back to PAGED mode if the page dimensions could not be fetched efficiently
     */
    React.useEffect(() => {
        if (selectedChapter) {
            if (readingMode === MangaReadingMode.DOUBLE_PAGE && !pageContainerLoading && !pageContainerError && (!pageContainer?.pageDimensions || Object.keys(
                pageContainer.pageDimensions).length === 0)) {
                toast.error("Could not get page dimensions from this provider. Switching to paged mode.")
                setReadingMode(MangaReadingMode.PAGED)
            }
        }
    }, [selectedChapter, pageContainer, pageContainerLoading, pageContainerError, readingMode])

    /**
     * Update the progress when the user confirms
     */
    const { updateProgress, isUpdatingProgress } = useUpdateMangaProgress(entry.mediaId)

    /**
     * Get the previous and next chapters
     * Either can be undefined
     */
    const { previousChapter, nextChapter } = React.useMemo(() => {
        if (!chapterContainer?.chapters) return { previousChapter: undefined, nextChapter: undefined }

        const idx = chapterContainer.chapters.findIndex((chapter) => chapter.id === selectedChapter?.id)
        return {
            previousChapter: chapterContainer.chapters[idx - 1],
            nextChapter: chapterContainer.chapters[idx + 1],
        }
    }, [chapterContainer?.chapters, selectedChapter])

    /**
     * Check if the progress should be updated
     * i.e. User progress is less than the current chapter number
     */
    const shouldUpdateProgress = React.useMemo(() => {
        const currentChapterNumber = chapterIdToNumbersMap.get(selectedChapter?.id || "")
        if (!currentChapterNumber) return false
        if (!entry.listData?.progress) return true
        return currentChapterNumber > entry.listData.progress
    }, [chapterIdToNumbersMap, entry, selectedChapter])

    /**
     * Reset the current page index when the pageContainer or chapterContainer changes
     * This signals that the user has switched chapters
     */
    const previousChapterId = React.useRef(selectedChapter?.id)
    React.useEffect(() => {
        // Avoid resetting the page index when we're still on the same chapter
        if (selectedChapter?.id !== previousChapterId.current) {
            setCurrentPageIndex(0)
            setCurrentPaginationMapIndex(0)
            previousChapterId.current = selectedChapter?.id
        }
    }, [pageContainer?.pages, chapterContainer?.chapters])

    // Navigation
    React.useEffect(() => {
        mousetrap.bind(kbsChapterLeft, () => {
            if (readingDirection === MangaReadingDirection.LTR) {
                setSelectedChapter(previousChapter)
            } else {
                setSelectedChapter(nextChapter)
            }
        })
        mousetrap.bind(kbsChapterRight, () => {
            if (readingDirection === MangaReadingDirection.RTL) {
                setSelectedChapter(previousChapter)
            } else {
                setSelectedChapter(nextChapter)
            }
        })

        return () => {
            mousetrap.unbind(kbsChapterLeft)
            mousetrap.unbind(kbsChapterRight)
        }
    }, [kbsChapterLeft, kbsChapterRight, paginationMap, readingDirection])

    return (
        <Drawer
            open={!!selectedChapter}
            onOpenChange={() => setSelectedChapter(undefined)}
            size="full"
            side="bottom"
            headerClass="absolute h-0"
            contentClass="p-0"
            closeButton={<></>}
        >

            <div
                className={cn(
                    "fixed top-2 left-2 z-[6] opacity-0 transition-opacity hidden duration-500",
                    (shouldUpdateProgress && isLastPage && !pageContainerLoading && !pageContainerError) && "block opacity-100",
                )}
                tabIndex={-1}
            >
                <Card className="max-w-[800px]">
                    <CardHeader>
                        Update progress to {chapterIdToNumbersMap.get(selectedChapter?.id || "")} / {entry?.media?.chapters || "-"}
                    </CardHeader>
                    <CardFooter>
                        <Button
                            onClick={() => {
                                updateProgress({
                                    chapterNumber: chapterIdToNumbersMap.get(selectedChapter?.id || "") || 0,
                                    mediaId: entry.mediaId,
                                    malId: entry.media?.idMal || undefined,
                                    totalChapters: chapterContainer?.chapters?.length || 0,
                                })
                                !!nextChapter && setSelectedChapter(nextChapter)
                            }}
                            className="w-full"
                            size="sm"
                            intent="success"
                            loading={isUpdatingProgress}
                            disabled={isUpdatingProgress}
                        >
                            Confirm
                        </Button>
                    </CardFooter>
                </Card>
            </div>

            <MangaReaderBar
                previousChapter={previousChapter}
                nextChapter={nextChapter}
                pageContainer={pageContainer}
                entry={entry}
            />


            <div className="max-h-[calc(100dvh-3rem)] h-full" tabIndex={-1}>
                {pageContainerError ? (
                    <LuffyError
                        title="Failed to load pages"
                    >
                        <p>An error occurred while trying to load pages for this chapter.</p>
                        <p>Reload the page, reload sources or change the source.</p>
                    </LuffyError>
                ) : (pageContainerLoading)
                    ? (<LoadingSpinner containerClass="h-full" />)
                    : (readingMode === MangaReadingMode.LONG_STRIP
                        ? (<MangaVerticalReader pageContainer={pageContainer} />)
                        : (readingMode === MangaReadingMode.PAGED || readingMode === MangaReadingMode.DOUBLE_PAGE)
                            ? (<MangaHorizontalReader pageContainer={pageContainer} />) : null)}
            </div>
        </Drawer>
    )
}
