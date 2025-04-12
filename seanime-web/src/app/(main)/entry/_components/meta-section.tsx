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
import { AnimeOnlinestreamButton } from "@/app/(main)/entry/_components/anime-onlinestream-button"
import { NextAiringEpisode } from "@/app/(main)/entry/_components/next-airing-episode"
import { useAnimeEntryPageView } from "@/app/(main)/entry/_containers/anime-entry-page"
import { DebridStreamButton } from "@/app/(main)/entry/_containers/debrid-stream/debrid-stream-button"
import { AnimeEntryDropdownMenu } from "@/app/(main)/entry/_containers/entry-actions/anime-entry-dropdown-menu"
import { AnimeEntrySilenceToggle } from "@/app/(main)/entry/_containers/entry-actions/anime-entry-silence-toggle"
import { TorrentSearchButton } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-button"
import { TorrentStreamButton } from "@/app/(main)/entry/_containers/torrent-stream/torrent-stream-button"
import { SeaLink } from "@/components/shared/sea-link"
import { Button, ButtonProps, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { TORRENT_CLIENT } from "@/lib/server/settings"
import { ThemeMediaPageInfoBoxSize, useThemeSettings } from "@/lib/theme/hooks"
import React from "react"
import { SiAnilist } from "react-icons/si"
import { PluginAnimePageButtons } from "../../_features/plugin/actions/plugin-actions"

export function AnimeMetaActionButton({ className, ...rest }: ButtonProps) {
    const ts = useThemeSettings()
    return <Button
        className={cn(
            "w-full",
            ts.mediaPageBannerInfoBoxSize === ThemeMediaPageInfoBoxSize.Fluid && "lg:w-full lg:max-w-[280px]",
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

    if (!entry.media) return null

    const { hasTorrentProvider } = useHasTorrentProvider()
    const { hasDebridService } = useHasDebridService()
    const { currentView, isTorrentStreamingView, isDebridStreamingView, isOnlineStreamingView } = useAnimeEntryPageView()

    const ActionButtons = () => (
        <div
            data-anime-meta-section-action-buttons
            className={cn(
                "w-full flex flex-wrap gap-4 items-center",
                ts.mediaPageBannerInfoBoxSize === ThemeMediaPageInfoBoxSize.Fluid && "w-auto flex-nowrap",
            )}
        >

            <div className="flex items-center gap-4 justify-center w-full lg:w-fit" data-anime-meta-section-action-buttons-inner-container>

                <SeaLink href={`https://anilist.co/anime/${entry.mediaId}`} target="_blank">
                    <IconButton intent="gray-link" className="px-0" icon={<SiAnilist className="text-lg" />} />
                </SeaLink>

                {!!entry?.media?.trailer?.id && <TrailerModal
                    trailerId={entry?.media?.trailer?.id} trigger={
                    <Button intent="gray-link" className="px-0">
                        Trailer
                    </Button>}
                />}
            </div>

            {ts.mediaPageBannerInfoBoxSize !== ThemeMediaPageInfoBoxSize.Fluid &&
                <div className="flex-1 hidden lg:flex" data-anime-meta-section-action-buttons-spacer></div>}

            <div className="flex items-center gap-4 justify-center w-full lg:w-fit" data-anime-meta-section-action-buttons-inner-container>
                <AnimeAutoDownloaderButton entry={entry} size="md" />

                {!!entry.libraryData && <>
                    <MediaSyncTrackButton mediaId={entry.mediaId} type="anime" size="md" />
                    <AnimeEntrySilenceToggle mediaId={entry.mediaId} size="md" />
                    <ToggleLockFilesButton
                        allFilesLocked={entry.libraryData.allFilesLocked}
                        mediaId={entry.mediaId}
                        size="md"
                    />
                </>}
                <AnimeEntryDropdownMenu entry={entry} />
            </div>

            <PluginAnimePageButtons media={entry.media!} />
        </div>
    )

    const Details = () => (
        <div
            data-anime-meta-section-details
            className={cn(
                "flex gap-3 flex-wrap items-center",
                ts.mediaPageBannerInfoBoxSize === ThemeMediaPageInfoBoxSize.Fluid && "justify-center lg:justify-start lg:max-w-[65vw]",
            )}
        >
            <MediaEntryAudienceScore meanScore={details?.meanScore} badgeClass="bg-transparent" />

            <AnimeEntryStudio studios={details?.studios} />

            <MediaEntryGenresList genres={details?.genres} />

            <div
                data-anime-meta-section-rankings-container
                className={cn(
                    ts.mediaPageBannerInfoBoxSize === ThemeMediaPageInfoBoxSize.Fluid ? "w-full" : "contents",
                )}
            >
                <AnimeEntryRankings rankings={details?.rankings} />
            </div>
        </div>
    )

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
                    {ts.mediaPageBannerInfoBoxSize === ThemeMediaPageInfoBoxSize.Fluid && <Details />}
                </MediaPageHeaderEntryDetails>

                {ts.mediaPageBannerInfoBoxSize !== ThemeMediaPageInfoBoxSize.Fluid && <Details />}

                <div
                    data-anime-meta-section-buttons-container
                    className={cn(
                        "flex flex-col lg:flex-row w-full gap-3 items-center",
                        ts.mediaPageBannerInfoBoxSize === ThemeMediaPageInfoBoxSize.Fluid && "flex-wrap",
                    )}
                >

                    {ts.mediaPageBannerInfoBoxSize === ThemeMediaPageInfoBoxSize.Fluid && <ActionButtons />}

                    {(
                        entry.media.status !== "NOT_YET_RELEASED"
                        && currentView === "library"
                        && hasTorrentProvider
                        && (
                            serverStatus?.settings?.torrent?.defaultTorrentClient !== TORRENT_CLIENT.NONE
                            || hasDebridService
                        )
                    ) && (
                        <TorrentSearchButton
                            entry={entry}
                        />
                    )}

                    <TorrentStreamButton
                        entry={entry}
                    />

                    <DebridStreamButton
                        entry={entry}
                    />

                    <AnimeOnlinestreamButton entry={entry} />

                </div>

                <NextAiringEpisode media={entry.media} />

                {entry.downloadInfo?.hasInaccurateSchedule && <p
                    className={cn(
                        "text-[--muted] text-sm text-center mb-3",
                        ts.mediaPageBannerInfoBoxSize === ThemeMediaPageInfoBoxSize.Fluid && "text-left",
                    )}
                    data-anime-meta-section-inaccurate-schedule-message
                >
                    <span className="block">Could not retrieve accurate scheduling information for this show.</span>
                    <span className="block text-[--muted]">Please check the schedule online for more information.</span>
                </p>}

                {ts.mediaPageBannerInfoBoxSize !== ThemeMediaPageInfoBoxSize.Fluid && <ActionButtons />}

                {(!entry.anidbId || entry.anidbId === 0) && (
                    <p
                        className={cn(
                            "text-center text-red-300 opacity-50",
                            ts.mediaPageBannerInfoBoxSize === ThemeMediaPageInfoBoxSize.Fluid && "text-left",
                        )}
                        data-anime-meta-section-no-metadata-message
                    >
                        No metadata found on AniDB
                    </p>
                )}


            </MediaPageHeaderDetailsContainer>

        </MediaPageHeader>

    )

}
