import { Nullish } from "@/api/generated/types"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import { FaShareFromSquare } from "react-icons/fa6"
import { PiVideoFill } from "react-icons/pi"

export const enum ElectronPlaybackMethod {
    NativePlayer = "nativePlayer", // Desktop media player or Integrated player (media streaming)
    Default = "default", // Desktop media player, media streaming or external player link
}

export const __playback_electronPlaybackMethodAtom = atomWithStorage<string>("sea-playback-electron-playback-method",
    ElectronPlaybackMethod.NativePlayer)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const enum PlaybackDownloadedMedia {
    Default = "default", // Desktop media player or Integrated player (media streaming)
    ExternalPlayerLink = "externalPlayerLink", // External player link
}


export const playbackDownloadedMediaOptions = [
    {
        label: <div className="flex items-center gap-4 md:gap-2 w-full">
            <PiVideoFill className="text-2xl flex-none" />
            <p className="max-w-[90%]">Desktop media player or Transcoding / Direct Play</p>
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
    const [electronPlaybackMethod, setElectronPlaybackMethod] = useAtom(__playback_electronPlaybackMethodAtom)
    return {
        downloadedMediaPlayback,
        setDownloadedMediaPlayback,
        torrentStreamingPlayback,
        setTorrentStreamingPlayback,
        electronPlaybackMethod,
        setElectronPlaybackMethod,
    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const __playback_externalPlayerLink = atomWithStorage<string>("sea-playback-external-player-link", "")
export const __playback_externalPlayerLink_encodePath = atomWithStorage<boolean>("sea-playback-external-player-link-encode-path", false)

export function useExternalPlayerLink() {
    const [externalPlayerLink, setExternalPlayerLink] = useAtom(__playback_externalPlayerLink)
    const [encodePath, setEncodePath] = useAtom(__playback_externalPlayerLink_encodePath)
    return {
        externalPlayerLink,
        setExternalPlayerLink,
        encodePath,
        setEncodePath,
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
