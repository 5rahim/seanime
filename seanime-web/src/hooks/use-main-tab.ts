import { websocketAtom } from "@/app/(main)/_atoms/websocket.atoms"
import { logger } from "@/lib/helpers/debug"
import { __isElectronDesktop__, __isTauriDesktop__ } from "@/types/constants"
import { useAtomValue } from "jotai"
import { useEffect, useRef, useState } from "react"

const log = logger("TAB")
const CHANNEL_NAME = "main-tab-election"
const TAB_ID = `${Date.now()}-${Math.random().toString(36).slice(2)}`
const IS_DESKTOP = __isElectronDesktop__ || __isTauriDesktop__

export function useMainTab(): boolean {
    const [isMainTab, setIsMainTab] = useState(false)
    const channelRef = useRef<BroadcastChannel | null>(null)
    const socket = useAtomValue(websocketAtom)

    useEffect(() => {
        const channel = new BroadcastChannel(CHANNEL_NAME)
        channelRef.current = channel

        const claimMainTab = () => {
            if (document.visibilityState === "visible") {
                // Claim via BroadcastChannel for other web tabs
                channel.postMessage({ type: "claim", tabId: TAB_ID })
                log.info("Claimed main tab")
                setIsMainTab(true)

                // Also claim via WebSocket so backend broadcasts to all clients (including Electron)
                if (socket?.readyState === WebSocket.OPEN) {
                    socket.send(JSON.stringify({
                        type: "main-tab-claim",
                        payload: { tabId: TAB_ID, isDesktop: IS_DESKTOP },
                    }))
                }
            }
        }

        // web tabs -> web tabs
        const handleBroadcastMessage = (event: MessageEvent) => {
            if (event.data.type === "claim" && event.data.tabId !== TAB_ID) {
                // Desktop is isolated so it shouldn't receive claims from the broadcast channel
                // tldr this shouldn't happen but just in case, ignore claim
                if (IS_DESKTOP) {
                    return
                }
                // Another tab claimed main, we yield
                setIsMainTab(false)
                log.warn("Yielded")
            }
        }

        const handleWebSocketMessage = (event: MessageEvent) => {
            try {
                // web tabs -> desktop tab || desktop tab -> web tabs
                const data = JSON.parse(event.data) as { type: string; payload?: { tabId: string; isDesktop: boolean } }
                if (
                    data.type === "main-tab-claim"
                    && data.payload?.tabId !== TAB_ID
                    && (IS_DESKTOP || data.payload?.isDesktop) // Yield only if we're the desktop tab or the desktop tab claimed main
                    && !(IS_DESKTOP && data.payload?.isDesktop) // Don't yield if we're desktop and another desktop tab claimed main
                ) {
                    setIsMainTab(false)
                    log.warn("Yielded")
                }
            }
            catch (e) {
                // Ignore parsing errors
            }
        }

        channel.addEventListener("message", handleBroadcastMessage)
        document.addEventListener("visibilitychange", claimMainTab)
        window.addEventListener("focus", claimMainTab)

        if (socket) {
            socket.addEventListener("message", handleWebSocketMessage)
        }

        // Claim on mount if visible
        claimMainTab()

        return () => {
            channel.removeEventListener("message", handleBroadcastMessage)
            document.removeEventListener("visibilitychange", claimMainTab)
            window.removeEventListener("focus", claimMainTab)
            if (socket) {
                socket.removeEventListener("message", handleWebSocketMessage)
            }
            channel.close()
        }
    }, [socket])

    return isMainTab
}

// export function useMainTab() {
//     const [isLeader, setIsLeader] = useState(false)
//
//     useEffect(() => {
//         const controller = new AbortController()
//         navigator.locks.request(
//             "main_tab",
//             {
//                 signal: controller.signal,
//                 // steal: true,
//             },
//             async () => {
//                 // If this callback runs, we've acquired the lock
//                 setIsLeader(true)
//
//                 // return a promise that never resolves to hold the lock
//                 // until the component unmounts (signal aborts)
//                 return new Promise<void>((resolve) => {
//                     controller.signal.onabort = () => resolve()
//                 })
//             },
//         ).catch()
//
//         return () => {
//             controller.abort()
//             setIsLeader(false)
//         }
//     }, [])
//
//     console.warn(isLeader)
//
//     return isLeader
// }
