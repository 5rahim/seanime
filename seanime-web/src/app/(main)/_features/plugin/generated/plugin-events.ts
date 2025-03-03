// This file is auto-generated. Do not edit.
import { useWebsocketPluginMessageListener, useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { useCallback } from "react"

export enum PluginClientEvents {
    TrayRender = "tray:render",
    TrayRenderAll = "tray:render-all",
    TrayOpened = "tray:opened",
    TrayClosed = "tray:closed",
    FormSubmitted = "form:submitted",
    ScreenChanged = "screen:changed",
    HandlerTriggered = "handler:triggered",
    FieldRefSendValue = "field-ref:send-value",
}

export enum PluginServerEvents {
    TrayUpdated = "tray:updated",
    FormReset = "form:reset",
    FormSetValues = "form:set-values",
    FieldRefSetValue = "field-ref:set-value",
    ScreenNavigateTo = "screen:navigate-to",
}

/////////////////////////////////////////////////////////////////////////////////////
// Client to server
/////////////////////////////////////////////////////////////////////////////////////

export type Plugin_Client_TrayRenderEventPayload = {}

export function usePluginSendTrayRenderEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendTrayRenderEvent = useCallback((payload: Plugin_Client_TrayRenderEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.TrayRender, payload, extensionID)
    }, [])

    return {
        sendTrayRenderEvent,
    }
}

export type Plugin_Client_TrayRenderAllEventPayload = {}

export function usePluginSendTrayRenderAllEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendTrayRenderAllEvent = useCallback((payload: Plugin_Client_TrayRenderAllEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.TrayRenderAll, payload, extensionID)
    }, [])

    return {
        sendTrayRenderAllEvent,
    }
}

export type Plugin_Client_TrayOpenedEventPayload = {}

export function usePluginSendTrayOpenedEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendTrayOpenedEvent = useCallback((payload: Plugin_Client_TrayOpenedEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.TrayOpened, payload, extensionID)
    }, [])

    return {
        sendTrayOpenedEvent,
    }
}

export type Plugin_Client_TrayClosedEventPayload = {}

export function usePluginSendTrayClosedEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendTrayClosedEvent = useCallback((payload: Plugin_Client_TrayClosedEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.TrayClosed, payload, extensionID)
    }, [])

    return {
        sendTrayClosedEvent,
    }
}

export type Plugin_Client_FormSubmittedEventPayload = {
    formName: string
    data: Record<string, any>
}

export function usePluginSendFormSubmittedEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendFormSubmittedEvent = useCallback((payload: Plugin_Client_FormSubmittedEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.FormSubmitted, payload, extensionID)
    }, [])

    return {
        sendFormSubmittedEvent,
    }
}

export type Plugin_Client_ScreenChangedEventPayload = {
    pathname: string
    query: string
}

export function usePluginSendScreenChangedEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendScreenChangedEvent = useCallback((payload: Plugin_Client_ScreenChangedEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.ScreenChanged, payload, extensionID)
    }, [])

    return {
        sendScreenChangedEvent,
    }
}

export type Plugin_Client_HandlerTriggeredEventPayload = {
    handlerName: string
    event: Record<string, any>
}

export function usePluginSendHandlerTriggeredEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendHandlerTriggeredEvent = useCallback((payload: Plugin_Client_HandlerTriggeredEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.HandlerTriggered, payload, extensionID)
    }, [])

    return {
        sendHandlerTriggeredEvent,
    }
}

export type Plugin_Client_FieldRefSendValueEventPayload = {
    fieldRef: string
    value: any
}

export function usePluginSendFieldRefSendValueEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendFieldRefSendValueEvent = useCallback((payload: Plugin_Client_FieldRefSendValueEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.FieldRefSendValue, payload, extensionID)
    }, [])

    return {
        sendFieldRefSendValueEvent,
    }
}

/////////////////////////////////////////////////////////////////////////////////////
// Server to client
/////////////////////////////////////////////////////////////////////////////////////

export type Plugin_Server_TrayUpdatedEventPayload = {
    components: any
}

export function usePluginListenTrayUpdatedEvent(cb: (payload: Plugin_Server_TrayUpdatedEventPayload) => void, extensionID: string) {
    return useWebsocketPluginMessageListener<Plugin_Server_TrayUpdatedEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.TrayUpdated,
        onMessage: cb,
    })
}

export type Plugin_Server_FormResetEventPayload = {
    formName: string
    fieldToReset: string
}

export function usePluginListenFormResetEvent(cb: (payload: Plugin_Server_FormResetEventPayload) => void, extensionID: string) {
    return useWebsocketPluginMessageListener<Plugin_Server_FormResetEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.FormReset,
        onMessage: cb,
    })
}

export type Plugin_Server_FormSetValuesEventPayload = {
    formName: string
    data: Record<string, any>
}

export function usePluginListenFormSetValuesEvent(cb: (payload: Plugin_Server_FormSetValuesEventPayload) => void, extensionID: string) {
    return useWebsocketPluginMessageListener<Plugin_Server_FormSetValuesEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.FormSetValues,
        onMessage: cb,
    })
}

export type Plugin_Server_FieldRefSetValueEventPayload = {
    fieldRef: string
    value: any
}

export function usePluginListenFieldRefSetValueEvent(cb: (payload: Plugin_Server_FieldRefSetValueEventPayload) => void, extensionID: string) {
    return useWebsocketPluginMessageListener<Plugin_Server_FieldRefSetValueEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.FieldRefSetValue,
        onMessage: cb,
    })
}

export type Plugin_Server_ScreenNavigateToEventPayload = {
    path: string
}

export function usePluginListenScreenNavigateToEvent(cb: (payload: Plugin_Server_ScreenNavigateToEventPayload) => void, extensionID: string) {
    return useWebsocketPluginMessageListener<Plugin_Server_ScreenNavigateToEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.ScreenNavigateTo,
        onMessage: cb,
    })
}

