import { getServerBaseUrl } from "@/api/client/server-url"
import { websocketAtom, WebSocketContext } from "@/app/(main)/_atoms/websocket.atoms"
import { __openDrawersAtom } from "@/components/ui/drawer"
import { atom, useAtomValue } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"
import { LuLoader } from "react-icons/lu"
import { RemoveScrollBar } from "react-remove-scroll-bar"
import { useEffectOnce } from "react-use"


function uuidv4(): string {
    // @ts-ignore
    return ([1e7] + -1e3 + -4e3 + -8e3 + -1e11).replace(/[018]/g, (c) =>
        (c ^ (crypto.getRandomValues(new Uint8Array(1))[0] & (15 >> (c / 4)))).toString(16),
    )
}

export const websocketConnectedAtom = atom(false)

export function WebsocketProvider({ children }: { children: React.ReactNode }) {
    const [socket, setSocket] = useAtom(websocketAtom)
    const [isConnected, setIsConnected] = useAtom(websocketConnectedAtom)
    const openDrawers = useAtomValue(__openDrawersAtom)

    useEffectOnce(() => {

        function connectWebSocket() {
            const newSocket = new WebSocket(`${document.location.protocol == "https:"
                ? "wss"
                : "ws"}://${getServerBaseUrl(true)}/events?id=${uuidv4()}`)

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
        <>
            {openDrawers.length > 0 && <RemoveScrollBar />}
            <WebSocketContext.Provider value={socket}>
                {!isConnected && <div
                    className="fixed right-4 bottom-4 bg-gray-900 border text-[--muted] text-sm py-3 px-5 font-semibold rounded-md z-[100] flex gap-2 items-center"
                >
                    <LuLoader className="text-brand-200 animate-spin text-lg" />
                    Establishing connection...
                </div>}
                {children}
            </WebSocketContext.Provider>
        </>
    )

}
