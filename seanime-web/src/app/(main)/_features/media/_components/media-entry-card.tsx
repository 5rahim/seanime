import { AL_BaseAnime, AL_BaseManga, Anime_EntryLibraryData, Anime_EntryListData, Manga_EntryListData } from "@/api/generated/types"
import { getAtomicLibraryEntryAtom } from "@/app/(main)/_atoms/anime-library-collection.atoms"
import { usePlayNext } from "@/app/(main)/_atoms/playback.atoms"
import { ToggleLockFilesButton } from "@/app/(main)/_features/anime/_containers/toggle-lock-files-button"
import {
    __mediaEntryCard_hoveredPopupId,
    AnimeEntryCardNextAiring,
    MediaEntryCardBody,
    MediaEntryCardContainer,
    MediaEntryCardHoverPopup,
    MediaEntryCardHoverPopupBanner,
    MediaEntryCardHoverPopupBody,
    MediaEntryCardHoverPopupFooter,
    MediaEntryCardHoverPopupTitleSection,
    MediaEntryCardOverlay,
    MediaEntryCardTitleSection,
} from "@/app/(main)/_features/media/_components/media-entry-card-components"
import { MediaEntryAudienceScore } from "@/app/(main)/_features/media/_components/media-entry-metadata-components"
import { MediaEntryProgressBadge } from "@/app/(main)/_features/media/_components/media-entry-progress-badge"
import { MediaEntryScoreBadge } from "@/app/(main)/_features/media/_components/media-entry-score-badge"
import { AnilistMediaEntryModal } from "@/app/(main)/_features/media/_containers/anilist-media-entry-modal"
import { useAnilistUserAnimeListData } from "@/app/(main)/_hooks/anilist-collection-loader"
import { useMissingEpisodes } from "@/app/(main)/_hooks/missing-episodes-loader"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { SeaLink } from "@/components/shared/sea-link"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { useAtom } from "jotai"
import { useSetAtom } from "jotai/react"
import capitalize from "lodash/capitalize"
import { usePathname, useRouter } from "next/navigation"
import React, { useState } from "react"
import { BiPlay } from "react-icons/bi"
import { IoLibrarySharp } from "react-icons/io5"
import { RiCalendarLine } from "react-icons/ri"

type MediaEntryCardBaseProps = {
    overlay?: React.ReactNode
    withAudienceScore?: boolean
    containerClassName?: string
    showListDataButton?: boolean
}

type MediaEntryCardProps<T extends "anime" | "manga"> = {
    type: T
    media: T extends "anime" ? AL_BaseAnime : T extends "manga" ? AL_BaseManga : never
    // Anime-only
    listData?: T extends "anime" ? Anime_EntryListData : T extends "manga" ? Manga_EntryListData : never
    showLibraryBadge?: T extends "anime" ? boolean : never
    showTrailer?: T extends "anime" ? boolean : never
    libraryData?: T extends "anime" ? Anime_EntryLibraryData : never
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

    const router = useRouter()
    const missingEpisodes = useMissingEpisodes()
    const [listData, setListData] = useState<Anime_EntryListData | undefined>(_listData)
    const [libraryData, setLibraryData] = useState<Anime_EntryLibraryData | undefined>(_libraryData)
    const setActionPopupHover = useSetAtom(__mediaEntryCard_hoveredPopupId)

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

    const pathname = usePathname()
    //
    // // Dynamically refresh data when LibraryCollection is updated
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

    const listDataFromCollection = useAnilistUserAnimeListData(media.id)

    React.useEffect(() => {
        if (listDataFromCollection && !_listData) {
            setListData(listDataFromCollection)
        }
    }, [listDataFromCollection, _listData])

    const { setPlayNext } = usePlayNext()
    const handleWatchButtonClicked = React.useCallback(() => {
        if ((!!listData?.progress && (listData?.status !== "COMPLETED"))) {
            setPlayNext(media.id, () => {
                router.push(`/entry?id=${media.id}`)
            })
        } else {
            router.push(`/entry?id=${media.id}`)
        }
    }, [])

    if (!media) return null

    return (
        <MediaEntryCardContainer mRef={ref} className={props.containerClassName}>

            <MediaEntryCardOverlay overlay={overlay} />

            {/*ACTION POPUP*/}
            <MediaEntryCardHoverPopup
                onMouseEnter={() => setActionPopupHover(media.id)}
                onMouseLeave={() => setActionPopupHover(undefined)}
                coverImage={media.bannerImage || media.coverImage?.extraLarge || ""}
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
                        <AnimeEntryCardNextAiring nextAiring={(media as AL_BaseAnime).nextAiringEpisode} />
                    )}

                    {type === "anime" && <div className="py-1">
                        {/*<Link*/}
                        {/*    href={`/entry?id=${media.id}${(!!listData?.progress && (listData?.status !== "COMPLETED")) ? "&playNext=true" : ""}`}*/}
                        {/*    tabIndex={-1}*/}
                        {/*>*/}
                        <Button
                            leftIcon={<BiPlay className="text-2xl" />}
                            intent="white"
                            size="md"
                            className="w-full text-md"
                            tabIndex={-1}
                            onClick={handleWatchButtonClicked}
                        >
                            {!!listData?.progress && (listData?.status === "CURRENT" || listData?.status === "PAUSED")
                                ? "Continue watching"
                                : "Watch"}
                        </Button>
                        {/*</Link>*/}
                    </div>}

                    {type === "manga" && <SeaLink
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
                    </SeaLink>}

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
                <div className="absolute z-[10] left-0 bottom-0">
                    <MediaEntryProgressBadge
                        progress={listData?.progress}
                        progressTotal={progressTotal}
                    />
                </div>
                <div className="absolute z-[10] right-1 bottom-1">
                    <MediaEntryScoreBadge
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


