import { useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { usePathname } from "next/navigation"
import { useEffect } from "react"

export function PluginManager() {
    const { sendPluginMessage } = useWebsocketSender()

    const pathname = usePathname()

    useEffect(() => {
        sendPluginMessage("screen:changed", {
            pathname: pathname,
            query: window.location.search,
        })
    }, [pathname])

    useEffect(() => {
        // sendPluginMessage("tray:render-all", {})
    }, [])

    return null
}
