import { WebSocketContext } from "@/app/(main)/_atoms/websocket.atoms"
import { clientIdAtom } from "@/app/websocket-provider"
import { logger } from "@/lib/helpers/debug"
import { SeaWebsocketEvent } from "@/lib/server/queries.types"
import { WSEvents } from "@/lib/server/ws-events"
import { useAtom } from "jotai"
import { useContext, useEffect } from "react"
import useUpdateEffect from "react-use/lib/useUpdateEffect"

export function useWebsocketSender() {
    const socket = useContext(WebSocketContext)
    const [clientId] = useAtom(clientIdAtom)


    function sendMessage<TData>(data: SeaWebsocketEvent<TData>) {
        if (socket && socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify({ ...data, clientId: clientId }))
        } else {
            logger("Websocket").error(`Socket is not open, cannot send ${data.type}`)
        }
    }

    return {
        sendMessage,
    }
}

export function useWebsocketSendEffect<TData>(data: SeaWebsocketEvent<TData>, ...deps: any[]) {
    const { sendMessage } = useWebsocketSender()

    useUpdateEffect(() => {
        sendMessage(data)
    }, [...deps])
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export type WebSocketMessageListener<TData> = {
    type: WSEvents | string
    onMessage: (data: TData) => void
}

export function useWebsocketMessageListener<TData = unknown>({ type, onMessage }: WebSocketMessageListener<TData>) {
    const socket = useContext(WebSocketContext)

    useEffect(() => {
        if (socket) {
            const messageHandler = (event: MessageEvent) => {
                try {
                    const parsed = JSON.parse(event.data) as SeaWebsocketEvent<TData>
                    if (!!parsed.type && parsed.type === type) {
                        onMessage(parsed.payload)
                    }
                }
                catch (e) {
                    logger("Websocket").error("Error parsing message", e)
                }
            }

            socket.addEventListener("message", messageHandler)

            return () => {
                socket.removeEventListener("message", messageHandler)
            }
        }
    }, [socket, onMessage])

    return null
}
