import { AL_BaseAnime, AL_BaseManga, AL_MediaStatus, Anime_EntryListData, Manga_EntryListData, Nullish } from "@/api/generated/types"
import { TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE } from "@/app/(main)/_features/custom-ui/styles"
import { AnilistMediaEntryModal } from "@/app/(main)/_features/media/_containers/anilist-media-entry-modal"
import { imageShimmer } from "@/components/shared/image-helpers"
import { TextGenerateEffect } from "@/components/shared/text-generate-effect"
import { Badge } from "@/components/ui/badge"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Popover } from "@/components/ui/popover"
import { Tooltip } from "@/components/ui/tooltip"
import { getScoreColor } from "@/lib/helpers/score"
import { getImageUrl } from "@/lib/server/assets"
import { ThemeMediaPageBannerSize, ThemeMediaPageBannerType, ThemeMediaPageInfoBoxSize, useIsMobile, useThemeSettings } from "@/lib/theme/hooks"
import capitalize from "lodash/capitalize"
import { motion } from "motion/react"
import Image from "next/image"
import React from "react"
import { BiCalendarAlt, BiSolidStar, BiStar } from "react-icons/bi"
import { MdOutlineSegment } from "react-icons/md"
import { RiSignalTowerFill } from "react-icons/ri"
import { useWindowScroll, useWindowSize } from "react-use"

const MotionImage = motion.create(Image)

type MediaPageHeaderProps = {
    children?: React.ReactNode
    backgroundImage?: string
    coverImage?: string
}

