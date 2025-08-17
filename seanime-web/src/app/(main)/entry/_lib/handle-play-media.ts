import { getExternalPlayerURL } from "@/api/client/external-player-link"
import { getServerBaseUrl } from "@/api/client/server-url"
import { Anime_Episode } from "@/api/generated/types"
import { useDirectstreamPlayLocalFile } from "@/api/hooks/directstream.hooks"
import { useNakamaPlayVideo } from "@/api/hooks/nakama.hooks"
import { usePlaybackPlayVideo, usePlaybackStartManualTracking } from "@/api/hooks/playback_manager.hooks"
import {
    ElectronPlaybackMethod,
    PlaybackDownloadedMedia,
    useCurrentDevicePlaybackSettings,
    useExternalPlayerLink,
} from "@/app/(main)/_atoms/playback.atoms"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useTorrentStreamAutoplay } from "@/app/(main)/entry/_containers/torrent-stream/_lib/handle-torrent-stream"
import { useMediastreamActiveOnDevice, useMediastreamCurrentFile } from "@/app/(main)/mediastream/_lib/mediastream.atoms"
import { clientIdAtom } from "@/app/websocket-provider"
import { openTab } from "@/lib/helpers/browser"
import { logger } from "@/lib/helpers/debug"
import { __isElectronDesktop__ } from "@/types/constants"
import { useAtomValue } from "jotai"
import { useRouter } from "next/navigation"
import React from "react"
import { toast } from "sonner"

export function useHandlePlayMedia() {
    const router = useRouter()
    const serverStatus = useServerStatus()
    const clientId = useAtomValue(clientIdAtom)

    const { activeOnDevice: mediastreamActiveOnDevice } = useMediastreamActiveOnDevice()
    const { setFilePath: setMediastreamFilePath } = useMediastreamCurrentFile()

    const { mutate: startManualTracking, isPending: isStarting } = usePlaybackStartManualTracking()

    const { downloadedMediaPlayback, electronPlaybackMethod } = useCurrentDevicePlaybackSettings()
    const { externalPlayerLink } = useExternalPlayerLink()

    // Play using desktop external player
    const { mutate: playVideo } = usePlaybackPlayVideo()
    const { mutate: playNakamaVideo } = useNakamaPlayVideo()

    const { mutate: directstreamPlayLocalFile } = useDirectstreamPlayLocalFile()

    const { setTorrentstreamAutoplayInfo } = useTorrentStreamAutoplay()

    function playMediaFile({ path, mediaId, episode }: { path: string, mediaId: number, episode: Anime_Episode }) {
        const anidbEpisode = episode.localFile?.metadata?.aniDBEpisode ?? ""

        setTorrentstreamAutoplayInfo(null)

        if (episode._isNakamaEpisode) {
            // If external player link is set, open the media file in the external player
            if (downloadedMediaPlayback === PlaybackDownloadedMedia.ExternalPlayerLink) {
                let urlToSend = getServerBaseUrl() + "/api/v1/nakama/stream?type=file&path=" + Buffer.from(path).toString("base64")
                logger("PLAY MEDIA").info("Opening external player", externalPlayerLink, "URL", urlToSend)

                // If the external player link includes a query parameter, we need to encode the URL to prevent query parameter conflicts
                if (externalPlayerLink.includes("?")) {
                    urlToSend = encodeURIComponent(urlToSend)
                }

                openTab(getExternalPlayerURL(externalPlayerLink, urlToSend))

                if (episode?.progressNumber && episode.type === "main") {
                    logger("PLAY MEDIA").error("Starting manual tracking for nakama file")
                    // Start manual tracking
                    React.startTransition(() => {
                        startManualTracking({
                            mediaId: mediaId,
                            episodeNumber: episode?.progressNumber,
                            clientId: clientId || "",
                        })
                    })
                } else {
                    logger("PLAY MEDIA").warning("No manual tracking, progress number is not set for nakama file")
                }
            }
            return playNakamaVideo({ path, mediaId, anidbEpisode })
        }

        logger("PLAY MEDIA").info("Playing media file", path)

        //
        // Electron native player
        //
        if (__isElectronDesktop__ && electronPlaybackMethod === ElectronPlaybackMethod.NativePlayer) {
            directstreamPlayLocalFile({ path, clientId: clientId ?? "" })
            return
        }

        // If external player link is set, open the media file in the external player
        if (downloadedMediaPlayback === PlaybackDownloadedMedia.ExternalPlayerLink) {
            if (!externalPlayerLink) {
                toast.error("External player link is not set.")
                return
            }

            logger("PLAY MEDIA").info("Opening media file in external player", externalPlayerLink, path)

            setMediastreamFilePath(path)
            React.startTransition(() => {
                router.push(`/medialinks?id=${mediaId}`)
            })
            return
        }

        // Handle media streaming
        if (serverStatus?.mediastreamSettings?.transcodeEnabled && mediastreamActiveOnDevice) {
            setMediastreamFilePath(path)
            React.startTransition(() => {
                router.push(`/mediastream?id=${mediaId}`)
            })
            return
        }

        return playVideo({ path })
    }

    return {
        playMediaFile,
    }
}
