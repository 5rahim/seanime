import { HibikeTorrent_AnimeTorrent, HibikeTorrent_BatchEpisodeFiles, Torrentstream_PlaybackType } from "@/api/generated/types"
import { useDebridStartStream } from "@/api/hooks/debrid.hooks"
import {
    ElectronPlaybackMethod,
    PlaybackTorrentStreaming,
    useCurrentDevicePlaybackSettings,
    useExternalPlayerLink,
} from "@/app/(main)/_atoms/playback.atoms"
import { __debridstream_stateAtom } from "@/app/(main)/entry/_containers/debrid-stream/debrid-stream-overlay"
import { clientIdAtom } from "@/app/websocket-provider"
import { logger } from "@/lib/helpers/debug"
import { __isElectronDesktop__ } from "@/types/constants"
import { useQueryClient } from "@tanstack/react-query"
import { useAtomValue } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"

type DebridStreamSelectionProps = {
    torrent: HibikeTorrent_AnimeTorrent
    mediaId: number
    episodeNumber: number
    aniDBEpisode: string
    chosenFileId: string
    batchEpisodeFiles: HibikeTorrent_BatchEpisodeFiles | undefined
}
type DebridStreamAutoSelectProps = {
    mediaId: number
    episodeNumber: number
    aniDBEpisode: string
}

export function useHandleStartDebridStream() {

    const { mutate, isPending } = useDebridStartStream()
    const qc = useQueryClient()

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
    }, [torrentStreamingPlayback, externalPlayerLink, electronPlaybackMethod])

    const handleStreamSelection = React.useCallback((params: DebridStreamSelectionProps) => {
        logger("DEBRID STREAM SELECTION").info("Starting debrid stream", params)
        mutate({
            mediaId: params.mediaId,
            episodeNumber: params.episodeNumber,
            torrent: params.torrent,
            aniDBEpisode: params.aniDBEpisode,
            fileId: params.chosenFileId,
            playbackType: playbackType,
            clientId: clientId || "",
            autoSelect: false,
            batchEpisodeFiles: params.batchEpisodeFiles,
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
            mediaId: params.mediaId,
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
