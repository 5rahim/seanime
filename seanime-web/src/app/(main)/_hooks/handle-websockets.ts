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

    // Store message queue in a ref so it persists across rerenders and socket changes
    const messageQueue = useRef<SeaWebsocketEvent<any>[]>([])
    const processingQueueRef = useRef<NodeJS.Timeout | null>(null)

    // Keep a local latest reference to socket to ensure we're using the most recent one
    const latestSocketRef = useRef<WebSocket | null>(null)

    // Update the socket ref whenever the socket changes
    useEffect(() => {
        latestSocketRef.current = socket

        // Log socket changes
        // logger("WebsocketSender").info(`Socket updated: ${socket ? getReadyStateString(socket.readyState) : "null"}`)

        // When socket becomes available and open, immediately process any queued messages
        if (socket && socket.readyState === WebSocket.OPEN && messageQueue.current.length > 0) {
            // logger("WebsocketSender").info(`New socket connected with ${messageQueue.current.length} queued messages, processing immediately`)
            setTimeout(() => processQueue(), 100) // Small delay to ensure socket is fully established
        }
    }, [socket])

    // Add debug log for socket state changes
    useEffect(() => {
        if (socket) {
            // logger("WebsocketSender").info(`Socket state changed: ${getReadyStateString(socket.readyState)}`)

            // If socket becomes open, process queue immediately
            if (socket.readyState === WebSocket.OPEN && messageQueue.current.length > 0) {
                // logger("WebsocketSender").info(`Socket became OPEN with ${messageQueue.current.length} messages in queue`)
                setTimeout(() => processQueue(), 100) // Small delay to ensure socket is fully established
            }
        }
    }, [socket?.readyState])

    function getReadyStateString(state?: number): string {
        if (state === undefined) return "UNDEFINED"
        switch (state) {
            case WebSocket.CONNECTING:
                return "CONNECTING"
            case WebSocket.OPEN:
                return "OPEN"
            case WebSocket.CLOSING:
                return "CLOSING"
            case WebSocket.CLOSED:
                return "CLOSED"
            default:
                return `UNKNOWN (${state})`
        }
    }

    function sendMessage<TData>(data: SeaWebsocketEvent<TData>) {
        // Always use the latest socket reference
        const currentSocket = latestSocketRef.current

        if (currentSocket && currentSocket.readyState === WebSocket.OPEN) {
            try {
                const message = JSON.stringify({ ...data, clientId: clientId })
                currentSocket.send(message)
                // logger("WebsocketSender").info(`Sent message of type ${data.type}`);
                return true
            }
            catch (e) {
                // logger("WebsocketSender").error(`Failed to send message of type ${data.type}`, e)
                messageQueue.current.push(data)
                return false
            }
        } else {
            if (messageQueue.current.length < 500) { // Limit queue size to prevent memory issues
                messageQueue.current.push(data)
                logger("WebsocketSender")
                    .info(`Queued message of type ${data.type}, queue size: ${messageQueue.current.length}, socket state: ${currentSocket
                        ? getReadyStateString(currentSocket.readyState)
                        : "null"}`)
            } else {
                logger("WebsocketSender").warning(`Message queue full (500), dropping message of type ${data.type}`)
            }

            // Always ensure queue processor is running
            ensureQueueProcessorIsRunning()
            return false
        }
    }

    function ensureQueueProcessorIsRunning() {
        if (!processingQueueRef.current) {
            processQueue()
        }
    }

    function processQueue() {
        // Clear any existing processor
        if (processingQueueRef.current) {
            clearTimeout(processingQueueRef.current)
            processingQueueRef.current = null
        }

        // Always use the latest socket reference
        const currentSocket = latestSocketRef.current

        // Process the queue if socket is connected
        if (currentSocket && currentSocket.readyState === WebSocket.OPEN && messageQueue.current.length > 0) {
            // logger("WebsocketSender").info(`Processing ${messageQueue.current.length} queued messages`)

            // Create a copy of the queue to avoid modification issues during iteration
            const queueCopy = [...messageQueue.current]
            const successfulMessages: number[] = []

            // Try to send all queued messages
            queueCopy.forEach((message, index) => {
                try {
                    const messageStr = JSON.stringify({ ...message, clientId: clientId })
                    currentSocket.send(messageStr)
                    successfulMessages.push(index)
                    // logger("WebsocketSender").info(`Successfully sent queued message of type ${message.type}`);
                }
                catch (e) {
                    // logger("WebsocketSender").error(`Failed to send queued message of type ${message.type}`, e);
                }
            })

            // Remove successfully sent messages
            if (successfulMessages.length > 0) {
                // Create a new array without the successfully sent messages
                messageQueue.current = queueCopy.filter((_, index) => !successfulMessages.includes(index))
                // logger("WebsocketSender").info(`Sent ${successfulMessages.length}/${queueCopy.length} queued messages, ${messageQueue.current.length} remaining`)
            }
        } else {
            // const reason = !currentSocket ? "no socket" :
            //                currentSocket.readyState !== WebSocket.OPEN ? `socket not open (${getReadyStateString(currentSocket.readyState)})` :
            //                "no messages in queue";
            // logger("WebsocketSender").info(`Skipped queue processing: ${reason}`);
        }

        // Always schedule next run if there are messages or socket isn't ready
        const shouldReschedule = messageQueue.current.length > 0 || !currentSocket || currentSocket.readyState !== WebSocket.OPEN

        if (shouldReschedule) {
            processingQueueRef.current = setTimeout(() => {
                processQueue()
            }, 1000) // Process every second for faster recovery
        } else {
            processingQueueRef.current = null
        }
    }

    // Process queue whenever connection status changes
    useEffect(() => {
        if (isConnected && latestSocketRef.current?.readyState === WebSocket.OPEN) {
            // logger("WebsocketSender").info(`Connection reestablished, processing message queue (${messageQueue.current.length} messages)`)
            // Force immediate processing with a small delay to ensure everything is ready
            setTimeout(() => processQueue(), 100)
        }

        return () => {
            if (processingQueueRef.current) {
                clearTimeout(processingQueueRef.current)
                processingQueueRef.current = null
            }
        }
    }, [isConnected]);

    return {
        sendMessage,
        sendPluginMessage: (type: string, payload: any, extensionId?: string) => {
            return sendMessage({
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
