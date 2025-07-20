// This file is auto-generated. Do not edit.
import { useWebsocketPluginMessageListener, useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { useCallback } from "react"

export enum PluginClientEvents {
    RenderTray = "tray:render",
    ListTrayIcons = "tray:list-icons",
    TrayOpened = "tray:opened",
    TrayClosed = "tray:closed",
    TrayClicked = "tray:clicked",
    ListCommandPalettes = "command-palette:list",
    CommandPaletteOpened = "command-palette:opened",
    CommandPaletteClosed = "command-palette:closed",
    RenderCommandPalette = "command-palette:render",
    CommandPaletteInput = "command-palette:input",
    CommandPaletteItemSelected = "command-palette:item-selected",
    ActionRenderAnimePageButtons = "action:anime-page-buttons:render",
    ActionRenderAnimePageDropdownItems = "action:anime-page-dropdown-items:render",
    ActionRenderMangaPageButtons = "action:manga-page-buttons:render",
    ActionRenderMediaCardContextMenuItems = "action:media-card-context-menu-items:render",
    ActionRenderAnimeLibraryDropdownItems = "action:anime-library-dropdown-items:render",
    ActionRenderEpisodeCardContextMenuItems = "action:episode-card-context-menu-items:render",
    ActionRenderEpisodeGridItemMenuItems = "action:episode-grid-item-menu-items:render",
    ActionClicked = "action:clicked",
    FormSubmitted = "form:submitted",
    ScreenChanged = "screen:changed",
    EventHandlerTriggered = "handler:triggered",
    FieldRefSendValue = "field-ref:send-value",
    DOMQueryResult = "dom:query-result",
    DOMQueryOneResult = "dom:query-one-result",
    DOMObserveResult = "dom:observe-result",
    DOMStopObserve = "dom:stop-observe",
    DOMCreateResult = "dom:create-result",
    DOMElementUpdated = "dom:element-updated",
    DOMEventTriggered = "dom:event-triggered",
    DOMReady = "dom:ready",
}

export enum PluginServerEvents {
    TrayUpdated = "tray:updated",
    TrayIcon = "tray:icon",
    TrayBadgeUpdated = "tray:badge-updated",
    TrayOpen = "tray:open",
    TrayClose = "tray:close",
    CommandPaletteInfo = "command-palette:info",
    CommandPaletteUpdated = "command-palette:updated",
    CommandPaletteOpen = "command-palette:open",
    CommandPaletteClose = "command-palette:close",
    CommandPaletteGetInput = "command-palette:get-input",
    CommandPaletteSetInput = "command-palette:set-input",
    ActionRenderAnimePageButtons = "action:anime-page-buttons:updated",
    ActionRenderAnimePageDropdownItems = "action:anime-page-dropdown-items:updated",
    ActionRenderMangaPageButtons = "action:manga-page-buttons:updated",
    ActionRenderMediaCardContextMenuItems = "action:media-card-context-menu-items:updated",
    ActionRenderEpisodeCardContextMenuItems = "action:episode-card-context-menu-items:updated",
    ActionRenderEpisodeGridItemMenuItems = "action:episode-grid-item-menu-items:updated",
    ActionRenderAnimeLibraryDropdownItems = "action:anime-library-dropdown-items:updated",
    FormReset = "form:reset",
    FormSetValues = "form:set-values",
    FieldRefSetValue = "field-ref:set-value",
    FatalError = "fatal-error",
    ScreenNavigateTo = "screen:navigate-to",
    ScreenReload = "screen:reload",
    ScreenGetCurrent = "screen:get-current",
    DOMQuery = "dom:query",
    DOMQueryOne = "dom:query-one",
    DOMObserve = "dom:observe",
    DOMStopObserve = "dom:stop-observe",
    DOMCreate = "dom:create",
    DOMManipulate = "dom:manipulate",
    DOMObserveInView = "dom:observe-in-view",
}

/////////////////////////////////////////////////////////////////////////////////////
// Client to server
/////////////////////////////////////////////////////////////////////////////////////

export type Plugin_Client_RenderTrayEventPayload = {}

export function usePluginSendRenderTrayEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendRenderTrayEvent = useCallback((payload: Plugin_Client_RenderTrayEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.RenderTray, payload, extensionID)
    }, [])

    return {
        sendRenderTrayEvent,
    }
}

