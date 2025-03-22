import { useWebsocketPluginMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { useEffect } from "react"
import { useDOMManager } from "./dom-manager"
import { PluginServerEvents } from "./generated/plugin-events"

export function PluginHandler({ extensionId }: { extensionId: string }) {
    // DOM Manager
    const {
        handleDOMQuery,
        handleDOMQueryOne,
        handleDOMObserve,
        handleDOMStopObserve,
        handleDOMCreate,
        handleDOMManipulate,
        cleanup: cleanupDOMManager,
    } = useDOMManager(extensionId)

    // Listen for DOM events
    useWebsocketPluginMessageListener({
        extensionId,
        type: PluginServerEvents.DOMQuery,
        onMessage: (payload: any) => {
            handleDOMQuery(payload.selector, payload.requestId)
        },
    })

    useWebsocketPluginMessageListener({
        extensionId,
        type: PluginServerEvents.DOMQueryOne,
        onMessage: (payload: any) => {
            handleDOMQueryOne(payload.selector, payload.requestId)
        },
    })

    useWebsocketPluginMessageListener({
        extensionId,
        type: PluginServerEvents.DOMObserve,
        onMessage: (payload: any) => {
            handleDOMObserve(payload.selector, payload.observerID)
        },
    })

    useWebsocketPluginMessageListener({
        extensionId,
        type: PluginServerEvents.DOMStopObserve,
        onMessage: (payload: any) => {
            handleDOMStopObserve(payload.observerID)
        },
    })

    useWebsocketPluginMessageListener({
        extensionId,
        type: PluginServerEvents.DOMCreate,
        onMessage: (payload: any) => {
            handleDOMCreate(payload.tagName, payload.requestId)
        },
    })

    useWebsocketPluginMessageListener({
        extensionId,
        type: PluginServerEvents.DOMManipulate,
        onMessage: (payload: any) => {
            handleDOMManipulate({
                elementId: payload.elementID,
                action: payload.action,
                params: payload.params,
            })
        },
    })

    // No need for cleanup useEffect anymore as the useDOMManager hook handles its own cleanup
    
    return null
}