export function MediaPageHeader(props: MediaPageHeaderProps) {

    const {
        children,
        backgroundImage,
        coverImage,
        ...rest
    } = props

    const ts = useThemeSettings()
    const { y } = useWindowScroll()
    const { isMobile } = useIsMobile()

    const bannerImage = backgroundImage || coverImage
    const shouldHideBanner = (
        (ts.mediaPageBannerType === ThemeMediaPageBannerType.HideWhenUnavailable && !backgroundImage)
        || ts.mediaPageBannerType === ThemeMediaPageBannerType.Hide
    )
    const shouldBlurBanner = (ts.mediaPageBannerType === ThemeMediaPageBannerType.BlurWhenUnavailable && !backgroundImage) ||
        ts.mediaPageBannerType === ThemeMediaPageBannerType.Blur

    const shouldDimBanner = (ts.mediaPageBannerType === ThemeMediaPageBannerType.DimWhenUnavailable && !backgroundImage) ||
        ts.mediaPageBannerType === ThemeMediaPageBannerType.Dim

    const shouldShowBlurredBackground = ts.enableMediaPageBlurredBackground && (
        y > 100
        || (shouldHideBanner && !ts.libraryScreenCustomBackgroundImage)
    )


    return (
        <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 1, ease: "easeOut" }}
            className="__meta-page-header relative group/media-page-header"
            data-media-page-header
        >

            {/*<div*/}
            {/*    className={cn(MediaPageHeaderAnatomy.fadeBg({ size }))}*/}
            {/*/>*/}

            {(ts.enableMediaPageBlurredBackground) && <div
                data-media-page-header-blurred-background
                className={cn(
                    "fixed opacity-0 transition-opacity duration-1000 top-0 left-0 w-full h-full z-[4] bg-[--background] rounded-xl",
                    shouldShowBlurredBackground && "opacity-100",
                )}
            >
                <Image
                    data-media-page-header-blurred-background-image
                    src={getImageUrl(bannerImage || "")}
                    alt={""}
                    fill
                    quality={100}
                    sizes="20rem"
                    className={cn(
                        "object-cover object-bottom transition opacity-10",
                        ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && "object-left",
                    )}
                />

                <div
                    data-media-page-header-blurred-background-blur
                    className="absolute top-0 w-full h-full backdrop-blur-2xl z-[2]"
                ></div>
            </div>}

            {children}

            <div
                data-media-page-header-banner
                className={cn(
                    "w-full scroll-locked-offset flex-none object-cover object-center z-[3] bg-[--background] h-[20rem]",
                    ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small ? "lg:h-[28rem]" : "h-[20rem] lg:h-[32rem] 2xl:h-[36.5rem]",
                    ts.libraryScreenCustomBackgroundImage ? "absolute -top-[5rem]" : "fixed transition-opacity top-0 duration-1000",
                    !ts.libraryScreenCustomBackgroundImage && y > 100 && (ts.enableMediaPageBlurredBackground ? "opacity-0" : shouldDimBanner
                        ? "opacity-15"
                        : (y > 300 ? "opacity-5" : "opacity-15")),
                    !ts.disableSidebarTransparency && TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE,
                    shouldHideBanner && "bg-transparent",
                )}
                // style={{
                //     opacity: !ts.libraryScreenCustomBackgroundImage && y > 100 ? (ts.enableMediaPageBlurredBackground ? 0 : shouldDimBanner ? 0.15
                // : 1  - Math.min(y * 0.005, 0.9) ) : 1, }}
            >
                {/*TOP FADE*/}
                <div
                    data-media-page-header-banner-top-gradient
                    className={cn(
                        "w-full absolute z-[2] top-0 h-[8rem] opacity-40 bg-gradient-to-b from-[--background] to-transparent via",
                    )}
                />

                {/*BOTTOM OVERFLOW FADE*/}
                <div
                    data-media-page-header-banner-bottom-gradient
                    className={cn(
                        "w-full z-[2] absolute scroll-locked-offset bottom-[-5rem] h-[5em] bg-gradient-to-b from-[--background] via-transparent via-100% to-transparent",
                        !ts.disableSidebarTransparency && TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE,
                        shouldHideBanner && "hidden",
                    )}
                />

                <motion.div
                    data-media-page-header-banner-image-container
                    className={cn(
                        "absolute top-0 left-0 scroll-locked-offset w-full h-full overflow-hidden",
                        // shouldBlurBanner && "blur-xl",
                        shouldHideBanner && "hidden",
                    )}
                    initial={{ scale: 1, y: 0 }}
                    animate={{
                        scale: !ts.libraryScreenCustomBackgroundImage ? Math.min(1 + y * 0.0002, 1.03) : 1,
                        y: isMobile ? 0 : Math.max(y * -0.9, -10),
                    }}
                    exit={{ scale: 1, y: 0 }}
                    transition={{ duration: 0.6, ease: "easeOut" }}
                >
                    {(!!bannerImage) && <MotionImage
                        data-media-page-header-banner-image
                        src={getImageUrl(bannerImage || "")}
                        alt="banner image"
                        fill
                        quality={100}
                        priority
                        sizes="100vw"
                        className={cn(
                            "object-cover object-center scroll-locked-offset z-[1]",
                            // shouldDimBanner && "!opacity-30",
                        )}
                        initial={{ scale: 1.05, x: 0, y: -10, opacity: 0 }}
                        animate={{ scale: 1, x: 0, y: 1, opacity: shouldDimBanner ? 0.3 : 1 }}
                        transition={{ duration: 0.6, delay: 0.2, ease: "easeOut" }}
                    />}

                    {shouldBlurBanner && <div
                        data-media-page-header-banner-blur
                        className="absolute top-0 w-full h-full backdrop-blur-xl z-[2] "
                    ></div>}

                    {/*LEFT MASK*/}
                    <div
                        data-media-page-header-banner-left-gradient
                        className={cn(
                            "hidden lg:block max-w-[60rem] xl:max-w-[100rem] w-full z-[2] h-full absolute left-0 bg-gradient-to-r from-[--background]  transition-opacity to-transparent",
                            "opacity-85 duration-1000",
                            // y > 300 && "opacity-70",
                        )}
                    />
                    <div
                        data-media-page-header-banner-right-gradient
                        className={cn(
                            "hidden lg:block max-w-[60rem] xl:max-w-[80rem] w-full z-[2] h-full absolute left-0 bg-gradient-to-r from-[--background] from-25% transition-opacity to-transparent",
                            "opacity-50 duration-500",
                        )}
                    />
                </motion.div>

                {/*BOTTOM FADE*/}
                <div
                    data-media-page-header-banner-bottom-gradient
                    className={cn(
                        "w-full z-[3] absolute bottom-0 h-[50%] bg-gradient-to-t from-[--background] via-transparent via-100% to-transparent",
                        shouldHideBanner && "hidden",
                    )}
                />

                <div
                    data-media-page-header-banner-dim
                    className={cn(
                        "absolute h-full w-full block lg:hidden bg-[--background] opacity-70 z-[2]",
                        shouldHideBanner && "hidden",
                    )}
                />

            </div>
        </motion.div>
    )
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type MediaPageHeaderDetailsContainerProps = {
    children?: React.ReactNode
}

