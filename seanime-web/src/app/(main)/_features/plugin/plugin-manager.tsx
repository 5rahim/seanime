import { useListExtensionData } from "@/api/hooks/extensions.hooks"
import { usePathname, useRouter, useSearchParams } from "next/navigation"
import { useEffect } from "react"
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
        {extensions?.filter(e => e.type === "plugin").map(extension => (
            <PluginHandler key={extension.id} extensionId={extension.id} />
        ))}
        <PluginCommandPalettes />
    </>
}
