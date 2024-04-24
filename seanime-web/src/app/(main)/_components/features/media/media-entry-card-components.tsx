import { AL_BaseMedia_NextAiringEpisode, AL_MediaListStatus, AL_MediaStatus } from "@/api/generated/types"
import { AnimeListItemBottomGradient } from "@/components/shared/custom-ui/item-bottom-gradients"
import { imageShimmer } from "@/components/shared/styling/image-helpers"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/components/ui/core/styling"
import { Tooltip } from "@/components/ui/tooltip"
import { addSeconds, formatDistanceToNow } from "date-fns"
import capitalize from "lodash/capitalize"
import startCase from "lodash/startCase"
import Image from "next/image"
import Link from "next/link"
import React from "react"
import { BiCalendarAlt } from "react-icons/bi"
import { IoLibrarySharp } from "react-icons/io5"
import { RiSignalTowerLine } from "react-icons/ri"

type MediaEntryCardContainerProps = {
    children?: React.ReactNode
} & React.HTMLAttributes<HTMLDivElement>

export function MediaEntryCardContainer(props: MediaEntryCardContainerProps) {

    const {
        children,
        className,
        ...rest
    } = props

    return (
        <div
            className={cn(
                "h-full col-span-1 group/anime-list-item relative flex flex-col place-content-stretch focus-visible:outline-0 flex-none",
                className,
            )}
            {...rest}
        >
            {children}
        </div>
    )
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type MediaEntryCardOverlayProps = {
    overlay?: React.ReactNode
}

export function MediaEntryCardOverlay(props: MediaEntryCardOverlayProps) {

    const {
        overlay,
        ...rest
    } = props

    return (
        <div
            className={cn(
                "absolute z-[14] top-0 left-0 w-full",
            )}
        >{overlay}</div>
    )
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type MediaEntryCardHoverPopupProps = {
    children?: React.ReactNode
} & React.HTMLAttributes<HTMLDivElement>

export function MediaEntryCardHoverPopup(props: MediaEntryCardHoverPopupProps) {

    const {
        children,
        className,
        ...rest
    } = props

    return (
        <div
            className={cn(
                "absolute z-[15] bg-gray-950 opacity-0 scale-70 border",
                "group-hover/anime-list-item:opacity-100 group-hover/anime-list-item:scale-100",
                "group-focus-visible/anime-list-item:opacity-100 group-focus-visible/anime-list-item:scale-100",
                "focus-visible:opacity-100 focus-visible:scale-100",
                "h-[105%] w-[100%] -top-[5%] rounded-md transition ease-in-out",
                "focus-visible:ring-2 ring-brand-400 focus-visible:outline-0",
                "hidden lg:block", // Hide on small screens
            )}
            tabIndex={0}
            {...rest}
        >
            <div className="p-2 h-full w-full flex flex-col justify-between">
                {children}
            </div>
        </div>
    )
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type MediaEntryCardHoverPopupBodyProps = {
    children?: React.ReactNode
} & React.HTMLAttributes<HTMLDivElement>

export function MediaEntryCardHoverPopupBody(props: MediaEntryCardHoverPopupBodyProps) {

    const {
        children,
        className,
        ...rest
    } = props

    return (
        <div
            className={cn(
                "space-y-1",
                className,
            )}
            {...rest}
        >
            {children}
        </div>
    )
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type MediaEntryCardHoverPopupFooterProps = {
    children?: React.ReactNode
} & React.HTMLAttributes<HTMLDivElement>

export function MediaEntryCardHoverPopupFooter(props: MediaEntryCardHoverPopupFooterProps) {

    const {
        children,
        className,
        ...rest
    } = props

    return (
        <div
            className={cn(
                "flex gap-2",
                className,
            )}
            {...rest}
        >
            {children}
        </div>
    )
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type MediaEntryCardHoverPopupTitleSectionProps = {
    link: string
    title: string
    season?: string
    year?: number
    format?: string
}

export function MediaEntryCardHoverPopupTitleSection(props: MediaEntryCardHoverPopupTitleSectionProps) {

    const {
        link,
        title,
        season,
        year,
        format,
        ...rest
    } = props

    return (
        <>
            <div>
                <Link
                    href={link}
                    className="text-center text-pretty font-medium text-sm lg:text-base px-4 leading-0 line-clamp-2 hover:text-brand-100"
                >
                    {title}
                </Link>
            </div>
            {!!year && <div>
                <p className="justify-center text-sm text-[--muted] flex w-full gap-1 items-center">
                    {startCase(format || "")} - <BiCalendarAlt /> {capitalize(season ?? "")} {year}
                </p>
            </div>}
        </>
    )
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type MediaEntryCardNextAiringProps = {
    nextAiring: AL_BaseMedia_NextAiringEpisode | undefined
}

export function MediaEntryCardNextAiring(props: MediaEntryCardNextAiringProps) {

    const {
        nextAiring,
        ...rest
    } = props

    if (!nextAiring) return null

    return (
        <>
            <div className="flex gap-1 items-center justify-center">
                <p className="text-xs min-[2000px]:text-md">Next episode:</p>
                <Tooltip
                    className="bg-gray-200 text-gray-800 font-semibold mb-1"
                    trigger={
                        <p className="text-justify font-normal text-xs min-[2000px]:text-md">
                            <Badge
                                size="sm"
                            >{nextAiring?.episode}</Badge>
                        </p>
                    }
                >{formatDistanceToNow(addSeconds(new Date(), nextAiring?.timeUntilAiring),
                    { addSuffix: true })}</Tooltip>
            </div>
        </>
    )
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type MediaEntryCardBodyProps = {
    link: string
    type: "anime" | "manga"
    title: string
    season?: string
    listStatus?: AL_MediaListStatus
    status?: AL_MediaStatus
    showProgressBar?: boolean
    progress?: number
    progressTotal?: number
    startDate?: { year?: number, month?: number, day?: number }
    bannerImage?: string
    isAdult?: boolean
    showLibraryBadge?: boolean
    children?: React.ReactNode
    blurAdultContent?: boolean
}

export function MediaEntryCardBody(props: MediaEntryCardBodyProps) {

    const {
        link,
        type,
        title,
        season,
        listStatus,
        status,
        showProgressBar,
        progress,
        progressTotal,
        startDate,
        bannerImage,
        isAdult,
        showLibraryBadge,
        children,
        blurAdultContent,
        ...rest
    } = props


    return (
        <>
            <Link
                href={link}
                className="w-full relative"
            >
                <div className="aspect-[6/7] flex-none rounded-md border object-cover object-center relative overflow-hidden">

                    {/*[CUSTOM UI] BOTTOM GRADIENT*/}
                    <AnimeListItemBottomGradient />

                    {(showProgressBar && progress && progressTotal) && <div className="absolute top-0 w-full h-1 z-[2] bg-gray-700 left-0">
                        <div
                            className={cn(
                                "h-1 absolute z-[2] left-0 bg-gray-200 transition-all",
                                {
                                    "bg-brand-400": listStatus === "CURRENT",
                                    "bg-gray-400": listStatus !== "CURRENT",
                                },
                            )}
                            style={{
                                width: `${String(Math.ceil((progress / progressTotal) * 100))}%`,
                            }}
                        ></div>
                    </div>}

                    {(showLibraryBadge) &&
                        <div className="absolute z-[1] left-0 top-0">
                            <Badge
                                size="xl" intent="warning-solid"
                                className="rounded-md rounded-bl-none rounded-tr-none text-orange-900"
                            ><IoLibrarySharp /></Badge>
                        </div>}

                    {/*RELEASING BADGE*/}
                    {status === "RELEASING" && <div className="absolute z-[10] right-1 top-2">
                        <Tooltip
                            trigger={<Badge intent="primary-solid" size="lg"><RiSignalTowerLine /></Badge>}
                        >
                            Airing
                        </Tooltip>
                    </div>}

                    {/*NOT YET RELEASED BADGE*/}
                    {status === "NOT_YET_RELEASED" && <div className="absolute z-[10] right-1 top-1">
                        {(!!startDate && !!startDate?.year) && <Tooltip
                            trigger={<Badge intent="gray-solid" size="lg"><RiSignalTowerLine /></Badge>}
                        >
                            {!!startDate?.year ?
                                new Intl.DateTimeFormat("en-US", {
                                    year: "numeric",
                                    month: "short",
                                    day: "numeric",
                                }).format(new Date(startDate.year, startDate?.month || 0, startDate?.day || 0))
                                : "-"
                            }
                        </Tooltip>}
                    </div>}


                    {children}
                    {/*<ProgressBadge media={media} listData={listData} />*/}
                    {/*<ScoreBadge listData={listData} />*/}

                    <Image
                        src={bannerImage || ""}
                        alt={""}
                        fill
                        placeholder={imageShimmer(700, 475)}
                        quality={100}
                        sizes="20rem"
                        className="object-cover object-center group-hover/anime-list-item:scale-125 transition"
                    />

                    {(blurAdultContent && isAdult) && <div
                        className="absolute top-0 w-full h-full backdrop-blur-xl z-[3] border-4"
                    ></div>}
                </div>
            </Link>
        </>
    )
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type MediaEntryCardTitleSectionProps = {
    title: string
    season?: string
    year?: number
    format?: string
}

export function MediaEntryCardTitleSection(props: MediaEntryCardTitleSectionProps) {

    const {
        title,
        season,
        year,
        format,
        ...rest
    } = props

    return (
        <>
            <div className="pt-2 space-y-2 flex flex-col justify-between h-full">
                <div>
                    <p className="text-center font-semibold text-sm lg:text-md min-[2000px]:text-lg line-clamp-3">{title}</p>
                </div>
                {(!!season || !!year) && <div>
                    <p className="text-sm text-[--muted] inline-flex gap-1 items-center">
                        <BiCalendarAlt />{capitalize(season ?? "")} {year}
                    </p>
                </div>}
            </div>
        </>
    )
}
