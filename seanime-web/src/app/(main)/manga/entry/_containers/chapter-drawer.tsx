import { useMangaPageContainer } from "@/app/(main)/manga/_lib/queries"
import { MangaChapterContainer, MangaChapterDetails, MangaEntry, MangaPageContainer } from "@/app/(main)/manga/_lib/types"
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
import { atom } from "jotai"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React from "react"
import { AiOutlineArrowLeft, AiOutlineArrowRight } from "react-icons/ai"
import { BiCog } from "react-icons/bi"
import { useMount } from "react-use"
import { toast } from "sonner"

type ChapterDrawerProps = {
    entry: MangaEntry
    chapterContainer: MangaChapterContainer
    chapterIdToNumbersMap: Map<string, number>
}

// export const __manga_selectedChapterAtom = atom<MangaChapterDetails | undefined>(undefined)
export const __manga_selectedChapterAtom = atomWithStorage<MangaChapterDetails | undefined>("sea-manga-chapter",
    undefined,
    undefined,
    { getOnInit: true })


const isLastPageAtom = atom(false)

export function ChapterDrawer(props: ChapterDrawerProps) {

    const {
        entry,
        chapterContainer,
        chapterIdToNumbersMap,
        ...rest
    } = props

    const qc = useQueryClient()

    const [selectedChapter, setSelectedChapter] = useAtom(__manga_selectedChapterAtom)

    const [readingMode, setReadingMode] = useAtom(readingModeAtom)
    const readingDirection = useAtomValue(readingDirectionAtom)
    const isLastPage = useAtomValue(isLastPageAtom)

    const { pageContainer, pageContainerLoading, pageContainerError } = useMangaPageContainer(String(entry?.media?.id || "0"), selectedChapter?.id)

    // If the reading mode is set to double page but
    // the pageContainer doesn't have page dimensions, switch to paged mode
    React.useEffect(() => {
        if (selectedChapter) {
            if (readingMode === ReadingMode.DOUBLE_PAGE && !pageContainerLoading && !pageContainerError && !pageContainer?.pageDimensions) {
                toast.error("Could not efficiently get page dimensions from this provider. Switching to paged mode.")
                setReadingMode(ReadingMode.PAGED)
            }
        }
    }, [selectedChapter, pageContainer, pageContainerLoading, pageContainerError, readingMode])

    const { mutate: updateProgress, isPending: isUpdatingProgress } = useSeaMutation<boolean, {
        chapterNumber: number,
        mediaId: number,
        totalChapters: number,
    }>({
        endpoint: SeaEndpoints.UPDATE_MANGA_PROGRESS,
        mutationKey: ["update-manga-progress", entry.mediaId],
        method: "post",
        onSuccess: () => {
            qc.refetchQueries({ queryKey: ["get-manga-entry", Number(entry.mediaId)] })
            qc.refetchQueries({ queryKey: ["get-manga-collection"] })
            toast.success("Progress updated")
        },
    })

    const { previousChapter, nextChapter } = React.useMemo(() => {
        if (!chapterContainer?.chapters) return { previousChapter: undefined, nextChapter: undefined }

        const idx = chapterContainer.chapters.findIndex((chapter) => chapter.id === selectedChapter?.id)
        return {
            previousChapter: chapterContainer.chapters[idx - 1],
            nextChapter: chapterContainer.chapters[idx + 1],
        }
    }, [chapterContainer?.chapters, selectedChapter])

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
                    "fixed top-2 left-2 z-[6] hidden",
                    (shouldUpdateProgress && isLastPage && !pageContainerLoading && !pageContainerError) && "block",
                )}
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
                            }}
                            className="w-full animate-pulse"
                            size="sm"
                            intent="success"
                            loading={isUpdatingProgress}
                            disabled={isUpdatingProgress}
                        >
                            Update progress
                        </Button>
                    </CardFooter>
                </Card>
            </div>

            <div className="fixed bottom-0 w-full h-12 gap-4 flex items-center px-4 z-[10] bg-[#0c0c0c]">
                <IconButton icon={<AiOutlineArrowLeft />} rounded intent="white-outline" size="xs" onClick={() => setSelectedChapter(undefined)} />
                <h4 className="flex gap-1 items-center">
                    <span className="max-w-[180px] text-ellipsis truncate block">{entry?.media?.title?.userPreferred}</span> - {selectedChapter?.title || ""}
                </h4>
                <div className="flex flex-1"></div>

                <div className="flex gap-2 items-center">
                    {(readingDirection === ReadingDirection.RTL && (readingMode === ReadingMode.PAGED || readingMode === ReadingMode.DOUBLE_PAGE)) ? (
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
                    ) : (
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
                <p>
                    {readingModeOptions.find((option) => option.value === readingMode)?.label}, {readingDirectionOptions.find((option) => option.value === readingDirection)?.label}
                </p>
                <ChapterReadingSettings />
            </div>


            <div className="max-h-[calc(100dvh-8rem)]">
                {pageContainerError ? (
                    <LuffyError
                        title="Failed to load pages"
                    >
                        <p>An error occurred while trying to load the pages for this chapter.</p>
                    </LuffyError>
                ) : (pageContainerLoading)
                    ? <LoadingSpinner /> :
                    (
                        readingMode === ReadingMode.LONG_STRIP ? (
                                <VerticalReadingMode pageContainer={pageContainer} />
                            )
                            :
                            readingMode === ReadingMode.PAGED || readingMode === ReadingMode.DOUBLE_PAGE ? (
                                    <HorizontalReadingMode pageContainer={pageContainer} />
                                ) :
                                null
                    )}
            </div>


        </Drawer>
    )
}

