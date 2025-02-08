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

type AutoplayInfo = {
    allEpisodes: Anime_Episode[]
    entry: Anime_Entry
    episodeNumber: number
    aniDBEpisode: string
    type: "torrentstream" | "debridstream"
}
const __stream_autoplayAtom = atom<AutoplayInfo | null>(null)
const __stream_autoplaySelectedTorrentAtom = atom<HibikeTorrent_AnimeTorrent | null>(null)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export function useTorrentStreamAutoplay() {
    const [info, setInfo] = useAtom(__stream_autoplayAtom)

    const { handleAutoSelectTorrentStream, handleManualTorrentStreamSelection } = useHandleStartTorrentStream()
    const [selectedTorrent, setSelectedTorrent] = useAtom(__stream_autoplaySelectedTorrentAtom)

    function handleAutoplayNextTorrentstreamEpisode() {
        if (!info) return
        const { entry, episodeNumber, aniDBEpisode, allEpisodes } = info

        if (selectedTorrent?.isBatch) {
            // If the user provided a torrent, use it
            handleManualTorrentStreamSelection({
                entry,
                episodeNumber: episodeNumber,
                aniDBEpisode: aniDBEpisode,
                torrent: selectedTorrent,
                chosenFileIndex: undefined,
            })
        } else {
            // Otherwise, use the auto-select function
            handleAutoSelectTorrentStream({ entry, episodeNumber: episodeNumber, aniDBEpisode })
        }

        const nextEpisode = allEpisodes?.find(e => e.episodeNumber === episodeNumber + 1)
        if (nextEpisode && !!nextEpisode.aniDBEpisode) {
            setInfo({
                allEpisodes,
                entry,
                episodeNumber: nextEpisode.episodeNumber,
                aniDBEpisode: nextEpisode.aniDBEpisode,
                type: "torrentstream",
            })
        } else {
            setInfo(null)
        }

        toast.info("Requesting next episode")
    }


    return {
        hasNextTorrentstreamEpisode: !!info && info.type === "torrentstream",
        setTorrentstreamAutoplayInfo: setInfo,
        autoplayNextTorrentstreamEpisode: handleAutoplayNextTorrentstreamEpisode,
        resetTorrentstreamAutoplayInfo: () => setInfo(null),
        setTorrentstreamAutoplaySelectedTorrent: setSelectedTorrent,
    }
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export function useDebridStreamAutoplay() {
    const [info, setInfo] = useAtom(__stream_autoplayAtom)

    const { handleAutoSelectStream, handleStreamSelection } = useHandleStartDebridStream()
    const [selectedTorrent, setSelectedTorrent] = useAtom(__stream_autoplaySelectedTorrentAtom)

    function handleAutoplayNextTorrentstreamEpisode() {
        if (!info) return
        const { entry, episodeNumber, aniDBEpisode, allEpisodes } = info

        if (selectedTorrent?.isBatch) {
            // If the user provided a torrent, use it
            handleStreamSelection({
                entry,
                episodeNumber: episodeNumber,
                aniDBEpisode: aniDBEpisode,
                torrent: selectedTorrent,
                chosenFileId: "",
            })
        } else {
            // Otherwise, use the auto-select function
            handleAutoSelectStream({ entry, episodeNumber: episodeNumber, aniDBEpisode })
        }

        const nextEpisode = allEpisodes?.find(e => e.episodeNumber === episodeNumber + 1)
        if (nextEpisode && !!nextEpisode.aniDBEpisode) {
            setInfo({
                allEpisodes,
                entry,
                episodeNumber: nextEpisode.episodeNumber,
                aniDBEpisode: nextEpisode.aniDBEpisode,
                type: "debridstream",
            })
        } else {
            setInfo(null)
        }

        toast.info("Requesting next episode")
    }


    return {
        hasNextDebridstreamEpisode: !!info && info.type === "debridstream",
        setDebridstreamAutoplayInfo: setInfo,
        autoplayNextDebridstreamEpisode: handleAutoplayNextTorrentstreamEpisode,
        resetDebridstreamAutoplayInfo: () => setInfo(null),
        setDebridstreamAutoplaySelectedTorrent: setSelectedTorrent,
    }
}
