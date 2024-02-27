import { cva } from "class-variance-authority"
import * as React from "react"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"
import { ChartColor, ColorPalette } from "./color-theme"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const LegendAnatomy = defineStyleAnatomy({
    legend: cva([
        "UI-Legend__legend",
        "flex flex-wrap overflow-hidden truncate",
    ]),
    legendItem: cva([
        "UI-Legend__legendItem",
        "inline-flex items-center truncate mr-4",
    ]),
    dot: cva([
        "UI-Legend__dot",
        "shrink-0",
        "flex-none h-3 w-3 bg-gray rounded-full shadow-sm mr-2",
    ]),
    label: cva([
        "UI-Legend__label",
        "whitespace-nowrap truncate text-sm font-medium text-[--muted]",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * LegendItem
 * -----------------------------------------------------------------------------------------------*/

export type LegendItemProps = {
    name: string
    color: ChartColor
    dotClass?: string
    labelClass?: string
    legendItemClass?: string
}

const LegendItem = ({ name, color, dotClass, legendItemClass, labelClass }: LegendItemProps) => (
    <li className={cn(LegendAnatomy.legendItem(), legendItemClass)}>
        <svg
            className={cn(LegendAnatomy.dot(), dotClass)}
            style={{ color: `var(--${color})` }}
            fill="currentColor"
            viewBox="0 0 8 8"
        >
            <circle cx={4} cy={4} r={4} />
        </svg>
        <p className={cn(LegendAnatomy.label(), labelClass)}>
            {name}
        </p>
    </li>
)

/* -------------------------------------------------------------------------------------------------
 * Legend
 * -----------------------------------------------------------------------------------------------*/

export type LegendProps = React.ComponentPropsWithRef<"ol"> & ComponentAnatomy<typeof LegendAnatomy> & {
    categories: string[]
    colors?: ChartColor[]
}

export const Legend = React.forwardRef<HTMLOListElement, LegendProps>((props, ref) => {
    const {
        categories,
        colors = ColorPalette,
        className,
        legendClass,
        legendItemClass,
        labelClass,
        dotClass,
        ...rest
    } = props
    return (
        <ol
            ref={ref}
            className={cn(
                LegendAnatomy.legend(),
                legendClass,
                className,
            )}
            {...rest}
        >
            {categories.map((category, idx) => (
                <LegendItem
                    key={`item-${idx}`}
                    name={category}
                    color={colors[idx] ?? "brand"}
                    dotClass={dotClass}
                    legendItemClass={legendItemClass}
                    labelClass={labelClass}
                />
            ))}
        </ol>
    )
})

Legend.displayName = "Legend"