export type Plugin_Client_ListTrayIconsEventPayload = {}

export function usePluginSendListTrayIconsEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendListTrayIconsEvent = useCallback((payload: Plugin_Client_ListTrayIconsEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.ListTrayIcons, payload, extensionID)
    }, [])

    return {
        sendListTrayIconsEvent,
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

export type Plugin_Client_TrayClickedEventPayload = {}

export function usePluginSendTrayClickedEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendTrayClickedEvent = useCallback((payload: Plugin_Client_TrayClickedEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.TrayClicked, payload, extensionID)
    }, [])

    return {
        sendTrayClickedEvent,
    }
}

export type Plugin_Client_ListCommandPalettesEventPayload = {}

export function usePluginSendListCommandPalettesEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendListCommandPalettesEvent = useCallback((payload: Plugin_Client_ListCommandPalettesEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.ListCommandPalettes, payload, extensionID)
    }, [])

    return {
        sendListCommandPalettesEvent,
    }
}

export type Plugin_Client_CommandPaletteOpenedEventPayload = {}

export function usePluginSendCommandPaletteOpenedEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendCommandPaletteOpenedEvent = useCallback((payload: Plugin_Client_CommandPaletteOpenedEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.CommandPaletteOpened, payload, extensionID)
    }, [])

    return {
        sendCommandPaletteOpenedEvent,
    }
}

export type Plugin_Client_CommandPaletteClosedEventPayload = {}

export function usePluginSendCommandPaletteClosedEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendCommandPaletteClosedEvent = useCallback((payload: Plugin_Client_CommandPaletteClosedEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.CommandPaletteClosed, payload, extensionID)
    }, [])

    return {
        sendCommandPaletteClosedEvent,
    }
}

export type Plugin_Client_RenderCommandPaletteEventPayload = {}

export function usePluginSendRenderCommandPaletteEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendRenderCommandPaletteEvent = useCallback((payload: Plugin_Client_RenderCommandPaletteEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.RenderCommandPalette, payload, extensionID)
    }, [])

    return {
        sendRenderCommandPaletteEvent,
    }
}

export type Plugin_Client_CommandPaletteInputEventPayload = {
    value: string
}

export function usePluginSendCommandPaletteInputEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendCommandPaletteInputEvent = useCallback((payload: Plugin_Client_CommandPaletteInputEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.CommandPaletteInput, payload, extensionID)
    }, [])

    return {
        sendCommandPaletteInputEvent,
    }
}

export type Plugin_Client_CommandPaletteItemSelectedEventPayload = {
    itemId: string
}

export function usePluginSendCommandPaletteItemSelectedEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendCommandPaletteItemSelectedEvent = useCallback((payload: Plugin_Client_CommandPaletteItemSelectedEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.CommandPaletteItemSelected, payload, extensionID)
    }, [])

    return {
        sendCommandPaletteItemSelectedEvent,
    }
}

export type Plugin_Client_ActionRenderAnimePageButtonsEventPayload = {}

export function usePluginSendActionRenderAnimePageButtonsEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendActionRenderAnimePageButtonsEvent = useCallback((payload: Plugin_Client_ActionRenderAnimePageButtonsEventPayload,
        extensionID?: string,
    ) => {
        sendPluginMessage(PluginClientEvents.ActionRenderAnimePageButtons, payload, extensionID)
    }, [])

    return {
        sendActionRenderAnimePageButtonsEvent,
    }
}

export type Plugin_Client_ActionRenderAnimePageDropdownItemsEventPayload = {}

export function usePluginSendActionRenderAnimePageDropdownItemsEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendActionRenderAnimePageDropdownItemsEvent = useCallback((payload: Plugin_Client_ActionRenderAnimePageDropdownItemsEventPayload,
        extensionID?: string,
    ) => {
        sendPluginMessage(PluginClientEvents.ActionRenderAnimePageDropdownItems, payload, extensionID)
    }, [])

    return {
        sendActionRenderAnimePageDropdownItemsEvent,
    }
}

export type Plugin_Client_ActionRenderMangaPageButtonsEventPayload = {}

