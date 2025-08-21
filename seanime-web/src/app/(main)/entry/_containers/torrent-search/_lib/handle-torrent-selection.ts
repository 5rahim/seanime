import { Anime_Entry, Anime_Episode } from "@/api/generated/types"
import { useHandleStartDebridStream } from "@/app/(main)/entry/_containers/debrid-stream/_lib/handle-debrid-stream"
import { __torrentSearch_selectedTorrentsAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-container"
import { __torrentSearch_selectionAtom, TorrentSelectionType } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import {
    useDebridStreamAutoplay,
    useHandleStartTorrentStream,
    useTorrentStreamAutoplay,
} from "@/app/(main)/entry/_containers/torrent-stream/_lib/handle-torrent-stream"
import { __torrentSearch_torrentstreamSelectedTorrentAtom } from "@/app/(main)/entry/_containers/torrent-stream/torrent-stream-file-selection-modal"
import { atom, useSetAtom } from "jotai/index"
import { useAtom } from "jotai/react"
import React from "react"

const __torrentSearch_streamingSelectedEpisodeAtom = atom<Anime_Episode | null>(null)

export function useTorrentSearchSelectedStreamEpisode() {
    const [value, setter] = useAtom(__torrentSearch_streamingSelectedEpisodeAtom)

    return {
        torrentStreamingSelectedEpisode: value,
        setTorrentStreamingSelectedEpisode: setter,
    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export function useTorrentSearchSelection({ type = "download", entry }: { type: TorrentSelectionType | undefined, entry: Anime_Entry }) {

    const [selectedTorrents, setSelectedTorrents] = useAtom(__torrentSearch_selectedTorrentsAtom)
    const { handleManualTorrentStreamSelection } = useHandleStartTorrentStream()
    const { handleStreamSelection } = useHandleStartDebridStream()
    const { torrentStreamingSelectedEpisode } = useTorrentSearchSelectedStreamEpisode()
    const setTorrentstreamSelectedTorrent = useSetAtom(__torrentSearch_torrentstreamSelectedTorrentAtom)
    const [, setDrawerOpen] = useAtom(__torrentSearch_selectionAtom)
    const { setDebridstreamAutoplaySelectedTorrent } = useDebridStreamAutoplay()
    const { setTorrentstreamAutoplaySelectedTorrent } = useTorrentStreamAutoplay()

    const onTorrentValidated = () => {
        console.log("onTorrentValidated", torrentStreamingSelectedEpisode)
        if (type === "torrentstream-select") {
            if (selectedTorrents.length && !!torrentStreamingSelectedEpisode?.aniDBEpisode) {
                setTorrentstreamAutoplaySelectedTorrent(selectedTorrents[0])
                handleManualTorrentStreamSelection({
                    torrent: selectedTorrents[0],
                    entry,
                    aniDBEpisode: torrentStreamingSelectedEpisode.aniDBEpisode,
                    episodeNumber: torrentStreamingSelectedEpisode.episodeNumber,
                    chosenFileIndex: undefined,
                })
                setDrawerOpen(undefined)
                React.startTransition(() => {
                    setSelectedTorrents([])
                })
            }
        } else if (type === "torrentstream-select-file") {
            // Open the drawer to select the file
            if (selectedTorrents.length && !!torrentStreamingSelectedEpisode?.aniDBEpisode) {
                // This opens the file selection drawer
                setTorrentstreamSelectedTorrent(selectedTorrents[0])
                React.startTransition(() => {
                    setSelectedTorrents([])
                })
            }
        } else if (type === "debridstream-select") {
            // Start debrid stream with auto file selection
            if (selectedTorrents.length && !!torrentStreamingSelectedEpisode?.aniDBEpisode) {
                setDebridstreamAutoplaySelectedTorrent(selectedTorrents[0])
                handleStreamSelection({
                    torrent: selectedTorrents[0],
                    entry,
                    aniDBEpisode: torrentStreamingSelectedEpisode.aniDBEpisode,
                    episodeNumber: torrentStreamingSelectedEpisode.episodeNumber,
                    chosenFileId: "",
                })
                setDrawerOpen(undefined)
                React.startTransition(() => {
                    setSelectedTorrents([])
                })
            }
        } else if (type === "debridstream-select-file") {
            // Open the drawer to select the file
            if (selectedTorrents.length && !!torrentStreamingSelectedEpisode?.aniDBEpisode) {
                // This opens the file selection drawer
                setTorrentstreamSelectedTorrent(selectedTorrents[0])
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
