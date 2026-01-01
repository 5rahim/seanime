import { useIsomorphicLayoutEffect } from "@/components/ui/core/hooks"
import { RefObject, useState } from "react"

export interface ElementMeasurements {
    width: number
    height: number
    top: number
    left: number
    bottom: number
    right: number
    x: number
    y: number
}

export function useMeasureElement(ref: RefObject<HTMLElement>) {
    const [measurements, setMeasurements] = useState<ElementMeasurements>({
        width: 0,
        height: 0,
        top: 0,
        left: 0,
        bottom: 0,
        right: 0,
        x: 0,
        y: 0,
    })

    useIsomorphicLayoutEffect(() => {
        const element = ref.current
        if (!element) return

        const resizeObserver = new ResizeObserver((entries) => {
            for (const entry of entries) {
                const rect = entry.target.getBoundingClientRect()
                setMeasurements({
                    width: rect.width,
                    height: rect.height,
                    top: rect.top,
                    left: rect.left,
                    bottom: rect.bottom,
                    right: rect.right,
                    x: rect.x,
                    y: rect.y,
                })
            }
        })

        resizeObserver.observe(element)

        return () => {
            resizeObserver.disconnect()
        }
    }, [ref])

    return measurements
}