export function MediaPageHeaderDetailsContainer(props: MediaPageHeaderDetailsContainerProps) {

    const {
        children,
        ...rest
    } = props

    const ts = useThemeSettings()
    const { y } = useWindowScroll()
    const { width } = useWindowSize()

    return (
        <>
            <motion.div
                initial={{ opacity: 1, y: 0 }}
                animate={{
                    opacity: (width >= 1024 && y > 400) ? Math.max(1 - y * 0.006, 0.1) : 1,
                    y: (width >= 1024 && y > 200) ? Math.max(y * -0.05, -40) : 0,
                }}
                transition={{ duration: 0.4, ease: "easeOut" }}
                className="relative z-[4]"
            >
                <motion.div
                    initial={{ opacity: 0, x: -20 }}
                    animate={{ opacity: 1, x: 0 }}
                    exit={{ opacity: 0, x: -20 }}
                    transition={{ duration: 0.7, delay: 0.4 }}
                    className="relative z-[4]"
                    data-media-page-header-details-container
                >
                    <div
                        data-media-page-header-details-inner-container
                        className={cn(
                            "space-y-8 p-6 sm:p-8 relative",
                            ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && "p-6 sm:py-4 sm:px-8",
                            ts.mediaPageBannerInfoBoxSize === ThemeMediaPageInfoBoxSize.Fluid
                                ? "w-full"
                                : "lg:max-w-[100%] xl:max-w-[80%] 2xl:max-w-[65rem] 5xl:max-w-[80rem]",
                        )}
                    >
                        <motion.div
                            {...{
                                initial: { opacity: 0 },
                                animate: { opacity: 1 },
                                exit: { opacity: 0 },
                                transition: {
                                    type: "spring",
                                    damping: 20,
                                    stiffness: 100,
                                    delay: 0.1,
                                },
                            }}
                            className="space-y-4"
                            data-media-page-header-details-motion-container
                        >

                            {children}

                        </motion.div>

                    </div>
                </motion.div>
            </motion.div>
        </>
    )
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////


type MediaPageHeaderEntryDetailsProps = {
    children?: React.ReactNode
    coverImage?: string
    color?: string
    title?: string
    englishTitle?: string
    romajiTitle?: string
    startDate?: { year?: number, month?: number }
    season?: string
    progressTotal?: number
    status?: AL_MediaStatus
    description?: string
    smallerTitle?: boolean

    listData?: Anime_EntryListData | Manga_EntryListData
    media: AL_BaseAnime | AL_BaseManga
    type: "anime" | "manga"
    offlineAnilistAnimeEntryModal?: React.ReactNode
}

export function MediaPageHeaderEntryDetails(props: MediaPageHeaderEntryDetailsProps) {

    const {
        children,
        coverImage,
        title,
        englishTitle,
        romajiTitle,
        startDate,
        season,
        progressTotal,
        status,
        description,
        color,
        smallerTitle,

        listData,
        media,
        type,
        offlineAnilistAnimeEntryModal,
        ...rest
    } = props

    const ts = useThemeSettings()
    const { y } = useWindowScroll()

    return (
        <>
            <div className="flex flex-col lg:flex-row gap-8" data-media-page-header-entry-details>

                {!!coverImage && <motion.div
                    initial={{ opacity: 0 }}
                    animate={{
                        opacity: 1,
                        // scale: Math.max(1 - y * 0.0002, 0.96),
                        // y: Math.max(y * -0.1, -10)
                    }}
                    transition={{ duration: 0.3 }}
                    data-media-page-header-entry-details-cover-image-container
                    className={cn(
                        "flex-none aspect-[6/8] max-w-[150px] mx-auto lg:m-0 h-auto sm:max-w-[200px] lg:max-w-[230px] w-full relative rounded-[--radius-md] overflow-hidden bg-[--background] shadow-md block",
                        ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && "max-w-[150px] lg:m-0 h-auto sm:max-w-[195px] lg:max-w-[210px] -top-1",
                        ts.mediaPageBannerInfoBoxSize === ThemeMediaPageInfoBoxSize.Fluid && "lg:max-w-[270px]",
                        (ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && ts.mediaPageBannerInfoBoxSize === ThemeMediaPageInfoBoxSize.Fluid) && "lg:max-w-[220px]",
                    )}
                >
                    <motion.div
                        initial={{ scale: 1.1, x: -10 }}
                        animate={{ scale: 1, x: 0 }}
                        transition={{ duration: 0.6, delay: 0.3, ease: "easeOut" }}
                        className="w-full h-full"
                    >
                        <MotionImage
                            data-media-page-header-entry-details-cover-image
                            src={getImageUrl(coverImage)}
                            alt="cover image"
                            fill
                            priority
                            placeholder={imageShimmer(700, 475)}
                            className="object-cover object-center"
                            initial={{ scale: 1.1, x: 0 }}
                            animate={{ scale: Math.min(1 + y * 0.0002, 1.05), x: 0 }}
                            transition={{ duration: 0.3, ease: "easeOut" }}
                        />
                    </motion.div>
                </motion.div>}


                <div
                    data-media-page-header-entry-details-content
                    className={cn(
                        "space-y-2 lg:space-y-4",
                        (ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small || ts.mediaPageBannerInfoBoxSize === ThemeMediaPageInfoBoxSize.Fluid) && "lg:space-y-3",
                    )}
                >
                    {/*TITLE*/}
                    <div className="space-y-2" data-media-page-header-entry-details-title-container>
                        <TextGenerateEffect
                            className={cn(
                                "[text-shadow:_0_1px_10px_rgb(0_0_0_/_20%)] text-white line-clamp-2 pb-1 text-center lg:text-left text-pretty text-3xl 2xl:text-5xl xl:max-w-[50vw]",
                                smallerTitle && "text-3xl 2xl:text-3xl",
                            )}
                            words={title || ""}
                        />
                        {(!!englishTitle && title?.toLowerCase() !== englishTitle?.toLowerCase()) &&
                            <h4 className="text-gray-200 line-clamp-1 text-center lg:text-left xl:max-w-[50vw]">{englishTitle}</h4>}
                        {(!!romajiTitle && title?.toLowerCase() !== romajiTitle?.toLowerCase()) &&
                            <h4 className="text-gray-200 line-clamp-1 text-center lg:text-left xl:max-w-[50vw]">{romajiTitle}</h4>}
                    </div>

                    {/*DATE*/}
                    {!!startDate?.year && (
                        <div
                            className="flex gap-4 items-center flex-wrap justify-center lg:justify-start"
                            data-media-page-header-entry-details-date-container
                        >
                            <p className="text-lg text-white flex gap-1 items-center">
                                <BiCalendarAlt /> {new Intl.DateTimeFormat("en-US", {
                                year: "numeric",
                                month: "short",
                            }).format(new Date(startDate?.year || 0, startDate?.month ? startDate?.month - 1 : 0))}{!!season
                                ? ` - ${capitalize(season)}`
                                : ""}
                            </p>

                            {((status !== "FINISHED" && type === "anime") || type === "manga") && <Badge
                                size="lg"
                                intent={status === "RELEASING" ? "primary" : "gray"}
                                className="bg-transparent border-transparent dark:text-brand-200 px-0 rounded-none"
                                leftIcon={<RiSignalTowerFill />}
                                data-media-page-header-entry-details-date-badge
                            >
                                {capitalize(status || "")?.replaceAll("_", " ")}
                            </Badge>}

                            {ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && <Popover
                                trigger={
                                    <IconButton
                                        intent="white-subtle"
                                        className="rounded-full"
                                        size="sm"
                                        icon={<MdOutlineSegment />}
                                    />
                                }
                                className="max-w-[40rem] bg-[--background] p-4 w-[20rem] lg:w-[40rem] text-md"
                            >
                                <span className="transition-colors">{description?.replace(/(<([^>]+)>)/ig, "")}</span>
                            </Popover>}
                        </div>
                    )}


                    {/*LIST*/}
                    <div className="flex gap-2 md:gap-4 items-center justify-center lg:justify-start" data-media-page-header-entry-details-more-info>

                        <MediaPageHeaderScoreAndProgress
                            score={listData?.score}
                            progress={listData?.progress}
                            episodes={progressTotal}
                        />

                        <AnilistMediaEntryModal listData={listData} media={media} type={type} />

                        {(listData?.status || listData?.repeat) &&
                            <div
                                data-media-page-header-entry-details-status
                                className="text-base text-white md:text-lg flex items-center"
                            >{capitalize(listData?.status === "CURRENT"
                                ? type === "anime" ? "watching" : "reading"
                                : listData?.status)}
                                {listData?.repeat && <Tooltip
                                    trigger={<Badge
                                        size="md"
                                        intent="gray"
                                        className="ml-3"
                                        data-media-page-header-entry-details-repeating-badge
                                    >
                                        {listData?.repeat}

                                    </Badge>}
                                >
                                    {listData?.repeat} {type === "anime" ? "rewatch" : "reread"}{listData?.repeat > 1
                                    ? type === "anime" ? "es" : "s"
                                    : ""}
                                </Tooltip>}
                            </div>}

                    </div>

                    {ts.mediaPageBannerSize !== ThemeMediaPageBannerSize.Small && <Popover
                        trigger={<div
                            className={cn(
                                "cursor-pointer max-h-16 line-clamp-3 col-span-2 left-[-.5rem] text-[--muted] 2xl:max-w-[50vw] hover:text-white transition-colors duration-500 text-sm pr-2",
                                "bg-transparent rounded-[--radius-md] text-center lg:text-left",
                            )}
                            data-media-page-header-details-description-trigger
                        >
                            {description?.replace(/(<([^>]+)>)/ig, "")}
                        </div>}
                        className="max-w-[40rem] bg-[--background] p-4 w-[20rem] lg:w-[40rem] text-md"
                        data-media-page-header-details-description-popover
                    >
                        <span className="transition-colors">{description?.replace(/(<([^>]+)>)/ig, "")}</span>
                    </Popover>}

                    {children}

                </div>

            </div>
        </>
    )
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export function MediaPageHeaderScoreAndProgress({ score, progress, episodes }: {
    score: Nullish<number>,
    progress: Nullish<number>,
    episodes: Nullish<number>,
}) {

    return (
        <>
            {!!score && <Badge
                leftIcon={score >= 90 ? <BiSolidStar className="text-sm" /> : <BiStar className="text-sm" />}
                size="xl"
                intent="unstyled"
                className={getScoreColor(score, "user")}
                data-media-page-header-score-badge
            >
                {score / 10}
            </Badge>}
            <Badge
                size="xl"
                intent="basic"
                className="!text-xl font-bold !text-white px-0 gap-0 rounded-none"
                data-media-page-header-progress-badge
            >
                <span data-media-page-header-progress-badge-progress>{`${progress ?? 0}`}</span><span
                data-media-page-header-progress-total
                className={cn(
                    (!progress || progress !== episodes) && "opacity-60",
                )}
            >/{episodes || "-"}</span>
            </Badge>
        </>
    )

}
