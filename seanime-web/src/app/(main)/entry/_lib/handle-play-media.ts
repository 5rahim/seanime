import { usePlaybackPlayVideo } from "@/api/hooks/playback_manager.hooks"
import { PlaybackDownloadedMedia, useCurrentDevicePlaybackSettings, useExternalPlayerLink } from "@/app/(main)/_atoms/playback.atoms"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useMediastreamActiveOnDevice, useMediastreamCurrentFile } from "@/app/(main)/mediastream/_lib/mediastream.atoms"
import { logger } from "@/lib/helpers/debug"
import { useRouter } from "next/navigation"
import React from "react"
import { toast } from "sonner"

export function useHandlePlayMedia() {
    const router = useRouter()
    const serverStatus = useServerStatus()

    const { activeOnDevice: mediastreamActiveOnDevice } = useMediastreamActiveOnDevice()
    const { setFilePath: setMediastreamFilePath } = useMediastreamCurrentFile()

    const { downloadedMediaPlayback } = useCurrentDevicePlaybackSettings()
    const { externalPlayerLink } = useExternalPlayerLink()

    // Play using desktop external player
    const { mutate: playVideo } = usePlaybackPlayVideo()


    function playMediaFile({ path, mediaId }: { path: string, mediaId: number }) {

        logger("PLAY_MEDIA").info("Playing media file", path)

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
