import { getAtomicLibraryEntryAtom } from "@/atoms/collection"
import { AnilistMediaEntryModal } from "@/components/shared/anilist-media-entry-modal"
import { imageShimmer } from "@/components/shared/image-helpers"
import { Badge } from "@/components/ui/badge"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core"
import { Tooltip } from "@/components/ui/tooltip"
import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import { useMediaEntryBulkAction } from "@/lib/server/hooks/library"
import { MediaEntryLibraryData, MediaEntryListData } from "@/lib/server/types"
import { BiCalendarAlt } from "@react-icons/all-files/bi/BiCalendarAlt"
import { BiLockOpenAlt } from "@react-icons/all-files/bi/BiLockOpenAlt"
import { BiPlay } from "@react-icons/all-files/bi/BiPlay"
import { BiStar } from "@react-icons/all-files/bi/BiStar"
import { IoLibrarySharp } from "@react-icons/all-files/io5/IoLibrarySharp"
import { RiSignalTowerLine } from "@react-icons/all-files/ri/RiSignalTowerLine"
import { VscVerified } from "@react-icons/all-files/vsc/VscVerified"
import { addSeconds, formatDistanceToNow } from "date-fns"
import { useAtom } from "jotai"
import capitalize from "lodash/capitalize"
import startCase from "lodash/startCase"
import Image from "next/image"
import Link from "next/link"
import { usePathname } from "next/navigation"
import React, { memo, useEffect, useLayoutEffect, useState } from "react"

type AnimeListItemProps = {
    media: BaseMediaFragment,
    listData?: MediaEntryListData
    libraryData?: MediaEntryLibraryData
    showLibraryBadge?: boolean
} & {
    containerClassName?: string
}

export const AnimeListItem = ((props: AnimeListItemProps) => {

    const { media, listData: _listData, libraryData: _libraryData } = props

    const [listData, setListData] = useState(_listData)
    const [libraryData, setLibraryData] = useState(_libraryData)

    const [__atomicLibraryCollection, getAtomicLibraryEntry] = useAtom(getAtomicLibraryEntryAtom)

    const showLibraryBadge = !!libraryData && !!props.showLibraryBadge
    const showProgressBar = (!!listData?.progress && media?.episodes && listData?.status !== "COMPLETED")

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

            {/*ACTION POPUP*/}
            <div className={cn(
                "absolute z-20 bg-gray-900 opacity-0 scale-70 border border-[--border]",
                "group-hover/anime-list-item:opacity-100 group-hover/anime-list-item:scale-100",
                "group-focus-visible/anime-list-item:opacity-100 group-focus-visible/anime-list-item:scale-100",
                "focus-visible:opacity-100 focus-visible:scale-100",
                "h-[105%] w-[100%] -top-[5%] rounded-md transition ease-in-out",
                "focus-visible:ring-2 ring-brand-400 focus-visible:outline-0 ",
            )} tabIndex={0}>
                <div className={"p-2 h-full w-full flex flex-col justify-between"}>
                    {/*METADATA SECTION*/}
                    <div className={"space-y-1"}>
                        <div className={"aspect-[4/2] relative rounded-md overflow-hidden mb-2"}>
                            {showProgressBar && <div className={"absolute top-0 w-full h-1 z-[2] bg-gray-700 left-0"}>
                                <div
                                    className={cn(
                                        "h-1 absolute z-[2] left-0 bg-gray-200 transition-all",
                                        {
                                            "bg-brand-400": listData?.status === "CURRENT",
                                            "bg-gray-400": listData?.status !== "CURRENT",
                                        },
                                    )}
                                    style={{ width: `${String(Math.ceil((listData.progress! / media.episodes!) * 100))}%` }}
                                ></div>
                            </div>}

                            {(!!media.bannerImage || !!media.coverImage?.large) ? <Image
                                src={media.bannerImage || media.coverImage?.large || ""}
                                alt={""}
                                fill
                                placeholder={imageShimmer(700, 475)}
                                quality={100}
                                sizes="20rem"
                                className="object-cover object-center transition"
                            /> : <div
                                className={"h-full block absolute w-full bg-gradient-to-t from-gray-800 to-transparent"}></div>}
                        </div>
                        <div>
                            {/*<Tooltip trigger={*/}
                            {/*    <p className={"text-center font-medium text-sm min-[2000px]:text-lg px-4 truncate text-ellipsis"}>{media.title?.userPreferred}</p>*/}
                            {/*}>{media.title?.userPreferred}</Tooltip>*/}
                            <Link
                                href={`/entry?id=${media.id}`}
                                className={"text-center font-medium text-sm lg:text-lg min-[2000px]:text-lg px-4 line-clamp-2"}
                            >{media.title?.userPreferred}</Link>
                        </div>
                        {!!media.startDate?.year && <div>
                            <p className={"justify-center text-sm text-[--muted] flex w-full gap-1 items-center"}>
                                {startCase(media.format || "")} - <BiCalendarAlt/> {new Intl.DateTimeFormat("en-US", {
                                year: "numeric",
                                month: "short",
                            }).format(new Date(media.startDate?.year || 0, media.startDate?.month || 0))} - {capitalize(media.season ?? "")}
                            </p>
                        </div>}

                        {!!media.nextAiringEpisode && (
                            <div className={"flex gap-1 items-center justify-center"}>
                                <p className={"text-xs min-[2000px]:text-md"}>Next episode:</p>
                                <Tooltip
                                    tooltipClassName={"bg-gray-200 text-gray-800 font-semibold mb-1"}
                                    trigger={
                                        <p className={"text-justify font-normal text-xs min-[2000px]:text-md"}>
                                            <Badge
                                                size={"sm"}>{media.nextAiringEpisode?.episode}</Badge>
                                        </p>
                                    }>{formatDistanceToNow(addSeconds(new Date(), media.nextAiringEpisode?.timeUntilAiring), { addSuffix: true })}</Tooltip>
                            </div>
                        )}

                        <MainActionButton media={media} listData={listData}/>

                        {(listData?.status && props.showLibraryBadge === undefined) &&
                            <p className={"text-center"}>{listData?.status === "CURRENT" ? "Watching" : capitalize(listData?.status ?? "")}</p>}

                    </div>
                    <div className={"flex gap-2"}>
                        {!!libraryData &&
                            <LockFilesButton mediaId={media.id} allFilesLocked={libraryData.allFilesLocked}/>}
                        <AnilistMediaEntryModal listData={listData} media={media}/>
                    </div>
                </div>
            </div>

            <div
                className="aspect-[6/7] flex-none rounded-md border border-[--border] object-cover object-center relative overflow-hidden"
            >

                {/*BOTTOM GRADIENT*/}
                <div
                    className={"z-[5] absolute bottom-0 w-full h-[50%] bg-gradient-to-t from-black to-transparent"}
                />

                {showProgressBar && <div className={"absolute top-0 w-full h-1 z-[2] bg-gray-700 left-0"}>
                    <div className={cn(
                        "h-1 absolute z-[2] left-0 bg-gray-200 transition-all",
                        {
                            "bg-brand-400": listData?.status === "CURRENT",
                            "bg-gray-400": listData?.status !== "CURRENT",
                        },
                    )}
                         style={{ width: `${String(Math.ceil((listData.progress! / media.episodes!) * 100))}%` }}></div>
                </div>}

                {(showLibraryBadge) &&
                    <div className={"absolute z-[1] left-0 top-0"}>
                        <Badge size={"xl"} intent={"warning-solid"}
                               className={"rounded-md rounded-bl-none rounded-tr-none text-orange-900"}><IoLibrarySharp/></Badge>
                    </div>}

                {/*RELEASING BADGE*/}
                {media.status === "RELEASING" && <div className={"absolute z-10 right-1 top-2"}>
                    <Tooltip
                        trigger={<Badge intent={"primary-solid"} size={"lg"}><RiSignalTowerLine/></Badge>}>
                        Airing
                    </Tooltip>
                </div>}

                {/*NOT YET RELEASED BADGE*/}
                {media.status === "NOT_YET_RELEASED" && <div className={"absolute z-10 right-1 top-1"}>
                    <Tooltip
                        trigger={<Badge intent={"gray-solid"} size={"lg"}><RiSignalTowerLine/></Badge>}>
                        {!!media.startDate?.year ?
                            new Intl.DateTimeFormat("en-US", {
                                year: "numeric",
                                month: "short",
                                day: "numeric",
                            }).format(new Date(media.startDate.year, media.startDate?.month || 0, media.startDate?.day || 0))
                            : "-"
                        }
                    </Tooltip>
                </div>}


                <ProgressBadge media={media} listData={listData}/>
                <ScoreBadge listData={listData}/>

                <Image
                    src={media.coverImage?.extraLarge || ""}
                    alt={""}
                    fill
                    placeholder={imageShimmer(700, 475)}
                    quality={100}
                    priority
                    sizes="20rem"
                    className="object-cover object-center group-hover/anime-list-item:scale-125 transition"
                />
            </div>
            <div className={"pt-2 space-y-2 flex flex-col justify-between h-full"}>
                <div>
                    <p className={"text-center font-semibold text-sm lg:text-md min-[2000px]:text-lg line-clamp-3"}>{media.title?.userPreferred}</p>
                </div>
                <div>
                    <div>
                        <p className={"text-sm text-[--muted] inline-flex gap-1 items-center"}>
                            <BiCalendarAlt/>{capitalize(media.season ?? "")} {media.startDate?.year ? new Intl.DateTimeFormat("en-US", {
                            year: "numeric",
                        }).format(new Date(media.startDate?.year || 0, media.startDate?.month || 0)) : "-"}
                        </p>
                    </div>
                </div>
            </div>

        </div>
    )
})

