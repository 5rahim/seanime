import { useDiscordMangaPresence } from "@/app/(main)/manga/_lib/discord-manga-presence"
import { useMangaPageContainer } from "@/app/(main)/manga/_lib/manga.hooks"
import { MangaChapterContainer, MangaChapterDetails, MangaEntry } from "@/app/(main)/manga/_lib/manga.types"
import {
    __manga_isLastPageAtom,
    __manga_readingDirectionAtom,
    __manga_readingModeAtom,
    MangaHorizontalReader,
    MangaReadingDirection,
    MangaReadingMode,
    mangaReadingModeOptions,
    MangaVerticalReader,
    readingDirectionOptions,
} from "@/app/(main)/manga/entry/_containers/chapter-drawer/chapter-reader"
import { LuffyError } from "@/components/shared/luffy-error"
import { Button, IconButton } from "@/components/ui/button"
import { Card, CardFooter, CardHeader } from "@/components/ui/card"
import { cn } from "@/components/ui/core/styling"
import { Drawer } from "@/components/ui/drawer"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { RadioGroup } from "@/components/ui/radio-group"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { useQueryClient } from "@tanstack/react-query"
import { useAtom, useAtomValue } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React from "react"
import { AiOutlineArrowLeft, AiOutlineArrowRight } from "react-icons/ai"
import { BiCog } from "react-icons/bi"
import { MdMenuBook } from "react-icons/md"
import { PiArrowCircleLeftDuotone, PiArrowCircleRightDuotone, PiScrollDuotone } from "react-icons/pi"
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


