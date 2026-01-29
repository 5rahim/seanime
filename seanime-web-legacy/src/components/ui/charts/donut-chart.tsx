"use client"

import * as React from "react"
import { Pie, PieChart as ReChartsDonutChart, ResponsiveContainer, Sector, Tooltip } from "recharts"
import { cn } from "../core/styling"
import { ChartTooltipFrame, ChartTooltipRow } from "./chart-tooltip"
import { ChartColor, ColorPalette } from "./color-theme"
import { ChartValueFormatter } from "./types"
import { defaultValueFormatter, parseChartData, parseChartLabelInput } from "./utils"

/* -------------------------------------------------------------------------------------------------
 * DonutChart
 * -----------------------------------------------------------------------------------------------*/

export type DonutChartProps = React.HTMLAttributes<HTMLDivElement> & {
    /**
     * The data to be displayed in the chart.
     * An array of objects. Each object represents a data point.
     */
    data: any[]
    /**
     * The key containing the quantitative chart values.
     */
    category: string
    /**
     * The key to map the data to the axis.
     * e.g. "value"
     */
    index: string
    /**
     * Color palette to be used in the chart.
     */
    colors?: ChartColor[]
    /**
     * The type of chart to display
     * @default "donut"
     */
    variant?: "donut" | "pie"
    /**
     * Changes the text formatting of the label.
     * This only works when the variant is "donut".
     */
    valueFormatter?: ChartValueFormatter
    /**
     * The text to be placed the center of the donut chart.
     * Only available when variant "donut".
     */
    label?: string
    /**
     * If true, the label will be displayed in the center of the chart
     * @default true
     */
    showLabel?: boolean
    /**
     * If true, the chart will animate when rendered
     */
    showAnimation?: boolean
    /**
     * If true, a tooltip will be displayed when hovering over a data point
     * @default true
     */
    showTooltip?: boolean
    /**
     * The element to be displayed when there is no data
     * @default <></>
     */
    emptyDisplay?: React.ReactElement
}

export const DonutChart = React.forwardRef<HTMLDivElement, DonutChartProps>((props, ref) => {
    const {
        data = [],
        category,
        index,
        colors = ColorPalette,
        variant = "donut",
        valueFormatter = defaultValueFormatter,
        label,
        showLabel = true,
        showAnimation = true,
        showTooltip = true,
        className,
        emptyDisplay = <></>,
        ...other
    } = props
    const isDonut = variant == "donut"

    const parsedLabelInput = parseChartLabelInput(label, valueFormatter, data, category)

    return (
        <div ref={ref} className={cn("w-full h-44", className)} {...other}>
            <ResponsiveContainer width="100%" height="100%">
                {data?.length ? (
                    <ReChartsDonutChart>
                        {showLabel && isDonut ? (
                            <text
                                x="50%"
                                y="50%"
                                textAnchor="middle"
                                dominantBaseline="middle"
                                className="fill-[--foreground] dark:fill-[--foreground] font-semibold"
                            >
                                {parsedLabelInput}
                            </text>
                        ) : null}
                        <Pie
                            data={parseChartData(data, colors)}
                            cx="50%"
                            cy="50%"
                            startAngle={90}
                            endAngle={-270}
                            innerRadius={isDonut ? "75%" : "0%"}
                            outerRadius="100%"
                            paddingAngle={0}
                            stroke=""
                            strokeLinejoin="round"
                            dataKey={category}
                            nameKey={index}
                            isAnimationActive={showAnimation}
                            inactiveShape={renderInactiveShape}
                            style={{ outline: "none" }}
                            className="stroke-[--background] dark:stroke-[--background]"
                        />
                        <Tooltip
                            cursorStyle={{ outline: "none" }}
                            wrapperStyle={{ outline: "none" }}
                            isAnimationActive={false}
                            content={showTooltip ? ({ active, payload }) => (
                                <DonutChartTooltip
                                    active={active}
                                    payload={payload}
                                    valueFormatter={valueFormatter}
                                />
                            ) : <></>}
                        />
                    </ReChartsDonutChart>
                ) : emptyDisplay}
            </ResponsiveContainer>
        </div>
    )
})

DonutChart.displayName = "DonutChart"

const renderInactiveShape = (props: any) => {
    const {
        cx,
        cy,
        // midAngle,
        innerRadius,
        outerRadius,
        startAngle,
        endAngle,
        // fill,
        // payload,
        // percent,
        // value,
        // activeIndex,
        className,
    } = props

    return (
        <g>
            <Sector
                cx={cx}
                cy={cy}
                innerRadius={innerRadius}
                outerRadius={outerRadius}
                startAngle={startAngle}
                endAngle={endAngle}
                className={className}
                fill=""
                opacity={0.3}
                style={{ outline: "none" }}
            />
        </g>
    )
}

/* -------------------------------------------------------------------------------------------------
 * DonutChartTooltip
 * -----------------------------------------------------------------------------------------------*/

type DonutChartTooltipProps = {
    active?: boolean
    payload: any
    valueFormatter: ChartValueFormatter
}

const DonutChartTooltip = ({ active, payload, valueFormatter }: DonutChartTooltipProps) => {
    if (active && payload[0]) {
        const payloadRow = payload[0]
        return (
            <ChartTooltipFrame>
                <div className={cn("py-2 px-2")}>
                    <ChartTooltipRow
                        value={valueFormatter(payloadRow.value)}
                        name={payloadRow.name}
                        color={payloadRow.payload.color}
                    />
                </div>
            </ChartTooltipFrame>
        )
    }
    return null
}

DonutChartTooltip.displayName = "DonutChartTooltip"
