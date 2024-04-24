import { AL_BaseManga, AL_BaseMedia, AL_MediaListStatus, Anime_MediaEntryLibraryData, Anime_MediaEntryListData } from "@/api/generated/types"
import { useAnimeEntryBulkAction } from "@/api/hooks/anime_entries.hooks"
import { MediaEntryListData } from "@/app/(main)/(library)/_lib/anime-library.types"
import { getAtomicLibraryEntryAtom } from "@/app/(main)/_atoms/anime-library-collection.atoms"
import {
    MediaEntryCardBody,
    MediaEntryCardContainer,
    MediaEntryCardHoverPopup,
    MediaEntryCardHoverPopupBody,
    MediaEntryCardHoverPopupFooter,
    MediaEntryCardHoverPopupTitleSection,
    MediaEntryCardNextAiring,
    MediaEntryCardOverlay,
    MediaEntryCardTitleSection,
} from "@/app/(main)/_components/features/media/media-entry-card-components"
import { useServerStatus } from "@/app/(main)/_hooks/server-status.hooks"
import { AnimeEntryAudienceScore } from "@/app/(main)/entry/_containers/meta-section/_components/anime-entry-metadata-components"
import { AnilistMediaEntryModal } from "@/components/shared/anilist-media-entry-modal"
import { imageShimmer } from "@/components/shared/styling/image-helpers"
import { TrailerModal } from "@/components/shared/trailer-modal"
import { Badge } from "@/components/ui/badge"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Tooltip } from "@/components/ui/tooltip"
import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import { atom, useAtom } from "jotai"
import { useSetAtom } from "jotai/react"
import capitalize from "lodash/capitalize"
import Image from "next/image"
import Link from "next/link"
import { usePathname } from "next/navigation"
import React, { memo, useState } from "react"
import { BiLockOpenAlt, BiPlay, BiStar } from "react-icons/bi"
import { IoLibrarySharp } from "react-icons/io5"
import { VscVerified } from "react-icons/vsc"

type MediaEntryCardBaseProps = {
    overlay?: React.ReactNode
    withAudienceScore?: boolean
    containerClassName?: string
    showListDataButton?: boolean
}

type MediaEntryCardProps<T extends "anime" | "manga"> = {
    type: T
    media: T extends "anime" ? AL_BaseMedia : T extends "manga" ? AL_BaseManga : never
    // Anime-only
    listData?: T extends "anime" ? Anime_MediaEntryListData : never
    showLibraryBadge?: T extends "anime" ? boolean : never
    showTrailer?: T extends "anime" ? boolean : never
    libraryData?: T extends "anime" ? Anime_MediaEntryLibraryData : never
} & MediaEntryCardBaseProps

const actionPopupHoverAtom = atom<number | undefined>(undefined)

