import { ChartValueFormatter } from "../charts"
import { UIColor } from "../core"

/* -------------------------------------------------------------------------------------------------
 * Chart Utils
 * -----------------------------------------------------------------------------------------------*/

/**
 * @internal
 */
export const constructCategoryColors = (
    categories: string[],
    colors: UIColor[],
): Map<string, UIColor> => {
    const categoryColors = new Map<string, UIColor>()
    categories.forEach((category, idx) => {
        categoryColors.set(category, colors[idx] ?? "gray")
    })
    return categoryColors
}

/**
 * @internal
 */
export const getYAxisDomain = (
    autoMinValue: boolean,
    minValue: number | undefined,
    maxValue: number | undefined,
) => {
    const minDomain = autoMinValue ? "auto" : minValue ?? 0
    const maxDomain = maxValue ?? "auto"
    return [minDomain, maxDomain]
}

export const defaultValueFormatter: ChartValueFormatter = (value: number) => value.toString()

/* -------------------------------------------------------------------------------------------------
 * DonutChart Utils
 * -----------------------------------------------------------------------------------------------*/

export const parseChartData = (data: any[], colors: UIColor[]) =>
    data.map((dataPoint: any, idx: number) => {
        const baseColor = idx < colors.length ? colors[idx] : "brand"
        return {
            ...dataPoint,
            // explicitly adding color key if not present for tooltip coloring
            color: baseColor,
            fill: `var(--${baseColor})`, // Color
        }
    })

const sumNumericArray = (arr: number[]) =>
    arr.reduce((prefixSum, num) => prefixSum + num, 0)

const calculateDefaultLabel = (data: any[], category: string) =>
    sumNumericArray(data.map((dataPoint) => dataPoint[category]))

export const parseChartLabelInput = (
    labelInput: string | undefined,
    valueFormatter: ChartValueFormatter,
    data: any[],
    category: string,
) => (labelInput ? labelInput : valueFormatter(calculateDefaultLabel(data, category)))
