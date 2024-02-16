"use client"
import { BulkToggleLockButton } from "@/app/(main)/entry/_containers/episode-section/bulk-toggle-lock-button"
import { EpisodeItem } from "@/app/(main)/entry/_containers/episode-section/episode-item"
import { EpisodeSectionDropdownMenu } from "@/app/(main)/entry/_containers/episode-section/episode-section-dropdown-menu"
import { ProgressTracking } from "@/app/(main)/entry/_containers/episode-section/progress-tracking"
import { UndownloadedEpisodeList } from "@/app/(main)/entry/_containers/episode-section/undownloaded-episode-list"
import { useMediaPlayer, usePlayNextVideoOnMount } from "@/app/(main)/entry/_lib/media-player"
import { LargeEpisodeListItem } from "@/components/shared/large-episode-list-item"
import { Slider } from "@/components/shared/slider"
import { Alert } from "@/components/ui/alert"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { Divider } from "@/components/ui/divider"
import { MediaEntry, MediaEntryEpisode } from "@/lib/server/types"
import { FiPlayCircle } from "@react-icons/all-files/fi/FiPlayCircle"
import { formatDistanceToNow, isBefore, subYears } from "date-fns"
import { memo, useMemo } from "react"

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
        return <div className={"space-y-10"}>
            {media?.status !== "NOT_YET_RELEASED" ? <p>Not in your library</p> : <p>Not yet released</p>}
            <div className={"overflow-y-auto pt-4 lg:pt-0 space-y-10"}>
                <UndownloadedEpisodeList
                    downloadInfo={entry.downloadInfo}
                    media={media}
                />
            </div>
        </div>
    }

    return (
        <>
            <AppLayoutStack spacing={"lg"}>

                <div className={"mb-8 flex flex-col md:flex-row items-center justify-between"}>

                    <div className={"flex items-center gap-8"}>
                        <h2>{media.format === "MOVIE" ? "Movie" : "Episodes"}</h2>
                        {!!entry.nextEpisode && <>
                            <Button
                                size={"lg"}
                                intent={"white"}
                                rightIcon={<FiPlayCircle/>}
                                iconClassName={"text-2xl"}
                                onClick={() => playVideo({ path: entry.nextEpisode?.localFile?.path ?? "" })}
                            >
                                {media.format === "MOVIE" ? "Watch" : "Play next episode"}
                            </Button>
                        </>}
                    </div>

                    {!!entry.libraryData && <div className={"space-x-4 flex items-center"}>
                        <ProgressTracking entry={entry}/>
                        <BulkToggleLockButton entry={entry}/>
                        <EpisodeSectionDropdownMenu entry={entry} />
                    </div>}

                </div>

                {!entry.episodes?.length && <p>Not in your library</p>}

                {hasInvalidEpisodes && <Alert
                    intent="alert"
                    description={"Some episodes are invalid. Update the metadata to fix this."}
                />}


                {episodesToWatch.length > 0 && (
                    <>
                        <Slider>
                            {episodesToWatch.map(episode => (
                                <SliderEpisodeItem
                                    key={episode.localFile?.path || ""}
                                    episode={episode}
                                    onPlay={playVideo}
                                />
                            ))}
                        </Slider>
                        <Divider/>
                    </>
                )}


                <div className="space-y-10">
                    <div className={"grid grid-cols-1 md:grid-cols-2 gap-4"}>
                        {mainEpisodes.map(episode => (
                            <EpisodeItem
                                key={episode.localFile?.path || ""}
                                episode={episode}
                                media={media}
                                isWatched={!!entry.listData?.progress && entry.listData.progress >= episode.progressNumber}
                                onPlay={playVideo}
                            />
                        ))}
                    </div>

                    <UndownloadedEpisodeList
                        downloadInfo={entry.downloadInfo}
                        media={media}
                    />

                    {specialEpisodes.length > 0 && <>
                        <Divider/>
                        <h3>Specials</h3>
                        <div className={"grid grid-cols-1 md:grid-cols-2 gap-4"}>
                            {specialEpisodes.map(episode => (
                                <EpisodeItem
                                    key={episode.localFile?.path || ""}
                                    episode={episode}
                                    media={media}
                                    onPlay={playVideo}
                                />
                            ))}
                        </div>
                    </>}

                    {ncEpisodes.length > 0 && <>
                        <Divider/>
                        <h3>Others</h3>
                        <div className={"grid grid-cols-1 md:grid-cols-2 gap-4"}>
                            {ncEpisodes.map(episode => (
                                <EpisodeItem
                                    key={episode.localFile?.path || ""}
                                    episode={episode}
                                    media={media}
                                    onPlay={playVideo}
                                />
                            ))}
                        </div>
                    </>}

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

    return (
        <LargeEpisodeListItem
            image={episode.episodeMetadata?.image}
            title={<span>{episode.displayTitle} {!!episode.basicMedia?.episodes &&
                (episode.basicMedia.episodes != 1 &&
                    <span className={"opacity-40"}>/{` `}{episode.basicMedia.episodes - offset}</span>)}</span>}
            topTitle={episode.episodeTitle}
            actionIcon={undefined}
            meta={(date) ? (!mediaIsOlder ? `${formatDistanceToNow(date, { addSuffix: true })}` : new Intl.DateTimeFormat("en-US", {
                day: "2-digit",
                month: "2-digit",
                year: "2-digit",
            }).format(date)) : undefined}
            onClick={() => onPlay({ path: episode.localFile?.path ?? "" })}
        />
    )
})
