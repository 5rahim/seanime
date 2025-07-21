import {
    AL_BaseAnime,
    AL_BaseManga,
    Anime_EntryLibraryData,
    Anime_EntryListData,
    Anime_NakamaEntryLibraryData,
    Manga_EntryListData,
} from "@/api/generated/types"
import { getAtomicLibraryEntryAtom } from "@/app/(main)/_atoms/anime-library-collection.atoms"
import { usePlayNext } from "@/app/(main)/_atoms/playback.atoms"
import { AnimeEntryCardUnwatchedBadge } from "@/app/(main)/_features/anime/_containers/anime-entry-card-unwatched-badge"
import { ToggleLockFilesButton } from "@/app/(main)/_features/anime/_containers/toggle-lock-files-button"
import { SeaContextMenu } from "@/app/(main)/_features/context-menu/sea-context-menu"
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
import { useMediaPreviewModal } from "@/app/(main)/_features/media/_containers/media-preview-modal"
import { useAnilistUserAnimeListData } from "@/app/(main)/_hooks/anilist-collection-loader"
import { useMissingEpisodes } from "@/app/(main)/_hooks/missing-episodes-loader"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { MangaEntryCardUnreadBadge } from "@/app/(main)/manga/_containers/manga-entry-card-unread-badge"
import { SeaLink } from "@/components/shared/sea-link"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { ContextMenuGroup, ContextMenuItem, ContextMenuLabel, ContextMenuTrigger } from "@/components/ui/context-menu"
import { useAtom } from "jotai"
import { useSetAtom } from "jotai/react"
import capitalize from "lodash/capitalize"
import { usePathname, useRouter } from "next/navigation"
import React, { useState } from "react"
import { BiPlay } from "react-icons/bi"
import { IoLibrarySharp } from "react-icons/io5"
import { RiCalendarLine } from "react-icons/ri"
import { PluginMediaCardContextMenuItems } from "../../plugin/actions/plugin-actions"

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
    nakamaLibraryData?: T extends "anime" ? Anime_NakamaEntryLibraryData : never
    hideUnseenCountBadge?: boolean
    hideAnilistEntryEditButton?: boolean
} & MediaEntryCardBaseProps

