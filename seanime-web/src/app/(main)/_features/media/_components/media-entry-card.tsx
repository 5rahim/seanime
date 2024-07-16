import { AL_BaseAnime, AL_BaseManga, Anime_AnimeEntryLibraryData, Anime_AnimeEntryListData, Manga_EntryListData } from "@/api/generated/types"
import { getAtomicLibraryEntryAtom } from "@/app/(main)/_atoms/anime-library-collection.atoms"
import { ToggleLockFilesButton } from "@/app/(main)/_features/anime/_containers/toggle-lock-files-button"
import {
    __animeEntryCard_hoveredPopupId,
    AnimeEntryCardBody,
    AnimeEntryCardContainer,
    AnimeEntryCardHoverPopup,
    AnimeEntryCardHoverPopupBanner,
    AnimeEntryCardHoverPopupBody,
    AnimeEntryCardHoverPopupFooter,
    AnimeEntryCardHoverPopupTitleSection,
    AnimeEntryCardNextAiring,
    AnimeEntryCardOverlay,
    AnimeEntryCardTitleSection,
} from "@/app/(main)/_features/media/_components/media-entry-card-components"
import { AnimeEntryAudienceScore } from "@/app/(main)/_features/media/_components/media-entry-metadata-components"
import { AnimeEntryProgressBadge } from "@/app/(main)/_features/media/_components/media-entry-progress-badge"
import { AnimeEntryScoreBadge } from "@/app/(main)/_features/media/_components/media-entry-score-badge"
import { AnilistAnimeEntryModal } from "@/app/(main)/_features/media/_containers/anilist-media-entry-modal"
import { useMissingEpisodes } from "@/app/(main)/_hooks/missing-episodes-loader"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { useAtom } from "jotai"
import { useSetAtom } from "jotai/react"
import capitalize from "lodash/capitalize"
import Link from "next/link"
import { usePathname } from "next/navigation"
import React, { useState } from "react"
import { BiPlay } from "react-icons/bi"
import { IoLibrarySharp } from "react-icons/io5"
import { RiCalendarLine } from "react-icons/ri"

type AnimeEntryCardBaseProps = {
    overlay?: React.ReactNode
    withAudienceScore?: boolean
    containerClassName?: string
    showListDataButton?: boolean
}

type AnimeEntryCardProps<T extends "anime" | "manga"> = {
    type: T
    media: T extends "anime" ? AL_BaseAnime : T extends "manga" ? AL_BaseManga : never
    // Anime-only
    listData?: T extends "anime" ? Anime_AnimeEntryListData : T extends "manga" ? Manga_EntryListData : never
    showLibraryBadge?: T extends "anime" ? boolean : never
    showTrailer?: T extends "anime" ? boolean : never
    libraryData?: T extends "anime" ? Anime_AnimeEntryLibraryData : never
} & AnimeEntryCardBaseProps

