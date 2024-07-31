import { Anime_AnimeEntry, Anime_AnimeEntryEpisode } from "@/api/generated/types"
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
import { Switch } from "@/components/ui/switch"
import { useThemeSettings } from "@/lib/theme/hooks"
import { useSetAtom } from "jotai/react"
import React, { useMemo } from "react"

type TorrentStreamPageProps = {
    children?: React.ReactNode
    entry: Anime_AnimeEntry
}

export function TorrentStreamPage(props: TorrentStreamPageProps) {

    const {
        children,
        entry,
        ...rest
    } = props

    const serverStatus = useServerStatus()
    const ts = useThemeSettings()

    const [autoSelect, setAutoSelect] = React.useState(serverStatus?.torrentstreamSettings?.autoSelect)

    const [manuallySelectFile, setManuallySelectFile] = React.useState(true)

    /**
     * Get all episodes to watch
     */
    const { data: episodeCollection, isLoading } = useGetTorrentstreamEpisodeCollection(entry.mediaId)

    React.useLayoutEffect(() => {
        // Set auto-select to the server status value
        if (!episodeCollection?.hasMappingError) {
            setAutoSelect(serverStatus?.torrentstreamSettings?.autoSelect)
        } else {
            // Fall back to manual select if no download info (no AniZip data)
            setAutoSelect(false)
        }
    }, [serverStatus?.torrentstreamSettings?.autoSelect, episodeCollection])

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
    const handleEpisodeClick = (episode: Anime_AnimeEntryEpisode) => {
        if (isPending) return

        setTorrentStreamingSelectedEpisode(episode)

        React.startTransition(() => {
            if (autoSelect) {
                if (episode.aniDBEpisode) {
                    handleAutoSelectTorrentStream({
                        entry,
                        episodeNumber: episode.episodeNumber,
                        aniDBEpisode: episode.aniDBEpisode,
                    })
                }
            } else if (!manuallySelectFile) {
                setTorrentSearchEpisode(episode.episodeNumber)
                React.startTransition(() => {
                    setTorrentDrawerIsOpen("select")
                })
            } else {
                setTorrentSearchEpisode(episode.episodeNumber)
                React.startTransition(() => {
                    setTorrentDrawerIsOpen("select-file")
                })
            }
        })
        // toast.info("Starting torrent stream...")
    }

    if (!entry.media) return null
    if (isLoading) return <LoadingSpinner />

    return (
        <AppLayoutStack>
            <div className="absolute right-0 top-[-3rem]">
                <h2 className="text-xl lg:text-3xl">Torrent streaming</h2>
            </div>

            <div className="flex flex-col md:flex-row gap-4">
                <Switch
                    label="Auto-select"
                    value={autoSelect}
                    onValueChange={v => {
                        setAutoSelect(v)
                    }}
                    help="Automatically select the best torrent and file to stream"
                    fieldClass="w-fit"
                />

                {!autoSelect && (
                    <Switch
                        label="Manually select file"
                        value={manuallySelectFile}
                        onValueChange={v => {
                            setManuallySelectFile(v)
                        }}
                        help="Manually select the file to stream after selecting a torrent"
                        fieldClass="w-fit"
                    />
                )}
            </div>

            {episodeCollection?.hasMappingError && (
                <div className="">
                    <p className="text-red-200 opacity-50">
                        No metadata info available for this anime. You may need to manually select the file to stream.
                    </p>
                </div>

            )}

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
                                meta={!!episode.episodeMetadata?.length
                                    ? `${episode.episodeMetadata?.length}m`
                                    : episode.episodeMetadata?.airDate ?? undefined}
                                isInvalid={episode.isInvalid}
                                progressTotal={episode.baseAnime?.episodes}
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
                        media={episode?.baseAnime as any}
                        title={episode?.displayTitle || episode?.baseAnime?.title?.userPreferred || ""}
                        image={episode?.episodeMetadata?.image || episode?.baseAnime?.coverImage?.large}
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
