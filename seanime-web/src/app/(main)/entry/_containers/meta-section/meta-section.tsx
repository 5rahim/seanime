"use client"
import React, { useMemo } from "react"
import { BiCalendarAlt } from "@react-icons/all-files/bi/BiCalendarAlt"
import { AnilistMediaEntryModal } from "@/components/shared/anilist-media-entry-modal"
import { Badge } from "@/components/ui/badge"
import { Accordion } from "@/components/ui/accordion"
import Image from "next/image"
import Link from "next/link"
import capitalize from "lodash/capitalize"
import { MediaEntry } from "@/lib/server/types"
import { BaseMediaFragment, MediaDetailsByIdQuery } from "@/lib/anilist/gql/graphql"
import { BiHeart } from "@react-icons/all-files/bi/BiHeart"
import { AiFillStar } from "@react-icons/all-files/ai/AiFillStar"
import { AiOutlineStar } from "@react-icons/all-files/ai/AiOutlineStar"
import { AiOutlineHeart } from "@react-icons/all-files/ai/AiOutlineHeart"
import { ScoreProgressBadges } from "@/app/(main)/entry/_containers/meta-section/score-progress-badges"
import formatDistanceToNow from "date-fns/formatDistanceToNow"
import addSeconds from "date-fns/addSeconds"
import { Divider } from "@/components/ui/divider"
import { Button } from "@/components/ui/button"
import { BiDownload } from "@react-icons/all-files/bi/BiDownload"
import { FiSearch } from "@react-icons/all-files/fi/FiSearch"
import { useSetAtom } from "jotai/react"
import { torrentSearchDrawerIsOpenAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"

export function MetaSection(props: { entry: MediaEntry, details: MediaDetailsByIdQuery["Media"] }) {

    const { entry, details } = props

    const relations = (entry.media?.relations?.edges?.map(edge => edge) || [])
        .filter(Boolean)
        .filter(n => (n.node?.format === "TV" || n.node?.format === "OVA" || n.node?.format === "MOVIE") && (n.relationType === "PREQUEL" || n.relationType === "SEQUEL" || n.relationType === "PARENT" || n.relationType === "SIDE_STORY" || n.relationType === "ALTERNATIVE" || n.relationType === "ADAPTATION"))

    const seasonMostPopular = details?.rankings?.find(r => (!!r?.season || !!r?.year) && r?.type === "POPULAR" && r.rank <= 10)
    const allTimeHighestRated = details?.rankings?.find(r => !!r?.allTime && r?.type === "RATED" && r.rank <= 100)
    const seasonHighestRated = details?.rankings?.find(r => (!!r?.season || !!r?.year) && r?.type === "RATED" && r.rank <= 5)
    const allTimeMostPopular = details?.rankings?.find(r => !!r?.allTime && r?.type === "POPULAR" && r.rank <= 100)

    if (!entry.media) return null

    return (
        <div className={"space-y-8 pb-10"}>
            <div className={"space-y-8 p-8 rounded-xl bg-gray-900 bg-opacity-80 drop-shadow-md relative"}>
                <div className={"space-y-4"}>

                    {/*TITLE*/}
                    <div className={"space-y-2"}>
                        <h1 className={"[text-shadow:_0_1px_10px_rgb(0_0_0_/_20%)]"}>{entry.media.title?.userPreferred}</h1>
                        {entry.media.title?.userPreferred?.toLowerCase() !== entry.media.title?.english?.toLowerCase() &&
                            <h4 className={"text-gray-400"}>{entry.media.title?.english}</h4>}
                    </div>

                    {/*SEASON*/}
                    {!!entry.media.season ? (
                            <div>
                                <p className={"text-lg text-gray-200 flex w-full gap-1 items-center"}>
                                    <BiCalendarAlt/> {new Intl.DateTimeFormat("en-US", {
                                    year: "numeric",
                                    month: "short",
                                }).format(new Date(entry.media.startDate?.year || 0, entry.media.startDate?.month || 0))} - {capitalize(entry.media.season ?? "")}
                                </p>
                            </div>
                        ) :
                        (
                            <p className={"text-lg text-gray-200 flex w-full gap-1 items-center"}>

                            </p>
                        )}

                    {/*PROGRESS*/}
                    <div className={"flex gap-4 items-center"}>
                        <ScoreProgressBadges
                            score={entry.listData?.score}
                            progress={entry.listData?.progress}
                            episodes={entry.media.episodes}
                        />
                        <AnilistMediaEntryModal listData={entry.listData} media={entry.media}/>
                        <p className="text-lg">{capitalize(entry.listData?.status === "CURRENT" ? "Watching" : entry.listData?.status)}</p>
                    </div>

                    <p className={"max-h-24 overflow-y-auto"}>{details?.description?.replace(/(<([^>]+)>)/ig, "")}</p>

                    {/*STUDIO*/}
                    {!!details?.studios?.nodes && <div>
                        <span className={"font-bold"}>Studio</span>
                        <Badge
                            className={"ml-2"} size={"lg"}
                            intent={"gray"}
                            badgeClassName={"rounded-md"}
                        >
                            {details?.studios?.nodes?.[0]?.name}
                        </Badge>
                    </div>}


                    {/*BADGES*/}
                    <div className={"items-center flex flex-wrap gap-2"}>
                        {!!(details?.meanScore) && (
                            <Badge
                                className={""}
                                size={"lg"}
                                intent={details.meanScore >= 70 ? details.meanScore >= 85 ? "primary" : "success" : "warning"}
                                leftIcon={<BiHeart/>}
                            >{details.meanScore / 10}</Badge>
                        )}
                        {details?.genres?.map(genre => {
                            return <Badge key={genre!} className={"mr-2"} size={"lg"}>{genre}</Badge>
                        })}
                    </div>

                    {/*AWARDS*/}
                    {(!!allTimeHighestRated || !!seasonMostPopular) && <div className={"flex flex-wrap gap-2"}>
                        {allTimeHighestRated && <Badge
                            className={""} size={"lg"}
                            intent={"gray"}
                            leftIcon={<AiFillStar/>}
                            iconClassName={"text-yellow-500"}
                            badgeClassName={"rounded-md"}
                        >
                            #{String(allTimeHighestRated.rank)} Highest
                            Rated {allTimeHighestRated.format !== "TV" ? `${allTimeHighestRated.format}` : ""} of All
                            Time
                        </Badge>}
                        {seasonHighestRated && <Badge
                            className={""} size={"lg"}
                            intent={"gray"}
                            leftIcon={<AiOutlineStar/>}
                            iconClassName={"text-yellow-500"}
                            badgeClassName={"rounded-md"}
                        >
                            #{String(seasonHighestRated.rank)} Highest
                            Rated {seasonHighestRated.format !== "TV" ? `${seasonHighestRated.format}` : ""} of {capitalize(seasonHighestRated.season!)} {seasonHighestRated.year}
                        </Badge>}
                        {seasonMostPopular && <Badge
                            className={""} size={"lg"}
                            intent={"gray"}
                            leftIcon={<AiOutlineHeart/>}
                            iconClassName={"text-pink-500"}
                            badgeClassName={"rounded-md"}
                        >
                            #{(String(seasonMostPopular.rank))} Most
                            Popular {seasonMostPopular.format !== "TV" ? `${seasonMostPopular.format}` : ""} of {capitalize(seasonMostPopular.season!)} {seasonMostPopular.year}
                        </Badge>}
                    </div>}

                    {entry.media.status !== "NOT_YET_RELEASED" && (
                        <TorrentSearchButton
                            entry={entry}
                        />
                    )}

                    <Divider className="dark:border-gray-800"/>

                    <NextAiringEpisode media={entry.media}/>

                    <div className="w-full flex justify-end">
                        <Link href={`https://anilist.co/anime/${entry.mediaId}`} target="_blank">Open on AniList</Link>
                    </div>

                </div>

            </div>

            <Accordion
                containerClassName={"hidden md:block"}
                triggerClassName={"bg-gray-900 bg-opacity-80 dark:bg-gray-900 dark:bg-opacity-80 hover:bg-gray-800 dark:hover:bg-gray-800 hover:bg-opacity-100 dark:hover:bg-opacity-100"}>
                {relations.length > 0 && (
                    <Accordion.Item title={"Relations"} defaultOpen={true}>
                        <div className={"grid grid-cols-4 gap-4 p-4"}>
                            {relations.slice(0, 4).map(edge => {
                                return <div key={edge.node?.id} className={"col-span-1"}>
                                    <Link href={`/entry?id=${edge.node?.id}`}>
                                        {edge.node?.coverImage?.large && <div
                                            className="h-64 w-full flex-none rounded-md object-cover object-center relative overflow-hidden group/anime-list-item">
                                            <Image
                                                src={edge.node?.coverImage.large}
                                                alt={""}
                                                fill
                                                quality={80}
                                                priority
                                                sizes="10rem"
                                                className="object-cover object-center group-hover/anime-list-item:scale-110 transition"
                                            />
                                            <div
                                                className={"z-[5] absolute bottom-0 w-full h-[60%] bg-gradient-to-t from-black to-transparent"}
                                            />
                                            <Badge
                                                className={"absolute left-2 top-2 font-semibold rounded-md text-[.95rem]"}
                                                intent={"white-solid"}>{edge.node?.format === "MOVIE" ? capitalize(edge.relationType || "").replace("_", " ") + " (Movie)" : capitalize(edge.relationType || "").replace("_", " ")}</Badge>
                                            <div className={"p-2 z-[5] absolute bottom-0 w-full "}>
                                                <p className={"font-semibold line-clamp-2 overflow-hidden"}>{edge.node?.title?.userPreferred}</p>
                                            </div>
                                        </div>}
                                    </Link>
                                </div>
                            })}
                        </div>
                    </Accordion.Item>
                )}
                <Accordion.Item title={"Recommendations"}>
                    <div className={"grid grid-cols-4 gap-4 p-4"}>
                        {details?.recommendations?.edges?.map(edge => edge?.node?.mediaRecommendation).filter(Boolean).map(media => {
                            return <div key={media.id} className={"col-span-1"}>
                                <Link href={`/entry?id=${media.id}`}>
                                    {media.coverImage?.large && <div
                                        className="h-64 w-full flex-none rounded-md object-cover object-center relative overflow-hidden group/anime-list-item">
                                        <Image
                                            src={media.coverImage.large}
                                            alt={""}
                                            fill
                                            quality={80}
                                            priority
                                            sizes="10rem"
                                            className="object-cover object-center group-hover/anime-list-item:scale-110 transition"
                                        />
                                        <div
                                            className={"z-[5] absolute bottom-0 w-full h-[60%] bg-gradient-to-t from-black to-transparent"}
                                        />
                                        <div className={"p-2 z-[5] absolute bottom-0 w-full "}>
                                            <p className={"font-semibold line-clamp-2 overflow-hidden"}>{media.title?.userPreferred}</p>
                                        </div>
                                    </div>}
                                </Link>
                            </div>
                        })}
                    </div>
                </Accordion.Item>
            </Accordion>

        </div>
    )

}

export function TorrentSearchButton({ entry }: { entry: MediaEntry }) {

    const setter = useSetAtom(torrentSearchDrawerIsOpenAtom)
    const count = entry.downloadInfo?.episodesToDownload?.length
    const isMovie = useMemo(() => entry.media?.format === "MOVIE", [entry.media?.format])

    return (
        <div>
            {entry.downloadInfo?.hasInaccurateSchedule && <p className={"text-orange-200 text-center mb-3"}>
                <span className={"block"}>Could not retrieve accurate scheduling information for this show.</span>
                <span className={"block text-[--muted]"}>Please check the schedule online for more information.</span>
            </p>}
            <Button
                className={"w-full"}
                intent={!entry.downloadInfo?.hasInaccurateSchedule ? (!!count ? "white" : "white-subtle") : "warning-subtle"}
                size={"lg"}
                leftIcon={(!!count) ? <BiDownload/> : <FiSearch/>}
                iconClassName={"text-2xl"}
                onClick={() => setter(true)}
            >
                {(!entry.downloadInfo?.hasInaccurateSchedule && !!count) ? <>
                    {(!isMovie) && `Download ${entry.downloadInfo?.batchAll ? "batch /" : "next"} ${count > 1 ? `${count} episodes` : "episode"}`}
                    {(isMovie) && `Download movie`}
                </> : <>
                    Search torrents
                </>}
            </Button>
        </div>
    )
}


export function NextAiringEpisode(props: { media: BaseMediaFragment }) {
    return <>
        {!!props.media.nextAiringEpisode && (
            <div className={"flex gap-2 items-center justify-center"}>
                <p className={"text-xl min-[2000px]:text-xl"}>Next
                    episode {formatDistanceToNow(addSeconds(new Date(), props.media.nextAiringEpisode?.timeUntilAiring), { addSuffix: true })}:</p>

                <p className={"text-justify font-normal text-xl min-[2000px]:text-xl"}>
                    <Badge
                        size={"lg"}>{props.media.nextAiringEpisode?.episode}</Badge>
                </p>

            </div>
        )}
    </>
}