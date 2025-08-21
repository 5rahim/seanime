import { Anime_Entry, Anime_Episode } from "@/api/generated/types"
import { useAutoPlaySelectedTorrent } from "@/app/(main)/_features/autoplay/autoplay"
import { useHandleStartDebridStream } from "@/app/(main)/entry/_containers/debrid-stream/_lib/handle-debrid-stream"
import { __torrentSearch_selectedTorrentsAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-container"
import { __torrentSearch_selectionAtom, TorrentSelectionType } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { useHandleStartTorrentStream } from "@/app/(main)/entry/_containers/torrent-stream/_lib/handle-torrent-stream"
import { __torrentSearch_fileSelectionTorrentAtom } from "@/app/(main)/entry/_containers/torrent-stream/torrent-stream-file-selection-modal"
import { atom, useSetAtom } from "jotai/index"
import { useAtom } from "jotai/react"
import React from "react"

const __torrentSearch_streamingSelectedEpisodeAtom = atom<Anime_Episode | null>(null)

// Stores the currently selected episode for torrent stream
export function useTorrentSearchSelectedStreamEpisode() {
    const [value, setter] = useAtom(__torrentSearch_streamingSelectedEpisodeAtom)

    return {
        torrentStreamingSelectedEpisode: value,
        setTorrentStreamingSelectedEpisode: setter,
    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export function useTorrentSearchSelection({ type = "download", entry }: { type: TorrentSelectionType | undefined, entry: Anime_Entry }) {
    const { handleStreamSelection: handleTorrentstreamSelection } = useHandleStartTorrentStream()
    const { handleStreamSelection: handleDebridstreamSelection } = useHandleStartDebridStream()

    // Currently selected torrents
    const [selectedTorrents, setSelectedTorrents] = useAtom(__torrentSearch_selectedTorrentsAtom)
    // Get the currently selected episode
    const { torrentStreamingSelectedEpisode } = useTorrentSearchSelectedStreamEpisode()
    // Sets the selected torrent for file selection
    const setTorrentstreamSelectedTorrent = useSetAtom(__torrentSearch_fileSelectionTorrentAtom)
    // Sets the selected torrent for auto play
    const { setAutoPlayTorrent } = useAutoPlaySelectedTorrent()
    const setTorrentSearch = useSetAtom(__torrentSearch_selectionAtom)

    const onTorrentValidated = () => {
        const torrent = selectedTorrents[0]
        console.log("onTorrentValidated", torrentStreamingSelectedEpisode)
        // User manually selected a torrent
        if (type === "torrentstream-select") {
            if (!!torrent && !!torrentStreamingSelectedEpisode?.aniDBEpisode) {
                // Store the selected torrent
                setAutoPlayTorrent(torrent)
                // Start torrent stream with auto file selection
                handleTorrentstreamSelection({
                    torrent: torrent,
                    entry,
                    aniDBEpisode: torrentStreamingSelectedEpisode.aniDBEpisode,
                    episodeNumber: torrentStreamingSelectedEpisode.episodeNumber,
                    chosenFileIndex: undefined,
                })
                // Close torrent search
                setTorrentSearch(undefined)
                React.startTransition(() => {
                    setSelectedTorrents([])
                })
            }
        } else if (type === "torrentstream-select-file") {
            // Open the drawer to select the file
            if (!!torrent && !!torrentStreamingSelectedEpisode?.aniDBEpisode) {
                // This opens the file selection drawer
                setTorrentstreamSelectedTorrent(torrent)
                React.startTransition(() => {
                    setSelectedTorrents([])
                })
            }
        } else if (type === "debridstream-select") {
            // Start debrid stream with auto file selection
            if (selectedTorrents.length && !!torrentStreamingSelectedEpisode?.aniDBEpisode) {
                // Store the selected torrent
                setAutoPlayTorrent(torrent)
                // Start debrid stream with auto file selection
                handleDebridstreamSelection({
                    torrent: torrent,
                    entry,
                    aniDBEpisode: torrentStreamingSelectedEpisode.aniDBEpisode,
                    episodeNumber: torrentStreamingSelectedEpisode.episodeNumber,
                    chosenFileId: "",
                })
                // Close torrent search
                setTorrentSearch(undefined)
                React.startTransition(() => {
                    setSelectedTorrents([])
                })
            }
        } else if (type === "debridstream-select-file") {
            // Open the drawer to select the file
            if (selectedTorrents.length && !!torrentStreamingSelectedEpisode?.aniDBEpisode) {
                // This opens the file selection drawer
                setTorrentstreamSelectedTorrent(torrent)
                React.startTransition(() => {
                    setSelectedTorrents([])
                })
            }
        }
    }

    return {
        onTorrentValidated,
    }
}
