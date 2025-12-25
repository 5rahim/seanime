import { useWebsocketMessageListener, useWebsocketPluginMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { useIsMainTabRef } from "@/app/websocket-provider"
import { WSEvents } from "@/lib/server/ws-events"
import { useDOMManager } from "./dom-manager"
import {
    Plugin_Server_DOMCreateEventPayload,
    Plugin_Server_DOMManipulateEventPayload,
    Plugin_Server_DOMObserveEventPayload,
    Plugin_Server_DOMObserveInViewEventPayload,
    Plugin_Server_DOMQueryEventPayload,
    Plugin_Server_DOMQueryOneEventPayload,
    Plugin_Server_DOMStopObserveEventPayload,
    PluginServerEvents,
} from "./generated/plugin-events"

export function PluginHandler({ extensionId, onUnloaded }: { extensionId: string, onUnloaded: () => void }) {
    const isMainTabRef = useIsMainTabRef()

    // DOM Manager
    const {
        handleDOMQuery,
        handleDOMQueryOne,
        handleDOMObserve,
        handleDOMObserveInView,
        handleDOMStopObserve,
        handleDOMCreate,
        handleDOMManipulate,
        cleanup: cleanupDOMManager,
    } = useDOMManager(extensionId)

    useWebsocketMessageListener({
        type: WSEvents.PLUGIN_UNLOADED,
        onMessage: (_extensionId) => {
            if (_extensionId === extensionId) {
                cleanupDOMManager()
                onUnloaded()
            }
        },
    })

    useWebsocketPluginMessageListener({
        extensionId,
        type: PluginServerEvents.DOMQuery,
        onMessage: (payload: Plugin_Server_DOMQueryEventPayload) => {
            if (!isMainTabRef.current) return
            handleDOMQuery(payload)
        },
    })

    useWebsocketPluginMessageListener({
        extensionId,
        type: PluginServerEvents.DOMQueryOne,
        onMessage: (payload: Plugin_Server_DOMQueryOneEventPayload) => {
            if (!isMainTabRef.current) return
            handleDOMQueryOne(payload)
        },
    })

    useWebsocketPluginMessageListener({
        extensionId,
        type: PluginServerEvents.DOMObserve,
        onMessage: (payload: Plugin_Server_DOMObserveEventPayload) => {
            if (!isMainTabRef.current) return
            handleDOMObserve(payload)
        },
    })

    useWebsocketPluginMessageListener({
        extensionId,
        type: PluginServerEvents.DOMObserveInView,
        onMessage: (payload: Plugin_Server_DOMObserveInViewEventPayload) => {
            if (!isMainTabRef.current) return
            handleDOMObserveInView(payload)
        },
    })

    useWebsocketPluginMessageListener({
        extensionId,
        type: PluginServerEvents.DOMStopObserve,
        onMessage: (payload: Plugin_Server_DOMStopObserveEventPayload) => {
            if (!isMainTabRef.current) return
            handleDOMStopObserve(payload)
        },
    })

    useWebsocketPluginMessageListener({
        extensionId,
        type: PluginServerEvents.DOMCreate,
        onMessage: (payload: Plugin_Server_DOMCreateEventPayload) => {
            if (!isMainTabRef.current) return
            handleDOMCreate(payload)
        },
    })

    useWebsocketPluginMessageListener({
        extensionId,
        type: PluginServerEvents.DOMManipulate,
        onMessage: (payload: Plugin_Server_DOMManipulateEventPayload) => {
            if (!isMainTabRef.current) {
                return
            }
            handleDOMManipulate(payload)
        },
    })

    return null
}
