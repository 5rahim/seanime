import { getServerBaseUrl } from "@/api/client/server-url"
import { websocketAtom, WebSocketContext } from "@/app/(main)/_atoms/websocket.atoms"
import { TauriRestartServerPrompt } from "@/app/(main)/_tauri/tauri-restart-server-prompt"
import { __openDrawersAtom } from "@/components/ui/drawer"
import { atom, useAtomValue } from "jotai"
import { useAtom, useSetAtom } from "jotai/react"
import React from "react"
import { useCookies } from "react-cookie"
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
export const websocketConnectionErrorCountAtom = atom(0)

export const clientIdAtom = atom<string | null>(null)

export function WebsocketProvider({ children }: { children: React.ReactNode }) {
    const [socket, setSocket] = useAtom(websocketAtom)
    const [isConnected, setIsConnected] = useAtom(websocketConnectedAtom)
    const setConnectionErrorCount = useSetAtom(websocketConnectionErrorCountAtom)
    const openDrawers = useAtomValue(__openDrawersAtom)

    const [cookies, setCookie, removeCookie] = useCookies(["Seanime-Client-Id"])

    const [, setClientId] = useAtom(clientIdAtom)
    React.useEffect(() => {
        if (cookies["Seanime-Client-Id"]) {
            setClientId(cookies["Seanime-Client-Id"])
        }
    }, [cookies])


    useEffectOnce(() => {
        function connectWebSocket() {
            const newSocket = new WebSocket(`${document.location.protocol == "https:"
                ? "wss"
                : "ws"}://${getServerBaseUrl(true)}/events?id=${cookies["Seanime-Client-Id"] || uuidv4()}`)

            newSocket.addEventListener("open", () => {
                console.log("WebSocket connection opened")
                setIsConnected(true)
                setConnectionErrorCount(0)
            })

            newSocket.addEventListener("close", () => {
                console.log("WebSocket connection closed")
                setIsConnected(false)
                // Reconnect after a delay
                setConnectionErrorCount(count => count + 1)
                setTimeout(connectWebSocket, 1000)
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
            {process.env.NEXT_PUBLIC_PLATFORM === "desktop" && (
                <TauriRestartServerPrompt />
            )}
            <WebSocketContext.Provider value={socket}>
                {!isConnected && <div
                    className="fixed right-4 bottom-4 bg-gray-900 border text-[--muted] text-sm py-3 px-5 font-semibold rounded-[--radius-md] z-[100] flex gap-2 items-center"
                >
                    <LuLoader className="text-brand-200 animate-spin text-lg" />
                    Establishing connection...
                </div>}
                {children}
            </WebSocketContext.Provider>
        </>
    )

}