export function MediaEntryCard<T extends "anime" | "manga">(props: MediaEntryCardProps<T>) {

    const {
        media,
        listData: _listData,
        libraryData: _libraryData,
        nakamaLibraryData,
        overlay,
        showListDataButton,
        showTrailer: _showTrailer,
        type,
        withAudienceScore = true,
        hideUnseenCountBadge = false,
        hideAnilistEntryEditButton = false,
    } = props

    const router = useRouter()
    const serverStatus = useServerStatus()
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

    const MANGA_LINK = serverStatus?.isOffline ? `/offline/entry/manga?id=${media.id}` : `/manga/entry?id=${media.id}`
    const ANIME_LINK = serverStatus?.isOffline ? `/offline/entry/anime?id=${media.id}` : `/entry?id=${media.id}`

    const link = React.useMemo(() => {
        return type === "anime" ? ANIME_LINK : MANGA_LINK
    }, [serverStatus?.isOffline, type])

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
                router.push(ANIME_LINK)
            })
        } else {
            router.push(ANIME_LINK)
        }
    }, [listData?.progress, listData?.status, media.id])

    const onPopupMouseEnter = React.useCallback(() => {
        setActionPopupHover(media.id)
    }, [media.id])

    const onPopupMouseLeave = React.useCallback(() => {
        setActionPopupHover(undefined)
    }, [media.id])

    const { setPreviewModalMediaId } = useMediaPreviewModal()

    if (!media) return null

    return (
        <MediaEntryCardContainer
            data-media-id={media.id}
            data-media-mal-id={media.idMal}
            data-media-type={type}
            mRef={ref}
            className={props.containerClassName}
            data-list-data={JSON.stringify(listData)}
        >

            <MediaEntryCardOverlay overlay={overlay} />

            <SeaContextMenu
                content={<ContextMenuGroup>
                    <ContextMenuLabel className="text-[--muted] line-clamp-1 py-0 my-2">
                        {media.title?.userPreferred}
                    </ContextMenuLabel>
                    {!serverStatus?.isOffline && <ContextMenuItem
                        onClick={() => {
                            setPreviewModalMediaId(media.id!, type)
                        }}
                    >
                        Preview
                    </ContextMenuItem>}

                    <PluginMediaCardContextMenuItems for={type} media={media} />
                </ContextMenuGroup>}
            >
                <ContextMenuTrigger>

                    {/*ACTION POPUP*/}
                    <MediaEntryCardHoverPopup
                        onMouseEnter={onPopupMouseEnter}
                        onMouseLeave={onPopupMouseLeave}
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
                                year={(media as AL_BaseAnime).seasonYear ?? media.startDate?.year}
                                season={media.season}
                                format={media.format}
                                link={link}
                            />

                            {type === "anime" && (
                                <AnimeEntryCardNextAiring nextAiring={(media as AL_BaseAnime).nextAiringEpisode} />
                            )}

                            {type === "anime" && <div className="py-1">
                                <Button
                                    leftIcon={<BiPlay className="text-2xl" />}
                                    intent="gray-subtle"
                                    size="sm"
                                    className="w-full text-sm"
                                    tabIndex={-1}
                                    onClick={handleWatchButtonClicked}
                                >
                                    {!!listData?.progress && (listData?.status === "CURRENT" || listData?.status === "PAUSED")
                                        ? "Continue watching"
                                        : "Watch"}
                                </Button>
                            </div>}

                            {type === "manga" && <SeaLink
                                href={MANGA_LINK}
                            >
                                <Button
                                    leftIcon={<IoLibrarySharp />}
                                    intent="gray-subtle"
                                    size="sm"
                                    className="w-full text-sm mt-2"
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

                            {(type === "anime" && !!libraryData) &&
                                <ToggleLockFilesButton mediaId={media.id} allFilesLocked={libraryData.allFilesLocked} />}

                            {!hideAnilistEntryEditButton && <AnilistMediaEntryModal listData={listData} media={media} type={type} forceModal />}

                            {withAudienceScore &&
                                <MediaEntryAudienceScore
                                    meanScore={media.meanScore}
                                />}

                        </MediaEntryCardHoverPopupFooter>
                    </MediaEntryCardHoverPopup>
                </ContextMenuTrigger>
            </SeaContextMenu>


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
                <div data-media-entry-card-body-progress-badge-container className="absolute z-[10] left-0 bottom-0 flex items-end">
                    <MediaEntryProgressBadge
                        progress={listData?.progress}
                        progressTotal={progressTotal}
                        forceShowTotal={type === "manga"}
                        // forceShowProgress={listData?.status === "CURRENT"}
                        top={!hideUnseenCountBadge ? <>

                            {(type === "anime" && (listData?.status === "CURRENT" || listData?.status === "REPEATING")) && (
                                <AnimeEntryCardUnwatchedBadge
                                    progress={listData?.progress || 0}
                                    media={media}
                                    libraryData={libraryData}
                                    nakamaLibraryData={nakamaLibraryData}
                                />
                            )}
                            {type === "manga" &&
                                <MangaEntryCardUnreadBadge mediaId={media.id} progress={listData?.progress} progressTotal={progressTotal} />}
                        </> : null}
                    />
                </div>
                <div data-media-entry-card-body-score-badge-container className="absolute z-[10] right-0 bottom-0">
                    <MediaEntryScoreBadge
                        isMediaCard
                        score={listData?.score}
                    />
                </div>
                {(type === "anime" && !!libraryData && missingEpisodes.find(n => n.baseAnime?.id === media.id)) && (
                    <div
                        data-media-entry-card-body-missing-episodes-badge-container
                        className="absolute z-[10] w-full flex justify-center left-1 bottom-0"
                    >
                        <Badge
                            className="font-semibold animate-pulse text-white bg-gray-950 !bg-opacity-90 rounded-[--radius-md] text-base rounded-bl-none rounded-br-none"
                            intent="gray-solid"
                            size="xl"
                        ><RiCalendarLine /></Badge>
                    </div>
                )}

            </MediaEntryCardBody>

            <MediaEntryCardTitleSection
                title={media.title?.userPreferred || ""}
                year={(media as AL_BaseAnime).seasonYear ?? media.startDate?.year}
                season={media.season}
                format={media.format}
            />

        </MediaEntryCardContainer>
    )
}


