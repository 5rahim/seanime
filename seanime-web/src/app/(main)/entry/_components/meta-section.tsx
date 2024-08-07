"use client"
import { AL_AnimeDetailsById_Media, Anime_AnimeEntry } from "@/api/generated/types"
import { TrailerModal } from "@/app/(main)/_features/anime/_components/trailer-modal"
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
import { useHasTorrentProvider } from "@/app/(main)/_hooks/use-server-status"
import { EntryOnlinestreamButton } from "@/app/(main)/entry/_components/entry-onlinestream-button"
import { NextAiringEpisode } from "@/app/(main)/entry/_components/next-airing-episode"
import { __anime_torrentStreamingViewActiveAtom } from "@/app/(main)/entry/_containers/anime-entry-page"
import { AnimeEntryDropdownMenu } from "@/app/(main)/entry/_containers/entry-actions/anime-entry-dropdown-menu"
import { AnimeEntrySilenceToggle } from "@/app/(main)/entry/_containers/entry-actions/anime-entry-silence-toggle"
import { TorrentSearchButton } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-button"
import { TorrentStreamButton } from "@/app/(main)/entry/_containers/torrent-stream/torrent-stream-button"
import { Button } from "@/components/ui/button"
import { useAtomValue } from "jotai"
import Link from "next/link"
import React from "react"


export function MetaSection(props: { entry: Anime_AnimeEntry, details: AL_AnimeDetailsById_Media | undefined }) {

    const { entry, details } = props

    if (!entry.media) return null

    const { hasTorrentProvider } = useHasTorrentProvider()
    const isTorrentStreamingView = useAtomValue(__anime_torrentStreamingViewActiveAtom)

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
                        && hasTorrentProvider
                    ) && (
                        <TorrentSearchButton
                            entry={entry}
                        />
                    )}

                    {entry.media.status !== "NOT_YET_RELEASED" && (
                        <TorrentStreamButton
                            entry={entry}
                        />
                    )}
                </div>

                <NextAiringEpisode media={entry.media} />

                <div className="w-full flex justify-between flex-wrap gap-4 items-center">

                    <EntryOnlinestreamButton entry={entry} />

                    <Link href={`https://anilist.co/anime/${entry.mediaId}`} target="_blank">
                        <Button intent="gray-link" className="px-0">
                            AniList
                        </Button>
                    </Link>

                    {!!entry?.media?.trailer?.id && <TrailerModal
                        trailerId={entry?.media?.trailer?.id} trigger={
                        <Button intent="gray-link" className="px-0">
                            Trailer
                        </Button>}
                    />}


                    <div className="flex flex-1"></div>

                    {!!entry.libraryData && <>
                        <AnimeEntrySilenceToggle mediaId={entry.mediaId} />
                        <ToggleLockFilesButton
                            allFilesLocked={entry.libraryData.allFilesLocked}
                            mediaId={entry.mediaId}
                            size="lg"
                        />
                    </>}
                    <AnimeEntryDropdownMenu entry={entry} />
                </div>

                {(!entry.aniDBId || entry.aniDBId === 0) && (
                    <p className="text-center text-red-300 opacity-50">
                        No metadata found on AniDB
                    </p>
                )}


            </MediaPageHeaderDetailsContainer>

        </MediaPageHeader>

    )

}
