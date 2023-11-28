"use client"

import { cn, ComponentWithAnatomy, defineStyleAnatomy, mergeRefs, useUILocaleConfig } from "../core"
import { cva } from "class-variance-authority"
import React, { useId, useRef } from "react"
import { useDateRangePicker } from "react-aria"
import { DateRangePickerStateOptions, useDateRangePickerState } from "react-stately"
import { BasicField, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"
import { IconButton } from "../button"
import { RangeCalendar } from "../calendar"
import { InputAddon, InputAnatomy, inputContainerStyle, InputIcon, InputStyling } from "../input"
import { Modal } from "../modal"
import { DateField } from "./date-field"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const DateRangePickerAnatomy = defineStyleAnatomy({
    input: cva([
        "UI-DateRangePicker__input",
        "relative flex flex-wrap items-center gap-1 cursor-text",
        "group-focus-within:border-brand-500 group-focus-within:ring-1 group-focus-within:ring-[--ring]",
        "justify-between text-sm sm:text-base",
    ]),
    iconButton: cva([
        "UI-DateRangePicker__iconButton",
        "w-5 h-5 group-focus-within:text-brand-700",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * DateRangePicker
 * -----------------------------------------------------------------------------------------------*/

export interface DateRangePickerProps extends Omit<DateRangePickerStateOptions, "label">,
    ComponentWithAnatomy<typeof DateRangePickerAnatomy>,
    BasicFieldOptions,
    InputStyling {
    locale?: string
}

export const DateRangePicker = React.forwardRef<HTMLDivElement, DateRangePickerProps>((props, ref) => {

    const [{
        size,
        intent,
        leftAddon,
        leftIcon,
        rightIcon,
        rightAddon,
        inputClassName,
        iconButtonClassName,
        locale,
        ...datePickerProps
    }, basicFieldProps] = extractBasicFieldProps<DateRangePickerProps>(props, useId())

    const { countryLocale } = useUILocaleConfig()
    const _locale = locale ?? countryLocale

    const state = useDateRangePickerState(datePickerProps)
    const _ref = mergeRefs(ref, useRef<HTMLDivElement>(null))
    const {
        groupProps,
        labelProps,
        startFieldProps,
        endFieldProps,
        buttonProps,
        dialogProps,
        calendarProps,
    } = useDateRangePicker({ ...datePickerProps, "aria-label": basicFieldProps.name ?? "no-label" }, state, _ref)

    const { onPress, onFocusChange, ...restButtonProps } = buttonProps

    return (
        <BasicField
            {...basicFieldProps}
            labelProps={labelProps}
        >
            <div {...groupProps} ref={_ref} className={cn("flex group", inputContainerStyle())}>

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
                        DateRangePickerAnatomy.input(),
                        inputClassName,
                    )}
                >
                    <div className="flex">
                        <DateField {...startFieldProps} locale={_locale}/>
                        <span aria-hidden="true" className="px-1"> – </span>
                        <DateField {...endFieldProps} locale={_locale}/>
                    </div>
                    <IconButton
                        intent="gray-basic"
                        size="xs"
                        {...restButtonProps}
                        icon={<svg
                            xmlns="http://www.w3.org/2000/svg"
                            fill="currentColor"
                            viewBox="0 0 24 24"
                            className={cn(DateRangePickerAnatomy.iconButton(), iconButtonClassName)}
                        >
                            <path
                                d="M3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6a2 2 0 0 0-2-2h-2V2h-2v2H9V2H7v2H5a2 2 0 0 0-2 2zm16 14H5V8h14z"></path>
                        </svg>}
                        onClick={e => onPress && onPress(e as any)}
                    />


                </div>

                <InputAddon addon={rightAddon} rightIcon={rightIcon} leftIcon={leftAddon} size={size} side={"right"}/>
                <InputIcon icon={rightIcon} size={size} side={"right"}/>

                <Modal
                    size="xl" isOpen={state.isOpen} onClose={state.close} isClosable
                    {...dialogProps}
                >
                    <div className="flex justify-center"><RangeCalendar {...calendarProps} locale={_locale}/></div>
                </Modal>

            </div>
        </BasicField>)
})

DateRangePicker.displayName = "DateRangePicker"
