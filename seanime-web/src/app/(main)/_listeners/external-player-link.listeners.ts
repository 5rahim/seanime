import { usePlaybackStartManualTracking } from "@/api/hooks/playback_manager.hooks"
import { useExternalPlayerLink } from "@/app/(main)/_atoms/playback.atoms"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { clientIdAtom } from "@/app/websocket-provider"
import { ExternalPlayerLink } from "@/lib/external-player-link/external-player-link"
import { openTab } from "@/lib/helpers/browser"
import { logger } from "@/lib/helpers/debug"
import { WSEvents } from "@/lib/server/ws-events"
import { useAtomValue } from "jotai"
import { toast } from "sonner"

type ExternalPlayerLinkEventProps = {
    url: string
    mediaId: number
    episodeNumber: number
    mediaTitle?: string
}

export function useExternalPlayerLinkListener() {

    const clientId = useAtomValue(clientIdAtom)
    const { externalPlayerLink } = useExternalPlayerLink()

    const { mutate: startManualTracking } = usePlaybackStartManualTracking()

    useWebsocketMessageListener<ExternalPlayerLinkEventProps>({
        type: WSEvents.EXTERNAL_PLAYER_OPEN_URL,
        onMessage: data => {
            if (!externalPlayerLink?.length) {
                toast.error("External player link is not set.")
                return
            }

            toast.info("Opening media file in external player.")

            logger("EXTERNAL PLAYER LINK").info("Opening external player", data)

            const link = new ExternalPlayerLink(externalPlayerLink)
            link.setEpisodeNumber(data.episodeNumber)
            link.setMediaTitle(data.mediaTitle)
            link.setUrl(data.url)
            openTab(link.getFullUrl())

            if (data.mediaId != 0) {
                logger("EXTERNAL PLAYER LINK").info("Starting manual tracking", {
                    mediaId: data.mediaId,
                    episodeNumber: data.episodeNumber,
                    clientId: clientId || "",
                })

                // Get the server to start asking the progress
                startManualTracking({
                    mediaId: data.mediaId,
                    episodeNumber: data.episodeNumber,
                    clientId: clientId || "",
                })
            } else {
                logger("EXTERNAL PLAYER LINK").info("No manual tracking", {
                    url: data.url,
                    mediaId: data.mediaId,
                    episodeNumber: data.episodeNumber,
                })
            }
        },
    })

}
