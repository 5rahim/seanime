import { PluginUI_WebviewSlot } from "@/api/generated/types"
import {
    Plugin_Server_WebviewIframeEventPayload,
    Plugin_Server_WebviewSyncStateEventPayload,
    usePluginListenWebviewIframeEvent,
    usePluginListenWebviewSyncStateEvent,
    usePluginSendEventHandlerTriggeredEvent,
    usePluginSendWebviewMountedEvent,
} from "@/app/(main)/_features/plugin/generated/plugin-events"
import { useIsMainTab, useIsMainTabRef } from "@/app/websocket-provider"
import { logger } from "@/lib/helpers/debug"
import { Portal } from "@radix-ui/react-portal"
import { useMap } from "@uidotdev/usehooks"
import React from "react"
import { useMount, useUnmount } from "react-use"

const log = logger("PLUGIN WEBVIEWS")

type PluginWebviewSlotProps = {
    slot: PluginUI_WebviewSlot
}

type IframeWebview = {
    webviewId: string
    extensionId: string
    src: string
    token: string // unique token for message verification
}

function getWebviewId(extensionId: string, slot: PluginUI_WebviewSlot) {
    return `${extensionId}-${slot}`
}

// generate a unique token for each webview instance
function generateWebviewToken() {
    return `${Date.now()}-${Math.random().toString(36).substring(2, 15)}`
}

// language=JavaScript
const generateBridgeScript = (token: string, parentOrigin: string, widgetId: string) => `
    const WEBVIEW_TOKEN = "${token}";
    const PARENT_ORIGIN = "${parentOrigin}";

    window.webview = {
        send: (event, payload) => {
            window.parent.postMessage({
                type: "plugin-webview-trigger",
                event,
                payload,
                webviewId: "${widgetId}",
                token: WEBVIEW_TOKEN
            }, PARENT_ORIGIN)
        },
        on: (event, callback) => {
            window.addEventListener("message", (e) => {
                const isTrustedOrigin = e.origin === PARENT_ORIGIN || e.origin === "null";
                if (!isTrustedOrigin || e.data.token !== WEBVIEW_TOKEN) return;

                if (e.data.type === "plugin-webview-sync" && e.data.key === event) {
                    callback(e.data.value)
                }
            })
        }
    };
`

export const processUserHtml = (userHtml: string, token: string, parentOrigin: string, widgetId: string): string => {
    const bridgeScript = generateBridgeScript(token, parentOrigin, widgetId)
    const scriptTag = `<script>${bridgeScript}</script>`

    // We prioritize injecting into <head> so the API is ready before body scripts run
    if (userHtml.includes("<head>")) {
        return userHtml.replace("<head>", `<head>\n${scriptTag}`)
    }

    // Fallback to body if head doesn't exist
    if (userHtml.includes("<body>")) {
        return userHtml.replace("<body>", `<body>\n${scriptTag}`)
    }

    // If the HTML is malformed or just a fragment, prepend it
    return scriptTag + userHtml
}

