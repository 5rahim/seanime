import { useAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"

export const enum PlaybackDownloadedMedia {
    Default = "default", // Built-in player
    ExternalPlayerLink = "externalPlayerLink",
}

export const playbackDownloadedMediaOptions = [
    { label: "Desktop media player / Built-in player (media streaming)", value: PlaybackDownloadedMedia.Default },
    { label: "External player link", value: PlaybackDownloadedMedia.ExternalPlayerLink },
]

export const __playback_downloadedMediaAtom = atomWithStorage<string>("sea-playback-downloaded-media", PlaybackDownloadedMedia.Default)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const enum PlaybackTorrentStreaming {
    Default = "default", // Desktop media player
    ExternalPlayerLink = "externalPlayerLink",
}

export const playbackTorrentStreamingOptions = [
    { label: "Desktop media player", value: PlaybackTorrentStreaming.Default },
    { label: "External player link", value: PlaybackTorrentStreaming.ExternalPlayerLink },
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
