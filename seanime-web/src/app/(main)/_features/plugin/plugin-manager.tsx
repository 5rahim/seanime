import { useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { usePathname, useRouter, useSearchParams } from "next/navigation"
import { useEffect } from "react"
import { usePluginListenScreenNavigateToEvent, usePluginSendScreenChangedEvent } from "./generated/plugin-events"

export function PluginManager() {
    const router = useRouter()
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

    usePluginListenScreenNavigateToEvent((event) => {
        if ([
            "/entry", "/anilist", "/search", "/manga",
            "/settings", "/auto-downloader", "/debrid", "/torrent-list", "/schedule", "/extensions", "/sync", "/discover",
            "/scan-summaries",

        ].some(path => event.path.startsWith(path))) {
            router.push(event.path)
        }
    }, "") // Listen to all plugins

    return <></>
}