// renders webviews at the given slot
export function PluginWebviewSlot({ slot }: PluginWebviewSlotProps) {

    const { sendWebviewMountedEvent } = usePluginSendWebviewMountedEvent()
    const { sendEventHandlerTriggeredEvent } = usePluginSendEventHandlerTriggeredEvent()
    const isMainTab = useIsMainTab()
    const isMainTabRef = useIsMainTabRef()
    const previousMainTab = React.useRef(isMainTabRef.current)

    const iframeWebviews = useMap()

    const mountedRef = React.useRef(false)
    useMount(() => {
        if (!isMainTabRef) return
        log.info("Mounting webview slot", slot)
        sendWebviewMountedEvent({ slot: slot })
        mountedRef.current = true
    })

    // remount the webviews when the main tab changes
    React.useEffect(() => {
        if (!mountedRef.current) return
        if (isMainTab && !previousMainTab.current) {
            log.info("Mounting webview slot because main tab changed", slot)
            sendWebviewMountedEvent({ slot: slot })
        }
        previousMainTab.current = isMainTab
    }, [isMainTab])

    // listen for messages from iframe webviews
    React.useEffect(() => {
        const handleMessage = (event: MessageEvent) => {
            // sandboxed iframes have 'null' origin, which is expected
            // we rely on token validation for security instead
            if (event.origin !== window.location.origin && event.origin !== "null") {
                log.warn("Rejected message from invalid origin", event.origin)
                return
            }

            const data = event.data
            if (data.type !== "plugin-webview-trigger") return

            // find the webview by webviewId and validate token
            const webview = iframeWebviews.get(data.webviewId) as IframeWebview | undefined

            if (!webview) {
                log.warn("Received message from unknown webview", data.webviewId)
                return
            }

            if (data.token !== webview.token) {
                log.warn("Rejected message with invalid token", { webviewId: data.webviewId })
                return
            }

            // forward the event to the server
            log.info("Forwarding webview event to server", {
                extensionId: webview.extensionId,
                event: data.event,
            })

            sendEventHandlerTriggeredEvent({
                handlerName: data.event,
                event: data.payload || {},
            }, webview.extensionId)
        }

        window.addEventListener("message", handleMessage)
        return () => window.removeEventListener("message", handleMessage)
    }, [iframeWebviews, slot, sendEventHandlerTriggeredEvent])

    useUnmount(() => {
        iframeWebviews.clear()
    })

    const setupIframeWebview = React.useCallback((extensionId: string, payload: Plugin_Server_WebviewIframeEventPayload) => {
        log.info("Setting up iframe webview", { extensionId, content: payload.content.slice(0, 100) + "..." })
        if (!isMainTabRef) return

        // If the iframe is already mounted, remove the script and add it back
        if (iframeWebviews.has(extensionId)) {
            document.getElementById(`webview-${extensionId}`)?.remove()
        }

        // generate unique token for this webview instance
        const token = generateWebviewToken()

        const webviewId = getWebviewId(extensionId, slot)

        // construct the HTML document for the iframe
        const srcDoc = processUserHtml(
            payload.content,
            token,
            window.location.origin,
            webviewId,
        )


        iframeWebviews.set(webviewId, { webviewId, extensionId, src: srcDoc, token } satisfies IframeWebview)
    }, [iframeWebviews])

    // Get the iframe event
    usePluginListenWebviewIframeEvent((payload, extensionId) => {
        if (!isMainTabRef) return
        if (payload.slot !== slot) return
        setupIframeWebview(extensionId, payload)
    }, "")

    // Listen for state sync events from the server
    usePluginListenWebviewSyncStateEvent((payload: Plugin_Server_WebviewSyncStateEventPayload, extensionId) => {
        if (!isMainTabRef.current) return

        // find the webview by webviewId
        const webview = iframeWebviews.get(payload.webviewId) as IframeWebview | undefined
        if (!webview) {
            log.warn("Received sync state for unknown webview", payload.webviewId)
            return
        }

        // get the iframe element
        const iframeElement = document.getElementById(`webview-${payload.webviewId}`) as HTMLIFrameElement | null
        if (!iframeElement || !iframeElement.contentWindow) {
            log.warn("Cannot find iframe element for webview", payload.webviewId)
            return
        }

        // send the state update to the iframe
        // log.info("Sending state update to iframe", { webviewId: payload.webviewId, key: payload.key })
        iframeElement.contentWindow.postMessage({
            type: "plugin-webview-sync",
            key: payload.key,
            value: payload.value,
            token: webview.token,
        }, "*")
    }, "")

    // Render the iframe
    if (slot === "fixed") {
        return (
            <>
                <Portal container={document.body} className="plugin-webview-portal">
                    {Array.from(iframeWebviews.values()).map(({ webviewId, src, token }) => (
                        <iframe
                            key={webviewId}
                            id={`webview-${webviewId}`}
                            srcDoc={src}
                            sandbox="allow-scripts" // BLOCKS: allow-same-origin, allow-forms, allow-popups
                            className="size-[20rem] border-none fixed top-0 left-0 z-[100]"
                        />
                    ))}
                </Portal>
            </>
        )
    }
    return null
}
