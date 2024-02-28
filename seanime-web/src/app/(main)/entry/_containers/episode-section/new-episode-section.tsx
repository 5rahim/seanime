"use client"
import { EpisodeListGrid } from "@/app/(main)/entry/_components/episode-list-grid"
import { EpisodeItem } from "@/app/(main)/entry/_containers/episode-section/episode-item"
import { UndownloadedEpisodeList } from "@/app/(main)/entry/_containers/episode-section/undownloaded-episode-list"
import { RelationsRecommendationsSection } from "@/app/(main)/entry/_containers/meta-section/_components/relations-recommendations-accordion"
import { useMediaPlayer, usePlayNextVideoOnMount } from "@/app/(main)/entry/_lib/media-player"
import { SliderEpisodeItem } from "@/components/shared/slider-episode-item"
import { Alert } from "@/components/ui/alert"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselItem } from "@/components/ui/carousel"
import { MediaDetailsByIdQuery } from "@/lib/anilist/gql/graphql"
import { MediaEntry } from "@/lib/server/types"
import React, { useMemo } from "react"


export function NewEpisodeSection(props: { entry: MediaEntry, details: MediaDetailsByIdQuery["Media"] }) {
    const { entry, details } = props
    const media = entry.media

    const { playVideo } = useMediaPlayer()

    usePlayNextVideoOnMount({
        onPlay: () => {
            if (entry.nextEpisode) {
                playVideo({ path: entry.nextEpisode.localFile?.path ?? "" })
            }
        },
    })

    const mainEpisodes = useMemo(() => {
        return entry.episodes?.filter(ep => ep.type === "main") ?? []
    }, [entry.episodes])

    const specialEpisodes = useMemo(() => {
        return entry.episodes?.filter(ep => ep.type === "special") ?? []
    }, [entry.episodes])

    const ncEpisodes = useMemo(() => {
        return entry.episodes?.filter(ep => ep.type === "nc") ?? []
    }, [entry.episodes])

    const hasInvalidEpisodes = useMemo(() => {
        return entry.episodes?.some(ep => ep.isInvalid) ?? false
    }, [entry.episodes])

    const episodesToWatch = useMemo(() => {
        const ret = mainEpisodes.filter(ep => {
            if (!entry.nextEpisode) {
                return true
            } else {
                return ep.progressNumber > (entry.listData?.progress ?? 0)
            }
        })
        return (!!entry.listData?.progress && !entry.nextEpisode) ? ret.reverse() : ret
    }, [mainEpisodes, entry.nextEpisode, entry.listData?.progress])

    if (!media) return null

    if (!!media && (!entry.listData || !entry.libraryData)) {
        return <div className="space-y-10">
            {media?.status !== "NOT_YET_RELEASED" ? <p>Not in your library</p> : <p>Not yet released</p>}
            <div className="overflow-y-auto pt-4 lg:pt-0 space-y-10">
                <UndownloadedEpisodeList
                    downloadInfo={entry.downloadInfo}
                    media={media}
                />
                <RelationsRecommendationsSection entry={entry} details={details} />
            </div>
        </div>
    }

    return (
        <>
            <AppLayoutStack spacing="lg">

                {!entry.episodes?.length && <p>Not in your library</p>}

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
                                        className="md:basis-1/2 lg:basis-1/2 2xl:basis-1/3 min-[2000px]:basis-1/4"
                                    >
                                        <SliderEpisodeItem
                                            key={episode.localFile?.path || ""}
                                            episode={episode}
                                            onPlay={playVideo}
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
                                onPlay={playVideo}
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
                                    onPlay={playVideo}
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
                                    onPlay={playVideo}
                                />
                            ))}
                        </EpisodeListGrid>
                    </>}

                    <RelationsRecommendationsSection entry={entry} details={details} />

                </div>
            </AppLayoutStack>
        </>
    )

}
