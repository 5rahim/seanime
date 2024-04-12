"use client"

import { MediaEntryEpisode } from "@/app/(main)/(library)/_lib/anime-library.types"
import { OfflineMetaSection } from "@/app/(main)/(offline)/offline/(entry)/_components/offline-meta-section"
import { useOfflineSnapshot } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot-context"
import { OfflineAnimeEntry, OfflineAssetMap } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.types"
import { offline_getAssetUrl } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.utils"
import { EpisodeItemIsolation } from "@/app/(main)/entry/_containers/episode-section/episode-item"
import { useMediaPlayer } from "@/app/(main)/entry/_lib/media-player"
import { EpisodeListItem } from "@/components/shared/episode-list-item"
import { LuffyError } from "@/components/shared/luffy-error"
import { SliderEpisodeItem } from "@/components/shared/slider-episode-item"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import { IconButton } from "@/components/ui/button"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselItem } from "@/components/ui/carousel"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import Image from "next/image"
import { useRouter, useSearchParams } from "next/navigation"
import React, { memo } from "react"
import { AiFillWarning } from "react-icons/ai"
import { MdInfo } from "react-icons/md"

export default function Page() {

    const router = useRouter()
    const mediaId = useSearchParams().get("id")
    const { snapshot, isLoading } = useOfflineSnapshot()

    const entry = React.useMemo(() => {
        return snapshot?.entries?.animeEntries?.find(n => n?.mediaId === Number(mediaId))
    }, [snapshot, mediaId])

    if (isLoading) return <LoadingSpinner />

    if (!entry) return <LuffyError title="Not found" />

    return (
        <>
            <OfflineMetaSection type="anime" entry={entry} assetMap={snapshot?.assetMap} />
            <PageWrapper className="p-4 relative">
                <EpisodeLists entry={entry} assetMap={snapshot?.assetMap} />
            </PageWrapper>
        </>
    )

}

type EpisodeListsProps = {
    children?: React.ReactNode
    entry: OfflineAnimeEntry
    assetMap: OfflineAssetMap | undefined
}

export function EpisodeLists(props: EpisodeListsProps) {

    const {
        children,
        entry,
        assetMap,
        ...rest
    } = props

    const episodes = React.useMemo(() => {
        if (!entry.episodes) return []

        return entry.episodes.filter(Boolean).map(ep => {
            return {
                ...ep,
                episodeMetadata: {
                    ...ep.episodeMetadata,
                    image: offline_getAssetUrl(ep.episodeMetadata?.image, assetMap),
                },
            }
        })
    }, [entry.episodes, assetMap])

    const mainEpisodes = React.useMemo(() => {
        return episodes.filter(ep => ep.type === "main") ?? []
    }, [episodes])

    const specialEpisodes = React.useMemo(() => {
        return episodes.filter(ep => ep.type === "special") ?? []
    }, [episodes])

    const ncEpisodes = React.useMemo(() => {
        return episodes.filter(ep => ep.type === "nc") ?? []
    }, [episodes])

    const episodesToWatch = React.useMemo(() => {
        return mainEpisodes.filter(ep => {
            return ep.progressNumber > (entry.listData?.progress ?? 0)
        })
    }, [mainEpisodes, entry.listData?.progress])

    const { playVideo } = useMediaPlayer()

    return (
        <div className="space-y-10">
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

            <div className="space-y-10 pb-10">
                <h2>Episodes</h2>
                <div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3 min-[2000px]:grid-cols-4">
                    {mainEpisodes.map(episode => (
                        <EpisodeItem
                            key={episode.localFile?.path || ""}
                            episode={episode}
                            media={entry.media!}
                            isWatched={!!entry.listData?.progress && entry.listData.progress >= episode.progressNumber}
                            onPlay={playVideo}
                        />
                    ))}
                </div>

                {specialEpisodes.length > 0 && <>
                    <h2>Specials</h2>
                    <div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3 min-[2000px]:grid-cols-4">
                        {specialEpisodes.map(episode => (
                            <EpisodeItem
                                key={episode.localFile?.path || ""}
                                episode={episode}
                                media={entry.media!}
                                onPlay={playVideo}
                            />
                        ))}
                    </div>
                </>}

                {ncEpisodes.length > 0 && <>
                    <h2>Others</h2>
                    <div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3 min-[2000px]:grid-cols-4">
                        {ncEpisodes.map(episode => (
                            <EpisodeItem
                                key={episode.localFile?.path || ""}
                                episode={episode}
                                media={entry.media!}
                                onPlay={playVideo}
                            />
                        ))}
                    </div>
                </>}

            </div>
        </div>
    )
}

const EpisodeItem = memo(({ episode, media, isWatched, onPlay }: {
    episode: MediaEntryEpisode,
    media: BaseMediaFragment,
    onPlay: ({ path }: { path: string }) => void,
    isWatched?: boolean
}) => {

    return (
        <EpisodeItemIsolation.Provider>
            <EpisodeListItem
                media={media}
                image={episode.episodeMetadata?.image}
                onClick={() => onPlay({ path: episode.localFile?.path ?? "" })}
                isInvalid={episode.isInvalid}
                title={episode.displayTitle}
                episodeTitle={episode.episodeTitle}
                fileName={episode.localFile?.name}
                isWatched={episode.progressNumber > 0 && isWatched}
                action={<>
                    <Modal
                        trigger={<IconButton
                            icon={<MdInfo />}
                            className="opacity-30 hover:opacity-100 transform-opacity"
                            intent="gray-basic"
                            size="xs"
                        />}
                        title={episode.displayTitle}
                        contentClass="max-w-2xl"
                        titleClass="text-xl"
                    >

                        {episode.episodeMetadata?.image && <div
                            className="h-[8rem] w-full flex-none object-cover object-center overflow-hidden absolute left-0 top-0 z-[-1]"
                        >
                            <Image
                                src={episode.episodeMetadata?.image}
                                alt="banner"
                                fill
                                quality={80}
                                priority
                                sizes="20rem"
                                className="object-cover object-center opacity-30"
                            />
                            <div
                                className="z-[5] absolute bottom-0 w-full h-[80%] bg-gradient-to-t from-[--background] to-transparent"
                            />
                        </div>}

                        <div className="space-y-4">
                            <p className="text-lg line-clamp-2 font-semibold">
                                {episode.episodeTitle}
                                {episode.isInvalid && <AiFillWarning />}
                            </p>
                            <p className="text-[--muted]">
                                {episode.episodeMetadata?.airDate || "Unknown airing date"} - {episode.episodeMetadata?.length || "N/A"} minutes
                            </p>
                            <p className="text-[--muted]">
                                {episode.episodeMetadata?.summary?.replaceAll("`", "'") || "No summary"}
                            </p>
                        </div>

                    </Modal>
                </>}
            />
        </EpisodeItemIsolation.Provider>
    )

})
