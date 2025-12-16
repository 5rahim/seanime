import { Anime_Entry, Anime_Episode } from "@/api/generated/types"
import { useGetAnimeEpisodeCollection } from "@/api/hooks/anime.hooks"
import { useGetTorrentstreamBatchHistory } from "@/api/hooks/torrentstream.hooks"
import { useAutoPlaySelectedTorrent, useTorrentstreamAutoplay } from "@/app/(main)/_features/autoplay/autoplay"

import { useSeaCommandInject } from "@/app/(main)/_features/sea-command/use-inject"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useTorrentSearchSelectedStreamEpisode } from "@/app/(main)/entry/_containers/torrent-search/_lib/handle-torrent-selection"
import {
    __torrentSearch_selectionAtom,
    __torrentSearch_selectionEpisodeAtom,
} from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { TorrentStreamEpisodeSection } from "@/app/(main)/entry/_containers/torrent-stream/_components/torrent-stream-episode-section"
import { useHandleStartTorrentStream } from "@/app/(main)/entry/_containers/torrent-stream/_lib/handle-torrent-stream"
import { ForcePlaybackMethod, useForcePlaybackMethod } from "@/app/(main)/entry/_lib/handle-play-media"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { IconButton } from "@/components/ui/button"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Popover } from "@/components/ui/popover"
import { Switch } from "@/components/ui/switch"
import { logger } from "@/lib/helpers/debug"
import { atom } from "jotai"
import { useAtom, useSetAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React from "react"
import { AiOutlineExclamationCircle } from "react-icons/ai"
import { BiX } from "react-icons/bi"

type TorrentStreamPageProps = {
    children?: React.ReactNode
    entry: Anime_Entry
    bottomSection?: React.ReactNode
}

export const __torrentStream_autoSelectFileAtom = atomWithStorage("sea-torrentstream-auto-select-file", true)
export const __torrentStream_currentSessionAutoSelectAtom = atom(false)

export function TorrentStreamPage(props: TorrentStreamPageProps) {

    const {
        children,
        entry,
        bottomSection,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    const [autoSelect, setAutoSelect] = useAtom(__torrentStream_currentSessionAutoSelectAtom)

    const [autoSelectFile, setAutoSelectFile] = useAtom(__torrentStream_autoSelectFileAtom)

    /**
     * Get all episodes to watch
     */
    const { data: episodeCollection, isLoading } = useGetAnimeEpisodeCollection(entry.mediaId)

    React.useLayoutEffect(() => {
        // Set auto-select to the server status value
        if (!episodeCollection?.hasMappingError) {
            setAutoSelect(serverStatus?.torrentstreamSettings?.autoSelect ?? false)
        } else {
            // Fall back to manual select if no download info (no Animap data)
            setAutoSelect(false)
        }
    }, [serverStatus?.torrentstreamSettings?.autoSelect, episodeCollection])

    const setTorrentSearchSelection = useSetAtom(__torrentSearch_selectionAtom)
    const setTorrentSearchEpisode = useSetAtom(__torrentSearch_selectionEpisodeAtom)

    // Stores the episode that was clicked
    const { setTorrentSearchStreamEpisode } = useTorrentSearchSelectedStreamEpisode()


    /**
     * Handle auto-select
     */
    const { handleAutoSelectStream, handleStreamSelection, isPending, isUsingNativePlayer } = useHandleStartTorrentStream()
    const { setTorrentstreamAutoplayInfo } = useTorrentstreamAutoplay()

    const { setAutoPlayTorrent } = useAutoPlaySelectedTorrent()

    const { forcePlaybackMethodFn } = useForcePlaybackMethod()

    const { data: batchHistory } = useGetTorrentstreamBatchHistory(entry?.mediaId, true)

    const [usePreviousBatch, setUsePreviousBatch] = React.useState(false)

    React.useEffect(() => {
        setUsePreviousBatch(!!batchHistory?.torrent?.isBatch)
    }, [batchHistory])

    // Function to set the torrent stream autoplay info
    // It checks if there is a next episode and if it has aniDBEpisode
    // If so, it sets the autoplay info
    // Otherwise, it resets the autoplay info
    function handleSetTorrentstreamAutoplayInfo(episode: Anime_Episode | undefined) {
        if (!episode || !episode.aniDBEpisode || !episodeCollection?.episodes) return
        const nextEpisode = episodeCollection?.episodes?.find(e => e.episodeNumber === episode.episodeNumber + 1)
        logger("TORRENTSTREAM").info("Auto select, Next episode", nextEpisode)
        if (nextEpisode && !!nextEpisode.aniDBEpisode) {
            setTorrentstreamAutoplayInfo({
                allEpisodes: episodeCollection?.episodes,
                entry: entry,
                episodeNumber: nextEpisode.episodeNumber,
                aniDBEpisode: nextEpisode.aniDBEpisode,
                type: "torrentstream",
            })
        } else {
            setTorrentstreamAutoplayInfo(null)
        }
    }

    function handleAutoSelect(entry: Anime_Entry, episode: Anime_Episode | undefined) {
        if (isPending || !episode || !episode.aniDBEpisode || !episodeCollection?.episodes) return
        // Start the torrent stream
        handleAutoSelectStream({
            mediaId: entry.mediaId,
            episodeNumber: episode.episodeNumber,
            aniDBEpisode: episode.aniDBEpisode,
        })

        // Set the torrent stream autoplay info
        handleSetTorrentstreamAutoplayInfo(episode)
    }

    function handlePlayNextEpisodeOnMount(episode: Anime_Episode) {
        if (autoSelect) {
            handleAutoSelect(entry, episode)
        } else {
            handleEpisodeClick(episode)
        }
    }

    /**
     * Handle episode click
     * - If auto-select is enabled, send the streaming request
     * - If auto-select is disabled, open the torrent drawer
     */
        // const setTorrentStreamLoader = useSetTorrentStreamLoader()
    const handleEpisodeClick = (episode: Anime_Episode, forcePlaybackMethod?: ForcePlaybackMethod) => {
            if (isPending) return

            console.log("handleEpisodeClick", episode, forcePlaybackMethod)

            setTorrentSearchStreamEpisode(episode)

            React.startTransition(() => {
                // If auto-select is enabled, send the streaming request
                if (autoSelect) {
                    forcePlaybackMethodFn(forcePlaybackMethod, () => {
                        handleAutoSelect(entry, episode)
                    })
                } else {

                    let started = false

                    // If we're using the previous batch
                    if (usePreviousBatch && batchHistory?.torrent && episode.aniDBEpisode) {

                        // Store the batch for auto play
                        setAutoPlayTorrent(batchHistory?.torrent, entry, batchHistory.batchEpisodeFiles)

                        if (autoSelectFile) {
                            forcePlaybackMethodFn(forcePlaybackMethod, () => {
                                handleStreamSelection({
                                    mediaId: entry.mediaId,
                                    episodeNumber: episode.episodeNumber,
                                    aniDBEpisode: episode.aniDBEpisode!,
                                    torrent: batchHistory.torrent!,
                                    chosenFileIndex: undefined,
                                    batchEpisodeFiles: undefined,
                                })
                            })
                            started = true
                        } else {
                            // Only auto select the file index if the user is trying to watch the next episode
                            if (batchHistory?.batchEpisodeFiles) {
                                let fileIndex: number | undefined = undefined

                                console.log("handleEpisodeClick (batchHistory)",
                                    batchHistory?.batchEpisodeFiles,
                                    episode.aniDBEpisode,
                                    episode.episodeNumber)

                                if (batchHistory?.batchEpisodeFiles.currentAniDBEpisode === episode.aniDBEpisode) {
                                    fileIndex = batchHistory.batchEpisodeFiles.current
                                } else {
                                    // guess index based on the last selected file
                                    const offset = episode.episodeNumber - batchHistory.batchEpisodeFiles.currentEpisodeNumber
                                    const file = batchHistory.batchEpisodeFiles.files?.find(f => f.index === (batchHistory.batchEpisodeFiles?.current || 0) + offset)
                                    if (file) {
                                        fileIndex = file.index
                                        console.log("handleEpisodeClick (batchHistory) found file", file)
                                    }
                                }

                                if (fileIndex !== undefined) {
                                    forcePlaybackMethodFn(forcePlaybackMethod, () => {
                                        handleStreamSelection({
                                            mediaId: entry.mediaId,
                                            episodeNumber: episode.episodeNumber,
                                            aniDBEpisode: episode.aniDBEpisode!,
                                            torrent: batchHistory.torrent!,
                                            chosenFileIndex: fileIndex,
                                            batchEpisodeFiles: (batchHistory.batchEpisodeFiles) ? {
                                                ...batchHistory.batchEpisodeFiles!,
                                                files: batchHistory.batchEpisodeFiles!.files!,
                                                current: fileIndex!,
                                                currentAniDBEpisode: episode.aniDBEpisode!,
                                                currentEpisodeNumber: episode.episodeNumber,
                                            } : undefined,
                                        })
                                    })
                                    started = true
                                }
                            }
                        }
                    }

                    if (!started) {
                        setTorrentSearchEpisode(episode.episodeNumber)
                        forcePlaybackMethodFn(forcePlaybackMethod, () => {
                            // If auto-select file is enabled, open the torrent drawer
                            if (autoSelectFile) {
                                setTorrentSearchSelection("torrentstream-select")
                            } else { // Otherwise, open the torrent drawer
                                setTorrentSearchSelection("torrentstream-select-file")
                            }
                        })
                    }
                    // Set the torrent stream autoplay info
                    handleSetTorrentstreamAutoplayInfo(episode)

                }
            })
            // toast.info("Starting torrent stream...")
        }

    const { inject, remove } = useSeaCommandInject()

    // Inject episodes into command palette when they're loaded
    React.useEffect(() => {
        if (!episodeCollection?.episodes?.length) return

        inject("torrent-stream-episodes", {
            items: episodeCollection.episodes.map(episode => ({
                id: `episode-${episode.episodeNumber}`,
                value: `${episode.episodeNumber}`,
                heading: "Episodes",
                render: () => (
                    <div className="flex gap-1 items-center w-full">
                        <p className="max-w-[70%] truncate">{episode.displayTitle}</p>
                        {!!episode.episodeTitle && (
                            <p className="text-[--muted] flex-1 truncate">- {episode.episodeTitle}</p>
                        )}
                    </div>
                ),
                onSelect: () => handleEpisodeClick(episode),
            })),
            // Optional custom filter
            filter: ({ item, input }) => {
                if (!input) return true
                return item.value.toLowerCase().includes(input.toLowerCase())
            },
        })

        return () => remove("torrent-stream-episodes")
    }, [episodeCollection?.episodes])

    if (!entry.media) return null
    if (isLoading) return <LoadingSpinner />

    return (
        <>


            <PageWrapper
                data-anime-entry-page-torrent-stream-view
                key="torrent-streaming-episodes"
                className="relative 2xl:order-first pb-10 lg:pt-0"
                {...{
                    initial: { opacity: 0, y: 60 },
                    animate: { opacity: 1, y: 0 },
                    exit: { opacity: 0, scale: 0.99 },
                    transition: {
                        duration: 0.35,
                    },
                }}
            >
                <div className="h-10 lg:h-0" />
                <AppLayoutStack data-torrent-stream-page>
                    {/*<div className="absolute right-0 top-[-3rem]" data-torrent-stream-page-title-container>*/}
                    {/*    <h2 className="text-xl lg:text-3xl flex items-center gap-3">Torrent streaming</h2>*/}
                    {/*</div>*/}

                    <div
                        className="flex flex-col flex-wrap lg:flex-nowrap items-start md:items-center md:flex-row gap-2 md:gap-6 2xl:py-0 lg:h-12"
                        data-torrent-stream-page-content-actions-container
                    >
                        <Switch
                            label="Auto-select"
                            value={autoSelect}
                            onValueChange={v => {
                                setAutoSelect(v)
                            }}
                            // moreHelp="Automatically select the best torrent and file to stream"
                            fieldClass="w-fit flex-none"
                        />

                        {!autoSelect && !usePreviousBatch && (
                            <Switch
                                label="Auto-select file"
                                value={autoSelectFile}
                                onValueChange={v => {
                                    setAutoSelectFile(v)
                                }}
                                moreHelp="The episode file will be automatically selected from your chosen batch torrent"
                                fieldClass="w-fit flex-none"
                                disabled={!autoSelect && usePreviousBatch}
                            />
                        )}

                        {(!autoSelect && usePreviousBatch && batchHistory) && (
                            <div className="relative w-full xl:max-w-[20rem] group/torrent-stream-batch-history">
                                <div className="rounded-full max-w-[20rem]">
                                    <div className="flex items-center gap-2">
                                        <div className="flex flex-none items-center justify-center">
                                            <IconButton
                                                intent="alert-glass"
                                                icon={<BiX />}
                                                size="xs"
                                                onClick={() => setUsePreviousBatch(false)}
                                                className="rounded-full"
                                            />
                                        </div>
                                        <div className="flex-1 flex items-center gap-2">
                                            <div className="flex items-center flex-none gap-1">Auto-selecting from previous torrent
                                                <Popover
                                                    className="text-sm"
                                                    trigger={
                                                        <AiOutlineExclamationCircle className="transition-opacity opacity-45 hover:opacity-90 cursor-pointer" />}
                                                >
                                                    {batchHistory.torrent?.name}
                                                </Popover>
                                            </div>
                                            <p className="line-clamp-1 text-[--muted] text-xs tracking-wide w-0 transition-all duration-300 ease-in-out group-hover/torrent-stream-batch-history:w-[20rem]">

                                            </p>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        )}
                    </div>

                    {episodeCollection?.hasMappingError && (
                        <div data-torrent-stream-page-no-metadata-message-container>
                            <p className="text-red-200 opacity-50">
                                No metadata info available for this anime. You may need to manually select the file to stream.
                            </p>
                        </div>

                    )}

                    <TorrentStreamEpisodeSection
                        episodeCollection={episodeCollection}
                        entry={entry}
                        onEpisodeClick={handleEpisodeClick}
                        onPlayExternallyEpisodeClick={!isUsingNativePlayer ? undefined : (episode) => {
                            handleEpisodeClick(episode, "playbackmanager")
                        }}
                        onPlayNextEpisodeOnMount={handlePlayNextEpisodeOnMount}
                        bottomSection={bottomSection}
                    />
                </AppLayoutStack>
            </PageWrapper>
        </>
    )
}
