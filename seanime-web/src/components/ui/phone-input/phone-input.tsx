"use client"

import { cva } from "class-variance-authority"
import { CountryCode, E164Number, parsePhoneNumber } from "libphonenumber-js"
import * as React from "react"
import PhoneInputPrimitive, { Country } from "react-phone-number-input"
import { BasicField, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"
import { extractInputPartProps, InputAddon, InputAnatomy, InputContainer, InputIcon, InputStyling } from "../input"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

const PhoneInputAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-PhoneInput__root",
        "rounded-l-none z-[2]",
    ]),
    container: cva([
        "UI-PhoneInput__container",
        "relative flex items-center w-full",
    ]),
    countrySelect: cva([
        "UI-PhoneInput__countrySelect",
        "w-[3rem] z-[3] relative flex-none cursor-pointer truncate rounded-r-none border-r-transparent opacity-0",
        "focus-visible:opacity-100 transition duration-200 ease-in-out",
    ], {
        variants: {
            hasLeftAddon: {
                true: "rounded-l-none",
                false: null,
            },
        },
    }),
    flagSelect: cva([
        "UI-PhoneInput__flagSelect",
        "absolute top-0 left-0 w-[3rem] z-[0] flex-none cursor-pointer truncate rounded-r-none border-r-0",
    ], {
        variants: {
            hasLeftAddon: {
                true: "rounded-l-none border-l-0",
                false: null,
            },
        },
    }),
    flagImage: cva([
        "UI-PhoneInput__flagImage",
        "w-6 absolute h-full inset-y-0 z-[0]",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * PhoneInput
 * -----------------------------------------------------------------------------------------------*/

export type PhoneInputProps = Omit<React.ComponentPropsWithoutRef<"input">, "value" | "size"> &
    ComponentAnatomy<typeof PhoneInputAnatomy> &
    InputStyling &
    BasicFieldOptions & {
    /**
     * The phone number value.
     */
    value?: string
    /**
     * Default phone number when uncontrolled.
     */
    defaultValue?: string
    /**
     * The default country to select if the value is empty.
     */
    defaultCountry?: CountryCode
    /**
     * Callback fired when the phone number value changes.
     */
    onValueChange?: (value: E164Number | undefined) => void
    /**
     * Callback fired when the country changes.
     */
    onCountryChange?: (country: Country) => void
    /**
     * The countries to display in the dropdown.
     */
    countries?: CountryCode[]
}

export type { CountryCode, E164Number, Country }

export const PhoneInput = React.forwardRef<HTMLInputElement, PhoneInputProps>((props, ref) => {

    const [props1, basicFieldProps] = extractBasicFieldProps<PhoneInputProps>(props, React.useId())

    const [{
        size,
        intent,
        rightAddon,
        rightIcon,
        leftAddon,
        leftIcon,
        className,
        value: controlledValue,
        onValueChange,
        defaultCountry,
        onCountryChange,
        countries,
        defaultValue,
        /**/
        countrySelectClass,
        flagSelectClass,
        flagImageClass,
        containerClass,
        ...rest
    }, {
        inputContainerProps,
        rightAddonProps,
        rightIconProps,
        leftIconProps,
        leftAddonProps,
    }] = extractInputPartProps<PhoneInputProps>({
        ...props1,
        size: props1.size ?? "md",
        intent: props1.intent ?? "basic",
        rightAddon: props1.rightAddon,
        rightIcon: props1.rightIcon,
    })

    const isFirst = React.useRef(true)

    const _defaults = React.useMemo(() => {
        try {
            return {
                phoneNumber: controlledValue ?? defaultValue,
                parsedNumber: parsePhoneNumber((controlledValue ?? defaultValue) || "", defaultCountry),
            }
        }
        catch (e) {
            return {
                phoneNumber: controlledValue ?? defaultValue,
                parsedNumber: undefined,
            }
        }
    }, [])

    const [_value, _setValue] = React.useState<E164Number | undefined>(_defaults.phoneNumber)

    const handleOnValueChange = React.useCallback((value: E164Number | undefined) => {
        _setValue(value)
        onValueChange?.(value)
    }, [])

    const handleOnCountryChange = React.useCallback((country: Country) => {
        onCountryChange?.(country)
    }, [])

    React.useEffect(() => {
        if (!defaultValue || !isFirst.current) {
            _setValue(controlledValue)
        }
        isFirst.current = false
    }, [controlledValue])


    return (
        <BasicField {...basicFieldProps}>
            <InputContainer {...inputContainerProps}>
                <InputAddon {...leftAddonProps} />
                {leftAddon && <InputIcon {...leftIconProps} />}

                <PhoneInputPrimitive
                    ref={ref as any}
                    id={basicFieldProps.id}
                    // name={basicFieldProps.name}
                    className={cn(
                        PhoneInputAnatomy.container(),
                        containerClass,
                    )}
                    countries={countries}
                    defaultCountry={defaultCountry || _defaults.parsedNumber?.country}
                    onCountryChange={handleOnCountryChange}
                    addInternationalOption={false}
                    international={false}
                    disabled={basicFieldProps.disabled || basicFieldProps.readonly}
                    countrySelectProps={{
                        name: basicFieldProps.name + "_country",
                        className: cn(
                            "form-select",
                            InputAnatomy.root({
                                size,
                                intent,
                                hasError: !!basicFieldProps.error,
                                isDisabled: !!basicFieldProps.disabled,
                                hasLeftAddon: !!leftAddon,
                                hasLeftIcon: !!leftIcon,
                            }),
                            PhoneInputAnatomy.countrySelect({
                                hasLeftAddon: !!leftAddon,
                            }),
                        ),
                        disabled: basicFieldProps.disabled || basicFieldProps.readonly,
                        "data-disabled": basicFieldProps.disabled,
                        "data-readonly": basicFieldProps.readonly,
                        "aria-readonly": basicFieldProps.readonly,
                    }}
                    numberInputProps={{
                        className: cn(
                            "form-input",
                            InputAnatomy.root({
                                size,
                                intent,
                                hasError: !!basicFieldProps.error,
                                isDisabled: !!basicFieldProps.disabled,
                                hasRightAddon: !!rightAddon,
                                hasRightIcon: !!rightIcon,
                            }),
                            PhoneInputAnatomy.root(),
                            className,
                        ),
                        disabled: basicFieldProps.disabled || basicFieldProps.readonly,
                        required: basicFieldProps.required,
                        "data-disabled": basicFieldProps.disabled,
                        "data-readonly": basicFieldProps.readonly,
                        "aria-readonly": basicFieldProps.readonly,
                        ...rest,
                    }}
                    flagComponent={flag => (
                        <button
                            className={cn(
                                InputAnatomy.root({
                                    size,
                                    intent,
                                    hasError: !!basicFieldProps.error,
                                    isDisabled: !!basicFieldProps.disabled,
                                }),
                                PhoneInputAnatomy.flagSelect({
                                    hasLeftAddon: !!leftAddon,
                                }),
                                flagSelectClass,
                            )}
                            disabled={basicFieldProps.disabled || basicFieldProps.readonly}
                            data-disabled={basicFieldProps.disabled}
                            tabIndex={-1}
                        >
                            <img
                                aria-hidden="true"
                                className={cn(PhoneInputAnatomy.flagImage(), flagImageClass)}
                                src={flag.flagUrl?.replace("{XX}", flag.country)}
                                alt={flag.country}
                            />
                        </button>
                    )}
                    value={_value}
                    onChange={handleOnValueChange}
                />

                <input
                    type="text"
                    value={_value || ""}
                    name={basicFieldProps.name}
                    aria-hidden="true"
                    hidden
                    tabIndex={-1}
                    onChange={() => {}}
                />

                <InputAddon {...rightAddonProps} />
                <InputIcon {...rightIconProps} />
            </InputContainer>
        </BasicField>
    )

})

PhoneInput.displayName = "PhoneInput"
