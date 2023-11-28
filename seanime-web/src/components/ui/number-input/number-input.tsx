"use client"

import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import * as numberInput from "@zag-js/number-input"
import { normalizeProps, useMachine } from "@zag-js/react"
import { cva } from "class-variance-authority"
import React, { useEffect, useId } from "react"
import { BasicField, extractBasicFieldProps } from "../basic-field"
import { InputAddon, InputAnatomy, inputContainerStyle, InputIcon, InputStyling } from "../input"
import type { TextInputProps } from "../text-input"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const NumberInputAnatomy = defineStyleAnatomy({
    input: cva("UI-NumberInput__input", {
        variants: {
            discrete: {
                true: false,
                false: "text-center rounded-none border-l-transparent border-r-transparent hover:border-l-transparent hover:border-r-transparent",
            },
        },
        defaultVariants: {
            discrete: false,
        },
    }),
    control: cva([
            "UI-NumberInput__control",
            "flex flex-none items-center justify-center w-10 border shadow-sm text-lg font-medium",
            "disabled:shadow-none disabled:pointer-events-none",
            "transition",
            "bg-[--paper] hover:bg-gray-50 dark:hover:bg-gray-800 border-[--border] disabled:!bg-gray-50 disabled:!bg-gray-50 disabled:text-gray-300 disabled:border-gray-200",
            "dark:disabled:!bg-gray-800 dark:disabled:border-gray-800 dark:disabled:text-gray-700",
        ], {
            variants: {
                size: { sm: "", md: "", lg: "" },
                position: { left: null, right: null },
                hasLeftAddon: {
                    true: "border-l-0",
                    false: null,
                },
                hasRightAddon: {
                    true: "border-r-0",
                    false: null,
                },
            },
            compoundVariants: [
                { hasRightAddon: false, hasLeftAddon: false, position: "left", className: "rounded-bl-md rounded-tl-md" },
                { hasRightAddon: false, hasLeftAddon: false, position: "right", className: "rounded-br-md rounded-tr-md" },
            ],
            defaultVariants: {
                size: "md",
                hasLeftAddon: false,
                hasRightAddon: false,
            },
        },
    ),
})

/* -------------------------------------------------------------------------------------------------
 * NumberInput
 * -----------------------------------------------------------------------------------------------*/

export interface NumberInputProps extends Omit<TextInputProps, "defaultValue" | "onChange" | "value">, InputStyling,
    ComponentWithAnatomy<typeof NumberInputAnatomy> {
    defaultValue?: number
    value?: number
    onChange?: (value: number) => void
    min?: number
    max?: number
    minFractionDigits?: number
    maxFractionDigits?: number
    precision?: number
    step?: number
    allowMouseWheel?: boolean
    fullWidth?: boolean
    discrete?: boolean
}

export const NumberInput = React.forwardRef<HTMLInputElement, NumberInputProps>((props, ref) => {

    const [{
        children,
        className,
        size,
        intent,
        leftAddon,
        leftIcon,
        rightAddon,
        rightIcon,
        defaultValue = 0,
        placeholder,
        onChange,
        fullWidth,
        discrete,
        value,
        controlClassName,
        inputClassName,
        min = 0, max, minFractionDigits, maxFractionDigits = 2, precision, step, allowMouseWheel = true,
        ...rest
    }, basicFieldProps] = extractBasicFieldProps<NumberInputProps>(props, useId())

    const [state, send] = useMachine(numberInput.machine({
        id: basicFieldProps.id,
        name: basicFieldProps.name,
        disabled: basicFieldProps.isDisabled,
        readOnly: basicFieldProps.isReadOnly,
        value: value ? String(value) : undefined,
        min,
        max,
        minFractionDigits,
        maxFractionDigits,
        step,
        allowMouseWheel,
        clampValueOnBlur: true,
        onChange: (v) => {
            onChange && onChange(v.valueAsNumber)
        },
    }))

    const api = numberInput.connect(state, send, normalizeProps)

    useEffect(() => {
        if (!value) {
            api.setValue(defaultValue)
        }
    }, [])

    useEffect(() => {
        value && api.setValue(value)
    }, [value])

    return (
        <>
            <BasicField
                {...api.rootProps}
                {...basicFieldProps}
            >
                <div className={cn(inputContainerStyle())}>

                    <InputAddon addon={leftAddon} rightIcon={rightIcon} leftIcon={leftIcon} size={size} side={"left"}/>
                    <InputIcon icon={leftIcon} size={size} side={"left"}/>

                    {!discrete && (
                        <button
                            className={cn(NumberInputAnatomy.control({
                                size,
                                position: "left",
                                hasLeftAddon: !!leftAddon || !!leftIcon,
                            }), controlClassName)}
                            {...api.decrementTriggerProps}>
                            -
                        </button>
                    )}

                    <input
                        type="number"
                        name={basicFieldProps.name}
                        className={cn(
                            "form-input",
                            InputAnatomy.input({
                                size,
                                intent,
                                hasError: !!basicFieldProps.error,
                                untouchable: !!basicFieldProps.isDisabled,
                                hasRightAddon: !!rightAddon || !discrete,
                                hasRightIcon: !!rightIcon,
                                hasLeftAddon: !!leftAddon || !discrete,
                                hasLeftIcon: !!leftIcon,
                            }),
                            NumberInputAnatomy.input({ discrete }),
                            inputClassName,
                            className,
                        )}
                        disabled={basicFieldProps.isDisabled}
                        {...api.inputProps}
                        {...rest}
                        ref={ref}
                    />

                    {!discrete && (
                        <button
                            className={cn(NumberInputAnatomy.control({
                                size,
                                position: "right",
                                hasRightAddon: !!rightAddon || !!rightIcon,
                            }), controlClassName)}
                            {...api.incrementTriggerProps}
                        >
                            +
                        </button>
                    )}

                    <InputAddon addon={rightAddon} rightIcon={rightIcon} leftIcon={leftAddon} size={size}
                                side={"right"}/>
                    <InputIcon icon={rightIcon} size={size} side={"right"}/>

                </div>
            </BasicField>
        </>
    )

})

NumberInput.displayName = "NumberInput"