import { WebSocketContext } from "@/app/(main)/_atoms/websocket.atoms"
import { clientIdAtom, websocketConnectedAtom } from "@/app/websocket-provider"
import { logger } from "@/lib/helpers/debug"
import { SeaWebsocketEvent, SeaWebsocketPluginEvent } from "@/lib/server/queries.types"
import { WSEvents } from "@/lib/server/ws-events"
import { useAtom } from "jotai"
import { useContext, useEffect, useRef } from "react"
import useUpdateEffect from "react-use/lib/useUpdateEffect"

export function useWebsocketSender() {
    const socket = useContext(WebSocketContext)
    const [clientId] = useAtom(clientIdAtom)
    const [isConnected] = useAtom(websocketConnectedAtom)

    const messageQueue = useRef<SeaWebsocketEvent<any>[]>([])

    function sendMessage<TData>(data: SeaWebsocketEvent<TData>) {
        if (socket && socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify({ ...data, clientId: clientId }))
        } else {
            messageQueue.current.push(data)
        }
        console.log("messageQueue", messageQueue.current)
    }

    useEffect(() => {
        if (socket && socket.readyState === WebSocket.OPEN) {
            messageQueue.current.splice(0).forEach(message => {
                sendMessage(message)
            })
        }
    }, [socket, isConnected])

    return {
        sendMessage,
        sendPluginMessage: (type: string, payload: any, extensionId?: string) => {
            sendMessage({
                type: "plugin",
                payload: {
                    type: type,
                    extensionId: extensionId,
                    payload: payload,
                },
            })
        },
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

export type WebSocketPluginMessageListener<TData> = {
    type: string
    extensionId: string // If empty, get message from all plugins
    onMessage: (data: TData, extensionId: string) => void
}

export function useWebsocketPluginMessageListener<TData = unknown>({ type, extensionId, onMessage }: WebSocketPluginMessageListener<TData>) {
    const socket = useContext(WebSocketContext)

    useEffect(() => {
        if (socket) {
            const messageHandler = (event: MessageEvent) => {
                try {
                    const parsed = JSON.parse(event.data) as SeaWebsocketEvent<TData>
                    if (!!parsed.type && parsed.type === "plugin") {
                        const message = parsed.payload as SeaWebsocketPluginEvent<TData>
                        // Plugins always send back their extension ID
                        // Invoke the callback only if the extension ID of the message matches the ID we're listening to
                        // OR if we're listening to all plugins (i.e. extensionId is "")
                        if (message.type === type && (message.extensionId === extensionId || extensionId === "")) {
                            onMessage(message.payload, message.extensionId)
                        }
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
