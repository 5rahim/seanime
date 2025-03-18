import { useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { PluginDOMElement, PluginDOMManipulateOptions } from "./generated/plugin-dom-types"
import { PluginClientEvents } from "./generated/plugin-events"

function uuidv4(): string {
    // @ts-ignore
    return ([1e7] + -1e3 + -4e3 + -8e3 + -1e11).replace(/[018]/g, (c) =>
        (c ^ (crypto.getRandomValues(new Uint8Array(1))[0] & (15 >> (c / 4)))).toString(16),
    )
}

/**
 * DOM Manager for plugins
 * Handles DOM manipulation requests from plugins
 */
export class DOMManager {
    extensionId: string
    private elementObservers: Map<string, { selector: string; callback: (elements: Element[]) => void }> = new Map()
    private eventListeners: Map<string, { elementId: string; eventType: string; callback: (event: Event) => void }> = new Map()
    private mutationObserver: MutationObserver | null = null
    private sendPluginMessage: (type: string, payload: any, extensionId?: string) => void

    constructor(extensionId: string) {
        this.extensionId = extensionId
        const { sendPluginMessage } = useWebsocketSender()
        // logger("DOMManager").info("DOMManager constructor", extensionId)

        sendPluginMessage(PluginClientEvents.DOMReady, {}, extensionId)

        this.sendPluginMessage = (type, payload) => {
            sendPluginMessage(type, payload, extensionId)
        }

        // Initialize mutation observer to watch for DOM changes
        this.initMutationObserver()
    }

    /**
     * Handle DOM query request from plugin
     */
    handleDOMQuery(selector: string, requestId: string) {
        const elements = document.querySelectorAll(selector)
        const domElements = Array.from(elements).map(e => this.elementToDOMElement(e))
        this.sendPluginMessage(PluginClientEvents.DOMQueryResult, {
            requestId,
            elements: domElements,
        })
    }

    /**
     * Handle DOM query one request from plugin
     */
    handleDOMQueryOne(selector: string, requestId: string) {
        const element = document.querySelector(selector)
        const domElement = element ? this.elementToDOMElement(element) : null

        this.sendPluginMessage(PluginClientEvents.DOMQueryOneResult, {
            requestId,
            element: domElement,
        })
    }

    /**
     * Handle DOM observe request from plugin
     */
    handleDOMObserve(selector: string, observerId: string) {
        // Store the observer
        this.elementObservers.set(observerId, {
            selector,
            callback: (elements) => {
                // This callback is called when elements matching the selector are found
                console.log(`Observer ${observerId} found ${elements.length} elements matching ${selector}`)
            },
        })

        // Immediately check for matching elements
        const elements = document.querySelectorAll(selector)
        if (elements.length > 0) {
            const domElements = Array.from(elements).map(e => this.elementToDOMElement(e))
            this.sendPluginMessage(PluginClientEvents.DOMObserveResult, {
                observerId,
                elements: domElements,
            })
        }
    }

    /**
     * Handle DOM stop observe request from plugin
     */
    handleDOMStopObserve(observerId: string) {
        this.elementObservers.delete(observerId)
    }

    /**
     * Handle DOM create element request from plugin
     */
    handleDOMCreate(tagName: string, requestId: string) {
        const element = document.createElement(tagName)
        element.id = `plugin-element-${uuidv4()}`

        // Add to a hidden container for now
        let container = document.getElementById("plugin-dom-container")
        if (!container) {
            container = document.createElement("div")
            container.id = "plugin-dom-container"
            container.style.display = "none"
            document.body.appendChild(container)
        }

        container.appendChild(element)

        this.sendPluginMessage(PluginClientEvents.DOMCreateResult, {
            requestId,
            element: this.elementToDOMElement(element),
        })
    }

    /**
     * Handle DOM manipulate request from plugin
     */
    handleDOMManipulate(options: PluginDOMManipulateOptions) {
        const { elementId, action, params } = options
        const element = document.getElementById(elementId)

        if (!element) {
            console.error(`Element with ID ${elementId} not found`)
            return
        }

        let result: any = null

        switch (action) {
            case "setAttribute":
                element.setAttribute(params.name, params.value)
                break
            case "removeAttribute":
                element.removeAttribute(params.name)
                break
            case "setTextContent":
                element.textContent = params.text
                break
            case "setInnerHTML":
                element.innerHTML = params.html
                break
            case "appendChild":
                const child = document.getElementById(params.childId)
                if (child) {
                    element.appendChild(child)
                }
                break
            case "removeChild":
                const childToRemove = document.getElementById(params.childId)
                if (childToRemove && element.contains(childToRemove)) {
                    element.removeChild(childToRemove)
                }
                break
            default:
                // Handle other actions based on the DOM methods in dom.go
                if (action === "getText") {
                    result = element.textContent
                } else if (action === "setText") {
                    element.textContent = params.text
                } else if (action === "getAttribute") {
                    result = element.getAttribute(params.name)
                } else if (action === "addClass") {
                    element.classList.add(params.className)
                } else if (action === "removeClass") {
                    element.classList.remove(params.className)
                } else if (action === "hasClass") {
                    result = element.classList.contains(params.className)
                } else if (action === "setStyle") {
                    element.style.setProperty(params.property, params.value)
                } else if (action === "getStyle") {
                    result = element.style.getPropertyValue(params.property)
                } else if (action === "getComputedStyle") {
                    result = window.getComputedStyle(element).getPropertyValue(params.property)
                } else if (action === "append") {
                    const childToAppend = document.getElementById(params.childId)
                    if (childToAppend) {
                        element.appendChild(childToAppend)
                    }
                } else if (action === "before") {
                    const siblingBefore = document.getElementById(params.siblingId)
                    if (siblingBefore && element.parentNode) {
                        element.parentNode.insertBefore(siblingBefore, element)
                    }
                } else if (action === "after") {
                    const siblingAfter = document.getElementById(params.siblingId)
                    if (siblingAfter && element.parentNode) {
                        element.parentNode.insertBefore(siblingAfter, element.nextSibling)
                    }
                } else if (action === "remove") {
                    element.remove()
                } else if (action === "getParent") {
                    result = element.parentElement ? this.elementToDOMElement(element.parentElement) : null
                } else if (action === "getChildren") {
                    result = Array.from(element.children).map(e => this.elementToDOMElement(e))
                } else if (action === "addEventListener") {
                    const listenerId = params.listenerId
                    const eventType = params.event

                    // Store the event listener
                    this.eventListeners.set(listenerId, {
                        elementId,
                        eventType,
                        callback: (event) => {
                            // Convert event to a serializable object
                            const eventData = this.eventToObject(event)

                            // Send the event to the plugin
                            this.sendPluginMessage(PluginClientEvents.DOMEventTriggered, {
                                elementId,
                                eventType,
                                event: eventData,
                            })
                        },
                    })

                    // Add the event listener
                    element.addEventListener(eventType, this.eventListeners.get(listenerId)!.callback)
                } else if (action === "removeEventListener") {
                    const listenerId = params.listenerId
                    const eventType = params.event

                    // Get the event listener
                    const listener = this.eventListeners.get(listenerId)
                    if (listener) {
                        // Remove the event listener
                        element.removeEventListener(eventType, listener.callback)
                        // Remove from the map
                        this.eventListeners.delete(listenerId)
                    }
                }
        }

        // Send the result back to the plugin
        this.sendPluginMessage(PluginClientEvents.DOMElementUpdated, {
            elementId,
            action,
            result,
        })
    }

    /**
     * Clean up resources
     */
    cleanup() {
        // Stop the mutation observer
        if (this.mutationObserver) {
            this.mutationObserver.disconnect()
        }

        // Remove all event listeners
        this.eventListeners.forEach((listener, listenerId) => {
            const element = document.getElementById(listener.elementId)
            if (element) {
                element.removeEventListener(listener.eventType, listener.callback)
            }
        })

        // Clear the maps
        this.elementObservers.clear()
        this.eventListeners.clear()
    }

    /**
     * Initialize the mutation observer to watch for DOM changes
     */
    private initMutationObserver() {
        if (typeof window === "undefined" || typeof MutationObserver === "undefined") return

        this.mutationObserver = new MutationObserver((mutations) => {
            // For each observer, check if any elements match the selector
            this.elementObservers.forEach((observer, observerId) => {
                const elements = document.querySelectorAll(observer.selector)
                if (elements.length > 0) {
                    // Convert elements to DOM elements
                    const domElements = Array.from(elements).map(e => this.elementToDOMElement(e))
                    // Call the callback
                    observer.callback(Array.from(elements))
                    // Send the elements to the plugin
                    this.sendPluginMessage(PluginClientEvents.DOMEventTriggered, {
                        observerId,
                        elements: domElements,
                    })
                }
            })
        })

        // Start observing the document with the configured parameters
        this.mutationObserver.observe(document.body, {
            childList: true,
            subtree: true,
            attributes: true,
            characterData: true,
        })
    }

    /**
     * Convert a DOM element to a serializable object
     */
    private elementToDOMElement(element: Element): PluginDOMElement {
        const attributes: Record<string, string> = {}

        // Get all attributes
        for (let i = 0; i < element.attributes.length; i++) {
            const attr = element.attributes[i]
            attributes[attr.name] = attr.value
        }

        // Ensure the element has an ID
        if (!attributes.id) {
            const id = `plugin-element-${uuidv4()}`
            element.setAttribute("id", id)
            attributes.id = id
        }

        return {
            id: attributes.id,
            tagName: element.tagName.toLowerCase(),
            attributes,
            textContent: element.textContent || undefined,
            innerHTML: element.innerHTML || undefined,
            children: Array.from(element.children).map(child => this.elementToDOMElement(child)),
        }
    }

    /**
     * Convert an event to a serializable object
     */
    private eventToObject(event: Event): Record<string, any> {
        const result: Record<string, any> = {
            type: event.type,
            bubbles: event.bubbles,
            cancelable: event.cancelable,
            composed: event.composed,
            timeStamp: event.timeStamp,
        }

        // Add properties from MouseEvent
        if (event instanceof MouseEvent) {
            result.clientX = event.clientX
            result.clientY = event.clientY
            result.screenX = event.screenX
            result.screenY = event.screenY
            result.altKey = event.altKey
            result.ctrlKey = event.ctrlKey
            result.shiftKey = event.shiftKey
            result.metaKey = event.metaKey
            result.button = event.button
            result.buttons = event.buttons
        }

        // Add properties from KeyboardEvent
        if (event instanceof KeyboardEvent) {
            result.key = event.key
            result.code = event.code
            result.location = event.location
            result.repeat = event.repeat
            result.altKey = event.altKey
            result.ctrlKey = event.ctrlKey
            result.shiftKey = event.shiftKey
            result.metaKey = event.metaKey
        }

        return result
    }
}

/**
 * Create a DOM manager hook
 */
export function useDOMManager(extensionId: string) {
    const domManager = new DOMManager(extensionId)

    return {
        handleDOMQuery: domManager.handleDOMQuery.bind(domManager),
        handleDOMQueryOne: domManager.handleDOMQueryOne.bind(domManager),
        handleDOMObserve: domManager.handleDOMObserve.bind(domManager),
        handleDOMStopObserve: domManager.handleDOMStopObserve.bind(domManager),
        handleDOMCreate: domManager.handleDOMCreate.bind(domManager),
        handleDOMManipulate: domManager.handleDOMManipulate.bind(domManager),
        cleanup: domManager.cleanup.bind(domManager),
    }
}
