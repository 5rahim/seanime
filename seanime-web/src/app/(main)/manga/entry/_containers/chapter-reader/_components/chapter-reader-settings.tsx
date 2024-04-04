"use client"
import {
    __manga_kbsChapterLeft,
    __manga_kbsChapterRight,
    __manga_kbsPageLeft,
    __manga_kbsPageRight,
    __manga_pageFitAtom,
    __manga_pageGapAtom,
    __manga_pageGapShadowAtom,
    __manga_pageStretchAtom,
    __manga_readingDirectionAtom,
    __manga_readingModeAtom,
    MANGA_DEFAULT_KBS,
    MANGA_KBS,
    MangaPageFit,
    MangaPageStretch,
    MangaReadingDirection,
    MangaReadingMode,
} from "@/app/(main)/manga/entry/_containers/chapter-reader/_lib/manga-chapter-reader.atoms"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Drawer } from "@/components/ui/drawer"
import { RadioGroup } from "@/components/ui/radio-group"
import { Separator } from "@/components/ui/separator"
import { Switch } from "@/components/ui/switch"
import { useAtom } from "jotai/react"
import React from "react"
import { BiCog } from "react-icons/bi"
import { FaRedo, FaRegImage } from "react-icons/fa"
import { GiResize } from "react-icons/gi"
import { MdMenuBook, MdOutlinePhotoSizeSelectLarge } from "react-icons/md"
import { PiArrowCircleLeftDuotone, PiArrowCircleRightDuotone, PiReadCvLogoLight, PiScrollDuotone } from "react-icons/pi"
import { TbArrowAutofitHeight, TbArrowAutofitWidth } from "react-icons/tb"

export type ChapterReaderSettingsProps = {
    children?: React.ReactNode
}

const radioGroupClasses = {
    itemClass: cn(
        "border-transparent absolute top-2 right-2 bg-transparent dark:bg-transparent dark:data-[state=unchecked]:bg-transparent",
        "data-[state=unchecked]:bg-transparent data-[state=unchecked]:hover:bg-transparent dark:data-[state=unchecked]:hover:bg-transparent",
        "focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:ring-offset-transparent",
    ),
    stackClass: "space-y-0 flex gap-2",
    itemIndicatorClass: "hidden",
    itemLabelClass: "font-normal tracking-wide line-clamp-1 truncate flex flex-col items-center data-[state=checked]:text-[--gray] cursor-pointer",
    itemContainerClass: cn(
        "items-start cursor-pointer transition border-transparent rounded-[--radius] py-1.5 px-3 w-full",
        "hover:bg-[--subtle] dark:bg-gray-900",
        "data-[state=checked]:bg-white dark:data-[state=checked]:bg-gray-950",
        "focus:ring-2 ring-transparent dark:ring-transparent outline-none ring-offset-1 ring-offset-[--background] focus-within:ring-2 transition",
        "border border-transparent data-[state=checked]:border-[--gray] data-[state=checked]:ring-offset-0",
        "w-fit",
    ),
}

export const MANGA_READING_MODE_OPTIONS = [
    {
        value: MangaReadingMode.LONG_STRIP,
        label: <span className="flex gap-2 items-center"><PiScrollDuotone className="text-xl" /> <span>Long Strip</span></span>,
    },
    {
        value: MangaReadingMode.PAGED,
        label: <span className="flex gap-2 items-center"><PiReadCvLogoLight className="text-xl" /> <span>Singe Page</span></span>,
    },
    {
        value: MangaReadingMode.DOUBLE_PAGE,
        label: <span className="flex gap-2 items-center"><MdMenuBook className="text-xl" /> <span>Double Page</span></span>,
    },
]

export const MANGA_READING_DIRECTION_OPTIONS = [
    {
        value: MangaReadingDirection.LTR,
        label: <span className="flex gap-2 items-center"><span>Left to Right</span> <PiArrowCircleRightDuotone className="text-2xl" /></span>,
    },
    {
        value: MangaReadingDirection.RTL,
        label: <span className="flex gap-2 items-center"><PiArrowCircleLeftDuotone className="text-2xl" /> <span>Right to Left</span></span>,
    },
]

export const MANGA_PAGE_FIT_OPTIONS = [
    {
        value: MangaPageFit.CONTAIN,
        label: <span className="flex gap-2 items-center"><TbArrowAutofitHeight className="text-xl" /> <span>Contain</span></span>,
    },
    {
        value: MangaPageFit.LARGER,
        label: <span className="flex gap-2 items-center"><TbArrowAutofitHeight className="text-xl" /> <span>Larger</span></span>,
    },
    {
        value: MangaPageFit.COVER,
        label: <span className="flex gap-2 items-center"><TbArrowAutofitWidth className="text-xl" /> <span>Cover</span></span>,
    },
    {
        value: MangaPageFit.TRUE_SIZE,
        label: <span className="flex gap-2 items-center"><FaRegImage className="text-xl" /> <span>True size</span></span>,
    },
]

