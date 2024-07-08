import { Anime_MediaEntry, Anime_MediaEntryEpisode } from "@/api/generated/types"
import { useGetTorrentstreamEpisodeCollection } from "@/api/hooks/torrentstream.hooks"
import { EpisodeCard } from "@/app/(main)/_features/anime/_components/episode-card"
import { EpisodeGridItem } from "@/app/(main)/_features/anime/_components/episode-grid-item"
import { MediaEpisodeInfoModal } from "@/app/(main)/_features/media/_components/media-episode-info-modal"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { EpisodeListGrid } from "@/app/(main)/entry/_components/episode-list-grid"
import {
    __torrentSearch_drawerEpisodeAtom,
    __torrentSearch_drawerIsOpenAtom,
} from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { useHandleStartTorrentStream } from "@/app/(main)/entry/_containers/torrent-stream/_lib/handle-torrent-stream"
import { useTorrentStreamingSelectedEpisode } from "@/app/(main)/entry/_lib/torrent-streaming.atoms"
import { episodeCardCarouselItemClass } from "@/components/shared/classnames"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselItem } from "@/components/ui/carousel"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { useThemeSettings } from "@/lib/theme/hooks"
import { useSetAtom } from "jotai/react"
import React, { useMemo } from "react"

type TorrentStreamPageProps = {
    children?: React.ReactNode
    entry: Anime_MediaEntry
}

export function TorrentStreamPage(props: TorrentStreamPageProps) {

    const {
        children,
        entry,
        ...rest
    } = props

    const serverStatus = useServerStatus()
    const ts = useThemeSettings()

    /**
     * Get all episodes to watch
     */
    const { data: episodeCollection, isLoading } = useGetTorrentstreamEpisodeCollection(entry.mediaId)

    /**
     * Organize episodes to watch
     */
    const episodesToWatch = useMemo(() => {
        if (!episodeCollection?.episodes) return []
        let ret = [...episodeCollection?.episodes]
        ret = ((!!entry.listData?.progress && !!entry.media?.episodes && entry.listData?.progress === entry.media?.episodes)
                ? ret?.reverse()
                : ret?.slice(entry.listData?.progress || 0)
        )?.slice(0, 30)
        return ret
    }, [episodeCollection?.episodes, entry.nextEpisode, entry.listData?.progress])


    const setTorrentDrawerIsOpen = useSetAtom(__torrentSearch_drawerIsOpenAtom)
    const setTorrentSearchEpisode = useSetAtom(__torrentSearch_drawerEpisodeAtom)

    /**
     * Handle start torrent stream
     */
    const { handleAutoSelectTorrentStream, isPending } = useHandleStartTorrentStream()

    // Stores the episode that was clicked
    const { setTorrentStreamingSelectedEpisode } = useTorrentStreamingSelectedEpisode()

    /**
     * Handle episode click
     * - If auto-select is enabled, send the streaming request
     * - If auto-select is disabled, open the torrent drawer
     */
    const handleEpisodeClick = (episode: Anime_MediaEntryEpisode) => {
        if (isPending) return

        setTorrentStreamingSelectedEpisode(episode)

        React.startTransition(() => {
            if (serverStatus?.torrentstreamSettings?.autoSelect) {
                if (episode.aniDBEpisode) {
                    handleAutoSelectTorrentStream({
                        entry,
                        episodeNumber: episode.episodeNumber,
                        aniDBEpisode: episode.aniDBEpisode,
                    })
                }
            } else {
                setTorrentSearchEpisode(episode.episodeNumber)
                React.startTransition(() => {
                    setTorrentDrawerIsOpen("select")
                })
            }
        })
        // toast.info("Starting torrent stream...")
    }

    if (!entry.media) return null
    if (isLoading) return <LoadingSpinner />

    return (
        <AppLayoutStack>
            <h2>Torrent streaming</h2>
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
                                image={episode.episodeMetadata?.image || episode.basicMedia?.bannerImage || episode.basicMedia?.coverImage?.extraLarge}
                                topTitle={episode.episodeTitle || episode?.basicMedia?.title?.userPreferred}
                                title={episode.displayTitle}
                                meta={episode.episodeMetadata?.airDate ?? undefined}
                                isInvalid={episode.isInvalid}
                                progressTotal={episode.basicMedia?.episodes}
                                progressNumber={episode.progressNumber}
                                episodeNumber={episode.episodeNumber}
                                hasDiscrepancy={episodeCollection?.episodes?.findIndex(e => e.type === "special") !== -1}
                                onClick={() => {
                                    handleEpisodeClick(episode)
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
                        media={episode?.basicMedia as any}
                        title={episode?.displayTitle || episode?.basicMedia?.title?.userPreferred || ""}
                        image={episode?.episodeMetadata?.image || episode?.basicMedia?.coverImage?.large}
                        episodeTitle={episode?.episodeTitle}
                        onClick={() => {
                            handleEpisodeClick(episode)
                        }}
                        description={episode?.episodeMetadata?.overview}
                        isFiller={episode?.episodeMetadata?.isFiller}
                        length={episode?.episodeMetadata?.length}
                        isWatched={!!entry.listData?.progress && entry.listData.progress >= episode?.progressNumber}
                        className="flex-none w-full"
                        action={<>
                            <MediaEpisodeInfoModal
                                title={episode.displayTitle}
                                image={episode.episodeMetadata?.image}
                                episodeTitle={episode.episodeTitle}
                                airDate={episode.episodeMetadata?.airDate}
                                length={episode.episodeMetadata?.length}
                                summary={episode.episodeMetadata?.summary}
                                isInvalid={episode.isInvalid}
                            />
                        </>}
                    />
                ))}
            </EpisodeListGrid>
        </AppLayoutStack>
    )
}
