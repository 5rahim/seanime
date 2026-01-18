import { PluginUI_WebviewOptions, PluginUI_WebviewSlot } from "@/api/generated/types"
import {
    Plugin_Server_WebviewIframeEventPayload,
    Plugin_Server_WebviewSyncStateEventPayload,
    usePluginListenWebviewCloseEvent,
    usePluginListenWebviewIframeEvent,
    usePluginListenWebviewSyncStateEvent,
    usePluginSendWebviewLoadedEvent,
    usePluginSendWebviewMountedEvent,
    usePluginSendWebviewPostMessageEvent,
    usePluginSendWebviewUnmountedEvent,
} from "@/app/(main)/_features/plugin/generated/plugin-events"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { useIsMainTab, useIsMainTabRef } from "@/app/websocket-provider"
import { cn } from "@/components/ui/core/styling"
import { useMeasureElement } from "@/hooks/use-measure-element"
import { logger } from "@/lib/helpers/debug"
import { WSEvents } from "@/lib/server/ws-events"
import { Portal } from "@radix-ui/react-portal"
import { useMap } from "@uidotdev/usehooks"
import { usePathname, useSearchParams } from "next/navigation"
import React from "react"
import { useMount, useUnmount } from "react-use"

const log = logger("PLUGIN WEBVIEWS")

type PluginWebviewSlotProps = {
    slot: PluginUI_WebviewSlot
}

type Position = "top-left" | "top-right" | "bottom-left" | "bottom-right"

type IframeWebview = {
    webviewId: string
    extensionId: string
    src: string
    slot: PluginUI_WebviewSlot
    token: string // unique token for message verification
    options?: PluginUI_WebviewOptions
    position?: { x: number; y: number }
    defaultPosition?: Position
    preservedSize?: { width: string; height: string }
}

type MessageFromWebview = {
    webviewId: string
    type: WebviewMessageType
    event: string
    token: string
    payload?: any
}


