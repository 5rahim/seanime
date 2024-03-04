"use client"

import * as React from "react"
import { CartesianGrid, Legend, Line, LineChart as ReChartsLineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts"
import type { AxisDomain } from "recharts/types/util/types"
import { cn } from "../core/styling"
import { ChartLegend } from "./chart-legend"
import { ChartTooltip } from "./chart-tooltip"
import { ColorPalette } from "./color-theme"
import { BaseChartProps, ChartCurveType } from "./types"
import { constructCategoryColors, defaultValueFormatter, getYAxisDomain } from "./utils"

/* -------------------------------------------------------------------------------------------------
 * LineChart
 * -----------------------------------------------------------------------------------------------*/

export type LineChartProps = React.ComponentPropsWithRef<"div"> & BaseChartProps & {
    /**
     * The type of curve to use for the line
     * @default "linear"
     */
    curveType?: ChartCurveType
    /**
     * Connect null data points
     * @default false
     */
    connectNulls?: boolean
    /**
     * Angle the x-axis labels
     * @default false
     */
    angledLabels?: boolean
    /**
     * Interval type for x-axis labels
     * @default "preserveStartEnd"
     */
    intervalType?: "preserveStart" | "preserveEnd" | "preserveStartEnd" | "equidistantPreserveStart"
}


export const LineChart = React.forwardRef<HTMLDivElement, LineChartProps>((props, ref) => {

    const {
        className,
        curveType = "linear",
        connectNulls = false,
        angledLabels,
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
        intervalType = "preserveStartEnd",
        emptyDisplay = <></>,
        ...rest
    } = props

    const [legendHeight, setLegendHeight] = React.useState(60)

    const categoryColors = constructCategoryColors(categories, colors)
    const yAxisDomain = getYAxisDomain(autoMinValue, minValue, maxValue)

    return (
        <div
            ref={ref}
            className={cn("w-full h-80", className)}
            {...rest}
        >
            <ResponsiveContainer width="100%" height="100%">
                {data?.length ? (
                    <ReChartsLineChart data={data}>
                        {showGridLines ? (
                            <CartesianGrid
                                strokeDasharray="3 3"
                                horizontal={true}
                                vertical={false}
                                className="stroke-gray-300 dark:stroke-gray-600"
                            />
                        ) : null}
                        <XAxis
                            hide={!showXAxis}
                            dataKey={index}
                            tick={{ transform: "translate(0, 8)" }}
                            ticks={startEndOnly ? [data[0][index], data[data.length - 1][index]] : undefined}
                            className="font-medium text-[--muted] text-xs"
                            interval={intervalType}
                            axisLine={false}
                            tickLine={false}
                            padding={{ left: 10, right: 10 }}
                            minTickGap={5}
                            textAnchor={angledLabels ? "end" : "middle"}
                            angle={angledLabels ? -40 : undefined}
                        />
                        <YAxis
                            width={yAxisWidth}
                            hide={!showYAxis}
                            axisLine={false}
                            tickLine={false}
                            type="number"
                            textAnchor="end"
                            domain={yAxisDomain as AxisDomain}
                            tick={{ transform: "translate(-3, 0)" }}
                            className="font-medium text-[--muted] text-xs"
                            tickFormatter={valueFormatter}
                            allowDecimals={allowDecimals}
                        />
                        <Tooltip
                            wrapperStyle={{ outline: "none" }}
                            isAnimationActive={false}
                            cursor={{ stroke: "var(--gray)", strokeWidth: 1 }}
                            position={{ y: 0 }}
                            content={showTooltip ? ({ active, payload, label }) => (
                                <ChartTooltip
                                    active={active}
                                    payload={payload}
                                    label={label}
                                    valueFormatter={valueFormatter}
                                    categoryColors={categoryColors}
                                />
                            ) : <></>}
                        />

                        {categories.map((category) => (
                            <Line
                                key={category}
                                name={category}
                                type={curveType}
                                dataKey={category}
                                stroke={`var(--${categoryColors.get(category)})`}
                                strokeWidth={2}
                                dot={false}
                                isAnimationActive={showAnimation}
                                connectNulls={connectNulls}
                            />
                        ))}

                        {showLegend ? (
                            <Legend
                                verticalAlign="bottom"
                                height={legendHeight}
                                content={({ payload }) => ChartLegend({ payload }, categoryColors, setLegendHeight)}
                            />
                        ) : null}

                    </ReChartsLineChart>
                ) : emptyDisplay}
            </ResponsiveContainer>
        </div>
    )

})

LineChart.displayName = "LineChart"
