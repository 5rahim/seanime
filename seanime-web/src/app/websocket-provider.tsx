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

    // Refs to manage connection state
    const heartbeatRef = React.useRef<NodeJS.Timeout | null>(null)
    const pingIntervalRef = React.useRef<NodeJS.Timeout | null>(null)
    const reconnectTimeoutRef = React.useRef<NodeJS.Timeout | null>(null)
    const lastPongRef = React.useRef<number>(Date.now())
    const socketRef = React.useRef<WebSocket | null>(null)
    const wasDisconnected = React.useRef<boolean>(false)
    const initialConnection = React.useRef<boolean>(true)

    React.useEffect(() => {
        logger("WebsocketProvider").info("Seanime-Client-Id", cookies["Seanime-Client-Id"])
        if (cookies["Seanime-Client-Id"]) {
            setClientId(cookies["Seanime-Client-Id"])
        }
    }, [cookies])

    // Effect to handle page reload on reconnection
    /* React.useEffect(() => {
     // If we're connected now and were previously disconnected (not the first connection)
     if (isConnected && wasDisconnected.current && !initialConnection.current) {
     logger("WebsocketProvider").info("Connection re-established, reloading page")
     // Add a small delay to allow for other components to process the connection status
     setTimeout(() => {
     window.location.reload()
     }, 100)
     }

     // Update the wasDisconnected ref when connection status changes
     if (!isConnected && !initialConnection.current) {
     wasDisconnected.current = true
     }

     // After first connection, set initialConnection to false
     if (isConnected && initialConnection.current) {
     initialConnection.current = false
     }
     }, [isConnected]) */

    useEffectOnce(() => {
        function clearAllIntervals() {
            if (heartbeatRef.current) {
                clearInterval(heartbeatRef.current)
                heartbeatRef.current = null
            }
            if (pingIntervalRef.current) {
                clearInterval(pingIntervalRef.current)
                pingIntervalRef.current = null
            }
            if (reconnectTimeoutRef.current) {
                clearTimeout(reconnectTimeoutRef.current)
                reconnectTimeoutRef.current = null
            }
        }

        function connectWebSocket() {
            // Clear existing connection attempts
            clearAllIntervals()

            // Close any existing socket
            if (socketRef.current && socketRef.current.readyState !== WebSocket.CLOSED) {
                try {
                    socketRef.current.close()
                }
                catch (e) {
                    // Ignore errors on closing
                }
            }

            const wsUrl = `${document.location.protocol == "https:" ? "wss" : "ws"}://${getServerBaseUrl(true)}/events`
            const clientId = cookies["Seanime-Client-Id"] || uuidv4()

            try {
                socketRef.current = new WebSocket(`${wsUrl}?id=${clientId}`)

                // Reset the last pong timestamp whenever we connect
                lastPongRef.current = Date.now()

                socketRef.current.addEventListener("open", () => {
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
                        const timeSinceLastPong = Date.now() - lastPongRef.current

                        // If no pong received for 45 seconds (3 missed pings), consider connection dead
                        if (timeSinceLastPong > 45000) {
                            logger("WebsocketProvider").error(`No pong response for ${Math.round(timeSinceLastPong / 1000)}s, reconnecting`)
                            reconnectSocket()
                            return
                        }

                        if (socketRef.current?.readyState !== WebSocket.OPEN) {
                            logger("WebsocketProvider").error("Heartbeat check failed, reconnecting")
                            reconnectSocket()
                        }
                    }, 15000) // check every 15 seconds

                    // Implement a ping mechanism to keep the connection alive
                    // Start the ping interval slightly offset from the heartbeat to avoid race conditions
                    setTimeout(() => {
                        pingIntervalRef.current = setInterval(() => {
                            if (socketRef.current?.readyState === WebSocket.OPEN) {
                                try {
                                    const timestamp = Date.now()
                                    // Send a ping message to keep the connection alive
                                    socketRef.current?.send(JSON.stringify({
                                        type: "ping",
                                        payload: { timestamp },
                                        clientId: clientId,
                                    }))
                                }
                                catch (e) {
                                    logger("WebsocketProvider").error("Failed to send ping", e)
                                    reconnectSocket()
                                }
                            } else {
                                logger("WebsocketProvider").error("Failed to send ping, WebSocket not open", socketRef.current?.readyState)
                                // reconnectSocket()
                            }
                        }, 15000) // ping every 15 seconds
                    }, 5000) // Start ping interval 5 seconds after heartbeat to offset them
                })

                // Add message handler for pong responses
                socketRef.current?.addEventListener("message", (event) => {
                    try {
                        const data = JSON.parse(event.data) as { type: string; payload?: any }
                        if (data.type === "pong") {
                            // Update the last pong timestamp
                            lastPongRef.current = Date.now()
                            // For debugging purposes
                            // logger("WebsocketProvider").info("Pong received, timestamp updated", lastPongRef.current)
                        }
                    }
                    catch (e) {
                    }
                })

                socketRef.current?.addEventListener("close", (event) => {
                    logger("WebsocketProvider").info(`WebSocket connection closed: ${event.code} ${event.reason}`)
                    handleDisconnection()
                })

                socketRef.current?.addEventListener("error", (event) => {
                    logger("WebsocketProvider").error("WebSocket encountered an error:", event)
                    reconnectSocket()
                })

                setSocket(socketRef.current)
            }
            catch (e) {
                logger("WebsocketProvider").error("Failed to create WebSocket connection:", e)
                handleDisconnection()
            }
        }

        function handleDisconnection() {
            clearAllIntervals()
            setIsConnected(false)
            scheduleReconnect()
        }

        function reconnectSocket() {
            if (socketRef.current) {
                try {
                    socketRef.current.close()
                }
                catch (e) {
                    // Ignore errors on closing
                }
            }
            handleDisconnection()
        }

        function scheduleReconnect() {
            // Reconnect after a delay with exponential backoff
            setConnectionErrorCount(count => {
                const newCount = count + 1
                // Calculate backoff time (1s, 2s, max 3s)
                const backoffTime = Math.min(Math.pow(2, Math.min(newCount - 1, 10)) * 1000, 3000)

                logger("WebsocketProvider").info(`Reconnecting in ${backoffTime}ms (attempt ${newCount})`)

                reconnectTimeoutRef.current = setTimeout(() => {
                    connectWebSocket()
                }, backoffTime)

                return newCount
            })
        }

        if (!socket || socket.readyState === WebSocket.CLOSED) {
            // If the socket is not set or the connection is closed, initiate a new connection
            connectWebSocket()
        }

        return () => {
            if (socketRef.current) {
                try {
                    socketRef.current.close()
                }
                catch (e) {
                    // Ignore errors on closing
                }
            }
            setIsConnected(false)
            // Cleanup all intervals on unmount
            clearAllIntervals()
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

