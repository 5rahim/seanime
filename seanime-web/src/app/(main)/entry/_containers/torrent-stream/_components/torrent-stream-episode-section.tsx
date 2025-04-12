import { Anime_Entry, Anime_Episode, Torrentstream_EpisodeCollection } from "@/api/generated/types"
import { getEpisodeMinutesRemaining, getEpisodePercentageComplete, useGetContinuityWatchHistory } from "@/api/hooks/continuity.hooks"
import { EpisodeCard } from "@/app/(main)/_features/anime/_components/episode-card"
import { EpisodeGridItem } from "@/app/(main)/_features/anime/_components/episode-grid-item"
import { MediaEpisodeInfoModal } from "@/app/(main)/_features/media/_components/media-episode-info-modal"
import { EpisodeListGrid } from "@/app/(main)/entry/_components/episode-list-grid"
import { usePlayNextVideoOnMount } from "@/app/(main)/entry/_lib/handle-play-on-mount"
import { episodeCardCarouselItemClass } from "@/components/shared/classnames"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselItem } from "@/components/ui/carousel"
import { useThemeSettings } from "@/lib/theme/hooks"
import React, { useMemo } from "react"

type TorrentStreamEpisodeSectionProps = {
    entry: Anime_Entry
    episodeCollection: Torrentstream_EpisodeCollection | undefined
    onEpisodeClick: (episode: Anime_Episode) => void
    onPlayNextEpisodeOnMount: (episode: Anime_Episode) => void
    bottomSection?: React.ReactNode
}

export function TorrentStreamEpisodeSection(props: TorrentStreamEpisodeSectionProps) {
    const ts = useThemeSettings()

    const {
        entry,
        episodeCollection,
        onEpisodeClick,
        onPlayNextEpisodeOnMount,
        bottomSection,
        ...rest
    } = props

    const { data: watchHistory } = useGetContinuityWatchHistory()

    /**
     * Organize episodes to watch
     */
    const episodesToWatch = useMemo(() => {
        if (!episodeCollection?.episodes) return []
        let ret = [...episodeCollection?.episodes]
        ret = ((!!entry.listData?.progress && !!entry.media?.episodes && entry.listData?.progress === entry.media?.episodes)
                ? ret?.reverse()
                : ret?.slice(entry.listData?.progress || 0)
        )?.slice(0, 30) || []
        return ret
    }, [episodeCollection?.episodes, entry.nextEpisode, entry.listData?.progress])

    /**
     * Play next episode on mount if requested
     */
    usePlayNextVideoOnMount({
        onPlay: () => {
            onPlayNextEpisodeOnMount(episodesToWatch[0])
        },
    }, !!episodesToWatch[0])

    if (!entry || !episodeCollection) return null

    return (
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
                                // meta={episode.episodeMetadata?.airDate ?? undefined}
                                isInvalid={episode.isInvalid}
                                progressTotal={episode.baseAnime?.episodes}
                                progressNumber={episode.progressNumber}
                                episodeNumber={episode.episodeNumber}
                                length={episode.episodeMetadata?.length}
                                percentageComplete={getEpisodePercentageComplete(watchHistory, entry.mediaId, episode.episodeNumber)}
                                minutesRemaining={getEpisodeMinutesRemaining(watchHistory, entry.mediaId, episode.episodeNumber)}
                                hasDiscrepancy={episodeCollection?.episodes?.findIndex(e => e.type === "special") !== -1}
                                onClick={() => {
                                    onEpisodeClick(episode)
                                }}
                                anime={{
                                    id: entry.mediaId,
                                    image: episode.baseAnime?.coverImage?.medium,
                                    title: episode?.baseAnime?.title?.userPreferred,
                                }}
                            />
                        </CarouselItem>
                    ))}
                </CarouselContent>
            </Carousel>

            <EpisodeListGrid>
                {episodeCollection?.episodes?.map(episode => (
                    <EpisodeGridItem
                        key={episode.episodeNumber + episode.displayTitle}
                        media={episode?.baseAnime as any}
                        title={episode?.displayTitle || episode?.baseAnime?.title?.userPreferred || ""}
                        image={episode?.episodeMetadata?.image || episode?.baseAnime?.coverImage?.large}
                        episodeTitle={episode?.episodeTitle}
                        onClick={() => {
                            onEpisodeClick(episode)
                        }}
                        description={episode?.episodeMetadata?.overview}
                        isFiller={episode?.episodeMetadata?.isFiller}
                        length={episode?.episodeMetadata?.length}
                        isWatched={!!entry.listData?.progress && entry.listData.progress >= episode?.progressNumber}
                        className="flex-none w-full"
                        episodeNumber={episode.episodeNumber}
                        progressNumber={episode.progressNumber}
                        action={<>
                            <MediaEpisodeInfoModal
                                title={episode.displayTitle}
                                image={episode.episodeMetadata?.image}
                                episodeTitle={episode.episodeTitle}
                                airDate={episode.episodeMetadata?.airDate}
                                length={episode.episodeMetadata?.length}
                                summary={episode.episodeMetadata?.overview}
                                isInvalid={episode.isInvalid}
                            />
                        </>}
                    />
                ))}
            </EpisodeListGrid>

            {bottomSection}
        </>
    )
}
