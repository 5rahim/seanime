import { useCallback, useState } from "react"
import { useEventListener, useIsomorphicLayoutEffect } from "../core"

export function useDataGridSize<T extends HTMLElement = HTMLDivElement>(): [
    (node: T | null) => void,
    { width: number, height: number },
] {
    const [ref, setRef] = useState<T | null>(null)
    const [size, setSize] = useState<{ width: number, height: number }>({
        width: 0,
        height: 0,
    })

    const handleSize = useCallback(() => {
        setSize({
            width: ref?.offsetWidth || 0,
            height: ref?.offsetHeight || 0,
        })

    }, [ref?.offsetHeight, ref?.offsetWidth])

    useEventListener("resize", handleSize)

    useIsomorphicLayoutEffect(() => {
        handleSize()
    }, [ref?.offsetHeight, ref?.offsetWidth])

    return [setRef, size]
}