const MainActionButton = (props: { media: BaseMediaFragment, listData?: MediaEntryListData }) => {
    const progress = props.listData?.progress
    const status = props.listData?.status
    return (
        <>
            <div>
                <div className={"py-1"}>
                    <Link
                        href={`/entry?id=${props.media.id}${(!!progress && (status !== "COMPLETED")) ? "&playNext=true" : ""}`}>
                        <Button
                            leftIcon={<BiPlay/>}
                            intent={"white"}
                            size={"md"}
                            className={"w-full text-md"}
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
        <Tooltip trigger={
            <IconButton
                icon={allFilesLocked ? <VscVerified/> : <BiLockOpenAlt/>}
                intent={allFilesLocked ? "success" : "warning-subtle"}
                size={"sm"}
                className={"hover:opacity-60"}
                isLoading={isPending}
                onClick={() => toggleLock(mediaId)}
            />
        }>
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
        <div className={"absolute z-10 right-1 bottom-1"}>
            <div className={cn(
                "backdrop-blur-lg inline-flex items-center justify-center gap-1 w-12 h-7 rounded-full font-bold bg-opacity-70 drop-shadow-sm shadow-lg",
                scoreColor,
            )}>
                <BiStar/> {(score === 0) ? "-" : score}
            </div>
        </div>
    )
}

const ProgressBadge = (props: { media: BaseMediaFragment, listData?: MediaEntryListData }) => {

    const progress = props.listData?.progress
    const episodes = props.media.episodes

    if (!props.listData || !progress) return null

    return (
        <div className={"absolute z-10 left-1 bottom-1"}>
            <Badge size={"lg"}>
                {progress}/{episodes ?? "-"}
            </Badge>
        </div>
    )
}
