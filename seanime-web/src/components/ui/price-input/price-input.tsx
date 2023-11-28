"use client"

import Dinero, { Currency } from "dinero.js"
import React, { ChangeEvent, useCallback, useId, useMemo, useState } from "react"
import { extractBasicFieldProps } from "../basic-field"
import { useUILocaleConfig } from "../core"
import { TextInput, TextInputProps } from "../text-input"
import { currencies } from "./currencies"
import { padTrimValue, removeNonNumericCharacters, replaceDecimalSeparator, sanitizeValue } from "./utils"
import localeConfig from "./locales"

/* -------------------------------------------------------------------------------------------------
 * PriceInput
 * -----------------------------------------------------------------------------------------------*/

export interface PriceInputProps extends Omit<TextInputProps, "value" | "onChange" | "defaultValue"> {
    value?: number
    defaultValue?: number
    onChange?: (value: number) => void
    locale?: string
    currency?: Currency
}

export const PriceInput = React.forwardRef<HTMLInputElement, PriceInputProps>((props, ref) => {

    const [{
        value,
        defaultValue = 0,
        locale,
        currency,
        onChange,
        ...rest
    }, basicFieldProps] = extractBasicFieldProps<PriceInputProps>(props, useId())

    const { locale: lng, country } = useUILocaleConfig()

    // 1. Get language and currency
    const _locale = locale ?? lng
    const _currency = currency ?? currencies[country]
    const config = useMemo(() => localeConfig[_locale] ?? { decimalSeparator: ".", groupSeparator: "," }, [_locale])

    const _decimalSeparator = config.decimalSeparator
    const _groupSeparator = config.groupSeparator

    // /!\ Change dinero.toFormat() options if you change this
    const _decimalSpace = 2

    const _multiplier = Math.pow(10, _decimalSpace) // eg: decimalSpace = 2 => 100

    // 2. Track the amount (int)
    const [amount, setAmount] = useState<number>(value ?? defaultValue)
    // 3. Track editing state
    const [isEditing, setIsEditing] = useState(false)
    // 4. Dinero object depends on amount
    const dineroObject = Dinero({ amount: amount, currency: _currency, precision: _decimalSpace }).setLocale(_locale)
    // 5. Get formatted value (string) from dinero object
    const formattedValue = dineroObject.toFormat()
    // 6. Track user input (what the user sees), the initial state is formatted
    const [inputValue, setInputValue] = useState(formatNumber(dineroObject.toUnit().toString(), _locale, _decimalSpace))


    const toFloat = useCallback((value: string) => {
        // 1. We remove prefixes, group separators and extra decimals (keep local decimal separator) (eg: 4,555.999 -> 4555.99)
        let _sanitizedValue = sanitizeValue({
            value: value,
            groupSeparator: _groupSeparator,
            decimalSeparator: _decimalSeparator,
            decimalsLimit: _decimalSpace
        })
        // 2. Convert local decimal to '.' if needed, in order to parse it (eg: fr2,5 -> 2.5)
        let _valueWithCorrectDecimal = _decimalSeparator !== "." ? replaceDecimalSeparator(_sanitizedValue, _decimalSeparator, false) : _sanitizedValue
        // 3. Keep decimal space before parsing to float (eg: 2.5 -> 2.50)
        return parseFloat(padTrimValue(_valueWithCorrectDecimal, ".", _decimalSpace))
    }, [_decimalSeparator, _groupSeparator])

    function handleOnChange(event: ChangeEvent<HTMLInputElement>) {
        let _amount = 0
        let _value = ""
        try {
            _value = removeNonNumericCharacters(event.target.value ?? "0")
            // Convert the value entered to a float
            if (_value.length > 0) {
                _amount = toFloat(_value)
            }
        } catch (e) {
            setInputValue("0")
            setAmount(_amount ?? 0)
            onChange && onChange(_amount ?? 0)
        }
        // Maintain the precision (eg: precision 2 => 400 -> 40000)
        const _fixed = parseInt((_amount * _multiplier).toFixed(_decimalSpace))
        setAmount(_fixed) // Update dinero object (#4)
        onChange && onChange(_fixed)
        setInputValue(_value) // Update displayed input (#6)
    }


    return (
        <>
            <TextInput
                value={isEditing ? inputValue : formattedValue}
                onChange={handleOnChange}
                onBlur={() => {
                    setInputValue(v => formatNumber(dineroObject.toUnit().toString(), _locale, _decimalSpace))
                    setIsEditing(false)
                }}
                onFocus={() => {
                    setIsEditing(true)
                }}
                ref={ref}
                {...basicFieldProps}
                {...rest}
            />
        </>
    )

})

/* -------------------------------------------------------------------------------------------------
 * Helper functions
 * -----------------------------------------------------------------------------------------------*/

/**
 * @param input
 * @param lang
 * @param decimalSpace
 */
function formatNumber(input: string | undefined, lang: string, decimalSpace: number): string {
    // Parse the input string to a number
    let inputAsNumber = parseFloat(input ?? "0")
    if (isNaN(inputAsNumber)) {
        // If the input is not a valid number, return an empty string
        return "0"
    }
    // Use the Intl object to format the number with 2 decimal places
    const res = new Intl.NumberFormat(lang, {
        minimumFractionDigits: decimalSpace,
        maximumFractionDigits: decimalSpace,
    }).format(inputAsNumber)

    return res
}

PriceInput.displayName = "PriceInput"
