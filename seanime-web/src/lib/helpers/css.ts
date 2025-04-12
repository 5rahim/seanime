const PIXELS_PER_INCH = 96
const MILLIMETRES_PER_INCH = 25.4
const POINTS_PER_INCH = 72
const PICAS_PER_INCH = 6

export function getStyle(
    element: HTMLElement,
    property: keyof CSSStyleDeclaration,
): string {
    const view = element.ownerDocument?.defaultView || window
    const style = view.getComputedStyle(element)
    return (
        style.getPropertyValue(property as string) || (style[property] as string)
    )
}

function fontSize(element?: HTMLElement | null): string {
    return element
        ? getStyle(element, "fontSize") || fontSize(element.parentElement)
        : getStyle(window.document.documentElement, "fontSize")
}

function parse(providedLength?: string | null): [number, string] {
    const length = providedLength || "0"

    // Check if it's a calc expression
    if (length.trim().startsWith("calc(")) {
        return parseCalc(length)
    }

    const value = Number.parseFloat(length)
    const match = length.match(/[\d-.]+(\w+)$/)
    const unit = match?.[1] ?? ""
    return [value, unit.toLowerCase()]
}

function parseCalc(calcExpression: string): [number, string] {
    // For calc expressions, we'll return a placeholder value and unit
    // The actual calculation will be done in evaluateCalcExpression
    return [0, "calc"]
}

function evaluateCalcExpression(calcExpression: string, element?: HTMLElement | null): number {
    // Extract the content inside calc()
    const content = calcExpression.replace(/^calc\(\s*/, "").replace(/\s*\)$/, "")

    // Replace all CSS length values with their pixel equivalents
    const pixelExpression = content.replace(/(-?[\d.]+[a-z%]+)/g, (match) => {
        // Skip if it's already a number without unit
        if (!isNaN(Number(match))) return match

        return getPixelsFromLength(match, element).toString()
    })

    try {
        // Safely evaluate the mathematical expression with all units converted to pixels
        // First normalize the expression to handle CSS math operators
        const normalizedExpression = pixelExpression
            .replace(/\s+/g, " ")           // Normalize whitespace
            .replace(/\s*([+\-*/()])\s*/g, "$1") // Remove spaces around operators

        return Function(`'use strict'; return (${normalizedExpression})`)()
    }
    catch (error) {
        console.error("Error evaluating calc() expression:", error)
        return 0
    }
}

export function getPixelsFromLength(length: string, element?: HTMLElement | null): number {
    // If the length is a calc expression, we need to evaluate each part
    if (length.trim().startsWith("calc(")) {
        return evaluateCalcExpression(length, element)
    }

    const view = element?.ownerDocument?.defaultView ?? window
    const root = view.document.documentElement || view.document.body

    const [value, unit] = parse(length)

    switch (unit) {
        case "rem":
            return value * getPixelsFromLength(fontSize(window.document.documentElement))

        case "em":
            return value * getPixelsFromLength(fontSize(element), element?.parentElement)

        case "in":
            return value * PIXELS_PER_INCH

        case "q":
            return (value * PIXELS_PER_INCH) / MILLIMETRES_PER_INCH / 4

        case "mm":
            return (value * PIXELS_PER_INCH) / MILLIMETRES_PER_INCH

        case "cm":
            return (value * PIXELS_PER_INCH * 10) / MILLIMETRES_PER_INCH

        case "pt":
            return (value * PIXELS_PER_INCH) / POINTS_PER_INCH

        case "pc":
            return (value * PIXELS_PER_INCH) / PICAS_PER_INCH

        case "vh":
            return (value * view.innerHeight || root.clientWidth) / 100

        case "vw":
            return (value * view.innerWidth || root.clientHeight) / 100

        case "vmin":
            return (
                (value *
                    Math.min(
                        view.innerWidth || root.clientWidth,
                        view.innerHeight || root.clientHeight,
                    )) /
                100
            )

        case "vmax":
            return (
                (value *
                    Math.max(
                        view.innerWidth || root.clientWidth,
                        view.innerHeight || root.clientHeight,
                    )) /
                100
            )

        default:
            return value
    }
}
