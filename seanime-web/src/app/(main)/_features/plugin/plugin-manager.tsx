import { useListExtensionData } from "@/api/hooks/extensions.hooks"
import { WSEvents } from "@/lib/server/ws-events"
import { usePathname, useRouter, useSearchParams } from "next/navigation"
import { startTransition, useEffect, useState } from "react"
import { useWebsocketMessageListener } from "../../_hooks/handle-websockets"
import { PluginCommandPalettes } from "./command/plugin-command-palettes"
import {
    usePluginListenScreenGetCurrentEvent,
    usePluginListenScreenNavigateToEvent,
    usePluginListenScreenReloadEvent,
    usePluginSendScreenChangedEvent,
} from "./generated/plugin-events"
import { PluginHandler } from "./plugin-handler"

export function PluginManager() {
    const router = useRouter()
    const pathname = usePathname()
    const searchParams = useSearchParams()
    const { sendScreenChangedEvent } = usePluginSendScreenChangedEvent()

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
        sendScreenChangedEvent({
            pathname: pathname,
            query: window.location.search,
        })
    }, [pathname, searchParams])

    usePluginListenScreenGetCurrentEvent((event, extensionId) => {
        sendScreenChangedEvent({
            pathname: pathname,
            query: window.location.search,
        }, extensionId)
    }, "") // Listen to all plugins

    usePluginListenScreenNavigateToEvent((event) => {
        if ([
            "/entry", "/anilist", "/search", "/manga",
            "/settings", "/auto-downloader", "/debrid", "/torrent-list",
            "/schedule", "/extensions", "/sync", "/discover",
            "/scan-summaries",

        ].some(path => event.path.startsWith(path))) {
            router.push(event.path)
        }
    }, "") // Listen to all plugins

    usePluginListenScreenReloadEvent((event) => {
        router.refresh()
    }, "") // Listen to all plugins

    return <>
        {/* Render plugin handlers for each extension */}
        {extensions?.filter(e => e.type === "plugin" && !unloadedExtensions.includes(e.id)).map(extension => (
            <PluginHandler key={extension.id} extensionId={extension.id} onUnloaded={() => setUnloadedExtensions(prev => [...prev, extension.id])} />
        ))}
        <PluginCommandPalettes />
    </>
}
