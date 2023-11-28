"use client"

import React, { useState } from "react"
import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva } from "class-variance-authority"
import {
    Bar,
    BarChart as ReChartsBarChart,
    CartesianGrid,
    Legend,
    ResponsiveContainer,
    Tooltip,
    XAxis,
    YAxis,
} from "recharts"
import { BaseChartProps } from "./types"
import { constructCategoryColors, defaultValueFormatter, getYAxisDomain } from "./utils"
import type { AxisDomain } from "recharts/types/util/types"
import { ChartTooltip } from "./chart-tooltip"
import { ColorPalette } from "../core/color-theme"
import { ChartLegend } from "./chart-legend"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const BarChartAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-BarChart__root",
    ])
})

/* -------------------------------------------------------------------------------------------------
 * BarChart
 * -----------------------------------------------------------------------------------------------*/

export interface BarChartProps extends React.ComponentPropsWithRef<"div">, ComponentWithAnatomy<typeof BarChartAnatomy>,
    BaseChartProps {
    layout?: "vertical" | "horizontal";
    stack?: boolean;
    relative?: boolean;
}

export const BarChart: React.FC<BarChartProps> = React.forwardRef<HTMLDivElement, BarChartProps>((props, ref) => {

    const {
        children,
        rootClassName,
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
        showGradient = true,
        autoMinValue = false,
        minValue,
        maxValue,
        allowDecimals = true,
        noDataText,
        ...rest
    } = props

    const [legendHeight, setLegendHeight] = useState(60)

    const categoryColors = constructCategoryColors(categories, colors)
    const yAxisDomain = getYAxisDomain(autoMinValue, minValue, maxValue)

    return (
        <div
            className={cn(BarChartAnatomy.root(), rootClassName, className)}
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
                            />
                        ) : null}

                        {layout !== "vertical" ? (
                            <XAxis
                                hide={!showXAxis}
                                dataKey={index}
                                interval="preserveStartEnd"
                                tick={{ transform: "translate(0, 6)" }} // Padding between labels and axis
                                ticks={startEndOnly ? [data[0][index], data[data.length - 1][index]] : undefined}
                                style={{
                                    fontSize: "12px",
                                    fontFamily: "Inter; Helvetica",
                                    marginTop: "20px",
                                }}
                                tickLine={false}
                                axisLine={false}
                            />
                        ) : (
                            <XAxis
                                hide={!showXAxis}
                                type="number"
                                tick={{ transform: "translate(-3, 0)" }}
                                domain={yAxisDomain as AxisDomain}
                                style={{
                                    fontSize: "12px",
                                    fontFamily: "Inter; Helvetica",
                                }}
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
                                style={{
                                    fontSize: "12px",
                                    fontFamily: "Inter; Helvetica",
                                }}
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
                                style={{
                                    fontSize: "12px",
                                    fontFamily: "Inter; Helvetica",
                                }}
                            />
                        )}
                        {showTooltip ? (
                            <Tooltip
                                wrapperStyle={{ outline: "none" }}
                                isAnimationActive={false}
                                cursor={{ fill: "#d1d5db", opacity: "0.15" }}
                                content={({ active, payload, label }) => (
                                    <ChartTooltip
                                        active={active}
                                        payload={payload}
                                        label={label}
                                        valueFormatter={valueFormatter}
                                        categoryColors={categoryColors}
                                    />
                                )}
                                position={{ y: 0 }}
                            />
                        ) : null}

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
                ) : (
                    <div>...</div>
                )}
            </ResponsiveContainer>
        </div>
    )

})

BarChart.displayName = "BarChart"
