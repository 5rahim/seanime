"use client"
import { MediaEntrySilenceToggle } from "@/app/(main)/entry/_components/media-entry-silence-toggle"
import { NextAiringEpisode } from "@/app/(main)/entry/_containers/meta-section/_components/next-airing-episode"
import { ScoreProgressBadges } from "@/app/(main)/entry/_containers/meta-section/_components/score-progress-badges"
import { TorrentSearchButton } from "@/app/(main)/entry/_containers/meta-section/_components/torrent-search-button"
import { getMediaDetailsStats } from "@/app/(main)/entry/_containers/meta-section/helpers"
import { serverStatusAtom } from "@/atoms/server-status"
import { AnilistMediaEntryModal } from "@/components/shared/anilist-media-entry-modal"
import { Badge } from "@/components/ui/badge"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Disclosure, DisclosureContent, DisclosureItem, DisclosureTrigger } from "@/components/ui/disclosure"
import { ScrollArea } from "@/components/ui/scroll-area"
import { MediaDetailsByIdQuery } from "@/lib/anilist/gql/graphql"
import { MediaEntry } from "@/lib/server/types"
import { motion } from "framer-motion"
import { useAtomValue } from "jotai/react"
import capitalize from "lodash/capitalize"
import Image from "next/image"
import Link from "next/link"
import React, { useMemo } from "react"
import { AiFillStar, AiOutlineHeart, AiOutlineStar } from "react-icons/ai"
import { BiCalendarAlt, BiChevronDown, BiHeart } from "react-icons/bi"


