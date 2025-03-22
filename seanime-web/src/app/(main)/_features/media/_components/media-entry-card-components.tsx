import { AL_BaseAnime_NextAiringEpisode, AL_MediaListStatus, AL_MediaStatus } from "@/api/generated/types"
import { MediaCardBodyBottomGradient } from "@/app/(main)/_features/custom-ui/item-bottom-gradients"
import { MediaEntryProgressBadge } from "@/app/(main)/_features/media/_components/media-entry-progress-badge"
import { __ui_fixBorderRenderingArtifacts } from "@/app/(main)/settings/_containers/ui-settings"
import { imageShimmer } from "@/components/shared/image-helpers"
import { SeaLink } from "@/components/shared/sea-link"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/components/ui/core/styling"
import { Tooltip } from "@/components/ui/tooltip"
import { getImageUrl } from "@/lib/server/assets"
import { useThemeSettings } from "@/lib/theme/hooks"
import { addSeconds, formatDistanceToNow } from "date-fns"
import { atom, useAtom } from "jotai/index"
import { useAtomValue } from "jotai/react"
import capitalize from "lodash/capitalize"
import startCase from "lodash/startCase"
import Image from "next/image"
import React, { memo } from "react"
import { BiCalendarAlt } from "react-icons/bi"
import { IoLibrarySharp } from "react-icons/io5"
import { RiSignalTowerLine } from "react-icons/ri"

type MediaEntryCardContainerProps = {
    children?: React.ReactNode
    mRef?: React.RefObject<HTMLDivElement>
} & React.HTMLAttributes<HTMLDivElement>