export const MANGA_PAGE_STRETCH_OPTIONS = [
    {
        value: MangaPageStretch.NONE,
        label: <span className="flex gap-2 items-center"><MdOutlinePhotoSizeSelectLarge className="text-xl" /> <span>None</span></span>,
    },
    {
        value: MangaPageStretch.STRETCH,
        label: <span className="flex gap-2 items-center"><GiResize className="text-xl" /> <span>Stretch</span></span>,
    },
]

const defaultSettings = {
    [MangaReadingMode.LONG_STRIP]: {
        pageFit: MangaPageFit.LARGER,
        pageStretch: MangaPageStretch.NONE,
    },
    [MangaReadingMode.PAGED]: {
        pageFit: MangaPageFit.CONTAIN,
        pageStretch: MangaPageStretch.NONE,
    },
    [MangaReadingMode.DOUBLE_PAGE]: {
        pageFit: MangaPageFit.CONTAIN,
        pageStretch: MangaPageStretch.NONE,
    },
}

export function ChapterReaderSettings(props: ChapterReaderSettingsProps) {

    const {
        children,
        ...rest
    } = props

    const [readingDirection, setReadingDirection] = useAtom(__manga_readingDirectionAtom)
    const [readingMode, setReadingMode] = useAtom(__manga_readingModeAtom)
    const [pageFit, setPageFit] = useAtom(__manga_pageFitAtom)
    const [pageStretch, setPageStretch] = useAtom(__manga_pageStretchAtom)
    const [pageGap, setPageGap] = useAtom(__manga_pageGapAtom)
    const [pageGapShadow, setPageGapShadow] = useAtom(__manga_pageGapShadowAtom)

    const [kbsChapterLeft, setKbsChapterLeft] = useAtom(__manga_kbsChapterLeft)
    const [kbsChapterRight, setKbsChapterRight] = useAtom(__manga_kbsChapterRight)
    const [kbsPageLeft, setKbsPageLeft] = useAtom(__manga_kbsPageLeft)
    const [kbsPageRight, setKbsPageRight] = useAtom(__manga_kbsPageRight)

    const isDefaultSettings =
        pageFit === defaultSettings[readingMode].pageFit &&
        pageStretch === defaultSettings[readingMode].pageStretch

    const resetKeyDefault = React.useCallback((key: string) => {
        switch (key) {
            case MANGA_KBS.kbsChapterLeft:
                setKbsChapterLeft(MANGA_DEFAULT_KBS[key])
                break
            case MANGA_KBS.kbsChapterRight:
                setKbsChapterRight(MANGA_DEFAULT_KBS[key])
                break
            case MANGA_KBS.kbsPageLeft:
                setKbsPageLeft(MANGA_DEFAULT_KBS[key])
                break
            case MANGA_KBS.kbsPageRight:
                setKbsPageRight(MANGA_DEFAULT_KBS[key])
                break
        }
    }, [])

    const resetKbsDefaultIfConflict = (currentKey: string, value: string) => {
        for (const key of Object.values(MANGA_KBS)) {
            if (key !== currentKey) {
                const otherValue = {
                    [MANGA_KBS.kbsChapterLeft]: kbsChapterLeft,
                    [MANGA_KBS.kbsChapterRight]: kbsChapterRight,
                    [MANGA_KBS.kbsPageLeft]: kbsPageLeft,
                    [MANGA_KBS.kbsPageRight]: kbsPageRight,
                }[key]
                console.log(currentKey, key, otherValue, value)
                if (otherValue === value) {
                    resetKeyDefault(key)
                }
            }
        }
    }

    const setKbs = (e: React.KeyboardEvent, kbs: string) => {
        e.preventDefault()
        e.stopPropagation()

        const specialKeys = ["Control", "Shift", "Meta", "Command", "Alt", "Option"]
        if (!specialKeys.includes(e.key)) {
            const keyStr = `${e.metaKey ? "meta+" : ""}${e.ctrlKey ? "ctrl+" : ""}${e.altKey ? "alt+" : ""}${e.shiftKey
                ? "shift+"
                : ""}${e.key.toLowerCase()
                .replace("arrow", "")
                .replace("insert", "ins")
                .replace("delete", "del")
                .replace(" ", "space")
                .replace("+", "plus")}`

            const kbsSetter = {
                [MANGA_KBS.kbsChapterLeft]: setKbsChapterLeft,
                [MANGA_KBS.kbsChapterRight]: setKbsChapterRight,
                [MANGA_KBS.kbsPageLeft]: setKbsPageLeft,
                [MANGA_KBS.kbsPageRight]: setKbsPageRight,
            }

            kbsSetter[kbs]?.(keyStr)
            resetKbsDefaultIfConflict(kbs, keyStr)
        }
    }

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
                size="lg"
            >
                <div className="space-y-4 py-4">
                    <RadioGroup
                        {...radioGroupClasses}
                        label="Reading Mode"
                        options={MANGA_READING_MODE_OPTIONS}
                        value={readingMode}
                        onValueChange={(value) => setReadingMode(value)}
                    />

                    <div
                        className={cn(
                            readingMode === MangaReadingMode.LONG_STRIP && "opacity-50 pointer-events-none",
                        )}
                    >
                        <RadioGroup
                            {...radioGroupClasses}
                            label="Reading Direction"
                            options={MANGA_READING_DIRECTION_OPTIONS}
                            value={readingDirection}
                            onValueChange={(value) => setReadingDirection(value)}
                        />
                    </div>

                    <RadioGroup
                        {...radioGroupClasses}
                        label="Page Fit"
                        options={MANGA_PAGE_FIT_OPTIONS}
                        value={pageFit}
                        onValueChange={(value) => setPageFit(value)}
                        help={<>
                            <p>'Contain': Fit Height</p>
                            <p>'Larger': Fit Height with some overflow for larger images</p>
                            <p>'Cover': Fit Width</p>
                            <p>'True Size': No scaling, raw sizes</p>
                        </>}
                    />

                    <div
                        className={cn(
                            (readingMode !== MangaReadingMode.LONG_STRIP || (pageFit !== MangaPageFit.LARGER && pageFit !== MangaPageFit.CONTAIN)) && "opacity-50 pointer-events-none",
                        )}
                    >
                        <RadioGroup
                            {...radioGroupClasses}
                            label="Page Stretch"
                            options={MANGA_PAGE_STRETCH_OPTIONS}
                            value={pageStretch}
                            onValueChange={(value) => setPageStretch(value)}
                            help="'Stretch' forces pages to have the same width in 'Long Strip' mode."
                        />
                    </div>

                    <div className="flex gap-4 flex-wrap items-center">
                        <Switch
                            label="Page Gap"
                            value={pageGap}
                            onValueChange={setPageGap}
                            fieldClass="w-fit"
                            size="sm"
                        />
                        <Switch
                            label="Page Gap Shadow"
                            value={pageGapShadow}
                            onValueChange={setPageGapShadow}
                            fieldClass="w-fit"
                            disabled={!pageGap}
                            size="sm"
                        />
                    </div>


                    <Button
                        size="sm" className="rounded-full" intent="white-subtle"
                        disabled={isDefaultSettings}
                        onClick={() => {
                            setPageFit(prev => defaultSettings[readingMode].pageFit)
                            setPageStretch(prev => defaultSettings[readingMode].pageStretch)
                        }}
                    >
                        Reset defaults
                        for <span className="w-2"></span> {MANGA_READING_MODE_OPTIONS.find((option) => option.value === readingMode)?.label}
                    </Button>

                    <Separator />

                    <h4>Keyboard shortcuts</h4>

                    {[
                        {
                            key: MANGA_KBS.kbsChapterLeft,
                            label: "Chapter Left",
                            value: kbsChapterLeft,
                        },
                        {
                            key: MANGA_KBS.kbsChapterRight,
                            label: "Chapter Right",
                            value: kbsChapterRight,
                        },
                        {
                            key: MANGA_KBS.kbsPageLeft,
                            label: "Page Left",
                            value: kbsPageLeft,
                        },
                        {
                            key: MANGA_KBS.kbsPageRight,
                            label: "Page Right",
                            value: kbsPageRight,
                        },
                    ].map(item => {
                        return (
                            <div className="flex gap-2 items-center" key={item.key}>
                                <label className="text-[--gray]">{item.label}</label>
                                <Button
                                    onKeyDownCapture={(e) => setKbs(e, item.key)}
                                    className="focus:ring-2 focus:ring-[--brand] focus:ring-offset-1"
                                    size="sm"
                                    intent="white-subtle"
                                >
                                    {item.value}
                                </Button>
                                {
                                    item.value !== (MANGA_DEFAULT_KBS as any)[item.key] && (
                                        <Button
                                            onClick={() => {
                                                resetKeyDefault(item.key)
                                            }}
                                            className="rounded-full"
                                            size="sm"
                                            intent="white-basic"
                                            leftIcon={<FaRedo />}
                                        >
                                            Reset
                                        </Button>
                                    )
                                }
                            </div>
                        )
                    })}
                </div>
            </Drawer>
        </>
    )
}
