import { getServerBaseUrl } from "@/api/client/server-url"
import { websocketAtom, WebSocketContext } from "@/app/(main)/_atoms/websocket.atoms"
import { TauriRestartServerPrompt } from "@/app/(main)/_tauri/tauri-restart-server-prompt"
import { __openDrawersAtom } from "@/components/ui/drawer"
import { logger } from "@/lib/helpers/debug"
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

    // Added heartbeatRef to periodically check connection health
    const heartbeatRef = React.useRef<any>(null)

    React.useEffect(() => {
        logger("WebsocketProvider").info("Seanime-Client-Id", cookies["Seanime-Client-Id"])
        if (cookies["Seanime-Client-Id"]) {
            setClientId(cookies["Seanime-Client-Id"])
        }
    }, [cookies])

    useEffectOnce(() => {
        function connectWebSocket() {
            const wsUrl = `${document.location.protocol == "https:" ? "wss" : "ws"}://${getServerBaseUrl(true)}/events`
            const clientId = cookies["Seanime-Client-Id"] || uuidv4()

            const newSocket = new WebSocket(`${wsUrl}?id=${clientId}`)

            newSocket.addEventListener("open", () => {
                logger("WebsocketProvider").info("WebSocket connection opened")
                setIsConnected(true)
                setConnectionErrorCount(0)

                // Set cookie if it doesn't exist
                if (!cookies["Seanime-Client-Id"]) {
                    setCookie("Seanime-Client-Id", clientId, {
                        path: "/",
                        sameSite: "lax",
                        secure: false,
                        maxAge: 24 * 60 * 60, // 24 hours
                    })
                }

                // Start heartbeat interval to detect silent disconnections
                heartbeatRef.current = setInterval(() => {
                    if (newSocket.readyState !== WebSocket.OPEN) {
                        logger("WebsocketProvider").error("Heartbeat check failed, closing connection")
                        newSocket.close()
                    }
                }, 30000) // check every 30 seconds
            })

            newSocket.addEventListener("close", () => {
                // Clear heartbeat interval
                if (heartbeatRef.current) {
                    clearInterval(heartbeatRef.current)
                    heartbeatRef.current = null
                }
                logger("WebsocketProvider").info("WebSocket connection closed")
                setIsConnected(false)
                // Reconnect after a delay
                setConnectionErrorCount(count => count + 1)
                setTimeout(connectWebSocket, 1000)
            })

            newSocket.addEventListener("error", (event) => {
                logger("WebsocketProvider").error("WebSocket encountered an error:", event)
                newSocket.close()
            })

            setSocket(newSocket)

            return newSocket
        }

        if (!socket || socket.readyState === WebSocket.CLOSED) {
            // If the socket is not set or the connection is closed, initiate a new connection
            connectWebSocket()
        }

        return () => {
            if (socket) {
                socket.close()
                setIsConnected(false)
            }
            // Cleanup heartbeat on unmount
            if (heartbeatRef.current) {
                clearInterval(heartbeatRef.current)
                heartbeatRef.current = null
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
