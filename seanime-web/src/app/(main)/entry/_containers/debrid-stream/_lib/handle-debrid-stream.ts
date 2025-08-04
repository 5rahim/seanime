import { Anime_Entry, HibikeTorrent_AnimeTorrent, Torrentstream_PlaybackType } from "@/api/generated/types"
import { useDebridStartStream } from "@/api/hooks/debrid.hooks"
import {
    ElectronPlaybackMethod,
    PlaybackTorrentStreaming,
    useCurrentDevicePlaybackSettings,
    useExternalPlayerLink,
} from "@/app/(main)/_atoms/playback.atoms"
import { __debridstream_stateAtom } from "@/app/(main)/entry/_containers/debrid-stream/debrid-stream-overlay"
import { clientIdAtom } from "@/app/websocket-provider"
import { __isElectronDesktop__ } from "@/types/constants"
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
type DebridStreamAutoSelectProps = {
    entry: Anime_Entry
    episodeNumber: number
    aniDBEpisode: string
}

export function useHandleStartDebridStream() {

    const { mutate, isPending } = useDebridStartStream()

    const { torrentStreamingPlayback, electronPlaybackMethod } = useCurrentDevicePlaybackSettings()
    const { externalPlayerLink } = useExternalPlayerLink()
    const clientId = useAtomValue(clientIdAtom)

    const [state, setState] = useAtom(__debridstream_stateAtom)

    const playbackType = React.useMemo<Torrentstream_PlaybackType>(() => {
        if (__isElectronDesktop__ && electronPlaybackMethod === ElectronPlaybackMethod.NativePlayer) {
            return "nativeplayer"
        }
        if (!!externalPlayerLink?.length && torrentStreamingPlayback === PlaybackTorrentStreaming.ExternalPlayerLink) {
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
            autoSelect: false,
        }, {
            onSuccess: () => {
            },
            onError: () => {
                setState(null)
            },
        })
    }, [playbackType, clientId])

    const handleAutoSelectStream = React.useCallback((params: DebridStreamAutoSelectProps) => {
        mutate({
            mediaId: params.entry.mediaId,
            episodeNumber: params.episodeNumber,
            torrent: undefined,
            aniDBEpisode: params.aniDBEpisode,
            fileId: "",
            playbackType: playbackType,
            clientId: clientId || "",
            autoSelect: true,
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
        handleAutoSelectStream,
        isPending,
    }
}
