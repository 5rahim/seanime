import * as React from "react"

/* -------------------------------------------------------------------------------------------------
 * useEventListener
 * -----------------------------------------------------------------------------------------------*/

export function useEventListener<
    KW extends keyof WindowEventMap,
    KH extends keyof HTMLElementEventMap,
    KM extends keyof MediaQueryListEventMap,
    T extends HTMLElement | MediaQueryList | void = void,
>(
    eventName: KW | KH | KM,
    handler: (
        event:
            | WindowEventMap[KW]
            | HTMLElementEventMap[KH]
            | MediaQueryListEventMap[KM]
            | Event,
    ) => void,
    element?: React.RefObject<T>,
    options?: boolean | AddEventListenerOptions,
) {
    // Create a ref that stores handler
    const savedHandler = React.useRef(handler)

    useIsomorphicLayoutEffect(() => {
        savedHandler.current = handler
    }, [handler])

    React.useEffect(() => {
        // Define the listening target
        const targetElement: T | Window = element?.current ?? window

        if (!(targetElement && targetElement.addEventListener)) return

        // Create event listener that calls handler function stored in ref
        const listener: typeof handler = event => savedHandler.current(event)

        targetElement.addEventListener(eventName, listener, options)

        // Remove event listener on cleanup
        return () => {
            targetElement.removeEventListener(eventName, listener, options)
        }
    }, [eventName, element, options])
}


/* -------------------------------------------------------------------------------------------------
 * useIsomorphicLayoutEffect
 * -----------------------------------------------------------------------------------------------*/

export const useIsomorphicLayoutEffect = typeof window !== "undefined" ? React.useLayoutEffect : React.useEffect

/* -------------------------------------------------------------------------------------------------
 * useUpdateEffect
 * -----------------------------------------------------------------------------------------------*/

export function useUpdateEffect(effect: React.EffectCallback, deps?: React.DependencyList) {
    const isInitialMount = React.useRef(true)

    React.useEffect(() => {
        if (isInitialMount.current) {
            isInitialMount.current = false
        } else {
            return effect()
        }
    }, deps)
}
