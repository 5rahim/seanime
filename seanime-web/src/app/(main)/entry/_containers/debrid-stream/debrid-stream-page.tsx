import { Anime_Entry, Anime_Episode } from "@/api/generated/types"
import { useGetTorrentstreamEpisodeCollection } from "@/api/hooks/torrentstream.hooks"
import {
    __torrentSearch_drawerEpisodeAtom,
    __torrentSearch_drawerIsOpenAtom,
} from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { TorrentStreamEpisodeSection } from "@/app/(main)/entry/_containers/torrent-stream/_components/torrent-stream-episode-section"
import { useTorrentStreamingSelectedEpisode } from "@/app/(main)/entry/_lib/torrent-streaming.atoms"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
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

    /**
     * Get all episodes to watch
     */
    const { data: episodeCollection, isLoading } = useGetTorrentstreamEpisodeCollection(entry.mediaId)


    const setTorrentSearchDrawerOpen = useSetAtom(__torrentSearch_drawerIsOpenAtom)
    const setTorrentSearchEpisode = useSetAtom(__torrentSearch_drawerEpisodeAtom)

    // Stores the episode that was clicked
    const { setTorrentStreamingSelectedEpisode } = useTorrentStreamingSelectedEpisode()

    function handlePlayNextEpisodeOnMount(episode: Anime_Episode) {
        handleEpisodeClick(episode)
    }

    const handleEpisodeClick = (episode: Anime_Episode) => {

        setTorrentStreamingSelectedEpisode(episode)

        React.startTransition(() => {
            setTorrentSearchEpisode(episode.episodeNumber)
            React.startTransition(() => {
                setTorrentSearchDrawerOpen("debrid-stream")
            })
        })
    }

    if (!entry.media) return null
    if (isLoading) return <LoadingSpinner />

    return (
        <>
            <AppLayoutStack>
                <div className="absolute right-0 top-[-3rem]">
                    <h2 className="text-xl lg:text-3xl flex items-center gap-3">Debrid streaming</h2>
                </div>

                <div className="flex flex-col md:flex-row gap-4 h-10">
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
