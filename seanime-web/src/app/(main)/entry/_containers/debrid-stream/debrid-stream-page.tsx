import { Anime_Entry, Anime_Episode } from "@/api/generated/types"
import { useGetAnimeEpisodeCollection } from "@/api/hooks/anime.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useHandleStartDebridStream } from "@/app/(main)/entry/_containers/debrid-stream/_lib/handle-debrid-stream"
import { useTorrentSearchSelectedStreamEpisode } from "@/app/(main)/entry/_containers/torrent-search/_lib/handle-torrent-selection"
import {
    __torrentSearch_selectionAtom,
    __torrentSearch_selectionEpisodeAtom,
} from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { TorrentStreamEpisodeSection } from "@/app/(main)/entry/_containers/torrent-stream/_components/torrent-stream-episode-section"
import { useDebridStreamAutoplay } from "@/app/(main)/entry/_containers/torrent-stream/_lib/handle-torrent-stream"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Switch } from "@/components/ui/switch"
import { logger } from "@/lib/helpers/debug"
import { useAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React from "react"

type DebridStreamPageProps = {
    children?: React.ReactNode
    entry: Anime_Entry
    bottomSection?: React.ReactNode
}

const autoSelectFileAtom = atomWithStorage("sea-debridstream-manually-select-file", false)

// DEVNOTE: This page uses some utility functions from the TorrentStream feature

export function DebridStreamPage(props: DebridStreamPageProps) {

    const {
        children,
        entry,
        bottomSection,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    // State to manage auto-select setting
    const [autoSelect, setAutoSelect] = React.useState(serverStatus?.debridSettings?.streamAutoSelect)
    const [autoSelectFile, setAutoSelectFile] = useAtom(autoSelectFileAtom)

    /**
     * Get all episodes to watch
     */
    const { data: episodeCollection, isLoading } = useGetAnimeEpisodeCollection(entry.mediaId)

    React.useLayoutEffect(() => {
        // Set auto-select to the server status value
        if (!episodeCollection?.hasMappingError) {
            setAutoSelect(serverStatus?.debridSettings?.streamAutoSelect)
        } else {
            // Fall back to manual select if no download info (no Animap data)
            setAutoSelect(false)
        }
    }, [serverStatus?.torrentstreamSettings?.autoSelect, episodeCollection])

    // Atoms to control the torrent search drawer state
    const [, setTorrentSearchDrawerOpen] = useAtom(__torrentSearch_selectionAtom)
    const [, setTorrentSearchEpisode] = useAtom(__torrentSearch_selectionEpisodeAtom)

    // Stores the episode that was clicked
    const { setTorrentStreamingSelectedEpisode } = useTorrentSearchSelectedStreamEpisode()

    // Function to handle playing the next episode on mount
    function handlePlayNextEpisodeOnMount(episode: Anime_Episode) {
        if (autoSelect) {
            handleAutoSelect(entry, episode)
        } else {
            handleEpisodeClick(episode)
        }
    }

    // Hook to handle starting the debrid stream
    const { handleAutoSelectStream } = useHandleStartDebridStream()

    // Hook to manage debrid stream autoplay information
    const { setDebridstreamAutoplayInfo } = useDebridStreamAutoplay()

    // Function to set the debrid stream autoplay info
    // It checks if there is a next episode and if it has aniDBEpisode
    // If so, it sets the autoplay info
    // Otherwise, it resets the autoplay info
    function handleSetDebridstreamAutoplayInfo(episode: Anime_Episode | undefined) {
        if (!episode || !episode.aniDBEpisode || !episodeCollection?.episodes) return
        const nextEpisode = episodeCollection?.episodes?.find(e => e.episodeNumber === episode.episodeNumber + 1)
        logger("TORRENTSTREAM").info("Auto select, Next episode", nextEpisode)
        if (nextEpisode && !!nextEpisode.aniDBEpisode) {
            setDebridstreamAutoplayInfo({
                allEpisodes: episodeCollection?.episodes,
                entry: entry,
                episodeNumber: nextEpisode.episodeNumber,
                aniDBEpisode: nextEpisode.aniDBEpisode,
                type: "debridstream",
            })
        } else {
            setDebridstreamAutoplayInfo(null)
        }
    }

    // Function to handle auto-selecting an episode
    function handleAutoSelect(entry: Anime_Entry, episode: Anime_Episode | undefined) {
        if (!episode || !episode.aniDBEpisode || !episodeCollection?.episodes) return
        // Start the debrid stream
        handleAutoSelectStream({
            entry: entry,
            episodeNumber: episode.episodeNumber,
            aniDBEpisode: episode.aniDBEpisode,
        })

        // Set the debrid stream autoplay info
        handleSetDebridstreamAutoplayInfo(episode)
    }

    // Function to handle episode click events
    const handleEpisodeClick = (episode: Anime_Episode) => {
        if (!episode || !episode.aniDBEpisode) return

        setTorrentStreamingSelectedEpisode(episode)

        if (autoSelect) {
            handleAutoSelect(entry, episode)
        } else {
            React.startTransition(() => {
                setTorrentSearchEpisode(episode.episodeNumber)
                React.startTransition(() => {
                    // If auto-select file is enabled, open the debrid stream select drawer
                    if (autoSelectFile) {
                        setTorrentSearchDrawerOpen("debridstream-select")

                        // Set the debrid stream autoplay info
                        handleSetDebridstreamAutoplayInfo(episode)
                    } else {
                        // Otherwise, open the debrid stream select file drawer
                        setTorrentSearchDrawerOpen("debridstream-select-file")
                    }
                })
            })
        }
    }

    if (!entry.media) return null
    if (isLoading) return <LoadingSpinner />

    return (
        <>
            <PageWrapper
                data-anime-entry-page-debrid-stream-view
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
                <AppLayoutStack data-debrid-stream-page>
                    <div className="absolute right-0 top-[-3rem]" data-debrid-stream-page-title-container>
                        <h2 className="text-xl lg:text-3xl flex items-center gap-3">Debrid streaming</h2>
                    </div>

                    <div className="flex flex-col md:flex-row gap-6 pb-6 2xl:py-0" data-debrid-stream-page-content-actions-container>
                        <Switch
                            label="Auto-select"
                            value={autoSelect}
                            onValueChange={v => {
                                setAutoSelect(v)
                            }}
                            // help="Automatically select the best torrent and file to stream"
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
                        <div data-debrid-stream-page-no-metadata-message-container>
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
            </PageWrapper>
        </>
    )
}