export function usePluginSendActionRenderMangaPageButtonsEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendActionRenderMangaPageButtonsEvent = useCallback((payload: Plugin_Client_ActionRenderMangaPageButtonsEventPayload,
        extensionID?: string,
    ) => {
        sendPluginMessage(PluginClientEvents.ActionRenderMangaPageButtons, payload, extensionID)
    }, [])

    return {
        sendActionRenderMangaPageButtonsEvent,
    }
}

export type Plugin_Client_ActionRenderMediaCardContextMenuItemsEventPayload = {}

export function usePluginSendActionRenderMediaCardContextMenuItemsEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendActionRenderMediaCardContextMenuItemsEvent = useCallback((payload: Plugin_Client_ActionRenderMediaCardContextMenuItemsEventPayload,
        extensionID?: string,
    ) => {
        sendPluginMessage(PluginClientEvents.ActionRenderMediaCardContextMenuItems, payload, extensionID)
    }, [])

    return {
        sendActionRenderMediaCardContextMenuItemsEvent,
    }
}

export type Plugin_Client_ActionRenderAnimeLibraryDropdownItemsEventPayload = {}

export function usePluginSendActionRenderAnimeLibraryDropdownItemsEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendActionRenderAnimeLibraryDropdownItemsEvent = useCallback((payload: Plugin_Client_ActionRenderAnimeLibraryDropdownItemsEventPayload,
        extensionID?: string,
    ) => {
        sendPluginMessage(PluginClientEvents.ActionRenderAnimeLibraryDropdownItems, payload, extensionID)
    }, [])

    return {
        sendActionRenderAnimeLibraryDropdownItemsEvent,
    }
}

export type Plugin_Client_ActionRenderEpisodeCardContextMenuItemsEventPayload = {}

export function usePluginSendActionRenderEpisodeCardContextMenuItemsEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendActionRenderEpisodeCardContextMenuItemsEvent = useCallback((payload: Plugin_Client_ActionRenderEpisodeCardContextMenuItemsEventPayload,
        extensionID?: string,
    ) => {
        sendPluginMessage(PluginClientEvents.ActionRenderEpisodeCardContextMenuItems, payload, extensionID)
    }, [])

    return {
        sendActionRenderEpisodeCardContextMenuItemsEvent,
    }
}

export type Plugin_Client_ActionRenderEpisodeGridItemMenuItemsEventPayload = {}

export function usePluginSendActionRenderEpisodeGridItemMenuItemsEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendActionRenderEpisodeGridItemMenuItemsEvent = useCallback((payload: Plugin_Client_ActionRenderEpisodeGridItemMenuItemsEventPayload,
        extensionID?: string,
    ) => {
        sendPluginMessage(PluginClientEvents.ActionRenderEpisodeGridItemMenuItems, payload, extensionID)
    }, [])

    return {
        sendActionRenderEpisodeGridItemMenuItemsEvent,
    }
}

export type Plugin_Client_ActionClickedEventPayload = {
    actionId: string
    event: Record<string, any>
}

export function usePluginSendActionClickedEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendActionClickedEvent = useCallback((payload: Plugin_Client_ActionClickedEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.ActionClicked, payload, extensionID)
    }, [])

    return {
        sendActionClickedEvent,
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

export type Plugin_Client_EventHandlerTriggeredEventPayload = {
    handlerName: string
    event: Record<string, any>
}

export function usePluginSendEventHandlerTriggeredEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendEventHandlerTriggeredEvent = useCallback((payload: Plugin_Client_EventHandlerTriggeredEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.EventHandlerTriggered, payload, extensionID)
    }, [])

    return {
        sendEventHandlerTriggeredEvent,
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

export type Plugin_Client_DOMQueryResultEventPayload = {
    requestId: string
    elements: Array<any>
}

export function usePluginSendDOMQueryResultEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendDOMQueryResultEvent = useCallback((payload: Plugin_Client_DOMQueryResultEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.DOMQueryResult, payload, extensionID)
    }, [])

    return {
        sendDOMQueryResultEvent,
    }
}

export type Plugin_Client_DOMQueryOneResultEventPayload = {
    requestId: string
    element: any
}

export function usePluginSendDOMQueryOneResultEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendDOMQueryOneResultEvent = useCallback((payload: Plugin_Client_DOMQueryOneResultEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.DOMQueryOneResult, payload, extensionID)
    }, [])

    return {
        sendDOMQueryOneResultEvent,
    }
}

