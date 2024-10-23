import { Anime_Entry, HibikeTorrent_AnimeTorrent, Torrentstream_PlaybackType } from "@/api/generated/types"
import { useDebridStartStream } from "@/api/hooks/debrid.hooks"
import { PlaybackTorrentStreaming, useCurrentDevicePlaybackSettings, useExternalPlayerLink } from "@/app/(main)/_atoms/playback.atoms"
import { __debridstream_stateAtom } from "@/app/(main)/entry/_containers/debrid-stream/debrid-stream-overlay"
import { clientIdAtom } from "@/app/websocket-provider"
import { useAtomValue } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"

type DebridStreamSelectionProps = {
    torrent: HibikeTorrent_AnimeTorrent
    entry: Anime_Entry
    episodeNumber: number
    aniDBEpisode: string
    chosenFileId: string
}

export function useHandleStartDebridStream() {

    const { mutate, isPending } = useDebridStartStream()

    const { torrentStreamingPlayback } = useCurrentDevicePlaybackSettings()
    const { externalPlayerLink } = useExternalPlayerLink()
    const clientId = useAtomValue(clientIdAtom)

    const [state, setState] = useAtom(__debridstream_stateAtom)

    const playbackType = React.useMemo<Torrentstream_PlaybackType>(() => {
        if (!externalPlayerLink?.length) {
            return "default"
        }
        if (torrentStreamingPlayback === PlaybackTorrentStreaming.ExternalPlayerLink) {
            return "externalPlayerLink"
        }
        return "default"
    }, [torrentStreamingPlayback, externalPlayerLink])

    const handleStreamSelection = React.useCallback((params: DebridStreamSelectionProps) => {
        mutate({
            mediaId: params.entry.mediaId,
            episodeNumber: params.episodeNumber,
            torrent: params.torrent,
            aniDBEpisode: params.aniDBEpisode,
            fileId: params.chosenFileId,
            playbackType: playbackType,
            clientId: clientId || "",
        }, {
            onSuccess: () => {
            },
            onError: () => {
                setState(null)
            },
        })
    }, [playbackType, clientId])

    return {
        handleStreamSelection,
        isPending,
    }
}
