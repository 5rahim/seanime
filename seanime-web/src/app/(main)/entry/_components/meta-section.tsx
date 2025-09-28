"use client"
import { AL_AnimeDetailsById_Media, Anime_Entry } from "@/api/generated/types"
import { TrailerModal } from "@/app/(main)/_features/anime/_components/trailer-modal"
import { AnimeAutoDownloaderButton } from "@/app/(main)/_features/anime/_containers/anime-auto-downloader-button"
import { ToggleLockFilesButton } from "@/app/(main)/_features/anime/_containers/toggle-lock-files-button"
import { AnimeEntryStudio } from "@/app/(main)/_features/media/_components/anime-entry-studio"
import {
    AnimeEntryRankings,
    MediaEntryAudienceScore,
    MediaEntryGenresList,
} from "@/app/(main)/_features/media/_components/media-entry-metadata-components"
import {
    MediaPageHeader,
    MediaPageHeaderDetailsContainer,
    MediaPageHeaderEntryDetails,
} from "@/app/(main)/_features/media/_components/media-page-header-components"
import { MediaSyncTrackButton } from "@/app/(main)/_features/media/_containers/media-sync-track-button"
import { useHasDebridService, useHasTorrentProvider, useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { NextAiringEpisode } from "@/app/(main)/entry/_components/next-airing-episode"
import { EntrySectionTabs, useAnimeEntryPageView } from "@/app/(main)/entry/_containers/anime-entry-page"
import { AnimeEntryDropdownMenu } from "@/app/(main)/entry/_containers/entry-actions/anime-entry-dropdown-menu"
import { AnimeEntrySilenceToggle } from "@/app/(main)/entry/_containers/entry-actions/anime-entry-silence-toggle"
import { TorrentSearchButton } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-button"
import { SeaLink } from "@/components/shared/sea-link"
import { Badge } from "@/components/ui/badge"
import { Button, ButtonProps, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Tooltip } from "@/components/ui/tooltip"
import { TORRENT_CLIENT } from "@/lib/server/settings"
import { getCustomSourceExtensionId, getCustomSourceMediaSiteUrl, isCustomSource } from "@/lib/server/utils"
import { useThemeSettings } from "@/lib/theme/hooks"
import React from "react"
import { BiExtension } from "react-icons/bi"
import { IoInformationCircle } from "react-icons/io5"
import { LuExternalLink } from "react-icons/lu"
import { MdOutlineConnectWithoutContact } from "react-icons/md"
import { SiAnilist } from "react-icons/si"
import { useNakamaStatus } from "../../_features/nakama/nakama-manager"
import { PluginAnimePageButtons } from "../../_features/plugin/actions/plugin-actions"

export function AnimeMetaActionButton({ className, ...rest }: ButtonProps) {
    const ts = useThemeSettings()
    return <Button
        className={cn(
            "w-full",
            "lg:w-full lg:max-w-[280px]",
            className,
        )}
        {...rest}
        // intent="gray-outline"
    />
}

export function MetaSection(props: { entry: Anime_Entry, details: AL_AnimeDetailsById_Media | undefined }) {
    const serverStatus = useServerStatus()
    const { entry, details } = props
    const ts = useThemeSettings()
    const nakamaStatus = useNakamaStatus()

    if (!entry.media) return null

    const { hasTorrentProvider } = useHasTorrentProvider()
    const { hasDebridService } = useHasDebridService()
    const { currentView, isLibraryView, isTorrentStreamingView, isDebridStreamingView, isOnlineStreamingView } = useAnimeEntryPageView()

    const listData = entry.listData
    const type = "anime"

    return (
        <MediaPageHeader
            backgroundImage={entry.media?.bannerImage}
            coverImage={entry.media?.coverImage?.extraLarge}
        >

            <MediaPageHeaderDetailsContainer>

                <MediaPageHeaderEntryDetails
                    coverImage={entry.media?.coverImage?.extraLarge || entry.media?.coverImage?.large}
                    title={entry.media?.title?.userPreferred}
                    color={entry.media?.coverImage?.color}
                    englishTitle={entry.media?.title?.english}
                    romajiTitle={entry.media?.title?.romaji}
                    startDate={entry.media?.startDate}
                    season={entry.media?.season}
                    progressTotal={entry.media?.episodes}
                    status={entry.media?.status}
                    description={entry.media?.description}
                    listData={entry.listData}
                    media={entry.media}
                    type="anime"
                >
                    <div
                        data-anime-meta-section-details
                        className={cn(
                            "flex gap-3 flex-wrap items-center",
                            "justify-center lg:justify-start lg:max-w-[65vw]",
                        )}
                    >

                        <MediaEntryAudienceScore meanScore={entry.media?.meanScore} badgeClass="bg-transparent" />


                        {!isCustomSource(entry.mediaId) ? <AnimeEntryStudio studios={details?.studios} /> : (
                            <Badge
                                size="lg"
                                intent="gray"
                                className="rounded-full px-0 border-transparent bg-transparent transition-all hover:bg-transparent hover:text-white hover:-translate-y-0.5"
                                data-anime-entry-studio-badge
                            >
                                {details?.studios?.nodes?.[0]?.name}
                            </Badge>
                        )}

                        <MediaEntryGenresList genres={details?.genres} />

                        <div
                            data-anime-meta-section-rankings-container
                            className={cn(
                                "w-full",
                            )}
                        >
                            <AnimeEntryRankings rankings={details?.rankings} />
                        </div>
                    </div>
                </MediaPageHeaderEntryDetails>


                <div
                    data-anime-meta-section-buttons-container
                    className={cn(
                        "flex flex-row w-full gap-3 items-center justify-center lg:justify-start lg:max-w-[65vw]",
                        "flex-wrap",
                    )}
                >

                    {isCustomSource(entry.mediaId) && (
                        <Tooltip
                            trigger={<div>
                                <SeaLink href={`/custom-sources?provider=${getCustomSourceExtensionId(entry.media)}`}>
                                    <IconButton size="sm" intent="gray-link" className="px-0" icon={<BiExtension className="text-lg" />} />
                                </SeaLink>
                            </div>}
                        >
                            Custom source
                        </Tooltip>
                    )}

                    {!isCustomSource(entry.mediaId) && <SeaLink href={`https://anilist.co/anime/${entry.mediaId}`} target="_blank">
                        <IconButton size="sm" intent="gray-link" className="px-0" icon={<SiAnilist className="text-lg" />} />
                    </SeaLink>}

                    {isCustomSource(entry.mediaId) && !!getCustomSourceMediaSiteUrl(entry.media) && <Tooltip
                        trigger={<div>
                            <SeaLink href={getCustomSourceMediaSiteUrl(entry.media)!} target="_blank">
                                <IconButton size="sm" intent="gray-link" className="px-0" icon={<LuExternalLink className="text-lg" />} />
                            </SeaLink>
                        </div>}
                    >
                        Open in website
                    </Tooltip>}

                    {!!entry?.media?.trailer?.id && <TrailerModal
                        trailerId={entry?.media?.trailer?.id} trigger={
                        <Button size="sm" intent="gray-link" className="px-0">
                            Trailer
                        </Button>}
                    />}

                    <AnimeAutoDownloaderButton entry={entry} size="md" />

                    {isLibraryView && !entry._isNakamaEntry && !!entry.libraryData && <>
                        <MediaSyncTrackButton mediaId={entry.mediaId} type="anime" size="md" />
                        <AnimeEntrySilenceToggle mediaId={entry.mediaId} size="md" />
                        <ToggleLockFilesButton
                            allFilesLocked={entry.libraryData.allFilesLocked}
                            mediaId={entry.mediaId}
                            size="md"
                        />
                    </>}
                    <AnimeEntryDropdownMenu entry={entry} />


                    {(
                        entry.media.status !== "NOT_YET_RELEASED"
                        && currentView === "library"
                        && hasTorrentProvider
                        && (
                            serverStatus?.settings?.torrent?.defaultTorrentClient !== TORRENT_CLIENT.NONE
                            || hasDebridService
                        )
                        && !entry._isNakamaEntry
                    ) && (
                        <TorrentSearchButton
                            entry={entry}
                        />
                    )}

                    {entry._isNakamaEntry && currentView === "library" &&
                        <div className="flex items-center gap-2 h-10 px-4 border rounded-md flex-none">
                            <MdOutlineConnectWithoutContact className="size-6 animate-pulse text-[--blue]" />
                            <span className="text-sm tracking-wide">Shared by {nakamaStatus?.hostConnectionStatus?.username}</span>
                        </div>}

                    <PluginAnimePageButtons media={entry.media!} />

                </div>

                <EntrySectionTabs entry={entry} />

                <NextAiringEpisode media={entry.media} />

                {entry.downloadInfo?.hasInaccurateSchedule && <p
                    className={cn(
                        "text-[--muted] text-sm text-center mb-3",
                        "text-left",
                    )}
                    data-anime-meta-section-inaccurate-schedule-message
                >
                    <span className="block">Could not retrieve accurate scheduling information for this show.</span>
                    <span className="block text-[--muted]">Please check the schedule online for more information.</span>
                </p>}


                {(!entry.anidbId || entry.anidbId === 0) && !isCustomSource(entry.mediaId) && entry.media?.status !== "NOT_YET_RELEASED" && (
                    <p
                        className={cn(
                            "text-center text-gray-200 opacity-50 text-sm flex gap-1 items-center",
                            "text-left",
                        )}
                        data-anime-meta-section-no-metadata-message
                    >
                        <IoInformationCircle />
                        Episode metadata retrieval not available for this entry.
                    </p>
                )}


            </MediaPageHeaderDetailsContainer>

        </MediaPageHeader>

    )

}