export type Plugin_Client_DOMObserveResultEventPayload = {
    observerId: string
    elements: Array<any>
}

export function usePluginSendDOMObserveResultEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendDOMObserveResultEvent = useCallback((payload: Plugin_Client_DOMObserveResultEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.DOMObserveResult, payload, extensionID)
    }, [])

    return {
        sendDOMObserveResultEvent,
    }
}

export type Plugin_Client_DOMStopObserveEventPayload = {
    observerId: string
}

export function usePluginSendDOMStopObserveEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendDOMStopObserveEvent = useCallback((payload: Plugin_Client_DOMStopObserveEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.DOMStopObserve, payload, extensionID)
    }, [])

    return {
        sendDOMStopObserveEvent,
    }
}

export type Plugin_Client_DOMCreateResultEventPayload = {
    requestId: string
    element: any
}

export function usePluginSendDOMCreateResultEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendDOMCreateResultEvent = useCallback((payload: Plugin_Client_DOMCreateResultEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.DOMCreateResult, payload, extensionID)
    }, [])

    return {
        sendDOMCreateResultEvent,
    }
}

export type Plugin_Client_DOMElementUpdatedEventPayload = {
    elementId: string
    action: string
    result: any
    requestId: string
}

export function usePluginSendDOMElementUpdatedEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendDOMElementUpdatedEvent = useCallback((payload: Plugin_Client_DOMElementUpdatedEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.DOMElementUpdated, payload, extensionID)
    }, [])

    return {
        sendDOMElementUpdatedEvent,
    }
}

export type Plugin_Client_DOMEventTriggeredEventPayload = {
    elementId: string
    eventType: string
    event: Record<string, any>
}

export function usePluginSendDOMEventTriggeredEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendDOMEventTriggeredEvent = useCallback((payload: Plugin_Client_DOMEventTriggeredEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.DOMEventTriggered, payload, extensionID)
    }, [])

    return {
        sendDOMEventTriggeredEvent,
    }
}

export type Plugin_Client_DOMReadyEventPayload = {}

export function usePluginSendDOMReadyEvent() {
    const { sendPluginMessage } = useWebsocketSender()

    const sendDOMReadyEvent = useCallback((payload: Plugin_Client_DOMReadyEventPayload, extensionID?: string) => {
        sendPluginMessage(PluginClientEvents.DOMReady, payload, extensionID)
    }, [])

    return {
        sendDOMReadyEvent,
    }
}

/////////////////////////////////////////////////////////////////////////////////////
// Server to client
/////////////////////////////////////////////////////////////////////////////////////

export type Plugin_Server_TrayUpdatedEventPayload = {
    components: any
}

export function usePluginListenTrayUpdatedEvent(cb: (payload: Plugin_Server_TrayUpdatedEventPayload, extensionId: string) => void,
    extensionID: string,
) {
    return useWebsocketPluginMessageListener<Plugin_Server_TrayUpdatedEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.TrayUpdated,
        onMessage: cb,
    })
}

export type Plugin_Server_TrayIconEventPayload = {
    extensionId: string
    extensionName: string
    iconUrl: string
    withContent: boolean
    tooltipText: string
    badgeNumber: number
    badgeIntent: string
    width: string
    minHeight: string
}

export function usePluginListenTrayIconEvent(cb: (payload: Plugin_Server_TrayIconEventPayload, extensionId: string) => void, extensionID: string) {
    return useWebsocketPluginMessageListener<Plugin_Server_TrayIconEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.TrayIcon,
        onMessage: cb,
    })
}

export type Plugin_Server_TrayBadgeUpdatedEventPayload = {
    badgeNumber: number
    badgeIntent: string
}

export function usePluginListenTrayBadgeUpdatedEvent(cb: (payload: Plugin_Server_TrayBadgeUpdatedEventPayload, extensionId: string) => void,
    extensionID: string,
) {
    return useWebsocketPluginMessageListener<Plugin_Server_TrayBadgeUpdatedEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.TrayBadgeUpdated,
        onMessage: cb,
    })
}

export type Plugin_Server_TrayOpenEventPayload = {
    extensionId: string
}

export function usePluginListenTrayOpenEvent(cb: (payload: Plugin_Server_TrayOpenEventPayload, extensionId: string) => void, extensionID: string) {
    return useWebsocketPluginMessageListener<Plugin_Server_TrayOpenEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.TrayOpen,
        onMessage: cb,
    })
}

