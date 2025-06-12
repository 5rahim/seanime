import { Anime_Episode } from "@/api/generated/types"
import { useDirectstreamPlayLocalFile } from "@/api/hooks/directstream.hooks"
import { useNakamaPlayVideo } from "@/api/hooks/nakama.hooks"
import { usePlaybackPlayVideo } from "@/api/hooks/playback_manager.hooks"
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
            return playNakamaVideo({ path, mediaId, anidbEpisode })
        }

        logger("PLAY_MEDIA").info("Playing media file", path)

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

            logger("PLAY_MEDIA").info("Opening media file in external player", externalPlayerLink, path)

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
