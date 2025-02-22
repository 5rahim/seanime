import { useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { usePathname } from "next/navigation"
import { useEffect } from "react"

export function PluginManager() {
    const { sendMessage } = useWebsocketSender()

    const pathname = usePathname()

    useEffect(() => {
        sendMessage({
            type: "plugin",
            payload: {
                type: "screenChanged",
                payload: {
                    pathname: pathname,
                    query: window.location.search,
                },
            },
        })
    }, [pathname])

    return null
}
