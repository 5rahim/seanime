import { useMediaEntryBulkAction } from "@/app/(main)/(library)/_containers/bulk-actions/_lib/media-entry-bulk-actions"
import { MediaEntryLibraryData, MediaEntryListData } from "@/app/(main)/(library)/_lib/anime-library.types"
import { getAtomicLibraryEntryAtom } from "@/app/(main)/_lib/anime-library-collection.atoms"
import { AnimeEntryAudienceScore } from "@/app/(main)/entry/_containers/meta-section/_components/anime-entry-metadata-components"
import { serverStatusAtom } from "@/atoms/server-status"
import { AnilistMediaEntryModal } from "@/components/shared/anilist-media-entry-modal"
import { AnimeListItemBottomGradient } from "@/components/shared/custom-ui/item-bottom-gradients"
import { imageShimmer } from "@/components/shared/styling/image-helpers"
import { TrailerModal } from "@/components/shared/trailer-modal"
import { Badge } from "@/components/ui/badge"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Tooltip } from "@/components/ui/tooltip"
import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import { addSeconds, formatDistanceToNow } from "date-fns"
import { atom, useAtom } from "jotai"
import { useAtomValue, useSetAtom } from "jotai/react"
import capitalize from "lodash/capitalize"
import startCase from "lodash/startCase"
import Image from "next/image"
import Link from "next/link"
import { usePathname } from "next/navigation"
import React, { memo, useEffect, useLayoutEffect, useState } from "react"
import { BiCalendarAlt, BiLockOpenAlt, BiPlay, BiStar } from "react-icons/bi"
import { IoLibrarySharp } from "react-icons/io5"
import { RiSignalTowerLine } from "react-icons/ri"
import { VscVerified } from "react-icons/vsc"

type AnimeListItemProps = {
    media: BaseMediaFragment,
    listData?: MediaEntryListData
    libraryData?: MediaEntryLibraryData
    showLibraryBadge?: boolean
    overlay?: React.ReactNode
    showListDataButton?: boolean
    showTrailer?: boolean
    isManga?: boolean
    withAudienceScore?: boolean
} & {
    containerClassName?: string
}

const actionPopupHoverAtom = atom<number | undefined>(undefined)

