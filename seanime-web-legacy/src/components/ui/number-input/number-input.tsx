"use client"

import type { IntlTranslations } from "@zag-js/number-input"
import * as numberInput from "@zag-js/number-input"
import { normalizeProps, useMachine } from "@zag-js/react"
import { cva, VariantProps } from "class-variance-authority"
import * as React from "react"
import { BasicField, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"
import { IconButton } from "../button"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"
import { extractInputPartProps, InputAddon, InputAnatomy, InputContainer, InputIcon, InputStyling } from "../input"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const NumberInputAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-NumberInput__root",
        "z-[2]",
    ], {
        variants: {
            hideControls: {
                true: false,
                false: "border-r border-r-transparent hover:border-r-[--border]",
            },
            size: {
                sm: null,
                md: null,
                lg: null,
            },
            intent: {
                basic: null,
                filled: null,
                unstyled: "border-r-0 hover:border-r-transparent",
            },
        },
        defaultVariants: {
            hideControls: false,
        },
    }),
    control: cva([
        "UI-NumberInput__control",
        "rounded-none h-[50%] ring-inset",
    ]),
    controlsContainer: cva([
        "UI-NumberInput__controlsContainer",
        "form-input w-auto p-0 flex flex-col items-stretch justify-center overflow-hidden max-h-full",
        "border-l-0 relative z-[1]",
        "shadow-xs",
    ], {
        variants: {
            size: {
                sm: "h-8",
                md: "h-10",
                lg: "h-12",
            },
            intent: {
                basic: null,
                filled: "hover:bg-gray-100",
                unstyled: null,
            },
            hasRightAddon: {
                true: "border-r-0",
                false: null,
            },
        },
    }),
    chevronIcon: cva([
        "UI-Combobox__chevronIcon",
        "h-4 w-4 shrink-0",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * NumberInput
 * -----------------------------------------------------------------------------------------------*/

export type NumberInputProps = Omit<React.ComponentPropsWithoutRef<"input">, "value" | "size" | "defaultValue"> &
    ComponentAnatomy<typeof NumberInputAnatomy> &
    Omit<VariantProps<typeof NumberInputAnatomy.root>, "size" | "intent"> &
    BasicFieldOptions &
    InputStyling & {
    /**
     * The value of the input
     */
    value?: number | string
    /**
     * The callback to handle value changes
     */
    onValueChange?: (value: number, valueAsString: string) => void
    /**
     * Default value when uncontrolled
     */
    defaultValue?: number | string
    /**
     * The minimum value of the input
     */
    min?: number
    /**
     * The maximum value of the input
     */
    max?: number
    /**
     * The amount to increment or decrement the value by
     */
    step?: number
    /**
     * Whether to allow mouse wheel to change the value
     */
    allowMouseWheel?: boolean
    /**
     * Whether to allow the value overflow the min/max range
     */
    allowOverflow?: boolean
    /**
     * Whether to hide the controls
     */
    hideControls?: boolean
    /**
     * The format options for the value
     */
    formatOptions?: Intl.NumberFormatOptions
    /**
     * Whether to clamp the value when the input loses focus (blur)
     */
    clampValueOnBlur?: boolean
    /**
     * Accessibility
     *
     * Specifies the localized strings that identifies the accessibility elements and their states
     */
    translations?: IntlTranslations,
    /**
     * The current locale. Based on the BCP 47 definition.
     */
    locale?: string
    /**
     * The document's text/writing direction.
     */
    dir?: "ltr" | "rtl"
}

export const NumberInput = React.forwardRef<HTMLInputElement, NumberInputProps>((props, ref) => {

    const [props1, basicFieldProps] = extractBasicFieldProps<NumberInputProps>(props, React.useId())

    const [{
        controlClass,
        controlsContainerClass,
        chevronIconClass,
        className,
        children,
        /**/
        size,
        intent,
        leftAddon,
        leftIcon,
        rightAddon,
        rightIcon,
        placeholder,
        onValueChange,
        hideControls,
        value: controlledValue,
        min = 0,
        max,
        step,
        allowMouseWheel = true,
        formatOptions = { maximumFractionDigits: 2 },
        clampValueOnBlur = true,
        translations,
        locale,
        dir,
        defaultValue,
        ...rest
    }, {
        inputContainerProps,
        leftAddonProps,
        leftIconProps,
        rightAddonProps,
        rightIconProps,
    }] = extractInputPartProps<NumberInputProps>({
        ...props1,
        size: props1.size ?? "md",
        intent: props1.intent ?? "basic",
        leftAddon: props1.leftAddon,
        leftIcon: props1.leftIcon,
        rightAddon: props1.rightAddon,
        rightIcon: props1.rightIcon,
    })
    const service = useMachine(numberInput.machine, {
        id: basicFieldProps.id,
        name: basicFieldProps.name,
        disabled: basicFieldProps.disabled,
        readOnly: basicFieldProps.readonly,
        value: controlledValue ? String(controlledValue) : (defaultValue ? String(defaultValue) : undefined),
        min,
        max,
        step,
        allowMouseWheel,
        formatOptions,
        clampValueOnBlur,
        translations,
        locale,
        dir,
        onValueChange: (details: { valueAsNumber: number; value: string }) => {
            onValueChange?.(details.valueAsNumber, details.value)
        },
    })

    const api = numberInput.connect(service, normalizeProps)

    const isFirst = React.useRef(true)

    React.useEffect(() => {
        if (!isFirst.current) {
            if (typeof controlledValue === "string" && !isNaN(Number(controlledValue))) {
                api.setValue(Number(controlledValue))
            } else if (typeof controlledValue === "number") {
                api.setValue(controlledValue)
            } else if (controlledValue === undefined) {
                api.setValue(min)
            }
        }
        isFirst.current = false
    }, [controlledValue])

    return (
        <BasicField
            {...basicFieldProps}
            id={api.getInputProps().id}
        >
            <InputContainer {...inputContainerProps}>
                <InputAddon {...leftAddonProps} />
                <InputIcon {...leftIconProps} />

                <input
                    ref={ref}
                    type="number"
                    name={basicFieldProps.name}
                    className={cn(
                        "form-input",
                        InputAnatomy.root({
                            size,
                            intent,
                            hasError: !!basicFieldProps.error,
                            isDisabled: !!basicFieldProps.disabled,
                            hasRightAddon: !!rightAddon || !hideControls,
                            hasRightIcon: !!rightIcon,
                            hasLeftAddon: !!leftAddon,
                            hasLeftIcon: !!leftIcon,
                        }),
                        NumberInputAnatomy.root({ hideControls, intent, size }),
                        className,
                    )}
                    disabled={basicFieldProps.disabled || basicFieldProps.readonly}
                    data-disabled={basicFieldProps.disabled}
                    data-readonly={basicFieldProps.readonly}
                    aria-readonly={basicFieldProps.readonly}
                    required={basicFieldProps.required}
                    {...api.getInputProps()}
                    {...rest}
                />

                {!hideControls && (<div
                    className={cn(
                        InputAnatomy.root({
                            size,
                            intent,
                            hasError: !!basicFieldProps.error,
                            isDisabled: !!basicFieldProps.disabled,
                            hasRightAddon: !!rightAddon,
                            hasRightIcon: !!rightIcon,
                            hasLeftAddon: true,
                        }),
                        NumberInputAnatomy.controlsContainer({
                            size,
                            intent,
                            hasRightAddon: !!rightAddon,
                        }),
                        controlsContainerClass,
                    )}
                >
                    <IconButton
                        intent="gray-basic"
                        size="sm"
                        className={cn(
                            NumberInputAnatomy.control(),
                            controlClass,
                        )}
                        {...api.getIncrementTriggerProps()}
                        data-readonly={basicFieldProps.readonly}
                        data-disabled={basicFieldProps.disabled || api.getIncrementTriggerProps().disabled}
                        disabled={basicFieldProps.disabled || basicFieldProps.readonly || api.getIncrementTriggerProps().disabled}
                        tabIndex={0}
                        icon={<svg
                            xmlns="http://www.w3.org/2000/svg"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            strokeWidth="2"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            className={cn(NumberInputAnatomy.chevronIcon(), "rotate-180", chevronIconClass)}
                        >
                            <path d="m6 9 6 6 6-6" />
                        </svg>}
                    />
                    <IconButton
                        intent="gray-basic"
                        size="sm"
                        className={cn(
                            NumberInputAnatomy.control(),
                            controlClass,
                        )}
                        {...api.getDecrementTriggerProps()}
                        data-readonly={basicFieldProps.readonly}
                        data-disabled={basicFieldProps.disabled || api.getDecrementTriggerProps().disabled}
                        disabled={basicFieldProps.disabled || basicFieldProps.readonly || api.getDecrementTriggerProps().disabled}
                        tabIndex={0}
                        icon={<svg
                            xmlns="http://www.w3.org/2000/svg"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            strokeWidth="2"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            className={cn(NumberInputAnatomy.chevronIcon(), chevronIconClass)}
                        >
                            <path d="m6 9 6 6 6-6" />
                        </svg>}
                    />
                </div>)}

                <InputAddon {...rightAddonProps} />
                <InputIcon
                    {...rightIconProps}
                    className={cn(
                        "z-3",
                        rightIconProps.className,
                        !rightAddon ? "mr-6" : null,
                    )}
                />
            </InputContainer>
        </BasicField>
    )

})

NumberInput.displayName = "NumberInput"