export function ChapterDrawer(props: ChapterDrawerProps) {

    const {
        entry,
        chapterContainer,
        chapterIdToNumbersMap,
        ...rest
    } = props


    const qc = useQueryClient()

    // Discord rich presence
    useDiscordMangaPresence(entry)

    const [selectedChapter, setSelectedChapter] = useAtom(__manga_selectedChapterAtom)

    const [readingMode, setReadingMode] = useAtom(__manga_readingModeAtom)
    const readingDirection = useAtomValue(__manga_readingDirectionAtom)
    const isLastPage = useAtomValue(__manga_isLastPageAtom)

    const { pageContainer, pageContainerLoading, pageContainerError } = useMangaPageContainer(String(entry?.media?.id || "0"), selectedChapter?.id)

    // If the reading mode is set to double page but
    // the pageContainer doesn't have page dimensions, switch to paged mode
    React.useEffect(() => {
        if (selectedChapter) {
            if (readingMode === MangaReadingMode.DOUBLE_PAGE && !pageContainerLoading && !pageContainerError && !pageContainer?.pageDimensions) {
                toast.error("Could not efficiently get page dimensions from this provider. Switching to paged mode.")
                setReadingMode(MangaReadingMode.PAGED)
            }
        }
    }, [selectedChapter, pageContainer, pageContainerLoading, pageContainerError, readingMode])

    // Update the progress when the last page is reached
    const { mutate: updateProgress, isPending: isUpdatingProgress } = useSeaMutation<boolean, {
        chapterNumber: number,
        mediaId: number,
        totalChapters: number,
    }>({
        endpoint: SeaEndpoints.UPDATE_MANGA_PROGRESS,
        mutationKey: ["update-manga-progress", entry.mediaId],
        method: "post",
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["get-manga-entry", Number(entry.mediaId)] })
            await qc.refetchQueries({ queryKey: ["get-manga-collection"] })
            toast.success("Progress updated")
        },
    })

    // Get the previous and next chapter
    const { previousChapter, nextChapter } = React.useMemo(() => {
        if (!chapterContainer?.chapters) return { previousChapter: undefined, nextChapter: undefined }

        const idx = chapterContainer.chapters.findIndex((chapter) => chapter.id === selectedChapter?.id)
        return {
            previousChapter: chapterContainer.chapters[idx - 1],
            nextChapter: chapterContainer.chapters[idx + 1],
        }
    }, [chapterContainer?.chapters, selectedChapter])

    // Check if the progress should be updated
    const shouldUpdateProgress = React.useMemo(() => {
        const currentChapterNumber = chapterIdToNumbersMap.get(selectedChapter?.id || "")
        if (!currentChapterNumber) return false
        if (!entry.listData?.progress) return true
        return currentChapterNumber > entry.listData.progress
    }, [chapterIdToNumbersMap, entry, selectedChapter])

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

            <div className="fixed bottom-0 w-full h-12 gap-4 flex items-center px-4 z-[10] bg-[#0c0c0c]">

                <IconButton
                    icon={<AiOutlineArrowLeft />}
                    rounded
                    intent="white-outline"
                    size="xs"
                    onClick={() => setSelectedChapter(undefined)}
                />

                <h4 className="flex gap-1 items-center">
                    <span className="max-w-[180px] text-ellipsis truncate block">{entry?.media?.title?.userPreferred}</span> - {selectedChapter?.title || ""}
                </h4>
                <div className="flex flex-1"></div>

                <div className="flex gap-2 items-center">
                    {(readingDirection === MangaReadingDirection.RTL && (readingMode === MangaReadingMode.PAGED || readingMode === MangaReadingMode.DOUBLE_PAGE))
                        ? (
                            <>
                                {!!nextChapter && <Button
                                    leftIcon={<AiOutlineArrowLeft />}
                                    rounded
                                    intent="white-outline"
                                    size="sm"
                                    onClick={() => setSelectedChapter(nextChapter)}
                                >Next chapter</Button>}
                                {!!previousChapter && <Button
                                    rightIcon={<AiOutlineArrowRight />}
                                    rounded
                                    intent="gray-outline"
                                    size="sm"
                                    onClick={() => setSelectedChapter(previousChapter)}
                                >Previous chapter</Button>}
                            </>
                        )
                        : (
                            <>
                                {!!previousChapter && <Button
                                    leftIcon={<AiOutlineArrowLeft />}
                                    rounded
                                    intent="gray-outline"
                                    size="sm"
                                    onClick={() => setSelectedChapter(previousChapter)}
                                >Previous chapter</Button>}
                                {!!nextChapter && <Button
                                    rightIcon={<AiOutlineArrowRight />}
                                    rounded
                                    intent="white-outline"
                                    size="sm"
                                    onClick={() => setSelectedChapter(nextChapter)}
                                >Next chapter</Button>}
                            </>
                        )}
                </div>

                <div className="flex flex-1"></div>
                <p className="flex gap-4 items-center">
                    <span className="flex items-center gap-1">
                        {readingMode === MangaReadingMode.LONG_STRIP && (<PiScrollDuotone className="text-2xl" />)}
                        {readingMode === MangaReadingMode.PAGED && (<MdMenuBook className="text-2xl" />)}
                        {mangaReadingModeOptions.find((option) => option.value === readingMode)?.label}
                    </span>
                    {readingMode !== MangaReadingMode.LONG_STRIP && (
                        <span className="flex items-center gap-1">
                            {readingDirection === MangaReadingDirection.RTL && (<PiArrowCircleLeftDuotone className="text-2xl" />)}
                            <span>{readingDirectionOptions.find((option) => option.value === readingDirection)?.label}</span>
                            {readingDirection === MangaReadingDirection.LTR && (<PiArrowCircleRightDuotone className="text-2xl" />)}
                        </span>
                    )}
                </p>
                <ChapterReadingSettings />
            </div>


            <div className="max-h-[calc(100dvh-3rem)]" tabIndex={-1}>
                {pageContainerError ? (
                    <LuffyError
                        title="Failed to load pages"
                    >
                        <p>An error occurred while trying to load the pages for this chapter.</p>
                    </LuffyError>
                ) : (pageContainerLoading)
                    ? (<LoadingSpinner />)
                    : (readingMode === MangaReadingMode.LONG_STRIP
                        ? (<MangaVerticalReader pageContainer={pageContainer} />)
                        : (readingMode === MangaReadingMode.PAGED || readingMode === MangaReadingMode.DOUBLE_PAGE)
                            ? (<MangaHorizontalReader pageContainer={pageContainer} />) : null)}
            </div>
        </Drawer>
    )
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////


