import { useMangaPageContainer } from "@/app/(main)/manga/_lib/queries"
import { MangaChapterContainer, MangaChapterDetails, MangaEntry, MangaPageContainer } from "@/app/(main)/manga/_lib/types"
import { LuffyError } from "@/components/shared/luffy-error"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Drawer } from "@/components/ui/drawer"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { RadioGroup } from "@/components/ui/radio-group"
import { logger } from "@/lib/helpers/debug"
import { atomWithImmer } from "jotai-immer"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"
import { BiCog } from "react-icons/bi"
import { useMount } from "react-use"
import { toast } from "sonner"

type ChapterDrawerProps = {
    entry: MangaEntry
    chapterContainer: MangaChapterContainer
}

// export const __manga_selectedChapterAtom = atom<MangaChapterDetails | undefined>(undefined)
export const __manga_selectedChapterAtom = atomWithStorage<MangaChapterDetails | undefined>("sea-manga-chapter",
    undefined,
    undefined,
    { getOnInit: true })

const currentPagesAtom = atomWithImmer<{ chapterId: string, pages: number[] } | undefined>(undefined)

export function ChapterDrawer(props: ChapterDrawerProps) {

    const {
        entry,
        chapterContainer,
        ...rest
    } = props

    const [selectedChapter, setSelectedChapter] = useAtom(__manga_selectedChapterAtom)

    const [readingMode, setReadingMode] = useAtom(readingModeAtom)
    const readingDirection = useAtomValue(readingDirectionAtom)

    const { pageContainer, pageContainerLoading, pageContainerError } = useMangaPageContainer(String(entry?.media?.id || "0"), selectedChapter?.id)

    // If the reading mode is set to double page but
    // the pageContainer doesn't have page dimensions, switch to paged mode
    React.useEffect(() => {
        if (readingMode === ReadingMode.DOUBLE_PAGE && !pageContainerLoading && !pageContainerError && !pageContainer?.pageDimensions) {
            toast.error("Could not efficiently get page dimensions from this provider. Switching to paged mode.")
            setReadingMode(ReadingMode.PAGED)
        }
    }, [pageContainer, pageContainerLoading, pageContainerError, readingMode])

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

            <div className="fixed bottom-0 w-full h-12 gap-4 flex items-center px-4 z-[10] bg-[#0c0c0c]">
                <IconButton icon={<AiOutlineArrowLeft />} rounded intent="white-outline" size="xs" onClick={() => setSelectedChapter(undefined)} />
                <h4>
                    {entry?.media?.title?.userPreferred} - {selectedChapter?.title || ""}
                </h4>
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
        console.log(map)
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
            <div className="absolute w-full h-[calc(100dvh-60px)] flex z-[5]">
                <div className="h-full w-full flex flex-1" onClick={() => onPaginate("left")} />
                <div className="h-full w-full flex flex-1" onClick={() => onPaginate("right")} />
            </div>
            <div
                className={cn(
                    twoDisplayed && readingMode === ReadingMode.DOUBLE_PAGE && "flex space-x-2 transition-transform duration-300",
                    twoDisplayed && readingMode === ReadingMode.DOUBLE_PAGE && readingDirection === ReadingDirection.RTL && "flex-row-reverse",
                )}
            >
                {pageContainer?.pages?.map((page, index) => (
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
                        <img
                            src={page.url} alt={`Page ${index}`} className={cn(
                            "w-full h-full inset-0 object-contain object-center select-none",
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

    const setCurrentPages = useSetAtom(currentPagesAtom)
    const [currentDivId, setCurrentDivId] = React.useState("")
    const containerRef = React.useRef<HTMLDivElement>(null)

    const divs = React.useMemo(() => containerRef.current?.querySelectorAll(".scroll-div"), [containerRef.current])
    // Function to handle scroll event
    const handleScroll = () => {
        if (!!containerRef.current) {
            divs?.forEach((div) => {
                if (isElementXPercentInViewport(div)) {
                    setCurrentDivId(div.id)
                }
            })
        }
    }

    // Add scroll event listener when component mounts
    React.useEffect(() => {
        containerRef.current?.addEventListener("scroll", handleScroll)
        return () => containerRef.current?.removeEventListener("scroll", handleScroll)
    }, [containerRef.current])

    React.useEffect(() => {
        logger("manga").info("currentPageId: ", currentDivId)
        if (currentDivId.length > 0 && currentDivId !== "page-0") {
            const pageIndex = Number(currentDivId.split("-")[1])
            setCurrentPages((draft) => {
                if (draft) {
                    draft.pages = [pageIndex]
                }
                return
            })
        }
    }, [currentDivId])

    return (
        <div
            className="container h-[calc(100dvh-60px)] overflow-y-auto overflow-x-hidden max-w-[1400px] px-4 space-y-4 select-none relative"
            ref={containerRef}
        >
            <div className="absolute w-full h-full z-[5]">

            </div>
            {pageContainer?.pages?.map((page, index) => (
                <div key={page.url} className="w-full scroll-div min-h-[200px]" id={`page-${index}`}>
                    <img src={page.url} alt={`Page ${index}`} className="max-w-full h-auto mx-auto select-none" />
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