const enum WebviewMessageType {
    SyncState = "plugin-webview-sync",
    Trigger = "plugin-webview-trigger",
    Resize = "plugin-webview-resize",
    RequestContainerWidth = "plugin-webview-request-container-width",
    ContainerWidth = "plugin-webview-container-width",
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
    const WEBVIEW_TOKEN = "${token}"
    const PARENT_ORIGIN = "${parentOrigin}"

    window.webview = {
        // Send a message to Seanime
        send: (event, payload) => {
            window.parent.postMessage({
                type: "${WebviewMessageType.Trigger}",
                event,
                payload,
                webviewId: "${widgetId}",
                token: WEBVIEW_TOKEN
            }, PARENT_ORIGIN)
        },
        // Receive messages from Seanime
        on: (event, callback) => {
            const handler = (e) => {
                const isTrustedOrigin = e.origin === PARENT_ORIGIN || e.origin === "null"
                if (!isTrustedOrigin || e.data.token !== WEBVIEW_TOKEN) return
                // State syncing
                if (e.data.type === "${WebviewMessageType.SyncState}" && e.data.key === event) {
                    callback(e.data.value)
                }
                // Handle response from requestContainerWidth()
                if (e.data.type === "${WebviewMessageType.ContainerWidth}") {
                    if (event === "containerWidth") {
                        callback(e.data.width)
                    }
                }
            }
            window.addEventListener("message", handler)
            // Return cancel function
            return () => {
                window.removeEventListener("message", handler)
            }
        },
        // Get the width of the container the webview is in.
        requestContainerWidth: () => {
            window.parent.postMessage({
                type: "${WebviewMessageType.RequestContainerWidth}",
                webviewId: "${widgetId}",
                token: WEBVIEW_TOKEN
            }, PARENT_ORIGIN)
        },
        // Notify Seanime when the webview body changed size
        _onResizeObserved: () => {
            const height = document.body.scrollHeight
            const width = document.body.scrollWidth
            window.parent.postMessage({
                type: "${WebviewMessageType.Resize}",
                webviewId: "${widgetId}",
                width: width,
                height: height,
                token: WEBVIEW_TOKEN
            }, PARENT_ORIGIN)
        }
    }

    // Notify Seanime to resize iframe when content changes
    if (window.ResizeObserver) {
        window.addEventListener("load", () => {
            const resizeObserver = new ResizeObserver(() => {
                if (window.webview) window.webview._onResizeObserved()
            })
            if (document.body) {
                resizeObserver.observe(document.body)
            }
        })
    }
`

export const processUserHtml = (userHtml: string, token: string, parentOrigin: string, widgetId: string): string => {
    // minify
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
    const pathname = usePathname()
    const searchParams = useSearchParams()

    const { sendWebviewMountedEvent } = usePluginSendWebviewMountedEvent()
    const { sendWebviewUnmountedEvent } = usePluginSendWebviewUnmountedEvent()
    const { sendWebviewPostMessageEvent } = usePluginSendWebviewPostMessageEvent()
    const isMainTab = useIsMainTab()
    const isMainTabRef = useIsMainTabRef()
    const previousMainTab = React.useRef(isMainTabRef.current)

    const iframeWebviews = useMap() as Map<string, IframeWebview>

    const iframeWebviewsRef = React.useRef(iframeWebviews)

    React.useEffect(() => { iframeWebviewsRef.current = iframeWebviews }, [iframeWebviews])

    const mountedRef = React.useRef(false)
    useMount(() => {
        if (!isMainTabRef) return
        log.info("Mounting webview slot", slot)
        sendWebviewMountedEvent({ slot: slot })
        mountedRef.current = true
    })
    useUnmount(() => {
        if (!isMainTabRef) return
        log.info("Unmounting webview slot", slot)
        sendWebviewUnmountedEvent({ slot: slot })
        mountedRef.current = false
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

    function getWebviewIframeElement(webviewId: string): HTMLIFrameElement | undefined {
        return document.getElementById(`webview-${webviewId}`) as HTMLIFrameElement | undefined
    }

    // listen for messages from iframe webviews
    React.useEffect(() => {
        const handleMessage = (event: MessageEvent) => {
            // sandboxed iframes have 'null' origin, which is expected
            // we rely on token validation for security instead
            if (event.origin !== window.location.origin && event.origin !== "null") {
                log.warn("Rejected message from invalid origin", event.origin)
                return
            }

            const data = event.data as MessageFromWebview & any

            const currentMap = iframeWebviewsRef.current

            // find the webview by webviewId and validate token
            const webview = currentMap.get(data.webviewId) as IframeWebview | undefined

            if (!webview) {
                // log.warn("Received message from unknown webview", data.webviewId)
                return
            }

            if (data.token !== webview.token) {
                log.warn("Rejected message with invalid token", { webviewId: data.webviewId })
                return
            }

            // Handle resize notifications
            if (data.type === WebviewMessageType.Resize) {
                if (webview.options?.autoHeight) {
                    const iframe = getWebviewIframeElement(data.webviewId)
                    if (iframe) {
                        iframe.style.height = `${data.height}px`
                        if (!webview.options?.fullWidth) {
                            iframe.style.width = `${data.width}px`
                        }
                    }
                }
                return
            }

            // Handle container width requests
            if (data.type === WebviewMessageType.RequestContainerWidth) {
                const iframe = getWebviewIframeElement(data.webviewId)
                if (iframe && iframe.contentWindow) {
                    const container = iframe.parentElement
                    const containerWidth = container?.clientWidth || window.innerWidth
                    iframe.contentWindow.postMessage({
                        type: WebviewMessageType.ContainerWidth,
                        width: containerWidth,
                        token: webview.token,
                    }, "*")
                }
                return
            }

            if (data.type !== WebviewMessageType.Trigger) return

            // forward the event to the server
            log.info("Forwarding webview event to server", {
                extensionId: webview.extensionId,
                event: data.event,
            })

            sendWebviewPostMessageEvent({
                slot: webview.slot,
                eventName: data.event,
                event: data.payload || {},
            }, webview.extensionId)
        }

        window.addEventListener("message", handleMessage)
        return () => window.removeEventListener("message", handleMessage)
    }, [slot])

    useUnmount(() => {
        iframeWebviews.clear()
    })

    useWebsocketMessageListener({
        type: WSEvents.PLUGIN_UNLOADED,
        onMessage: (extensionId) => {
            if (!isMainTabRef) return
            iframeWebviews.forEach((webview, webviewId) => {
                if (webview.extensionId === extensionId) {
                    iframeWebviews.delete(webviewId)
                }
            })
            // setTimeout(() => {
            //     sendWebviewMountedEvent({ slot: slot })
            // }, 500)
        },
    })

    const handleUpdatePosition = React.useCallback((id: string, x: number, y: number) => {
        const wv = iframeWebviews.get(id)
        if (wv) iframeWebviews.set(id, { ...wv, position: { x, y } })
    }, [iframeWebviews])

    const handleUpdateSize = React.useCallback((id: string, width: string, height: string) => {
        const wv = iframeWebviews.get(id)
        if (wv) iframeWebviews.set(id, { ...wv, preservedSize: { width, height } })
    }, [iframeWebviews])

    const handleClose = React.useCallback((id: string) => {
        iframeWebviews.delete(id)
        document.getElementById(`webview-${id}`)?.remove()
    }, [iframeWebviews])

    const setupIframeWebview = React.useCallback((extensionId: string, payload: Plugin_Server_WebviewIframeEventPayload) => {
        log.info("Setting up iframe webview", { extensionId })
        if (!isMainTabRef) return

        const webviewId = getWebviewId(extensionId, slot)

        // If the iframe is already mounted, preserve position/size if draggable/resizable
        const existingWebview = iframeWebviews.get(webviewId) as IframeWebview | undefined
        const preservedPosition = existingWebview?.position
        const preservedSize = existingWebview?.preservedSize

        // Remove old iframe
        if (iframeWebviews.has(webviewId)) {
            iframeWebviews.delete(webviewId)
        }

        // generate unique token for this webview instance
        const token = generateWebviewToken()

        // construct the HTML document for the iframe
        const srcDoc = processUserHtml(
            payload.content,
            token,
            window.location.origin,
            webviewId,
        )

        const options = payload.options as PluginUI_WebviewOptions | undefined

        iframeWebviews.set(webviewId, {
            webviewId,
            extensionId,
            src: srcDoc,
            token,
            options,
            position: preservedPosition,
            preservedSize: preservedSize,
            slot,
        } satisfies IframeWebview)
    }, [iframeWebviews])

    // Get the iframe event
    usePluginListenWebviewIframeEvent((payload, extensionId) => {
        if (!isMainTabRef) return
        if (payload.slot !== slot) return
        if (payload.slot === "screen") {
            if (pathname !== "/webview") return
            if (extensionId !== searchParams.get("id")) return
        }
        setupIframeWebview(extensionId, payload)
    }, "")

    // Listen for state sync events from the server
    // i.e., when
    // - webview.channel.sync("count", count)
    // - webview.channel.send("foo", $store.get("foo"))
    usePluginListenWebviewSyncStateEvent((payload: Plugin_Server_WebviewSyncStateEventPayload, extensionId) => {
        if (!isMainTabRef.current) return

        // find the webview by webviewId
        const webview = iframeWebviews.get(payload.webviewId) as IframeWebview | undefined
        if (!webview) {
            return
        }

        // get the iframe element
        const iframeElement = getWebviewIframeElement(webview.webviewId)
        if (!iframeElement || !iframeElement.contentWindow) {
            log.warn("Cannot find iframe element for webview", payload.webviewId)
            return
        }

        // send the state update to the iframe
        // log.info("Sending state update to iframe", { webviewId: payload.webviewId, key: payload.key })
        iframeElement.contentWindow.postMessage({
            type: WebviewMessageType.SyncState,
            key: payload.key,
            value: payload.value,
            token: webview.token,
        }, "*")
    }, "")

    usePluginListenWebviewCloseEvent((payload, extensionId) => {
        if (!isMainTabRef) return
        const webview = iframeWebviews.get(payload.webviewId) as IframeWebview | undefined
        if (webview) {
            iframeWebviews.delete(webview.webviewId)
            document.getElementById(`webview-${webview.webviewId}`)?.remove()
        }
    }, "")

    // Render the iframe
    if (slot === "fixed") {
        return (
            <>
                <Portal container={document.body} className="plugin-webview-portal">
                    {Array.from(iframeWebviews.values()).map((webview) => (
                        <WebviewIframe
                            key={webview.webviewId}
                            webview={webview}
                            onUpdatePosition={handleUpdatePosition}
                            onClose={handleClose}
                        />
                    ))}
                </Portal>
            </>
        )
    }
    return <>
        {Array.from(iframeWebviews.values()).map((webview) => (
            <WebviewIframe
                key={webview.webviewId}
                webview={webview}
                onUpdatePosition={handleUpdatePosition}
                onClose={handleClose}
            />
        ))}
    </>
}

type WebviewIframeProps = {
    webview: IframeWebview
    onUpdatePosition: (id: string, x: number, y: number) => void
    onClose: (id: string) => void
}

function WebviewIframe({ webview, onUpdatePosition, onClose }: WebviewIframeProps) {
    const { sendWebviewLoadedEvent } = usePluginSendWebviewLoadedEvent()
    const options = webview.options || {}

    const iframeRef = React.useRef<HTMLIFrameElement>(null)
    const [isDragging, setIsDragging] = React.useState(false)
    const dragStartPos = React.useRef({ x: 0, y: 0, elemX: 0, elemY: 0 })

    const viewportPaddingRef = React.useRef({ paddingTop: 0, paddingRight: 0, paddingBottom: 0, paddingLeft: 80 })

    function getXPosition(x: number) {
        return Math.max(viewportPaddingRef.current.paddingLeft,
            Math.min(x, window.innerWidth - (iframeRef.current?.offsetWidth || 0) - viewportPaddingRef.current.paddingRight))
    }

    function getYPosition(y: number) {
        return Math.max(viewportPaddingRef.current.paddingTop,
            Math.min(y, window.innerHeight - (iframeRef.current?.offsetHeight || 0) - viewportPaddingRef.current.paddingBottom))
    }

    const defaultPosition = React.useMemo(() => {
        let defaultX = getXPosition(options?.window?.defaultX || 0)
        let defaultY = getYPosition(options?.window?.defaultY || 0)
        if (options?.window?.defaultPosition && window) {
            const padding = 10
            const vpWidth = window.innerWidth
            const vpHeight = window.innerHeight
            switch (options.window.defaultPosition) {
                case "top-left":
                    defaultX = padding
                    defaultY = padding
                    break
                case "top-right":
                    defaultX = vpWidth - (iframeRef.current?.offsetWidth || 0) - padding
                    defaultY = padding
                    break
                case "bottom-left":
                    defaultX = padding
                    defaultY = vpHeight - (iframeRef.current?.offsetHeight || 0) - padding
                    break
                case "bottom-right":
                    defaultX = vpWidth - (iframeRef.current?.offsetWidth || 0) - padding
                    defaultY = vpHeight - (iframeRef.current?.offsetHeight || 0) - padding
                    break
            }
        }
        return { x: defaultX, y: defaultY }
    }, [iframeRef.current])


    const position = webview.position || defaultPosition
    // const size = webview.preservedSize

    const customStyle = React.useMemo(() => {
        return options.style?.split(";")?.reduce((acc, rule) => {
            const [key, value] = rule.split(":").map(s => s.trim())
            if (key && value) {
                // Convert kebab-case to camelCase
                const camelKey = key.replace(/-([a-z])/g, g => g[1].toUpperCase())
                ;(acc as any)[camelKey] = value
            }
            return acc
        }, {} as React.CSSProperties)
    }, [options.style])

    // Build inline styles
    const buildStyles = React.useMemo((): React.CSSProperties => {
        const styles: React.CSSProperties = {
            position: webview.slot === "fixed" ? "fixed" : "relative",
            border: "none",
            zIndex: options.zIndex || (webview.slot === "fixed" ? 100 : 5),
            background: "transparent",
        }

        if (webview.slot === "fixed" && !webview.options?.window?.frameless) {
            styles.overflow = "hidden"
            styles.borderRadius = "1rem"
            styles.borderColor = "var(--border)"
            styles.borderStyle = "solid"
            styles.borderWidth = "1px"
            styles.backgroundColor = "var(--background)"
            styles.boxShadow = "var(--tw-ring-offset-shadow, 0 0 #0000), var(--tw-ring-shadow, 0 0 #0000), var(--tw-shadow)"
        }

        if (options.fullWidth) {
            styles.width = "100%"
        } else if (options.width) {
            styles.width = options.width
        }

        if (options.height) {
            styles.height = options.height
        } else if (options.autoHeight) {
            styles.height = "auto"
        }

        if (options.maxWidth) styles.maxWidth = options.maxWidth
        if (options.maxHeight) styles.maxHeight = options.maxHeight

        if (webview.slot === "fixed") {
            if (options?.window?.draggable) {
                styles.left = `${position.x}px`
                styles.top = `${position.y}px`
            } else {
                styles.left = options?.window?.defaultX !== undefined ? `${options?.window?.defaultX}px` : "0"
                styles.top = options?.window?.defaultY !== undefined ? `${options?.window?.defaultY}px` : "0"
            }
        }

        // Parse custom style string
        if (customStyle) {
            Object.assign(styles, customStyle)
        }

        return styles
    }, [webview.slot, options, position, customStyle])

    // Dragging logic
    const handleMouseDown = React.useCallback((e: React.MouseEvent) => {
        if (!options?.window?.draggable) return
        e.preventDefault()
        setIsDragging(true)
        dragStartPos.current = {
            x: e.clientX,
            y: e.clientY,
            elemX: position.x,
            elemY: position.y,
        }
    }, [options?.window?.draggable, position])

    React.useEffect(() => {
        if (!isDragging) return

        const handleMouseMove = (e: MouseEvent) => {
            const deltaX = e.clientX - dragStartPos.current.x
            const deltaY = e.clientY - dragStartPos.current.y

            let newX = getXPosition(dragStartPos.current.elemX + deltaX)
            let newY = getYPosition(dragStartPos.current.elemY + deltaY)

            onUpdatePosition(webview.webviewId, newX, newY)
        }

        const handleMouseUp = () => setIsDragging(false)

        document.addEventListener("mousemove", handleMouseMove)
        document.addEventListener("mouseup", handleMouseUp)

        return () => {
            document.removeEventListener("mousemove", handleMouseMove)
            document.removeEventListener("mouseup", handleMouseUp)
        }
    }, [isDragging, onUpdatePosition, webview.webviewId])

    // Tell the plugin that the webview is loaded
    const handleIframeLoaded = () => {
        log.info("Loaded iframe webview", webview.webviewId)
        sendWebviewLoadedEvent({ slot: webview.slot }, webview.extensionId)
    }

    const { width: iframeWidth } = useMeasureElement(iframeRef, [webview.options?.hidden])

    if (webview.options?.hidden) return null

    return (
        <div
            data-webview-container={webview.webviewId}
            style={{
                ...(webview.slot === "fixed" ? {
                    position: "fixed",
                    left: position.x,
                    top: position.y,
                    zIndex: options.zIndex || (webview.slot === "fixed" ? 100 : 5),
                } : {
                    display: "block",
                    width: "100%",
                }),
            }}
        >

            {!!options?.window?.draggable && webview.slot === "fixed" && <div
                data-plugin-webview-el="drag-handle"
                onMouseDown={handleMouseDown}
                className="absolute top-0 left-0 right-0 h-8 cursor-move z-[9999]" // bg-gradient-to-b from-black/20 to-transparent
                style={{ pointerEvents: "auto", width: iframeWidth }}
            />}

            <iframe
                ref={iframeRef}
                id={`webview-${webview.webviewId}`}
                srcDoc={webview.src}
                sandbox="allow-scripts allow-forms"
                style={buildStyles}
                onLoad={handleIframeLoaded}
                className={cn(
                    // (isResizing) && "pointer-events-none",
                    (isDragging) && "pointer-events-none",
                    options.className,
                )}
            />
        </div>
    )
}
