"use client"

import React, { useState } from "react"
import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva } from "class-variance-authority"
import {
    Area,
    AreaChart as ReChartsAreaChart,
    CartesianGrid,
    Legend,
    ResponsiveContainer,
    Tooltip,
    XAxis,
    YAxis,
} from "recharts"
import { BaseChartProps, ChartCurveType } from "./types"
import { constructCategoryColors, defaultValueFormatter, getYAxisDomain } from "./utils"
import type { AxisDomain } from "recharts/types/util/types"
import { ChartTooltip } from "./chart-tooltip"
import { ColorPalette } from "../core/color-theme"
import { ChartLegend } from "./chart-legend"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const AreaChartAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-AreaChart__root",
    ])
})

/* -------------------------------------------------------------------------------------------------
 * AreaChart
 * -----------------------------------------------------------------------------------------------*/

export interface AreaChartProps extends React.ComponentPropsWithRef<"div">,
    ComponentWithAnatomy<typeof AreaChartAnatomy>,
    BaseChartProps {
    stack?: boolean
    curveType?: ChartCurveType
    connectNulls?: boolean
    // Display dots for each data point
    showDots?: boolean
}

export const AreaChart: React.FC<AreaChartProps> = React.forwardRef<HTMLDivElement, AreaChartProps>((props, ref) => {

    const {
        rootClassName,
        className,
        stack = false,
        curveType = "linear",
        connectNulls = false,
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
        showDots = true,
        noDataText,
        ...rest
    } = props

    const [legendHeight, setLegendHeight] = useState(60)

    const categoryColors = constructCategoryColors(categories, colors)
    const yAxisDomain = getYAxisDomain(autoMinValue, minValue, maxValue)

    return (
        <div
            className={cn(AreaChartAnatomy.root(), rootClassName, className)}
            {...rest}
            ref={ref}
        >
            <ResponsiveContainer width={"100%"} height={"100%"}>
                {data?.length ? (
                    <ReChartsAreaChart data={data}>
                        {showGridLines ? (
                            <CartesianGrid strokeDasharray="3 3" horizontal={true} vertical={false}/>
                        ) : null}
                        <XAxis
                            hide={!showXAxis}
                            dataKey={index}
                            tick={{ transform: "translate(0, 8)" }}
                            ticks={startEndOnly ? [data[0][index], data[data.length - 1][index]] : undefined}
                            style={{
                                fontSize: ".75rem",
                                fontFamily: "Inter; Helvetica",
                                color: "red",
                            }}
                            interval="preserveStartEnd"
                            axisLine={false}
                            tickLine={false}
                            padding={{ left: 10, right: 10 }}
                            minTickGap={5}
                            spacing={120}
                            // textAnchor="end"
                            // angle={-40}
                        />
                        <YAxis
                            width={yAxisWidth}
                            hide={!showYAxis}
                            axisLine={false}
                            tickLine={false}
                            type="number"
                            domain={yAxisDomain as AxisDomain}
                            tick={{ transform: "translate(-3, 0)" }}
                            style={{
                                fontSize: ".8rem",
                                fontFamily: "Inter; Helvetica",
                            }}
                            tickFormatter={valueFormatter}
                            allowDecimals={allowDecimals}
                        />
                        {showTooltip ? (
                            <Tooltip
                                wrapperStyle={{ outline: "none" }}
                                isAnimationActive={false}
                                cursor={{ stroke: "#ddd", strokeWidth: 2 }}
                                position={{ y: 0 }}
                                content={({ active, payload, label }) => (
                                    <ChartTooltip
                                        active={active}
                                        payload={payload}
                                        label={label}
                                        valueFormatter={valueFormatter}
                                        categoryColors={categoryColors}
                                    />
                                )}
                            />
                        ) : null}

                        {categories.map((category) => {
                            const hexColor = `var(--${categoryColors.get(category)})`
                            return (
                                <defs key={category}>
                                    {showGradient ? (
                                        <linearGradient id={categoryColors.get(category)} x1="0" y1="0" x2="0" y2="1">
                                            <stop offset="5%" stopColor={hexColor} stopOpacity={0.2}/>
                                            <stop offset="95%" stopColor={hexColor} stopOpacity={0}/>
                                        </linearGradient>
                                    ) : (
                                        <linearGradient id={categoryColors.get(category)} x1="0" y1="0" x2="0" y2="1">
                                            <stop stopColor={hexColor} stopOpacity={0.3}/>
                                        </linearGradient>
                                    )}
                                </defs>
                            )
                        })}

                        {categories.map((category) => (
                            <Area
                                key={category}
                                name={category}
                                type={curveType}
                                dataKey={category}
                                stroke={`var(--${categoryColors.get(category)})`}
                                fill={`url(#${categoryColors.get(category)})`}
                                strokeWidth={2}
                                dot={showDots}
                                isAnimationActive={showAnimation}
                                stackId={stack ? "a" : undefined}
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

                    </ReChartsAreaChart>
                ) : (
                    <div>...</div>
                )}
            </ResponsiveContainer>
        </div>
    )

})

AreaChart.displayName = "AreaChart"
