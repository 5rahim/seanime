import { Anime_Entry, Anime_Episode, HibikeTorrent_AnimeTorrent, Torrentstream_PlaybackType } from "@/api/generated/types"
import { useTorrentstreamStartStream } from "@/api/hooks/torrentstream.hooks"
import { PlaybackTorrentStreaming, useCurrentDevicePlaybackSettings, useExternalPlayerLink } from "@/app/(main)/_atoms/playback.atoms"
import { useHandleStartDebridStream } from "@/app/(main)/entry/_containers/debrid-stream/_lib/handle-debrid-stream"
import {
    __torrentstream__loadingStateAtom,
    __torrentstream__stateAtom,
    TorrentStreamState,
} from "@/app/(main)/entry/_containers/torrent-stream/torrent-stream-overlay"
import { clientIdAtom } from "@/app/websocket-provider"
import { useAtomValue } from "jotai"
import { atom } from "jotai/index"
import { useAtom, useSetAtom } from "jotai/react"
import React from "react"
import { toast } from "sonner"

type ManualTorrentStreamSelectionProps = {
    torrent: HibikeTorrent_AnimeTorrent
    entry: Anime_Entry
    episodeNumber: number
    aniDBEpisode: string
    chosenFileIndex: number | undefined | null
}
type AutoSelectTorrentStreamProps = {
    entry: Anime_Entry
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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type TorrentStreamAutoplayInfo = {
    allEpisodes: Anime_Episode[]
    entry: Anime_Entry
    episodeNumber: number
    aniDBEpisode: string
}
const __torrentstream_autoplayAtom = atom<TorrentStreamAutoplayInfo | null>(null)

export function useTorrentStreamAutoplay() {
    const [info, setInfo] = useAtom(__torrentstream_autoplayAtom)

    const { handleAutoSelectTorrentStream } = useHandleStartTorrentStream()

    function handleAutoplayNextTorrentstreamEpisode() {
        if (!info) return
        const { entry, episodeNumber, aniDBEpisode, allEpisodes } = info
        handleAutoSelectTorrentStream({ entry, episodeNumber: episodeNumber, aniDBEpisode })

        const nextEpisode = allEpisodes?.find(e => e.episodeNumber === episodeNumber + 1)
        if (nextEpisode && !!nextEpisode.aniDBEpisode) {
            setInfo({
                allEpisodes,
                entry,
                episodeNumber: nextEpisode.episodeNumber,
                aniDBEpisode: nextEpisode.aniDBEpisode,
            })
        } else {
            setInfo(null)
        }

        toast.info("Requesting next torrent")
    }

    return {
        hasNextTorrentstreamEpisode: !!info,
        setTorrentstreamAutoplayInfo: setInfo,
        autoplayNextTorrentstreamEpisode: handleAutoplayNextTorrentstreamEpisode,
        resetTorrentstreamAutoplayInfo: () => setInfo(null),
    }
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type DebridStreamAutoplayInfo = {
    allEpisodes: Anime_Episode[]
    entry: Anime_Entry
    episodeNumber: number
    aniDBEpisode: string
}
const __debridstream_autoplayAtom = atom<DebridStreamAutoplayInfo | null>(null)

export function useDebridStreamAutoplay() {
    const [info, setInfo] = useAtom(__debridstream_autoplayAtom)

    const { handleAutoSelectStream } = useHandleStartDebridStream()

    function handleAutoplayNextTorrentstreamEpisode() {
        if (!info) return
        const { entry, episodeNumber, aniDBEpisode, allEpisodes } = info
        handleAutoSelectStream({ entry, episodeNumber: episodeNumber, aniDBEpisode })

        const nextEpisode = allEpisodes?.find(e => e.episodeNumber === episodeNumber + 1)
        if (nextEpisode && !!nextEpisode.aniDBEpisode) {
            setInfo({
                allEpisodes,
                entry,
                episodeNumber: nextEpisode.episodeNumber,
                aniDBEpisode: nextEpisode.aniDBEpisode,
            })
        } else {
            setInfo(null)
        }

        toast.info("Requesting next torrent")
    }

    return {
        hasNextDebridstreamEpisode: !!info,
        setDebridstreamAutoplayInfo: setInfo,
        autoplayNextDebridstreamEpisode: handleAutoplayNextTorrentstreamEpisode,
        resetDebridstreamAutoplayInfo: () => setInfo(null),
    }
}
