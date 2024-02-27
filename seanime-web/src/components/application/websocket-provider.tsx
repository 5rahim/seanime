import { websocketAtom, WebSocketContext } from "@/atoms/websocket"
import { Spinner } from "@/components/ui/loading-spinner"
import { useAtom } from "jotai/react"
import React, { useState } from "react"
import { useEffectOnce } from "react-use"

export function WebsocketProvider({ children }: { children: React.ReactNode }) {
    const [socket, setSocket] = useAtom(websocketAtom)
    const [isConnected, setIsConnected] = useState(false)

    useEffectOnce(() => {

        function connectWebSocket() {
            const newSocket = new WebSocket(`ws://${process.env.NODE_ENV === "development"
                ? `${window.location.hostname}:43211`
                : window.location.host}/events`)

            newSocket.addEventListener("open", () => {
                console.log("WebSocket connection opened")
                setIsConnected(true)
            })

            newSocket.addEventListener("close", () => {
                console.log("WebSocket connection closed")
                setIsConnected(false)
                // Reconnect after a delay
                setTimeout(connectWebSocket, 3000)
            })

            setSocket(newSocket)

            return newSocket
        }

        if (!socket || socket.readyState === WebSocket.CLOSED) {
            // If the socket is not set or the connection is closed, initiate a new connection
            const newSocket = connectWebSocket()
        }

        return () => {
            if (socket) {
                socket.close()
                setIsConnected(false)
            }
        }
    })

    return (
        <WebSocketContext.Provider value={socket}>
            {!isConnected && <div
                className="fixed right-4 bottom-4 bg-gray-800 border  text-gray-100 py-3 px-5 font-semibold rounded-md z-[100] flex gap-2 items-center"
            >
                <Spinner className="w-5 h-5"/>
                Websocket connection
            </div>}
            {children}
        </WebSocketContext.Provider>
    )

}
