"use client"

import * as React from "react"
import { Bar, BarChart as ReChartsBarChart, CartesianGrid, Legend, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts"
import type { AxisDomain } from "recharts/types/util/types"
import { cn } from "../core/styling"
import { ChartLegend } from "./chart-legend"
import { ChartTooltip } from "./chart-tooltip"
import { ColorPalette } from "./color-theme"
import { BaseChartProps } from "./types"
import { constructCategoryColors, defaultValueFormatter, getYAxisDomain } from "./utils"


/* -------------------------------------------------------------------------------------------------
 * BarChart
 * -----------------------------------------------------------------------------------------------*/

export type BarChartProps = React.ComponentPropsWithRef<"div"> & BaseChartProps & {
    /**
     * Display bars vertically or horizontally
     */
    layout?: "vertical" | "horizontal"
    /**
     * If true, the bars will be stacked
     */
    stack?: boolean
    /**
     * Display bars as a percentage of the total
     */
    relative?: boolean
    /**
     * Interval type for x-axis labels
     * @default "equidistantPreserveStart"
     */
    intervalType?: "preserveStart" | "preserveEnd" | "preserveStartEnd" | "equidistantPreserveStart"
}

export const BarChart = React.forwardRef<HTMLDivElement, BarChartProps>((props, ref) => {

    const {
        children,
        className,
        layout = "horizontal",
        stack = false,
        relative = false,
        /**/
        data = [],
        categories = [],
        index,
        colors = ColorPalette,
        valueFormatter = defaultValueFormatter,
        startEndOnly = false,
        showXAxis = true,
        showYAxis = true,
        yAxisWidth = 56,
        showAnimation = true,
        showTooltip = true,
        showLegend = true,
        showGridLines = true,
        autoMinValue = false,
        minValue,
        maxValue,
        allowDecimals = true,
        intervalType = "equidistantPreserveStart",
        emptyDisplay = <></>,
        ...rest
    } = props

    const [legendHeight, setLegendHeight] = React.useState(60)

    const categoryColors = constructCategoryColors(categories, colors)
    const yAxisDomain = getYAxisDomain(autoMinValue, minValue, maxValue)

    return (
        <div
            className={cn("w-full h-80", className)}
            {...rest}
            ref={ref}
        >
            <ResponsiveContainer width="100%" height="100%">
                {data?.length ? (
                    <ReChartsBarChart
                        data={data}
                        stackOffset={relative ? "expand" : "none"}
                        layout={layout === "vertical" ? "vertical" : "horizontal"}
                    >
                        {showGridLines ? (
                            <CartesianGrid
                                strokeDasharray="3 3"
                                horizontal={layout !== "vertical"}
                                vertical={layout === "vertical"}
                                className="stroke-gray-300 dark:stroke-gray-600"
                            />
                        ) : null}

                        {layout !== "vertical" ? (
                            <XAxis
                                hide={!showXAxis}
                                dataKey={index}
                                interval="preserveStartEnd"
                                tick={{ transform: "translate(0, 6)" }} // Padding between labels and axis
                                ticks={startEndOnly ? [data[0][index], data[data.length - 1][index]] : undefined}
                                className="font-medium text-[--muted] text-xs mt-4"
                                tickLine={false}
                                axisLine={false}
                            />
                        ) : (
                            <XAxis
                                hide={!showXAxis}
                                type="number"
                                tick={{ transform: "translate(-3, 0)" }}
                                domain={yAxisDomain as AxisDomain}
                                className="font-medium text-[--muted] text-xs"
                                tickLine={false}
                                axisLine={false}
                                tickFormatter={valueFormatter}
                                padding={{ left: 10, right: 10 }}
                                minTickGap={5}
                                allowDecimals={allowDecimals}
                            />
                        )}
                        {layout !== "vertical" ? (
                            <YAxis
                                width={yAxisWidth}
                                hide={!showYAxis}
                                axisLine={false}
                                tickLine={false}
                                type="number"
                                domain={yAxisDomain as AxisDomain}
                                tick={{ transform: "translate(-3, 0)" }}
                                className="font-medium text-[--muted] text-xs"
                                tickFormatter={
                                    relative ? (value: number) => `${(value * 100).toString()} %` : valueFormatter
                                }
                                allowDecimals={allowDecimals}
                            />
                        ) : (
                            <YAxis
                                width={yAxisWidth}
                                hide={!showYAxis}
                                dataKey={index}
                                axisLine={false}
                                tickLine={false}
                                ticks={startEndOnly ? [data[0][index], data[data.length - 1][index]] : undefined}
                                type="category"
                                interval="preserveStartEnd"
                                tick={{ transform: "translate(0, 6)" }}
                                className="font-medium text-[--muted] text-xs"
                            />
                        )}
                        <Tooltip
                            wrapperStyle={{
                                outline: "none",
                            }}
                            cursor={{
                                fill: "var(--gray)",
                                opacity: 0.05,
                            }}
                            isAnimationActive={false}
                            content={showTooltip ? ({ active, payload, label }) => (
                                <ChartTooltip
                                    active={active}
                                    payload={payload}
                                    label={label}
                                    valueFormatter={valueFormatter}
                                    categoryColors={categoryColors}
                                />
                            ) : <></>}
                            position={{ y: 0 }}
                        />

                        {categories.map((category) => (
                            <Bar
                                key={category}
                                name={category}
                                type="linear"
                                stackId={stack || relative ? "a" : undefined}
                                dataKey={category}
                                fill={`var(--${categoryColors.get(category)})`}
                                isAnimationActive={showAnimation}
                            />
                        ))}

                        {showLegend ? (
                            <Legend
                                verticalAlign="bottom"
                                height={legendHeight}
                                content={({ payload }) => ChartLegend({ payload }, categoryColors, setLegendHeight)}
                            />
                        ) : null}
                    </ReChartsBarChart>
                ) : emptyDisplay}
            </ResponsiveContainer>
        </div>
    )

})

BarChart.displayName = "BarChart"