export type Plugin_Server_TrayCloseEventPayload = {
    extensionId: string
}

export function usePluginListenTrayCloseEvent(cb: (payload: Plugin_Server_TrayCloseEventPayload, extensionId: string) => void, extensionID: string) {
    return useWebsocketPluginMessageListener<Plugin_Server_TrayCloseEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.TrayClose,
        onMessage: cb,
    })
}

export type Plugin_Server_CommandPaletteInfoEventPayload = {
    placeholder: string
    keyboardShortcut: string
}

export function usePluginListenCommandPaletteInfoEvent(cb: (payload: Plugin_Server_CommandPaletteInfoEventPayload, extensionId: string) => void,
    extensionID: string,
) {
    return useWebsocketPluginMessageListener<Plugin_Server_CommandPaletteInfoEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.CommandPaletteInfo,
        onMessage: cb,
    })
}

export type Plugin_Server_CommandPaletteUpdatedEventPayload = {
    placeholder: string
    items: any
}

export function usePluginListenCommandPaletteUpdatedEvent(cb: (payload: Plugin_Server_CommandPaletteUpdatedEventPayload, extensionId: string) => void,
    extensionID: string,
) {
    return useWebsocketPluginMessageListener<Plugin_Server_CommandPaletteUpdatedEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.CommandPaletteUpdated,
        onMessage: cb,
    })
}

export type Plugin_Server_CommandPaletteOpenEventPayload = {}

export function usePluginListenCommandPaletteOpenEvent(cb: (payload: Plugin_Server_CommandPaletteOpenEventPayload, extensionId: string) => void,
    extensionID: string,
) {
    return useWebsocketPluginMessageListener<Plugin_Server_CommandPaletteOpenEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.CommandPaletteOpen,
        onMessage: cb,
    })
}

export type Plugin_Server_CommandPaletteCloseEventPayload = {}

export function usePluginListenCommandPaletteCloseEvent(cb: (payload: Plugin_Server_CommandPaletteCloseEventPayload, extensionId: string) => void,
    extensionID: string,
) {
    return useWebsocketPluginMessageListener<Plugin_Server_CommandPaletteCloseEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.CommandPaletteClose,
        onMessage: cb,
    })
}

export type Plugin_Server_CommandPaletteGetInputEventPayload = {}

export function usePluginListenCommandPaletteGetInputEvent(cb: (payload: Plugin_Server_CommandPaletteGetInputEventPayload,
    extensionId: string,
) => void, extensionID: string) {
    return useWebsocketPluginMessageListener<Plugin_Server_CommandPaletteGetInputEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.CommandPaletteGetInput,
        onMessage: cb,
    })
}

export type Plugin_Server_CommandPaletteSetInputEventPayload = {
    value: string
}

export function usePluginListenCommandPaletteSetInputEvent(cb: (payload: Plugin_Server_CommandPaletteSetInputEventPayload,
    extensionId: string,
) => void, extensionID: string) {
    return useWebsocketPluginMessageListener<Plugin_Server_CommandPaletteSetInputEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.CommandPaletteSetInput,
        onMessage: cb,
    })
}

export type Plugin_Server_ActionRenderAnimePageButtonsEventPayload = {
    buttons: any
}

export function usePluginListenActionRenderAnimePageButtonsEvent(cb: (payload: Plugin_Server_ActionRenderAnimePageButtonsEventPayload,
    extensionId: string,
) => void, extensionID: string) {
    return useWebsocketPluginMessageListener<Plugin_Server_ActionRenderAnimePageButtonsEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.ActionRenderAnimePageButtons,
        onMessage: cb,
    })
}

export type Plugin_Server_ActionRenderAnimePageDropdownItemsEventPayload = {
    items: any
}

export function usePluginListenActionRenderAnimePageDropdownItemsEvent(cb: (payload: Plugin_Server_ActionRenderAnimePageDropdownItemsEventPayload,
    extensionId: string,
) => void, extensionID: string) {
    return useWebsocketPluginMessageListener<Plugin_Server_ActionRenderAnimePageDropdownItemsEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.ActionRenderAnimePageDropdownItems,
        onMessage: cb,
    })
}

export type Plugin_Server_ActionRenderMangaPageButtonsEventPayload = {
    buttons: any
}