export function NewMetaSection(props: { entry: MediaEntry, details: MediaDetailsByIdQuery["Media"] }) {

    const { entry, details } = props

    const status = useAtomValue(serverStatusAtom)
    const hideAudienceScore = useMemo(() => status?.settings?.anilist?.hideAudienceScore ?? false, [status?.settings?.anilist?.hideAudienceScore])

    const {
        seasonHighestRated,
        seasonMostPopular,
        allTimeHighestRated,
    } = getMediaDetailsStats(details)

    if (!entry.media) return null

    return (
        <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 1, delay: 0.2 }}
            className="__header relative bg-[--background]"
        >

            <motion.div
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -20 }}
                transition={{ duration: 0.7, delay: 0.4 }}
                className="pb-[1rem] relative z-[4]"
            >
                <div className="space-y-8 p-6 sm:p-8 lg:max-w-[50%] 2xl:max-w-[60rem] relative">
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

                            {entry.media.coverImage?.large && <div
                                className="flex-none w-[200px] relative rounded-md overflow-hidden bg-[--background] shadow-md border hidden 2xl:block"
                            >
                                <Image
                                    src={entry.media.coverImage.large}
                                    alt="cover image"
                                    fill
                                    priority
                                    className="object-cover object-center"
                                />
                            </div>}


                            <div className="space-y-4">
                                {/*TITLE*/}
                                <div className="space-y-2">
                                    <h1 className="[text-shadow:_0_1px_10px_rgb(0_0_0_/_20%)] text-center md:text-left text-pretty text-3xl lg:text-4xl">{entry.media.title?.userPreferred}</h1>
                                    {entry.media.title?.userPreferred?.toLowerCase() !== entry.media.title?.english?.toLowerCase() &&
                                        <h4 className="text-gray-400 text-center md:text-left">{entry.media.title?.english}</h4>}
                                    {entry.media.title?.userPreferred?.toLowerCase() !== entry.media.title?.romaji?.toLowerCase() &&
                                        <h4 className="text-gray-400 text-center md:text-left">{entry.media.title?.romaji}</h4>}
                                </div>

                                {/*SEASON*/}
                                {!!entry.media.season ? (
                                        <div>
                                            <p className="text-lg text-gray-200 flex w-full gap-1 items-center">
                                                <BiCalendarAlt /> {new Intl.DateTimeFormat("en-US", {
                                                year: "numeric",
                                                month: "short",
                                            }).format(new Date(entry.media.startDate?.year || 0,
                                                entry.media.startDate?.month || 0))} - {capitalize(entry.media.season ?? "")}
                                            </p>
                                        </div>
                                    ) :
                                    (
                                        <p className="text-lg text-gray-200 flex w-full gap-1 items-center">

                                        </p>
                                    )}

                                {/*PROGRESS*/}
                                <div className="flex gap-2 md:gap-4 items-center">
                                    <ScoreProgressBadges
                                        score={entry.listData?.score}
                                        progress={entry.listData?.progress}
                                        episodes={entry.media.episodes}
                                    />
                                    <AnilistMediaEntryModal listData={entry.listData} media={entry.media} />
                                    <p className="text-base md:text-lg">{capitalize(entry.listData?.status === "CURRENT"
                                        ? "Watching"
                                        : entry.listData?.status)}</p>
                                </div>

                                <ScrollArea className="h-16 text-[--muted] text-sm pr-2">{details?.description?.replace(/(<([^>]+)>)/ig,
                                    "")}</ScrollArea>
                            </div>

                        </div>

                        <Disclosure type="multiple" className="space-y-4" defaultValue={[]}>
                            <DisclosureItem value="item-1" className="space-y-2">

                                {/*STUDIO*/}
                                {!!details?.studios?.nodes && <div className="flex gap-2 items-center">
                                    {(!!details?.meanScore && !hideAudienceScore) && (
                                        <Badge
                                            className=""
                                            size="lg"
                                            intent={details.meanScore >= 70 ? details.meanScore >= 85 ? "primary" : "success" : "warning"}
                                            leftIcon={<BiHeart />}
                                        >{details.meanScore / 10}</Badge>
                                    )}
                                    <Badge
                                        size="lg"
                                        intent="gray"
                                        className="rounded-full border-transparent"
                                    >
                                        {details?.studios?.nodes?.[0]?.name}
                                    </Badge>
                                    <DisclosureTrigger>
                                        <IconButton className="rounded-full" size="sm" intent="gray-basic" icon={<BiChevronDown />} />
                                    </DisclosureTrigger>
                                </div>}

                                <DisclosureContent className="space-y-2">
                                    {/*BADGES*/}
                                    <div className="items-center flex flex-wrap gap-2">
                                        {details?.genres?.map(genre => {
                                            return <Badge key={genre!} className="mr-2 border-transparent" size="lg">{genre}</Badge>
                                        })}
                                    </div>

                                    {/*AWARDS*/}
                                    {(!!allTimeHighestRated || !!seasonMostPopular) && <div className="flex-wrap gap-2 hidden md:flex">
                                        {allTimeHighestRated && <Badge
                                            size="lg"
                                            intent="gray"
                                            leftIcon={<AiFillStar />}
                                            iconClass="text-yellow-500"
                                            className="rounded-md border-transparent px-2"
                                        >
                                            #{String(allTimeHighestRated.rank)} Highest
                                            Rated {allTimeHighestRated.format !== "TV" ? `${allTimeHighestRated.format}` : ""} of All
                                            Time
                                        </Badge>}
                                        {seasonHighestRated && <Badge
                                            size="lg"
                                            intent="gray"
                                            leftIcon={<AiOutlineStar />}
                                            iconClass="text-yellow-500"
                                            className="rounded-md border-transparent px-2"
                                        >
                                            #{String(seasonHighestRated.rank)} Highest
                                            Rated {seasonHighestRated.format !== "TV"
                                            ? `${seasonHighestRated.format}`
                                            : ""} of {capitalize(seasonHighestRated.season!)} {seasonHighestRated.year}
                                        </Badge>}
                                        {seasonMostPopular && <Badge
                                            size="lg"
                                            intent="gray"
                                            leftIcon={<AiOutlineHeart />}
                                            iconClass="text-pink-500"
                                            className="rounded-md border-transparent px-2"
                                        >
                                            #{(String(seasonMostPopular.rank))} Most
                                            Popular {seasonMostPopular.format !== "TV"
                                            ? `${seasonMostPopular.format}`
                                            : ""} of {capitalize(seasonMostPopular.season!)} {seasonMostPopular.year}
                                        </Badge>}
                                    </div>}
                                </DisclosureContent>
                            </DisclosureItem>
                        </Disclosure>


                        {entry.media.status !== "NOT_YET_RELEASED" && (
                            <TorrentSearchButton
                                entry={entry}
                            />
                        )}

                        <NextAiringEpisode media={entry.media} />

                        <div className="w-full flex gap-4 items-center">
                            <Link href={`https://anilist.co/anime/${entry.mediaId}`} target="_blank">
                                <Button intent="gray-link" className="px-0">
                                    Open on AniList
                                </Button>
                            </Link>
                            {!!entry.libraryData && <MediaEntrySilenceToggle mediaId={entry.mediaId} />}
                        </div>

                        {(!entry.aniDBId || entry.aniDBId === 0) && (
                            <p className="text-center text-red-300 opacity-50">
                                No mapping found for AniDB. The episodes will have no metadata.
                            </p>
                        )}

                    </motion.div>

                </div>
            </motion.div>

            <div
                className="h-[40rem] w-full flex-none object-cover object-center absolute -top-[5rem] overflow-hidden"
            >
                <div
                    className="w-full absolute z-[2] top-0 h-[8rem] bg-gradient-to-b from-[rgba(0,0,0,0.8)] to-transparent via"
                />
                <div className="absolute lg:left-[16rem] w-full h-full">
                    {(!!entry.media?.bannerImage || !!entry.media?.coverImage?.extraLarge) && <Image
                        src={entry.media?.bannerImage || entry.media?.coverImage?.extraLarge || ""}
                        alt="banner image"
                        fill
                        quality={100}
                        priority
                        sizes="100vw"
                        className="object-cover object-center z-[1]"
                    />}
                    {/*LEFT MASK*/}
                    <div
                        className="hidden lg:block w-[20rem] z-[2] h-full absolute left-0 bg-gradient-to-r from-[--background] via-[--background] via-opacity-50 via-10% to-transparent"
                    />
                </div>
                <div
                    className="w-full z-[2] absolute bottom-0 h-[20rem] bg-gradient-to-t from-[--background] via-[--background] via-opacity-50 via-10% to-transparent"
                />

                <Image
                    src={"/mask-2.png"}
                    alt="mask"
                    fill
                    quality={100}
                    priority
                    sizes="100vw"
                    className={cn(
                        "object-cover object-left z-[2] transition-opacity duration-1000 opacity-90 lg:opacity-90",
                    )}
                />

                {/*<div className="absolute w-full -left-[5rem] h-full">*/}
                {/*    <Image*/}
                {/*        src={"/mask-2.png"}*/}
                {/*        alt="mask"*/}
                {/*        fill*/}
                {/*        quality={100}*/}
                {/*        priority*/}
                {/*        sizes="100vw"*/}
                {/*        className={cn(*/}
                {/*            "object-cover object-left z-[2] transition-opacity opacity-100",*/}
                {/*        )}*/}
                {/*    />*/}
                {/*</div>*/}

            </div>
        </motion.div>

    )

}