export const AnimeListItem = ((props: AnimeListItemProps) => {

    const serverStatus = useAtomValue(serverStatusAtom)
    const {
        media,
        listData: _listData,
        libraryData: _libraryData,
        overlay,
        showListDataButton,
        showTrailer: _showTrailer,
        isManga,
        withAudienceScore = true,
    } = props

    const [listData, setListData] = useState(_listData)
    const [libraryData, setLibraryData] = useState(_libraryData)
    const setActionPopupHover = useSetAtom(actionPopupHoverAtom)

    const [__atomicLibraryCollection, getAtomicLibraryEntry] = useAtom(getAtomicLibraryEntryAtom)

    const showLibraryBadge = !!libraryData && !!props.showLibraryBadge
    const showProgressBar = (!!listData?.progress && (!isManga ? !!media?.episodes : !!(media as any)?.chapters) && listData?.status !== "COMPLETED")
    const showTrailer = _showTrailer && !libraryData && !media?.isAdult // Show trailer only if libraryData is not available

    const link = !isManga ? `/entry?id=${media.id}` : `/manga/entry?id=${media.id}`

    // For pages where listData or libraryData is not accessible (where LibraryCollection is not fetched),
    // use cached LibraryCollection
    useEffect(() => {
        if (!_listData || !_libraryData) {
            const entry = getAtomicLibraryEntry(media.id)
            if (!_listData) {
                setListData(entry?.listData)
            }
            if (!_libraryData) {
                setLibraryData(entry?.libraryData)
            }
        }
    }, [getAtomicLibraryEntry])

    const pathname = usePathname()

    // Dynamically refresh data when LibraryCollection is updated
    useEffect(() => {
        if (pathname !== "/") {
            const entry = getAtomicLibraryEntry(media.id)
            if (!_listData) {
                setListData(entry?.listData)
            }
            if (!_libraryData) {
                setLibraryData(entry?.libraryData)
            }
        }
    }, [pathname, __atomicLibraryCollection])

    useLayoutEffect(() => {
        setListData(_listData)
    }, [_listData])

    useLayoutEffect(() => {
        setLibraryData(_libraryData)
    }, [_libraryData])

    if (!media) return null

    return (
        <div
            className={cn(
                "h-full col-span-1 group/anime-list-item relative flex flex-col place-content-stretch focus-visible:outline-0 flex-none",
                props.containerClassName,
            )}
        >

            {overlay && <div
                className={cn(
                    "absolute z-[14] top-0 left-0 w-full",
                )}
            >{overlay}</div>}

            {/*ACTION POPUP*/}
            <div
                className={cn(
                    "absolute z-[15] bg-gray-950 opacity-0 scale-70 border",
                    "group-hover/anime-list-item:opacity-100 group-hover/anime-list-item:scale-100",
                    "group-focus-visible/anime-list-item:opacity-100 group-focus-visible/anime-list-item:scale-100",
                    "focus-visible:opacity-100 focus-visible:scale-100",
                    "h-[105%] w-[100%] -top-[5%] rounded-md transition ease-in-out",
                    "focus-visible:ring-2 ring-brand-400 focus-visible:outline-0",
                    "hidden lg:block", // Hide on small screens
                )} tabIndex={0}
                onMouseEnter={() => setActionPopupHover(media.id)}
                onMouseLeave={() => setActionPopupHover(undefined)}
            >
                <div className="p-2 h-full w-full flex flex-col justify-between">
                    {/*METADATA SECTION*/}
                    <div className="space-y-1">

                        <ActionPopupImage
                            media={media}
                            listData={listData}
                            showProgressBar={showProgressBar}
                            showTrailer={showTrailer || false}
                            link={link}
                            isManga={isManga}
                        />

                        <div>
                            {/*<Tooltip trigger={*/}
                            {/*    <p className="text-center font-medium text-sm min-[2000px]:text-lg px-4 line-clamp-1">{media.title?.userPreferred}</p>*/}
                            {/*}>{media.title?.userPreferred}</Tooltip>*/}
                            <Link
                                href={link}
                                className="text-center text-pretty font-medium text-sm lg:text-base px-4 leading-0 line-clamp-2 hover:text-brand-100"
                            >
                                {media.title?.userPreferred}
                            </Link>
                        </div>
                        {!!media.startDate?.year && <div>
                            <p className="justify-center text-sm text-[--muted] flex w-full gap-1 items-center">
                                {startCase(media.format || "")} - <BiCalendarAlt /> {capitalize(media.season ?? "")} {media.startDate?.year}
                            </p>
                        </div>}

                        {!!media.nextAiringEpisode && (
                            <div className="flex gap-1 items-center justify-center">
                                <p className="text-xs min-[2000px]:text-md">Next episode:</p>
                                <Tooltip
                                    className="bg-gray-200 text-gray-800 font-semibold mb-1"
                                    trigger={
                                        <p className="text-justify font-normal text-xs min-[2000px]:text-md">
                                            <Badge
                                                size="sm"
                                            >{media.nextAiringEpisode?.episode}</Badge>
                                        </p>
                                    }
                                >{formatDistanceToNow(addSeconds(new Date(), media.nextAiringEpisode?.timeUntilAiring),
                                    { addSuffix: true })}</Tooltip>
                            </div>
                        )}

                        {!isManga && <MainActionButton media={media} listData={listData} />}
                        {isManga && <Link
                            href={`/manga/entry?id=${props.media.id}`}
                        >
                            <Button
                                leftIcon={<IoLibrarySharp />}
                                intent="white"
                                size="md"
                                className="w-full text-md mt-2"
                            >
                                Read
                            </Button>
                        </Link>}

                        {(listData?.status && props.showLibraryBadge === undefined) &&
                            <p className="text-center">{listData?.status === "CURRENT" ? "Current" : capitalize(listData?.status ?? "")}</p>}

                    </div>
                    <div className="flex gap-2">

                        {!!libraryData && <LockFilesButton mediaId={media.id} allFilesLocked={libraryData.allFilesLocked} />}

                        {showListDataButton && <AnilistMediaEntryModal listData={listData} media={media} type={!isManga ? "anime" : "manga"} />}

                        {withAudienceScore &&
                            <AnimeEntryAudienceScore
                                meanScore={media.meanScore}
                                hideAudienceScore={serverStatus?.settings?.anilist?.hideAudienceScore}
                            />}

                    </div>
                </div>
            </div>

            <Link
                href={link}
                className="w-full relative"
            >
                <div className="aspect-[6/7] flex-none rounded-md border object-cover object-center relative overflow-hidden">

                    {/*[CUSTOM UI] BOTTOM GRADIENT*/}
                    <AnimeListItemBottomGradient />

                    {showProgressBar && <div className="absolute top-0 w-full h-1 z-[2] bg-gray-700 left-0">
                        <div
                            className={cn(
                                "h-1 absolute z-[2] left-0 bg-gray-200 transition-all",
                                {
                                    "bg-brand-400": listData?.status === "CURRENT",
                                    "bg-gray-400": listData?.status !== "CURRENT",
                                },
                            )}
                            style={{
                                width: `${String(Math.ceil((listData.progress! / (!isManga ? media?.episodes : (media as any)?.chapters)!) * 100))}%`,
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
                    {media.status === "RELEASING" && <div className="absolute z-[10] right-1 top-2">
                        <Tooltip
                            trigger={<Badge intent="primary-solid" size="lg"><RiSignalTowerLine /></Badge>}
                        >
                            Airing
                        </Tooltip>
                    </div>}

                    {/*NOT YET RELEASED BADGE*/}
                    {media.status === "NOT_YET_RELEASED" && <div className="absolute z-[10] right-1 top-1">
                        {(!!media.startDate && !!media.startDate?.year) && <Tooltip
                            trigger={<Badge intent="gray-solid" size="lg"><RiSignalTowerLine /></Badge>}
                        >
                            {!!media.startDate?.year ?
                                new Intl.DateTimeFormat("en-US", {
                                    year: "numeric",
                                    month: "short",
                                    day: "numeric",
                                }).format(new Date(media.startDate.year, media.startDate?.month || 0, media.startDate?.day || 0))
                                : "-"
                            }
                        </Tooltip>}
                    </div>}


                    <ProgressBadge media={media} listData={listData} />
                    <ScoreBadge listData={listData} />

                    <Image
                        src={media.coverImage?.extraLarge || ""}
                        alt={""}
                        fill
                        placeholder={imageShimmer(700, 475)}
                        quality={100}
                        sizes="20rem"
                        className="object-cover object-center group-hover/anime-list-item:scale-125 transition"
                    />

                    {serverStatus?.settings?.anilist?.blurAdultContent && media.isAdult && <div
                        className="absolute top-0 w-full h-full backdrop-blur-xl z-[3] border-4"
                    ></div>}
                </div>
            </Link>
            <div className="pt-2 space-y-2 flex flex-col justify-between h-full">
                <div>
                    <p className="text-center font-semibold text-sm lg:text-md min-[2000px]:text-lg line-clamp-3">{media.title?.userPreferred}</p>
                </div>
                {(!!media.season || !!media.startDate?.year) && <div>
                    <p className="text-sm text-[--muted] inline-flex gap-1 items-center">
                        <BiCalendarAlt />{capitalize(media.season ?? "")} {media.startDate?.year}
                    </p>
                </div>}
            </div>

        </div>
    )
})

const ActionPopupImage = ({ media, showProgressBar, listData, showTrailer, link, isManga }: {
    media: BaseMediaFragment,
    listData: MediaEntryListData | undefined,
    showProgressBar: boolean
    showTrailer: boolean
    link: string
    isManga?: boolean
}) => {

    const serverStatus = useAtomValue(serverStatusAtom)
    const [trailerLoaded, setTrailerLoaded] = React.useState(false)
    const [actionPopupHoverId] = useAtom(actionPopupHoverAtom)
    const actionPopupHover = actionPopupHoverId === media.id
    const [trailerEnabled, setTrailerEnabled] = useState(!!media?.trailer?.id && !serverStatus?.settings?.library?.disableAnimeCardTrailers && showTrailer)

    React.useEffect(() => {
        setTrailerEnabled(!!media?.trailer?.id && !serverStatus?.settings?.library?.disableAnimeCardTrailers && showTrailer)
    }, [!!media?.trailer?.id, !serverStatus?.settings?.library?.disableAnimeCardTrailers, showTrailer])

    const Content = (
        <div className="aspect-[4/2] relative rounded-md overflow-hidden mb-2 cursor-pointer">
            {(showProgressBar && listData) && <div className="absolute top-0 w-full h-1 z-[2] bg-gray-700 left-0">
                <div
                    className={cn(
                        "h-1 absolute z-[2] left-0 bg-gray-200 transition-all",
                        {
                            "bg-brand-400": listData?.status === "CURRENT",
                            "bg-gray-400": listData?.status !== "CURRENT",
                        },
                    )}
                    style={{ width: `${String(Math.ceil((listData?.progress! / (!isManga ? media?.episodes : (media as any)?.chapters)!) * 100))}%` }}
                ></div>
            </div>}

            {(!!media.bannerImage || !!media.coverImage?.large) ? <Image
                src={media.bannerImage || media.coverImage?.large || ""}
                alt={""}
                fill
                placeholder={imageShimmer(700, 475)}
                quality={100}
                sizes="20rem"
                className={cn(
                    "object-cover object-center transition",
                    trailerLoaded && "hidden",
                )}
            /> : <div
                className="h-full block absolute w-full bg-gradient-to-t from-gray-800 to-transparent"
            ></div>}

            {serverStatus?.settings?.anilist?.blurAdultContent && media.isAdult && <div
                className="absolute top-0 w-full h-full backdrop-blur-xl z-[3] border-2"
            ></div>}

            {(trailerEnabled && actionPopupHover) && <video
                src={`https://yewtu.be/latest_version?id=${media?.trailer?.id}&itag=18`}
                className={cn(
                    "aspect-video w-full absolute left-0",
                    !trailerLoaded && "hidden",
                )}
                playsInline
                preload="none"
                loop
                autoPlay
                muted
                onLoadedData={() => setTrailerLoaded(true)}
                onError={() => setTrailerEnabled(false)}
            />}

            <div
                className="w-full absolute bottom-0 h-[4rem] bg-gradient-to-t from-gray-950 to-transparent z-[2]"
            />
        </div>
    )

    if (!trailerEnabled) {
        return <Link href={link}>{Content}</Link>
    } else {
        return (
            <TrailerModal
                trailerId={media.trailer?.id}
                trigger={Content}
            />
        )
    }

}

const MainActionButton = (props: { media: BaseMediaFragment, listData?: MediaEntryListData }) => {
    const progress = props.listData?.progress
    const status = props.listData?.status
    return (
        <>
            <div>
                <div className="py-1">
                    <Link
                        href={`/entry?id=${props.media.id}${(!!progress && (status !== "COMPLETED")) ? "&playNext=true" : ""}`}
                    >
                        <Button
                            leftIcon={<BiPlay className="text-2xl" />}
                            intent="white"
                            size="md"
                            className="w-full text-md"
                        >
                            {!!progress && (status === "CURRENT" || status === "PAUSED") ? "Continue watching" : "Watch"}
                        </Button>
                    </Link>
                </div>
            </div>
        </>
    )
}

const LockFilesButton = memo(({ mediaId, allFilesLocked }: { mediaId: number, allFilesLocked: boolean }) => {

    const { toggleLock, isPending } = useMediaEntryBulkAction()

    return (
        <Tooltip
            trigger={
                <IconButton
                    icon={allFilesLocked ? <VscVerified /> : <BiLockOpenAlt />}
                    intent={allFilesLocked ? "success" : "warning-subtle"}
                    size="sm"
                    className="hover:opacity-60"
                    loading={isPending}
                    onClick={() => toggleLock(mediaId)}
                />
            }
        >
            {allFilesLocked ? "Unlock all files" : "Lock all files"}
        </Tooltip>
    )
})

const ScoreBadge = (props: { listData?: MediaEntryListData }) => {

    const score = props.listData?.score

    if (!props.listData || !score) return null

    const scoreColor = score ? (
        score < 5 ? "bg-red-500" :
            score < 7 ? "bg-orange-500" :
                score < 9 ? "bg-green-500" :
                    "bg-brand-500 text-white bg-opacity-80"
    ) : ""

    return (
        <div className="absolute z-[10] right-1 bottom-1">
            <div
                className={cn(
                    "backdrop-blur-lg inline-flex items-center justify-center gap-1 w-14 h-7 rounded-full font-bold bg-opacity-70 drop-shadow-sm shadow-lg",
                    scoreColor,
                )}
            >
                <BiStar /> {(score === 0) ? "-" : score}
            </div>
        </div>
    )
}

const ProgressBadge = (props: { media: BaseMediaFragment, listData?: MediaEntryListData }) => {

    const progress = props.listData?.progress
    const episodes = props.media.episodes || (props.media as any)?.chapters

    if (!props.listData || !progress) return null

    return (
        <div className="absolute z-[10] left-1 bottom-1">
            <Badge size="lg" className="rounded-md px-1.5">
                {progress}{!!episodes ? `/${episodes}` : ""}
            </Badge>
        </div>
    )
}