export function usePluginListenActionRenderMangaPageButtonsEvent(cb: (payload: Plugin_Server_ActionRenderMangaPageButtonsEventPayload,
    extensionId: string,
) => void, extensionID: string) {
    return useWebsocketPluginMessageListener<Plugin_Server_ActionRenderMangaPageButtonsEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.ActionRenderMangaPageButtons,
        onMessage: cb,
    })
}

export type Plugin_Server_ActionRenderMediaCardContextMenuItemsEventPayload = {
    items: any
}

export function usePluginListenActionRenderMediaCardContextMenuItemsEvent(cb: (payload: Plugin_Server_ActionRenderMediaCardContextMenuItemsEventPayload,
    extensionId: string,
) => void, extensionID: string) {
    return useWebsocketPluginMessageListener<Plugin_Server_ActionRenderMediaCardContextMenuItemsEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.ActionRenderMediaCardContextMenuItems,
        onMessage: cb,
    })
}

export type Plugin_Server_ActionRenderEpisodeCardContextMenuItemsEventPayload = {
    items: any
}

export function usePluginListenActionRenderEpisodeCardContextMenuItemsEvent(cb: (payload: Plugin_Server_ActionRenderEpisodeCardContextMenuItemsEventPayload,
    extensionId: string,
) => void, extensionID: string) {
    return useWebsocketPluginMessageListener<Plugin_Server_ActionRenderEpisodeCardContextMenuItemsEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.ActionRenderEpisodeCardContextMenuItems,
        onMessage: cb,
    })
}

export type Plugin_Server_ActionRenderEpisodeGridItemMenuItemsEventPayload = {
    items: any
}

export function usePluginListenActionRenderEpisodeGridItemMenuItemsEvent(cb: (payload: Plugin_Server_ActionRenderEpisodeGridItemMenuItemsEventPayload,
    extensionId: string,
) => void, extensionID: string) {
    return useWebsocketPluginMessageListener<Plugin_Server_ActionRenderEpisodeGridItemMenuItemsEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.ActionRenderEpisodeGridItemMenuItems,
        onMessage: cb,
    })
}

export type Plugin_Server_ActionRenderAnimeLibraryDropdownItemsEventPayload = {
    items: any
}

export function usePluginListenActionRenderAnimeLibraryDropdownItemsEvent(cb: (payload: Plugin_Server_ActionRenderAnimeLibraryDropdownItemsEventPayload,
    extensionId: string,
) => void, extensionID: string) {
    return useWebsocketPluginMessageListener<Plugin_Server_ActionRenderAnimeLibraryDropdownItemsEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.ActionRenderAnimeLibraryDropdownItems,
        onMessage: cb,
    })
}

export type Plugin_Server_FormResetEventPayload = {
    formName: string
    fieldToReset: string
}

export function usePluginListenFormResetEvent(cb: (payload: Plugin_Server_FormResetEventPayload, extensionId: string) => void, extensionID: string) {
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

export function usePluginListenFormSetValuesEvent(cb: (payload: Plugin_Server_FormSetValuesEventPayload, extensionId: string) => void,
    extensionID: string,
) {
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

export function usePluginListenFieldRefSetValueEvent(cb: (payload: Plugin_Server_FieldRefSetValueEventPayload, extensionId: string) => void,
    extensionID: string,
) {
    return useWebsocketPluginMessageListener<Plugin_Server_FieldRefSetValueEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.FieldRefSetValue,
        onMessage: cb,
    })
}

export type Plugin_Server_FatalErrorEventPayload = {
    error: string
}

export function usePluginListenFatalErrorEvent(cb: (payload: Plugin_Server_FatalErrorEventPayload, extensionId: string) => void,
    extensionID: string,
) {
    return useWebsocketPluginMessageListener<Plugin_Server_FatalErrorEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.FatalError,
        onMessage: cb,
    })
}

export type Plugin_Server_ScreenNavigateToEventPayload = {
    path: string
}

export function usePluginListenScreenNavigateToEvent(cb: (payload: Plugin_Server_ScreenNavigateToEventPayload, extensionId: string) => void,
    extensionID: string,
) {
    return useWebsocketPluginMessageListener<Plugin_Server_ScreenNavigateToEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.ScreenNavigateTo,
        onMessage: cb,
    })
}

