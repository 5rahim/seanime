import { useListExtensionData } from "@/api/hooks/extensions.hooks"
import { useIsMainTabRef } from "@/app/websocket-provider"
import { useDebounce } from "@/hooks/use-debounce"
import { WSEvents } from "@/lib/server/ws-events"
import { usePathname, useRouter, useSearchParams } from "next/navigation"
import { startTransition, useEffect, useState } from "react"
import { useWindowSize } from "react-use"
import { useWebsocketMessageListener } from "../../_hooks/handle-websockets"
import { PluginCommandPalettes } from "./command/plugin-command-palettes"
import {
    usePluginListenDOMGetViewportSizeEvent,
    usePluginListenScreenGetCurrentEvent,
    usePluginListenScreenNavigateToEvent,
    usePluginListenScreenReloadEvent,
    usePluginSendDOMViewportSizeEvent,
    usePluginSendScreenChangedEvent,
} from "./generated/plugin-events"
import { PluginHandler } from "./plugin-handler"

export function PluginManager() {
    const router = useRouter()
    const pathname = usePathname()
    const searchParams = useSearchParams()
    const { sendScreenChangedEvent } = usePluginSendScreenChangedEvent()
    const isMainTabRef = useIsMainTabRef()

    const { data: extensions } = useListExtensionData()

    const [unloadedExtensions, setUnloadedExtensions] = useState<string[]>([])

    useWebsocketMessageListener({
        type: WSEvents.PLUGIN_LOADED,
        onMessage: (extensionId: string) => {
            startTransition(() => {
                setUnloadedExtensions(prev => prev.filter(id => id !== extensionId))
            })
        },
    })

    useEffect(() => {
        if (!isMainTabRef.current) return
        sendScreenChangedEvent({
            pathname: pathname,
            query: window.location.search,
        })
    }, [pathname, searchParams])

    usePluginListenScreenGetCurrentEvent((event, extensionId) => {
        if (!isMainTabRef.current) return
        sendScreenChangedEvent({
            pathname: pathname,
            query: window.location.search,
        }, extensionId)
    }, "") // Listen to all plugins

    usePluginListenScreenNavigateToEvent((event) => {
        if (!isMainTabRef.current) return
        if ([
            "/entry", "/lists", "/search", "/manga",
            "/settings", "/auto-downloader", "/debrid", "/torrent-list",
            "/schedule", "/extensions", "/sync", "/discover",
            "/scan-summaries", "/webview",

        ].some(path => event.path.startsWith(path))) {
            router.push(event.path)
        }
    }, "") // Listen to all plugins

    usePluginListenScreenReloadEvent((event) => {
        if (!isMainTabRef.current) return
        router.refresh()
    }, "") // Listen to all plugins

    return <>
        {/* Render plugin handlers for each extension */}
        <SizeEvents />
        {extensions?.filter(e => e.type === "plugin" && !unloadedExtensions.includes(e.id)).map(extension => (
            <PluginHandler key={extension.id} extensionId={extension.id} onUnloaded={() => setUnloadedExtensions(prev => [...prev, extension.id])} />
        ))}
        <PluginCommandPalettes />
    </>
}

function SizeEvents() {
    const isMainTabRef = useIsMainTabRef()
    const { sendDOMViewportSizeEvent } = usePluginSendDOMViewportSizeEvent()

    const { width, height } = useWindowSize()
    const debounceWindowSize = useDebounce({ width, height }, 400)

    useEffect(() => {
        if (!isMainTabRef.current) return
        sendDOMViewportSizeEvent({
            width: debounceWindowSize.width,
            height: debounceWindowSize.height,
        })
    }, [debounceWindowSize])

    usePluginListenDOMGetViewportSizeEvent((event, extensionId) => {
        if (!isMainTabRef.current) return
        sendDOMViewportSizeEvent({
            width: debounceWindowSize.width,
            height: debounceWindowSize.height,
        }, extensionId)
    }, "")

    return null
}
