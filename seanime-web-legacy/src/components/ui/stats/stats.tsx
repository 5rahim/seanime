import { cva, VariantProps } from "class-variance-authority"
import * as React from "react"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const StatsAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Stats__root",
        "grid grid-cols-1 divide-y divide-[--border] overflow-hidden md:grid-cols-3 md:divide-y-0 md:divide-x",
    ], {
        variants: {
            size: {
                sm: null, md: null, lg: null,
            },
        },
        defaultVariants: {
            size: "md",
        },
    }),
    item: cva([
        "UI-Stats__item",
        "relative",
    ], {
        variants: {
            size: {
                sm: "p-3 sm:p-4",
                md: "p-4 sm:p-6",
                lg: "p-4 sm:p-7",
            },
        },
    }),
    name: cva([
        "UI-Stats__name",
        "text-sm font-normal text-[--muted]",
    ], {
        variants: {
            size: {
                sm: "text-xs",
                md: "text-sm",
                lg: "text-base",
            },
        },
    }),
    value: cva([
        "UI-Stats__value",
        "mt-1 flex items-baseline md:block lg:flex font-semibold",
    ], {
        variants: {
            size: {
                sm: "text-xl md:text-2xl",
                md: "text-2xl md:text-3xl",
                lg: "text-3xl md:text-4xl",
            },
        },
    }),
    unit: cva([
        "UI-Stats__unit",
        "ml-2 text-sm font-medium text-[--muted]",
    ]),
    trend: cva([
        "UI-Stats__trend",
        "inline-flex items-baseline text-sm font-medium",
        "data-[trend=up]:text-[--green] data-[trend=down]:text-[--red]",
    ]),
    icon: cva([
        "UI-Stats__icon",
        "absolute top-5 right-5 opacity-30",
    ], {
        variants: {
            size: {
                sm: "text-xl sm:text-2xl",
                md: "text-2xl sm:text-3xl",
                lg: "text-3xl sm:text-4xl",
            },
        },
    }),
})

/* -------------------------------------------------------------------------------------------------
 * Stats
 * -----------------------------------------------------------------------------------------------*/

export type StatsItem = {
    name: string,
    value: string | number,
    unit?: string | number,
    change?: string | number,
    trend?: "up" | "down",
    icon?: React.ReactElement
}

export type StatsProps = React.ComponentPropsWithRef<"dl"> &
    ComponentAnatomy<typeof StatsAnatomy> &
    VariantProps<typeof StatsAnatomy.root> & {
    children?: React.ReactNode,
    items: StatsItem[]
}

export const Stats = React.forwardRef<HTMLDListElement, StatsProps>((props, ref) => {

    const {
        children,
        itemClass,
        nameClass,
        valueClass,
        unitClass,
        trendClass,
        iconClass,
        className,
        items,
        size = "md",
        ...rest
    } = props

    return (
        <dl
            ref={ref}
            className={cn(StatsAnatomy.root({ size }), className)}
            {...rest}
        >
            {items.map((item) => (
                <div key={item.name} className={cn(StatsAnatomy.item({ size }), itemClass)}>

                    <dt className={cn(StatsAnatomy.name({ size }), nameClass)}>{item.name}</dt>

                    <dd className={cn(StatsAnatomy.value({ size }), valueClass)}>
                        {item.value}
                        <span className={cn(StatsAnatomy.unit(), unitClass)}>{item.unit}</span>
                    </dd>

                    {(!!item.change || !!item.trend) &&
                        <div
                            className={cn(StatsAnatomy.trend(), trendClass)}
                            data-trend={item.trend}
                        >
                            {item.trend && <span> {item.trend === "up" ? "+" : "-"}</span>}
                            {item.change}
                        </div>
                    }

                    <div className={cn(StatsAnatomy.icon({ size }), iconClass)}>
                        {item.icon}
                    </div>

                </div>
            ))}
        </dl>
    )

})

Stats.displayName = "Stats"
