import { Anime_AnimeEntry, HibikeTorrent_AnimeTorrent, Torrentstream_PlaybackType } from "@/api/generated/types"
import { useTorrentstreamStartStream } from "@/api/hooks/torrentstream.hooks"
import { PlaybackTorrentStreaming, useCurrentDevicePlaybackSettings, useExternalPlayerLink } from "@/app/(main)/_atoms/playback.atoms"
import {
    __torrentstream__loadingStateAtom,
    __torrentstream__stateAtom,
    TorrentStreamState,
} from "@/app/(main)/entry/_containers/torrent-stream/torrent-stream-overlay"
import { clientIdAtom } from "@/app/websocket-provider"
import { useAtomValue } from "jotai"
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

    const { torrentStreamingPlayback } = useCurrentDevicePlaybackSettings()
    const { externalPlayerLink } = useExternalPlayerLink()
    const clientId = useAtomValue(clientIdAtom)

    const playbackType = React.useMemo<Torrentstream_PlaybackType>(() => {
        if (!externalPlayerLink?.length) {
            return "default"
        }
        if (torrentStreamingPlayback === PlaybackTorrentStreaming.ExternalPlayerLink) {
            return "externalPlayerLink"
        }
        return "default"
    }, [torrentStreamingPlayback, externalPlayerLink])

    const handleManualTorrentStreamSelection = React.useCallback((params: ManualTorrentStreamSelectionProps) => {
        mutate({
            mediaId: params.entry.mediaId,
            episodeNumber: params.episodeNumber,
            torrent: params.torrent,
            aniDBEpisode: params.aniDBEpisode,
            autoSelect: false,
            fileIndex: params.chosenFileIndex ?? undefined,
            playbackType: playbackType,
            clientId: clientId || "",
        }, {
            onSuccess: () => {
                // setLoadingState(null)
            },
            onError: () => {
                setLoadingState(null)
                setState(TorrentStreamState.Stopped)
            },
        })
    }, [playbackType, clientId])

    const handleAutoSelectTorrentStream = React.useCallback((params: AutoSelectTorrentStreamProps) => {
        mutate({
            mediaId: params.entry.mediaId,
            episodeNumber: params.episodeNumber,
            aniDBEpisode: params.aniDBEpisode,
            autoSelect: true,
            torrent: undefined,
            playbackType: playbackType,
            clientId: clientId || "",
        }, {
            onError: () => {
                setLoadingState(null)
                setState(TorrentStreamState.Stopped)
            },
        })
    }, [playbackType, clientId])

    return {
        handleManualTorrentStreamSelection,
        handleAutoSelectTorrentStream,
        isPending,
    }
}
