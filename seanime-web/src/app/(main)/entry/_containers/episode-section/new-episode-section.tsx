"use client"
import { EpisodeListGrid } from "@/app/(main)/entry/_components/episode-list-grid"
import { EpisodeItem } from "@/app/(main)/entry/_containers/episode-section/episode-item"
import { UndownloadedEpisodeList } from "@/app/(main)/entry/_containers/episode-section/undownloaded-episode-list"
import { RelationsRecommendationsSection } from "@/app/(main)/entry/_containers/meta-section/_components/relations-recommendations-accordion"
import { useMediaPlayer, usePlayNextVideoOnMount } from "@/app/(main)/entry/_lib/media-player"
import { imageShimmer } from "@/components/shared/image-helpers"
import { Alert } from "@/components/ui/alert"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselItem } from "@/components/ui/carousel"
import { cn } from "@/components/ui/core/styling"
import { MediaDetailsByIdQuery } from "@/lib/anilist/gql/graphql"
import { MediaEntry, MediaEntryEpisode } from "@/lib/server/types"
import { isBefore, subYears } from "date-fns"
import Image from "next/image"
import React, { memo, useMemo } from "react"
import { AiFillPlayCircle } from "react-icons/ai"


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

                {/*<div className="mb-8 flex flex-col md:flex-row md:items-center justify-between">*/}

                {/*    <div className="flex flex-col md:flex-row items-center gap-4 md:gap-8">*/}
                {/*        <h2>{media.format === "MOVIE" ? "Movie" : "Episodes"}</h2>*/}
                {/*        {!!entry.nextEpisode && <>*/}
                {/*            <Button*/}
                {/*                size="lg"*/}
                {/*                intent="white"*/}
                {/*                rightIcon={<FiPlayCircle/>}*/}
                {/*                iconClass="text-2xl"*/}
                {/*                onClick={() => playVideo({ path: entry.nextEpisode?.localFile?.path ?? "" })}*/}
                {/*            >*/}
                {/*                {media.format === "MOVIE" ? "Watch" : "Play next episode"}*/}
                {/*            </Button>*/}
                {/*        </>}*/}
                {/*    </div>*/}

                {/*    {!!entry.libraryData && <div className="space-x-4 flex justify-center items-center mt-4 md:mt-0">*/}
                {/*        <ProgressTracking entry={entry}/>*/}
                {/*        <BulkToggleLockButton entry={entry}/>*/}
                {/*        <EpisodeSectionDropdownMenu entry={entry} />*/}
                {/*    </div>}*/}

                {/*</div>*/}

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


const SliderEpisodeItem = memo(({ episode, onPlay }: {
    episode: MediaEntryEpisode,
    onPlay: ({ path }: { path: string }) => void
}) => {

    const date = episode.episodeMetadata?.airDate ? new Date(episode.episodeMetadata.airDate) : undefined
    const mediaIsOlder = useMemo(() => date ? isBefore(date, subYears(new Date(), 2)) : undefined, [])
    const offset = episode.progressNumber - episode.episodeNumber
    const title = episode.displayTitle

    return (
        <div
            className={cn(
                "rounded-md border overflow-hidden aspect-[4/2] relative flex items-end flex-none group/missed-episode-item cursor-pointer",
                "user-select-none",
                "w-full",
            )}
            onClick={() => onPlay({ path: episode.localFile?.path ?? "" })}
        >
            <div className="absolute w-full h-full overflow-hidden z-[1]">
                {!!episode.episodeMetadata?.image ? <Image
                    src={episode.episodeMetadata?.image}
                    alt={""}
                    fill
                    quality={100}
                    placeholder={imageShimmer(700, 475)}
                    sizes="20rem"
                    className="object-cover object-center transition"
                /> : <div
                    className="h-full block absolute w-full bg-gradient-to-t from-gray-800 to-transparent z-[2]"
                ></div>}
                <div
                    className="z-[1] absolute bottom-0 w-full h-full md:h-[80%] bg-gradient-to-t from-[--background] to-transparent"
                />
            </div>
            <div
                className={cn(
                    "group-hover/missed-episode-item:opacity-100 text-6xl text-gray-200",
                    "cursor-pointer opacity-0 transition-opacity bg-gray-950 bg-opacity-60 z-[2] absolute w-[105%] h-[105%] items-center justify-center",
                    "hidden md:flex",
                )}
            >
                <AiFillPlayCircle className="opacity-50" />
            </div>
            <div className="relative z-[3] w-full p-4 space-y-1">
                <p className="w-[80%] line-clamp-1 text-[--muted] font-semibold">{episode.episodeTitle}</p>
                <div className="w-full justify-between flex items-center">
                    <p className="text-base md:text-xl lg:text-2xl font-semibold line-clamp-2">
                        <span>{episode.displayTitle} {!!episode.basicMedia?.episodes &&
                            (episode.basicMedia.episodes != 1 &&
                                <span className="opacity-40">/{` `}{episode.basicMedia.episodes - offset}</span>)}
                        </span>
                    </p>
                    <p className="text-[--muted] text-sm md:text-base">{episode.episodeMetadata?.length + "m" || ""}</p>
                </div>
                {episode.isInvalid && <p className="text-red-300">No metadata found</p>}
            </div>
        </div>
    )
})
