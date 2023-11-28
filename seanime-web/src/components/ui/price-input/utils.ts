interface SanitizeValueProps {
    value: string,
    groupSeparator?: string
    decimalSeparator?: string
    allowDecimals?: boolean,
    allowNegativeValue?: boolean
    decimalsLimit?: number
    disableAbbreviations?: boolean
    prefix?: string
    transformRawValue?: (raw: string) => string
}

/**
 * Remove group separator (eg: 1,000 -> 1000)
 */
export const removeSeparators = (value: string, separator = ","): string => {
    const reg = new RegExp(escapeRegExp(separator), "g")
    return value.replace(reg, "")
}
/**
 * Add group separator (eg: 2000 -> 2,000)
 */
export const addSeparators = (value: string, separator = ","): string => {
    return value.replace(/\B(?=(\d{3})+(?!\d))/g, separator)
}
/**
 * Remove prefix, separators and extra decimals from value
 * @author https://github.com/cchanxzy/react-currency-input-field/blob/master/src/components/utils/cleanValue.ts
 */
export const sanitizeValue = ({
                                  value,
                                  groupSeparator = ",",
                                  decimalSeparator = ".",
                                  allowDecimals = true,
                                  decimalsLimit = 2,
                                  allowNegativeValue = false,
                                  prefix = "",
                                  transformRawValue = (rawValue) => rawValue,
                              }: SanitizeValueProps): string => {
    const transformedValue = transformRawValue(value)

    if (transformedValue === "-") {
        return transformedValue
    }

    const reg = new RegExp(`((^|\\D)-\\d)|(-${escapeRegExp(prefix)})`)
    const isNegative = reg.test(transformedValue)

    // Is there a digit before the prefix? eg: 1$
    const [prefixWithValue, preValue] = RegExp(`(\\d+)-?${escapeRegExp(prefix)}`).exec(value) || []
    const withoutPrefix = prefix
        ? prefixWithValue
            ? transformedValue.replace(prefixWithValue, "").concat(preValue ?? "")
            : transformedValue.replace(prefix, "")
        : transformedValue
    const withoutSeparators = removeSeparators(withoutPrefix, groupSeparator)
    const withoutInvalidChars = removeNonNumericCharacters(withoutSeparators, [
        groupSeparator,
        decimalSeparator,
    ])

    let valueOnly = withoutInvalidChars

    const includeNegative = isNegative && allowNegativeValue ? "-" : ""

    if (decimalSeparator && valueOnly.includes(decimalSeparator)) {
        const [int, decimals] = withoutInvalidChars.split(decimalSeparator)
        const trimmedDecimals = decimalsLimit && decimals ? decimals.slice(0, decimalsLimit) : decimals
        const includeDecimals = allowDecimals ? `${decimalSeparator}${trimmedDecimals}` : ""

        return `${includeNegative}${int}${includeDecimals}`
    }

    return `${includeNegative}${valueOnly}`
}

/**
 * Remove incorrect characters
 * @param input
 * @param validChars
 */
export function removeNonNumericCharacters(input: string, validChars = [",", "."]): string {
    const chars = escapeRegExp(validChars.join(""))
    const reg = new RegExp(`[^\\d${chars}]`, "gi")
    return input.replace(reg, "")
}

/**
 * @author https://github.com/cchanxzy/react-currency-input-field/blob/master/src/components/utils/padTrimValue.ts
 * @param value
 * @param decimalSeparator
 * @param decimalScale
 */
export const padTrimValue = (
    value: string,
    decimalSeparator = ".",
    decimalScale: number = 2
): string => {
    if (decimalScale === undefined || value === "" || value === undefined) {
        return value
    }

    if (!value.match(/\d/g)) {
        return ""
    }

    const [int, decimals] = value.split(decimalSeparator)

    if (decimalScale === 0 && int) {
        return int
    }

    let newValue = decimals || ""

    if (newValue.length < decimalScale) {
        while (newValue.length < decimalScale) {
            newValue += "0"
        }
    } else {
        newValue = newValue.slice(0, decimalScale)
    }

    return `${int}${decimalSeparator}${newValue}`
}


/**
 * Set decimal separator to '.' so the string can be converted to an integer later on
 */
export const replaceDecimalSeparator = (
    value: string,
    decimalSeparator: string,
    isNegative: boolean = false
): string => {
    let newValue = value
    if (decimalSeparator && decimalSeparator !== ".") {
        newValue = newValue.replace(RegExp(escapeRegExp(decimalSeparator), "g"), ".")
        if (isNegative && decimalSeparator === "-") {
            newValue = `-${newValue.slice(1)}`
        }
    }
    return newValue
}

/* -------------------------------------------------------------------------------------------------
 * Escape RegExp
 * -----------------------------------------------------------------------------------------------*/

const reRegExpChar = /[\\^$.*+?()[\]{}|]/g
const reHasRegExpChar = RegExp(reRegExpChar.source)

/**
 * Escapes the `RegExp` special characters "^", "$", "\", ".", "*", "+",
 * "?", "(", ")", "[", "]", "{", "}", and "|" in `string`.
 *
 * @link https://github.com/lodash/lodash/blob/master/escapeRegExp.js
 * @param {string} [value=''] The string to escape.
 * @returns {string} Returns the escaped string.
 * @example
 *
 * escapeRegExp('[lodash](https://lodash.com/)')
 * // => '\[lodash\]\(https://lodash\.com/\)'
 */
export function escapeRegExp(value: string) {
    return (value && reHasRegExpChar.test(value))
        ? value.replace(reRegExpChar, "\\$&")
        : (value || "")
}
