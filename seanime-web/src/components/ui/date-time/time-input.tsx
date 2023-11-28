"use client"

import { cn, ComponentWithAnatomy, defineStyleAnatomy, mergeRefs, useUILocaleConfig } from "../core"
import { cva } from "class-variance-authority"
import React, { useId, useRef } from "react"
import { useTimeField } from "react-aria"
import { TimeFieldStateOptions, useTimeFieldState } from "react-stately"
import { BasicField, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"
import { InputAddon, InputAnatomy, inputContainerStyle, InputIcon, InputStyling } from "../input"
import { DateSegmentComponent } from "./date-field"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const TimeInputAnatomy = defineStyleAnatomy({
    input: cva([
        "UI-TimeInput__input",
        "relative flex flex-wrap items-center gap-1 cursor-text",
        "group-focus-within:border-brand-500 group-focus-within:ring-1 group-focus-within:ring-[--ring]",
        "!w-fit",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * TimeInput
 * -----------------------------------------------------------------------------------------------*/

export interface TimeInputProps extends Omit<TimeFieldStateOptions, "locale" | "label">,
    ComponentWithAnatomy<typeof TimeInputAnatomy>,
    BasicFieldOptions,
    InputStyling {
    locale?: string
}

export const TimeInput = React.forwardRef<HTMLDivElement, TimeInputProps>((props, ref) => {

    const [{
        size,
        intent,
        leftAddon,
        leftIcon = <span>
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" className="w-4 h-4">
                <path
                    d="M8 0a8 8 0 1 1 0 16A8 8 0 0 1 8 0ZM1.5 8a6.5 6.5 0 1 0 13 0 6.5 6.5 0 0 0-13 0Zm7-3.25v2.992l2.028.812a.75.75 0 0 1-.557 1.392l-2.5-1A.751.751 0 0 1 7 8.25v-3.5a.75.75 0 0 1 1.5 0Z"></path>
            </svg>
        </span>,
        rightIcon,
        rightAddon,
        inputClassName,
        locale,
        ...datePickerProps
    }, basicFieldProps] = extractBasicFieldProps<TimeInputProps>(props, useId())

    const { countryLocale } = useUILocaleConfig()
    const state = useTimeFieldState({
        ...datePickerProps,
        locale: locale ?? countryLocale,
    })

    const _ref = mergeRefs(ref, useRef<HTMLDivElement>(null))
    const { labelProps, fieldProps } = useTimeField(datePickerProps, state, _ref)

    return (
        <BasicField
            {...basicFieldProps}
            labelProps={labelProps}
        >
            <div className={cn(inputContainerStyle(), "!w-fit")}>

                <InputAddon addon={leftAddon} rightIcon={rightIcon} leftIcon={leftIcon} size={size} side={"left"}/>
                <InputIcon icon={leftIcon} size={size} side={"left"}/>

                <div
                    className={cn(
                        "form-input",
                        InputAnatomy.input({
                            size,
                            intent,
                            hasError: !!basicFieldProps.error,
                            untouchable: !!basicFieldProps.isDisabled,
                            hasRightAddon: !!rightAddon,
                            hasRightIcon: !!rightIcon,
                            hasLeftAddon: !!leftAddon,
                            hasLeftIcon: !!leftIcon,
                        }),
                        TimeInputAnatomy.input(),
                        inputClassName,
                    )}
                    {...fieldProps}
                    ref={_ref}
                >
                    {state.segments.map((segment, i) => (
                        <DateSegmentComponent key={i} segment={segment} state={state}/>
                    ))}
                </div>

                <InputAddon addon={rightAddon} rightIcon={rightIcon} leftIcon={leftAddon} size={size} side={"right"}/>
                <InputIcon icon={rightIcon} size={size} side={"right"}/>

            </div>
        </BasicField>
    )

})

TimeInput.displayName = "TimeInput"
