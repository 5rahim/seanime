import { AL_BaseManga, Manga_ChapterContainer, Manga_EntryListData } from "@/api/generated/types"
import { useGetMangaEntryPages, useUpdateMangaProgress } from "@/api/hooks/manga.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { MangaHorizontalReader } from "@/app/(main)/manga/_containers/chapter-reader/_components/chapter-horizontal-reader"
import { MangaVerticalReader } from "@/app/(main)/manga/_containers/chapter-reader/_components/chapter-vertical-reader"
import { MangaReaderBar } from "@/app/(main)/manga/_containers/chapter-reader/manga-reader-bar"
import {
    useCurrentChapter,
    useHandleChapterPagination,
    useSetCurrentChapter,
    useSwitchSettingsWithKeys,
} from "@/app/(main)/manga/_lib/handle-chapter-reader"
import { useDiscordMangaPresence } from "@/app/(main)/manga/_lib/handle-discord-manga-presence"
import {
    __manga_currentPageIndexAtom,
    __manga_currentPaginationMapIndexAtom,
    __manga_hiddenBarAtom,
    __manga_isLastPageAtom,
    __manga_kbsChapterLeft,
    __manga_kbsChapterRight,
    __manga_paginationMapAtom,
    __manga_readingDirectionAtom,
    __manga_readingModeAtom,
    MangaReadingDirection,
    MangaReadingMode,
} from "@/app/(main)/manga/_lib/manga-chapter-reader.atoms"
import { LuffyError } from "@/components/shared/luffy-error"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Drawer } from "@/components/ui/drawer"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import mousetrap from "mousetrap"
import React from "react"
import { TbLayoutBottombarExpandFilled } from "react-icons/tb"
import { toast } from "sonner"

type ChapterDrawerProps = {
    entry: { media?: AL_BaseManga | undefined, mediaId: number, listData?: Manga_EntryListData }
    chapterContainer: Manga_ChapterContainer
    chapterIdToNumbersMap: Map<string, number>
}


