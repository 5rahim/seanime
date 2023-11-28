export type ChartValueFormatter = {
    (value: number): string
}

export type ChartCurveType = "linear" | "natural" | "step"

export type ChartHorizontalPosition = "left" | "right"

export type ChartVerticalPosition = "top" | "bottom"

export interface BaseChartProps extends React.HTMLAttributes<HTMLDivElement> {
    data: any[] | null | undefined
    categories: string[]
    index: string
    // Choose the color for each category
    colors?: string[]
    // Change the formatting
    valueFormatter?: ChartValueFormatter
    // Show only the first and last elements in the x-axis. Great for smaller charts or sparklines.
    startEndOnly?: boolean
    showXAxis?: boolean
    showYAxis?: boolean
    yAxisWidth?: number
    // Sets an animation to the chart when it is loaded.
    showAnimation?: boolean
    showTooltip?: boolean
    showGradient?: boolean
    showLegend?: boolean
    showGridLines?: boolean
    // Adjusts the minimum value in relation to the magnitude of the data.
    autoMinValue?: boolean
    // Sets the minimum value of the shown chart data.
    minValue?: number
    // Sets the maximum value of the shown chart data.
    maxValue?: number
    // Controls if the ticks of a numeric axis are displayed as decimals or not.
    allowDecimals?: boolean
    // The displayed text when the data is empty.
    noDataText?: string
}
