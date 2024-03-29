import { useMangaPageContainer } from "@/app/(main)/manga/_lib/queries"
import { MangaChapterContainer, MangaChapterDetails, MangaEntry, MangaPageContainer } from "@/app/(main)/manga/_lib/types"
import { LuffyError } from "@/components/shared/luffy-error"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Drawer } from "@/components/ui/drawer"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Select } from "@/components/ui/select"
import { logger } from "@/lib/helpers/debug"
import { atomWithImmer } from "jotai-immer"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"
import { BiCog } from "react-icons/bi"

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

    const readingMode = useAtomValue(readingModeAtom)
    const readingDirection = useAtomValue(readingDirectionAtom)

    const { pageContainer, pageContainerLoading, pageContainerError } = useMangaPageContainer(String(entry?.media?.id || "0"), selectedChapter?.id)

    const [currentPages, setCurrentPages] = useAtom(currentPagesAtom) // Used for PAGED and DOUBLE_PAGE reading modes

    React.useEffect(() => {
        setCurrentPages((draft) => {
            if (!!selectedChapter && (!draft || draft?.chapterId !== selectedChapter?.id)) {
                return {
                    chapterId: selectedChapter?.id || "",
                    pages: readingMode === ReadingMode.PAGED ? [0] : readingMode === ReadingMode.DOUBLE_PAGE ? [0, 1] : [],
                }
            } else {
                return draft
            }
        })
    }, [selectedChapter, pageContainer])

    React.useEffect(() => {
        setCurrentPages((draft) => {
            if (!!pageContainer?.pages?.length && readingMode === ReadingMode.DOUBLE_PAGE && draft?.pages.length === 1) {
                if (draft && draft.pages[0] + 1 < pageContainer?.pages?.length) {
                    draft.pages.push(draft.pages[0] + 1)
                }
                return
            }
        })
    }, [readingMode, pageContainer?.pages?.length])

    React.useEffect(() => {
        logger("manga").info("currentPages: ", currentPages)
    }, [currentPages])

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
                {!!currentPages?.pages?.length && (
                    readingMode === ReadingMode.DOUBLE_PAGE ? (
                        <p className="text-[--muted]">
                            {currentPages?.pages?.length > 1
                                ? `${currentPages?.pages[0] + 1}-${currentPages?.pages[1] + 1}`
                                : currentPages?.pages[0] + 1} / {pageContainer?.pages?.length}
                        </p>
                    ) : (
                        <p className="text-[--muted]">
                            {currentPages?.pages[0] + 1} / {pageContainer?.pages?.length}
                        </p>
                    )
                )}
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
    const [currentPages, setCurrentPages] = useAtom(currentPagesAtom) // Used for PAGED and DOUBLE_PAGE reading modes

    const readingDirection = useAtomValue(readingDirectionAtom)

    const onPaginate = (dir: "left" | "right") => {
        setCurrentPages((draft) => {
            if (draft && pageContainer?.pages?.length) {
                const shouldDecrement = dir === "left" && readingDirection === ReadingDirection.LTR || dir === "right" && readingDirection === ReadingDirection.RTL
                if (shouldDecrement && !draft?.pages.includes(0)) {
                    if (readingMode === ReadingMode.DOUBLE_PAGE) {
                        draft.pages = draft.pages.map((page) => page - 2)
                    } else {
                        draft.pages = draft.pages.map((page) => page - 1)
                    }
                }
                if (!shouldDecrement && !draft?.pages.includes(pageContainer?.pages?.length - 1)) {
                    if (readingMode === ReadingMode.DOUBLE_PAGE) {
                        draft.pages = draft.pages.map((page) => page + 2)
                    } else {
                        draft.pages = draft.pages.map((page) => page + 1)
                    }
                }

                if (draft.pages.length === 1 && readingMode === ReadingMode.DOUBLE_PAGE) {
                    draft.pages.push(draft.pages[0] + 1)
                }

                draft.pages = draft.pages.filter((page) => page >= 0 && page < pageContainer?.pages?.length!)
                return
            }
        })
    }

    const twoDisplayed = readingMode === ReadingMode.DOUBLE_PAGE && currentPages?.pages?.length === 2

    return (
        <div
            className="h-[calc(100dvh-60px)] overflow-y-hidden overflow-x-hidden w-full px-4 space-y-4 select-none relative"
            ref={containerRef}
        >
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
                            "w-full h-[calc(100dvh-60px)] scroll-div min-h-[200px] relative",
                            !currentPages?.pages?.includes(index) && "hidden",
                            twoDisplayed && readingDirection === ReadingDirection.RTL && readingMode === ReadingMode.DOUBLE_PAGE && currentPages?.pages?.[0] === index && "before:content-[''] before:absolute before:w-[3%] before:z-[5] before:h-full before:[background:_linear-gradient(-90deg,_rgba(17,_17,_17,_0)_0,_rgba(17,_17,_17,_.3)_100%)]",
                            twoDisplayed && readingDirection === ReadingDirection.RTL && readingMode === ReadingMode.DOUBLE_PAGE && currentPages?.pages?.[1] === index && "before:content-[''] before:absolute before:right-0 before:w-[3%] before:z-[5] before:h-full before:[background:_linear-gradient(90deg,_rgba(17,_17,_17,_0)_0,_rgba(17,_17,_17,_.3)_100%)]",
                            twoDisplayed && readingDirection === ReadingDirection.LTR && readingMode === ReadingMode.DOUBLE_PAGE && currentPages?.pages?.[1] === index && "before:content-[''] before:absolute before:w-[3%] before:z-[5] before:h-full before:[background:_linear-gradient(-90deg,_rgba(17,_17,_17,_0)_0,_rgba(17,_17,_17,_.3)_100%)]",
                            twoDisplayed && readingDirection === ReadingDirection.LTR && readingMode === ReadingMode.DOUBLE_PAGE && currentPages?.pages?.[0] === index && "before:content-[''] before:absolute before:right-0 before:w-[3%] before:z-[5] before:h-full before:[background:_linear-gradient(90deg,_rgba(17,_17,_17,_0)_0,_rgba(17,_17,_17,_.3)_100%)]",
                        )}
                        id={`page-${index}`}
                    >
                        <img
                            src={page.url} alt={`Page ${index}`} className={cn(
                            "w-full h-full inset-0 object-contain object-center select-none",
                            twoDisplayed && readingDirection === ReadingDirection.RTL && readingMode === ReadingMode.DOUBLE_PAGE && currentPages?.pages?.[0] === index && "[object-position:0%_50%] before:content-['']",
                            twoDisplayed && readingDirection === ReadingDirection.RTL && readingMode === ReadingMode.DOUBLE_PAGE && currentPages?.pages?.[1] === index && "[object-position:100%_50%]",
                            twoDisplayed && readingDirection === ReadingDirection.LTR && readingMode === ReadingMode.DOUBLE_PAGE && currentPages?.pages?.[0] === index && "[object-position:100%_50%]",
                            twoDisplayed && readingDirection === ReadingDirection.LTR && readingMode === ReadingMode.DOUBLE_PAGE && currentPages?.pages?.[1] === index && "[object-position:0%_50%]",
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

                <div className="space-y-2 py-4">


                    <h4>Reading Mode</h4>

                    <Select options={readingModeOptions} value={readingMode} onValueChange={(value) => setReadingMode(value as any)} />

                    {readingMode !== ReadingMode.LONG_STRIP && (
                        <>
                            <h4>Reading direction</h4>

                            <Select
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
