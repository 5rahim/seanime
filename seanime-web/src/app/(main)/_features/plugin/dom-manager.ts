import { useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { logger } from "@/lib/helpers/debug"
import { useEffect, useRef } from "react"
import { PluginDOMElement } from "./generated/plugin-dom-types"
import {
    Plugin_Server_DOMCreateEventPayload,
    Plugin_Server_DOMManipulateEventPayload,
    Plugin_Server_DOMObserveEventPayload,
    Plugin_Server_DOMQueryEventPayload,
    Plugin_Server_DOMQueryOneEventPayload,
    Plugin_Server_DOMStopObserveEventPayload,
    PluginClientEvents,
} from "./generated/plugin-events"

function uuidv4(): string {
    // @ts-ignore
    return ([1e7] + -1e3 + -4e3 + -8e3 + -1e11).replace(/[018]/g, (c) =>
        (c ^ (crypto.getRandomValues(new Uint8Array(1))[0] & (15 >> (c / 4)))).toString(16),
    )
}

type ElementToDOMElementOptions = {
    withInnerHTML?: boolean
    identifyChildren?: boolean
}

/**
 * DOM Manager for plugins
 * Handles DOM manipulation requests from plugins
 */
export function useDOMManager(extensionId: string) {
    const { sendPluginMessage } = useWebsocketSender()

    const elementObserversRef = useRef<Map<string, {
        selector: string;
        withInnerHTML?: boolean;
        identifyChildren?: boolean;
        callback: (elements: Element[]) => void
    }>>(new Map())
    const observedElementsRef = useRef<Map<string, Set<string>>>(new Map()) // Track observed elements by observerId
    const eventListenersRef = useRef<Map<string, { elementId: string; eventType: string; callback: (event: Event) => void }>>(new Map())
    const mutationObserverRef = useRef<MutationObserver | null>(null)
    const disposedRef = useRef<boolean>(false)
    const domReadySentRef = useRef<boolean>(false)

    const safeSendPluginMessage = (type: string, payload: any) => {
        if (disposedRef.current) return // Prevent sending messages if disposed
        sendPluginMessage(type, payload, extensionId)
    }

    // Send DOM ready event when document is loaded
    const sendDOMReadyEvent = () => {
        if (disposedRef.current || domReadySentRef.current) return
        domReadySentRef.current = true
        safeSendPluginMessage(PluginClientEvents.DOMReady, {})
    }

    // Convert a DOM element to a serializable object
    const elementToDOMElement = (element: Element, options?: ElementToDOMElementOptions): PluginDOMElement => {
        const attributes: Record<string, string> = {}

        // Get all attributes
        for (let i = 0; i < element.attributes.length; i++) {
            const attr = element.attributes[i]
            attributes[attr.name] = attr.value
        }

        // Ensure the element has an ID
        if (!element.id) {
            const id = `plugin-element-${uuidv4()}`
            element.setAttribute("id", id)
            attributes.id = id
        } else {
            attributes.id = element.id
        }

        // Add dataset as attributes with data- prefix
        if (element instanceof HTMLElement) {
            for (const key in element.dataset) {
                if (Object.prototype.hasOwnProperty.call(element.dataset, key)) {
                    attributes[`data-${key}`] = element.dataset[key] || ""
                }
            }
        }

        // If identifyChildren is true, assign IDs to all children recursively
        if (options?.identifyChildren) {
            // Get all descendants (not just direct children)
            element.querySelectorAll("*").forEach(child => {
                if (!child.id) {
                    const childId = `plugin-element-${uuidv4()}`
                    child.setAttribute("id", childId)
                }
            })
        }

        return {
            id: attributes.id,
            tagName: element.tagName.toLowerCase(),
            attributes,
            // textContent: element.textContent || undefined,
            innerHTML: options?.withInnerHTML ? element.innerHTML : undefined,
            children: [],
            // children: Array.from(element.children).map(child => elementToDOMElement(child)),
        }
    }

    // Convert an event to a serializable object
    const eventToObject = (event: Event): Record<string, any> => {
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

    // Initialize mutation observer to watch for DOM changes
    const initMutationObserver = () => {
        if (typeof window === "undefined" || typeof MutationObserver === "undefined") return

        mutationObserverRef.current = new MutationObserver((mutations) => {
            if (disposedRef.current) return // Skip processing if disposed

            // Process each mutation to find modified elements that match our selectors
            const processedElements = new Set<Element>()

            mutations.forEach(mutation => {
                // Handle added nodes
                if (mutation.type === "childList") {
                    mutation.addedNodes.forEach(node => {
                        if (node instanceof Element) {
                            processedElements.add(node)
                            // Also check descendant elements
                            node.querySelectorAll("*").forEach(el => processedElements.add(el))
                        }
                    })
                }

                // Handle modified nodes (attributes or character data)
                if (mutation.type === "attributes" || mutation.type === "characterData") {
                    const target = mutation.target instanceof Element ?
                        mutation.target :
                        mutation.target.parentElement

                    if (target) processedElements.add(target)
                }
            })

            // Check each observer against processed elements
            elementObserversRef.current.forEach((observer, observerId) => {
                // Track newly matched elements for this observer
                const matchedElements: Element[] = []
                const observedSet = observedElementsRef.current.get(observerId) || new Set()

                // Check if any of the processed elements match our selector
                processedElements.forEach(element => {
                    if (element.matches(observer.selector)) {
                        // Only assign ID if element matches the selector
                        if (!element.id) {
                            element.id = `plugin-element-${uuidv4()}`
                        }
                        matchedElements.push(element)
                    }
                })

                // Also do a general query to catch any elements that might match but weren't directly modified
                document.querySelectorAll(observer.selector).forEach(element => {
                    const id = element.id

                    // If element matches and doesn't have an ID, assign one
                    if (!id) {
                        element.id = `plugin-element-${uuidv4()}`
                    }

                    // If we haven't seen this element before, add it
                    if (!observedSet.has(element.id) && !matchedElements.includes(element)) {
                        matchedElements.push(element)
                    }
                })

                if (matchedElements.length > 0) {
                    // Convert to DOM elements
                    const domElements = matchedElements.map(e => {
                        // Ensure ID
                        if (!e.id) e.id = `plugin-element-${uuidv4()}`
                        return elementToDOMElement(e, {
                            withInnerHTML: observer.withInnerHTML,
                            identifyChildren: observer.identifyChildren,
                        })
                    })

                    // Update observed set with any new elements
                    domElements.forEach(elem => observedSet.add(elem.id))
                    observedElementsRef.current.set(observerId, observedSet)

                    // Call the callback
                    observer.callback(matchedElements)

                    // Send the elements to the plugin
                    safeSendPluginMessage(PluginClientEvents.DOMObserveResult, {
                        observerId,
                        elements: domElements,
                    })
                }
            })
        })

        // Start observing the document with the configured parameters
        mutationObserverRef.current.observe(document.body, {
            childList: true,
            subtree: true,
            attributes: true,
            characterData: true,
        })
    }

    // Handler functions
    const handleDOMQuery = (payload: Plugin_Server_DOMQueryEventPayload) => {
        const { selector, requestId, withInnerHTML, identifyChildren } = payload
        if (disposedRef.current) return
        const elements = document.querySelectorAll(selector)
        const domElements = Array.from(elements).map(e => elementToDOMElement(e, { withInnerHTML, identifyChildren }))
        safeSendPluginMessage(PluginClientEvents.DOMQueryResult, {
            requestId,
            elements: domElements,
        })
    }

    const handleDOMQueryOne = (payload: Plugin_Server_DOMQueryOneEventPayload) => {
        const { selector, requestId, withInnerHTML, identifyChildren } = payload
        if (disposedRef.current) return
        const element = document.querySelector(selector)
        const domElement = element ? elementToDOMElement(element, { withInnerHTML, identifyChildren }) : null

        safeSendPluginMessage(PluginClientEvents.DOMQueryOneResult, {
            requestId,
            element: domElement,
        })
    }

    const handleDOMObserve = (payload: Plugin_Server_DOMObserveEventPayload) => {
        const { selector, observerId, withInnerHTML, identifyChildren } = payload
        if (disposedRef.current) return

        // console.log(`Registering observer ${observerId} for selector ${selector}`)

        // Initialize set to track observed elements for this observer
        observedElementsRef.current.set(observerId, new Set())

        // Store the observer
        elementObserversRef.current.set(observerId, {
            selector,
            withInnerHTML,
            identifyChildren,
            callback: (elements) => {
                // This callback is called when elements matching the selector are found
                // console.log(`Observer ${observerId} callback with ${elements.length} elements matching ${selector}`, elements.map(e => e.id))
            },
        })

        // Immediately check for matching elements
        const elements = document.querySelectorAll(selector)
        if (elements.length > 0) {
            // Ensure each element has an ID and add to matched set
            const matchedElements = Array.from(elements).map(element => {
                if (!element.id) {
                    element.id = `plugin-element-${uuidv4()}`
                }
                return element
            })

            // Convert to DOM elements for sending to plugin
            const domElements = matchedElements.map(e => elementToDOMElement(e, { withInnerHTML, identifyChildren }))

            // Track these elements as observed
            const observedSet = observedElementsRef.current.get(observerId)!
            domElements.forEach(elem => observedSet.add(elem.id))

            // Call the callback
            elementObserversRef.current.get(observerId)?.callback(matchedElements)

            // Send matched elements to the plugin
            safeSendPluginMessage(PluginClientEvents.DOMObserveResult, {
                observerId,
                elements: domElements,
            })
        }
    }

    const handleDOMStopObserve = (payload: Plugin_Server_DOMStopObserveEventPayload) => {
        const { observerId } = payload
        elementObserversRef.current.delete(observerId)
        observedElementsRef.current.delete(observerId)
    }

    const handleDOMCreate = (payload: Plugin_Server_DOMCreateEventPayload) => {
        const { tagName, requestId } = payload
        if (disposedRef.current) return
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

        safeSendPluginMessage(PluginClientEvents.DOMCreateResult, {
            requestId,
            element: elementToDOMElement(element),
        })
    }

    const handleDOMManipulate = (payload: Plugin_Server_DOMManipulateEventPayload) => {
        if (disposedRef.current) return
        const { elementId, action, params, requestId } = payload
        const element = document.getElementById(elementId)

        if (!element) {
            // console.error(`Element with ID ${elementId} not found`)
            safeSendPluginMessage(PluginClientEvents.DOMElementUpdated, {
                elementId,
                action,
                result: undefined,
                requestId,
            })
            return
        }

        let result: any = null

        // Utility to safely store original value in data-original attribute
        const storeOriginalValue = (el: Element, type: string, key: string, value: any) => {
            if (!(el instanceof HTMLElement)) return

            let originalData: Record<string, Record<string, any>> = {}
            if (el.dataset.original) {
                try {
                    originalData = JSON.parse(el.dataset.original) as Record<string, Record<string, any>>
                }
                catch {
                    originalData = {}
                }
            }

            // Initialize type record if it doesn't exist
            if (!originalData[type]) {
                originalData[type] = {}
            }

            // Only store the value if it's not already set for this type+key
            if (originalData[type][key] === undefined) {
                originalData[type][key] = value
                el.dataset.original = JSON.stringify(originalData)
            }
        }

        switch (action) {
            case "setAttribute":
                // Store previous attribute value
                if (element instanceof HTMLElement && params.name) {
                    const prevValue = element.getAttribute(params.name)
                    storeOriginalValue(element, "attribute", params.name, prevValue)
                }

                element.setAttribute(params.name, params.value)
                result = true
                break
            case "removeAttribute":
                // Store previous attribute value
                if (element instanceof HTMLElement && params.name) {
                    const prevValue = element.getAttribute(params.name)
                    storeOriginalValue(element, "attribute", params.name, prevValue)
                }

                element.removeAttribute(params.name)
                break
            case "setInnerHTML":
                // Store previous HTML
                if (element instanceof HTMLElement) {
                    storeOriginalValue(element, "html", "innerHTML", element.innerHTML)
                }

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
            case "getText":
                result = element.textContent
                break
            case "setText":
                // Store previous text content
                if (element instanceof HTMLElement) {
                    storeOriginalValue(element, "text", "textContent", element.textContent)
                }

                element.textContent = params.text
                break
            case "getAttribute":
                result = element.getAttribute(params.name)
                break
            case "getAttributes":
                result = {}
                for (let i = 0; i < element.attributes.length; i++) {
                    const attr = element.attributes[i]
                    result[attr.name] = attr.value
                }
                break
            case "hasAttribute":
                result = element.hasAttribute(params.name)
                break
            case "getProperty":
                result = (element as any)[params.name]
                break
            case "setProperty":
                // Store previous property value
                if (element instanceof HTMLElement && params.name) {
                    storeOriginalValue(element, "property", params.name, (element as any)[params.name])
                }

                (element as any)[params.name] = params.value
                break
            case "addClass":
                element.classList.add(params.className)
                break
            case "removeClass":
                // Store previous class presence
                if (element instanceof HTMLElement && params.className) {
                    storeOriginalValue(element, "class", params.className, element.classList.contains(params.className))
                }

                element.classList.remove(params.className)
                break
            case "hasClass":
                result = element.classList.contains(params.className)
                break
            case "setStyle":
                // Store previous style value
                if (element instanceof HTMLElement && params.property) {
                    storeOriginalValue(element, "style", params.property, element.style.getPropertyValue(params.property))
                }

                element.style.setProperty(params.property, params.value)
                break
            case "getStyle":
                if (params.property) {
                    result = element.style.getPropertyValue(params.property)
                } else {
                    result = {}
                    for (let i = 0; i < element.style.length; i++) {
                        const prop = element.style[i]
                        result[prop] = element.style.getPropertyValue(prop)
                    }
                }
                break
            case "getComputedStyle":
                result = window.getComputedStyle(element).getPropertyValue(params.property)
                break
            case "append":
                const childToAppend = document.getElementById(params.childId)
                if (childToAppend) {
                    element.appendChild(childToAppend)
                }
                break
            case "before":
                const siblingBefore = document.getElementById(params.siblingId)
                if (siblingBefore && element.parentNode) {
                    element.parentNode.insertBefore(siblingBefore, element)
                }
                break
            case "after":
                const siblingAfter = document.getElementById(params.siblingId)
                if (siblingAfter && element.parentNode) {
                    element.parentNode.insertBefore(siblingAfter, element.nextSibling)
                }
                break
            case "remove":
                element.remove()
                break
            case "getParent":
                result = element.parentElement ? elementToDOMElement(element.parentElement, {
                    withInnerHTML: params.withInnerHTML,
                    identifyChildren: params.identifyChildren,
                }) : null
                break
            case "getChildren":
                result = Array.from(element.children).map(e => elementToDOMElement(e, {
                    withInnerHTML: params.withInnerHTML,
                    identifyChildren: params.identifyChildren,
                }))
                break
            case "query":
                // Find elements within the current element using the provided selector
                const queryElements = element.querySelectorAll(params.selector)
                const queryDomElements = Array.from(queryElements).map(e => elementToDOMElement(e, {
                    withInnerHTML: params.withInnerHTML,
                    identifyChildren: params.identifyChildren,
                }))

                // Send the results back using the DOMQueryResult event
                safeSendPluginMessage(PluginClientEvents.DOMQueryResult, {
                    requestId: params.requestId,
                    elements: queryDomElements,
                })
                return // Return early since we're sending separate event
            case "queryOne":
                // Find a single element within the current element using the provided selector
                const queryOneElement = element.querySelector(params.selector)
                const _queryOneElements = element.querySelectorAll(params.selector)
                const queryOneDomElement = queryOneElement ? elementToDOMElement(queryOneElement, {
                    withInnerHTML: params.withInnerHTML,
                    identifyChildren: params.identifyChildren,
                }) : null

                // Send the result back using the DOMQueryOneResult event
                safeSendPluginMessage(PluginClientEvents.DOMQueryOneResult, {
                    requestId: params.requestId,
                    element: queryOneDomElement,
                })
                return // Return early since we're sending separate event
            case "addEventListener":
                const listenerId = params.listenerId
                const eventType = params.event

                // Store the event listener
                eventListenersRef.current.set(listenerId, {
                    elementId,
                    eventType,
                    callback: (event) => {
                        // Convert event to a serializable object
                        const eventData = eventToObject(event)

                        // Send the event to the plugin
                        safeSendPluginMessage(PluginClientEvents.DOMEventTriggered, {
                            elementId,
                            eventType,
                            event: eventData,
                        })
                    },
                })

                // Add the event listener
                element.addEventListener(eventType, eventListenersRef.current.get(listenerId)!.callback)
                break
            case "removeEventListener":
                const listenerIdToRemove = params.listenerId
                const eventTypeToRemove = params.event

                // Get the event listener
                const listener = eventListenersRef.current.get(listenerIdToRemove)
                if (listener) {
                    // Remove the event listener
                    element.removeEventListener(eventTypeToRemove, listener.callback)
                    // Remove from the map
                    eventListenersRef.current.delete(listenerIdToRemove)
                }
                break
            case "getDataAttribute":
                if (element instanceof HTMLElement) {
                    result = element.dataset[params.key]
                }
                break
            case "getDataAttributes":
                if (element instanceof HTMLElement) {
                    result = { ...element.dataset }
                } else {
                    result = {}
                }
                break
            case "setDataAttribute":
                if (element instanceof HTMLElement) {
                    // Store previous data attribute value
                    if (params.key) {
                        storeOriginalValue(element, "dataset", params.key, element.dataset[params.key])
                    }

                    element.dataset[params.key] = params.value
                }
                break
            case "removeDataAttribute":
                if (element instanceof HTMLElement) {
                    // Store previous data attribute value
                    if (params.key) {
                        storeOriginalValue(element, "dataset", params.key, element.dataset[params.key])
                    }

                    delete element.dataset[params.key]
                }
                break
            case "hasDataAttribute":
                if (element instanceof HTMLElement) {
                    result = params.key in element.dataset
                } else {
                    result = false
                }
                break
            case "hasStyle":
                result = element.style.getPropertyValue(params.property) !== ""
                break
            case "removeStyle":
                // Store previous style value
                if (element instanceof HTMLElement && params.property) {
                    storeOriginalValue(element, "style", params.property, element.style.getPropertyValue(params.property))
                }

                element.style.removeProperty(params.property)
                break
            default:
                console.warn(`Unknown DOM action: ${action}`)
        }

        // console.log(`DOMElementUpdated: ${elementId} ${action} ${requestId}`)

        // Send the result back to the plugin
        safeSendPluginMessage(PluginClientEvents.DOMElementUpdated, {
            elementId,
            action,
            result,
            requestId,
        })
    }

    const cleanup = () => {
        logger("DOMManager").info("Cleaning up DOMManager for extension", extensionId)
        // Mark as disposed to prevent further message sending
        disposedRef.current = true
        domReadySentRef.current = false

        // Stop the mutation observer
        if (mutationObserverRef.current) {
            mutationObserverRef.current.disconnect()
            mutationObserverRef.current = null
        }

        // Remove all event listeners
        eventListenersRef.current.forEach((listener, listenerId) => {
            const element = document.getElementById(listener.elementId)
            if (element) {
                element.removeEventListener(listener.eventType, listener.callback)
            }
        })

        // Clear the maps
        elementObserversRef.current.clear()
        eventListenersRef.current.clear()
        observedElementsRef.current.clear()

        // Remove plugin container if it exists
        const container = document.getElementById("plugin-dom-container")
        if (container) {
            container.remove()
        }
    }

    useEffect(() => {
        logger("DOMManager").info("DOMManager hook initialized for extension", extensionId)

        // Send DOM ready event if document is already loaded
        if (document.readyState === "complete") {
            sendDOMReadyEvent()
        } else {
            // Otherwise wait for the document to be loaded
            window.addEventListener("load", sendDOMReadyEvent)
        }

        // Initialize mutation observer
        initMutationObserver()

        // Cleanup function
        return () => {
            cleanup()
            // Remove load event listener if added
            if (!domReadySentRef.current) {
                window.removeEventListener("load", sendDOMReadyEvent)
            }
        }
    }, [extensionId])

    return {
        handleDOMQuery,
        handleDOMQueryOne,
        handleDOMObserve,
        handleDOMStopObserve,
        handleDOMCreate,
        handleDOMManipulate,
        cleanup,
    }
}
