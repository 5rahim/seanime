"use client"

import { AL_BaseMedia, Anime_MediaEntryEpisode, Offline_AnimeEntry, Offline_AssetMapImageMap } from "@/api/generated/types"
import { usePlayVideo } from "@/api/hooks/mediaplayer.hooks"
import { OfflineMetaSection } from "@/app/(main)/(offline)/offline/(entry)/_components/offline-meta-section"
import { useOfflineSnapshot } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot-context"
import { offline_getAssetUrl } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.utils"
import { EpisodeCard } from "@/app/(main)/_features/anime/_components/episode-card"
import { EpisodeGridItem } from "@/app/(main)/_features/anime/_components/episode-grid-item"
import { EpisodeItemIsolation } from "@/app/(main)/entry/_containers/episode-list/episode-item"
import { usePlayNextVideoOnMount } from "@/app/(main)/entry/_lib/media-player"
import { LuffyError } from "@/components/shared/luffy-error"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import { IconButton } from "@/components/ui/button"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselItem } from "@/components/ui/carousel"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import Image from "next/image"
import { useRouter, useSearchParams } from "next/navigation"
import React, { memo } from "react"
import { AiFillWarning } from "react-icons/ai"
import { MdInfo } from "react-icons/md"

export const dynamic = "force-static"

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
    entry: Offline_AnimeEntry
    assetMap: Offline_AssetMapImageMap | undefined
}

function EpisodeLists(props: EpisodeListsProps) {

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

    const { mutate: playVideo } = usePlayVideo()

    usePlayNextVideoOnMount({
        onPlay: () => {
            if (episodesToWatch.length > 0) {
                playVideo({ path: episodesToWatch[0].localFile?.path ?? "" })
            }
        },
    })

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
                                    <EpisodeCard
                                        key={episode.localFile?.path || ""}
                                        image={episode.episodeMetadata?.image || episode.basicMedia?.bannerImage || episode.basicMedia?.coverImage?.extraLarge}
                                        topTitle={episode.episodeTitle || episode?.basicMedia?.title?.userPreferred}
                                        title={episode.displayTitle}
                                        meta={episode.episodeMetadata?.airDate ?? undefined}
                                        isInvalid={episode.isInvalid}
                                        progressTotal={episode.basicMedia?.episodes}
                                        progressNumber={episode.progressNumber}
                                        episodeNumber={episode.episodeNumber}
                                        onClick={() => playVideo({ path: episode.localFile?.path ?? "" })}
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
    episode: Anime_MediaEntryEpisode,
    media: AL_BaseMedia,
    onPlay: ({ path }: { path: string }) => void,
    isWatched?: boolean
}) => {

    return (
        <EpisodeItemIsolation.Provider>
            <EpisodeGridItem
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
