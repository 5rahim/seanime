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
import { EntryOnlinestreamButton } from "@/app/(main)/entry/_components/entry-onlinestream-button"
import { NextAiringEpisode } from "@/app/(main)/entry/_components/next-airing-episode"
import { __anime_debridStreamingViewActiveAtom, __anime_torrentStreamingViewActiveAtom } from "@/app/(main)/entry/_containers/anime-entry-page"
import { DebridStreamButton } from "@/app/(main)/entry/_containers/debrid-stream/debrid-stream-button"
import { AnimeEntryDropdownMenu } from "@/app/(main)/entry/_containers/entry-actions/anime-entry-dropdown-menu"
import { AnimeEntrySilenceToggle } from "@/app/(main)/entry/_containers/entry-actions/anime-entry-silence-toggle"
import { TorrentSearchButton } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-button"
import { TorrentStreamButton } from "@/app/(main)/entry/_containers/torrent-stream/torrent-stream-button"
import { SeaLink } from "@/components/shared/sea-link"
import { Button, IconButton } from "@/components/ui/button"
import { useAtomValue } from "jotai"
import React from "react"
import { SiAnilist } from "react-icons/si"


export function MetaSection(props: { entry: Anime_Entry, details: AL_AnimeDetailsById_Media | undefined }) {
    const serverStatus = useServerStatus()
    const { entry, details } = props

    if (!entry.media) return null

    const { hasTorrentProvider } = useHasTorrentProvider()
    const { hasDebridService } = useHasDebridService()
    const isTorrentStreamingView = useAtomValue(__anime_torrentStreamingViewActiveAtom)
    const isDebridStreamingView = useAtomValue(__anime_debridStreamingViewActiveAtom)

    return (
        <MediaPageHeader
            backgroundImage={entry.media?.bannerImage || entry.media?.coverImage?.extraLarge}
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
                />


                <div className="flex gap-2 flex-wrap items-center">
                    <MediaEntryAudienceScore meanScore={details?.meanScore} />

                    <AnimeEntryStudio studios={details?.studios} />

                    <MediaEntryGenresList genres={details?.genres} />

                    <AnimeEntryRankings rankings={details?.rankings} />
                </div>


                <div className="flex flex-col lg:flex-row w-full gap-3">
                    {(
                        entry.media.status !== "NOT_YET_RELEASED"
                        && !isTorrentStreamingView
                        && !isDebridStreamingView
                        && hasTorrentProvider
                    ) && (
                        <TorrentSearchButton
                            entry={entry}
                        />
                    )}

                    {(entry.media.status !== "NOT_YET_RELEASED"
                        && !isDebridStreamingView
                    ) && (
                        <TorrentStreamButton
                            entry={entry}
                        />
                    )}

                    {(entry.media.status !== "NOT_YET_RELEASED"
                        && !isTorrentStreamingView
                        && hasDebridService
                    ) && (
                        <DebridStreamButton
                            entry={entry}
                        />
                    )}
                </div>

                <NextAiringEpisode media={entry.media} />

                <div className="w-full flex flex-wrap gap-4 items-center">

                    <div className="flex items-center gap-4 justify-center w-full lg:w-fit">
                        <EntryOnlinestreamButton entry={entry} />

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

                    <div className="flex-1 hidden lg:flex"></div>

                    <div className="flex items-center gap-4 justify-center w-full lg:w-fit">
                        <AnimeAutoDownloaderButton entry={entry} size="lg" />

                        {!!entry.libraryData && <>
                            <MediaSyncTrackButton mediaId={entry.mediaId} type="anime" size="lg" />
                            <AnimeEntrySilenceToggle mediaId={entry.mediaId} />
                            <ToggleLockFilesButton
                                allFilesLocked={entry.libraryData.allFilesLocked}
                                mediaId={entry.mediaId}
                                size="lg"
                            />
                        </>}
                        <AnimeEntryDropdownMenu entry={entry} />
                    </div>
                </div>

                {(!entry.anidbId || entry.anidbId === 0) && (
                    <p className="text-center text-red-300 opacity-50">
                        No metadata found on AniDB
                    </p>
                )}


            </MediaPageHeaderDetailsContainer>

        </MediaPageHeader>

    )

}
