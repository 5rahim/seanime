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
import { useLibraryExplorer } from "@/app/(main)/_features/library-explorer/library-explorer.atoms"
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
import { usePlaylistEditorManager } from "@/app/(main)/_features/playlists/lib/playlist-editor-manager"
import { useAnilistUserAnimeListData } from "@/app/(main)/_hooks/anilist-collection-loader"
import { useMissingEpisodes } from "@/app/(main)/_hooks/missing-episodes-loader"
import { useHasTorrentOrDebridInclusion, useServerStatus } from "@/app/(main)/_hooks/use-server-status"
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
import { BiAddToQueue, BiPlay } from "react-icons/bi"
import { IoLibrarySharp } from "react-icons/io5"
import { LuEye, LuFolderTree } from "react-icons/lu"
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
    onClick?: () => void
    hideReleasingBadge?: boolean
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
        onClick,
        hideReleasingBadge = false,
    } = props

    const router = useRouter()
    const serverStatus = useServerStatus()
    const { hasStreamingEnabled } = useHasTorrentOrDebridInclusion()
    const missingEpisodes = useMissingEpisodes()

    const prevListDataRef = React.useRef(_listData)
    const prevLibraryDataRef = React.useRef(_libraryData)

    const [listData, setListData] = useState<Anime_EntryListData | undefined>(_listData)
    const [libraryData, setLibraryData] = useState<Anime_EntryLibraryData | undefined>(_libraryData)
    const setActionPopupHover = useSetAtom(__mediaEntryCard_hoveredPopupId)

    const { selectMediaAndOpenEditor } = usePlaylistEditorManager()

    const [__atomicLibraryCollection, getAtomicLibraryEntry] = useAtom(getAtomicLibraryEntryAtom)

    const showLibraryBadge = !!libraryData && !!props.showLibraryBadge

    const mediaId = media.id
    const mediaEpisodes = (media as AL_BaseAnime)?.episodes
    const mediaChapters = (media as AL_BaseManga)?.chapters
    const mediaIsAdult = media?.isAdult

    const showProgressBar = React.useMemo(() => {
        return !!listData?.progress
        && type === "anime" ? !!mediaEpisodes : !!mediaChapters
            && listData?.status !== "COMPLETED"
    }, [listData?.progress, mediaEpisodes, mediaChapters, listData?.status, type])

    const showTrailer = React.useMemo(() => _showTrailer && !libraryData && !mediaIsAdult, [_showTrailer, libraryData, mediaIsAdult])

    const MANGA_LINK = React.useMemo(() =>
            serverStatus?.isOffline ? `/offline/entry/manga?id=${mediaId}` : `/manga/entry?id=${mediaId}`,
        [serverStatus?.isOffline, mediaId],
    )
    const ANIME_LINK = React.useMemo(() =>
            serverStatus?.isOffline ? `/offline/entry/anime?id=${mediaId}` : `/entry?id=${mediaId}`,
        [serverStatus?.isOffline, mediaId],
    )

    const link = React.useMemo(() => {
        return type === "anime" ? ANIME_LINK : MANGA_LINK
    }, [ANIME_LINK, MANGA_LINK, type])

    const progressTotal = type === "anime" ? (media as AL_BaseAnime)?.episodes : (media as AL_BaseManga)?.chapters

    const pathname = usePathname()

    React.useEffect(() => {
        if (_listData !== prevListDataRef.current) {
            prevListDataRef.current = _listData
            setListData(_listData)
        }
    }, [_listData])

    React.useEffect(() => {
        if (_libraryData !== prevLibraryDataRef.current) {
            prevLibraryDataRef.current = _libraryData
            setLibraryData(_libraryData)
        }
    }, [_libraryData])

    // Dynamically refresh data when LibraryCollection is updated
    React.useEffect(() => {
        if (pathname !== "/") {
            const entry = getAtomicLibraryEntry(mediaId)
            if (!_listData) {
                setListData(entry?.listData)
            }
            if (!_libraryData) {
                setLibraryData(entry?.libraryData)
            }
        }
    }, [pathname, __atomicLibraryCollection])

    const listDataFromCollection = useAnilistUserAnimeListData(mediaId)

    React.useEffect(() => {
        if (listDataFromCollection && !_listData && listDataFromCollection !== listData) {
            setListData(listDataFromCollection)
        }
    }, [listDataFromCollection, _listData, listData])

    const { setPlayNext } = usePlayNext()
    const handleWatchButtonClicked = React.useCallback(() => {
        if ((!!listData?.progress && (listData?.status !== "COMPLETED"))) {
            setPlayNext(mediaId, () => {
                router.push(ANIME_LINK)
            })
        } else {
            router.push(ANIME_LINK)
        }
    }, [listData?.progress, listData?.status, mediaId, ANIME_LINK, setPlayNext, router])

    const onPopupMouseEnter = React.useCallback(() => {
        setActionPopupHover(mediaId)
    }, [mediaId, setActionPopupHover])

    const onPopupMouseLeave = React.useCallback(() => {
        setActionPopupHover(undefined)
    }, [setActionPopupHover])

    const { setPreviewModalMediaId } = useMediaPreviewModal()
    const { openDirInLibraryExplorer } = useLibraryExplorer()

    const [hoveringTitle, setHoveringTitle] = useState(false)
    const [isHoveringCard, setIsHoveringCard] = useState(false)
    const [shouldRenderPopup, setShouldRenderPopup] = useState(false)

    // Handle delayed unmount for exit animation
    React.useEffect(() => {
        if (isHoveringCard) {
            setShouldRenderPopup(true)
            return
        } else {
            // Delay unmount to allow exit animation
            const timer = setTimeout(() => {
                setShouldRenderPopup(false)
            }, 35) // Match animation duration
            return () => clearTimeout(timer)
        }
    }, [isHoveringCard])

    const handlePreviewClick = React.useCallback(() => {
        setPreviewModalMediaId(mediaId, type)
    }, [mediaId, type, setPreviewModalMediaId])

    const handleAddToPlaylistClick = React.useCallback(() => {
        selectMediaAndOpenEditor(mediaId)
    }, [mediaId, selectMediaAndOpenEditor])

    const handleOpenInExplorerClick = React.useCallback(() => {
        if (libraryData?.sharedPath) {
            openDirInLibraryExplorer(libraryData.sharedPath)
        }
    }, [libraryData?.sharedPath, openDirInLibraryExplorer])

    const stringifiedListData = React.useMemo(() => JSON.stringify(listData), [listData])

    if (!media) return null

    return (
        <MediaEntryCardContainer
            data-media-id={media.id}
            data-media-mal-id={media.idMal}
            data-media-type={type}
            className={props.containerClassName}
            data-list-data={stringifiedListData}
            onMouseEnter={() => setIsHoveringCard(true)}
            onMouseLeave={() => setIsHoveringCard(false)}
        >

            <MediaEntryCardOverlay overlay={overlay} />

            <SeaContextMenu
                content={<ContextMenuGroup>
                    <ContextMenuLabel className="text-[--muted] line-clamp-1 py-0 my-2">
                        {media.title?.userPreferred}
                    </ContextMenuLabel>
                    {!serverStatus?.isOffline && <ContextMenuItem
                        onClick={handlePreviewClick}
                    >
                        <LuEye /> Preview
                    </ContextMenuItem>}
                    {(libraryData || nakamaLibraryData || (listData && hasStreamingEnabled)) && <ContextMenuItem
                        onClick={handleAddToPlaylistClick}
                    >
                        <BiAddToQueue /> Add to Playlist
                    </ContextMenuItem>}
                    {(!!libraryData) && <ContextMenuItem
                        onClick={handleOpenInExplorerClick}
                    >
                        <LuFolderTree /> Open in Library Explorer
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
                        shouldRenderPopup={shouldRenderPopup}
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
                                onClick={onClick}
                            />

                            <MediaEntryCardHoverPopupTitleSection
                                title={media.title?.userPreferred || ""}
                                year={(media as AL_BaseAnime).seasonYear ?? media.startDate?.year}
                                season={media.season}
                                format={media.format}
                                link={link}
                                onClick={onClick}
                                // onHover={() => setHoveringTitle(true)}
                                // onHoverLeave={() => setHoveringTitle(false)}
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
                                href={!onClick ? MANGA_LINK : undefined}
                                onClick={onClick}
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

                            {/*{hoveringTitle && <div>*/}
                            {/*    <p*/}
                            {/*        data-media-entry-card-hover-popup-title-section-year-season*/}
                            {/*        className="justify-center text-center text-xs text-[--muted] flex w-full gap-1 items-center px-4 leading-0 line-clamp-2"*/}
                            {/*    >*/}
                            {/*        {(media.title?.english && media.title?.userPreferred !== media.title?.english) ? `${startCase(media.title?.english)}` : null}*/}
                            {/*        {(media.title?.romaji && media.title?.userPreferred !== media.title?.romaji) ? `${startCase(media.title?.romaji)}` : null}*/}
                            {/*    </p>*/}
                            {/*</div>}*/}

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
                onClick={onClick}
                hideReleasingBadge={hideReleasingBadge}
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