export function ChapterReaderDrawer(props: ChapterDrawerProps) {

    const {
        entry,
        chapterContainer,
        chapterIdToNumbersMap,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    // Discord rich presence
    useDiscordMangaPresence(entry)

    const currentChapter = useCurrentChapter()
    const setCurrentChapter = useSetCurrentChapter()

    const setCurrentPageIndex = useSetAtom(__manga_currentPageIndexAtom)
    const setCurrentPaginationMapIndex = useSetAtom(__manga_currentPaginationMapIndexAtom)

    const [readingMode, setReadingMode] = useAtom(__manga_readingModeAtom)
    const isLastPage = useAtomValue(__manga_isLastPageAtom)
    const kbsChapterLeft = useAtomValue(__manga_kbsChapterLeft)
    const kbsChapterRight = useAtomValue(__manga_kbsChapterRight)
    const paginationMap = useAtomValue(__manga_paginationMapAtom)
    const readingDirection = useAtomValue(__manga_readingDirectionAtom)

    const [hiddenBar, setHideBar] = useAtom(__manga_hiddenBarAtom)

    useSwitchSettingsWithKeys()

    /**
     * Get the pages
     */
    const {
        data: pageContainer,
        isLoading: pageContainerLoading,
        isError: pageContainerError,
        refetch: retryFetchPageContainer,
    } = useGetMangaEntryPages({
        mediaId: entry?.media?.id,
        chapterId: currentChapter?.chapterId,
        // provider: chapterContainer.provider as Manga_Provider,
        provider: currentChapter?.provider,
        doublePage: readingMode === MangaReadingMode.DOUBLE_PAGE,
    })

    /**
     * Update the progress when the user confirms
     */
    const { mutate: updateProgress, isPending: isUpdatingProgress } = useUpdateMangaProgress(entry.mediaId)

    /**
     * Switch back to PAGED mode if the page dimensions could not be fetched efficiently
     */
    React.useEffect(() => {
        if (currentChapter) {
            if (
                readingMode === MangaReadingMode.DOUBLE_PAGE &&
                !pageContainerLoading &&
                !pageContainerError &&
                (!pageContainer?.pageDimensions || Object.keys(pageContainer.pageDimensions).length === 0)
            ) {
                toast.error("Could not get page dimensions from this provider. Switching to paged mode.")
                setReadingMode(MangaReadingMode.PAGED)
            }
        }
    }, [currentChapter, pageContainer, pageContainerLoading, pageContainerError, readingMode])


    /**
     * Get the previous and next chapters
     * Either can be undefined
     */
    const { previousChapter, nextChapter, goToChapter } = useHandleChapterPagination(entry.mediaId, chapterContainer)

    /**
     * Check if the progress should be updated
     * i.e. User progress is less than the current chapter number
     */
    const shouldUpdateProgress = React.useMemo(() => {
        const currentChapterNumber = chapterIdToNumbersMap.get(currentChapter?.chapterId || "")
        if (!currentChapterNumber) return false
        if (!entry.listData?.progress) return true
        return currentChapterNumber > entry.listData.progress
    }, [chapterIdToNumbersMap, entry, currentChapter])

    const handleUpdateProgress = () => {
        if (shouldUpdateProgress && !isUpdatingProgress) {

            updateProgress({
                chapterNumber: chapterIdToNumbersMap.get(currentChapter?.chapterId || "") || 0,
                mediaId: entry.mediaId,
                malId: entry.media?.idMal || undefined,
                totalChapters: entry.media?.chapters || 0,
            }, {
                onSuccess: () => {
                    goToChapter("next")
                },
            })

        }
    }

    /**
     * Reset the current page index when the pageContainer or chapterContainer changes
     * This signals that the user has switched chapters
     */
    const previousChapterId = React.useRef(currentChapter?.chapterId)
    React.useEffect(() => {
        // Avoid resetting the page index when we're still on the same chapter
        if (currentChapter?.chapterId !== previousChapterId.current) {
            setCurrentPageIndex(0)
            setCurrentPaginationMapIndex(0)
            previousChapterId.current = currentChapter?.chapterId
        }
    }, [pageContainer?.pages, chapterContainer?.chapters])

    // Progress update keyboard shortcuts
    React.useEffect(() => {
        mousetrap.bind("u", () => {
            handleUpdateProgress()
        })

        return () => {
            mousetrap.unbind("u")
        }
    }, [pageContainer?.pages, chapterContainer?.chapters, shouldUpdateProgress, isLastPage])

    // Hide bar shortcut
    React.useEffect(() => {
        mousetrap.bind("b", () => {
            setHideBar((prev) => !prev)
        })

        return () => {
            mousetrap.unbind("b")
        }
    }, [])

    // Navigation
    React.useEffect(() => {
        mousetrap.bind(kbsChapterLeft, () => {
            if (readingDirection === MangaReadingDirection.LTR) {
                goToChapter("previous")
            } else {
                goToChapter("next")
            }
        })
        mousetrap.bind(kbsChapterRight, () => {
            if (readingDirection === MangaReadingDirection.RTL) {
                goToChapter("previous")
            } else {
                goToChapter("next")
            }
        })

        return () => {
            mousetrap.unbind(kbsChapterLeft)
            mousetrap.unbind(kbsChapterRight)
        }
    }, [kbsChapterLeft, kbsChapterRight, paginationMap, readingDirection, chapterContainer, previousChapter, nextChapter])


    return (
        <Drawer
            open={!!currentChapter}
            onOpenChange={() => setCurrentChapter(undefined)}
            size="full"
            side="bottom"
            headerClass="absolute h-0"
            contentClass={cn(
                "p-0 pt-0",
            )}
            closeButton={<></>}
        >

            <div
                className={cn(
                    "fixed left-0 w-full z-[6] opacity-0 transition-opacity hidden duration-500",
                    process.env.NEXT_PUBLIC_PLATFORM !== "desktop" && "top-0 justify-center",
                    process.env.NEXT_PUBLIC_PLATFORM === "desktop" && cn(
                        "bottom-12",
                        hiddenBar && "bottom-0 justify-left",
                    ),
                    (shouldUpdateProgress && isLastPage && !pageContainerLoading && !pageContainerError) && "flex opacity-100",
                )}
                tabIndex={-1}
            >
                <Button
                    onClick={handleUpdateProgress}
                    className={cn(
                        process.env.NEXT_PUBLIC_PLATFORM !== "desktop" && "rounded-tl-none rounded-tr-none",
                        process.env.NEXT_PUBLIC_PLATFORM === "desktop" && "rounded-bl-none rounded-br-none rounded-tl-none",
                    )}
                    size="md"
                    intent="success"
                    loading={isUpdatingProgress}
                    disabled={isUpdatingProgress}
                >
                    Update progress ({chapterIdToNumbersMap.get(currentChapter?.chapterId || "")} / {entry?.media?.chapters || "-"})
                </Button>
            </div>

            {/*Exit fullscreen button*/}
            {hiddenBar && <div className="fixed right-0 bottom-4 group/hiddenBarArea z-[10] px-4">
                <IconButton
                    rounded
                    icon={<TbLayoutBottombarExpandFilled />}
                    intent="white-outline"
                    size="sm"
                    onClick={() => setHideBar(false)}
                    className="lg:opacity-0 opacity-30 group-hover/hiddenBarArea:opacity-100 transition-opacity duration-200"
                />
            </div>}

            <MangaReaderBar
                previousChapter={previousChapter}
                nextChapter={nextChapter}
                goToChapter={goToChapter}
                pageContainer={pageContainer}
                entry={entry}
            />


            <div
                className={cn(
                    "max-h-[calc(100dvh-3rem)] h-full",
                    hiddenBar && "max-h-dvh",
                )} tabIndex={-1}
            >
                {pageContainerError ? (
                    <LuffyError
                        title="Failed to load pages"
                    >
                        <p>An error occurred while trying to load pages for this chapter.</p>
                        <p>Reload the page, reload sources or change the source.</p>

                        <div className="mt-2">
                            <Button intent="white" onClick={() => retryFetchPageContainer()}>
                                Retry
                            </Button>
                        </div>
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
