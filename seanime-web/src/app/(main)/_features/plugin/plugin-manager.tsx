import { useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { usePathname } from "next/navigation"
import { useEffect } from "react"
import { usePluginSendScreenChangedEvent } from "./generated/plugin-events"

export function PluginManager() {
    const pathname = usePathname()
    const { sendScreenChangedEvent } = usePluginSendScreenChangedEvent()
    const { sendPluginMessage } = useWebsocketSender()


    useEffect(() => {
        sendScreenChangedEvent({
            pathname: pathname,
            query: window.location.search,
        })
    }, [pathname])

    useEffect(() => {
        // sendPluginMessage("tray:render-all", {})
    }, [])

    return <></>
}
