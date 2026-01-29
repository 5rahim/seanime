"use client"

import * as React from "react"
import CurrencyInputPrimitive from "react-currency-input-field"
import { BasicField, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"
import { cn } from "../core/styling"
import { extractInputPartProps, InputAddon, InputAnatomy, InputContainer, InputIcon, InputStyling } from "../input"

/* -------------------------------------------------------------------------------------------------
 * CurrencyInput
 * -----------------------------------------------------------------------------------------------*/

export type CurrencyInputIntlConfig = {
    /**
     * e.g. en-US
     */
    locale: string
    /**
     * e.g. USD
     */
    currency?: string
}

export type CurrentInputValues = {
    /**
     * Value as float or null if empty
     * e.g. "1.99" -> 1.99 | "" -> null
     */
    float: number | null
    /**
     * Value after applying formatting
     * e.g. "1000000" -> "1,000,0000"
     */
    formatted: string
    /**
     * Non formatted value as string
     */
    value: string
}

export type CurrencyInputProps =
    Omit<React.ComponentPropsWithoutRef<"input">, "size" | "disabled" | "defaultValue"> &
    InputStyling &
    BasicFieldOptions & {
    /**
     * Allow decimals
     * @default true
     */
    allowDecimals?: boolean
    /**
     * Allow user to enter negative value
     * @default true
     */
    allowNegativeValue?: boolean
    /**
     * Maximum characters the user can enter
     */
    maxLength?: number
    /**
     * Limit length of decimals allowed
     * @default 2
     */
    decimalsLimit?: number
    /**
     * Specify decimal scale for padding/trimming
     * e.g. 1.5 -> 1.50 | 1.234 -> 1.23
     */
    decimalScale?: number
    /**
     * Default value if uncontrolled
     */
    defaultValue?: number | string
    /**
     * Value will always have the specified length of decimals
     * e.g. 123 -> 1.23
     * Note: This formatting only happens onBlur
     */
    fixedDecimalLength?: number
    /**
     * Placeholder if there is no value
     */
    placeholder?: string
    /**
     * Include a prefix
     * e.g. £
     */
    prefix?: string
    /**
     * Include a suffix
     * e.g. €
     */
    suffix?: string
    /**
     * Incremental value change on arrow down and arrow up key press
     */
    step?: number
    /**
     * Separator between integer part and fractional part of value.
     */
    decimalSeparator?: string
    /**
     * Separator between thousand, million and billion.
     */
    groupSeparator?: string
    /**
     * Disable auto adding separator between values
     * e.g. 1000 -> 1,000
     * @default false
     */
    disableGroupSeparators?: boolean
    /**
     * Disable abbreviations (m, k, b)
     * @default false
     */
    disableAbbreviations?: boolean
    /**
     * International locale config
     * e.g. { locale: 'ja-JP', currency: 'JPY' }
     * Any prefix, groupSeparator or decimalSeparator options passed in will override Intl Locale config
     */
    intlConfig?: CurrencyInputIntlConfig
    /**
     * Transform the raw value form the input before parsing
     */
    transformRawValue?: (rawValue: string) => string
    /**
     * Callback invoked when value changes
     */
    onValueChange?: (value: (string | undefined), values?: CurrentInputValues) => void
}

export const CurrencyInput = React.forwardRef<HTMLInputElement, CurrencyInputProps>((props, ref) => {

    const [props1, basicFieldProps] = extractBasicFieldProps<CurrencyInputProps>(props, React.useId())

    const [{
        size,
        intent,
        leftAddon,
        leftIcon,
        rightAddon,
        rightIcon,
        className,
        /**/
        value,
        onValueChange,
        transformRawValue,
        intlConfig,
        allowDecimals,
        allowNegativeValue,
        decimalsLimit,
        decimalScale,
        disabled,
        fixedDecimalLength,
        placeholder,
        prefix,
        suffix,
        step,
        decimalSeparator,
        groupSeparator,
        disableGroupSeparators,
        disableAbbreviations,
        defaultValue,
        ...rest
    }, {
        inputContainerProps,
        leftAddonProps,
        leftIconProps,
        rightAddonProps,
        rightIconProps,
    }] = extractInputPartProps<CurrencyInputProps>({
        ...props1,
        size: props1.size ?? "md",
        intent: props1.intent ?? "basic",
        leftAddon: props1.leftAddon,
        leftIcon: props1.leftIcon,
        rightAddon: props1.rightAddon,
        rightIcon: props1.rightIcon,
    })

    return (
        <BasicField{...basicFieldProps}>
            <InputContainer {...inputContainerProps}>
                <InputAddon {...leftAddonProps} />
                <InputIcon {...leftIconProps} />

                <CurrencyInputPrimitive
                    ref={ref}
                    id={basicFieldProps.id}
                    name={basicFieldProps.name}
                    defaultValue={defaultValue}
                    className={cn(
                        "form-input",
                        InputAnatomy.root({
                            size,
                            intent,
                            hasError: !!basicFieldProps.error,
                            isDisabled: !!basicFieldProps.disabled,
                            isReadonly: !!basicFieldProps.readonly,
                            hasRightAddon: !!rightAddon,
                            hasRightIcon: !!rightIcon,
                            hasLeftAddon: !!leftAddon,
                            hasLeftIcon: !!leftIcon,
                        }),
                        className,
                    )}
                    disabled={basicFieldProps.disabled || basicFieldProps.readonly}
                    data-disabled={basicFieldProps.disabled}
                    required={basicFieldProps.required}
                    value={value}
                    onValueChange={(value, _, values) => onValueChange?.(value, values)}
                    transformRawValue={transformRawValue}
                    intlConfig={intlConfig}
                    allowDecimals={allowDecimals}
                    allowNegativeValue={allowNegativeValue}
                    decimalsLimit={decimalsLimit}
                    decimalScale={decimalScale}
                    fixedDecimalLength={fixedDecimalLength}
                    placeholder={placeholder}
                    prefix={prefix}
                    suffix={suffix}
                    step={step}
                    decimalSeparator={decimalSeparator}
                    groupSeparator={groupSeparator}
                    disableGroupSeparators={disableGroupSeparators}
                    disableAbbreviations={disableAbbreviations}
                    {...rest}
                />

                <InputAddon {...rightAddonProps} />
                <InputIcon {...rightIconProps} />
            </InputContainer>
        </BasicField>
    )

})

CurrencyInput.displayName = "CurrencyInput"