export type ChapterReadingSettingsProps = {
    children?: React.ReactNode
}

export function ChapterReadingSettings(props: ChapterReadingSettingsProps) {

    const {
        children,
        ...rest
    } = props

    const [readingDirection, setReadingDirection] = useAtom(__manga_readingDirectionAtom)
    const [readingMode, setReadingMode] = useAtom(__manga_readingModeAtom)

    return (
        <>
            <Drawer
                trigger={
                    <IconButton
                        icon={<BiCog />}
                        intent="gray-basic"
                        className=""
                    />
                }
                title="Settings"
                allowOutsideInteraction={true}
            >
                <div className="space-y-4 py-4">
                    <RadioGroup
                        itemClass={cn(
                            "border-transparent absolute top-2 right-2 bg-transparent dark:bg-transparent dark:data-[state=unchecked]:bg-transparent",
                            "data-[state=unchecked]:bg-transparent data-[state=unchecked]:hover:bg-transparent dark:data-[state=unchecked]:hover:bg-transparent",
                            "focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:ring-offset-transparent",
                        )}
                        itemIndicatorClass="hidden"
                        itemLabelClass="font-normal tracking-wide line-clamp-1 truncate flex flex-col items-center data-[state=checked]:text-[--brand] cursor-pointer"
                        itemContainerClass={cn(
                            "items-start cursor-pointer transition border-transparent rounded-[--radius] py-1.5 px-2 w-full",
                            "hover:bg-[--subtle] dark:bg-gray-900",
                            "data-[state=checked]:bg-white dark:data-[state=checked]:bg-gray-950",
                            "focus:ring-2 ring-transparent dark:ring-transparent outline-none ring-offset-1 ring-offset-[--background] focus-within:ring-2 transition",
                            "border border-transparent data-[state=checked]:border-[--brand] data-[state=checked]:ring-offset-0",
                        )}
                        label="Reading mode"
                        options={mangaReadingModeOptions}
                        value={readingMode}
                        onValueChange={(value) => setReadingMode(value as any)}
                    />

                    {readingMode !== MangaReadingMode.LONG_STRIP && (
                        <>
                            <RadioGroup
                                itemClass={cn(
                                    "border-transparent absolute top-2 right-2 bg-transparent dark:bg-transparent dark:data-[state=unchecked]:bg-transparent",
                                    "data-[state=unchecked]:bg-transparent data-[state=unchecked]:hover:bg-transparent dark:data-[state=unchecked]:hover:bg-transparent",
                                    "focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:ring-offset-transparent",
                                )}
                                itemIndicatorClass="hidden"
                                itemLabelClass="font-normal tracking-wide line-clamp-1 truncate flex flex-col items-center data-[state=checked]:text-[--brand] cursor-pointer"
                                itemContainerClass={cn(
                                    "items-start cursor-pointer transition border-transparent rounded-[--radius] py-1.5 px-2 w-full",
                                    "hover:bg-[--subtle] dark:bg-gray-900",
                                    "data-[state=checked]:bg-white dark:data-[state=checked]:bg-gray-950",
                                    "focus:ring-2 ring-transparent dark:ring-transparent outline-none ring-offset-1 ring-offset-[--background] focus-within:ring-2 transition",
                                    "border border-transparent data-[state=checked]:border-[--brand] data-[state=checked]:ring-offset-0",
                                )}
                                label="Reading direction"
                                options={readingDirectionOptions}
                                value={readingDirection}
                                onValueChange={(value) => setReadingDirection(value as any)}
                            />
                        </>
                    )}
                </div>
            </Drawer>
        </>
    )
}
