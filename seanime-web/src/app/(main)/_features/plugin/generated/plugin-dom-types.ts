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
    action: "setAttribute" | "removeAttribute" | "setTextContent" | "setInnerHTML" | "appendChild" | "removeChild"
    params: Record<string, any>
}

export type PluginDOMEventData = {
    observerId: string
    elementId: string
    eventType: string
    data: any
}