export function MediaEntryCardContainer(props: MediaEntryCardContainerProps) {

    const {
        children,
        className,
        mRef,
        ...rest
    } = props

    return (
        <div
            data-media-entry-card-container
            ref={mRef}
            className={cn(
                "h-full col-span-1 group/media-entry-card relative flex flex-col place-content-stretch focus-visible:outline-0 flex-none",
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
            data-media-entry-card-overlay
            className={cn(
                "absolute z-[14] top-0 left-0 w-full",
            )}
        >{overlay}</div>
    )
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type MediaEntryCardHoverPopupProps = {
    children?: React.ReactNode
    coverImage?: string
} & React.HTMLAttributes<HTMLDivElement>

export function MediaEntryCardHoverPopup(props: MediaEntryCardHoverPopupProps) {

    const {
        children,
        className,
        coverImage,
        ...rest
    } = props

    const ts = useThemeSettings()
    const markBorderRenderingArtifacts = useAtomValue(__ui_fixBorderRenderingArtifacts)

    return (
        <div
            data-media-entry-card-hover-popup
            className={cn(
                !ts.enableMediaCardBlurredBackground ? "bg-[--media-card-popup-background]" : "bg-[--background]",
                "absolute z-[15] opacity-0 scale-100 border border-[rgb(255_255_255_/_5%)] duration-150",
                "group-hover/media-entry-card:opacity-100 group-hover/media-entry-card:scale-100",
                "group-focus-visible/media-entry-card:opacity-100 group-focus-visible/media-entry-card:scale-100",
                "focus-visible:opacity-100 focus-visible:scale-100",
                "h-[105%] w-[100%] -top-[5%] rounded-[0.7rem] transition ease-in-out",
                "focus-visible:ring-2 ring-brand-400 focus-visible:outline-0",
                "hidden lg:block", // Hide on small screens
                markBorderRenderingArtifacts && "w-[101%] -left-[0.5%]",
            )}
            {...rest}
        >
            {(ts.enableMediaCardBlurredBackground && !!coverImage) &&
                <div
                    data-media-entry-card-hover-popup-image-container
                    className="absolute top-0 left-0 w-full h-full rounded-[--radius] overflow-hidden"
                >
                    <Image
                        data-media-entry-card-hover-popup-image
                        src={getImageUrl(coverImage || "")}
                        alt={"cover image"}
                        fill
                        placeholder={imageShimmer(700, 475)}
                        quality={100}
                    sizes="20rem"
                    className="object-cover object-center transition opacity-20"
                />

                <div
                    data-media-entry-card-hover-popup-image-blur-overlay
                    className="absolute top-0 w-full h-full backdrop-blur-xl z-[0]"
                ></div>
            </div>}

            {ts.enableMediaCardBlurredBackground && <div
                data-media-entry-card-hover-popup-image-blur-gradient
                className="w-full absolute top-0 h-full opacity-60 bg-gradient-to-b from-70% from-[--background] to-transparent z-[2] rounded-[--radius]"
            />}

            <div data-media-entry-card-hover-popup-content className="p-2 h-full w-full flex flex-col justify-between relative z-[2]">
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
            data-media-entry-card-hover-popup-body
            className={cn(
                "space-y-1 select-none",
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
            data-media-entry-card-hover-popup-footer
            className={cn(
                "flex gap-2 items-center",
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
            <div data-media-entry-card-hover-popup-title className="select-none">
                <SeaLink
                    href={link}
                    className="text-center text-pretty font-medium text-sm lg:text-base px-4 leading-0 line-clamp-2 hover:text-brand-100"
                >
                    {title}
                </SeaLink>
            </div>
            {!!year && <div>
                <p
                    data-media-entry-card-hover-popup-title-section-year-season
                    className="justify-center text-sm text-[--muted] flex w-full gap-1 items-center"
                >
                    {startCase(format || "")} - <BiCalendarAlt /> {capitalize(season ?? "")} {year}
                </p>
            </div>}
        </>
    )
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type AnimeEntryCardNextAiringProps = {
    nextAiring: AL_BaseAnime_NextAiringEpisode | undefined
}

export function AnimeEntryCardNextAiring(props: AnimeEntryCardNextAiringProps) {

    const {
        nextAiring,
        ...rest
    } = props

    if (!nextAiring) return null

    return (
        <>
            <div data-anime-entry-card-next-airing-container className="flex gap-1 items-center justify-center">
                {/*<p className="text-xs min-[2000px]:text-md">Next episode:</p>*/}
                <p data-anime-entry-card-next-airing className="text-justify font-normal text-xs min-[2000px]:text-md">
                    Episode <span className="font-semibold">{nextAiring?.episode}</span> {formatDistanceToNow(addSeconds(new Date(),
                    nextAiring?.timeUntilAiring), { addSuffix: true })}
                    {/*<Badge*/}
                    {/*    size="sm"*/}
                    {/*    className="bg-transparent rounded-[--radius]"*/}
                    {/*>{nextAiring?.episode}</Badge>*/}
                </p>
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
            <SeaLink
                href={link}
                className="w-full relative focus-visible:ring-2 ring-[--brand]"
                data-media-entry-card-body-link
            >
                <div
                    data-media-entry-card-body
                    className={cn(
                        "media-entry-card__body aspect-[6/8] flex-none rounded-[--radius] object-cover object-center relative overflow-hidden select-none",
                    )}
                >

                    {/*[CUSTOM UI] BOTTOM GRADIENT*/}
                    <MediaCardBodyBottomGradient />

                    {(showProgressBar && progress && progressTotal) && (
                        <div
                            data-media-entry-card-body-progress-bar-container
                            className={cn(
                                "absolute top-0 w-full h-1 z-[2] bg-gray-700 left-0",
                                listStatus === "COMPLETED" && "hidden",
                            )}
                        >
                            <div
                                data-media-entry-card-body-progress-bar
                                className={cn(
                                    "h-1 absolute z-[2] left-0 bg-gray-200 transition-all",
                                    (listStatus === "CURRENT") ? "bg-brand-400" : "bg-gray-400",
                                )}
                                style={{
                                    width: `${String(Math.ceil((progress / progressTotal) * 100))}%`,
                                }}
                            ></div>
                        </div>
                    )}

                    {(showLibraryBadge) &&
                        <div data-media-entry-card-body-library-badge className="absolute z-[1] left-0 top-0">
                            <Badge
                                size="xl" intent="warning-solid"
                                className="rounded-[--radius] rounded-bl-none rounded-tr-none text-orange-900"
                            ><IoLibrarySharp /></Badge>
                        </div>}

                    {/*RELEASING BADGE*/}
                    {(status === "RELEASING" || status === "NOT_YET_RELEASED") &&
                        <div data-media-entry-card-body-releasing-badge-container className="absolute z-[10] right-1 top-2">
                        <Badge intent={status === "RELEASING" ? "primary-solid" : "zinc-solid"} size="lg"><RiSignalTowerLine /></Badge>
                    </div>}


                    {children}

                    <Image
                        data-media-entry-card-body-image
                        src={getImageUrl(bannerImage || "")}
                        alt={""}
                        fill
                        placeholder={imageShimmer(700, 475)}
                        quality={100}
                        sizes="20rem"
                        className={cn(
                            "object-cover object-center transition-transform",
                            "group-hover/media-entry-card:scale-110",
                            (blurAdultContent && isAdult) && "opacity-80",
                        )}
                    />

                    {(blurAdultContent && isAdult) && <div
                        data-media-entry-card-body-blur-adult-content-overlay
                        className="absolute top-0 w-[125%] h-[125%] -translate-x-[10%] -translate-y-[10%] backdrop-blur-xl z-[3] rounded-[--radius]"
                    ></div>}
                </div>
            </SeaLink>
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
        <div data-media-entry-card-title-section className="pt-2 space-y-1 flex flex-col justify-between h-full select-none">
            <div>
                <p
                    data-media-entry-card-title-section-title
                    className="text-pretty font-medium min-[2000px]:font-semibold text-sm lg:text-[1rem] min-[2000px]:text-lg line-clamp-2"
                >{title}</p>
            </div>
            {(!!season || !!year) && <div>
                <p data-media-entry-card-title-section-year-season className="text-sm text-[--muted] inline-flex gap-1 items-center">
                    {capitalize(season ?? "")} {year}
                </p>
            </div>}
        </div>
    )
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const __mediaEntryCard_hoveredPopupId = atom<number | undefined>(undefined)

export const MediaEntryCardHoverPopupBanner = memo(({
    trailerId,
    showProgressBar,
    mediaId,
    progress,
    progressTotal,
    showTrailer,
    disableAnimeCardTrailers,
    bannerImage,
    isAdult,
    blurAdultContent,
    link,
    listStatus,
    status,
}: {
    mediaId: number
    trailerId?: string
    progress?: number
    progressTotal?: number
    bannerImage?: string
    showProgressBar: boolean
    showTrailer?: boolean
    link: string
    disableAnimeCardTrailers?: boolean
    blurAdultContent?: boolean
    isAdult?: boolean
    listStatus?: AL_MediaListStatus
    status?: AL_MediaStatus
}) => {

    const [trailerLoaded, setTrailerLoaded] = React.useState(false)
    const [actionPopupHoverId] = useAtom(__mediaEntryCard_hoveredPopupId)
    const actionPopupHover = actionPopupHoverId === mediaId
    const [trailerEnabled, setTrailerEnabled] = React.useState(!!trailerId && !disableAnimeCardTrailers && showTrailer)

    const ts = useThemeSettings()

    React.useEffect(() => {
        setTrailerEnabled(!!trailerId && !disableAnimeCardTrailers && showTrailer)
    }, [!!trailerId, !disableAnimeCardTrailers, showTrailer])

    return <SeaLink tabIndex={-1} href={link} data-media-entry-card-hover-popup-banner-link>
        <div data-media-entry-card-hover-popup-banner-container className="aspect-[4/2] relative rounded-[--radius] mb-2 cursor-pointer">
            {(showProgressBar && progress && listStatus && progressTotal && progress !== progressTotal) &&
                <div
                    data-media-entry-card-hover-popup-banner-progress-bar-container
                    className="absolute rounded-[--radius] overflow-hidden top-0 w-full h-1 z-[2] bg-gray-700 left-0"
                >
                    <div
                        data-media-entry-card-hover-popup-banner-progress-bar
                        className={cn(
                            "h-1 absolute z-[2] left-0 bg-gray-200 transition-all",
                            (listStatus === "CURRENT" || listStatus === "COMPLETED") ? "bg-brand-400" : "bg-gray-400",
                        )}
                        style={{ width: `${String(Math.ceil((progress / progressTotal) * 100))}%` }}
                    ></div>
                </div>}

            {(status === "RELEASING" || status === "NOT_YET_RELEASED") &&
                <div data-media-entry-card-hover-popup-banner-releasing-badge-container className="absolute z-[10] right-1 top-2">
                <Tooltip
                    trigger={<Badge intent={status === "RELEASING" ? "primary-solid" : "zinc-solid"} size="lg"><RiSignalTowerLine /></Badge>}
                >
                    {status === "RELEASING" ? "Releasing" : "Not yet released"}
                </Tooltip>
            </div>}

            {(!!bannerImage) ? <Image
                data-media-entry-card-hover-popup-banner-image
                src={getImageUrl(bannerImage || "")}
                alt={"banner"}
                fill
                placeholder={imageShimmer(700, 475)}
                quality={100}
                sizes="20rem"
                className={cn(
                    "object-cover top-0 object-center rounded-[--radius] transition",
                    trailerLoaded && "hidden",
                )}
            /> : <div
                data-media-entry-card-hover-popup-banner-image-gradient
                className="h-full block absolute w-full bg-gradient-to-t from-gray-800 to-transparent"
            ></div>}

            {(blurAdultContent && isAdult) && <div
                data-media-entry-card-hover-popup-banner-blur-adult-content-overlay
                className="absolute top-0 w-full h-full backdrop-blur-xl z-[3] rounded-[--radius]"
            ></div>}

            <div data-media-entry-card-hover-popup-banner-progress-badge-container className="absolute z-[4] left-0 bottom-0">
                <MediaEntryProgressBadge
                    progress={progress}
                    progressTotal={progressTotal}
                    forceShowTotal
                />
            </div>

            {(trailerEnabled && actionPopupHover) && <div
                data-media-entry-card-hover-popup-banner-trailer-container
                className={cn(
                    "absolute w-full h-full overflow-hidden rounded-[--radius]",
                    !trailerLoaded && "hidden",
                )}
            >
                <iframe
                    data-media-entry-card-hover-popup-banner-trailer
                    src={`https://www.youtube-nocookie.com/embed/${trailerId}?autoplay=1&controls=0&mute=1&disablekb=1&loop=1&vq=medium&playlist=${trailerId}&cc_lang_pref=ja`}
                    className={cn(
                        "aspect-video w-full absolute left-0",
                    )}
                    // playsInline
                    // preload="none"
                    // loop
                    allow="autoplay"
                    // muted
                    onLoad={() => setTrailerLoaded(true)}
                    onError={() => setTrailerEnabled(false)}
                />
            </div>}

            {<div
                data-media-entry-card-hover-popup-banner-gradient
                className={cn(
                    "w-full absolute -bottom-1 h-[80%] from-10% bg-gradient-to-t from-[--media-card-popup-background] to-transparent z-[2]",
                    ts.enableMediaCardBlurredBackground && "from-[--background] from-0% bottom-0 rounded-[--radius] opacity-80",
                )}
            />}
        </div>
    </SeaLink>
})
