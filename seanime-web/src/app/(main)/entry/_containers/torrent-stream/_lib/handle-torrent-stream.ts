import { Anime_AnimeEntry, HibikeTorrent_AnimeTorrent } from "@/api/generated/types"
import { useTorrentstreamStartStream } from "@/api/hooks/torrentstream.hooks"
import {
    __torrentstream__loadingStateAtom,
    __torrentstream__stateAtom,
    TorrentStreamState,
} from "@/app/(main)/entry/_containers/torrent-stream/torrent-stream-overlay"
import { useSetAtom } from "jotai/react"
import React from "react"

type ManualTorrentStreamSelectionProps = {
    torrent: HibikeTorrent_AnimeTorrent
    entry: Anime_AnimeEntry
    episodeNumber: number
    aniDBEpisode: string
    chosenFileIndex: number | undefined | null
}
type AutoSelectTorrentStreamProps = {
    entry: Anime_AnimeEntry
    episodeNumber: number
    aniDBEpisode: string
}

export function useHandleStartTorrentStream() {

    const { mutate, isPending } = useTorrentstreamStartStream()

    const setLoadingState = useSetAtom(__torrentstream__loadingStateAtom)
    const setState = useSetAtom(__torrentstream__stateAtom)

    const handleManualTorrentStreamSelection = React.useCallback((params: ManualTorrentStreamSelectionProps) => {
        mutate({
            mediaId: params.entry.mediaId,
            episodeNumber: params.episodeNumber,
            torrent: params.torrent,
            aniDBEpisode: params.aniDBEpisode,
            autoSelect: false,
            fileIndex: params.chosenFileIndex ?? undefined,
        }, {
            onSuccess: () => {
                // setLoadingState(null)
            },
            onError: () => {
                setLoadingState(null)
                setState(TorrentStreamState.Stopped)
            },
        })
    }, [])

    const handleAutoSelectTorrentStream = React.useCallback((params: AutoSelectTorrentStreamProps) => {
        mutate({
            mediaId: params.entry.mediaId,
            episodeNumber: params.episodeNumber,
            aniDBEpisode: params.aniDBEpisode,
            autoSelect: true,
            torrent: undefined,
        }, {
            onError: () => {
                setLoadingState(null)
                setState(TorrentStreamState.Stopped)
            },
        })
    }, [])

    return {
        handleManualTorrentStreamSelection,
        handleAutoSelectTorrentStream,
        isPending,
    }
}
