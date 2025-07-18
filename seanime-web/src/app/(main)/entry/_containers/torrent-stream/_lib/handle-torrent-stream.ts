import { Anime_Entry, Anime_Episode, HibikeTorrent_AnimeTorrent, Torrentstream_PlaybackType } from "@/api/generated/types"
import { useTorrentstreamStartStream } from "@/api/hooks/torrentstream.hooks"
import {
    ElectronPlaybackMethod,
    PlaybackTorrentStreaming,
    useCurrentDevicePlaybackSettings,
    useExternalPlayerLink,
} from "@/app/(main)/_atoms/playback.atoms"
import { __autoplay_nextEpisodeAtom } from "@/app/(main)/_features/progress-tracking/_lib/autoplay"
import { useHandleStartDebridStream } from "@/app/(main)/entry/_containers/debrid-stream/_lib/handle-debrid-stream"
import {
    __torrentstream__isLoadedAtom,
    __torrentstream__loadingStateAtom,
} from "@/app/(main)/entry/_containers/torrent-stream/torrent-stream-overlay"
import { clientIdAtom } from "@/app/websocket-provider"
import { __isElectronDesktop__ } from "@/types/constants"
import { atom, useAtomValue } from "jotai"
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
    const setIsLoaded = useSetAtom(__torrentstream__isLoadedAtom)
    const { torrentStreamingPlayback, electronPlaybackMethod } = useCurrentDevicePlaybackSettings()
    const { externalPlayerLink } = useExternalPlayerLink()
    const clientId = useAtomValue(clientIdAtom)

    const playbackType = React.useMemo<Torrentstream_PlaybackType>(() => {
        if (__isElectronDesktop__ && electronPlaybackMethod === ElectronPlaybackMethod.NativePlayer) {
            return "nativeplayer"
        }
        if (!!externalPlayerLink?.length && torrentStreamingPlayback === PlaybackTorrentStreaming.ExternalPlayerLink) {
            return "externalPlayerLink"
        }
        return "default"
    }, [torrentStreamingPlayback, externalPlayerLink, electronPlaybackMethod])

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
                setIsLoaded(false)
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
                setIsLoaded(false)
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
    const [nextEpisode, setNextEpisode] = useAtom(__autoplay_nextEpisodeAtom)

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
            setNextEpisode(nextEpisode)
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
    const [nextEpisode, setNextEpisode] = useAtom(__autoplay_nextEpisodeAtom)

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
            setNextEpisode(nextEpisode)
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
