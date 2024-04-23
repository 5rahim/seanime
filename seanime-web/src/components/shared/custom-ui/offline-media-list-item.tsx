import { useMediaEntryBulkAction } from "@/app/(main)/(library)/_containers/bulk-actions/_lib/media-entry-bulk-actions"
import { OfflineAnilistMediaEntryModal } from "@/app/(main)/(offline)/offline/_components/offline-anilist-media-entry-modal"
import { OfflineAssetMap, OfflineListData } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.types"
import { offline_getAssetUrl } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.utils"
import { serverStatusAtom } from "@/app/(main)/_atoms/server-status"
import { AnimeEntryAudienceScore } from "@/app/(main)/entry/_containers/meta-section/_components/anime-entry-metadata-components"
import { AnimeListItemBottomGradient } from "@/components/shared/custom-ui/item-bottom-gradients"
import { imageShimmer } from "@/components/shared/styling/image-helpers"
import { Badge } from "@/components/ui/badge"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Tooltip } from "@/components/ui/tooltip"
import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import { useAtomValue } from "jotai/react"
import capitalize from "lodash/capitalize"
import startCase from "lodash/startCase"
import Image from "next/image"
import Link from "next/link"
import React, { memo } from "react"
import { BiCalendarAlt, BiLockOpenAlt, BiStar } from "react-icons/bi"
import { IoLibrarySharp } from "react-icons/io5"
import { VscVerified } from "react-icons/vsc"

type OfflineMediaListItemProps = {
    media: BaseMediaFragment,
    listData: OfflineListData | undefined
    overlay?: React.ReactNode
    isManga?: boolean
    withAudienceScore?: boolean
    assetMap: OfflineAssetMap | undefined
} & {
    containerClassName?: string
}

export const OfflineMediaListAtom = ((props: OfflineMediaListItemProps) => {

    const serverStatus = useAtomValue(serverStatusAtom)
    const {
        media,
        listData,
        overlay,
        isManga,
        withAudienceScore = true,
        assetMap,
    } = props


    const showLibraryBadge = false
    const showProgressBar = (!!listData?.progress && (isManga ? !!media?.episodes : !!(media as any)?.chapters) && listData?.status !== "COMPLETED")

    const link = !isManga ? `/offline/anime?id=${media.id}` : `/offline/manga?id=${media.id}`

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
            >
                <div className="p-2 h-full w-full flex flex-col justify-between">
                    {/*METADATA SECTION*/}
                    <div className="space-y-1">

                        <Link
                            href={link}
                        >
                            <div className="aspect-[4/2] relative rounded-md overflow-hidden mb-2 cursor-pointer">
                                {showProgressBar && <div className="absolute top-0 w-full h-1 z-[2] bg-gray-700 left-0">
                                    <div
                                        className={cn(
                                            "h-1 absolute z-[2] left-0 bg-gray-200 transition-all",
                                            {
                                                "bg-brand-400": listData?.status === "CURRENT",
                                                "bg-gray-400": listData?.status !== "CURRENT",
                                            },
                                        )}
                                        style={{ width: `${String(Math.ceil(((listData?.progress || 0) / (media.episodes || 1)) * 100))}%` }}
                                    ></div>
                                </div>}

                                {(!!media.bannerImage || !!media.coverImage?.large) ? <Image
                                    src={offline_getAssetUrl(media.bannerImage, assetMap) || offline_getAssetUrl(media.coverImage?.large,
                                        assetMap) || ""}
                                    alt={""}
                                    fill
                                    placeholder={imageShimmer(700, 475)}
                                    quality={100}
                                    sizes="20rem"
                                    className={cn(
                                        "object-cover object-center transition",
                                    )}
                                /> : <div
                                    className="h-full block absolute w-full bg-gradient-to-t from-gray-800 to-transparent"
                                ></div>}

                                {serverStatus?.settings?.anilist?.blurAdultContent && media.isAdult && <div
                                    className="absolute top-0 w-full h-full backdrop-blur-xl z-[3] border-2"
                                ></div>}

                                <div
                                    className="w-full absolute bottom-0 h-[4rem] bg-gradient-to-t from-gray-950 to-transparent z-[2]"
                                />
                            </div>
                        </Link>

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

                        <p className="text-center">{capitalize(listData?.status ?? "")}</p>

                    </div>
                    <div className="flex gap-2">

                        <OfflineAnilistMediaEntryModal
                            listData={listData}
                            assetMap={assetMap}
                            media={media}
                            type={!isManga ? "anime" : "manga"}
                        />

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
                                width: `${String(Math.ceil((listData.progress! / (isManga
                                    ? media?.episodes
                                    : (media as any)?.chapters)!) * 100))}%`,
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

                    <ProgressBadge media={media} listData={listData} />
                    <ScoreBadge listData={listData} />

                    <Image
                        src={offline_getAssetUrl(media.coverImage?.extraLarge, assetMap) || ""}
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

const ScoreBadge = (props: { listData?: OfflineListData }) => {

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

const ProgressBadge = (props: { media: BaseMediaFragment, listData?: OfflineListData }) => {

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
