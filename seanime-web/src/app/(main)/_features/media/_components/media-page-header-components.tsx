import { AL_BaseAnime, AL_BaseManga, AL_MediaStatus, Anime_EntryListData, Manga_EntryListData, Nullish } from "@/api/generated/types"
import { TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE } from "@/app/(main)/_features/custom-ui/styles"
import { AnilistMediaEntryModal } from "@/app/(main)/_features/media/_containers/anilist-media-entry-modal"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { TextGenerateEffect } from "@/components/shared/text-generate-effect"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/components/ui/core/styling"
import { ScrollArea } from "@/components/ui/scroll-area"
import { getScoreColor } from "@/lib/helpers/score"
import { getImageUrl } from "@/lib/server/assets"
import { useThemeSettings } from "@/lib/theme/hooks"
import { motion } from "framer-motion"
import capitalize from "lodash/capitalize"
import Image from "next/image"
import React from "react"
import { BiCalendarAlt, BiSolidStar, BiStar } from "react-icons/bi"
import { useWindowScroll } from "react-use"


type MediaPageHeaderProps = {
    children?: React.ReactNode
    backgroundImage?: string
}

export function MediaPageHeader(props: MediaPageHeaderProps) {

    const {
        children,
        backgroundImage,
        ...rest
    } = props

    const ts = useThemeSettings()
    const { y } = useWindowScroll()

    return (
        <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 1, delay: 0.2 }}
            className="__meta-page-header relative group/media-page-header"
        >

            {/*<div*/}
            {/*    className={cn(MediaPageHeaderAnatomy.fadeBg({ size }))}*/}
            {/*/>*/}

            {(ts.enableMediaPageBlurredBackground) && <div
                className={cn(
                    "fixed opacity-0 transition-opacity duration-1000 top-0 left-0 w-full h-full z-[4] bg-[--background] rounded-xl",
                    y > 100 && "opacity-100",
                )}
            >
                <Image
                    src={getImageUrl(backgroundImage || "")}
                    alt={""}
                    fill
                    quality={100}
                    sizes="20rem"
                    className="object-cover object-center transition opacity-10"
                />

                <div
                    className="absolute top-0 w-full h-full backdrop-blur-2xl z-[2] "
                ></div>
            </div>}

            {children}

            <div
                className={cn(
                    "w-full scroll-locked-offset flex-none object-cover object-center z-[3] bg-[--background] h-[20rem] lg:h-[32rem] 2xl:h-[40rem]",
                    ts.libraryScreenCustomBackgroundImage ? "absolute -top-[5rem]" : "fixed transition-opacity top-0 duration-1000",
                    !ts.libraryScreenCustomBackgroundImage && y > 100 && (ts.enableMediaPageBlurredBackground ? "opacity-0" : "opacity-5"),
                    !ts.disableSidebarTransparency && TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE,
                )}
            >
                {/*TOP FADE*/}
                <div
                    className="w-full absolute z-[2] top-0 h-[8rem] opacity-40 bg-gradient-to-b from-[--background] to-transparent via"
                />

                {/*BOTTOM OVERFLOW FADE*/}
                <div
                    className={cn(
                        "w-full z-[2] absolute scroll-locked-offset bottom-[-10rem] h-[10rem] bg-gradient-to-b from-[--background] via-transparent via-100% to-transparent",
                        !ts.disableSidebarTransparency && TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE,
                    )}
                />

                <div className="absolute top-0 left-0 scroll-locked-offset w-full h-full">
                    {(!!backgroundImage) && <Image
                        src={getImageUrl(backgroundImage || "")}
                        alt="banner image"
                        fill
                        quality={100}
                        priority
                        sizes="100vw"
                        className="object-cover object-center scroll-locked-offset z-[1]"
                    />}
                    {/*LEFT MASK*/}
                    <div
                        className={cn(
                            "hidden lg:block max-w-[60rem] xl:max-w-[100rem] w-full z-[2] h-full absolute left-0 bg-gradient-to-r from-[--background]  transition-opacity to-transparent",
                            "opacity-85 duration-1000",
                            // y > 300 && "opacity-70",
                        )}
                    />
                    <div
                        className={cn(
                            "hidden lg:block max-w-[60rem] xl:max-w-[80rem] w-full z-[2] h-full absolute left-0 bg-gradient-to-r from-[--background] from-25% transition-opacity to-transparent",
                            "opacity-50 duration-500",
                        )}
                    />
                </div>

                {/*BOTTOM FADE*/}
                <div
                    className="w-full z-[3] absolute bottom-0 h-[50%] bg-gradient-to-t from-[--background] via-transparent via-100% to-transparent"
                />

                <div className="absolute h-full w-full block lg:hidden bg-[--background] opacity-70 z-[2]" />

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

    return (
        <>
            <motion.div
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -20 }}
                transition={{ duration: 0.7, delay: 0.4 }}
                className="relative z-[4]"
            >
                <div className="space-y-8 p-6 sm:p-8 lg:max-w-[70%] 2xl:max-w-[60rem] relative">
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
                    >

                        {children}

                    </motion.div>

                </div>
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

        listData,
        media,
        type,
        offlineAnilistAnimeEntryModal,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    return (
        <>
            <div className="flex flex-col lg:flex-row gap-8">

                {!!coverImage && <div
                    className={cn(
                        "flex-none aspect-[6/8] max-w-[150px] mx-auto lg:m-0 h-auto sm:max-w-[200px] lg:max-w-[230px] w-full relative rounded-md overflow-hidden bg-[--background] shadow-md block",
                    )}
                >
                    <Image
                        src={getImageUrl(coverImage)}
                        alt="cover image"
                        fill
                        priority
                        className="object-cover object-center"
                    />
                </div>}


                <div className="space-y-2 lg:space-y-4">
                    {/*TITLE*/}
                    <div className="space-y-2">
                        <TextGenerateEffect
                            className="[text-shadow:_0_1px_10px_rgb(0_0_0_/_20%)] text-white line-clamp-2 pb-1 text-center lg:text-left text-pretty text-3xl 2xl:text-5xl"
                            words={title || ""}
                        />
                        {(!!englishTitle && title?.toLowerCase() !== englishTitle?.toLowerCase()) &&
                            <h4 className="text-gray-200 line-clamp-2 text-center lg:text-left">{englishTitle}</h4>}
                        {(!!romajiTitle && title?.toLowerCase() !== romajiTitle?.toLowerCase()) &&
                            <h4 className="text-gray-200 line-clamp-2 text-center lg:text-left">{romajiTitle}</h4>}
                    </div>

                    {/*DATE*/}
                    {!!startDate?.year && (
                        <div className="flex gap-4 items-center flex-wrap justify-center lg:justify-start">
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
                                intent={status === "RELEASING" ? "success" : "gray"}
                            >
                                {capitalize(status || "")?.replaceAll("_", " ")}
                            </Badge>}
                        </div>
                    )}


                    {/*LIST*/}
                    <div className="flex gap-2 md:gap-4 items-center justify-center lg:justify-start">

                        <MediaPageHeaderScoreAndProgress
                            score={listData?.score}
                            progress={listData?.progress}
                            episodes={progressTotal}
                        />

                        <AnilistMediaEntryModal listData={listData} media={media} type={type} />

                        <p className="text-base text-white md:text-lg">{capitalize(listData?.status === "CURRENT"
                            ? type === "anime" ? "watching" : "reading"
                            : listData?.status)}</p>
                    </div>

                    <ScrollArea
                        className={cn(
                            "h-20 col-span-2 p-2 left-[-.5rem] text-[--muted] hover:text-white transition-colors duration-500 text-sm pr-2",
                            "bg-transparent hover:bg-zinc-950/30 rounded-md text-center lg:text-left",
                        )}
                    >
                        {description?.replace(/(<([^>]+)>)/ig, "")}
                    </ScrollArea>

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
            >
                {score / 10}
            </Badge>}
            <Badge
                size="xl"
                className="!text-lg font-bold !text-white"
            >
                {`${progress ?? 0}/${episodes || "-"}`}
            </Badge>
        </>
    )

}
