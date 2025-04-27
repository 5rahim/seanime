import { useWebsocketMessageListener, useWebsocketPluginMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { WSEvents } from "@/lib/server/ws-events"
import { useDOMManager } from "./dom-manager"
import {
    Plugin_Server_DOMCreateEventPayload,
    Plugin_Server_DOMManipulateEventPayload,
    Plugin_Server_DOMObserveEventPayload,
    Plugin_Server_DOMQueryEventPayload,
    Plugin_Server_DOMQueryOneEventPayload,
    Plugin_Server_DOMStopObserveEventPayload,
    PluginServerEvents,
} from "./generated/plugin-events"

export function PluginHandler({ extensionId, onUnloaded }: { extensionId: string, onUnloaded: () => void }) {
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

    useWebsocketMessageListener({
        type: WSEvents.PLUGIN_UNLOADED,
        onMessage: (_extensionId) => {
            if (_extensionId === extensionId) {
                cleanupDOMManager()
                onUnloaded()
            }
        },
    })

    // Listen for DOM events
    useWebsocketPluginMessageListener({
        extensionId,
        type: PluginServerEvents.DOMQuery,
        onMessage: (payload: Plugin_Server_DOMQueryEventPayload) => {
            handleDOMQuery(payload)
        },
    })

    useWebsocketPluginMessageListener({
        extensionId,
        type: PluginServerEvents.DOMQueryOne,
        onMessage: (payload: Plugin_Server_DOMQueryOneEventPayload) => {
            handleDOMQueryOne(payload)
        },
    })

    useWebsocketPluginMessageListener({
        extensionId,
        type: PluginServerEvents.DOMObserve,
        onMessage: (payload: Plugin_Server_DOMObserveEventPayload) => {
            handleDOMObserve(payload)
        },
    })

    useWebsocketPluginMessageListener({
        extensionId,
        type: PluginServerEvents.DOMStopObserve,
        onMessage: (payload: Plugin_Server_DOMStopObserveEventPayload) => {
            handleDOMStopObserve(payload)
        },
    })

    useWebsocketPluginMessageListener({
        extensionId,
        type: PluginServerEvents.DOMCreate,
        onMessage: (payload: Plugin_Server_DOMCreateEventPayload) => {
            handleDOMCreate(payload)
        },
    })

    useWebsocketPluginMessageListener({
        extensionId,
        type: PluginServerEvents.DOMManipulate,
        onMessage: (payload: Plugin_Server_DOMManipulateEventPayload) => {
            handleDOMManipulate(payload)
        },
    })

    return null
}
