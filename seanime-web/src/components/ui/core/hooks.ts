"use client"

import React, { RefObject, useCallback, useEffect, useLayoutEffect, useRef, useState } from "react"

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
    element?: RefObject<T>,
    options?: boolean | AddEventListenerOptions,
) {
    // Create a ref that stores handler
    const savedHandler = useRef(handler)

    useIsomorphicLayoutEffect(() => {
        savedHandler.current = handler
    }, [handler])

    useEffect(() => {
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

export const useIsomorphicLayoutEffect = typeof window !== "undefined" ? useLayoutEffect : useEffect

/* -------------------------------------------------------------------------------------------------
 * useMediaQuery
 * -----------------------------------------------------------------------------------------------*/

export interface UseMediaQueryOptions {
    getInitialValueInEffect: boolean;
}

type MediaQueryCallback = (event: { matches: boolean; media: string }) => void;

/**
 * Older versions of Safari (shipped withCatalina and before) do not support addEventListener on matchMedia
 * https://stackoverflow.com/questions/56466261/matchmedia-addlistener-marked-as-deprecated-addeventlistener-equivalent
 * */
function attachMediaListener(query: MediaQueryList, callback: MediaQueryCallback) {
    try {
        query.addEventListener("change", callback)
        return () => query.removeEventListener("change", callback)
    } catch (e) {
        query.addListener(callback)
        return () => query.removeListener(callback)
    }
}

function getInitialValue(query: string, initialValue?: boolean) {
    if (typeof initialValue === "boolean") {
        return initialValue
    }

    if (typeof window !== "undefined" && "matchMedia" in window) {
        return window.matchMedia(query).matches
    }

    return false
}

/**
 * @author Mantine.js
 * @link https://github.com/mantinedev/mantine/blob/master/src/mantine-hooks/src/use-media-query/use-media-query.ts
 * @example
 * const matches = useMediaQuery('(min-width: 56.25em)')
 * @param query
 * @param initialValue
 * @param getInitialValueInEffect
 */
export function useMediaQuery(
    query: string,
    initialValue?: boolean,
    { getInitialValueInEffect }: UseMediaQueryOptions = {
        getInitialValueInEffect: true,
    }
) {
    const [matches, setMatches] = useState(
        getInitialValueInEffect ? initialValue : getInitialValue(query, initialValue)
    )
    const queryRef = useRef<MediaQueryList>()

    useEffect(() => {
        if ("matchMedia" in window) {
            queryRef.current = window.matchMedia(query)
            setMatches(queryRef.current.matches)
            return attachMediaListener(queryRef.current, (event) => setMatches(event.matches))
        }

        return undefined
    }, [query])

    return matches
}

/* -------------------------------------------------------------------------------------------------
 * useOutOfBounds
 * -----------------------------------------------------------------------------------------------*/


type Size = {
    width: number
    height: number
}

export function useOutOfBounds<T extends HTMLElement = HTMLDivElement>(): [
    (node: T | null) => void,
    { top: number, bottom: number, left: number, right: number },
    Size,
] {
    // Mutable values like 'ref.current' aren't valid dependencies
    // because mutating them doesn't re-render the component.
    // Instead, we use a state as a ref to be reactive.
    const [ref, setRef] = useState<T | null>(null)
    const [size, setSize] = useState<Size>({
        width: 0,
        height: 0,
    })
    const [outOfBounds, setOutOfBounds] = React.useState({
        top: 0,
        bottom: 0,
        left: 0,
        right: 0
    })

    const handleSize = useCallback(() => {

        const windowWidth = Math.min(document.documentElement.clientWidth, window.innerWidth)
        const windowHeight = Math.min(document.documentElement.clientHeight, window.innerHeight)

        const rect = ref?.getBoundingClientRect()

        if (rect && ref?.offsetHeight && ref.offsetWidth && ref?.offsetWidth > 0 && ref?.offsetHeight > 0) {
            let directions = {
                top: 0,
                bottom: 0,
                left: 0,
                right: 0
            }

            if (rect.top < 0) {
                directions.top = Math.abs(0 - rect.top)
            }

            if (rect.bottom > windowHeight) {
                directions.bottom = Math.abs(windowHeight - rect.bottom)
            }

            if (rect.left < 0) {
                directions.left = Math.abs(0 - rect.left)
            }

            if (rect.right > windowWidth) {
                directions.right = Math.abs(windowWidth - rect.right)
            }

            if (directions.top > 0 || directions.left > 0 || directions.bottom > 0 || directions.right > 0) {
                setOutOfBounds(directions)
            }
            // setOutOfBounds(prev => {
            //     if (prev.top !== directions.top || prev.right !== directions.right || prev.bottom !== directions.bottom || prev.left !== directions.left) {
            //         return directions
            //     }
            //     return prev
            // })
        }

        setSize({
            width: ref?.offsetWidth || 0,
            height: ref?.offsetHeight || 0,
        })

    }, [ref])

    useEventListener("resize", handleSize)
    useEventListener("keydown", handleSize)

    useIsomorphicLayoutEffect(() => {
        handleSize()
    }, [ref])

    return [setRef, outOfBounds, size]
}
