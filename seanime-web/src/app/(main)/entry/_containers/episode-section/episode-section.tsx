"use client"
import { EpisodeListGrid } from "@/app/(main)/entry/_components/episode-list-grid"
import { BulkToggleLockButton } from "@/app/(main)/entry/_containers/episode-section/bulk-toggle-lock-button"
import { EpisodeItem } from "@/app/(main)/entry/_containers/episode-section/episode-item"
import { EpisodeSectionDropdownMenu } from "@/app/(main)/entry/_containers/episode-section/episode-section-dropdown-menu"
import { UndownloadedEpisodeList } from "@/app/(main)/entry/_containers/episode-section/undownloaded-episode-list"
import { useMediaPlayer, usePlayNextVideoOnMount } from "@/app/(main)/entry/_lib/media-player"
import { SliderEpisodeItem } from "@/components/shared/slider-episode-item"
import { Alert } from "@/components/ui/alert"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselItem } from "@/components/ui/carousel"
import { Separator } from "@/components/ui/separator"
import { MediaEntry } from "@/lib/server/types"
import React, { useMemo } from "react"
import { FiPlayCircle } from "react-icons/fi"
import { IoLibrarySharp } from "react-icons/io5"

export function EpisodeSection(props: { entry: MediaEntry }) {
    const { entry } = props
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
            {media?.status !== "NOT_YET_RELEASED"
                ? <h4 className="text-yellow-50 flex items-center gap-2"><IoLibrarySharp /> Not in your library</h4>
                : <h5 className="text-yellow-50">Not yet released</h5>}
            <div className="overflow-y-auto pt-4 lg:pt-0 space-y-10">
                <UndownloadedEpisodeList
                    downloadInfo={entry.downloadInfo}
                    media={media}
                />
            </div>
        </div>
    }

    return (
        <>
            <AppLayoutStack spacing="lg">

                <div className="mb-8 mt-8 flex flex-col md:flex-row md:items-center justify-between">

                    <div className="flex flex-col md:flex-row items-center gap-4 md:gap-8">
                        <h2>{media.format === "MOVIE" ? "Movie" : "Episodes"}</h2>
                        {!!entry.nextEpisode && <>
                            <Button
                                size="lg"
                                intent="white"
                                rightIcon={<FiPlayCircle/>}
                                iconClass="text-2xl"
                                onClick={() => playVideo({ path: entry.nextEpisode?.localFile?.path ?? "" })}
                            >
                                {media.format === "MOVIE" ? "Watch" : "Play next episode"}
                            </Button>
                        </>}
                    </div>

                    {!!entry.libraryData && <div className="space-x-4 flex justify-center items-center mt-4 md:mt-0">
                        {/*<ProgressTracking entry={entry}/>*/}
                        <BulkToggleLockButton entry={entry}/>
                        <EpisodeSectionDropdownMenu entry={entry} />
                    </div>}

                </div>

                {hasInvalidEpisodes && <Alert
                    intent="alert"
                    description="Some episodes are invalid. Update the metadata to fix this."
                />}


                {episodesToWatch.length > 0 && (
                    <>
                        <Carousel
                            className="w-full max-w-full pt-4 relative"
                            gap="md"
                            opts={{
                                align: "start",
                            }}
                        >
                            <CarouselDotButtons className="-top-3" />
                            <CarouselContent>
                                {episodesToWatch.map((episode, idx) => (
                                    <CarouselItem
                                        key={episode?.localFile?.path || idx}
                                        className="md:basis-1/2 lg:basis-1/3 2xl:basis-1/2 min-[2000px]:basis-1/3"
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
                        <Separator />
                        <h3>Specials</h3>
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
                        <Separator />
                        <h3>Others</h3>
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

                </div>
            </AppLayoutStack>
        </>
    )

}
