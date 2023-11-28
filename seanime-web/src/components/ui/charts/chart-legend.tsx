"use client"

import { useEffect, useRef, useState } from "react"
import { UIColor } from "../core/color-theme"
import { Legend } from "./legend"

/* -------------------------------------------------------------------------------------------------
 * ChartLegend
 * -----------------------------------------------------------------------------------------------*/

export const ChartLegend = (
    { payload }: any,
    categoryColors: Map<string, UIColor>,
    setLegendHeight: React.Dispatch<React.SetStateAction<number>>,
) => {
    const legendRef = useRef<HTMLDivElement>(null)

    const [windowSize, setWindowSize] = useState<undefined | number>(undefined)

    useEffect(() => {
        const handleResize = () => {
            setWindowSize(window.innerWidth)
            const calculateHeight = (height: number | undefined) =>
                height ?
                    Number(height) + 20 // 20px extra padding
                    : 60 // default height
            setLegendHeight(calculateHeight(legendRef.current?.clientHeight))
        }
        handleResize()
        window.addEventListener("resize", handleResize)

        return () => window.removeEventListener("resize", handleResize)
    }, [windowSize])

    return (
        <div ref={legendRef} className="flex w-full items-center justify-center mt-4">
            <Legend
                categories={payload.map((entry: any) => entry.value)}
                colors={payload.map((entry: any) => categoryColors.get(entry.value))}
            />
        </div>
    )
}

ChartLegend.displayName = "ChartLegend"
