import {
    AL_BaseManga,
    AL_BaseMedia,
    AL_MediaStatus,
    Anime_MediaEntryListData,
    Manga_EntryListData,
    Nullish,
    Offline_ListData,
} from "@/api/generated/types"
import { AnilistMediaEntryModal } from "@/app/(main)/_features/media/_containers/anilist-media-entry-modal"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { TextGenerateEffect } from "@/components/shared/text-generate-effect"
import { Badge } from "@/components/ui/badge"
import { cn, defineStyleAnatomy } from "@/components/ui/core/styling"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useThemeSettings } from "@/lib/theme/hooks"
import { cva, VariantProps } from "class-variance-authority"
import { motion } from "framer-motion"
import capitalize from "lodash/capitalize"
import Image from "next/image"
import React from "react"
import { BiCalendarAlt, BiStar } from "react-icons/bi"
import { useWindowScroll } from "react-use"

export const MediaPageHeaderAnatomy = defineStyleAnatomy({
    fadeBg: cva([
        "__media-page-header-fade-bg",
        "w-full absolute z-[1] top-0",
        "opacity-100 bg-gradient-to-b from-[--background] via-[--background] via-80% to-transparent via",
    ], {
        variants: {
            size: {
                normal: "h-[35rem] lg:h-[35rem] 2xl:h-[45rem]",
                smaller: "h-[35rem]",
            },
        },
        defaultVariants: {
            size: "normal",
        },
    }),
    imageContainer: cva([
        "__media-page-header-image-container",
        " w-full flex-none object-cover object-center z-[3] overflow-hidden bg-[--background]",
    ], {
        variants: {
            flavor: {
                fixed: "fixed transition-opacity top-0 duration-1000",
                absolute: "absolute -top-[5rem]",
            },
            size: {
                normal: "h-[20rem] lg:h-[32rem] 2xl:h-[40rem]",
                smaller: "h-[20rem] lg:h-[30rem] 2xl:h-[30rem]",
            },
        },
        defaultVariants: {
            size: "normal",
        },
    }),
})

type MediaPageHeaderProps = {
    children?: React.ReactNode
    backgroundImage?: string
} & VariantProps<typeof MediaPageHeaderAnatomy.imageContainer>

export function MediaPageHeader(props: MediaPageHeaderProps) {

    const {
        children,
        backgroundImage,
        size,
        ...rest
    } = props

    const ts = useThemeSettings()
    const { y } = useWindowScroll()

    return (
        <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 1, delay: 0.2 }}
            className="__meta-page-header relative group/media-page-header"
        >

            <div
                className={cn(MediaPageHeaderAnatomy.fadeBg({ size }))}
            />

            {children}

            <div
                className={cn(
                    MediaPageHeaderAnatomy.imageContainer({
                        size,
                        flavor: ts.libraryScreenCustomBackgroundImage ? "absolute" : "fixed",
                    }),
                    !ts.libraryScreenCustomBackgroundImage && y > 100 && "opacity-5",
                )}
            >
                <div
                    className="w-full absolute z-[2] top-0 h-[8rem] opacity-40 bg-gradient-to-b from-[--background] to-transparent via"
                />
                <div className="absolute lg:left-[6rem] w-full h-full">
                    {(!!backgroundImage) && <Image
                        src={backgroundImage || ""}
                        alt="banner image"
                        fill
                        quality={100}
                        priority
                        sizes="100vw"
                        className="object-cover object-center z-[1]"
                    />}
                    {/*LEFT MASK*/}
                    <div
                        className="hidden lg:block w-[30rem] z-[2] h-full absolute left-0 bg-gradient-to-r from-[--background] via-[--background] via-opacity-50 via-10% to-transparent"
                    />
                </div>
                <div
                    className="w-full z-[3] absolute bottom-0 h-[5rem] bg-gradient-to-t from-[--background] via-transparent via-100% to-transparent"
                />

                <Image
                    src={"/mask-2.png"}
                    alt="mask"
                    fill
                    quality={100}
                    priority
                    sizes="100vw"
                    className={cn(
                        "hidden lg:block object-cover object-left z-[2] transition-opacity duration-1000 opacity-90 lg:opacity-70 lg:group-hover/meta-section:opacity-80",
                    )}
                />

                <div className="absolute h-full w-full block lg:hidden bg-gray-950 opacity-70 z-[2]" />

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
    title?: string
    englishTitle?: string
    romajiTitle?: string
    startDate?: { year?: number, month?: number }
    season?: string
    progressTotal?: number
    status?: AL_MediaStatus
    description?: string

    listData?: Anime_MediaEntryListData | Manga_EntryListData | Offline_ListData
    media: AL_BaseMedia | AL_BaseManga
    type: "anime" | "manga"
    offlineAnilistMediaEntryModal?: React.ReactNode
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

        listData,
        media,
        type,
        offlineAnilistMediaEntryModal,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    return (
        <>
            <div className="flex gap-8">

                {!!coverImage && <div
                    className="flex-none w-[200px] relative rounded-md overflow-hidden bg-[--background] shadow-md border hidden lg:block"
                >
                    <Image
                        src={coverImage}
                        alt="cover image"
                        fill
                        priority
                        className="object-cover object-center"
                    />
                </div>}


                <div className="space-y-4">
                    {/*TITLE*/}
                    <div className="space-y-2">
                        <TextGenerateEffect
                            className="[text-shadow:_0_1px_10px_rgb(0_0_0_/_20%)] line-clamp-2 pb-1 text-center lg:text-left text-pretty text-3xl 2xl:text-5xl"
                            words={title || ""}
                        />
                        {(!!englishTitle && title?.toLowerCase() !== englishTitle?.toLowerCase()) &&
                            <h4 className="text-[--muted] line-clamp-2 text-center lg:text-left">{englishTitle}</h4>}
                        {(!!romajiTitle && title?.toLowerCase() !== romajiTitle?.toLowerCase()) &&
                            <h4 className="text-[--muted] line-clamp-2 text-center lg:text-left">{romajiTitle}</h4>}
                    </div>

                    {/*DATE*/}
                    {!!startDate?.year && (
                        <div className="flex gap-4 items-center flex-wrap justify-center lg:justify-start">
                            <p className="text-lg text-gray-200 flex gap-1 items-center">
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

                        {!serverStatus?.isOffline ?
                            <AnilistMediaEntryModal listData={listData} media={media} type={type} /> :
                            offlineAnilistMediaEntryModal}

                        <p className="text-base md:text-lg">{capitalize(listData?.status === "CURRENT"
                            ? type === "anime" ? "watching" : "reading"
                            : listData?.status)}</p>
                    </div>

                    <ScrollArea className="h-16 text-[--muted] hover:text-gray-300 transition-colors duration-500 text-sm pr-2">{description?.replace(
                        /(<([^>]+)>)/ig,
                        "")}</ScrollArea>
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

    const scoreColor = score ? (
        score < 50 ? "bg-red-500" :
            score < 70 ? "bg-gray-500" :
                score < 85 ? "bg-green-500" :
                    "bg-brand-500 text-white"
    ) : ""

    return (
        <>
            {!!score && <Badge leftIcon={<BiStar />} size="xl" intent="primary-solid" className={scoreColor}>
                {score / 10}
            </Badge>}
            <Badge
                size="xl"
                className="!text-lg font-bold !text-yellow-50"
            >
                {`${progress ?? 0}/${episodes || "-"}`}
            </Badge>
        </>
    )

}
