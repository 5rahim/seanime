export type PluginDOMElement = {
    id: string
    tagName: string
    attributes: Record<string, string>
    children: PluginDOMElement[]
    textContent?: string
    innerHTML?: string
}

export type PluginDOMQueryResult = {
    elements: PluginDOMElement[]
}

export type PluginDOMQueryOneResult = {
    element?: PluginDOMElement
}

export type PluginDOMObserveResult = {
    observerId: string
}

export type PluginDOMCreateResult = {
    element: PluginDOMElement
}

export type PluginDOMManipulateOptions = {
    elementId: string
    requestId: string
    action: "setAttribute"
    | "removeAttribute"
    | "setInnerHTML"
    | "appendChild"
    | "removeChild"
    | "getText"
    | "setText"
    | "getAttribute"
    | "getAttributes"
    | "addClass"
    | "removeClass"
    | "hasClass"
    | "setStyle"
    | "getStyle"
    | "hasStyle"
    | "removeStyle"
    | "getComputedStyle"
    | "append"
    | "before"
    | "after"
    | "remove"
    | "getParent"
    | "getChildren"
    | "addEventListener"
    | "removeEventListener"
    | "getDataAttribute"
    | "getDataAttributes"
    | "setDataAttribute"
    | "removeDataAttribute"
    | "hasAttribute"
    | "hasDataAttribute"
    | "getProperty"
    | "setProperty"
        | "query"
        | "queryOne"
    params: Record<string, any>
}

export type PluginDOMEventData = {
    observerId: string
    elementId: string
    eventType: string
    data: any
}
