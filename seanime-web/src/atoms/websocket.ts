import { createContext, useContext, useEffect } from "react"
import { atom } from "jotai"
import { SeaWebsocketEvent } from "@/lib/server/types"
import { WSEvents } from "@/lib/server/endpoints"

export const WebSocketContext = createContext<WebSocket | null>(null)

export const websocketAtom = atom<WebSocket | null>(null)

export function useWebsocketSender() {
    const socket = useContext(WebSocketContext)

    const send = (message: string) => {
        if (socket && socket.readyState === WebSocket.OPEN) {
            socket.send(message)
        }
    }

    return {
        send,
    }

}

export type WebSocketListener<TData = any> = {
    onMessage: (data: TData) => void
}

export function useWebsocketListener<TData = any>({ onMessage }: WebSocketListener<TData>) {
    const socket = useContext(WebSocketContext)

    useEffect(() => {
        if (socket) {
            const messageHandler = (event: MessageEvent) => {
                onMessage(event.data)
            }

            socket.addEventListener("message", messageHandler)

            return () => {
                socket.removeEventListener("message", messageHandler)
            }
        }
    }, [socket, onMessage])

    return null
}

export type WebSocketMessageListener<TData> = {
    type: WSEvents
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
                } catch (e) {

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