type HorizontalReadingModeProps = {
    pageContainer: MangaPageContainer | undefined
}

function HorizontalReadingMode({ pageContainer }: HorizontalReadingModeProps) {

    const containerRef = React.useRef<HTMLDivElement>(null)
    const readingMode = useAtomValue(readingModeAtom)
    const [selectedChapter, setSelectedChapter] = useAtom(__manga_selectedChapterAtom)
    const setIsLastPage = useSetAtom(isLastPageAtom)

    const readingDirection = useAtomValue(readingDirectionAtom)

    const [currentMapIndex, setCurrentMapIndex] = React.useState<number>(0)

    const [hydrated, setHydrated] = React.useState(false)
    useMount(() => {
        setHydrated(true)
    })

    const paginationMap = React.useMemo(() => {
        setCurrentMapIndex(0)
        if (!pageContainer?.pages?.length) return new Map<number, number[]>()

        if (readingMode === ReadingMode.PAGED || !pageContainer.pageDimensions) {
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
        // e.g. [[0, 1], [2, 3], [4, 5], [6], [7, 8], ...]
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
    }, [pageContainer?.pages, readingMode, selectedChapter, hydrated])

    const onPaginate = React.useCallback((dir: "left" | "right") => {
        const shouldDecrement = dir === "left" && readingDirection === ReadingDirection.LTR || dir === "right" && readingDirection === ReadingDirection.RTL

        setCurrentMapIndex((draft) => {
            const newIdx = shouldDecrement ? draft - 1 : draft + 1
            if (paginationMap.has(newIdx)) {
                return newIdx
            }
            return draft
        })
    }, [paginationMap, readingDirection])

    React.useEffect(() => {
        setIsLastPage(paginationMap.size > 0 && currentMapIndex === paginationMap.size - 1)
    }, [currentMapIndex, paginationMap])

    const getSrc = (url: string) => {
        if (!pageContainer?.isDownloaded) {
            return url
        }

        return process.env.NODE_ENV === "development"
            ? `http://${window?.location?.hostname}:43211/manga-backups${url}`
            : `http://${window?.location?.host}/manga-backups${url}`
    }

    const currentPages = React.useMemo(() => paginationMap.get(currentMapIndex), [currentMapIndex, paginationMap])
    const twoDisplayed = readingMode === ReadingMode.DOUBLE_PAGE && currentPages?.length === 2

    return (
        <div
            className="h-[calc(100dvh-60px)] overflow-y-hidden overflow-x-hidden w-full px-4 space-y-4 select-none relative"
            ref={containerRef}
        >
            <div className="w-fit right-6 absolute z-[5] flex items-center bottom-2">
                {!!currentPages?.length && (
                    readingMode === ReadingMode.DOUBLE_PAGE ? (
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
            <div className="absolute w-full h-[calc(100dvh-60px)] flex z-[5] cursor-pointer">
                <div className="h-full w-full flex flex-1" onClick={() => onPaginate("left")} />
                <div className="h-full w-full flex flex-1" onClick={() => onPaginate("right")} />
            </div>
            <div
                className={cn(
                    twoDisplayed && readingMode === ReadingMode.DOUBLE_PAGE && "flex space-x-2 transition-transform duration-300",
                    // twoDisplayed && readingMode === ReadingMode.DOUBLE_PAGE && readingDirection === ReadingDirection.RTL && "flex-row-reverse",
                )}
            >
                {pageContainer?.pages?.toSorted((a, b) => a.index - b.index)?.map((page, index) => (
                    <div
                        key={page.url}
                        className={cn(
                            "w-full h-[calc(100dvh-60px)] scroll-div min-h-[200px] relative page",
                            !currentPages?.includes(index) ? "hidden" : "displayed",
                            twoDisplayed && readingDirection === ReadingDirection.RTL && readingMode === ReadingMode.DOUBLE_PAGE && currentPages?.[0] === index && "before:content-[''] before:absolute before:w-[3%] before:z-[5] before:h-full before:[background:_linear-gradient(-90deg,_rgba(17,_17,_17,_0)_0,_rgba(17,_17,_17,_.3)_100%)]",
                            twoDisplayed && readingDirection === ReadingDirection.RTL && readingMode === ReadingMode.DOUBLE_PAGE && currentPages?.[1] === index && "before:content-[''] before:absolute before:right-0 before:w-[3%] before:z-[5] before:h-full before:[background:_linear-gradient(90deg,_rgba(17,_17,_17,_0)_0,_rgba(17,_17,_17,_.3)_100%)]",
                            twoDisplayed && readingDirection === ReadingDirection.LTR && readingMode === ReadingMode.DOUBLE_PAGE && currentPages?.[1] === index && "before:content-[''] before:absolute before:w-[3%] before:z-[5] before:h-full before:[background:_linear-gradient(-90deg,_rgba(17,_17,_17,_0)_0,_rgba(17,_17,_17,_.3)_100%)]",
                            twoDisplayed && readingDirection === ReadingDirection.LTR && readingMode === ReadingMode.DOUBLE_PAGE && currentPages?.[0] === index && "before:content-[''] before:absolute before:right-0 before:w-[3%] before:z-[5] before:h-full before:[background:_linear-gradient(90deg,_rgba(17,_17,_17,_0)_0,_rgba(17,_17,_17,_.3)_100%)]",
                        )}
                        id={`page-${index}`}
                    >
                        {/*<LoadingSpinner containerClass="h-full absolute inset-0 z-[1] w-24 mx-auto" />*/}
                        <img
                            src={getSrc(page.url)} alt={`Page ${index}`} className={cn(
                            "w-full h-full inset-0 object-contain object-center select-none z-[4] relative",
                            twoDisplayed && readingDirection === ReadingDirection.RTL && readingMode === ReadingMode.DOUBLE_PAGE && currentPages?.[0] === index && "[object-position:0%_50%] before:content-['']",
                            twoDisplayed && readingDirection === ReadingDirection.RTL && readingMode === ReadingMode.DOUBLE_PAGE && currentPages?.[1] === index && "[object-position:100%_50%]",
                            twoDisplayed && readingDirection === ReadingDirection.LTR && readingMode === ReadingMode.DOUBLE_PAGE && currentPages?.[0] === index && "[object-position:100%_50%]",
                            twoDisplayed && readingDirection === ReadingDirection.LTR && readingMode === ReadingMode.DOUBLE_PAGE && currentPages?.[1] === index && "[object-position:0%_50%]",
                        )}
                        />
                    </div>
                ))}
            </div>

        </div>
    )
}

type VerticalReadingModeProps = {
    pageContainer: MangaPageContainer | undefined
}

// source: https://stackoverflow.com/a/51121566
const isElementXPercentInViewport = function (el: any, percentVisible = 50) {
    let
        rect = el.getBoundingClientRect(),
        windowHeight = (window.innerHeight || document.documentElement.clientHeight)

    return !(
        Math.floor(100 - (((rect.top >= 0 ? 0 : rect.top) / +-rect.height) * 100)) < percentVisible ||
        Math.floor(100 - ((rect.bottom - windowHeight) / rect.height) * 100) < percentVisible
    )
}


function VerticalReadingMode({ pageContainer }: VerticalReadingModeProps) {

    const containerRef = React.useRef<HTMLDivElement>(null)
    const setIsLastPage = useSetAtom(isLastPageAtom)

    // const divs = React.useMemo(() => containerRef.current?.querySelectorAll(".scroll-div"), [containerRef.current])
    // Function to handle scroll event
    const handleScroll = () => {
        if (!!containerRef.current) {
            const scrollTop = containerRef.current.scrollTop
            const scrollHeight = containerRef.current.scrollHeight
            const clientHeight = containerRef.current.clientHeight

            if (scrollTop > 1000 && (scrollTop + clientHeight >= scrollHeight - 1500)) {
                setIsLastPage(true)
            } else {
                setIsLastPage(false)
            }

            // divs?.forEach((div) => {
            //     if (isElementXPercentInViewport(div) && pageContainer?.pages?.length) {
            //         const idx = Number(div.id.split("-")[1])
            //         if (idx === pageContainer?.pages?.length - 1) {
            //             setIsLastPage(true)
            //         } else {
            //             setIsLastPage(false)
            //         }
            //     }
            // })
        }
    }

    // Add scroll event listener when component mounts
    React.useEffect(() => {
        containerRef.current?.addEventListener("scroll", handleScroll)
        return () => containerRef.current?.removeEventListener("scroll", handleScroll)
    }, [containerRef.current])


    return (
        <div
            className="container h-[calc(100dvh-60px)] overflow-y-auto overflow-x-hidden max-w-[1400px] px-4 space-y-4 select-none relative"
            ref={containerRef}
        >
            <div className="absolute w-full h-full z-[5]">

            </div>
            {pageContainer?.pages?.map((page, index) => (
                <div key={page.url} className="w-full scroll-div min-h-[200px] relative" id={`page-${index}`}>
                    <LoadingSpinner containerClass="h-full absolute inset-0 z-[1]" />
                    <img src={page.url} alt={`Page ${index}`} className="max-w-full h-auto mx-auto select-none z-[4] relative" />
                </div>
            ))}

        </div>
    )
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const enum ReadingDirection {
    LTR = "ltr",
    RTL = "rtl",
}

const readingDirectionOptions = [
    { value: ReadingDirection.LTR, label: "Left to right" },
    { value: ReadingDirection.RTL, label: "Right to left" },
]

const enum ReadingMode {
    LONG_STRIP = "long-strip",
    PAGED = "paged",
    DOUBLE_PAGE = "double-page",
}

const readingModeOptions = [
    { value: ReadingMode.LONG_STRIP, label: "Long strip" },
    { value: ReadingMode.PAGED, label: "Paged" },
    { value: ReadingMode.DOUBLE_PAGE, label: "Double page" },
]

const readingDirectionAtom = atomWithStorage<ReadingDirection>("sea-manga-reading-direction", ReadingDirection.LTR)
const readingModeAtom = atomWithStorage<ReadingMode>("sea-manga-reading-mode", ReadingMode.LONG_STRIP)


type ChapterReadingSettingsProps = {
    children?: React.ReactNode
}

export function ChapterReadingSettings(props: ChapterReadingSettingsProps) {

    const {
        children,
        ...rest
    } = props

    const [readingDirection, setReadingDirection] = useAtom(readingDirectionAtom)
    const [readingMode, setReadingMode] = useAtom(readingModeAtom)

    return (
        <>
            <Drawer
                side="left"
                trigger={
                    <IconButton
                        icon={<BiCog />}
                        intent="gray-basic"
                        className=""
                    />
                }
                title="Settings"
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
                        options={readingModeOptions}
                        value={readingMode}
                        onValueChange={(value) => setReadingMode(value as any)}
                    />

                    {readingMode !== ReadingMode.LONG_STRIP && (
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
