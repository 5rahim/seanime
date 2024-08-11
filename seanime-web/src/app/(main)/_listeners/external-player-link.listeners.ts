import { usePlaybackStartManualTracking } from "@/api/hooks/playback_manager.hooks"
import { useExternalPlayerLink } from "@/app/(main)/_atoms/playback.atoms"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { clientIdAtom } from "@/app/websocket-provider"
import { WSEvents } from "@/lib/server/ws-events"
import { useAtomValue } from "jotai"
import { toast } from "sonner"

type ExternalPlayerLinkEventProps = {
    url: string
    mediaId: number
    episodeNumber: number
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

            const url = externalPlayerLink.replace("{url}", data.url)
                .replace("127.0.0.1", window.location.hostname)
                .replace("localhost", window.location.hostname)

            window.open(url, "_blank")

            // Get the server to start asking the progress
            startManualTracking({
                mediaId: data.mediaId,
                episodeNumber: data.episodeNumber,
                clientId: clientId || "",
            })
        },
    })

}