export type Plugin_Server_ScreenReloadEventPayload = {}

export function usePluginListenScreenReloadEvent(cb: (payload: Plugin_Server_ScreenReloadEventPayload, extensionId: string) => void,
    extensionID: string,
) {
    return useWebsocketPluginMessageListener<Plugin_Server_ScreenReloadEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.ScreenReload,
        onMessage: cb,
    })
}

export type Plugin_Server_ScreenGetCurrentEventPayload = {}

export function usePluginListenScreenGetCurrentEvent(cb: (payload: Plugin_Server_ScreenGetCurrentEventPayload, extensionId: string) => void,
    extensionID: string,
) {
    return useWebsocketPluginMessageListener<Plugin_Server_ScreenGetCurrentEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.ScreenGetCurrent,
        onMessage: cb,
    })
}

export type Plugin_Server_DOMQueryEventPayload = {
    selector: string
    requestId: string
    withInnerHTML: boolean
    withOuterHTML: boolean
    identifyChildren: boolean
}

export function usePluginListenDOMQueryEvent(cb: (payload: Plugin_Server_DOMQueryEventPayload, extensionId: string) => void, extensionID: string) {
    return useWebsocketPluginMessageListener<Plugin_Server_DOMQueryEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.DOMQuery,
        onMessage: cb,
    })
}

export type Plugin_Server_DOMQueryOneEventPayload = {
    selector: string
    requestId: string
    withInnerHTML: boolean
    withOuterHTML: boolean
    identifyChildren: boolean
}

export function usePluginListenDOMQueryOneEvent(cb: (payload: Plugin_Server_DOMQueryOneEventPayload, extensionId: string) => void,
    extensionID: string,
) {
    return useWebsocketPluginMessageListener<Plugin_Server_DOMQueryOneEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.DOMQueryOne,
        onMessage: cb,
    })
}

export type Plugin_Server_DOMObserveEventPayload = {
    selector: string
    observerId: string
    withInnerHTML: boolean
    withOuterHTML: boolean
    identifyChildren: boolean
}

export function usePluginListenDOMObserveEvent(cb: (payload: Plugin_Server_DOMObserveEventPayload, extensionId: string) => void,
    extensionID: string,
) {
    return useWebsocketPluginMessageListener<Plugin_Server_DOMObserveEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.DOMObserve,
        onMessage: cb,
    })
}

export type Plugin_Server_DOMStopObserveEventPayload = {
    observerId: string
}

export function usePluginListenDOMStopObserveEvent(cb: (payload: Plugin_Server_DOMStopObserveEventPayload, extensionId: string) => void,
    extensionID: string,
) {
    return useWebsocketPluginMessageListener<Plugin_Server_DOMStopObserveEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.DOMStopObserve,
        onMessage: cb,
    })
}

export type Plugin_Server_DOMCreateEventPayload = {
    tagName: string
    requestId: string
}

export function usePluginListenDOMCreateEvent(cb: (payload: Plugin_Server_DOMCreateEventPayload, extensionId: string) => void, extensionID: string) {
    return useWebsocketPluginMessageListener<Plugin_Server_DOMCreateEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.DOMCreate,
        onMessage: cb,
    })
}

export type Plugin_Server_DOMManipulateEventPayload = {
    elementId: string
    action: string
    params: Record<string, any>
    requestId: string
}

export function usePluginListenDOMManipulateEvent(cb: (payload: Plugin_Server_DOMManipulateEventPayload, extensionId: string) => void,
    extensionID: string,
) {
    return useWebsocketPluginMessageListener<Plugin_Server_DOMManipulateEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.DOMManipulate,
        onMessage: cb,
    })
}

export type Plugin_Server_DOMObserveInViewEventPayload = {
    selector: string
    observerId: string
    withInnerHTML: boolean
    withOuterHTML: boolean
    identifyChildren: boolean
    margin: string
}

export function usePluginListenDOMObserveInViewEvent(cb: (payload: Plugin_Server_DOMObserveInViewEventPayload, extensionId: string) => void,
    extensionID: string,
) {
    return useWebsocketPluginMessageListener<Plugin_Server_DOMObserveInViewEventPayload>({
        extensionId: extensionID,
        type: PluginServerEvents.DOMObserveInView,
        onMessage: cb,
    })
}

