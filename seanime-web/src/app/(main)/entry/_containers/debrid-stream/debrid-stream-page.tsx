import { Anime_Entry, Anime_Episode } from "@/api/generated/types"
import { useGetTorrentstreamEpisodeCollection } from "@/api/hooks/torrentstream.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useHandleStartDebridStream } from "@/app/(main)/entry/_containers/debrid-stream/_lib/handle-debrid-stream"
import {
    __torrentSearch_drawerEpisodeAtom,
    __torrentSearch_drawerIsOpenAtom,
} from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { TorrentStreamEpisodeSection } from "@/app/(main)/entry/_containers/torrent-stream/_components/torrent-stream-episode-section"
import { useDebridStreamAutoplay } from "@/app/(main)/entry/_containers/torrent-stream/_lib/handle-torrent-stream"
import { useTorrentStreamingSelectedEpisode } from "@/app/(main)/entry/_lib/torrent-streaming.atoms"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Switch } from "@/components/ui/switch"
import { logger } from "@/lib/helpers/debug"
import { useSetAtom } from "jotai/react"
import React from "react"

type DebridStreamPageProps = {
    children?: React.ReactNode
    entry: Anime_Entry
    bottomSection?: React.ReactNode
}

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

    /**
     * Get all episodes to watch
     */
    const { data: episodeCollection, isLoading } = useGetTorrentstreamEpisodeCollection(entry.mediaId)

    React.useLayoutEffect(() => {
        // Set auto-select to the server status value
        if (!episodeCollection?.hasMappingError) {
            setAutoSelect(serverStatus?.debridSettings?.streamAutoSelect)
        } else {
            // Fall back to manual select if no download info (no AniZip data)
            setAutoSelect(false)
        }
    }, [serverStatus?.torrentstreamSettings?.autoSelect, episodeCollection])

    // Atoms to control the torrent search drawer state
    const setTorrentSearchDrawerOpen = useSetAtom(__torrentSearch_drawerIsOpenAtom)
    const setTorrentSearchEpisode = useSetAtom(__torrentSearch_drawerEpisodeAtom)

    // Stores the episode that was clicked
    const { setTorrentStreamingSelectedEpisode } = useTorrentStreamingSelectedEpisode()

    // Function to handle playing the next episode on mount
    function handlePlayNextEpisodeOnMount(episode: Anime_Episode) {
        if (autoSelect) {
            handleAutoSelect(entry, episode)
        } else {
            setTorrentStreamingSelectedEpisode(episode)
        }
    }

    // Hook to handle starting the debrid stream
    const { handleAutoSelectStream } = useHandleStartDebridStream()

    // Hook to manage debrid stream autoplay information
    const { setDebridstreamAutoplayInfo } = useDebridStreamAutoplay()

    // Function to handle auto-selecting an episode
    function handleAutoSelect(entry: Anime_Entry, episode: Anime_Episode | undefined) {
        if (!episode || !episode.aniDBEpisode || !episodeCollection?.episodes) return
        // Start the debrid stream
        handleAutoSelectStream({
            entry: entry,
            episodeNumber: episode.episodeNumber,
            aniDBEpisode: episode.aniDBEpisode,
        })
        // Check if next episode exists for autoplay
        const nextEpisode = episodeCollection?.episodes?.find(e => e.episodeNumber === episode.episodeNumber + 1)
        logger("TORRENTSTREAM").info("Auto select, Next episode", nextEpisode)
        if (nextEpisode && !!nextEpisode.aniDBEpisode) {
            setDebridstreamAutoplayInfo({
                allEpisodes: episodeCollection?.episodes,
                entry: entry,
                episodeNumber: nextEpisode.episodeNumber,
                aniDBEpisode: nextEpisode.aniDBEpisode,
            })
        } else {
            setDebridstreamAutoplayInfo(null)
        }
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
                    setTorrentSearchDrawerOpen("debrid-stream")
                })
            })
        }
    }

    if (!entry.media) return null
    if (isLoading) return <LoadingSpinner />

    return (
        <>
            <AppLayoutStack>
                <div className="absolute right-0 top-[-3rem]">
                    <h2 className="text-xl lg:text-3xl flex items-center gap-3">Debrid streaming</h2>
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
                </div>

                {episodeCollection?.hasMappingError && (
                    <div className="">
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
