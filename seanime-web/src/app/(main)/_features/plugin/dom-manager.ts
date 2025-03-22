import { useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { PluginDOMElement, PluginDOMManipulateOptions } from "./generated/plugin-dom-types"
import { PluginClientEvents } from "./generated/plugin-events"
import { useEffect, useRef } from "react"
import { logger } from "@/lib/helpers/debug"

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
export function useDOMManager(extensionId: string) {
    const { sendPluginMessage } = useWebsocketSender()
    
    const elementObserversRef = useRef<Map<string, { selector: string; callback: (elements: Element[]) => void }>>(new Map())
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
    const elementToDOMElement = (element: Element): PluginDOMElement => {
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

        // Add dataset as attributes with data- prefix
        if (element instanceof HTMLElement) {
            for (const key in element.dataset) {
                if (Object.prototype.hasOwnProperty.call(element.dataset, key)) {
                    attributes[`data-${key}`] = element.dataset[key] || ""
                }
            }
        }

        return {
            id: attributes.id,
            tagName: element.tagName.toLowerCase(),
            attributes,
            textContent: element.textContent || undefined,
            innerHTML: element.innerHTML || undefined,
            children: Array.from(element.children).map(child => elementToDOMElement(child)),
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
            
            // For each observer, check if any elements match the selector
            elementObserversRef.current.forEach((observer, observerId) => {
                const elements = document.querySelectorAll(observer.selector)
                if (elements.length > 0) {
                    // Get currently observed elements for this observer
                    const observedSet = observedElementsRef.current.get(observerId) || new Set()
                    
                    // Filter out elements that have already been observed
                    const newElements = Array.from(elements).filter(element => {
                        const id = element.id || `plugin-element-${uuidv4()}`
                        if (!element.id) element.id = id
                        return !observedSet.has(id)
                    })
                    
                    // If there are new elements, process them
                    if (newElements.length > 0) {
                        // Convert elements to DOM elements
                        const domElements = newElements.map(e => elementToDOMElement(e))
                        
                        // Add new elements to the observed set
                        domElements.forEach(elem => observedSet.add(elem.id))
                        observedElementsRef.current.set(observerId, observedSet)
                        
                        // Call the callback
                        observer.callback(newElements)
                        
                        // Send the elements to the plugin
                        safeSendPluginMessage(PluginClientEvents.DOMObserveResult, {
                            observerId,
                            elements: domElements,
                        })
                    }
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
    const handleDOMQuery = (selector: string, requestId: string) => {
        if (disposedRef.current) return
        const elements = document.querySelectorAll(selector)
        const domElements = Array.from(elements).map(e => elementToDOMElement(e))
        safeSendPluginMessage(PluginClientEvents.DOMQueryResult, {
            requestId,
            elements: domElements,
        })
    }

    const handleDOMQueryOne = (selector: string, requestId: string) => {
        if (disposedRef.current) return
        const element = document.querySelector(selector)
        const domElement = element ? elementToDOMElement(element) : null

        safeSendPluginMessage(PluginClientEvents.DOMQueryOneResult, {
            requestId,
            element: domElement,
        })
    }

    const handleDOMObserve = (selector: string, observerId: string) => {
        if (disposedRef.current) return
        
        // console.log(`Registering observer ${observerId} for selector ${selector}`)
        
        // Initialize set to track observed elements for this observer
        observedElementsRef.current.set(observerId, new Set())
        
        // Store the observer
        elementObserversRef.current.set(observerId, {
            selector,
            callback: (elements) => {
                // This callback is called when elements matching the selector are found
                // console.log(`Observer ${observerId} callback with ${elements.length} elements matching ${selector}`)
            },
        })

        // Immediately check for matching elements
        const elements = document.querySelectorAll(selector)
        if (elements.length > 0) {
            const domElements = Array.from(elements).map(e => elementToDOMElement(e))
            
            // Track these elements as observed
            const observedSet = observedElementsRef.current.get(observerId)!
            domElements.forEach(elem => observedSet.add(elem.id))
            
            safeSendPluginMessage(PluginClientEvents.DOMObserveResult, {
                observerId,
                elements: domElements,
            })
        }
    }

    const handleDOMStopObserve = (observerId: string) => {
        elementObserversRef.current.delete(observerId)
        observedElementsRef.current.delete(observerId)
    }

    const handleDOMCreate = (tagName: string, requestId: string) => {
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

    const handleDOMManipulate = (options: PluginDOMManipulateOptions) => {
        if (disposedRef.current) return
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
            case "getText":
                result = element.textContent
                break
            case "setText":
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
                (element as any)[params.name] = params.value
                break
            case "addClass":
                element.classList.add(params.className)
                break
            case "removeClass":
                element.classList.remove(params.className)
                break
            case "hasClass":
                result = element.classList.contains(params.className)
                break
            case "setStyle":
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
                result = element.parentElement ? elementToDOMElement(element.parentElement) : null
                break
            case "getChildren":
                result = Array.from(element.children).map(e => elementToDOMElement(e))
                break
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
                    element.dataset[params.key] = params.value
                }
                break
            case "removeDataAttribute":
                if (element instanceof HTMLElement) {
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
                element.style.removeProperty(params.property)
                break
            default:
                console.warn(`Unknown DOM action: ${action}`)
        }

        // Send the result back to the plugin
        safeSendPluginMessage(PluginClientEvents.DOMElementUpdated, {
            elementId,
            action,
            result,
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
