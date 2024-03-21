"use client"
import { EntryOnlinestreamButton } from "@/app/(main)/entry/_components/entry-onlinestream-button"
import { MediaEntrySilenceToggle } from "@/app/(main)/entry/_components/media-entry-silence-toggle"
import { NextAiringEpisode } from "@/app/(main)/entry/_containers/meta-section/_components/next-airing-episode"
import { ScoreProgressBadges } from "@/app/(main)/entry/_containers/meta-section/_components/score-progress-badges"
import { TorrentSearchButton } from "@/app/(main)/entry/_containers/meta-section/_components/torrent-search-button"
import { getMediaDetailsStats } from "@/app/(main)/entry/_containers/meta-section/helpers"
import { serverStatusAtom } from "@/atoms/server-status"
import { AnilistMediaEntryModal } from "@/components/shared/anilist-media-entry-modal"
import { TextGenerateEffect } from "@/components/shared/styling/text-generate-effect"
import { TrailerModal } from "@/components/shared/trailer-modal"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { MediaDetailsByIdQuery } from "@/lib/anilist/gql/graphql"
import { MediaEntry } from "@/lib/server/types"
import { motion } from "framer-motion"
import { useAtomValue } from "jotai/react"
import capitalize from "lodash/capitalize"
import Link from "next/link"
import React, { useMemo } from "react"
import { AiFillStar, AiOutlineHeart, AiOutlineStar } from "react-icons/ai"
import { BiCalendarAlt, BiHeart } from "react-icons/bi"

export function MetaSection(props: { entry: MediaEntry, details: MediaDetailsByIdQuery["Media"] }) {

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
        <div className="space-y-8">
            <div className="space-y-8 p-6 sm:p-8 rounded-xl bg-gray-950 bg-opacity-80 drop-shadow-md relative">
                <motion.div
                    {...{
                        initial: { opacity: 0 },
                        animate: { opacity: 1 },
                        exit: { opacity: 0 },
                        transition: {
                            delay: 0.3,
                            duration: 0.3,
                        },
                    }}
                    className="space-y-4"
                >
                    {/*TITLE*/}
                    <div className="space-y-2">
                        <TextGenerateEffect
                            className="[text-shadow:_0_1px_10px_rgb(0_0_0_/_20%)] line-clamp-2 text-center md:text-left text-pretty text-3xl lg:text-5xl"
                            words={entry.media.title?.userPreferred || ""}
                        />
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

                    <p className="max-h-24 text-[--muted] text-sm overflow-y-auto">{details?.description?.replace(/(<([^>]+)>)/ig, "")}</p>

                    {/*STUDIO*/}
                    {!!details?.studios?.nodes && <div>
                        <span className="font-bold">Studio</span>
                        <Badge
                            size="lg"
                            intent="gray"
                            className="ml-2 rounded-full border-transparent"
                        >
                            {details?.studios?.nodes?.[0]?.name}
                        </Badge>
                    </div>}


                    {/*BADGES*/}
                    <div className="items-center flex flex-wrap gap-2">
                        {(!!details?.meanScore && !hideAudienceScore) && (
                            <Badge
                                className="mr-2"
                                size="lg"
                                intent={details.meanScore >= 70 ? details.meanScore >= 85 ? "primary" : "success" : "warning"}
                                leftIcon={<BiHeart />}
                            >{details.meanScore / 10}</Badge>
                        )}
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

                    {entry.media.status !== "NOT_YET_RELEASED" && (
                        <TorrentSearchButton
                            entry={entry}
                        />
                    )}

                    <Separator className="dark:border-gray-800" />

                    <NextAiringEpisode media={entry.media} />

                    <div className="w-full flex gap-4 flex-wrap items-center">
                        <Link href={`https://anilist.co/anime/${entry.mediaId}`} target="_blank">Open on AniList</Link>

                        <div className="flex flex-1"></div>

                        <EntryOnlinestreamButton entry={entry} />

                        <TrailerModal
                            mediaId={entry.mediaId} trigger={
                            <Button intent="white-subtle">
                                Watch Trailer
                            </Button>
                        }
                        />
                        {!!entry.libraryData ? <MediaEntrySilenceToggle size="md" mediaId={entry.mediaId} /> : <div></div>}
                    </div>

                    {(!entry.aniDBId || entry.aniDBId === 0) && (
                        <p className="text-center text-red-300 opacity-60">
                            No metadata found on AniDB
                        </p>
                    )}

                </motion.div>

            </div>

        </div>
    )

}
