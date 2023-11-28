"use client"

import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import * as combobox from "@zag-js/combobox"
import { normalizeProps, useMachine } from "@zag-js/react"
import { cva } from "class-variance-authority"
import _find from "lodash/find"
import _isEmpty from "lodash/isEmpty"
import React, { startTransition, useEffect, useId, useMemo, useState } from "react"
import { BasicField, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"
import { InputAddon, InputAnatomy, inputContainerStyle, InputIcon, InputStyling } from "../input"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const ComboboxAnatomy = defineStyleAnatomy({
    menuContainer: cva([
        "UI-Combobox__menuContainer",
        "absolute z-10 -bottom-0.5",
        "left-0 translate-y-full max-h-56 w-full overflow-auto rounded-[--radius] p-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm",
        "bg-[--paper] border border-[--border]",
    ]),
    menuItem: cva([
        "UI-Combobox__menuItem",
        "relative cursor-pointer py-2 pl-3 pr-9 rounded-[--radius] data-[highlighted]:bg-[--highlight] text-base",
    ]),
    menuNoOptionText: cva([
        "UI-Combobox__menuNoOptionText",
        "text-base text-center py-1 text-gray-500 dark:text-gray-700",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Combobox
 * -----------------------------------------------------------------------------------------------*/

export interface ComboboxProps extends Omit<React.ComponentPropsWithRef<"input">, "onChange" | "size" | "defaultChecked">,
    BasicFieldOptions,
    InputStyling,
    ComponentWithAnatomy<typeof ComboboxAnatomy> {
    options: { value: string, label?: string }[]
    /**
     * Filter the specified options as the user is typing
     */
    withFiltering?: boolean
    /**
     * Get the value on of the input as the user is typing
     * @param value
     */
    onInputChange?: (value: string) => void
    /**
     * Get the selected value
     * @param value
     */
    onChange?: (value: string | undefined) => void
    placeholder?: string
    /**
     * Message to display when there are no options
     */
    noOptionsMessage?: string
    /**
     * Allow the user to enter custom values that are not specified in the options
     */
    allowCustomValue?: boolean
    /**
     * Allow the user to enter custom values that are not specified in the options
     */
    defaultValue?: string
    /**
     * Control the value
     */
    value?: string
    valueInputRef?: React.Ref<HTMLInputElement>
    /**
     * We can either return the value or label of the options.
     * Returning the label is useful if users can enter custom values or if the selection doesn't depend on IDs.
     */
    returnValueOrLabel?: "value" | "label"
}

export const Combobox = React.forwardRef<HTMLInputElement, ComboboxProps>((props, ref) => {

    const [{
        size,
        intent,
        leftIcon,
        leftAddon,
        rightIcon,
        rightAddon,
        children,
        className,
        options,
        withFiltering = true,
        placeholder,
        noOptionsMessage,
        allowCustomValue = false,
        onInputChange,
        valueInputRef,
        defaultValue,
        onChange,
        value,
        returnValueOrLabel = "value",
        menuContainerClassName,
        menuItemClassName,
        menuNoOptionTextClassName,
        ...rest
    }, { ...basicFieldProps }] = extractBasicFieldProps<ComboboxProps>(props, useId())

    const [data, setData] = useState(options)

    const [selectedValue, setSelectedValue] = useState<string | undefined>(undefined)

    const [state, send] = useMachine(
        combobox.machine({
            id: basicFieldProps.id,
            allowCustomValue: allowCustomValue,
            inputBehavior: "autohighlight",
            openOnClick: true,
            loop: true,
            blurOnSelect: true,
            placeholder: placeholder,
            onOpen() {
                startTransition(() => {
                    setData(options)
                })
            },
            onSelect: (details) => {
                startTransition(() => {
                    if (returnValueOrLabel === "value") {
                        setSelectedValue(details.value)
                        onChange && onChange(details.value)

                    } else if (returnValueOrLabel === "label") {
                        setSelectedValue(details.label)
                        onChange && onChange(details.label)
                    }
                })
            },
            onInputChange({ value }) {
                onInputChange && onInputChange(value)
                startTransition(() => {
                    if (withFiltering) {
                        const filtered = options.filter((item) => {
                                if (item.label) {
                                    return item.label.toLowerCase().includes(value.toLowerCase())
                                } else {
                                    return item.value.toLowerCase().includes(value.toLowerCase())
                                }
                            },
                        )
                        // Do not empty options if there is no 'noOptionsMessage'
                        setData(filtered.length > 0 ? filtered : noOptionsMessage ? [] : data)
                    }
                })
            },
        }),
    )

    const api = combobox.connect(state, send, normalizeProps)

    // Set default value
    useEffect(() => {
        if (returnValueOrLabel === "value") {
            if (defaultValue) {
                setSelectedValue(defaultValue)
                api.setInputValue(_find(options, ["value", defaultValue])?.label ?? "")
                api.setValue(_find(options, ["value", defaultValue])?.value ?? "")
            }
        }
        if (returnValueOrLabel === "label") {
            if (defaultValue) {
                setSelectedValue(_find(options, ["label", defaultValue])?.value ?? defaultValue)
                api.setInputValue(_find(options, ["label", defaultValue])?.label ?? defaultValue)
                api.setValue(_find(options, ["label", defaultValue])?.value ?? defaultValue)
            }
        }
    }, [defaultValue])

    // Control the state
    useEffect(() => {
        value && setSelectedValue(value)
    }, [value])

    const list = useMemo(() => {
        return withFiltering ? data : options
    }, [options, withFiltering, data])

    return (
        <>
            <BasicField
                {...basicFieldProps}
                ref={ref}
            >
                <input type="text" hidden value={selectedValue ?? ""} onChange={() => {
                }} ref={valueInputRef}/>

                <div {...api.rootProps}>
                    <div {...api.controlProps} className={cn(inputContainerStyle())}>

                        <InputAddon addon={leftAddon} rightIcon={rightIcon} leftIcon={leftIcon} size={size}
                                    side={"left"}/>
                        <InputIcon icon={leftIcon} size={size} side={"left"} props={api.triggerProps}/>

                        <input
                            className={cn(
                                "appearance-none",
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
                            )}
                            disabled={basicFieldProps.isDisabled}
                            onBlur={() => {
                                // If we do not allow custom values and the user blurs the input, we reset the input
                                startTransition(() => {
                                    if (!allowCustomValue) {
                                        if (options.length === 0 && !api.selectedValue || (api.selectedValue && api.selectedValue.length === 0)) {
                                            api.setInputValue("")
                                        }

                                        if (
                                            options.length > 0 &&
                                            (!_isEmpty(_find(options, ["value", api.selectedValue])?.label)
                                                || !_isEmpty(_find(options, ["value", api.selectedValue])?.value)
                                            )
                                        ) {
                                            api.selectedValue && api.setValue(api.selectedValue)
                                        }
                                    }
                                })
                            }}
                            {...rest}
                            ref={ref}
                            {...api.inputProps}
                        />

                        <InputAddon addon={rightAddon} rightIcon={rightIcon} leftIcon={leftAddon} size={size}
                                    side={"right"}/>
                        <InputIcon icon={rightIcon} size={size} side={"right"} props={api.triggerProps}/>

                    </div>
                </div>

                {/* Menu */}
                <div {...api.positionerProps} className="z-10">
                    {(!!noOptionsMessage || list.length > 0) && (
                        <ul
                            className={cn(ComboboxAnatomy.menuContainer(), menuContainerClassName)}
                            {...api.contentProps}
                        >
                            {(list.length === 0 && !!noOptionsMessage) &&
                                <div
                                    className={cn(ComboboxAnatomy.menuNoOptionText(), menuNoOptionTextClassName)}>{noOptionsMessage}</div>}
                            {list.map((item, index) => (
                                <li
                                    className={cn(
                                        ComboboxAnatomy.menuItem(),
                                        menuItemClassName,
                                    )}
                                    key={`combobox:${item.value}:${index}`}
                                    {...api.getOptionProps({
                                        label: item.label ?? item.value,
                                        value: item.value,
                                        index,
                                        disabled: basicFieldProps.isDisabled,
                                    })}
                                >
                                    {item.label ?? item.value}
                                </li>
                            ))}
                        </ul>
                    )}
                </div>
            </BasicField>
        </>
    )

})

Combobox.displayName = "Combobox"
