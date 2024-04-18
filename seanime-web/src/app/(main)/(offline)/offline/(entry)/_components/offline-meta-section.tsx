"use client"
import { OfflineAnilistMediaEntryModal } from "@/app/(main)/(offline)/offline/_components/offline-anilist-media-entry-modal"
import { OfflineAnimeEntry, OfflineAssetMap, OfflineMangaEntry } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.types"
import { offline_getAssetUrl } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.utils"
import { AnimeEntryAudienceScore } from "@/app/(main)/entry/_containers/meta-section/_components/anime-entry-metadata-components"
import { ScoreProgressBadges } from "@/app/(main)/entry/_containers/meta-section/_components/score-progress-badges"
import { serverStatusAtom } from "@/atoms/server-status"
import { TextGenerateEffect } from "@/components/shared/styling/text-generate-effect"
import { cn } from "@/components/ui/core/styling"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useThemeSettings } from "@/lib/theme/hooks"
import { motion } from "framer-motion"
import { useAtomValue } from "jotai/react"
import capitalize from "lodash/capitalize"
import Image from "next/image"
import React, { useMemo } from "react"
import { BiCalendarAlt } from "react-icons/bi"
import { useWindowScroll } from "react-use"

type OfflineMetaSectionProps<T extends "anime" | "manga"> = {
    type: T,
    entry: T extends "anime" ? OfflineAnimeEntry : OfflineMangaEntry
    assetMap: OfflineAssetMap | undefined
}

export function OfflineMetaSection<T extends "anime" | "manga">(props: OfflineMetaSectionProps<T>) {

    const { type, entry, assetMap } = props

    const status = useAtomValue(serverStatusAtom)
    const hideAudienceScore = useMemo(() => status?.settings?.anilist?.hideAudienceScore ?? false, [status?.settings?.anilist?.hideAudienceScore])

    const ts = useThemeSettings()
    const { y } = useWindowScroll()

    if (!entry?.media) return null

    return (
        <>
            <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 1, delay: 0.2 }}
                className="__header relative group/meta-section"
            >

                <div
                    className="META_SECTION_FADE_BG w-full absolute z-[1] top-0 h-[35rem] opacity-100 bg-gradient-to-b from-[--background] via-[--background] via-80% to-transparent via"
                />

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

                            <div className="flex gap-8">

                                {entry.media.coverImage?.extraLarge && <div
                                    className="flex-none w-[200px] relative min-h-[16rem] rounded-md overflow-hidden bg-[--background] shadow-md border hidden lg:block"
                                >
                                    <Image
                                        src={offline_getAssetUrl(entry.media.coverImage.extraLarge, assetMap) || ""}
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
                                            className="[text-shadow:_0_1px_10px_rgb(0_0_0_/_20%)] line-clamp-2 pb-1 text-center md:text-left text-pretty text-3xl lg:text-4xl 2xl:text-5xl"
                                            words={entry.media.title?.userPreferred || ""}
                                        />
                                        {(!!entry.media.title?.english && entry.media.title?.userPreferred?.toLowerCase() !== entry.media.title?.english?.toLowerCase()) &&
                                            <h4 className="text-gray-400 line-clamp-2 text-center md:text-left">{entry.media.title?.english}</h4>}
                                        {(!!entry.media.title?.romaji && entry.media.title?.userPreferred?.toLowerCase() !== entry.media.title?.romaji?.toLowerCase()) &&
                                            <h4 className="text-gray-400 line-clamp-2 text-center md:text-left">{entry.media.title?.romaji}</h4>}
                                    </div>

                                    {/*SEASON*/}
                                    {!!entry.media.startDate?.year && (
                                        <div className="flex gap-4 items-center flex-wrap">
                                            <p className="text-lg text-gray-200 flex gap-1 items-center">
                                                <BiCalendarAlt /> {new Intl.DateTimeFormat("en-US", {
                                                year: "numeric",
                                                month: "short",
                                            }).format(new Date(entry.media.startDate?.year || 0,
                                                entry.media.startDate?.month || 0))}
                                            </p>
                                        </div>
                                    )}

                                    {/*PROGRESS*/}
                                    <div className="flex gap-2 md:gap-4 items-center">
                                        <ScoreProgressBadges
                                            score={entry.listData?.score}
                                            progress={entry.listData?.progress}
                                            episodes={type === "anime" ? (entry.media as any)?.episodes : (entry.media as any)?.chapters}
                                        />

                                        <OfflineAnilistMediaEntryModal
                                            listData={entry.listData}
                                            assetMap={assetMap}
                                            media={entry.media}
                                            type={type}
                                        />

                                        <p className="text-base md:text-lg">{capitalize(entry.listData?.status === "CURRENT"
                                            ? (type === "anime" ? "Watching" : "Reading")
                                            : entry.listData?.status)}</p>
                                    </div>

                                    <ScrollArea className="h-16 text-[--muted] hover:text-gray-300 transition-colors duration-500 text-sm pr-2">{entry.media?.description?.replace(
                                        /(<([^>]+)>)/ig,
                                        "")}</ScrollArea>
                                </div>

                            </div>


                            <div className="flex gap-2 items-center">
                                <AnimeEntryAudienceScore meanScore={entry.media?.meanScore} hideAudienceScore={hideAudienceScore} />

                            </div>

                            {/*<AnimeEntryGenres genres={entry?.genres} />*/}

                            {/*<AnimeEntryRanks details={details} />*/}

                        </motion.div>

                    </div>
                </motion.div>

                <div
                    className={cn(
                        "h-[20rem] lg:h-[30rem] 2xl:h-[30rem] w-full flex-none object-cover object-center absolute z-[3] -top-[5rem] overflow-hidden bg-[--background]",
                        !ts.libraryScreenCustomBackgroundImage && cn(
                            "fixed transition-opacity top-0 duration-1000",
                            y > 100 && "opacity-10",
                        ),
                    )}
                >
                    <div
                        className="w-full absolute z-[2] top-0 h-[8rem] opacity-40 bg-gradient-to-b from-[--background] to-transparent via"
                    />
                    <div className="absolute lg:left-[6rem] w-full h-full">
                        {(!!entry.media?.bannerImage || !!entry.media?.coverImage?.extraLarge) && <Image
                            src={offline_getAssetUrl(entry.media.bannerImage, assetMap) || offline_getAssetUrl(entry.media.coverImage?.extraLarge,
                                assetMap) || ""}
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
        </>

    )

}
