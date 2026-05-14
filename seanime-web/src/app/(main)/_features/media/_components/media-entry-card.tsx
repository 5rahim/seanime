import {
    AL_BaseAnime,
    AL_BaseManga,
    Anime_EntryLibraryData,
    Anime_EntryListData,
    Anime_LibraryCollectionEntry,
    Anime_NakamaEntryLibraryData,
    Manga_EntryListData,
} from "@/api/generated/types"
import { getAnimeLibraryEntryAtom } from "@/app/(main)/_atoms/anime-library-collection.atoms"
import { getMangaCollectionEntryAtom } from "@/app/(main)/_atoms/manga-collection.atoms"
import { usePlayNext } from "@/app/(main)/_atoms/playback.atoms"
import { ToggleLockFilesButton } from "@/app/(main)/_features/anime-library/_containers/toggle-lock-files-button"
import { AnimeEntryCardUnwatchedBadge } from "@/app/(main)/_features/anime/_containers/anime-entry-card-unwatched-badge"
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
import { useHasMissingEpisodes } from "@/app/(main)/_hooks/missing-episodes-loader"
import { useHasTorrentOrDebridInclusion, useIsSimulatedUser, useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { MangaEntryCardUnreadBadge } from "@/app/(main)/manga/_containers/manga-entry-card-unread-badge"
import { SeaLink } from "@/components/shared/sea-link"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { ContextMenuGroup, ContextMenuItem, ContextMenuLabel, ContextMenuTrigger } from "@/components/ui/context-menu"
import { preloadMediaEntry } from "@/lib/entry-preloader"
import { useRouter } from "@/lib/navigation"
import { __navigationPreloadModeAtom, shouldWarmEntryOnIntent } from "@/lib/navigation-preload-settings"
import { useAtomValue, useSetAtom } from "jotai/react"
import capitalize from "lodash/capitalize"
import React, { useState } from "react"
import { BiAddToQueue, BiPlay } from "react-icons/bi"
import { LuBookOpen } from "react-icons/lu"
import { LuEye, LuFolderTree } from "react-icons/lu"
import { RiCalendarLine } from "react-icons/ri"
import { PluginMediaCardContextMenuItems } from "../../plugin/actions/plugin-actions"

type MediaEntryCardBaseProps = {
    overlay?: React.ReactNode
    withAudienceScore?: boolean
    containerClassName?: string
    showListDataButton?: boolean
}

type MediaEntryCardListData = Anime_EntryListData | Manga_EntryListData

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

function useMediaCollectionEntry(type: "anime" | "manga", mediaId: number) {
    const entryAtom = React.useMemo(() => {
        return type === "anime" ? getAnimeLibraryEntryAtom(mediaId) : getMangaCollectionEntryAtom(mediaId)
    }, [mediaId, type])

    return useAtomValue(entryAtom)
}

export function MediaEntryCard<T extends "anime" | "manga">(props: MediaEntryCardProps<T>) {

    const {
        media,
        listData: _listData,
        libraryData: _libraryData,
        nakamaLibraryData,
        overlay,
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
    const isSimulatedUser = useIsSimulatedUser()
    const { hasStreamingEnabled } = useHasTorrentOrDebridInclusion()
    const navigationPreloadMode = useAtomValue(__navigationPreloadModeAtom)
    const setActionPopupHover = useSetAtom(__mediaEntryCard_hoveredPopupId)

    const { selectMediaAndOpenEditor } = usePlaylistEditorManager()

    const mediaId = media.id
    const mediaEpisodes = (media as AL_BaseAnime)?.episodes
    const mediaChapters = (media as AL_BaseManga)?.chapters
    const mediaIsAdult = media?.isAdult
    const collectionEntry = useMediaCollectionEntry(type, mediaId)
    const animeListDataFromCollection = useAnilistUserAnimeListData(mediaId, type === "anime" && !_listData)
    const animeCollectionEntry = type === "anime" ? collectionEntry as Anime_LibraryCollectionEntry | undefined : undefined

    const listData = React.useMemo<MediaEntryCardListData | undefined>(() => {
        if (_listData) {
            return _listData
        }

        if (type === "anime") {
            return animeListDataFromCollection ?? animeCollectionEntry?.listData
        }

        return collectionEntry?.listData
    }, [_listData, type, animeListDataFromCollection, animeCollectionEntry, collectionEntry])

    const libraryData = React.useMemo(() => {
        if (_libraryData) {
            return _libraryData
        }

        if (type !== "anime") {
            return undefined
        }

        return animeCollectionEntry?.libraryData
    }, [_libraryData, type, animeCollectionEntry])

    const hasMissingEpisodes = useHasMissingEpisodes(mediaId, type === "anime" && !!libraryData)

    const showLibraryBadge = !!libraryData && !!props.showLibraryBadge

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

    const shouldWarmCollectionEntryOnViewport = !!collectionEntry || !!animeListDataFromCollection
    const shouldBypassPreloadBudget = !!collectionEntry || !!animeListDataFromCollection || !!_listData

    const progressTotal = type === "anime" ? (media as AL_BaseAnime)?.episodes : (media as AL_BaseManga)?.chapters

    const { setPlayNext } = usePlayNext()
    const handleWatchButtonClicked = React.useCallback(() => {
        setPlayNext(mediaId, () => {
            router.push(ANIME_LINK)
        })
    }, [mediaId, ANIME_LINK, setPlayNext, router])

    const onPopupMouseEnter = React.useCallback(() => {
        setActionPopupHover(mediaId)
    }, [mediaId, setActionPopupHover])

    const onPopupMouseLeave = React.useCallback(() => {
        setActionPopupHover(undefined)
    }, [setActionPopupHover])

    const { setPreviewModalMediaId } = useMediaPreviewModal()
    const { openDirInLibraryExplorer } = useLibraryExplorer()

    const [shouldRenderPopup, setShouldRenderPopup] = useState(false)
    const closePopupTimerRef = React.useRef<number | undefined>(undefined)

    const warmCardEntry = React.useCallback(() => {
        if (onClick || !shouldWarmEntryOnIntent(navigationPreloadMode, isSimulatedUser)) return
        preloadMediaEntry(link, { bypassBudget: shouldBypassPreloadBudget })
    }, [isSimulatedUser, link, navigationPreloadMode, onClick, shouldBypassPreloadBudget])

    const handleCardMouseEnter = React.useCallback(() => {
        if (closePopupTimerRef.current) {
            window.clearTimeout(closePopupTimerRef.current)
            closePopupTimerRef.current = undefined
        }
        setShouldRenderPopup(true)
        warmCardEntry()
    }, [warmCardEntry])

    const handleCardMouseLeave = React.useCallback(() => {
        if (closePopupTimerRef.current) {
            window.clearTimeout(closePopupTimerRef.current)
        }
        closePopupTimerRef.current = window.setTimeout(() => {
            setShouldRenderPopup(false)
            closePopupTimerRef.current = undefined
        }, 35)
    }, [])

    React.useEffect(() => {
        return () => {
            if (closePopupTimerRef.current) {
                window.clearTimeout(closePopupTimerRef.current)
            }
        }
    }, [])

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
            onPointerEnter={handleCardMouseEnter}
            onPointerLeave={handleCardMouseLeave}
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
                                bypassEntryPreloadBudget={shouldBypassPreloadBudget}
                                listStatus={listData?.status}
                                status={media.status}
                                onClick={onClick}
                            />

                            <MediaEntryCardHoverPopupTitleSection
                                title={media.title?.userPreferred || ""}
                                allTitles={media.title}
                                year={(media as AL_BaseAnime).seasonYear ?? media.startDate?.year}
                                season={media.season}
                                format={media.format}
                                link={link}
                                bypassEntryPreloadBudget={shouldBypassPreloadBudget}
                                onClick={onClick}
                                // onHover={() => setHoveringTitle(true)}
                                // onHoverLeave={() => setHoveringTitle(false)}
                            />

                            <div className="py-1 flex items-center justify-center w-full gap-2">
                                {type === "anime" && <Button
                                    leftIcon={<BiPlay className="text-2xl" />}
                                    intent="gray-subtle"
                                    size="sm"
                                    className="w-full"
                                    tabIndex={-1}
                                    onClick={handleWatchButtonClicked}
                                >
                                    {!!listData?.progress && (listData?.status === "CURRENT" || listData?.status === "PAUSED")
                                        ? "Continue"
                                        : "Watch"}
                                </Button>}

                                {type === "manga" && <SeaLink
                                    href={!onClick ? MANGA_LINK : undefined}
                                    bypassEntryPreloadBudget={shouldBypassPreloadBudget}
                                    onClick={onClick}
                                    className="block w-full"
                                >
                                    <Button
                                        leftIcon={<LuBookOpen />}
                                        intent="gray-subtle"
                                        size="sm"
                                        className="w-full"
                                        tabIndex={-1}
                                    >
                                        {!!listData?.progress && (listData?.status === "CURRENT" || listData?.status === "PAUSED")
                                            ? "Continue"
                                            : "Start Reading"}
                                    </Button>
                                </SeaLink>}
                            </div>

                            {type === "anime" && (
                                <AnimeEntryCardNextAiring nextAiring={(media as AL_BaseAnime).nextAiringEpisode} />
                            )}

                            {(listData?.status && listData?.status !== "CURRENT") &&
                                <p className="text-center text-xs text-[--muted] w-full">
                                    {capitalize(listData?.status ?? "")}
                                    {/*{listData?.status === "CURRENT" ? type === "anime" ? "Watching" : "Reading"*/}
                                    {/*    : capitalize(listData?.status ?? "")}*/}
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
                bypassEntryPreloadBudget={shouldBypassPreloadBudget}
                warmEntryOnViewport={shouldWarmCollectionEntryOnViewport}
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
                {(type === "anime" && !!libraryData && hasMissingEpisodes) && (
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


