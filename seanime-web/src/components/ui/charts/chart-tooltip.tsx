"use client"

import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { ChartValueFormatter } from "../charts/types"
import { cva } from "class-variance-authority"
import { UIColor } from "../core/color-theme"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const ChartTooltipAnatomy = defineStyleAnatomy({
    frame: cva([
        "UI-ChartTooltip__frame",
        "border border-[--border] bg-[--paper] p-2 rounded-[--radius]"
    ]),
    header: cva([
        "UI-ChartTooltip__header",
        "mb-2 font-semibold"
    ]),
    label: cva([
        "UI-ChartTooltip__label",
    ]),
    content: cva([
        "UI-ChartTooltip__content",
        "space-y-1"
    ]),
})

export const ChartTooltipRowAnatomy = defineStyleAnatomy({
    row: cva([
        "UI-ChartTooltip__row",
        "flex items-center justify-between space-x-8"
    ]),
    labelContainer: cva([
        "UI-ChartTooltip__labelContainer",
        "flex items-center space-x-2"
    ]),
    dot: cva([
        "UI-ChartTooltip__dot",
        "shrink-0",
        "h-3 w-3 bg-gray rounded-full shadow-sm"
    ]),
    value: cva([
        "UI-ChartTooltip__value",
        "font-semibold tabular-nums text-right whitespace-nowrap",
    ]),
    label: cva([
        "UI-ChartTooltip__label",
        "text-sm text-right whitespace-nowrap font-medium text-[--muted]"
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * ChartTooltipFrame
 * -----------------------------------------------------------------------------------------------*/

interface ChartTooltipFrameProps extends React.ComponentPropsWithoutRef<"div"> {
}

export const ChartTooltipFrame = ({ children, className }: ChartTooltipFrameProps) => (
    <div
        className={cn(ChartTooltipAnatomy.frame(), className)}
    >
        {children}
    </div>
)

/* -------------------------------------------------------------------------------------------------
 * ChartTooltipRow
 * -----------------------------------------------------------------------------------------------*/

export interface ChartTooltipRowProps extends ComponentWithAnatomy<typeof ChartTooltipRowAnatomy> {
    value: string;
    name: string;
    color: UIColor;
}

export const ChartTooltipRow = (
    {
        value,
        name,
        color,
        dotClassName,
        rowClassName,
        valueClassName,
        labelClassName,
        labelContainerClassName
    }: ChartTooltipRowProps) => (
    <div className={cn(ChartTooltipRowAnatomy.row(), rowClassName)}>
        <div className={cn(ChartTooltipRowAnatomy.labelContainer(), labelContainerClassName)}>
            <span
                className={cn(
                    ChartTooltipRowAnatomy.dot(),
                    dotClassName
                )}
                style={{ backgroundColor: `var(--${color})` }}
            />
            <p
                className={cn(
                    ChartTooltipRowAnatomy.label(),
                    labelClassName,
                )}
            >
                {name}
            </p>
        </div>
        <p
            className={cn(
                ChartTooltipRowAnatomy.value(),
                valueClassName
            )}
        >
            {value}
        </p>
    </div>
)

/* -------------------------------------------------------------------------------------------------
 * ChartTooltip
 * -----------------------------------------------------------------------------------------------*/

export interface ChartTooltipProps extends ComponentWithAnatomy<typeof ChartTooltipAnatomy> {
    active: boolean | undefined;
    payload: any;
    label: string;
    categoryColors: Map<string, UIColor>;
    valueFormatter: ChartValueFormatter;
}

export const ChartTooltip = (props: ChartTooltipProps) => {

    const {
        active,
        payload,
        label,
        categoryColors,
        valueFormatter,
        headerClassName,
        contentClassName,
        frameClassName,
        labelClassName,
    } = props
    if (active && payload) {
        return (
            <ChartTooltipFrame className={frameClassName}>
                <div
                    className={cn(
                        ChartTooltipAnatomy.header(),
                        headerClassName,
                    )}
                >
                    <p
                        className={cn(
                            ChartTooltipAnatomy.label(),
                            labelClassName,
                        )}
                    >
                        {label}
                    </p>
                </div>

                <div className={cn(
                    ChartTooltipAnatomy.content(),
                    contentClassName
                )}>
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