export function MediaEntryCard<T extends "anime" | "manga">(props: MediaEntryCardProps<T>) {

    const serverStatus = useServerStatus()
    const {
        media,
        listData: _listData,
        libraryData: _libraryData,
        overlay,
        showListDataButton,
        showTrailer: _showTrailer,
        type,
        withAudienceScore = true,
    } = props

    const [listData, setListData] = useState<Anime_MediaEntryListData | undefined>(_listData)
    const [libraryData, setLibraryData] = useState<Anime_MediaEntryLibraryData | undefined>(_libraryData)
    const setActionPopupHover = useSetAtom(actionPopupHoverAtom)

    const [__atomicLibraryCollection, getAtomicLibraryEntry] = useAtom(getAtomicLibraryEntryAtom)

    const showLibraryBadge = !!libraryData && !!props.showLibraryBadge

    const showProgressBar = React.useMemo(() => {
        return !!listData?.progress
        && type === "anime" ? !!(media as AL_BaseMedia)?.episodes : !!(media as AL_BaseManga)?.chapters
            && listData?.status !== "COMPLETED"
    }, [listData?.progress, media, listData?.status])

    const showTrailer = React.useMemo(() => _showTrailer && !libraryData && !media?.isAdult, [_showTrailer, libraryData, media])

    const link = type === "anime" ? `/entry?id=${media.id}` : `/manga/entry?id=${media.id}`

    // For pages where listData or libraryData is not accessible (where LibraryCollection is not fetched),
    // use cached LibraryCollection
    React.useEffect(() => {
        if (type === "anime" && !_listData || !_libraryData) {
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
    React.useEffect(() => {
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

    React.useLayoutEffect(() => {
        setListData(_listData)
    }, [_listData])

    React.useLayoutEffect(() => {
        setLibraryData(_libraryData)
    }, [_libraryData])

    if (!media) return null

    return (
        <MediaEntryCardContainer className={props.containerClassName}>

            <MediaEntryCardOverlay overlay={overlay} />

            {/*ACTION POPUP*/}
            <MediaEntryCardHoverPopup
                onMouseEnter={() => setActionPopupHover(media.id)}
                onMouseLeave={() => setActionPopupHover(undefined)}
            >

                {/*METADATA SECTION*/}
                <MediaEntryCardHoverPopupBody>

                    <ActionPopupImage
                        trailerId={(media as any)?.trailer?.id}
                        showProgressBar={showProgressBar}
                        mediaId={media.id}
                        progress={listData?.progress}
                        progressTotal={type === "anime" ? (media as AL_BaseMedia)?.episodes : (media as AL_BaseManga)?.chapters}
                        showTrailer={showTrailer}
                        disableAnimeCardTrailers={serverStatus?.settings?.library?.disableAnimeCardTrailers}
                        bannerImage={media.bannerImage || media.coverImage?.extraLarge}
                        isAdult={media.isAdult}
                        blurAdultContent={serverStatus?.settings?.anilist?.blurAdultContent}
                        link={link}
                        status={listData?.status}
                    />

                    <MediaEntryCardHoverPopupTitleSection
                        title={media.title?.userPreferred || ""}
                        year={media.startDate?.year}
                        season={media.season}
                        format={media.format}
                        link={link}
                    />


                    {type === "anime" && (
                        <MediaEntryCardNextAiring nextAiring={(media as AL_BaseMedia).nextAiringEpisode} />
                    )}

                    {type === "anime" && <MainActionButton media={media} listData={listData} />}

                    {type === "manga" && <Link
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

                </MediaEntryCardHoverPopupBody>

                <MediaEntryCardHoverPopupFooter>

                    {!!libraryData && <LockFilesButton mediaId={media.id} allFilesLocked={libraryData.allFilesLocked} />}

                    {showListDataButton && <AnilistMediaEntryModal listData={listData} media={media} type={type} />}

                    {withAudienceScore &&
                        <AnimeEntryAudienceScore
                            meanScore={media.meanScore}
                            hideAudienceScore={serverStatus?.settings?.anilist?.hideAudienceScore}
                        />}

                </MediaEntryCardHoverPopupFooter>
            </MediaEntryCardHoverPopup>


            <MediaEntryCardBody
                link={link}
                type={type}
                title={media.title?.userPreferred || ""}
                season={media.season}
                listStatus={listData?.status}
                status={media.status}
                showProgressBar={showProgressBar}
                progress={listData?.progress}
                progressTotal={type === "anime" ? (media as AL_BaseMedia)?.episodes : (media as AL_BaseManga)?.chapters}
                startDate={media.startDate}
                bannerImage={media.coverImage?.extraLarge || ""}
                isAdult={media.isAdult}
                showLibraryBadge={showLibraryBadge}
                blurAdultContent={serverStatus?.settings?.anilist?.blurAdultContent}
            >
                <ProgressBadge media={media} listData={listData} />
                <ScoreBadge listData={listData} />
            </MediaEntryCardBody>

            <MediaEntryCardTitleSection
                title={media.title?.userPreferred || ""}
                year={media.startDate?.year}
                season={media.season}
                format={media.format}
            />

        </MediaEntryCardContainer>
    )
}


const ActionPopupImage = ({
    trailerId,
    showProgressBar,
    mediaId,
    progress,
    progressTotal,
    showTrailer,
    disableAnimeCardTrailers,
    bannerImage,
    isAdult,
    blurAdultContent,
    link,
    status,
}: {
    mediaId: number
    trailerId?: string
    progress?: number
    progressTotal?: number
    bannerImage?: string
    showProgressBar: boolean
    showTrailer?: boolean
    link: string
    disableAnimeCardTrailers?: boolean
    blurAdultContent?: boolean
    isAdult?: boolean
    status?: AL_MediaListStatus
}) => {

    const [trailerLoaded, setTrailerLoaded] = React.useState(false)
    const [actionPopupHoverId] = useAtom(actionPopupHoverAtom)
    const actionPopupHover = actionPopupHoverId === mediaId
    const [trailerEnabled, setTrailerEnabled] = React.useState(!!trailerId && !disableAnimeCardTrailers && showTrailer)

    React.useEffect(() => {
        setTrailerEnabled(!!trailerId && !disableAnimeCardTrailers && showTrailer)
    }, [!!trailerId, !disableAnimeCardTrailers, showTrailer])

    const Content = (
        <div className="aspect-[4/2] relative rounded-md overflow-hidden mb-2 cursor-pointer">
            {(showProgressBar && progress && status && progressTotal) && <div className="absolute top-0 w-full h-1 z-[2] bg-gray-700 left-0">
                <div
                    className={cn(
                        "h-1 absolute z-[2] left-0 bg-gray-200 transition-all",
                        {
                            "bg-brand-400": status === "CURRENT",
                            "bg-gray-400": status !== "CURRENT",
                        },
                    )}
                    style={{ width: `${String(Math.ceil((progress / progressTotal) * 100))}%` }}
                ></div>
            </div>}

            {(!!bannerImage) ? <Image
                src={bannerImage || ""}
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

            {(blurAdultContent && isAdult) && <div
                className="absolute top-0 w-full h-full backdrop-blur-xl z-[3] border-2"
            ></div>}

            {(trailerEnabled && actionPopupHover) && <video
                src={`https://yewtu.be/latest_version?id=${trailerId}&itag=18`}
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
                trailerId={trailerId}
                trigger={Content}
            />
        )
    }

}

const MainActionButton = (props: { media: AL_BaseMedia, listData?: Anime_MediaEntryListData }) => {
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

    const { mutate: performBulkAction, isPending } = useAnimeEntryBulkAction(mediaId)

    return (
        <Tooltip
            trigger={
                <IconButton
                    icon={allFilesLocked ? <VscVerified /> : <BiLockOpenAlt />}
                    intent={allFilesLocked ? "success" : "warning-subtle"}
                    size="sm"
                    className="hover:opacity-60"
                    loading={isPending}
                    onClick={() => performBulkAction({
                        mediaId,
                        action: allFilesLocked ? "unlock" : "lock",
                    })}
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
