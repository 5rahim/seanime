import { Anime_Entry, Anime_Episode } from "@/api/generated/types"
import { useGetTorrentstreamEpisodeCollection } from "@/api/hooks/torrentstream.hooks"

import { useSeaCommandInject } from "@/app/(main)/_features/sea-command/use-inject"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import {
    __torrentSearch_drawerEpisodeAtom,
    __torrentSearch_drawerIsOpenAtom,
} from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { TorrentStreamEpisodeSection } from "@/app/(main)/entry/_containers/torrent-stream/_components/torrent-stream-episode-section"
import { useHandleStartTorrentStream, useTorrentStreamAutoplay } from "@/app/(main)/entry/_containers/torrent-stream/_lib/handle-torrent-stream"
import { useTorrentStreamingSelectedEpisode } from "@/app/(main)/entry/_lib/torrent-streaming.atoms"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Switch } from "@/components/ui/switch"
import { logger } from "@/lib/helpers/debug"
import { useAtom, useSetAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React from "react"

type TorrentStreamPageProps = {
    children?: React.ReactNode
    entry: Anime_Entry
    bottomSection?: React.ReactNode
}

const autoSelectFileAtom = atomWithStorage("sea-torrentstream-auto-select-file", true)

export function TorrentStreamPage(props: TorrentStreamPageProps) {

    const {
        children,
        entry,
        bottomSection,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    const [autoSelect, setAutoSelect] = React.useState(serverStatus?.torrentstreamSettings?.autoSelect)

    const [autoSelectFile, setAutoSelectFile] = useAtom(autoSelectFileAtom)

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

    const setTorrentDrawerIsOpen = useSetAtom(__torrentSearch_drawerIsOpenAtom)
    const setTorrentSearchEpisode = useSetAtom(__torrentSearch_drawerEpisodeAtom)

    // Stores the episode that was clicked
    const { setTorrentStreamingSelectedEpisode } = useTorrentStreamingSelectedEpisode()


    /**
     * Handle auto-select
     */
    const { handleAutoSelectTorrentStream, isPending } = useHandleStartTorrentStream()
    const { setTorrentstreamAutoplayInfo } = useTorrentStreamAutoplay()

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
        handleAutoSelectTorrentStream({
            entry: entry,
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
    const handleEpisodeClick = (episode: Anime_Episode) => {
            if (isPending) return

            setTorrentStreamingSelectedEpisode(episode)

            React.startTransition(() => {
                // If auto-select is enabled, send the streaming request
                if (autoSelect) {
                    handleAutoSelect(entry, episode)
                } else {

                    setTorrentSearchEpisode(episode.episodeNumber)
                    React.startTransition(() => {
                        // If auto-select file is enabled, open the torrent drawer
                        if (autoSelectFile) {
                            setTorrentDrawerIsOpen("select")

                            // Set the torrent stream autoplay info
                            handleSetTorrentstreamAutoplayInfo(episode)

                        } else { // Otherwise, open the torrent drawer
                        setTorrentDrawerIsOpen("select-file")
                        }
                    })

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
            <AppLayoutStack data-torrent-stream-page>
                <div className="absolute right-0 top-[-3rem]" data-torrent-stream-page-title-container>
                    <h2 className="text-xl lg:text-3xl flex items-center gap-3">Torrent streaming</h2>
                </div>

                <div className="flex flex-col md:flex-row gap-6 pb-6 2xl:py-0" data-torrent-stream-page-content-actions-container>
                    <Switch
                        label="Auto-select"
                        value={autoSelect}
                        onValueChange={v => {
                            setAutoSelect(v)
                        }}
                        // moreHelp="Automatically select the best torrent and file to stream"
                        fieldClass="w-fit"
                    />

                    {!autoSelect && (
                        <Switch
                            label="Auto-select file"
                            value={autoSelectFile}
                            onValueChange={v => {
                                setAutoSelectFile(v)
                            }}
                            moreHelp="The episode file will be automatically selected from your chosen batch torrent"
                            fieldClass="w-fit"
                        />
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
                    onPlayNextEpisodeOnMount={handlePlayNextEpisodeOnMount}
                    bottomSection={bottomSection}
                />
            </AppLayoutStack>
        </>
    )
}