export function AnimeEntryCard<T extends "anime" | "manga">(props: AnimeEntryCardProps<T>) {

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

    const missingEpisodes = useMissingEpisodes()
    const [listData, setListData] = useState<Anime_AnimeEntryListData | undefined>(_listData)
    const [libraryData, setLibraryData] = useState<Anime_AnimeEntryLibraryData | undefined>(_libraryData)
    const setActionPopupHover = useSetAtom(__animeEntryCard_hoveredPopupId)

    const ref = React.useRef<HTMLDivElement>(null)

    const [__atomicLibraryCollection, getAtomicLibraryEntry] = useAtom(getAtomicLibraryEntryAtom)

    const showLibraryBadge = !!libraryData && !!props.showLibraryBadge

    const showProgressBar = React.useMemo(() => {
        return !!listData?.progress
        && type === "anime" ? !!(media as AL_BaseAnime)?.episodes : !!(media as AL_BaseManga)?.chapters
            && listData?.status !== "COMPLETED"
    }, [listData?.progress, media, listData?.status])

    const showTrailer = React.useMemo(() => _showTrailer && !libraryData && !media?.isAdult, [_showTrailer, libraryData, media])

    const link = type === "anime" ? `/entry?id=${media.id}` : `/manga/entry?id=${media.id}`

    const progressTotal = type === "anime" ? (media as AL_BaseAnime)?.episodes : (media as AL_BaseManga)?.chapters

    // React.useEffect(() => {
    //     console.log("rendered", media.title?.userPreferred)
    // }, [])

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
        <AnimeEntryCardContainer mRef={ref} className={props.containerClassName}>

            <AnimeEntryCardOverlay overlay={overlay} />

            {/*ACTION POPUP*/}
            <AnimeEntryCardHoverPopup
                onMouseEnter={() => setActionPopupHover(media.id)}
                onMouseLeave={() => setActionPopupHover(undefined)}
            >

                {/*METADATA SECTION*/}
                <AnimeEntryCardHoverPopupBody>

                    <AnimeEntryCardHoverPopupBanner
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

                    <AnimeEntryCardHoverPopupTitleSection
                        title={media.title?.userPreferred || ""}
                        year={media.startDate?.year}
                        season={media.season}
                        format={media.format}
                        link={link}
                    />

                    {type === "anime" && (
                        <AnimeEntryCardNextAiring nextAiring={(media as AL_BaseAnime).nextAiringEpisode} />
                    )}

                    {type === "anime" && <div className="py-1">
                        <Link
                            href={`/entry?id=${media.id}${(!!listData?.progress && (listData?.status !== "COMPLETED")) ? "&playNext=true" : ""}`}
                            tabIndex={-1}
                        >
                            <Button
                                leftIcon={<BiPlay className="text-2xl" />}
                                intent="white"
                                size="md"
                                className="w-full text-md"
                                tabIndex={-1}
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
                            tabIndex={-1}
                        >
                            Read
                        </Button>
                    </Link>}

                    {(listData?.status) &&
                        <p className="text-center text-sm text-[--muted]">
                            {listData?.status === "CURRENT" ? type === "anime" ? "Watching" : "Reading"
                                : capitalize(listData?.status ?? "")}
                        </p>}

                </AnimeEntryCardHoverPopupBody>

                <AnimeEntryCardHoverPopupFooter>

                    {(type === "anime" && !!libraryData) && <ToggleLockFilesButton mediaId={media.id} allFilesLocked={libraryData.allFilesLocked} />}

                    {showListDataButton && <AnilistAnimeEntryModal listData={listData} media={media} type={type} />}

                    {withAudienceScore &&
                        <AnimeEntryAudienceScore
                            meanScore={media.meanScore}
                        />}

                </AnimeEntryCardHoverPopupFooter>
            </AnimeEntryCardHoverPopup>


            <AnimeEntryCardBody
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
                <div className="absolute z-[10] left-1 bottom-1">
                    <AnimeEntryProgressBadge
                        progress={listData?.progress}
                        progressTotal={progressTotal}
                    />
                </div>
                <div className="absolute z-[10] right-1 bottom-1">
                    <AnimeEntryScoreBadge
                        score={listData?.score}
                    />
                </div>
                {(type === "anime" && !!libraryData && missingEpisodes.find(n => n.baseAnime?.id === media.id)) && (
                    <div className="absolute z-[10] w-full flex justify-center left-1 bottom-0">
                        <Badge
                            className="font-semibold animate-pulse text-white bg-gray-950 !bg-opacity-90 rounded-md text-base rounded-bl-none rounded-br-none"
                            intent="gray-solid"
                            size="xl"
                        ><RiCalendarLine /></Badge>
                    </div>
                )}
            </AnimeEntryCardBody>

            <AnimeEntryCardTitleSection
                title={media.title?.userPreferred || ""}
                year={media.startDate?.year}
                season={media.season}
                format={media.format}
            />

        </AnimeEntryCardContainer>
    )
}


