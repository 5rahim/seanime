import {
    Plugin_Server_WebviewSidebarEventPayload,
    usePluginListenWebviewSidebarEvent,
    usePluginSendWebviewSidebarMountedEvent,
} from "@/app/(main)/_features/plugin/generated/plugin-events"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { useIsMainTab, useIsMainTabRef } from "@/app/websocket-provider"
import { VerticalMenuItem } from "@/components/ui/vertical-menu"
import { logger } from "@/lib/helpers/debug"
import { WSEvents } from "@/lib/server/ws-events"
import { useMap } from "@uidotdev/usehooks"
import { usePathname, useSearchParams } from "next/navigation"
import React, { useMemo } from "react"
import { useMount } from "react-use"

const log = logger("PLUGIN WEBVIEW SIDEBAR ITEMS")

// A plugin can have multiple webviews but only one sidebar item

type WebviewSidebarItem = {
    extensionId: string
    label: string
    icon: string
}

export function usePluginSidebarItems(): VerticalMenuItem[] {
    const { sendWebviewSidebarMountedEvent } = usePluginSendWebviewSidebarMountedEvent()
    const pathname = usePathname()
    const searchParams = useSearchParams()

    const isMainTab = useIsMainTab()
    const isMainTabRef = useIsMainTabRef()
    const previousMainTab = React.useRef(isMainTabRef.current)

    const items = useMap() as Map<string, WebviewSidebarItem>

    const itemsRef = React.useRef(items)

    React.useEffect(() => { itemsRef.current = items }, [items])

    useMount(() => {
        sendWebviewSidebarMountedEvent({})
    })

    const setupItem = React.useCallback((extensionId: string, payload: Plugin_Server_WebviewSidebarEventPayload) => {
        log.info("Setting up webview sidebar item", { extensionId })
        if (!isMainTabRef) return

        // Remove old iframe
        if (items.has(extensionId)) {
            items.delete(extensionId)
        }

        items.set(extensionId, {
            extensionId,
            label: payload.label,
            icon: payload.icon,
        } satisfies WebviewSidebarItem)
    }, [items])

    // Get the iframe event
    usePluginListenWebviewSidebarEvent((payload, extensionId) => {
        if (!isMainTabRef) return
        setupItem(extensionId, payload)
    }, "")

    // Remove specific item ts plugin is unloaded
    useWebsocketMessageListener({
        type: WSEvents.PLUGIN_UNLOADED,
        onMessage: (extensionId) => {
            items.forEach((webview) => {
                if (webview.extensionId === extensionId) {
                    items.delete(extensionId)
                }
            })
        },
    })

    const sortedItems = useMemo(() => {
        return Array.from(items.values()).sort((a, b) => a.extensionId.localeCompare(b.extensionId))
    }, [items.values()])

    return sortedItems.map((item) => {
        let svgContent = item.icon.trim()
        if (!svgContent.startsWith("<svg")) {
            // handle cases where the string includes the data URI prefix
            const cleanBase64 = item.icon?.replace(/^data:image\/svg\+xml;base64,/, "")
            // decode the base64 to raw svg xml
            try {
                svgContent = atob(cleanBase64)
            }
            catch (e) {
            }
        }
        return {
            name: item.label,
            href: `/webview?id=${item.extensionId}`,
            isCurrent: pathname === `/webview` && searchParams.get("id") === item.extensionId,
            iconType: (props: React.HTMLAttributes<HTMLSpanElement>) => (
                <span
                    {...props}
                    // display contents removes the span from the layout box tree
                    // allowing the svg to act as a direct child of the parent
                    style={{ display: "contents", ...props.style }}
                    dangerouslySetInnerHTML={{ __html: svgContent }}
                />
            ),
        } as VerticalMenuItem
    })
}
