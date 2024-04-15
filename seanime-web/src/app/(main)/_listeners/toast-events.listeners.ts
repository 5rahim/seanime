import { useWebsocketMessageListener } from "@/atoms/websocket"
import { WSEvents } from "@/lib/server/endpoints"
import { toast } from "sonner"

export function useToastEventListeners() {

    useWebsocketMessageListener<string>({
        type: WSEvents.INFO_TOAST, onMessage: data => {
            if (!!data) {
                toast.info(data)
            }
        },
    })

    useWebsocketMessageListener<string>({
        type: WSEvents.SUCCESS_TOAST, onMessage: data => {
            if (!!data) {
                toast.success(data)
            }
        },
    })

    useWebsocketMessageListener<string>({
        type: WSEvents.WARNING_TOAST, onMessage: data => {
            if (!!data) {
                toast.warning(data)
            }
        },
    })

    useWebsocketMessageListener<string>({
        type: WSEvents.ERROR_TOAST, onMessage: data => {
            if (!!data) {
                toast.error(data)
            }
        },
    })

}
