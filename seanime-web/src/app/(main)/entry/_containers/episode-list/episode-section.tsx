"use client"
import { AL_AnimeDetailsById_Media, Anime_AnimeEntry } from "@/api/generated/types"
import { getEpisodeMinutesRemaining, getEpisodePercentageComplete, useGetContinuityWatchHistory } from "@/api/hooks/continuity.hooks"
import { EpisodeCard } from "@/app/(main)/_features/anime/_components/episode-card"
import { EpisodeListGrid } from "@/app/(main)/entry/_components/episode-list-grid"
import { EpisodeItem } from "@/app/(main)/entry/_containers/episode-list/episode-item"
import { UndownloadedEpisodeList } from "@/app/(main)/entry/_containers/episode-list/undownloaded-episode-list"
import { useHandleEpisodeSection } from "@/app/(main)/entry/_lib/handle-episode-section"
import { episodeCardCarouselItemClass } from "@/components/shared/classnames"
import { Alert } from "@/components/ui/alert"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselItem } from "@/components/ui/carousel"
import { useThemeSettings } from "@/lib/theme/hooks"
import React from "react"
import { IoLibrarySharp } from "react-icons/io5"

type EpisodeSectionProps = {
    entry: Anime_AnimeEntry
    details: AL_AnimeDetailsById_Media | undefined
    bottomSection: React.ReactNode
}

export function EpisodeSection({ entry, details, bottomSection }: EpisodeSectionProps) {
    const ts = useThemeSettings()

    const {
        media,
        hasInvalidEpisodes,
        episodesToWatch,
        mainEpisodes,
        specialEpisodes,
        ncEpisodes,
        playMediaFile,
    } = useHandleEpisodeSection({ entry, details })

    const { data: watchHistory } = useGetContinuityWatchHistory()

    if (!media) return null

    if (!!media && (!entry.listData || !entry.libraryData)) {
        return <div className="space-y-10">
            {media?.status !== "NOT_YET_RELEASED"
                ? <h4 className="text-yellow-50 flex items-center gap-2"><IoLibrarySharp /> Not in your library</h4>
                : <h5 className="text-yellow-50">Not yet released</h5>}
            <div className="overflow-y-auto pt-4 lg:pt-0 space-y-10">
                <UndownloadedEpisodeList
                    downloadInfo={entry.downloadInfo}
                    media={media}
                />
                {bottomSection}
            </div>
        </div>
    }

    return (
        <>
            <AppLayoutStack spacing="lg">

                {hasInvalidEpisodes && <Alert
                    intent="alert"
                    description="Some episodes are invalid. Update the metadata to fix this."
                />}


                {episodesToWatch.length > 0 && (
                    <>
                        <Carousel
                            className="w-full max-w-full"
                            gap="md"
                            opts={{
                                align: "start",
                            }}
                        >
                            <CarouselDotButtons />
                            <CarouselContent>
                                {episodesToWatch.map((episode, idx) => (
                                    <CarouselItem
                                        key={episode?.localFile?.path || idx}
                                        className={episodeCardCarouselItemClass(ts.smallerEpisodeCarouselSize)}
                                    >
                                        <EpisodeCard
                                            key={episode.localFile?.path || ""}
                                            image={episode.episodeMetadata?.image || episode.baseAnime?.bannerImage || episode.baseAnime?.coverImage?.extraLarge}
                                            topTitle={episode.episodeTitle || episode?.baseAnime?.title?.userPreferred}
                                            title={episode.displayTitle}
                                            isInvalid={episode.isInvalid}
                                            progressTotal={episode.baseAnime?.episodes}
                                            progressNumber={episode.progressNumber}
                                            episodeNumber={episode.episodeNumber}
                                            length={episode.episodeMetadata?.length}
                                            percentageComplete={getEpisodePercentageComplete(watchHistory, entry.mediaId, episode.episodeNumber)}
                                            minutesRemaining={getEpisodeMinutesRemaining(watchHistory, entry.mediaId, episode.episodeNumber)}
                                            onClick={() => playMediaFile({ path: episode.localFile?.path ?? "", mediaId: entry.mediaId })}
                                        />
                                    </CarouselItem>
                                ))}
                            </CarouselContent>
                        </Carousel>
                    </>
                )}


                <div className="space-y-10">
                    <EpisodeListGrid>
                        {mainEpisodes.map(episode => (
                            <EpisodeItem
                                key={episode.localFile?.path || ""}
                                episode={episode}
                                media={media}
                                isWatched={!!entry.listData?.progress && entry.listData.progress >= episode.progressNumber}
                                onPlay={playMediaFile}
                                percentageComplete={getEpisodePercentageComplete(watchHistory, entry.mediaId, episode.episodeNumber)}
                                minutesRemaining={getEpisodeMinutesRemaining(watchHistory, entry.mediaId, episode.episodeNumber)}
                            />
                        ))}
                    </EpisodeListGrid>

                    <UndownloadedEpisodeList
                        downloadInfo={entry.downloadInfo}
                        media={media}
                    />

                    {specialEpisodes.length > 0 && <>
                        <h2>Specials</h2>
                        <EpisodeListGrid>
                            {specialEpisodes.map(episode => (
                                <EpisodeItem
                                    key={episode.localFile?.path || ""}
                                    episode={episode}
                                    media={media}
                                    onPlay={playMediaFile}
                                />
                            ))}
                        </EpisodeListGrid>
                    </>}

                    {ncEpisodes.length > 0 && <>
                        <h2>Others</h2>
                        <EpisodeListGrid>
                            {ncEpisodes.map(episode => (
                                <EpisodeItem
                                    key={episode.localFile?.path || ""}
                                    episode={episode}
                                    media={media}
                                    onPlay={playMediaFile}
                                />
                            ))}
                        </EpisodeListGrid>
                    </>}

                    {bottomSection}

                </div>
            </AppLayoutStack>
        </>
    )

}
