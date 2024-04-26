import { AL_BaseManga, AL_BaseMedia, Anime_MediaEntryLibraryData, Anime_MediaEntryListData, Manga_EntryListData } from "@/api/generated/types"
import { getAtomicLibraryEntryAtom } from "@/app/(main)/_atoms/anime-library-collection.atoms"
import { ToggleLockFilesButton } from "@/app/(main)/_features/anime/_containers/toggle-lock-files-button"
import {
    __mediaEntryCard_hoveredPopupId,
    MediaEntryCardBody,
    MediaEntryCardContainer,
    MediaEntryCardHoverPopup,
    MediaEntryCardHoverPopupBanner,
    MediaEntryCardHoverPopupBody,
    MediaEntryCardHoverPopupFooter,
    MediaEntryCardHoverPopupTitleSection,
    MediaEntryCardNextAiring,
    MediaEntryCardOverlay,
    MediaEntryCardTitleSection,
} from "@/app/(main)/_features/media/_components/media-entry-card-components"
import { MediaEntryAudienceScore } from "@/app/(main)/_features/media/_components/media-entry-metadata-components"
import { MediaEntryProgressBadge } from "@/app/(main)/_features/media/_components/media-entry-progress-badge"
import { MediaEntryScoreBadge } from "@/app/(main)/_features/media/_components/media-entry-score-badge"
import { AnilistMediaEntryModal } from "@/app/(main)/_features/media/_containers/anilist-media-entry-modal"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { Button } from "@/components/ui/button"
import { useAtom } from "jotai"
import { useSetAtom } from "jotai/react"
import capitalize from "lodash/capitalize"
import Link from "next/link"
import { usePathname } from "next/navigation"
import React, { useState } from "react"
import { BiPlay } from "react-icons/bi"
import { IoLibrarySharp } from "react-icons/io5"

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
    listData?: T extends "anime" ? Anime_MediaEntryListData : T extends "manga" ? Manga_EntryListData : never
    showLibraryBadge?: T extends "anime" ? boolean : never
    showTrailer?: T extends "anime" ? boolean : never
    libraryData?: T extends "anime" ? Anime_MediaEntryLibraryData : never
} & MediaEntryCardBaseProps

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
    const setActionPopupHover = useSetAtom(__mediaEntryCard_hoveredPopupId)

    const [__atomicLibraryCollection, getAtomicLibraryEntry] = useAtom(getAtomicLibraryEntryAtom)

    const showLibraryBadge = !!libraryData && !!props.showLibraryBadge

    const showProgressBar = React.useMemo(() => {
        return !!listData?.progress
        && type === "anime" ? !!(media as AL_BaseMedia)?.episodes : !!(media as AL_BaseManga)?.chapters
            && listData?.status !== "COMPLETED"
    }, [listData?.progress, media, listData?.status])

    const showTrailer = React.useMemo(() => _showTrailer && !libraryData && !media?.isAdult, [_showTrailer, libraryData, media])

    const link = type === "anime" ? `/entry?id=${media.id}` : `/manga/entry?id=${media.id}`

    const progressTotal = type === "anime" ? (media as AL_BaseMedia)?.episodes : (media as AL_BaseManga)?.chapters

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

                    <MediaEntryCardHoverPopupBanner
                        trailerId={(media as any)?.trailer?.id}
                        showProgressBar={showProgressBar}
                        mediaId={media.id}
                        progress={listData?.progress}
                        progressTotal={progressTotal}
                        showTrailer={showTrailer}
                        disableAnimeCardTrailers={serverStatus?.settings?.library?.disableAnimeCardTrailers}
                        bannerImage={media.bannerImage || media.coverImage?.extraLarge}
                        isAdult={media.isAdult}
                        blurAdultContent={serverStatus?.settings?.anilist?.blurAdultContent}
                        link={link}
                        listStatus={listData?.status}
                        status={media.status}
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

                    {type === "anime" && <div className="py-1">
                        <Link
                            href={`/entry?id=${media.id}${(!!listData?.progress && (listData?.status !== "COMPLETED")) ? "&playNext=true" : ""}`}
                        >
                            <Button
                                leftIcon={<BiPlay className="text-2xl" />}
                                intent="white"
                                size="md"
                                className="w-full text-md"
                            >
                                {!!listData?.progress && (listData?.status === "CURRENT" || listData?.status === "PAUSED")
                                    ? "Continue watching"
                                    : "Watch"}
                            </Button>
                        </Link>
                    </div>}

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

                    {(listData?.status) &&
                        <p className="text-center text-sm text-[--muted]">
                            {listData?.status === "CURRENT" ? type === "anime" ? "Watching" : "Reading"
                                : capitalize(listData?.status ?? "")}
                        </p>}

                </MediaEntryCardHoverPopupBody>

                <MediaEntryCardHoverPopupFooter>

                    {(type === "anime" && !!libraryData) && <ToggleLockFilesButton mediaId={media.id} allFilesLocked={libraryData.allFilesLocked} />}

                    {showListDataButton && <AnilistMediaEntryModal listData={listData} media={media} type={type} />}

                    {withAudienceScore &&
                        <MediaEntryAudienceScore
                            meanScore={media.meanScore}
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
                progressTotal={progressTotal}
                startDate={media.startDate}
                bannerImage={media.coverImage?.extraLarge || ""}
                isAdult={media.isAdult}
                showLibraryBadge={showLibraryBadge}
                blurAdultContent={serverStatus?.settings?.anilist?.blurAdultContent}
            >
                <div className="absolute z-[10] right-1 bottom-1">
                    <MediaEntryScoreBadge
                        score={listData?.score}
                    />
                </div>
                <div className="absolute z-[10] left-1 bottom-1">
                    <MediaEntryProgressBadge
                        progress={listData?.progress}
                        progressTotal={progressTotal}
                    />
                </div>
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


