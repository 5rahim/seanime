import * as React from "react"
import { ChartColor } from "./color-theme"

export type ChartValueFormatter = {
    (value: number): string
}

export type ChartCurveType = "linear" | "natural" | "step"

export type BaseChartProps = {
    /**
     * The data to be displayed in the chart.
     * An array of objects. Each object represents a data point.
     */
    data: any[] | null | undefined
    /**
     *  Data categories. Each string represents a key in a data object.
     *  e.g. ["Jan", "Feb", "Mar"]
     */
    categories: string[]
    /**
     * The key to map the data to the axis. It should match the key in the data object.
     * e.g. "value"
     */
    index: string
    /**
     * Color palette to be used in the chart.
     */
    colors?: ChartColor[]
    /**
     * Changes the text formatting for the y-axis values.
     */
    valueFormatter?: ChartValueFormatter
    /**
     * Show only the first and last elements in the x-axis. Great for smaller charts or sparklines.
     * @default false
     */
    startEndOnly?: boolean
    /**
     * Controls the visibility of the X axis.
     * @default true
     */
    showXAxis?: boolean
    /**
     * Controls the visibility of the Y axis.
     * @default true
     */
    showYAxis?: boolean
    /**
     * Controls width of the vertical axis.
     * @default 56
     */
    yAxisWidth?: number
    /**
     * Sets an animation to the chart when it is loaded.
     * @default true
     */
    showAnimation?: boolean
    /**
     * Controls the visibility of the tooltip.
     * @default true
     */
    showTooltip?: boolean
    /**
     * Controls the visibility of the legend.
     * @default true
     */
    showLegend?: boolean
    /**
     * Controls the visibility of the grid lines.
     * @default true
     */
    showGridLines?: boolean
    /**
     * Adjusts the minimum value in relation to the magnitude of the data.
     * @default false
     */
    autoMinValue?: boolean
    /**
     * Sets the minimum value of the shown chart data.
     */
    minValue?: number
    /**
     * Sets the maximum value of the shown chart data.
     */
    maxValue?: number
    /**
     * Controls if the ticks of a numeric axis are displayed as decimals or not.
     * @default true
     */
    allowDecimals?: boolean
    /**
     * Element to be displayed when there is no data.
     * @default `<></>`
     */
    emptyDisplay?: React.ReactElement
}
