import { Nullish } from "@/api/generated/types"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import { FaShareFromSquare } from "react-icons/fa6"
import { PiVideoFill } from "react-icons/pi"

export const enum PlaybackDownloadedMedia {
    Default = "default", // Built-in player
    ExternalPlayerLink = "externalPlayerLink",
}

export const playbackDownloadedMediaOptions = [
    {
        label: <div className="flex items-center gap-4 md:gap-2 w-full">
            <PiVideoFill className="text-2xl flex-none" />
            <p className="max-w-[90%]">Desktop media player or Integrated player (media streaming)</p>
        </div>, value: PlaybackDownloadedMedia.Default,
    },
    {
        label: <div className="flex items-center gap-4 md:gap-2 w-full">
            <FaShareFromSquare className="text-2xl flex-none" />
            <p className="max-w-[90%]">External player link</p>
        </div>, value: PlaybackDownloadedMedia.ExternalPlayerLink,
    },
]

export const __playback_downloadedMediaAtom = atomWithStorage<string>("sea-playback-downloaded-media", PlaybackDownloadedMedia.Default)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const enum PlaybackTorrentStreaming {
    Default = "default", // Desktop media player
    ExternalPlayerLink = "externalPlayerLink",
}

export const playbackTorrentStreamingOptions = [
    {
        label: <div className="flex items-center gap-4 md:gap-2 w-full">
            <PiVideoFill className="text-2xl flex-none" />
            <p className="max-w-[90%]">Desktop media player</p>
        </div>, value: PlaybackTorrentStreaming.Default,
    },
    {
        label: <div className="flex items-center gap-4 md:gap-2 w-full">
            <FaShareFromSquare className="text-2xl flex-none" />
            <p className="max-w-[90%]">External player link</p>
        </div>, value: PlaybackTorrentStreaming.ExternalPlayerLink,
    },
]


export const __playback_torrentStreamingAtom = atomWithStorage<string>("sea-playback-torrentstream", PlaybackTorrentStreaming.Default)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export function useCurrentDevicePlaybackSettings() {

    const [downloadedMediaPlayback, setDownloadedMediaPlayback] = useAtom(__playback_downloadedMediaAtom)
    const [torrentStreamingPlayback, setTorrentStreamingPlayback] = useAtom(__playback_torrentStreamingAtom)

    return {
        downloadedMediaPlayback,
        setDownloadedMediaPlayback,
        torrentStreamingPlayback,
        setTorrentStreamingPlayback,
    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const __playback_externalPlayerLink = atomWithStorage<string>("sea-playback-external-player-link", "")

export function useExternalPlayerLink() {
    const [externalPlayerLink, setExternalPlayerLink] = useAtom(__playback_externalPlayerLink)

    return {
        externalPlayerLink,
        setExternalPlayerLink,
    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const __playback_playNext = atom<number | null>(null)

export function usePlayNext() {
    const [playNext, _setPlayNext] = useAtom(__playback_playNext)

    function setPlayNext(ep: Nullish<number>, callback: () => void) {
        if (!ep) return
        _setPlayNext(ep)
        callback()
    }

    return {
        playNext,
        setPlayNext,
        resetPlayNext: () => _setPlayNext(null),
    }
}
