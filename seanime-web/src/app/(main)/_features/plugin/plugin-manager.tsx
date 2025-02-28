import { useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { usePathname, useSearchParams } from "next/navigation"
import { useEffect } from "react"
import { usePluginSendScreenChangedEvent } from "./generated/plugin-events"

export function PluginManager() {
    const pathname = usePathname()
    const searchParams = useSearchParams()
    const { sendScreenChangedEvent } = usePluginSendScreenChangedEvent()
    const { sendPluginMessage } = useWebsocketSender()


    useEffect(() => {
        sendScreenChangedEvent({
            pathname: pathname,
            query: window.location.search,
        })
    }, [pathname, searchParams])

    useEffect(() => {
        // sendPluginMessage("tray:render-all", {})
    }, [])

    return <></>
}
