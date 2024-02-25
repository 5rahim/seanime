"use client"

import { cva } from "class-variance-authority"
import * as React from "react"
import { ChartValueFormatter } from "../charts/types"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"
import { ChartColor } from "./color-theme"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const ChartTooltipAnatomy = defineStyleAnatomy({
    frame: cva([
        "UI-ChartTooltip__frame",
        "border bg-[--paper] p-2 rounded-[--radius]",
    ]),
    header: cva([
        "UI-ChartTooltip__header",
        "mb-2 font-semibold",
    ]),
    label: cva([
        "UI-ChartTooltip__label",
    ]),
    content: cva([
        "UI-ChartTooltip__content",
        "space-y-1",
    ]),
})

export const ChartTooltipRowAnatomy = defineStyleAnatomy({
    row: cva([
        "UI-ChartTooltip__row",
        "flex items-center justify-between space-x-8",
    ]),
    labelContainer: cva([
        "UI-ChartTooltip__labelContainer",
        "flex items-center space-x-2",
    ]),
    dot: cva([
        "UI-ChartTooltip__dot",
        "shrink-0",
        "h-3 w-3 bg-[--gray] rounded-full shadow-sm",
    ]),
    value: cva([
        "UI-ChartTooltip__value",
        "font-semibold tabular-nums text-right whitespace-nowrap",
    ]),
    label: cva([
        "UI-ChartTooltip__label",
        "text-sm text-right whitespace-nowrap font-medium text-[--foreground]",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * ChartTooltipFrame
 * -----------------------------------------------------------------------------------------------*/

export type ChartTooltipFrameProps = React.ComponentPropsWithoutRef<"div">

export const ChartTooltipFrame = ({ children, className }: ChartTooltipFrameProps) => (
    <div className={cn(ChartTooltipAnatomy.frame(), className)}>
        {children}
    </div>
)

/* -------------------------------------------------------------------------------------------------
 * ChartTooltipRow
 * -----------------------------------------------------------------------------------------------*/

export type ChartTooltipRowProps = ComponentAnatomy<typeof ChartTooltipRowAnatomy> & {
    value: string
    name: string
    color: ChartColor
}

export const ChartTooltipRow = (
    {
        value,
        name,
        color,
        dotClass,
        rowClass,
        valueClass,
        labelClass,
        labelContainerClass,
    }: ChartTooltipRowProps) => (
    <div className={cn(ChartTooltipRowAnatomy.row(), rowClass)}>
        <div className={cn(ChartTooltipRowAnatomy.labelContainer(), labelContainerClass)}>
            <span
                className={cn(ChartTooltipRowAnatomy.dot(), dotClass)}
                style={{ backgroundColor: `var(--${color})` }}
            />
            <p className={cn(ChartTooltipRowAnatomy.label(), labelClass)}>
                {name}
            </p>
        </div>
        <p className={cn(ChartTooltipRowAnatomy.value(), valueClass)}>
            {value}
        </p>
    </div>
)

/* -------------------------------------------------------------------------------------------------
 * ChartTooltip
 * -----------------------------------------------------------------------------------------------*/

export type ChartTooltipProps = ComponentAnatomy<typeof ChartTooltipAnatomy> & {
    active: boolean | undefined
    payload: any
    label: string
    categoryColors: Map<string, ChartColor>
    valueFormatter: ChartValueFormatter
}

export const ChartTooltip = (props: ChartTooltipProps) => {

    const {
        active,
        payload,
        label,
        categoryColors,
        valueFormatter,
        headerClass,
        contentClass,
        frameClass,
        labelClass,
    } = props
    if (active && payload) {
        return (
            <ChartTooltipFrame className={frameClass}>
                <div className={cn(ChartTooltipAnatomy.header(), headerClass)}>
                    <p className={cn(ChartTooltipAnatomy.label(), labelClass)}>
                        {label}
                    </p>
                </div>

                <div className={cn(ChartTooltipAnatomy.content(), contentClass)}>
                    {payload.map(({ value, name }: { value: number; name: string }, idx: number) => (
                        <ChartTooltipRow
                            key={`id-${idx}`}
                            value={valueFormatter(value)}
                            name={name}
                            color={categoryColors.get(name) ?? "brand"}
                        />
                    ))}
                </div>
            </ChartTooltipFrame>
        )
    }
    return null
}
