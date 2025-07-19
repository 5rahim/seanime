"use client"
import {
    __manga_doublePageOffsetAtom,
    __manga_entryReaderSettings,
    __manga_hiddenBarAtom,
    __manga_kbsChapterLeft,
    __manga_kbsChapterRight,
    __manga_kbsPageLeft,
    __manga_kbsPageRight,
    __manga_pageFitAtom,
    __manga_pageGapAtom,
    __manga_pageGapShadowAtom,
    __manga_pageOverflowContainerWidthAtom,
    __manga_pageStretchAtom,
    __manga_readerProgressBarAtom,
    __manga_readingDirectionAtom,
    __manga_readingModeAtom,
    MANGA_DEFAULT_KBS,
    MANGA_KBS_ATOM_KEYS,
    MANGA_SETTINGS_ATOM_KEYS,
    MangaPageFit,
    MangaPageStretch,
    MangaReadingDirection,
    MangaReadingMode,
} from "@/app/(main)/manga/_lib/manga-chapter-reader.atoms"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Drawer } from "@/components/ui/drawer"
import { DropdownMenu, DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { NumberInput } from "@/components/ui/number-input"
import { RadioGroup } from "@/components/ui/radio-group"
import { Separator } from "@/components/ui/separator"
import { Switch } from "@/components/ui/switch"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React, { useState } from "react"
import { AiOutlineColumnHeight, AiOutlineColumnWidth } from "react-icons/ai"
import { BiCog } from "react-icons/bi"
import { FaRedo, FaRegImage } from "react-icons/fa"
import { GiResize } from "react-icons/gi"
import { MdMenuBook, MdOutlinePhotoSizeSelectLarge } from "react-icons/md"
import { PiArrowCircleLeftDuotone, PiArrowCircleRightDuotone, PiReadCvLogoLight, PiScrollDuotone } from "react-icons/pi"
import { TbArrowAutofitHeight } from "react-icons/tb"
import { useWindowSize } from "react-use"
import { toast } from "sonner"

export type ChapterReaderSettingsProps = {
    mediaId: number
}

const radioGroupClasses = {
    itemClass: cn(
        "border-transparent absolute top-2 right-2 bg-transparent dark:bg-transparent dark:data-[state=unchecked]:bg-transparent",
        "data-[state=unchecked]:bg-transparent data-[state=unchecked]:hover:bg-transparent dark:data-[state=unchecked]:hover:bg-transparent",
        "focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:ring-offset-transparent",
    ),
    stackClass: "space-y-0 flex flex-wrap gap-2",
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
        label: <span className="flex gap-2 items-center"><PiReadCvLogoLight className="text-xl" /> <span>Single Page</span></span>,
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
        label: <span className="flex gap-2 items-center"><AiOutlineColumnHeight className="text-xl" /> <span>Contain</span></span>,
    },
    {
        value: MangaPageFit.LARGER,
        label: <span className="flex gap-2 items-center"><TbArrowAutofitHeight className="text-xl" /> <span>Overflow</span></span>,
    },
    {
        value: MangaPageFit.COVER,
        label: <span className="flex gap-2 items-center"><AiOutlineColumnWidth className="text-xl" /> <span>Cover</span></span>,
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


export const __manga__readerSettingsDrawerOpen = atom(false)

export function ChapterReaderSettings(props: ChapterReaderSettingsProps) {

    const {
        mediaId,
        ...rest
    } = props

    const [readingDirection, setReadingDirection] = useAtom(__manga_readingDirectionAtom)
    const [readingMode, setReadingMode] = useAtom(__manga_readingModeAtom)
    const [pageFit, setPageFit] = useAtom(__manga_pageFitAtom)
    const [pageStretch, setPageStretch] = useAtom(__manga_pageStretchAtom)
    const [pageGap, setPageGap] = useAtom(__manga_pageGapAtom)
    const [pageGapShadow, setPageGapShadow] = useAtom(__manga_pageGapShadowAtom)
    const [doublePageOffset, setDoublePageOffset] = useAtom(__manga_doublePageOffsetAtom)
    const [pageOverflowContainerWidth, setPageOverflowContainerWidth] = useAtom(__manga_pageOverflowContainerWidthAtom)
    //---
    const [readerProgressBar, setReaderProgressBar] = useAtom(__manga_readerProgressBarAtom)
    const [hiddenBar, setHideBar] = useAtom(__manga_hiddenBarAtom)

    const { width } = useWindowSize()
    const isMobile = width < 950

    const defaultSettings = React.useMemo(() => {
        if (isMobile) {
            return {
                [MangaReadingMode.LONG_STRIP]: {
                    pageFit: MangaPageFit.COVER,
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
        } else {
            return {
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
        }
    }, [isMobile])

    /**
     * Remember settings for current media
     */
    const [entrySettings, setEntrySettings] = useAtom(__manga_entryReaderSettings)

    const mounted = React.useRef(false)
    React.useEffect(() => {
        if (!mounted.current) {
            if (entrySettings[mediaId]) {
                const settings = entrySettings[mediaId]
                setReadingDirection(settings[MANGA_SETTINGS_ATOM_KEYS.readingDirection])
                setReadingMode(settings[MANGA_SETTINGS_ATOM_KEYS.readingMode])
                setPageFit(settings[MANGA_SETTINGS_ATOM_KEYS.pageFit])
                setPageStretch(settings[MANGA_SETTINGS_ATOM_KEYS.pageStretch])
                setPageGap(settings[MANGA_SETTINGS_ATOM_KEYS.pageGap])
                setPageGapShadow(settings[MANGA_SETTINGS_ATOM_KEYS.pageGapShadow])
                setDoublePageOffset(settings[MANGA_SETTINGS_ATOM_KEYS.doublePageOffset])
                setPageOverflowContainerWidth(settings[MANGA_SETTINGS_ATOM_KEYS.overflowPageContainerWidth])
            }
        }
        mounted.current = true
    }, [entrySettings[mediaId]])

    React.useEffect(() => {
        setEntrySettings(prev => ({
            ...prev,
            [mediaId]: {
                [MANGA_SETTINGS_ATOM_KEYS.readingDirection]: readingDirection,
                [MANGA_SETTINGS_ATOM_KEYS.readingMode]: readingMode,
                [MANGA_SETTINGS_ATOM_KEYS.pageFit]: pageFit,
                [MANGA_SETTINGS_ATOM_KEYS.pageStretch]: pageStretch,
                [MANGA_SETTINGS_ATOM_KEYS.pageGap]: pageGap,
                [MANGA_SETTINGS_ATOM_KEYS.pageGapShadow]: pageGapShadow,
                [MANGA_SETTINGS_ATOM_KEYS.doublePageOffset]: doublePageOffset,
                [MANGA_SETTINGS_ATOM_KEYS.overflowPageContainerWidth]: pageOverflowContainerWidth,
            },
        }))
    }, [readingDirection, readingMode, pageFit, pageStretch, pageGap, pageGapShadow, doublePageOffset, pageOverflowContainerWidth])

    const [kbsChapterLeft, setKbsChapterLeft] = useAtom(__manga_kbsChapterLeft)
    const [kbsChapterRight, setKbsChapterRight] = useAtom(__manga_kbsChapterRight)
    const [kbsPageLeft, setKbsPageLeft] = useAtom(__manga_kbsPageLeft)
    const [kbsPageRight, setKbsPageRight] = useAtom(__manga_kbsPageRight)

    const isDefaultSettings =
        pageFit === defaultSettings[readingMode].pageFit &&
        pageStretch === defaultSettings[readingMode].pageStretch

    const resetKeyDefault = React.useCallback((key: string) => {
        switch (key) {
            case MANGA_KBS_ATOM_KEYS.kbsChapterLeft:
                setKbsChapterLeft(MANGA_DEFAULT_KBS[key])
                break
            case MANGA_KBS_ATOM_KEYS.kbsChapterRight:
                setKbsChapterRight(MANGA_DEFAULT_KBS[key])
                break
            case MANGA_KBS_ATOM_KEYS.kbsPageLeft:
                setKbsPageLeft(MANGA_DEFAULT_KBS[key])
                break
            case MANGA_KBS_ATOM_KEYS.kbsPageRight:
                setKbsPageRight(MANGA_DEFAULT_KBS[key])
                break
        }
    }, [])

    const resetKbsDefaultIfConflict = (currentKey: string, value: string) => {
        for (const key of Object.values(MANGA_KBS_ATOM_KEYS)) {
            if (key !== currentKey) {
                const otherValue = {
                    [MANGA_KBS_ATOM_KEYS.kbsChapterLeft]: kbsChapterLeft,
                    [MANGA_KBS_ATOM_KEYS.kbsChapterRight]: kbsChapterRight,
                    [MANGA_KBS_ATOM_KEYS.kbsPageLeft]: kbsPageLeft,
                    [MANGA_KBS_ATOM_KEYS.kbsPageRight]: kbsPageRight,
                }[key]
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
                [MANGA_KBS_ATOM_KEYS.kbsChapterLeft]: setKbsChapterLeft,
                [MANGA_KBS_ATOM_KEYS.kbsChapterRight]: setKbsChapterRight,
                [MANGA_KBS_ATOM_KEYS.kbsPageLeft]: setKbsPageLeft,
                [MANGA_KBS_ATOM_KEYS.kbsPageRight]: setKbsPageRight,
            }

            kbsSetter[kbs]?.(keyStr)
            resetKbsDefaultIfConflict(kbs, keyStr)
        }
    }

    /**
     * Disabled double page on small screens
     */
    React.useEffect(() => {
        if (readingMode === MangaReadingMode.DOUBLE_PAGE && width < 950) {
            setReadingMode(prev => {
                toast.error("Double page mode is not supported on small screens.")
                return MangaReadingMode.LONG_STRIP
            })
        }
    }, [width, readingMode])

    function handleSetReadingMode(mode: string) {
        if (mode === MangaReadingMode.DOUBLE_PAGE && width < 950) {
            toast.error("Double page mode is not supported on small screens.")
            return
        }
        setReadingMode(mode)
    }

    const [open, setOpen] = useAtom(__manga__readerSettingsDrawerOpen)
    const [fullscreen, setFullscreen] = useState(false)

    function handleFullscreen() {
        const el = document.documentElement
        if (fullscreen && document.exitFullscreen) {
            document.exitFullscreen()
            setFullscreen(false)
        } else if (!fullscreen) {
            if (el.requestFullscreen) {
                el.requestFullscreen()
            } else if ((el as any).webkitRequestFullscreen) {
                (el as any).webkitRequestFullscreen()
            } else if ((el as any).msRequestFullscreen) {
                (el as any).msRequestFullscreen()
            }
            setFullscreen(true)
        }
    }
    return (
        <>
            <DropdownMenu
                trigger={<IconButton
                    data-chapter-reader-settings-dropdown-menu-trigger
                    icon={<BiCog />}
                    intent="gray-basic"
                    className="flex lg:hidden"
                />}
                className="block lg:hidden"
                data-chapter-reader-settings-dropdown-menu
            >
                <DropdownMenuItem
                    onClick={() => setOpen(true)}
                >Open settings</DropdownMenuItem>
                <DropdownMenuItem
                    onClick={handleFullscreen}
                >Toggle fullscreen</DropdownMenuItem>
                <DropdownMenuItem
                    onClick={() => setHideBar((prev) => !prev)}
                >{hiddenBar ? "Show" : "Hide"} bar</DropdownMenuItem>
            </DropdownMenu>

            <Drawer
                trigger={
                    <IconButton
                        icon={<BiCog />}
                        intent="gray-basic"
                        className="hidden lg:flex"
                    />
                }
                title="Settings"
                allowOutsideInteraction={false}
                open={open}
                onOpenChange={setOpen}
                size="lg"
                contentClass="z-[51]"
                data-chapter-reader-settings-drawer
            >
                <div className="space-y-4 py-4" data-chapter-reader-settings-drawer-content>

                    <RadioGroup
                        {...radioGroupClasses}
                        label="Reading Mode"
                        options={MANGA_READING_MODE_OPTIONS}
                        value={readingMode}
                        onValueChange={(value) => handleSetReadingMode(value)}
                    />

                    <div
                        className={cn(
                            readingMode !== MangaReadingMode.DOUBLE_PAGE && "hidden",
                        )}
                    >
                        <NumberInput
                            label="Offset"
                            value={doublePageOffset}
                            onValueChange={(value) => setDoublePageOffset(value)}
                        />
                    </div>
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
                    // help={<>
                    //     <p>'Contain': Fit Height</p>
                    //     <p>'Overflow': Height overflow</p>
                    //     <p>'Cover': Fit Width</p>
                    //     <p>'True Size': No scaling, raw sizes</p>
                    // </>}
                    />

                    {
                        pageFit === MangaPageFit.LARGER && (
                            <NumberInput
                                label="Page Container Width"
                                max={100}
                                min={0}
                                rightAddon="%"
                                value={pageOverflowContainerWidth}
                                onValueChange={(value) => setPageOverflowContainerWidth(value)}
                                disabled={readingMode === MangaReadingMode.DOUBLE_PAGE}
                            />
                        )
                    }

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
                            help="'Stretch' forces all pages to have the same width as the container in 'Long Strip' mode."
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
                        size="sm" className="rounded-full w-full" intent="white-subtle"
                        disabled={isDefaultSettings}
                        onClick={() => {
                            setPageFit(defaultSettings[readingMode].pageFit)
                            setPageStretch(defaultSettings[readingMode].pageStretch)
                        }}
                    >
                        <span className="flex flex-none items-center">
                            Reset defaults
                            for <span className="w-2"></span> {MANGA_READING_MODE_OPTIONS.find((option) => option.value === readingMode)?.label}
                        </span>
                    </Button>

                    <Separator />

                    <div className="flex items-center gap-4">
                        <Switch
                            label="Progress Bar"
                            value={readerProgressBar}
                            onValueChange={setReaderProgressBar}
                            fieldClass="w-fit"
                            size="sm"
                        />
                    </div>

                    <Separator />

                    {!isMobile && (
                        <>
                            <div>
                                <h4>Editable Keybindings</h4>
                                <p className="text-[--muted] text-xs">Click to edit</p>
                            </div>

                            {[
                                {
                                    key: MANGA_KBS_ATOM_KEYS.kbsChapterLeft,
                                    label: readingDirection === MangaReadingDirection.LTR ? "Previous chapter" : "Next chapter",
                                    value: kbsChapterLeft,
                                    // help: readingDirection === MangaReadingDirection.LTR ? "Previous chapter" : "Next chapter",
                                },
                                {
                                    key: MANGA_KBS_ATOM_KEYS.kbsChapterRight,
                                    label: readingDirection === MangaReadingDirection.LTR ? "Next chapter" : "Previous chapter",
                                    value: kbsChapterRight,
                                    // help: readingDirection === MangaReadingDirection.LTR ? "Next chapter" : "Previous chapter",
                                },
                                {
                                    key: MANGA_KBS_ATOM_KEYS.kbsPageLeft,
                                    label: readingDirection === MangaReadingDirection.LTR ? "Previous page" : "Next page",
                                    value: kbsPageLeft,
                                    // help: readingDirection === MangaReadingDirection.LTR ? "Previous page" : "Next page",
                                },
                                {
                                    key: MANGA_KBS_ATOM_KEYS.kbsPageRight,
                                    label: readingDirection === MangaReadingDirection.LTR ? "Next page" : "Previous page",
                                    value: kbsPageRight,
                                    // help: readingDirection === MangaReadingDirection.LTR ? "Next page" : "Previous page",
                                },
                            ].map(item => {
                                return (
                                    <div className="flex gap-2 items-center" key={item.key}>
                                        <div className="">
                                            <Button
                                                onKeyDownCapture={(e) => setKbs(e, item.key)}
                                                className="focus:ring-2 focus:ring-[--brand] focus:ring-offset-1 focus-visible:ring-2 focus-visible:ring-[--brand] focus-visible:ring-offset-1"
                                                size="sm"
                                                intent="primary-subtle"
                                                id={`chapter-reader-settings-kbs-${item.key}`}
                                                onClick={() => {
                                                    const el = document.getElementById(`chapter-reader-settings-kbs-${item.key}`)
                                                    if (el) {
                                                        el.focus()
                                                    }
                                                }}
                                            >
                                                {item.value}
                                            </Button>
                                        </div>
                                        <label className="text-[--gray]">
                                            <span className="font-semibold">{item.label}</span>
                                            {/*{!!item.help && <span className="ml-2 text-[--muted]">({item.help})</span>}*/}
                                        </label>
                                        {
                                            item.value !== (MANGA_DEFAULT_KBS as any)[item.key] && (
                                                <Button
                                                    onClick={() => {
                                                        resetKeyDefault(item.key)
                                                    }}
                                                    className="rounded-full"
                                                    size="sm"
                                                    intent="warning-subtle"
                                                    leftIcon={<FaRedo />}
                                                >
                                                    Reset
                                                </Button>
                                            )
                                        }
                                    </div>
                                )
                            })}

                            <Separator />

                            <h4>Keyboard Shortcuts</h4>

                            {[{
                                key: "u",
                                label: "Update progress and go to next chapter",
                            }, {
                                key: "b",
                                label: "Toggle bottom bar visibility",
                            }, {
                                key: "m",
                                label: "Switch reading mode",
                            }, {
                                key: "d",
                                label: "Switch reading direction",
                            }, {
                                key: "f",
                                label: "Switch page fit",
                            }, {
                                key: "s",
                                label: "Switch page stretch",
                            }, {
                                key: "shift+right",
                                label: "Increment double page offset",
                            }, {
                                key: "shift+left",
                                label: "Decrement double page offset",
                            }].map(item => {
                                return (
                                    <div className="flex gap-2 items-center" key={item.key}>
                                        <div>
                                            <Button
                                                size="sm"
                                                intent="white-subtle"
                                                className="pointer-events-none"
                                            >
                                                {item.key}
                                            </Button>
                                        </div>
                                        <p>{item.label}</p>
                                    </div>
                                )
                            })}
                        </>
                    )}
                </div>
            </Drawer>
        </>
    )
}
